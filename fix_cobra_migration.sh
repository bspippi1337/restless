#!/usr/bin/env bash
set -euo pipefail

echo "======================================="
echo "   RESTLESS COBRA MIGRATION FIXER"
echo "======================================="
echo

ROOT_DIR=$(pwd)
CMD_DIR="cmd/restless"
BACKUP_DIR="legacy_backup_$(date +%Y%m%d_%H%M%S)"

if [ ! -d "$CMD_DIR" ]; then
  echo "ERROR: $CMD_DIR not found."
  exit 1
fi

echo "==> Scanning for package conflicts..."

PACKAGES=$(grep -R "^package " "$CMD_DIR" | awk '{print $2}' | sort | uniq)

if [ "$(echo "$PACKAGES" | wc -l)" -gt 1 ]; then
  echo "Multiple packages detected:"
  echo "$PACKAGES"
  echo
  echo "==> Creating backup directory: $BACKUP_DIR"
  mkdir -p "$BACKUP_DIR"

  echo "==> Moving non-main packages out of $CMD_DIR..."
  for f in $(grep -R "^package " "$CMD_DIR" | grep -v "package main" | cut -d: -f1); do
    echo "  -> Moving $f"
    mv "$f" "$BACKUP_DIR/"
  done
else
  echo "No package conflicts detected."
fi

echo
echo "==> Checking for multiple main() functions..."

MAIN_COUNT=$(grep -R "func main" "$CMD_DIR" | wc -l)

if [ "$MAIN_COUNT" -gt 1 ]; then
  echo "WARNING: Multiple main() functions found:"
  grep -R "func main" "$CMD_DIR"
  echo
  echo "You may need to manually remove extra main() definitions."
else
  echo "Single main() detected. OK."
fi

echo
echo "==> Cleaning build cache..."
go clean -cache

echo
echo "==> Attempting build..."
if go build ./cmd/restless; then
  echo
  echo "======================================="
  echo " BUILD SUCCESSFUL"
  echo "======================================="
else
  echo
  echo "======================================="
  echo " BUILD FAILED"
  echo "======================================="
  exit 1
fi
