#!/bin/bash
set -e
curl -L https://github.com/bspippi1337/restless/releases/latest/download/restless-linux-amd64 \
  -o /usr/local/bin/restless
chmod +x /usr/local/bin/restless
echo "Restless installed."
