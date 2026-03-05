#!/usr/bin/env bash
set -e

PROJECT="restless"
ARCH="amd64"
VERSION="$(git describe --tags --always)"
MAINTAINER="Pippi <you@example.com>"
DESCRIPTION="Contract-aware API inspection engine"

ROOT="$(pwd)"
BUILD_DIR="$ROOT/.deb-build"
APT_DIR="$ROOT/apt-repo"
DIST="stable"
COMP="main"

fail() { echo "❌ $1"; exit 1; }
pass() { echo "✔ $1"; }

echo "== RESTLESS APT BUILDER =="

command -v dpkg-deb >/dev/null || fail "dpkg-deb missing"
command -v go >/dev/null || fail "go missing"

# Clean
rm -rf "$BUILD_DIR"
mkdir -p "$BUILD_DIR"

# 1️⃣ Build binary
echo "Building binary..."
go build -trimpath -ldflags="-s -w" -o "$BUILD_DIR/$PROJECT" ./cmd/restless || fail "Build failed"
pass "Binary built"

# 2️⃣ Debian package structure
PKG_DIR="$BUILD_DIR/${PROJECT}_${VERSION}_${ARCH}"
mkdir -p "$PKG_DIR/DEBIAN"
mkdir -p "$PKG_DIR/usr/local/bin"

cp "$BUILD_DIR/$PROJECT" "$PKG_DIR/usr/local/bin/"

# 3️⃣ Control file
cat > "$PKG_DIR/DEBIAN/control" <<EOF
Package: $PROJECT
Version: ${VERSION#v}
Section: utils
Priority: optional
Architecture: $ARCH
Maintainer: $MAINTAINER
Description: $DESCRIPTION
EOF

# 4️⃣ Permissions
chmod 755 "$PKG_DIR/usr/local/bin/$PROJECT"

# 5️⃣ Build .deb
echo "Building .deb..."
dpkg-deb --build "$PKG_DIR" >/dev/null
DEB_FILE="$BUILD_DIR/${PROJECT}_${VERSION#v}_${ARCH}.deb"
mv "$BUILD_DIR/${PROJECT}_${VERSION}_${ARCH}.deb" "$DEB_FILE"
pass ".deb created"

# 6️⃣ APT repo layout
POOL_DIR="$APT_DIR/pool/$COMP/${PROJECT:0:1}/$PROJECT"
BIN_DIR="$APT_DIR/dists/$DIST/$COMP/binary-$ARCH"

mkdir -p "$POOL_DIR"
mkdir -p "$BIN_DIR"

cp "$DEB_FILE" "$POOL_DIR/"

# 7️⃣ Generate Packages
echo "Generating Packages index..."
dpkg-scanpackages "$APT_DIR/pool" > "$BIN_DIR/Packages" 2>/dev/null
gzip -kf "$BIN_DIR/Packages"
pass "Packages index created"

# 8️⃣ Optional GPG signing
if command -v gpg >/dev/null; then
  echo "Signing repository..."
  gpg --default-key "$MAINTAINER" --output "$APT_DIR/dists/$DIST/Release.gpg" -ba "$BIN_DIR/Packages" || true
fi

pass "APT repo ready"

echo
echo "🎯 APT repository built at:"
echo "$APT_DIR"
echo
echo "Next steps:"
echo "1. Push apt-repo/ to GitHub Pages branch"
echo "2. Configure users to add repo"
