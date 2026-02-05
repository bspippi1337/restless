#!/usr/bin/env bash
set -euo pipefail

echo "[restless] Applying finishing kit…"

ts="$(date +%Y%m%d-%H%M%S)"
backup_dir=".finishkit-backup-$ts"
mkdir -p "$backup_dir"

backup_if_exists () {
  local p="$1"
  if [ -e "$p" ]; then
    mkdir -p "$backup_dir/$(dirname "$p")"
    cp -a "$p" "$backup_dir/$p"
    echo "  backed up: $p -> $backup_dir/$p"
  fi
}

# backup files that we overwrite
backup_if_exists README.md
backup_if_exists LICENSE
backup_if_exists .goreleaser.yaml
backup_if_exists .github/workflows/ci.yml
backup_if_exists .github/workflows/release.yml
backup_if_exists .github/workflows/npm-publish.yml
backup_if_exists brand
backup_if_exists npm
backup_if_exists PUBLISHING.md

# copy in new files
cp -a brand ./
cp -a .goreleaser.yaml ./
cp -a .github ./
cp -a npm ./
cp -a LICENSE ./
cp -a README.md ./
cp -a PUBLISHING.md ./

echo
echo "[restless] Done ✅"
echo "Backup: $backup_dir"
echo
echo "Next:"
echo "  git add -A"
echo "  git commit -m 'chore: finish kit (branding, release, npm wrapper)'"
echo "  git push"
echo
echo "Then set GitHub secret: NPM_TOKEN"
echo "And release: bump npm/package.json version, tag vX.Y.Z and push."
