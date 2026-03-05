#!/usr/bin/env bash
set -e

TARGET=${1:-https://api.github.com}

mkdir -p dist

echo "⚡ RESTLESS MAGISWARM"
echo "release the BLCKSWAN swarm"
echo

go build -buildvcs=false -o build/restless ./cmd/restless

./build/restless magiswarm "$TARGET" --max-requests 80 --out dist || true

REPORT=$(ls dist/magiswarm_*.json | head -n1)

python3 scripts/reactor_topology.py "$REPORT" > reactor/topology.txt

scripts/reactor_animation.sh reactor/topology.txt

echo
echo "⚡ reactor complete"
