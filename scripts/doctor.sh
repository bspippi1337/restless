#!/usr/bin/env sh
set -eu
echo "[doctor] gofmt"
gofmt -w . >/dev/null 2>&1 || true
echo "[doctor] mod tidy"
go mod tidy >/dev/null 2>&1 || true
echo "[doctor] ok"
