#!/usr/bin/env sh
set -eu
cd docs
python3 -m http.server 8080
