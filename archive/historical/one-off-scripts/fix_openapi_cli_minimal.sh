#!/usr/bin/env bash
set -euo pipefail

FILE="cmd/restless/openapi_cli.go"

echo "==> Removing broken WARNING lines"
sed -i '/WARNING:/d' "$FILE" || true

echo "==> Replacing RawPath loader correctly"

perl -0777 -i -pe '
s/spec,\s*err\s*:=\s*openapi\.LoadSpecFromFile\([^\)]*\)/openapiDir, err := openapi.DefaultDir()
if err != nil {
	fmt.Println("ERROR:", err)
	os.Exit(1)
}

specPath := filepath.Join(openapiDir, ra.ID+".json")

spec, err := openapi.LoadSpecFromFile(specPath)
/s
' "$FILE"

echo "==> Ensuring filepath import"
grep -q '"path/filepath"' "$FILE" || \
  sed -i '/import (/a \	"path/filepath"' "$FILE"

echo "==> Formatting"
go fmt ./...

echo "==> Building"
go build -o restless ./cmd/restless

echo "==> Testing"
go test ./...

echo "==> Commit"
git add "$FILE"
git commit -m "fix(openapi): load spec from disk via ID instead of RawPath (corrected)"

echo "Done."
