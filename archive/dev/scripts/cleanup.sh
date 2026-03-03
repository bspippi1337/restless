#!/usr/bin/env sh
set -eu
echo "==> cleanup: removing bin/ dist/ build/ logs/ and *.log"
rm -rf bin dist build logs .fixall-logs 2>/dev/null || true
find . -maxdepth 4 -type f -name "*.log" -delete 2>/dev/null || true
echo "[ OK ] cleanup complete"
