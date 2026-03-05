#!/usr/bin/env bash
set -euo pipefail

# Quickstart (v2)
# Build:
#   go build -o restless-v2 ./cmd/restless-v2
# Run:
#   ./restless-v2 -url https://httpbin.org/get
# Template vars:
#   ./restless-v2 -set token=abc123 -Hk Authorization -Hv 'Bearer {{token}}' -url https://httpbin.org/headers
# Bench:
#   ./restless-v2 -bench -c 20 -dur 3s -url https://httpbin.org/get

echo "OK - see comments in this file."
