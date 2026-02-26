#!/usr/bin/env bash
set -euo pipefail

# Records a clean asciinema demo for Restless.
# Requires: asciinema
# Usage: ./scripts/record_demo.sh

OUT="${1:-demo.cast}"

echo "Recording to: ${OUT}"
echo "Tip: run these commands during recording:"
echo "  restless probe https://httpbin.org"
echo "  restless list"
echo "  restless run GET /headers"
echo "  restless session"
echo

asciinema rec "${OUT}"
