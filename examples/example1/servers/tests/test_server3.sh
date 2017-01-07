#!/bin/bash
set -euo pipefail
cd "$( dirname "$0" )"
helpers/test.sh server3a:svc1 server3b:svc2 server3c:svc3 server3d:svc4
