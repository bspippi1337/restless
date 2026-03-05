#!/usr/bin/env bash
set -e

OUT=demo
CAST=$OUT.cast

echo "Recording demo..."

asciinema rec "$CAST" -c "
clear
echo '⚡ Restless API exploration'
sleep 1

echo
echo '$ restless discover https://api.github.com'
sleep 1
./build/restless discover https://api.github.com
sleep 2

echo
echo '$ restless map https://api.github.com'
sleep 1
./build/restless map https://api.github.com
sleep 2

echo
echo '$ restless auto https://api.github.com'
sleep 1
./build/restless auto https://api.github.com
sleep 2
" --overwrite

echo
echo "Rendering SVG animation..."

svg-term --in "$CAST" --out demo.svg --window --padding 20

echo
echo "Done:"
echo "demo.svg"
