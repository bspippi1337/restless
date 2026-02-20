#!/usr/bin/env sh
set -eu
cd "$(dirname "$0")/.."
cd docs
if command -v python3 >/dev/null 2>&1; then
  python3 -m http.server 8080
elif command -v python >/dev/null 2>&1; then
  python -m http.server 8080
else
  echo "Need python3 (or python) to serve docs locally."
  exit 1
fi
