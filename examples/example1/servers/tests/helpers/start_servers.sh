#!/bin/bash
#
# - builds all server executables if necessary
# - creates a "nano" docker network
# - starts the specified servers in their own containers with the
#   specified hostnames for the containers
# - waits for each server to become ready (until their /health-check becomes responsive)
#
# Note: In a real life scenario you can use docker-compose to start up and tear
# down several containers and setup network between them.
# In these simple test scripts I decided to depend only on docker-engine.
#

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

	# Removing any leftover containers from a previously failed execution.
	./shutdown_servers.sh
	./build_all.sh

	# Creating the "nano" docker network for the server containers.
	# The containers will be able to find each other using the embedded docker DNS.
	# A service name can be used as a hostname to resolve the IP address of the
	# server that contains the service.
	create_network

	docker pull "${IMAGE_NAME}"

	local MAPPING
	for MAPPING in "$@"; do
		start_server_container "${MAPPING}"
	done

	for MAPPING in "$@"; do
		wait_for_server_ready "${MAPPING}"
	done
}

function start_server_container() {
	local ORIG_IFS="${IFS}"
	local IFS=":"
	local MAPPING=( $1 )
	IFS="${ORIG_IFS}"

	local NAME="${MAPPING[0]}"
	local CNAME="${CONTAINER_NAME_PREFIX}${NAME}"

	local DOCKER_PARAMS=(
		run -dit
		--net "${NETWORK}"
		-h "${NAME}"
		--name "${CNAME}"
		-v "${BIN_DIR}":/servers
	)
	if [[ -t 1 ]]; then
		# making ctrl-c work if STDOUT is a terminal
		DOCKER_PARAMS+=( -it )
	fi

	local ALIAS
	for ALIAS in "${MAPPING[@]:1}"; do
		DOCKER_PARAMS+=( --network-alias "${ALIAS}" )
	done

	DOCKER_PARAMS+=(
		"${IMAGE_NAME}"
		"/servers/${NAME}"
	)

	echo "Starting ${NAME} ..."
	docker "${DOCKER_PARAMS[@]}" >/dev/null
}

function wait_for_server_ready() {
	local ORIG_IFS="${IFS}"
	local IFS=":"
	local MAPPING=( $1 )
	IFS="${ORIG_IFS}"

	local NAME="${MAPPING[0]}"
	local CNAME="${CONTAINER_NAME_PREFIX}${NAME}"

	echo -n "Waiting for ${NAME} to become ready "

	# Waiting until the /health-check endpoint of the server becomes responsive.
	local CMD="cd /root && wget -q http://localhost:8000/health-check >&/dev/null"
	local ATTEMPT
	for ATTEMPT in $( seq 5 ); do
		if docker exec "${CNAME}" /bin/sh -c "${CMD}"; then
			echo "OK"
			break
		fi
		if [[ "${ATTEMPT}" -eq 5 ]]; then
			echo "FAILED"
			return 1
		fi
		sleep 1
		echo -n "."
	done
}

function network_exists() {
	docker network inspect "${NETWORK}" >&/dev/null
}

function create_network() {
	if ! network_exists; then
		echo "Creating docker network ${NETWORK} ..."
		docker network create "${NETWORK}" >/dev/null
	fi
}

main "$@"
