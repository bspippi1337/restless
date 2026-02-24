#!/usr/bin/env bash
set -euo pipefail

PROJECT="restless"
APT_BASE="https://bspippi1337.github.io/restless"
BREW_TAP="bspippi1337/restless"

echo "==> Verifying APT repository..."

# Check InRelease exists
curl -fsSL "${APT_BASE}/dists/stable/InRelease" >/dev/null \
  && echo "✓ InRelease reachable" \
  || { echo "✗ InRelease missing"; exit 1; }

# Check Packages.gz exists
curl -fsSL "${APT_BASE}/dists/stable/main/binary-amd64/Packages.gz" >/dev/null \
  && echo "✓ Packages.gz reachable" \
  || { echo "✗ Packages.gz missing"; exit 1; }

# Check .deb exists
DEB_URL=$(curl -fsSL "${APT_BASE}/dists/stable/main/binary-amd64/Packages.gz" \
  | gunzip -c \
  | grep Filename | head -n1 | awk '{print $2}')

if [ -n "${DEB_URL}" ]; then
  curl -fsSL "${APT_BASE}/${DEB_URL}" >/dev/null \
    && echo "✓ .deb file reachable (${DEB_URL})" \
    || { echo "✗ .deb missing"; exit 1; }
else
  echo "✗ Could not parse .deb filename"
  exit 1
fi

echo
echo "==> Verifying Homebrew..."

if brew tap | grep -q "${BREW_TAP}"; then
  echo "✓ Tap already installed"
else
  brew tap "${BREW_TAP}" >/dev/null \
    && echo "✓ Tap added" \
    || { echo "✗ Could not add tap"; exit 1; }
fi

brew info "${PROJECT}" >/dev/null \
  && echo "✓ Brew formula OK" \
  || { echo "✗ Brew formula missing"; exit 1; }

echo
echo "======================================"
echo "All checks passed."
echo "APT and Homebrew look healthy."
echo "======================================"
