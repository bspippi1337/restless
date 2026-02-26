bash -c 'set -euo pipefail

PKG="restless"
REPO="bspippi1337/restless"
TAP_REPO="bspippi1337/homebrew-restless"

RAW_VERSION=$(git describe --tags --always)
CLEAN_VERSION=$(echo "$RAW_VERSION" | sed "s/^v//" | sed "s/-dirty//")
TAG="v${CLEAN_VERSION}"

ASSET="${PKG}_${CLEAN_VERSION}_linux_amd64.tar.gz"
RELEASE_URL="https://github.com/${REPO}/releases/download/${TAG}/${ASSET}"

echo "Using VERSION=$CLEAN_VERSION"
echo "Using TAG=$TAG"

echo "Ensuring release exists..."
gh release view "$TAG" >/dev/null 2>&1 || \
  gh release create "$TAG" -t "$TAG" -n "Restless $CLEAN_VERSION"

echo "Building asset..."
mkdir -p dist/releases
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o dist/${PKG}_linux_amd64 ./cmd/restless
tar -czf dist/releases/${ASSET} -C dist ${PKG}_linux_amd64

echo "Uploading asset..."
gh release upload "$TAG" dist/releases/${ASSET} --clobber

echo "Downloading asset for SHA..."
TMPFILE=$(mktemp)
curl -L "$RELEASE_URL" -o "$TMPFILE"

SHA=$(shasum -a 256 "$TMPFILE" | awk "{print \$1}")
echo "SHA=$SHA"

echo "Ensuring brew tap exists..."
gh repo view "$TAP_REPO" >/dev/null 2>&1 || \
  gh repo create "$TAP_REPO" --public

TMP_TAP=$(mktemp -d)
git clone "https://github.com/${TAP_REPO}.git" "$TMP_TAP" >/dev/null 2>&1 || true
mkdir -p "$TMP_TAP/Formula"

cat > "$TMP_TAP/Formula/restless.rb" <<RUBY
class Restless < Formula
  desc "Restless adaptive API client"
  homepage "https://github.com/${REPO}"
  version "${CLEAN_VERSION}"

  on_linux do
    url "${RELEASE_URL}"
    sha256 "${SHA}"
    def install
      bin.install "${PKG}_linux_amd64" => "restless"
    end
  end

  test do
    system "#{bin}/restless", "--help"
  end
end
RUBY

(
  cd "$TMP_TAP"
  git add -A
  git commit -m "restless ${CLEAN_VERSION}" >/dev/null || true
  git push
)

echo ""
echo "✅ Release fixed"
echo "✅ Brew formula updated"
echo "Release URL:"
echo "$RELEASE_URL"
'
