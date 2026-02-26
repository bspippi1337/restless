#!/usr/bin/env bash
set -euo pipefail

REPO="${1:-https://github.com/bspippi1337/restless.git}"
WORKDIR="${2:-/tmp/restless_clean_$(date +%s)}"

echo "======================================"
echo "        RESTLESS SAVE-ME BUILD"
echo "======================================"
echo
echo "Repo:    $REPO"
echo "Workdir: $WORKDIR"
echo

echo "==> Step 1: Clean environment"
unset GOWORK || true
unset GOFLAGS || true
unset GOPATH || true
unset GOMODCACHE || true
unset SSL_CERT_FILE || true
unset SSL_CERT_DIR || true

echo "==> Step 2: Fresh clone"
rm -rf "$WORKDIR"
git clone "$REPO" "$WORKDIR"

cd "$WORKDIR"

echo "==> Step 3: Remove go.work if present"
rm -f go.work || true

echo "==> Step 4: Force clean module state"
go clean -cache -modcache -work

echo "==> Step 5: Show Go env"
go env | grep -E 'GOMOD|GOWORK|GOPATH|GOROOT'

echo
echo "==> Step 6: Tidy"
go mod tidy

echo
echo "==> Step 7: Direct build (no Makefile)"
go build -x -o restless ./cmd/restless

echo
echo "==> Step 8: If Makefile exists, test it"
if [ -f Makefile ]; then
    make clean || true
    make build || true
fi

echo
echo "======================================"
echo "        SAVE-ME COMPLETE"
echo "Binary (if success):"
echo "  $WORKDIR/restless"
echo "======================================"
