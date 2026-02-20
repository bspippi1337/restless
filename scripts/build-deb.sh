#!/usr/bin/env bash
# scripts/build-deb.sh
#
# Build a .deb from a prebuilt binary.
#
# Usage:
#   scripts/build-deb.sh ./dist/restless
#
# Env vars (sane defaults):
#   PKG_NAME=restless
#   PKG_VERSION=0.1.0
#   PKG_ARCH=amd64              # amd64|arm64|all
#   PKG_SECTION=utils
#   PKG_PRIORITY=optional
#   PKG_MAINTAINER="You <you@example.com>"
#   PKG_DESCRIPTION="Restless universal API client"
#   PKG_HOMEPAGE="https://github.com/<you>/<repo>"
#   PKG_LICENSE="MIT"
#   INSTALL_PATH=/usr/bin       # or /usr/local/bin
#   OUT_DIR=dist
#
set -euo pipefail

need() { command -v "$1" >/dev/null 2>&1 || { echo "Missing dependency: $1" >&2; exit 2; }; }
need dpkg-deb
need gzip
need tar

if [[ "${1:-}" == "" ]]; then
  echo "Usage: $0 path/to/binary" >&2
  exit 2
fi

BIN_PATH="$1"
[[ -f "$BIN_PATH" ]] || { echo "Binary not found: $BIN_PATH" >&2; exit 2; }
[[ -x "$BIN_PATH" ]] || chmod +x "$BIN_PATH" || true

PKG_NAME="${PKG_NAME:-restless}"
PKG_VERSION="${PKG_VERSION:-0.1.0}"
PKG_ARCH="${PKG_ARCH:-amd64}"
PKG_SECTION="${PKG_SECTION:-utils}"
PKG_PRIORITY="${PKG_PRIORITY:-optional}"
PKG_MAINTAINER="${PKG_MAINTAINER:-You <you@example.com>}"
PKG_DESCRIPTION="${PKG_DESCRIPTION:-A handy CLI tool}"
PKG_HOMEPAGE="${PKG_HOMEPAGE:-}"
PKG_LICENSE="${PKG_LICENSE:-}"
INSTALL_PATH="${INSTALL_PATH:-/usr/bin}"
OUT_DIR="${OUT_DIR:-dist}"

WORKDIR="$(mktemp -d)"
trap 'rm -rf "$WORKDIR"' EXIT

PKGROOT="$WORKDIR/pkgroot"
mkdir -p "$PKGROOT/DEBIAN"
mkdir -p "$PKGROOT${INSTALL_PATH}"

install -m 0755 "$BIN_PATH" "$PKGROOT${INSTALL_PATH}/${PKG_NAME}"

{
  echo "Package: ${PKG_NAME}"
  echo "Version: ${PKG_VERSION}"
  echo "Section: ${PKG_SECTION}"
  echo "Priority: ${PKG_PRIORITY}"
  echo "Architecture: ${PKG_ARCH}"
  echo "Maintainer: ${PKG_MAINTAINER}"
  [[ -n "$PKG_HOMEPAGE" ]] && echo "Homepage: ${PKG_HOMEPAGE}"
  [[ -n "$PKG_LICENSE" ]] && echo "License: ${PKG_LICENSE}"
  echo "Description: ${PKG_DESCRIPTION}"
} > "$PKGROOT/DEBIAN/control"

mkdir -p "$OUT_DIR"
OUT_DEB="${OUT_DIR}/${PKG_NAME}_${PKG_VERSION}_${PKG_ARCH}.deb"

dpkg-deb --build "$PKGROOT" "$OUT_DEB" >/dev/null
echo "✅ Built: $OUT_DEB"
