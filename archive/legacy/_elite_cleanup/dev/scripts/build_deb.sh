
#!/usr/bin/env sh
set -eu

VERSION=${1:-0.0.0}
PKGDIR=build/deb/restless_$VERSION

mkdir -p $PKGDIR/DEBIAN
mkdir -p $PKGDIR/usr/local/bin

cp restless $PKGDIR/usr/local/bin/

cat > $PKGDIR/DEBIAN/control <<EOF
Package: restless
Version: $VERSION
Section: utils
Priority: optional
Architecture: amd64
Maintainer: Restless Maintainers
Description: Contract-aware API probing and validation tool
EOF

dpkg-deb --build $PKGDIR
