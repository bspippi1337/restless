#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")/.."
cd corec
make clean || true
make
echo "OK: built corec/bin/restless-core"
