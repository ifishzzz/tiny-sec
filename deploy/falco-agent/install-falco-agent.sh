#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  sudo bash install-falco-agent.sh \
    --server http://<gva-host>:8888 \
    --enroll-key <enroll-key> \
    [--script-base-url http://<gva-host>:8888/falco/agent/install] \
    [--agent-id agt-node-01] \
    [--agent-version 0.1.0] \
    [--provider aws] \
    [--region ap-southeast-1] \
    [--falco-driver-choice auto|kmod|ebpf|modern_ebpf|none] \
    [--falcoctl-enabled yes|no]

Notes:
  - If --script-base-url is not provided, place install-falco-agent.sh and falco-agent.sh in the same directory.
  - The installer supports common Linux hosts with systemd and package managers such as apt, dnf, yum, and zypper.
  - Falco package installation itself is executed later by platform tasks.
EOF
}

log() {
  echo "[$(date '+%F %T')] $*"
}

fail() {
  echo "$*" >&2
  exit 1
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
      apt-get update -y
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
      fail "unsupported package manager: ${pkg_manager}"
      ;;
  esac
}

normalize_arch() {
  case "$(uname -m)" in
    x86_64|amd64)
      echo "x86_64"
      ;;
    aarch64|arm64)
      echo "aarch64"
      ;;
    *)
      return 1
      ;;
  esac
}

primary_ip() {
  local ip_value=""
  if command -v hostname >/dev/null 2>&1; then
    ip_value="$(hostname -I 2>/dev/null | awk '{print $1}')"
  fi
  if [[ -z "${ip_value}" ]] && command -v ip >/dev/null 2>&1; then
    ip_value="$(ip route get 1.1.1.1 2>/dev/null | awk '{for (i=1; i<=NF; i++) if ($i == "src") {print $(i+1); exit}}')"
  fi
  echo "${ip_value}"
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ "${EUID}" -ne 0 ]]; then
  exec sudo bash "$0" "$@"
fi

SERVER_URL=""
ENROLL_KEY=""
SCRIPT_BASE_URL=""
AGENT_ID=""
AGENT_VERSION="0.1.0"
PROVIDER="aws"
REGION=""
FALCO_DRIVER_CHOICE=""
FALCOCTL_ENABLED="no"

while [[ $# -gt 0 ]]; do
  case "$1" in
    --server)
      SERVER_URL="${2:-}"
      shift 2
      ;;
    --enroll-key)
      ENROLL_KEY="${2:-}"
      shift 2
      ;;
    --script-base-url)
      SCRIPT_BASE_URL="${2:-}"
      shift 2
      ;;
    --agent-id)
      AGENT_ID="${2:-}"
      shift 2
      ;;
    --agent-version)
      AGENT_VERSION="${2:-}"
      shift 2
      ;;
    --provider)
      PROVIDER="${2:-}"
      shift 2
      ;;
    --region)
      REGION="${2:-}"
      shift 2
      ;;
    --falco-driver-choice)
      FALCO_DRIVER_CHOICE="${2:-}"
      shift 2
      ;;
    --falcoctl-enabled)
      FALCOCTL_ENABLED="${2:-}"
      shift 2
      ;;
    *)
      echo "unknown argument: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if [[ -z "${SERVER_URL}" || -z "${ENROLL_KEY}" ]]; then
  fail "--server and --enroll-key are required"
fi

if [[ -z "${AGENT_ID}" ]]; then
  AGENT_ID="agt-$(hostname)-$(date +%s)"
fi

ARCH="$(normalize_arch)" || fail "unsupported arch: $(uname -m), only x86_64 and aarch64 are supported"

if [[ ! -f /etc/os-release ]]; then
  fail "missing /etc/os-release"
fi

# shellcheck disable=SC1091
source /etc/os-release
HOST_OS="${PRETTY_NAME:-${ID:-linux}}"
PACKAGE_MANAGER="$(detect_package_manager)" || fail "unsupported host: no supported package manager found (apt/dnf/yum/zypper)"

if ! command -v systemctl >/dev/null 2>&1; then
  fail "systemctl is required"
fi

missing_packages=()
command -v curl >/dev/null 2>&1 || missing_packages+=("curl")
command -v jq >/dev/null 2>&1 || missing_packages+=("jq")

if [[ "${#missing_packages[@]}" -gt 0 ]]; then
  log "installing missing dependencies via ${PACKAGE_MANAGER}: ${missing_packages[*]}"
  install_packages "${PACKAGE_MANAGER}" "${missing_packages[@]}"
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AGENT_SCRIPT_SOURCE="${SCRIPT_DIR}/falco-agent.sh"
if [[ -n "${SCRIPT_BASE_URL}" ]]; then
  AGENT_SCRIPT_SOURCE="/tmp/falco-agent.sh"
  curl -fsSL "${SCRIPT_BASE_URL%/}/runtime" -o "${AGENT_SCRIPT_SOURCE}"
  chmod +x "${AGENT_SCRIPT_SOURCE}"
