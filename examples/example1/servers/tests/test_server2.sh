#!/bin/bash
set -euo pipefail
cd "$( dirname "$0" )"
helpers/test.sh server2a:svc2 server2b:svc3
