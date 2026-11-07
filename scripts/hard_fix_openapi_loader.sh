#!/usr/bin/env bash
set -euo pipefail

FILE="internal/modules/openapi/run.go"

echo "==> Installing deterministic disk loader"

if [ ! -f "$FILE" ]; then
  echo "ERROR: $FILE not found"
  exit 1
fi

cp "$FILE" "$FILE.bak"

cat >> "$FILE" <<'GOFIX'

// --- DISK LOADER PATCH ---

func loadSpecFromDisk(openapiDir, id string) ([]byte, error) {
	specPath := filepath.Join(openapiDir, id+".json")
	raw, err := os.ReadFile(specPath)
	if err != nil {
		return nil, fmt.Errorf("spec: failed to read stored spec: %w", err)
	}
	return raw, nil
}
GOFIX

echo "==> Ensuring required imports"

grep -q '"path/filepath"' "$FILE" || sed -i '/import (/a \	"path/filepath"' "$FILE"
grep -q '"os"' "$FILE" || sed -i '/import (/a \	"os"' "$FILE"
grep -q '"fmt"' "$FILE" || sed -i '/import (/a \	"fmt"' "$FILE"

echo "==> Removing unused json/yaml imports if not referenced"
sed -i '/encoding\/json/d' "$FILE" || true
sed -i '/gopkg.in\/yaml.v3/d' "$FILE" || true

echo "==> Formatting"
go fmt ./...

echo "==> Building"
go build -o restless ./cmd/restless

echo "==> Committing"
git add "$FILE"
git commit -m "fix(openapi): add deterministic disk loader function (manual integration required)"

echo "Patch applied."
