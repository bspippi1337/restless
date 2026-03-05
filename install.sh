#!/usr/bin/env sh
set -e

REPO="bspippi1337/restless"
BIN="restless"

OS=$(uname | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

URL="https://github.com/$REPO/releases/latest/download/${BIN}-${OS}-${ARCH}"

echo "Installing $BIN ($OS-$ARCH)..."

TMP=$(mktemp)

if command -v curl >/dev/null 2>&1; then
  curl -fsSL "$URL" -o "$TMP"
elif command -v wget >/dev/null 2>&1; then
  wget -qO "$TMP" "$URL"
else
  echo "Need curl or wget"
  exit 1
fi

chmod +x "$TMP"

if [ -w "/usr/local/bin" ]; then
  mv "$TMP" /usr/local/bin/$BIN
else
  sudo mv "$TMP" /usr/local/bin/$BIN
fi

echo
echo "✔ Restless installed."
echo
echo "Run:"
echo "  restless blckswan https://api.github.com"
echo
