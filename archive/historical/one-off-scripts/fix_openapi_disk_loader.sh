#!/usr/bin/env bash
set -euo pipefail

echo "==> Hardening OpenAPI run: forcing disk-only spec loading"

TARGET="internal/modules/openapi/run.go"

if [ ! -f "$TARGET" ]; then
  echo "ERROR: $TARGET not found"
  exit 1
fi

# Backup
cp "$TARGET" "$TARGET.bak"

# Replace spec loading logic with deterministic disk loader
awk '
BEGIN { replaced=0 }
/SPEC_LOADER_START/ { print; skip=1; next }
/SPEC_LOADER_END/ { skip=0; replaced=1; print diskLoader; next }
!skip { print }
END {
  if (replaced==0) {
    print ""
    print "// SPEC_LOADER_START"
    print "specPath := filepath.Join(openapiDir, id+\".json\")"
    print "raw, err := os.ReadFile(specPath)"
    print "if err != nil {"
    print "    return fmt.Errorf(\"spec: failed to read stored spec: %w\", err)"
    print "}"
    print ""
    print "// parse JSON first"
    print "var spec any"
    print "if err := json.Unmarshal(raw, &spec); err != nil {"
    print "    // try YAML fallback"
    print "    if err := yaml.Unmarshal(raw, &spec); err != nil {"
    print "        return fmt.Errorf(\"spec: failed to parse spec as JSON or YAML\")"
    print "    }"
    print "}"
    print "// SPEC_LOADER_END"
  }
}
' "$TARGET" > "$TARGET.tmp"

mv "$TARGET.tmp" "$TARGET"

echo "==> Ensuring imports"
grep -q '"encoding/json"' "$TARGET" || sed -i '/package/a import "encoding/json"' "$TARGET"
grep -q '"gopkg.in/yaml.v3"' "$TARGET" || sed -i '/package/a import "gopkg.in/yaml.v3"' "$TARGET"
grep -q '"path/filepath"' "$TARGET" || sed -i '/package/a import "path/filepath"' "$TARGET"
grep -q '"os"' "$TARGET" || sed -i '/package/a import "os"' "$TARGET"
grep -q '"fmt"' "$TARGET" || sed -i '/package/a import "fmt"' "$TARGET"

echo "==> Formatting"
go fmt ./...

echo "==> Tidy"
go mod tidy

echo "==> Testing"
go test ./...

echo "==> Building"
go build -o restless ./cmd/restless

echo "==> Committing fix"
git add "$TARGET"
git add go.mod go.sum || true
git commit -m "fix(openapi): enforce disk-only spec loading for run (teacher-stable)"
echo "==> Done"
