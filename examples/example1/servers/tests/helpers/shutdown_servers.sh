#!/bin/bash
#
# - removes all containers that contain "${CONTAINER_NAME_PREFIX}" in their names
# - removes the "nano" docker network
#

set -euo pipefail

cd "$( dirname "$0" )"
source env.sh

function main() {
	rm_containers
	rm_network
}

function rm_containers() {
	local CONTAINERS="$( docker ps -aqf name="${CONTAINER_NAME_PREFIX}" )"
	if [[ -n "${CONTAINERS[@]}" ]]; then
		docker rm -f ${CONTAINERS[@]} >/dev/null
	fi
}

function network_exists() {
	docker network inspect "${NETWORK}" >&/dev/null
}

function rm_network() {
	if network_exists; then
		docker network rm "${NETWORK}" >/dev/null
	fi
}

main
