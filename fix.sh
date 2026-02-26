#!/usr/bin/env bash
set -euo pipefail

# --- safety ---
if ! git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  echo "Not inside a git repo."
  exit 1
fi

if [[ -n "$(git status --porcelain)" ]]; then
  echo "Working tree not clean. Commit/stash first."
  git status --porcelain
  exit 1
fi

BR="optimize/ci-green"
git switch -c "$BR" 2>/dev/null || git switch "$BR"

mkdir -p .github/workflows

# --- CI workflow: proper YAML + cache + lint ---
cat > .github/workflows/ci.yml <<'YAML'
name: CI

on:
  push:
    branches: [ main ]
  pull_request:

permissions:
  contents: read

jobs:
  test:
    name: Test (${{ matrix.os }}, Go ${{ matrix.go }})
    runs-on: ${{ matrix.os }}
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go: ["1.22.x"]

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go }}
          cache: true

      - name: gofmt (must be clean)
        shell: bash
        run: |
          set -euo pipefail
          BAD="$(gofmt -l . | tr -d '\r')"
          if [[ -n "$BAD" ]]; then
            echo "gofmt needed on:"
            echo "$BAD"
            exit 1
          fi

      - name: go mod tidy (must be clean)
        shell: bash
        run: |
          set -euo pipefail
          go mod tidy
          git diff --exit-code

      - name: go test
        run: go test ./...

      - name: go vet
        run: go vet ./...

  lint:
    name: golangci-lint (ubuntu)
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: "1.22.x"
          cache: true

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest
          args: --timeout=5m
YAML

# --- golangci config: proper YAML ---
cat > .golangci.yml <<'YAML'
run:
  timeout: 5m

linters:
  enable:
    - govet
    - errcheck
    - staticcheck
    - revive
YAML

# --- Makefile: readable again ---
cat > Makefile <<'MAKE'
.PHONY: tidy test run build build-all clean doctor lint

tidy:
	go mod tidy

test:
	go test ./...

lint:
	golangci-lint run ./...

run:
	go run ./cmd/restless

build:
	mkdir -p bin
	go build -o bin/restless ./cmd/restless

build-all:
	mkdir -p dist
	GOOS=linux   GOARCH=amd64 go build -o dist/restless_linux_amd64 ./cmd/restless
	GOOS=darwin  GOARCH=amd64 go build -o dist/restless_darwin_amd64 ./cmd/restless
	GOOS=windows GOARCH=amd64 go build -o dist/restless_windows_amd64.exe ./cmd/restless

clean:
	rm -rf bin dist build logs

doctor:
	go run ./cmd/restless doctor
MAKE

# --- local sanity pass ---
gofmt -w .
go mod tidy
go test ./...

git add .github/workflows/ci.yml .golangci.yml Makefile go.mod go.sum 2>/dev/null || true
git add -A

git commit -m "ci: fix workflows + lint + tidy checks; restore readable config files" || true

echo
echo "✅ Done on branch: $BR"
echo "Next: push and open PR:"
echo "  git push -u origin $BR"
