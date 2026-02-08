#!/usr/bin/env bash
set -euo pipefail
python3 -m http.server 8765 --directory assets/brand/logo
