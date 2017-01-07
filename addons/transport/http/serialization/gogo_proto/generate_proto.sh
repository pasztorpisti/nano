#!/bin/bash
set -euo pipefail
cd "$( dirname "$0" )"

PKG="github.com/pasztorpisti/nano/addons/transport/http/serialization/gogo_proto"
protoc --gogofaster_out="M${PKG}/Transport.proto=${PKG}:." Transport.proto
