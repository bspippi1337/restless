#!/usr/bin/env bash
set -euo pipefail

echo "⚡ Restless doctor starting"
echo

ROOT="$(git rev-parse --show-toplevel 2>/dev/null || pwd)"
cd "$ROOT"

echo "repo: $ROOT"
echo

# ------------------------------------------------
# ensure goimports exists
# ------------------------------------------------

if ! command -v goimports >/dev/null; then
  echo "installing goimports..."
  go install golang.org/x/tools/cmd/goimports@latest
  export PATH="$PATH:$(go env GOPATH)/bin"
fi

# ------------------------------------------------
# fix formatting + imports
# ------------------------------------------------

echo "1️⃣ formatting code"

goimports -w .
gofmt -s -w .

# ------------------------------------------------
# tidy modules
# ------------------------------------------------

echo
echo "2️⃣ module hygiene"

go mod tidy

# ------------------------------------------------
# ensure CLI commands registered
# ------------------------------------------------

echo
echo "3️⃣ verifying CLI registration"

ROOTFILE="internal/cli/root.go"

fix_cmd() {
  local cmd="$1"
  if ! grep -q "$cmd" "$ROOTFILE"; then
    echo "adding $cmd"
    sed -i "/return cmd/i\\
\t$cmd
" "$ROOTFILE"
  fi
}

fix_cmd "cmd.AddCommand(NewAutoCmd())"
fix_cmd "cmd.AddCommand(NewSmartCmd())"
fix_cmd "cmd.AddCommand(NewSwarmCmd())"

gofmt -w "$ROOTFILE"

# ------------------------------------------------
# vet
# ------------------------------------------------

echo
echo "4️⃣ go vet"

go vet ./... || true

# ------------------------------------------------
# build
# ------------------------------------------------

echo
echo "5️⃣ building restless"

mkdir -p build

CGO_ENABLED=0 go build \
  -trimpath \
  -ldflags "-s -w" \
  -o build/restless \
  ./cmd/restless

echo "build OK"

# ------------------------------------------------
# generate completions
# ------------------------------------------------

echo
echo "6️⃣ generating completions"

mkdir -p dist

./build/restless completion bash > dist/restless.bash || true
./build/restless completion zsh > dist/_restless || true

# ------------------------------------------------
# run tests
# ------------------------------------------------

echo
echo "7️⃣ running tests"

go test ./... || true

# ------------------------------------------------
# verify commands exist
# ------------------------------------------------

echo
echo "8️⃣ verifying CLI commands"

./build/restless --help | grep -E "auto|smart|swarm" || {
  echo "ERROR: commands missing"
  exit 1
}

echo
echo "9️⃣ smoke tests"

./build/restless --version || true
./build/restless --help | head -n 20

echo
echo "✅ Restless repo is healthy"
echo
echo "try:"
echo
echo "  ./build/restless auto https://api.github.com"
echo