elif [[ ! -f "${AGENT_SCRIPT_SOURCE}" ]]; then
  fail "missing agent runtime script: ${AGENT_SCRIPT_SOURCE}"
fi

INSTALL_DIR="/opt/falco-agent"
CONFIG_DIR="/etc/falco-agent"
STATE_DIR="/var/lib/falco-agent"
mkdir -p "${INSTALL_DIR}" "${CONFIG_DIR}" "${STATE_DIR}"
install -m 0755 "${AGENT_SCRIPT_SOURCE}" "${INSTALL_DIR}/falco-agent.sh"

PRIMARY_IP="$(primary_ip)"
MACHINE_ID="$(cat /etc/machine-id 2>/dev/null || true)"
KERNEL_VERSION="$(uname -r)"

REGISTER_PAYLOAD="$(jq -nc \
  --arg enrollKey "${ENROLL_KEY}" \
  --arg agentId "${AGENT_ID}" \
  --arg hostname "$(hostname)" \
  --arg ip "${PRIMARY_IP}" \
  --arg instanceId "" \
  --arg provider "${PROVIDER}" \
  --arg region "${REGION}" \
  --arg os "${HOST_OS}" \
  --arg arch "${ARCH}" \
  --arg version "${AGENT_VERSION}" \
  --arg labels "{}" \
  --arg metadata "$(jq -nc --arg machineId "${MACHINE_ID}" --arg kernelVersion "${KERNEL_VERSION}" --arg packageManager "${PACKAGE_MANAGER}" '{machineId:$machineId, kernelVersion:$kernelVersion, packageManager:$packageManager}')" \
  '{enrollKey:$enrollKey, agentId:$agentId, hostname:$hostname, ip:$ip, instanceId:$instanceId, provider:$provider, region:$region, os:$os, arch:$arch, version:$version, labels:$labels, metadata:$metadata}')"

REGISTER_RESPONSE="$(curl -fsS -X POST "${SERVER_URL}/falco/agent/register" \
  -H "Content-Type: application/json" \
  -d "${REGISTER_PAYLOAD}")"

REGISTER_CODE="$(echo "${REGISTER_RESPONSE}" | jq -r '.code // 1')"
if [[ "${REGISTER_CODE}" != "0" ]]; then
  fail "agent register failed: ${REGISTER_RESPONSE}"
fi

ACCESS_TOKEN="$(echo "${REGISTER_RESPONSE}" | jq -r '.data.accessToken // ""')"
HEARTBEAT_INTERVAL="$(echo "${REGISTER_RESPONSE}" | jq -r '.data.heartbeatInterval // 30')"
TASK_PULL_INTERVAL="$(echo "${REGISTER_RESPONSE}" | jq -r '.data.taskPullInterval // 15')"
EVENT_UPLOAD_BATCH_SIZE="$(echo "${REGISTER_RESPONSE}" | jq -r '.data.eventUploadBatchSize // 200')"

cat > "${CONFIG_DIR}/agent.env" <<EOF
SERVER_URL=${SERVER_URL}
ENROLL_KEY=${ENROLL_KEY}
AGENT_ID=${AGENT_ID}
ACCESS_TOKEN=${ACCESS_TOKEN}
HEARTBEAT_INTERVAL=${HEARTBEAT_INTERVAL}
TASK_PULL_INTERVAL=${TASK_PULL_INTERVAL}
EVENT_UPLOAD_BATCH_SIZE=${EVENT_UPLOAD_BATCH_SIZE}
FALCO_FRONTEND=noninteractive
FALCO_DRIVER_CHOICE=${FALCO_DRIVER_CHOICE}
FALCOCTL_ENABLED=${FALCOCTL_ENABLED}
EOF

cat > /etc/systemd/system/falco-agent.service <<'EOF'
[Unit]
Description=Falco Agent
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=/opt/falco-agent/falco-agent.sh
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload
systemctl enable --now falco-agent
systemctl --no-pager --full status falco-agent || true

cat <<EOF
Falco agent installed successfully.
Host OS: ${HOST_OS}
Package manager: ${PACKAGE_MANAGER}
Config file: ${CONFIG_DIR}/agent.env
Runtime script: ${INSTALL_DIR}/falco-agent.sh
Service name: falco-agent

Quick checks:
  sudo systemctl status falco-agent
  sudo journalctl -u falco-agent -f
EOF
