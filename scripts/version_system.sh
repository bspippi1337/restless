#!/usr/bin/env bash
set -euo pipefail

VERSION="4.0.4"

echo "==> Installing centralized version system ($VERSION)"

mkdir -p internal/version

cat > internal/version/version.go <<EOT
package version

var Version = "$VERSION"
EOT

# Inject version flag in main if not exists
MAIN="cmd/restless/main.go"

if ! grep -q 'internal/version' "$MAIN"; then
  sed -i '/import (/a\
\t"github.com/bspippi1337/restless/internal/version"' "$MAIN"
fi

if ! grep -q 'case "--version"' "$MAIN"; then
  sed -i '/switch os.Args\[1\]/a\
\tcase "--version":\n\t\tfmt.Println("restless", version.Version)\n\t\treturn' "$MAIN"
fi

echo "==> Updating Makefile for ldflags injection"

if ! grep -q 'VERSION ?=' Makefile 2>/dev/null; then
cat >> Makefile <<'EOT'

VERSION ?= 4.0.4

build:
	go build -ldflags "-X github.com/bspippi1337/restless/internal/version.Version=$(VERSION)" -o restless ./cmd/restless

build-all:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/bspippi1337/restless/internal/version.Version=$(VERSION)" -o dist/restless_linux_amd64 ./cmd/restless
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-X github.com/bspippi1337/restless/internal/version.Version=$(VERSION)" -o dist/restless_linux_arm64 ./cmd/restless
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-X github.com/bspippi1337/restless/internal/version.Version=$(VERSION)" -o dist/restless_darwin_amd64 ./cmd/restless
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags "-X github.com/bspippi1337/restless/internal/version.Version=$(VERSION)" -o dist/restless_darwin_arm64 ./cmd/restless
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-X github.com/bspippi1337/restless/internal/version.Version=$(VERSION)" -o dist/restless_windows_amd64.exe ./cmd/restless
EOT
fi

echo "==> Formatting"
gofmt -w .

echo "==> Testing"
go test ./...

echo "==> Building"
make build

echo "==> Checking version output"
./restless --version

echo "✅ Version system installed"
