# Restless ⚡

Universal API blade: discover, probe, simulate, and export.

## Install (Debian / Ubuntu)

echo "deb [trusted=yes] https://bspippi1337.github.io/restless/ ./" | sudo tee /etc/apt/sources.list.d/restless.list
sudo apt update
sudo apt install restless

## Quickstart

restless probe https://api.github.com
restless simulate https://api.github.com
restless smart https://api.github.com

Architecture:
smartcmd → discover → engine → simulator → export → app
