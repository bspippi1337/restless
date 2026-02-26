#!/usr/bin/env bash
set -e

echo "======================================="
echo "        RESTLESS DOCTOR MODE"
echo "======================================="
echo

echo "==> Working directory:"
pwd
echo

echo "==> Git branch:"
git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "Not a git repo"
echo

echo "==> Git status:"
git status --short || true
echo

echo "==> Files in cmd/restless:"
ls -la cmd/restless || true
echo

echo "==> Packages in cmd/restless:"
grep -R "^package " cmd/restless || true
echo

echo "==> func main occurrences:"
grep -R "func main" cmd/restless || true
echo

echo "==> go.mod module path:"
grep "^module" go.mod || true
echo

echo "==> go version:"
go version
echo

echo "==> go mod tidy (dry check):"
go mod tidy -v || true
echo

echo "==> Attempting direct build:"
go build -v ./cmd/restless 2>&1 || true
echo

echo "==> Makefile build attempt:"
make 2>&1 || true
echo

echo "======================================="
echo "           DOCTOR COMPLETE"
echo "======================================="
