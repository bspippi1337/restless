
#!/usr/bin/env bash
set -euo pipefail

echo "🦢 Applying REAL BLCKSWAN patch"

rm -f internal/core/magiswarm/report.go 2>/dev/null || true

ROOT="internal/cli/root.go"
if [ ! -f "$ROOT" ]; then
  echo "ERROR: $ROOT not found"
  exit 1
fi

# Remove duplicates if present
sed -i '/NewBlckswanCmd()/d' "$ROOT"

# Insert once before first 'return cmd'
perl -0777 -i -pe 's/\n(\s*)return cmd/\n\1cmd.AddCommand(NewBlckswanCmd())\n\1return cmd/' "$ROOT"

gofmt -w internal/recon internal/topology internal/cli >/dev/null 2>&1 || true

mkdir -p build
go build -buildvcs=false -o build/restless ./cmd/restless

echo
echo "✅ built. Smoke:"
./build/restless blckswan https://api.github.com --max 40 --out dist || true
