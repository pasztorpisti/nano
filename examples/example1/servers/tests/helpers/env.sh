#!/bin/bash

IFS=$'\n'

THIS_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TESTS_DIR="$( cd "${THIS_DIR}/.." && pwd )"
BIN_DIR="${TESTS_DIR}/bin"

NANO_PKG="github.com/pasztorpisti/nano"
NANO_DIR="$( cd "${TESTS_DIR}/../../../.." && pwd )"

SERVERS_DIR="$( cd "${TESTS_DIR}/.." && pwd )"
SERVERS_PKG="${NANO_PKG}/examples/example1/servers"

# The name of the docker network to create. We create a separate docker
# network because the embedded docker DNS works only with custom networks.
# The default docker bridge works in legacy mode without DNS.
NETWORK="nano"

# Docker image to use to run the static linked server executables.
IMAGE_NAME="alpine:3.4"

# Docker image to use to build the server executables.
GOLANG_BUILD_IMAGE="golang:1.7.4"

# Every docker container we create will have this prefix. This makes mass
# deletion of our containers easier by filter for this.
CONTAINER_NAME_PREFIX="nano_test_"

# List of package names to build under the nano/examples/example1/servers directory.
ALL=(
	server1
	server2a
	server2b
	server3a
	server3b
	server3c
	server3d
	test_client
)
