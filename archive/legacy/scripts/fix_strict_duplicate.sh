#!/usr/bin/env bash
set -euo pipefail

FILE="internal/modules/openapi/run.go"

echo "==> Removing duplicate strictEnabled()"

# Remove all existing strictEnabled definitions
awk '
/func strictEnabled\(\)/ {skip=1}
skip && /^}/ {skip=0; next}
!skip {print}
' "$FILE" > "$FILE.tmp"

mv "$FILE.tmp" "$FILE"

# Ensure os import exists
if ! grep -q '"os"' "$FILE"; then
  sed -i '/^import (/a\	"os"' "$FILE"
fi

# Append single clean definition
cat >> "$FILE" <<'EOT'

func strictEnabled() bool {
	return os.Getenv("RESTLESS_STRICT") == "1"
}
EOT

echo "==> Formatting"
gofmt -w "$FILE"

echo "==> Building"
go build -o restless-v2 ./cmd/restless-v2

echo "✅ strictEnabled cleaned and rebuilt"
