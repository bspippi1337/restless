#!/usr/bin/env bash
# scripts/install-from-ghpages.sh
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/<user>/<repo>/main/scripts/install-from-ghpages.sh | bash
#
# Env vars:
#   APT_REPO_URL="https://<user>.github.io/<repo>/"
#   APT_CODENAME="stable"
#   APT_COMPONENT="main"
#   PKG_NAME="restless"
#
set -euo pipefail

APT_REPO_URL="${APT_REPO_URL:-}"
APT_CODENAME="${APT_CODENAME:-stable}"
APT_COMPONENT="${APT_COMPONENT:-main}"
PKG_NAME="${PKG_NAME:-restless}"

if [[ -z "$APT_REPO_URL" ]]; then
  echo "Set APT_REPO_URL, example:" >&2
  echo "  APT_REPO_URL=\"https://bspippi1337.github.io/restless/\" $0" >&2
  exit 2
fi

LIST_FILE="/etc/apt/sources.list.d/${PKG_NAME}.list"

echo "==> Adding APT source: ${APT_REPO_URL} ${APT_CODENAME} ${APT_COMPONENT}"
echo "deb [trusted=yes] ${APT_REPO_URL} ${APT_CODENAME} ${APT_COMPONENT}" | sudo tee "$LIST_FILE" >/dev/null

echo "==> apt update"
sudo apt-get update -y

echo "==> Installing: ${PKG_NAME}"
sudo apt-get install -y "${PKG_NAME}"

echo "✅ Installed. Try: ${PKG_NAME} --help"
