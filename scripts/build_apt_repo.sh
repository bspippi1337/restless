#!/usr/bin/env bash
set -euo pipefail

VERSION="${1:-0.0.0}"
DIST="${DIST:-stable}"
ARCH="${ARCH:-amd64}"

DEB="dist/restless_${VERSION}_${ARCH}.deb"
if [[ ! -f "$DEB" ]]; then
  echo "ERR: $DEB missing. Run: make deb"
  exit 1
fi

# Minimal apt repo structure (for GitHub Pages / static hosting)
rm -rf .apt-repo
mkdir -p ".apt-repo/dists/${DIST}/main/binary-${ARCH}"
mkdir -p ".apt-repo/pool/main/r/restless"

cp "$DEB" ".apt-repo/pool/main/r/restless/"

# Packages index
PKGDIR=".apt-repo/dists/${DIST}/main/binary-${ARCH}"
( cd .apt-repo && dpkg-scanpackages -m pool /dev/null ) > "${PKGDIR}/Packages"
gzip -9c "${PKGDIR}/Packages" > "${PKGDIR}/Packages.gz"

# Release file (unsigned, but works; you can add signing later)
cat > ".apt-repo/dists/${DIST}/Release" <<REL
Origin: restless
Label: restless
Suite: ${DIST}
Codename: ${DIST}
Architectures: ${ARCH}
Components: main
Description: Restless APT repo
Date: $(date -Ru)
REL

echo "APT repo staged in .apt-repo/"
echo "Tip: publish .apt-repo/ via GitHub Pages, S3, or nginx."
