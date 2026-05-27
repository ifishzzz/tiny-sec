#!/usr/bin/env bash
set -euo pipefail

CONFIG_FILE="${FALCO_AGENT_ENV_FILE:-/etc/falco-agent/agent.env}"
STATE_DIR="${FALCO_AGENT_STATE_DIR:-/var/lib/falco-agent}"
STATUS_REPORT_INTERVAL="${FALCO_AGENT_STATUS_INTERVAL:-60}"

if [[ ! -f "${CONFIG_FILE}" ]]; then
  echo "missing config file: ${CONFIG_FILE}" >&2
  exit 1
fi

# shellcheck disable=SC1090
source "${CONFIG_FILE}"

: "${SERVER_URL:?missing SERVER_URL}"
: "${AGENT_ID:?missing AGENT_ID}"
: "${ACCESS_TOKEN:?missing ACCESS_TOKEN}"

HEARTBEAT_INTERVAL="${HEARTBEAT_INTERVAL:-30}"
TASK_PULL_INTERVAL="${TASK_PULL_INTERVAL:-15}"
EVENT_UPLOAD_BATCH_SIZE="${EVENT_UPLOAD_BATCH_SIZE:-200}"
FALCO_FRONTEND="${FALCO_FRONTEND:-noninteractive}"
FALCO_DRIVER_CHOICE="${FALCO_DRIVER_CHOICE:-}"
FALCOCTL_ENABLED="${FALCOCTL_ENABLED:-no}"

mkdir -p "${STATE_DIR}"

if [[ -f /etc/os-release ]]; then
  # shellcheck disable=SC1091
  source /etc/os-release
fi

log() {
  echo "[$(date '+%F %T')] $*"
}

trim_summary() {
  local file_path="${1}"
  if [[ ! -f "${file_path}" ]]; then
    echo ""
    return 0
  fi
  tail -c 4000 "${file_path}" 2>/dev/null || true
}

primary_ip() {
  if command -v hostname >/dev/null 2>&1; then
    local ip_value
    ip_value="$(hostname -I 2>/dev/null | awk '{print $1}')"
    if [[ -n "${ip_value}" ]]; then
      echo "${ip_value}"
      return 0
    fi
  fi
  if command -v ip >/dev/null 2>&1; then
    ip route get 1.1.1.1 2>/dev/null | awk '{for (i=1; i<=NF; i++) if ($i == "src") {print $(i+1); exit}}'
    return 0
  fi
  echo ""
}

system_load() {
  awk '{print $1}' /proc/loadavg 2>/dev/null || echo "0"
}

memory_percent() {
  if command -v free >/dev/null 2>&1; then
    free | awk '/Mem:/ { if ($2 > 0) printf "%.2f", ($3 / $2) * 100; else printf "0" }'
    return 0
  fi
  echo "0"
}

cpu_percent() {
  if command -v top >/dev/null 2>&1; then
    top -bn1 2>/dev/null | awk -F'[, ]+' '/Cpu/ {for (i=1; i<=NF; i++) if ($i ~ /id/) {printf "%.2f", 100 - $(i-1); exit}}'
    return 0
  fi
  echo "0"
}

detect_package_manager() {
  if command -v apt-get >/dev/null 2>&1; then
    echo "apt"
    return 0
  fi
  if command -v dnf >/dev/null 2>&1; then
    echo "dnf"
    return 0
  fi
  if command -v yum >/dev/null 2>&1; then
    echo "yum"
    return 0
  fi
  if command -v zypper >/dev/null 2>&1; then
    echo "zypper"
    return 0
  fi
  return 1
}

install_packages() {
  local pkg_manager="$1"
  shift
  local packages=("$@")
  [[ "${#packages[@]}" -eq 0 ]] && return 0

  case "${pkg_manager}" in
    apt)
      export DEBIAN_FRONTEND=noninteractive
      apt-get update
      apt-get install -y "${packages[@]}"
      ;;
    dnf)
      dnf install -y "${packages[@]}"
      ;;
    yum)
      yum install -y "${packages[@]}"
      ;;
    zypper)
      zypper -n install "${packages[@]}"
      ;;
    *)
      echo "unsupported package manager: ${pkg_manager}" >&2
      return 1
      ;;
  esac
}

