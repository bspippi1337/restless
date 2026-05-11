#!/data/data/com.termux/files/usr/bin/bash

set -euo pipefail

MSG="${1:-checkpoint}"

if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "[!] Ikke et git repo"
  exit 1
fi

git add -A

if git diff --cached --quiet; then
  echo "[*] Ingen endringer å committe"
  exit 0
fi

STAMP=$(date +"%Y-%m-%d %H:%M:%S")

git commit -m "restless: ${MSG} (${STAMP})"

echo
echo "[+] Checkpoint committed"
