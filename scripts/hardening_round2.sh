#!/usr/bin/env bash
set -euo pipefail

echo "==> Hardening Round 2: Error normalization"

sed -i 's/fmt.Println("error:/fmt.Println("ERROR:/g' cmd/restless-v2/*.go || true
sed -i 's/fmt.Println("spec error:/fmt.Println("ERROR: spec:/g' cmd/restless-v2/*.go || true
sed -i 's/fmt.Println("build error:/fmt.Println("ERROR: build:/g' cmd/restless-v2/*.go || true
sed -i 's/fmt.Println("request error:/fmt.Println("ERROR: request:/g' cmd/restless-v2/*.go || true

gofmt -w cmd/restless-v2
go build -o restless-v2 ./cmd/restless-v2

echo "✅ Round 2 installed"
