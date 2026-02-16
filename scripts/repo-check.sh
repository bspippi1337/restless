#!/usr/bin/env sh
set -eu
echo "[repo-check] verifying repo references are bspippi1337/restless"
bad=$(grep -RIn --exclude-dir=.git --exclude=*.png --exclude=*.jpg --exclude=*.zip -E "bspippi1337/(lessmess|stressless|stressless-win|oldstressless|restless-win)" . || true)
if [ -n "$bad" ]; then
  echo "Found wrong repo refs:"
  echo "$bad"
  exit 1
fi
echo "OK"
