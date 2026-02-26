#!/usr/bin/env bash
set -euo pipefail

MAIN="cmd/restless/main.go"
VALIDATE="internal/validate/validate.go"

[ -f "$MAIN" ] || { echo "main.go not found"; exit 1; }

echo "==> Patching request mode..."

# Add timeout flag to request mode
perl -0777 -i -pe '
s/(body := fs.String\("d", "", "Body string"\))/\1\n\ttimeout := fs.Int("timeout", 7, "Timeout in seconds")/s
unless /timeout := fs\.Int\("timeout"/;
' "$MAIN"

# Replace RunOnce context
perl -0777 -i -pe '
s/resp, err := a\.RunOnce\(context\.Background\(\), req\)/ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout)*time.Second)\n\tdefer cancel()\n\n\tresp, err := a.RunOnce(ctx, req)/s
unless /context\.WithTimeout/;
' "$MAIN"

echo "==> Patching validate handler..."

perl -0777 -i -pe '
s/(jsonOut := fs.String\("json", false, "JSON output"\))/\1\n\ttimeout := fs.Int("timeout", 7, "Timeout in seconds")/s
unless /timeout := fs\.Int\("timeout"/;
' "$MAIN"

perl -0777 -i -pe '
s/ctx := context\.Background\(\)/ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout)*time.Second)\n\tdefer cancel()/s
unless /defer cancel\(\)/;
' "$MAIN"

echo "==> Ensuring time import..."

if ! grep -q '"time"' "$MAIN"; then
  perl -0777 -i -pe 's/import \(/import (\n\t"time"/;' "$MAIN"
fi

echo "==> Ensuring validate uses timeout..."

if [ -f "$VALIDATE" ]; then
  perl -0777 -i -pe '
  s/client := &http\.Client\s*\{\s*\}/client := &http.Client{ Timeout: opt.Timeout }/g
  ' "$VALIDATE"

  if ! grep -q '"time"' "$VALIDATE"; then
    perl -0777 -i -pe 's/import \(/import (\n\t"time"/;' "$VALIDATE"
  fi
fi

echo "==> Formatting..."
gofmt -w cmd internal

echo "==> Building..."
go build ./cmd/restless

echo "==> Committing..."
git add -A
git commit -m "feat: add global --timeout flag (default 7s)" || true

echo
echo "✓ Done."
echo
echo "Examples:"
echo "  restless probe https://api.github.com --timeout 3"
echo "  restless validate --spec openapi.yaml --base https://api.example.com --timeout 12"
