#!/usr/bin/env bash
set -euo pipefail

FILE="cmd/restless/openapi_cli.go"

echo "==> Installing proper disk-based spec loader"

cp "$FILE" "$FILE.bak"

# Replace spec loading block
awk '
BEGIN { replaced=0 }
/spec, err := openapi.LoadSpecFromFile/ {
    print "\t\topenapiDir, err := openapi.DefaultDir()"
    print "\t\tif err != nil {"
    print "\t\t\tfmt.Println(\"ERROR:\", err)"
    print "\t\t\tos.Exit(1)"
    print "\t\t}"
    print ""
    print "\t\tspecPath := filepath.Join(openapiDir, ra.ID+\".json\")"
    print "\t\traw, err := os.ReadFile(specPath)"
    print "\t\tif err != nil {"
    print "\t\t\tfmt.Println(\"ERROR: spec:\", err)"
    print "\t\t\tos.Exit(1)"
    print "\t\t}"
    print ""
    print "\t\tspec, err := openapi.LoadSpec(raw)"
    replaced=1
    next
}
{ print }
END {
    if (replaced==0) {
        print "WARNING: spec load line not replaced"
    }
}
' "$FILE" > "$FILE.tmp"

mv "$FILE.tmp" "$FILE"

# Ensure imports
grep -q '"path/filepath"' "$FILE" || sed -i '/import (/a \	"path/filepath"' "$FILE"
grep -q '"os"' "$FILE" || sed -i '/import (/a \	"os"' "$FILE"

echo "==> Formatting"
go fmt ./...

echo "==> Building"
go build -o restless ./cmd/restless

echo "==> Testing"
go test ./...

echo "==> Commit"
git add "$FILE"
git commit -m "fix(openapi): load spec from disk via ID instead of RawPath"

echo "Done."
