#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$(cd -- "${SCRIPT_DIR}/.." && pwd)"
RUNTIME_DIR="${APP_DIR}/.runtime"
BIN_DIR="${APP_DIR}/bin"
BIN_PATH="${BIN_DIR}/atthub-server"
PID_FILE="${RUNTIME_DIR}/atthub.pid"
LOG_FILE="${RUNTIME_DIR}/atthub.log"
ENV_FILE="${APP_DIR}/.env"

usage() {
	echo "Usage: $(basename "$0") {start|stop|restart|status|logs}"
}

ensure_dirs() {
	mkdir -p "${RUNTIME_DIR}" "${BIN_DIR}"
}

read_pid() {
	if [[ ! -f "${PID_FILE}" ]]; then
		return 1
	fi

	local pid
	pid="$(tr -d '[:space:]' <"${PID_FILE}")"
	if [[ ! "${pid}" =~ ^[0-9]+$ ]]; then
		return 1
	fi

	echo "${pid}"
}

is_pid_alive() {
	local pid="${1}"
	kill -0 "${pid}" 2>/dev/null
}

pid_matches_service() {
	local pid="${1}"
	local cmdline
	cmdline="$(ps -p "${pid}" -o command= 2>/dev/null || true)"
	[[ -n "${cmdline}" && "${cmdline}" == *"${BIN_PATH}"* ]]
}

is_running() {
	local pid
	if ! pid="$(read_pid)"; then
		return 1
	fi

	if is_pid_alive "${pid}" && pid_matches_service "${pid}"; then
		return 0
	fi

	rm -f "${PID_FILE}"
	return 1
}

load_env() {
	if [[ -f "${ENV_FILE}" ]]; then
		set -a
		# shellcheck source=/dev/null
		source "${ENV_FILE}"
		set +a
	fi
}

build_server() {
	(
		cd "${APP_DIR}"
		go build -o "${BIN_PATH}" ./cmd/server
	)
}

start_service() {
	ensure_dirs
	load_env

	if is_running; then
		local pid
		pid="$(read_pid)"
		echo "AttachmentHub already running (pid=${pid})"
		return 0
	fi

	build_server

	nohup "${BIN_PATH}" >>"${LOG_FILE}" 2>&1 &
	local pid=$!
	echo "${pid}" >"${PID_FILE}"

	sleep 0.3
	if ! is_pid_alive "${pid}" || ! pid_matches_service "${pid}"; then
		rm -f "${PID_FILE}"
		echo "AttachmentHub failed to start. Check logs: ${LOG_FILE}"
		tail -n 40 "${LOG_FILE}" || true
		return 1
	fi

	echo "AttachmentHub started (pid=${pid})"
	echo "Log file: ${LOG_FILE}"
}

stop_service() {
	if ! is_running; then
		echo "AttachmentHub is not running"
		return 0
	fi

	local pid
	pid="$(read_pid)"

	kill "${pid}" 2>/dev/null || true

	for _ in {1..20}; do
		if ! is_pid_alive "${pid}"; then
			rm -f "${PID_FILE}"
			echo "AttachmentHub stopped"
			return 0
		fi
		sleep 0.25
	done

	kill -9 "${pid}" 2>/dev/null || true
	rm -f "${PID_FILE}"
	echo "AttachmentHub force stopped"
}

status_service() {
	if is_running; then
		local pid
		pid="$(read_pid)"
		echo "AttachmentHub is running (pid=${pid})"
		echo "Log file: ${LOG_FILE}"
		return 0
	fi

	echo "AttachmentHub is not running"
	return 1
}

logs_service() {
	ensure_dirs
	touch "${LOG_FILE}"
	tail -f "${LOG_FILE}"
}

cmd="${1:-}"
case "${cmd}" in
start)
	start_service
	;;
stop)
	stop_service
	;;
restart)
	stop_service || true
	start_service
	;;
status)
	status_service
	;;
logs)
	logs_service
	;;
*)
	usage
	exit 1
	;;
esac
