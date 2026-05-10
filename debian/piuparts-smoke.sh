#!/bin/sh
set -eu

printf '[piuparts-smoke] install/remove cycle\n'

PKG=$(ls ../restless_*.deb | head -n1)

sudo dpkg -i "$PKG"
command -v restless >/dev/null
restless --help >/dev/null
sudo dpkg -r restless

printf '[piuparts-smoke] ok\n'
