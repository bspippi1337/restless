#!/usr/bin/env bash
set -euo pipefail

FILE="cmd/restless/openapi_cli.go"

echo "==> Cleaning previous broken patches"
sed -i '/WARNING: spec load line not replaced/d' "$FILE" || true

echo "==> Replacing RawPath loader with disk-based loader"

# Replace exact LoadSpecFromFile call safely
perl -0777 -i -pe '
s/spec,\s*err\s*:=\s*openapi\.LoadSpecFromFile\([^\)]*\)/
openapiDir, err := openapi.DefaultDir()
if err != nil {
	fmt.Println("ERROR:", err)
	os.Exit(1)
}

specPath := filepath.Join(openapiDir, ra.ID+".json")

raw, err := os.ReadFile(specPath)
if err != nil {
	fmt.Println("ERROR: spec:", err)
	os.Exit(1)
}

spec, err := openapi.LoadSpec(raw)
/s
' "$FILE"

echo "==> Ensuring required imports"

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
git commit -m "fix(openapi): load spec from disk via ID instead of RawPath"

echo "==> Done"
