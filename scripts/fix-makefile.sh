#!/usr/bin/env bash
set -euo pipefail

FILE="Makefile"

if [[ ! -f "$FILE" ]]; then
  echo "❌ No Makefile found in current directory"
  exit 1
fi

echo "🔍 Scanning Makefile..."

# Backup
cp "$FILE" "${FILE}.bak.$(date +%s)"
echo "📦 Backup created"

# Replace go build lines that do NOT already contain CGO_ENABLED
# Adds CGO_ENABLED=0 before GOOS or go build
sed -i '
/go build/ {
  /CGO_ENABLED=/! {
    s/GOOS=/CGO_ENABLED=0 GOOS=/g
    s/\tgo build/\tCGO_ENABLED=0 go build/g
  }
}
' "$FILE"

echo "🧹 Cleaning trailing double spaces..."
sed -i 's/  */ /g' "$FILE"

echo "🔧 Ensuring build-all target exists..."

if ! grep -q "^build-all:" "$FILE"; then
  cat >> "$FILE" <<'EOF'

build-all:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -o dist/restless_linux_amd64   ./cmd/restless
	CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -o dist/restless_linux_arm64   ./cmd/restless
	CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -o dist/restless_darwin_amd64  ./cmd/restless
	CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -o dist/restless_darwin_arm64  ./cmd/restless
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o dist/restless_windows_amd64.exe ./cmd/restless
EOF
  echo "➕ Added build-all target"
fi

echo "🧪 Testing Makefile syntax..."
make -n build-all >/dev/null 2>&1 || {
  echo "⚠️ Warning: Makefile may need manual adjustment"
}

echo "✅ Makefile patched for CGO-free cross-compilation"
echo "💡 You can now run: make build-all"
