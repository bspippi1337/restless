#!/usr/bin/env sh
set -eu

REPO="${1:-bspippi1337/restless}"
TAG="${2:-}"

if [ -z "$TAG" ]; then
  echo "Tag required for this simple installer."
  echo "Example: $0 bspippi1337/restless v0.2.0-alpha"
  exit 2
fi

ASSET="restless_linux_amd64"
URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}"

echo "Downloading: $URL"
if command -v wget >/dev/null 2>&1; then
  wget -O restless "$URL"
elif command -v curl >/dev/null 2>&1; then
  curl -L "$URL" -o restless
else
  echo "Need wget or curl."
  exit 1
fi

chmod +x restless
mkdir -p "$HOME/.local/bin"
mv restless "$HOME/.local/bin/restless"
echo "Installed to $HOME/.local/bin/restless"
echo "Run: restless"
