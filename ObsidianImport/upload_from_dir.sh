#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
DEFAULT_CONFIG="${SCRIPT_DIR}/config.env"
CONFIG_FILE="${DEFAULT_CONFIG}"
RUN_ONCE=0

usage() {
  cat <<'EOF'
Usage:
  ./upload_from_dir.sh [config_file] [--once]

Options:
  --once    Run one scan-upload cycle and exit.
EOF
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  usage
  exit 0
fi

if [[ "${1:-}" == "--once" ]]; then
  RUN_ONCE=1
elif [[ -n "${1:-}" ]]; then
  CONFIG_FILE="${1}"
fi

if [[ "${2:-}" == "--once" ]]; then
  RUN_ONCE=1
fi

if [[ ! -f "${CONFIG_FILE}" ]]; then
  echo "Config file not found: ${CONFIG_FILE}"
  echo "Please copy ${SCRIPT_DIR}/config.env.example to ${SCRIPT_DIR}/config.env and edit it."
  exit 1
fi

# shellcheck source=/dev/null
source "${CONFIG_FILE}"

if [[ -z "${API_URL:-}" ]]; then
  echo "Missing API_URL in ${CONFIG_FILE}"
  exit 1
fi

# Backward compatibility for typo from earlier notes.
if [[ -z "${SCAN_DIR:-}" && -n "${SCDN_DIR:-}" ]]; then
  SCAN_DIR="${SCDN_DIR}"
fi

if [[ -z "${SCAN_DIR:-}" ]]; then
  echo "Missing SCAN_DIR in ${CONFIG_FILE}"
  exit 1
fi

if [[ ! -d "${SCAN_DIR}" ]]; then
  echo "SCAN_DIR does not exist: ${SCAN_DIR}"
  exit 1
fi

API_BASE="${API_URL%/}"
if [[ ! "${API_BASE}" =~ ^https?://[^/]+$ ]]; then
  echo "Invalid API_URL in ${CONFIG_FILE}"
  echo "API_URL must be domain/ip + port only, without API path."
  echo "Example: API_URL=\"http://127.0.0.1:10001\""
  exit 1
fi

IMPORT_ENDPOINT="${API_BASE}/api/v1/attachments/import"
FAILED_DIR="${SCAN_DIR%/}/failed"
MIN_INTERVAL_SEC=60
MAX_INTERVAL_SEC=$((32 * 60))
next_interval_sec=${MIN_INTERVAL_SEC}
cycle_no=0

mkdir -p "${FAILED_DIR}"

echo "Using import endpoint: ${IMPORT_ENDPOINT}"
echo "Scanning directory: ${SCAN_DIR}"
echo "Failed uploads folder: ${FAILED_DIR}"
echo

trim_spaces() {
  local text="${1:-}"
  text="${text#"${text%%[![:space:]]*}"}"
  text="${text%"${text##*[![:space:]]}"}"
  printf '%s' "${text}"
}

extract_singlefile_url() {
  local file="${1}"
  local ext="${file##*.}"
  ext="$(printf '%s' "${ext}" | tr '[:upper:]' '[:lower:]')"
  if [[ "${ext}" != "html" && "${ext}" != "htm" ]]; then
    return 1
  fi

  local extracted
  extracted="$(
    head -n 4 "${file}" | awk '
      BEGIN { IGNORECASE = 1 }
      /^[[:space:]]*url[[:space:]]*:/ {
        sub(/^[[:space:]]*url[[:space:]]*:[[:space:]]*/, "", $0)
        gsub(/\r/, "", $0)
        print
        exit
      }'
  )"
  extracted="$(trim_spaces "${extracted}")"
  if [[ -n "${extracted}" ]]; then
    printf '%s' "${extracted}"
    return 0
  fi
  return 1
}

move_to_failed() {
  local src="${1}"
  local base_scan="${SCAN_DIR%/}"
  local rel="${src#"${base_scan}/"}"
  local target="${FAILED_DIR}/${rel}"

  if [[ "${rel}" == "${src}" ]]; then
    target="${FAILED_DIR}/$(basename "${src}")"
  fi

  mkdir -p "$(dirname "${target}")"

  if [[ -e "${target}" ]]; then
    local ts base ext
    ts="$(date +%Y%m%d%H%M%S)"
    if [[ "${target}" == *.* ]]; then
      base="${target%.*}"
      ext=".${target##*.}"
    else
      base="${target}"
      ext=""
    fi
    target="${base}.failed-${ts}${ext}"
  fi

  mv "${src}" "${target}"
  printf '%s' "${target}"
}

process_once() {
  cycle_total=0
  cycle_success=0
  cycle_failed=0

  while IFS= read -r -d '' file; do
    cycle_total=$((cycle_total + 1))

    upload_args=(
      -sS
      -X POST "${IMPORT_ENDPOINT}"
      -F "file=@${file}"
    )

    detected_url=""
    if detected_url="$(extract_singlefile_url "${file}" || true)"; then
      if [[ -n "${detected_url}" ]]; then
        upload_args+=(-F "url=${detected_url}")
      fi
    fi

    response_with_code="$(curl "${upload_args[@]}" -w $'\n%{http_code}' || true)"
    http_code="$(printf '%s\n' "${response_with_code}" | tail -n 1)"
    body="$(printf '%s\n' "${response_with_code}" | sed '$d')"

    if [[ "${http_code}" == "200" || "${http_code}" == "201" ]]; then
      cycle_success=$((cycle_success + 1))
      if [[ -n "${detected_url}" ]]; then
        echo "[OK] ${file} (url=${detected_url})"
      else
        echo "[OK] ${file}"
      fi
      continue
    fi

    cycle_failed=$((cycle_failed + 1))
    failed_target="$(move_to_failed "${file}")"
    echo "[FAIL] ${file} (HTTP ${http_code}) -> moved to ${failed_target}"
    if [[ -n "${body}" ]]; then
      echo "       ${body}"
    fi
  done < <(
    find "${SCAN_DIR}" \
      \( -path "${FAILED_DIR}" -o -path "${FAILED_DIR}/*" \) -prune -o \
      -type f \( -iname '*.pdf' -o -iname '*.html' -o -iname '*.htm' \) -print0
  )
}

trap 'echo; echo "Stopped by user."; exit 0' INT TERM

while true; do
  cycle_no=$((cycle_no + 1))
  echo "===== Cycle ${cycle_no} @ $(date '+%Y-%m-%d %H:%M:%S') ====="

  process_once

  echo "Cycle result: Total=${cycle_total}, Success=${cycle_success}, Failed=${cycle_failed}"

  if (( cycle_success > 0 )); then
    next_interval_sec=${MIN_INTERVAL_SEC}
  else
    next_interval_sec=$((next_interval_sec * 2))
    if (( next_interval_sec > MAX_INTERVAL_SEC )); then
      next_interval_sec=${MAX_INTERVAL_SEC}
    fi
  fi

  if (( RUN_ONCE == 1 )); then
    echo "Run once mode enabled, exit now."
    break
  fi

  echo "Next scan in $((next_interval_sec / 60)) minute(s)."
  echo
  sleep "${next_interval_sec}"
done
