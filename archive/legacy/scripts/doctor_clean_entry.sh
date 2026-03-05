#!/usr/bin/env bash
set -euo pipefail

echo "==> Removing legacy entry layers"

# Remove old interactive/smartcmd if they exist
rm -rf internal/interactive 2>/dev/null || true
rm -rf internal/smartcmd 2>/dev/null || true

echo "==> Removing restless-gui cmd (temporary, will restore clean later)"
rm -rf cmd/restless-gui 2>/dev/null || true

echo "==> Cleaning old compiled leftovers"
find . -name "*.bak" -delete
find . -name "*~" -delete

echo "==> Formatting + tidy"
gofmt -w .
go mod tidy

echo "==> Testing"
go test ./...

echo ""
echo "If green:"
echo "git add -A && git commit -m 'cleanup: remove legacy entry layers for v2 core stabilization'"
