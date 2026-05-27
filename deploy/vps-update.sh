#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
COMPOSE_DIR="${REPO_ROOT}/deploy/docker-compose"
COMPOSE_FILE="${COMPOSE_DIR}/docker-compose.yaml"

SERVER_HEALTH_URL="${SERVER_HEALTH_URL:-http://127.0.0.1:8888/health}"
WEB_HEALTH_URL="${WEB_HEALTH_URL:-http://127.0.0.1:8080/}"
WAIT_RETRIES="${WAIT_RETRIES:-30}"
WAIT_INTERVAL="${WAIT_INTERVAL:-2}"
LOG_TAIL="${LOG_TAIL:-80}"

TARGET="all"
BRANCH="main"
USE_PULL=0
GIT_REF=""
ALLOW_DIRTY=0
NO_CACHE=0
SKIP_HEALTHCHECK=0

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

info() {
  echo -e "${GREEN}[INFO]${NC} $*"
}

warn() {
  echo -e "${YELLOW}[WARN]${NC} $*"
}

die() {
  echo -e "${RED}[ERROR]${NC} $*" >&2
  exit 1
}

usage() {
  cat <<'EOF'
用法:
  bash deploy/vps-update.sh [server|web|all] [options]

说明:
  - 默认基于当前服务器工作树本地构建，不会强制 reset
  - 适合在 /opt/tiny-sec 这类已存在定制的仓库中执行
  - server 更新会自动确保 mysql / redis 已启动

选项:
  --pull                从 origin/<branch> 快进更新当前分支后再构建
  --branch <name>       配合 --pull 使用，默认 main
  --ref <git-ref>       切到指定 ref 的 detached HEAD 后再构建
  --allow-dirty         允许在有本地改动时继续执行（仅构建当前工作树）
  --no-cache            docker compose build 时禁用缓存
  --skip-healthcheck    跳过 HTTP 健康检查
  --tail <n>            最后输出日志行数，默认 80
  -h, --help            查看帮助

示例:
  sudo bash deploy/vps-update.sh all --pull
  sudo bash deploy/vps-update.sh server --no-cache
  sudo bash deploy/vps-update.sh web --tail 120
EOF
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "缺少命令: $1"
}

run_compose() {
  (
    cd "${COMPOSE_DIR}"
    docker compose -f "${COMPOSE_FILE}" "$@"
  )
}

git_has_tracked_changes() {
  ! git -C "${REPO_ROOT}" diff --quiet || ! git -C "${REPO_ROOT}" diff --cached --quiet
}

current_branch() {
  git -C "${REPO_ROOT}" rev-parse --abbrev-ref HEAD
}

current_revision() {
  git -C "${REPO_ROOT}" rev-parse --short HEAD
}

wait_for_container_state() {
  local container_name="$1"
  local expected_state="$2"
  local attempt state

  for attempt in $(seq 1 "${WAIT_RETRIES}"); do
    state="$(docker inspect -f '{{if .State.Health}}{{.State.Health.Status}}{{else}}{{.State.Status}}{{end}}' "${container_name}" 2>/dev/null || true)"
    if [[ "${state}" == "${expected_state}" ]]; then
      info "${container_name} 状态正常: ${state}"
      return 0
    fi
    sleep "${WAIT_INTERVAL}"
  done

  die "${container_name} 未进入期望状态: ${expected_state}，当前状态: ${state:-unknown}"
}

wait_for_http() {
  local name="$1"
  local url="$2"
  local attempt

  for attempt in $(seq 1 "${WAIT_RETRIES}"); do
    if curl -fsS --max-time 5 "${url}" >/dev/null; then
      info "${name} 健康检查通过: ${url}"
      return 0
    fi
    sleep "${WAIT_INTERVAL}"
  done

  return 1
}

print_logs() {
  local service
  for service in "$@"; do
    echo
    info "${service} 最近 ${LOG_TAIL} 行日志"
    docker logs --tail "${LOG_TAIL}" "gva-${service}" 2>&1 || true
  done
}

