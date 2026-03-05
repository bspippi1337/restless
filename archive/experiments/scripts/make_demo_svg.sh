#!/usr/bin/env bash
set -e

mkdir -p reactor

CAST=reactor/demo.cast
SVG=demo.svg

rm -f "$CAST" "$SVG"

asciinema rec "$CAST" \
  --overwrite \
  --command "./scripts/run_reactor_demo.sh"

if command -v svg-term >/dev/null 2>&1; then
  svg-term --in "$CAST" --out "$SVG"
fi
