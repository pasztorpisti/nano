#!/bin/bash
#
# Builds all server/test executables in a linux golang container if the
# executables haven't yet been built. You need only docker installation and
# access, no golang installation needed.
#

set -euo pipefail

cd "$( dirname "$0" )"
source env.sh

function main() {
	build_all_if_needed
}

function build_all_if_needed() {
	local NAME
	for NAME in "${ALL[@]}"; do
		if [[ ! -x "${BIN_DIR}/${NAME}" ]]; then
			# if at least one executable is missing then we rebuild
			build_all
			return
		fi
	done
	echo "The executables have already been built."
}

function build_all() {
	echo "Building all executables for linux ..."

	local TERM_PARAM=
	if [[ -t 1 ]]; then
		# making ctrl-c work if STDOUT is a terminal
		TERM_PARAM="-it"
	fi
	docker run --rm "${TERM_PARAM}" \
		-v "${NANO_DIR}:/go/src/${NANO_PKG}" \
		-w "/go/src/${SERVERS_PKG}/tests" \
		-e "HOST_UID=${UID}" \
			"${GOLANG_BUILD_IMAGE}" \
				./helpers/build_helper.sh do_it
}

main
