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
    [--region ap-southeast-1]

Notes:
  - If --script-base-url is not provided, place install-falco-agent.sh and falco-agent.sh in the same directory.
  - This installer targets Amazon Linux 2023 x86_64 first.
EOF
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
    *)
      echo "unknown argument: $1" >&2
      usage
      exit 1
      ;;
  esac
done

if [[ -z "${SERVER_URL}" || -z "${ENROLL_KEY}" ]]; then
  echo "--server and --enroll-key are required" >&2
  usage
  exit 1
fi

if [[ -z "${AGENT_ID}" ]]; then
  AGENT_ID="agt-$(hostname)-$(date +%s)"
fi

ARCH="$(uname -m)"
if [[ "${ARCH}" != "x86_64" ]]; then
  echo "unsupported arch: ${ARCH}, only x86_64 is supported for now" >&2
  exit 1
fi

if [[ ! -f /etc/os-release ]]; then
  echo "missing /etc/os-release" >&2
  exit 1
fi

# shellcheck disable=SC1091
source /etc/os-release
if [[ "${ID:-}" != "amzn" || "${VERSION_ID:-}" != "2023" ]]; then
  echo "unsupported OS: ${PRETTY_NAME:-unknown}, expected Amazon Linux 2023" >&2
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
AGENT_SCRIPT_SOURCE="${SCRIPT_DIR}/falco-agent.sh"
if [[ -n "${SCRIPT_BASE_URL}" ]]; then
  AGENT_SCRIPT_SOURCE="/tmp/falco-agent.sh"
  curl -fsSL "${SCRIPT_BASE_URL%/}/runtime" -o "${AGENT_SCRIPT_SOURCE}"
  chmod +x "${AGENT_SCRIPT_SOURCE}"
elif [[ ! -f "${AGENT_SCRIPT_SOURCE}" ]]; then
  echo "missing agent runtime script: ${AGENT_SCRIPT_SOURCE}" >&2
  exit 1
fi

if ! command -v systemctl >/dev/null 2>&1; then
  echo "systemctl is required" >&2
  exit 1
fi

if ! command -v curl >/dev/null 2>&1 || ! command -v jq >/dev/null 2>&1; then
  dnf install -y curl jq
fi

INSTALL_DIR="/opt/falco-agent"
CONFIG_DIR="/etc/falco-agent"
STATE_DIR="/var/lib/falco-agent"
mkdir -p "${INSTALL_DIR}" "${CONFIG_DIR}" "${STATE_DIR}"
install -m 0755 "${AGENT_SCRIPT_SOURCE}" "${INSTALL_DIR}/falco-agent.sh"

PRIMARY_IP="$(hostname -I 2>/dev/null | awk '{print $1}')"
MACHINE_ID="$(cat /etc/machine-id 2>/dev/null || true)"
KERNEL_VERSION="$(uname -r)"
HOST_OS="${PRETTY_NAME:-Amazon Linux 2023}"

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
  --arg metadata "$(jq -nc --arg machineId "${MACHINE_ID}" --arg kernelVersion "${KERNEL_VERSION}" '{machineId:$machineId, kernelVersion:$kernelVersion}')" \
  '{enrollKey:$enrollKey, agentId:$agentId, hostname:$hostname, ip:$ip, instanceId:$instanceId, provider:$provider, region:$region, os:$os, arch:$arch, version:$version, labels:$labels, metadata:$metadata}')"

REGISTER_RESPONSE="$(curl -fsS -X POST "${SERVER_URL}/falco/agent/register" \
  -H "Content-Type: application/json" \
  -d "${REGISTER_PAYLOAD}")"

REGISTER_CODE="$(echo "${REGISTER_RESPONSE}" | jq -r '.code // 1')"
if [[ "${REGISTER_CODE}" != "0" ]]; then
  echo "agent register failed: ${REGISTER_RESPONSE}" >&2
  exit 1
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
Config file: ${CONFIG_DIR}/agent.env
Runtime script: ${INSTALL_DIR}/falco-agent.sh
Service name: falco-agent

Quick checks:
  sudo systemctl status falco-agent
  sudo journalctl -u falco-agent -f
EOF