refresh_package_metadata() {
  local pkg_manager="$1"
  case "${pkg_manager}" in
    apt)
      apt-get update
      ;;
    dnf)
      dnf makecache -y
      ;;
    yum)
      yum makecache -y
      ;;
    zypper)
      zypper --gpg-auto-import-keys -n refresh
      ;;
    *)
      echo "unsupported package manager: ${pkg_manager}" >&2
      return 1
      ;;
  esac
}

ensure_falco_repo() {
  local pkg_manager="$1"

  case "${pkg_manager}" in
    apt)
      install_packages "${pkg_manager}" ca-certificates curl gnupg apt-transport-https
      if [[ ! -f /usr/share/keyrings/falco-archive-keyring.gpg ]]; then
        curl -fsSL https://falco.org/repo/falcosecurity-packages.asc | \
          gpg --dearmor -o /usr/share/keyrings/falco-archive-keyring.gpg
      fi
      if [[ ! -f /etc/apt/sources.list.d/falcosecurity.list ]]; then
        echo "deb [signed-by=/usr/share/keyrings/falco-archive-keyring.gpg] https://download.falco.org/packages/deb stable main" \
          > /etc/apt/sources.list.d/falcosecurity.list
      fi
      ;;
    dnf|yum)
      install_packages "${pkg_manager}" ca-certificates curl gnupg2
      rpm --import https://falco.org/repo/falcosecurity-packages.asc
      if [[ ! -f /etc/yum.repos.d/falcosecurity.repo ]]; then
        curl -fsSL https://falco.org/repo/falcosecurity-rpm.repo -o /etc/yum.repos.d/falcosecurity.repo
      fi
      ;;
    zypper)
      install_packages "${pkg_manager}" ca-certificates curl gpg2
      rpm --import https://falco.org/repo/falcosecurity-packages.asc
      if [[ ! -f /etc/zypp/repos.d/falcosecurity.repo ]]; then
        curl -fsSL https://falco.org/repo/falcosecurity-rpm.repo -o /etc/zypp/repos.d/falcosecurity.repo
      fi
      ;;
    *)
      echo "unsupported package manager: ${pkg_manager}" >&2
      return 1
      ;;
  esac
}

falco_install_env() {
  export FALCO_FRONTEND
  if [[ -n "${FALCO_DRIVER_CHOICE}" ]]; then
    export FALCO_DRIVER_CHOICE
  else
    unset FALCO_DRIVER_CHOICE 2>/dev/null || true
  fi
  export FALCOCTL_ENABLED
}

systemd_unit_exists() {
  local unit_name="${1%.service}"
  systemctl list-unit-files "${unit_name}.service" --no-legend 2>/dev/null | grep -q "^${unit_name}\.service"
}

resolve_falco_service_name() {
  local requested="${1:-falco}"
  requested="${requested%.service}"

  if systemd_unit_exists "${requested}"; then
    echo "${requested}"
    return 0
  fi

  local candidate
  for candidate in falco falco-kmod falco-modern-bpf falco-bpf falco-custom; do
    if systemd_unit_exists "${candidate}"; then
      echo "${candidate}"
      return 0
    fi
  done

  echo "${requested}"
}

falco_service_state() {
  if command -v systemctl >/dev/null 2>&1; then
    local service_name
    service_name="$(resolve_falco_service_name "falco")"
    systemctl is-active "${service_name}" 2>/dev/null || echo "inactive"
    return 0
  fi
  echo "unknown"
}