update_code_if_needed() {
  if [[ -z "${GIT_REF}" && "${USE_PULL}" -eq 0 ]]; then
    info "使用当前工作树构建，当前 revision: $(current_revision)"
    return 0
  fi

  if git_has_tracked_changes && [[ "${ALLOW_DIRTY}" -ne 1 ]]; then
    die "检测到已跟踪文件存在本地改动；如需继续，请先提交/清理，或显式传入 --allow-dirty"
  fi

  if [[ -n "${GIT_REF}" ]]; then
    info "拉取远端引用并切换到 detached HEAD: ${GIT_REF}"
    git -C "${REPO_ROOT}" fetch --all --tags --prune
    git -C "${REPO_ROOT}" checkout --detach "${GIT_REF}"
    info "当前 revision: $(current_revision)"
    return 0
  fi

  local branch_now
  branch_now="$(current_branch)"
  [[ "${branch_now}" == "${BRANCH}" ]] || die "--pull 模式要求当前分支为 ${BRANCH}，当前是 ${branch_now}"

  info "从 origin/${BRANCH} 快进更新代码"
  git -C "${REPO_ROOT}" fetch origin "${BRANCH}"
  git -C "${REPO_ROOT}" merge --ff-only "origin/${BRANCH}"
  info "当前 revision: $(current_revision)"
}

build_services() {
  local build_args=()
  [[ "${NO_CACHE}" -eq 1 ]] && build_args+=(--no-cache)

  case "${TARGET}" in
    server)
      run_compose build "${build_args[@]}" server
      ;;
    web)
      run_compose build "${build_args[@]}" web
      ;;
    all)
      run_compose build "${build_args[@]}" server web
      ;;
    *)
      die "未知目标: ${TARGET}"
      ;;
  esac
}

start_services() {
  case "${TARGET}" in
    server)
      info "确保 mysql / redis 已启动"
      run_compose up -d mysql redis
      wait_for_container_state "gva-mysql" "healthy"
      wait_for_container_state "gva-redis" "healthy"
      run_compose up -d server
      ;;
    web)
      run_compose up -d --no-deps web
      ;;
    all)
      info "确保 mysql / redis 已启动"
      run_compose up -d mysql redis
      wait_for_container_state "gva-mysql" "healthy"
      wait_for_container_state "gva-redis" "healthy"
      run_compose up -d server web
      ;;
    *)
      die "未知目标: ${TARGET}"
      ;;
  esac
}

healthcheck() {
  [[ "${SKIP_HEALTHCHECK}" -eq 1 ]] && {
    warn "已跳过健康检查"
    return 0
  }

  case "${TARGET}" in
    server)
      wait_for_http "server" "${SERVER_HEALTH_URL}" || {
        print_logs server
        die "server 健康检查失败"
      }
      ;;
    web)
      wait_for_http "web" "${WEB_HEALTH_URL}" || {
        print_logs web
        die "web 健康检查失败"
      }
      ;;
    all)
      wait_for_http "server" "${SERVER_HEALTH_URL}" || {
        print_logs server
        die "server 健康检查失败"
      }
      wait_for_http "web" "${WEB_HEALTH_URL}" || {
        print_logs server web
        die "web 健康检查失败"
      }
      ;;
  esac
}

parse_args() {
  if [[ $# -gt 0 ]]; then
    case "$1" in
      server|web|all)
        TARGET="$1"
        shift
        ;;
      -h|--help)
        usage
        exit 0
        ;;
    esac
  fi

  while [[ $# -gt 0 ]]; do
    case "$1" in
      --pull)
        USE_PULL=1
        ;;
      --branch)
        [[ $# -ge 2 ]] || die "--branch 需要参数"
        BRANCH="$2"
        shift
        ;;
      --ref)
        [[ $# -ge 2 ]] || die "--ref 需要参数"
        GIT_REF="$2"
        shift
        ;;
      --allow-dirty)
        ALLOW_DIRTY=1
        ;;
      --no-cache)
        NO_CACHE=1
        ;;
      --skip-healthcheck)
        SKIP_HEALTHCHECK=1
        ;;
      --tail)
        [[ $# -ge 2 ]] || die "--tail 需要参数"
        LOG_TAIL="$2"
        shift
        ;;
      -h|--help)
        usage
        exit 0
        ;;
      *)
        die "未知参数: $1"
        ;;
    esac
    shift
  done

  [[ "${USE_PULL}" -eq 1 && -n "${GIT_REF}" ]] && die "--pull 和 --ref 不能同时使用"
}

main() {
  parse_args "$@"

  require_cmd git
  require_cmd docker
  require_cmd curl

  [[ -f "${COMPOSE_FILE}" ]] || die "未找到 compose 文件: ${COMPOSE_FILE}"
  docker compose version >/dev/null 2>&1 || die "当前环境不可用 docker compose"

  info "仓库目录: ${REPO_ROOT}"
  info "更新目标: ${TARGET}"

  update_code_if_needed
  build_services
  start_services
  healthcheck

  echo
  info "容器状态"
  run_compose ps

  case "${TARGET}" in
    server)
      print_logs server
      ;;
    web)
      print_logs web
      ;;
    all)
      print_logs server web
      ;;
  esac

  echo
  info "更新完成，当前 revision: $(current_revision)"
}

main "$@"
