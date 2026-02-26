#!/usr/bin/env bash
set -euo pipefail

echo "==> Hardening Round 3: Strict mode"

cat >> internal/modules/openapi/run.go <<'EOT'

func strictEnabled() bool {
	return os.Getenv("RESTLESS_STRICT") == "1"
}
EOT

gofmt -w internal/modules/openapi/run.go
go build -o restless-v2 ./cmd/restless-v2

echo "✅ Round 3 installed"
