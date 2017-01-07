#!/bin/bash
set -euo pipefail
cd "$( dirname "$0" )"
helpers/test.sh server1:svc2
