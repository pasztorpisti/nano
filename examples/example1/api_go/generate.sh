#!/bin/bash
#
# This scripts generates api_go from the the language independent ../api dir.
# Requires protoc and go installations and the GOPATH env var to be set.
#

set -euo pipefail
IFS=$'\n'

cd "$( dirname "$0" )"
THIS_DIR="$( pwd )"

INPUT_ROOT_PKG="github.com/pasztorpisti/nano/examples/example1/api"
OUTPUT_ROOT_PKG="github.com/pasztorpisti/nano/examples/example1/api_go"

INPUT_ROOT_DIR="${GOPATH}/src/${INPUT_ROOT_PKG}"
OUTPUT_ROOT_DIR="${GOPATH}/src/${OUTPUT_ROOT_PKG}"

function main() {
	initialise

	if [[ $# -ge 1 ]]; then
		local SERVICES=( "$@" )
	else
		local SERVICES=()
		local DIR
		for DIR in $( find "${INPUT_ROOT_DIR}" -type d -depth 1 ); do
			SERVICES+=( "$( basename "${DIR}" )" )
		done
	fi

	local SERVICE
	for SERVICE in "${SERVICES[@]}"; do
		generate_service "${SERVICE}"
	done
}

function initialise() {
	if ! command_exists protoc; then
		>&2 echo "Please install the protoc command and put it onto your PATH."
		exit 1
	fi
	if ! command_exists go; then
		>&2 echo "Please install go and put it onto your PATH."
		exit 1
	fi

	# With bash 4.2 and newer we could use -v to check if the env var is set
	# but some distros have quite old bash and OSX has a 10 years old version!
	if [[ -z ${GOPATH+x} ]]; then
		>&2 echo "Please set the GOPATH env var."
		exit 1
	fi
	export PATH="${PATH}:${GOPATH}/bin"

	if ! command_exists gen_http_transport_config; then
		go install "github.com/pasztorpisti/nano/addons/transport/http/config/gen_http_transport_config"
	fi
	if ! command_exists protoc-gen-gogofaster; then
		go install "github.com/gogo/protobuf/protoc-gen-gogofaster"
	fi
}

function command_exists() {
	type "$1" >&/dev/null
}

function generate_service() {
	local SVC="$1"

	local INPUT_PKG="${INPUT_ROOT_PKG}/${SVC}"
	local OUTPUT_PKG="${OUTPUT_ROOT_PKG}/${SVC}"

	local INPUT_DIR="${INPUT_ROOT_DIR}/${SVC}"
	local OUTPUT_DIR="${OUTPUT_ROOT_DIR}/${SVC}"

	if [[ ! -d "${INPUT_DIR}" ]]; then
		>&2 echo "Input dir doesn't exist: ${INPUT_DIR}"
		return 1
	fi

	mkdir -p "${OUTPUT_DIR}"

	if [[ -r "${INPUT_DIR}/requests.proto" ]]; then
		echo "Processing ${INPUT_DIR}/requests.proto ..."
		local IMPORT_MAP=(
			"M${INPUT_PKG}/requests.proto=${OUTPUT_PKG}"
		)
		protoc \
			--proto_path="${INPUT_DIR}" \
			--gogofaster_out="$( IFS="," && echo "${IMPORT_MAP[*]}" ):${OUTPUT_DIR}" \
				"${INPUT_DIR}/requests.proto"
	fi

	if [[ -r "${INPUT_DIR}/http_transport_config.json" ]]; then
		echo "Processing ${INPUT_DIR}/http_transport_config.json ..."
		gen_http_transport_config \
			"${INPUT_DIR}/http_transport_config.json:${OUTPUT_DIR}/http_transport_config.go"
		gofmt -w "${OUTPUT_DIR}/http_transport_config.go"
	fi
}

main "$@"