falco_version() {
  if command -v falco >/dev/null 2>&1; then
    falco --version 2>/dev/null | awk '{print $NF}' | tail -n1
    return 0
  fi
  if command -v rpm >/dev/null 2>&1; then
    rpm -q falco --queryformat '%{VERSION}-%{RELEASE}\n' 2>/dev/null || true
    return 0
  fi
  if command -v dpkg-query >/dev/null 2>&1; then
    dpkg-query -W -f='${Version}\n' falco 2>/dev/null || true
    return 0
  fi
  echo ""
}

post_json() {
  local path="${1}"
  local payload="${2}"
  curl -fsS -X POST "${SERVER_URL}${path}" \
    -H "Content-Type: application/json" \
    -d "${payload}"
}

report_heartbeat() {
  local payload
  payload="$(jq -nc \
    --arg agentId "${AGENT_ID}" \
    --arg accessToken "${ACCESS_TOKEN}" \
    --arg status "online" \
    '{agentId:$agentId, accessToken:$accessToken, status:$status}')"
  post_json "/falco/agent/heartbeat" "${payload}" >/dev/null || true
}

report_status() {
  local payload
  payload="$(jq -nc \
    --arg agentId "${AGENT_ID}" \
    --arg accessToken "${ACCESS_TOKEN}" \
    --arg falcoStatus "$(falco_service_state)" \
    --argjson cpuPercent "$(cpu_percent)" \
    --argjson memoryPercent "$(memory_percent)" \
    --argjson load1 "$(system_load)" \
    --argjson eventCount 0 \
    '{agentId:$agentId, accessToken:$accessToken, cpuPercent:$cpuPercent, memoryPercent:$memoryPercent, load1:$load1, falcoStatus:$falcoStatus, eventCount:$eventCount}')"
  post_json "/falco/agent/status/report" "${payload}" >/dev/null || true
}

save_task_result() {
  local task_id="${1}"
  local status="${2}"
  local stage="${3}"
  local stdout_summary="${4}"
  local stderr_summary="${5}"
  local error_code="${6}"
  local error_message="${7}"
  local rule_package_version="${8}"

  local payload
  payload="$(jq -nc \
    --argjson taskId "${task_id}" \
    --arg agentId "${AGENT_ID}" \
    --arg accessToken "${ACCESS_TOKEN}" \
    --arg status "${status}" \
    --arg stage "${stage}" \
    --arg result "" \
    --arg stdoutSummary "${stdout_summary}" \
    --arg stderrSummary "${stderr_summary}" \
    --arg falcoVersion "$(falco_version)" \
    --arg rulePackageVersion "${rule_package_version}" \
    --arg serviceState "$(falco_service_state)" \
    --arg errorCode "${error_code}" \
    --arg errorMessage "${error_message}" \
    '{taskId:$taskId, agentId:$agentId, accessToken:$accessToken, status:$status, stage:$stage, result:$result, stdoutSummary:$stdoutSummary, stderrSummary:$stderrSummary, falcoVersion:$falcoVersion, rulePackageVersion:$rulePackageVersion, serviceState:$serviceState, errorCode:$errorCode, errorMessage:$errorMessage}')"
  post_json "/falco/agent/task/result" "${payload}" >/dev/null || true
}

prepare_falco_package_install() {
  local pkg_manager
  pkg_manager="$(detect_package_manager)" || {
    echo "no supported package manager found" >&2
    return 1
  }
  falco_install_env
  ensure_falco_repo "${pkg_manager}"
  refresh_package_metadata "${pkg_manager}"
  echo "${pkg_manager}"
}

install_falco_package() {
  local pkg_manager="$1"
  local falco_version_value="$2"

  case "${pkg_manager}" in
    apt)
      if [[ -n "${falco_version_value}" ]]; then
        apt-get install -y "falco=${falco_version_value}" || \
          apt-get install -y "falco=${falco_version_value}*" || \
          apt-get install -y falco
      else
        apt-get install -y falco
      fi
      ;;
    dnf)
      if [[ -n "${falco_version_value}" ]]; then
        dnf install -y "falco-${falco_version_value}" || dnf install -y falco
      else
        dnf install -y falco
      fi
      ;;
    yum)
      if [[ -n "${falco_version_value}" ]]; then
        yum install -y "falco-${falco_version_value}" || yum install -y falco
      else
        yum install -y falco
      fi
      ;;
    zypper)
      if [[ -n "${falco_version_value}" ]]; then
        zypper -n install "falco=${falco_version_value}" || zypper -n install falco
      else
        zypper -n install falco
      fi
      ;;
    *)
      echo "unsupported package manager: ${pkg_manager}" >&2
      return 1
      ;;
  esac
}

