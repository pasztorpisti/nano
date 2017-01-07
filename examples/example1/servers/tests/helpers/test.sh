#!/bin/bash
set -euo pipefail

cd "$( dirname "$0" )"
source env.sh

HELP_TEXT="
Usage: $0 mapping [mapping [mapping [...]]]

A mapping has the following format:
server_name[:network_alias[:network_alias[:...]]]
"

function main() {
	if [[ $# -eq 0 ]]; then
		>&2 echo "${HELP_TEXT}"
		exit 1
	fi

	trap ./shutdown_servers.sh EXIT
	./start_servers.sh "$@"
	run_test
	echo "SUCCESS."
}

function run_test() {
	echo "Running test against svc2 ..."
	local TERM_PARAM=
	if [[ -t 1 ]]; then
		# making ctrl-c work if STDOUT is a terminal
		TERM_PARAM="-it"
	fi

	# This is a very primitive test: we simply check whether two types of
	# requests succeed without checking the actual response values.
	# A real world test should perform a lot of tests and should check the
	# response values.
	docker run --rm "${TERM_PARAM}" \
		-v "${BIN_DIR}:/servers" \
		--net "${NETWORK}" \
		-w /servers \
			"${IMAGE_NAME}" \
				/bin/sh -c "/servers/test_client -addr svc2:8000 && /servers/test_client -addr svc2:8000 -req getreq"
}

main "$@"
