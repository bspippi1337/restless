#!/usr/bin/env bash
set -euo pipefail

RAW_VERSION="${1:-0.0.0}"
ARCH="${ARCH:-amd64}"

# --- Normalize version for Debian ---
# remove leading v
VER="${RAW_VERSION#v}"

# convert git describe patterns to Debian friendly
# 6.0.0-16-gabcdef -> 6.0.0+git16.gabcdef
VER=$(echo "$VER" | sed -E 's/-([0-9]+)-g([0-9a-f]+)/+git\1.g\2/')

# remove "-dirty"
VER="${VER/-dirty/}"

VERSION="$VER"

if [[ ! -x build/restless ]]; then
  echo "ERR: build/restless missing. Run: make build"
  exit 1
fi

rm -rf .deb-build
mkdir -p .deb-build/restless/DEBIAN
mkdir -p .deb-build/restless/usr/local/bin

cat > .deb-build/restless/DEBIAN/control <<EOF
Package: restless
Version: ${VERSION}
Section: utils
Priority: optional
Architecture: ${ARCH}
Maintainer: Pippi Tednes
Description: Terminal-first API discovery and exploration tool
EOF

install -m755 build/restless .deb-build/restless/usr/local/bin/restless

mkdir -p dist
dpkg-deb --build .deb-build/restless "dist/restless_${VERSION}_${ARCH}.deb"

echo "Debian package built:"
echo "dist/restless_${VERSION}_${ARCH}.deb"
