#!/bin/bash
#
# This script will execute inside a golang builder container.
#

set -euo pipefail

cd "$( dirname "$0" )"
source env.sh

# Since things inside the container execute as root we don't want to place
# the root-owned files directly to the mounted volumes. First we compile the
# root-owned executables into a temp dir and then we change their owner to
# ${HOST_UID} and then move them to the mounted volume.
TEMP_BIN="/temp_bin"

function main() {
	if [[ $# -ne 1 || $1 != "do_it" ]]; then
		>&2 echo "Don't run this manually."
		exit 1
	fi

	mkdir -p "${BIN_DIR}"

	local NAME
	for NAME in "${ALL[@]}"; do
		echo "building ${NAME} ..."
		build "${NAME}"
	done

	# The HOST_UID env is passed into the container by a docker command.
	chown -R "${HOST_UID}:${HOST_UID}" "${TEMP_BIN}"
	rm -rf "${BIN_DIR}"
	cp -rp "${TEMP_BIN}" "${BIN_DIR}"
}

function build() {
	local NAME="$1"
	local PKG="${SERVERS_PKG}/${NAME}"
	go get -d -v "${PKG}"
	go build \
		-o "${TEMP_BIN}/${NAME}" \
		-ldflags "-linkmode external -extldflags -static" \
			"${PKG}"
}

main "$@"
