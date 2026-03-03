#!/usr/bin/env bash
set -euo pipefail

PROJECT="restless"
DIST="stable"
COMP="main"
ARCHES=("amd64" "arm64")

GPG_KEY="5EA836A98EB9E38A51466BF6A0CB94CCA7E69627"
MAINTAINER="Restless Release <bspippi1337@gmail.com>"
HOMEPAGE="https://github.com/bspippi1337/restless"
DESCRIPTION="Contract-aware API inspection engine"

APT_DIR="apt-repo"
BUILD_DIR=".apt-build"
OUT_DIR="dist"

fail(){ echo "❌ $1" >&2; exit 1; }
pass(){ echo "✔ $1"; }
need(){ command -v "$1" >/dev/null 2>&1 || fail "Missing dependency: $1"; }

need go
need dpkg-deb
need dpkg-scanpackages
need apt-ftparchive
need gpg

version="$(git describe --tags --always)"
version="${version#v}"

echo "== Building version $version =="

rm -rf "$BUILD_DIR" "$OUT_DIR" "$APT_DIR"
mkdir -p "$BUILD_DIR" "$OUT_DIR"
mkdir -p "$APT_DIR/pool/$COMP/r/$PROJECT"

for arch in "${ARCHES[@]}"; do
  echo "Building linux/$arch"
  GOOS=linux GOARCH="$arch" CGO_ENABLED=0 \
    go build -trimpath -ldflags="-s -w" \
    -o "$BUILD_DIR/$PROJECT-$arch" ./cmd/restless

  pkgdir="$BUILD_DIR/${PROJECT}_${version}_${arch}"
  mkdir -p "$pkgdir/DEBIAN"
  mkdir -p "$pkgdir/usr/local/bin"

  install -m 0755 "$BUILD_DIR/$PROJECT-$arch" \
    "$pkgdir/usr/local/bin/$PROJECT"

  cat > "$pkgdir/DEBIAN/control" <<CONTROL
Package: $PROJECT
Version: $version
Section: utils
Priority: optional
Architecture: $arch
Maintainer: $MAINTAINER
Homepage: $HOMEPAGE
Description: $DESCRIPTION
CONTROL

  dpkg-deb --build "$pkgdir" \
    "$OUT_DIR/${PROJECT}_${version}_${arch}.deb" >/dev/null

  cp "$OUT_DIR/${PROJECT}_${version}_${arch}.deb" \
     "$APT_DIR/pool/$COMP/r/$PROJECT/"
done

pass ".deb packages built"

for arch in "${ARCHES[@]}"; do
  bindir="$APT_DIR/dists/$DIST/$COMP/binary-$arch"
  mkdir -p "$bindir"
  dpkg-scanpackages "$APT_DIR/pool" > "$bindir/Packages" 2>/dev/null
  gzip -kf "$bindir/Packages"
done

pass "Packages index generated"

mkdir -p "$APT_DIR/dists/$DIST"

cat > "$BUILD_DIR/release.conf" <<REL
APT::FTPArchive::Release {
  Origin "restless";
  Label "restless";
  Suite "$DIST";
  Codename "$DIST";
  Architectures "amd64 arm64";
  Components "$COMP";
  Description "$DESCRIPTION";
};
REL

apt-ftparchive -c "$BUILD_DIR/release.conf" \
  release "$APT_DIR/dists/$DIST" \
  > "$APT_DIR/dists/$DIST/Release"

gpg --batch --yes --local-user "$GPG_KEY" \
    --clearsign \
    -o "$APT_DIR/dists/$DIST/InRelease" \
    "$APT_DIR/dists/$DIST/Release"

gpg --batch --yes --local-user "$GPG_KEY" \
    -abs \
    -o "$APT_DIR/dists/$DIST/Release.gpg" \
    "$APT_DIR/dists/$DIST/Release"

gpg --export "$GPG_KEY" | gpg --dearmor -o "$APT_DIR/repo.gpg"

pass "Repository signed"

echo
echo "Publishing to gh-pages..."

git fetch origin
git branch gh-pages 2>/dev/null || true
git worktree add -f .apt-pages gh-pages

rm -rf .apt-pages/*
cp -a "$APT_DIR"/* .apt-pages/

cd .apt-pages
git add -A
git commit -m "chore(apt): signed multi-arch repo $version" || true
git push origin gh-pages
cd ..

git worktree remove .apt-pages

pass "Published to gh-pages"
echo
echo "APT repo ready."
