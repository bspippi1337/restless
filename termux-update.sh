#!/data/data/com.termux/files/usr/bin/sh
set -e

echo "==> Updating Restless repo with bundled docs"

if [ ! -d .git ]; then
  echo "❌ Not inside a git repo"
  exit 1
fi

mkdir -p docs
cp CLI-HELP.md docs/CLI-HELP.md
cp README.md README.md

git add docs/CLI-HELP.md README.md
git commit -m "docs: update CLI handbook + README" || true

echo "✅ Repo updated."
