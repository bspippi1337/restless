#!/usr/bin/env bash
set -euo pipefail

ROOT="/home/pippi/dev/restless"
BACKUP_DIR=$(ls -d "$ROOT"/legacy_backup_* 2>/dev/null | sort | tail -n 1 || true)

echo "========================================="
echo "   RESTLESS AUTO-RESTORE ENTRYPOINT"
echo "========================================="

cd "$ROOT"

if [ -z "$BACKUP_DIR" ]; then
    echo "❌ No legacy_backup_* directory found."
    exit 1
fi

echo "Using backup dir: $BACKUP_DIR"

SOURCE_FILE="$BACKUP_DIR/main_legacy.go"
TARGET_DIR="$ROOT/cmd/restless"
TARGET_FILE="$TARGET_DIR/main.go"

if [ ! -f "$SOURCE_FILE" ]; then
    echo "❌ $SOURCE_FILE not found."
    exit 1
fi

mkdir -p "$TARGET_DIR"

echo "Restoring entrypoint..."
cp "$SOURCE_FILE" "$TARGET_FILE"

echo "Cleaning Go workspace..."
rm -f go.work || true
unset GOWORK || true
go clean -cache -modcache -work || true

echo "Building..."
if go build -o bin/restless ./cmd/restless; then
    echo
    echo "✅ SUCCESS: entrypoint restored and binary built."
else
    echo
    echo "❌ Build failed. But entrypoint restored."
    exit 1
fi

echo "========================================="
echo "Done."
echo "Binary: bin/restless"
echo "========================================="
