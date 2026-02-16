#!/data/data/com.termux/files/usr/bin/sh
set -eu
command -v git >/dev/null 2>&1 || pkg install -y git
command -v go >/dev/null 2>&1 || pkg install -y golang
command -v make >/dev/null 2>&1 || pkg install -y make
make test
make build
echo "✅ OK"