upgrade_falco_package() {
  local pkg_manager="$1"
  local falco_version_value="$2"

  if [[ -n "${falco_version_value}" ]]; then
    install_falco_package "${pkg_manager}" "${falco_version_value}"
    return 0
  fi

  case "${pkg_manager}" in
    apt)
      apt-get install -y --only-upgrade falco || apt-get install -y falco
      ;;
    dnf)
      dnf upgrade -y falco || dnf install -y falco
      ;;
    yum)
      yum update -y falco || yum install -y falco
      ;;
    zypper)
      zypper -n update falco || zypper -n install falco
      ;;
    *)
      echo "unsupported package manager: ${pkg_manager}" >&2
      return 1
      ;;
  esac
}

resolve_previous_apt_falco_version() {
  apt-cache madison falco 2>/dev/null | awk 'NR==2 {print $3; exit}'
}

rollback_falco_package() {
  local pkg_manager="$1"
  local falco_version_value="$2"
  local previous_version=""

  case "${pkg_manager}" in
    apt)
      if [[ -z "${falco_version_value}" ]]; then
        previous_version="$(resolve_previous_apt_falco_version)"
        [[ -n "${previous_version}" ]] || {
          echo "unable to determine previous falco version from apt cache" >&2
          return 1
        }
        falco_version_value="${previous_version}"
      fi
      apt-get install -y --allow-downgrades "falco=${falco_version_value}" || \
        apt-get install -y --allow-downgrades "falco=${falco_version_value}*"
      ;;
    dnf)
      if [[ -n "${falco_version_value}" ]]; then
        dnf downgrade -y "falco-${falco_version_value}" || dnf install -y "falco-${falco_version_value}"
      else
        dnf downgrade -y falco
      fi
      ;;
    yum)
      if [[ -n "${falco_version_value}" ]]; then
        yum downgrade -y "falco-${falco_version_value}" || yum install -y "falco-${falco_version_value}"
      else
        yum downgrade -y falco
      fi
      ;;
    zypper)
      if [[ -n "${falco_version_value}" ]]; then
        zypper -n install --oldpackage "falco=${falco_version_value}" || \
          zypper -n install --oldpackage falco
      else
        zypper -n install --oldpackage falco
      fi
      ;;
    *)
      echo "unsupported package manager: ${pkg_manager}" >&2
      return 1
      ;;
  esac
}

enable_and_restart_service() {
  local service_name="${1}"
  service_name="$(resolve_falco_service_name "${service_name}")"
  systemctl enable "${service_name}"
  systemctl restart "${service_name}"
}

reload_or_restart_service() {
  local service_name="${1}"
  service_name="$(resolve_falco_service_name "${service_name}")"
  systemctl reload "${service_name}" || systemctl restart "${service_name}"
}

perform_task_action() {
  local action="$1"
  local falco_version_value="$2"
  local service_name="$3"
  local config_path="$4"

  local pkg_manager=""

  case "${action}" in
    falco.install)
      mkdir -p "$(dirname "${config_path}")"
      pkg_manager="$(prepare_falco_package_install)"
      install_falco_package "${pkg_manager}" "${falco_version_value}"
      enable_and_restart_service "${service_name}"
      ;;
    falco.upgrade)
      pkg_manager="$(prepare_falco_package_install)"
      upgrade_falco_package "${pkg_manager}" "${falco_version_value}"
      enable_and_restart_service "${service_name}"
      ;;
    falco.rollback)
      pkg_manager="$(prepare_falco_package_install)"
      rollback_falco_package "${pkg_manager}" "${falco_version_value}"
      enable_and_restart_service "${service_name}"
      ;;
    falco.reload|falco.rule_publish)
      reload_or_restart_service "${service_name}"
      ;;
    falco.restart)
      systemctl restart "$(resolve_falco_service_name "${service_name}")"
      ;;
    *)
      echo "unsupported action: ${action}" >&2
      return 1
      ;;
  esac
}

