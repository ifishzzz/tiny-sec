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
mkdir -p "${STATE_DIR}"

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

falco_service_state() {
  if command -v systemctl >/dev/null 2>&1; then
    systemctl is-active falco 2>/dev/null || echo "inactive"
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

run_shell_command() {
  local command_text="${1}"
  local stdout_file="${2}"
  local stderr_file="${3}"
  bash -lc "${command_text}" >"${stdout_file}" 2>"${stderr_file}"
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

  local command_text=""
  case "${action}" in
    falco.install)
      command_text="mkdir -p \"$(dirname "${config_path}")\" && dnf makecache -y && (dnf install -y \"falco-${falco_version_value}\" || dnf install -y falco) && systemctl enable ${service_name} && systemctl restart ${service_name}"
      ;;
    falco.upgrade)
      command_text="dnf makecache -y && (dnf upgrade -y falco || dnf install -y falco) && systemctl restart ${service_name}"
      ;;
    falco.rollback)
      command_text="dnf downgrade -y falco && systemctl restart ${service_name}"
      ;;
    falco.reload)
      command_text="systemctl reload ${service_name} || systemctl restart ${service_name}"
      ;;
    falco.restart)
      command_text="systemctl restart ${service_name}"
      ;;
    falco.rule_publish)
      command_text="systemctl reload ${service_name} || systemctl restart ${service_name}"
      ;;
    *)
      echo "unsupported action: ${action}" >"${stderr_file}"
      error_code="UNSUPPORTED_ACTION"
      error_message="unsupported action: ${action}"
      save_task_result "${task_id}" "failed" "rejected" "" "$(trim_summary "${stderr_file}")" "${error_code}" "${error_message}" "${rule_package_version}"
      rm -f "${stdout_file}" "${stderr_file}"
      return 0
      ;;
  esac

  if run_shell_command "${command_text}" "${stdout_file}" "${stderr_file}"; then
    save_task_result "${task_id}" "succeeded" "finished" "$(trim_summary "${stdout_file}")" "$(trim_summary "${stderr_file}")" "" "" "${rule_package_version}"
  else
    error_code="TASK_EXEC_FAILED"
    error_message="command execution failed"
    save_task_result "${task_id}" "failed" "finished" "$(trim_summary "${stdout_file}")" "$(trim_summary "${stderr_file}")" "${error_code}" "${error_message}" "${rule_package_version}"
  fi

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