execute_task() {
  local task_json="${1}"
  local task_id request_id action payload_json falco_version_value rule_package_version service_name config_path
  task_id="$(echo "${task_json}" | jq -r '.taskId')"
  request_id="$(echo "${task_json}" | jq -r '.requestId // ""')"
  action="$(echo "${task_json}" | jq -r '.action // ""')"
  payload_json="$(echo "${task_json}" | jq -c '.payload // {}')"
  falco_version_value="$(echo "${payload_json}" | jq -r '.falcoVersion // ""')"
  rule_package_version="$(echo "${payload_json}" | jq -r '.rulePackageVersion // ""')"
  service_name="$(echo "${payload_json}" | jq -r '.serviceName // "falco"')"
  config_path="$(echo "${payload_json}" | jq -r '.configPath // "/etc/falco/falco.yaml"')"

  local stdout_file stderr_file error_message error_code
  stdout_file="$(mktemp)"
  stderr_file="$(mktemp)"
  error_message=""
  error_code=""

  log "processing task ${task_id} (${request_id}) action=${action}"
  save_task_result "${task_id}" "running" "running" "" "" "" "" "${rule_package_version}"

  case "${action}" in
    falco.install|falco.upgrade|falco.rollback|falco.reload|falco.restart|falco.rule_publish)
      if perform_task_action "${action}" "${falco_version_value}" "${service_name}" "${config_path}" >"${stdout_file}" 2>"${stderr_file}"; then
        save_task_result "${task_id}" "succeeded" "finished" "$(trim_summary "${stdout_file}")" "$(trim_summary "${stderr_file}")" "" "" "${rule_package_version}"
      else
        error_code="TASK_EXEC_FAILED"
        error_message="command execution failed"
        save_task_result "${task_id}" "failed" "finished" "$(trim_summary "${stdout_file}")" "$(trim_summary "${stderr_file}")" "${error_code}" "${error_message}" "${rule_package_version}"
      fi
      ;;
    *)
      echo "unsupported action: ${action}" >"${stderr_file}"
      error_code="UNSUPPORTED_ACTION"
      error_message="unsupported action: ${action}"
      save_task_result "${task_id}" "failed" "rejected" "" "$(trim_summary "${stderr_file}")" "${error_code}" "${error_message}" "${rule_package_version}"
      ;;
  esac

  rm -f "${stdout_file}" "${stderr_file}"
}

poll_tasks() {
  local response list_json
  response="$(curl -fsS "${SERVER_URL}/falco/agent/task/pull?agentId=${AGENT_ID}&accessToken=${ACCESS_TOKEN}")" || return 0
  list_json="$(echo "${response}" | jq -c '.data.list // []' 2>/dev/null || echo '[]')"
  if [[ "${list_json}" == "[]" ]]; then
    return 0
  fi
  while IFS= read -r item; do
    [[ -z "${item}" ]] && continue
    execute_task "${item}"
  done < <(echo "${list_json}" | jq -c '.[]')
}

main() {
  local last_heartbeat=0
  local last_status_report=0

  while true; do
    local now
    now="$(date +%s)"

    if (( now - last_heartbeat >= HEARTBEAT_INTERVAL )); then
      report_heartbeat
      last_heartbeat="${now}"
    fi

    if (( now - last_status_report >= STATUS_REPORT_INTERVAL )); then
      report_status
      last_status_report="${now}"
    fi

    poll_tasks
    sleep "${TASK_PULL_INTERVAL}"
  done
}

main "$@"
