#!/usr/bin/env bash
set -euo pipefail

ROOT="$(pwd)"
OUT="$ROOT/hn_material"
BIN="$ROOT/build/restless"

mkdir -p "$OUT"

echo "== Restless HN material collector =="

echo
echo "1) System info"
{
echo "DATE: $(date -Iseconds)"
echo "HOST: $(uname -a)"
echo
echo "GO:"
go version || true
echo
echo "GIT:"
git rev-parse HEAD || true
git status || true
} > "$OUT/system.txt"

echo
echo "2) Building restless (if needed)"
if [ ! -x "$BIN" ]; then
    echo "Building..."
    go build -o "$BIN" ./cmd/restless || true
fi

if [ ! -x "$BIN" ]; then
    echo "restless binary not found. abort."
    exit 1
fi

echo
echo "3) CLI info"
{
"$BIN" --help || true
} > "$OUT/help.txt"

echo
echo "4) Running API scans"

APIS=(
"https://api.github.com"
"https://petstore.swagger.io/v2"
"https://api.spacexdata.com/v4"
)

for API in "${APIS[@]}"; do

    NAME=$(echo "$API" | sed 's#https\?://##; s#[^a-zA-Z0-9]#_#g')

    echo "Scanning $API"

    {
        echo "### SCAN $API"
        "$BIN" scan "$API"
    } > "$OUT/${NAME}_scan.txt" 2>&1 || true

    {
        echo "### MAP $API"
        "$BIN" map
    } > "$OUT/${NAME}_map.txt" 2>&1 || true

    {
        echo "### GRAPH $API"
        "$BIN" graph
    } > "$OUT/${NAME}_graph.txt" 2>&1 || true

done

echo
echo "5) Extracting nice snippets"

for f in "$OUT"/*_map.txt; do
    base=$(basename "$f")
    sed -n '1,80p' "$f" > "$OUT/snippet_$base"
done

echo
echo "6) Repo overview"

{
echo "FILES:"
git ls-files | head -n 200
echo
echo "TREE:"
ls -R | head -n 400
} > "$OUT/repo_overview.txt"

echo
echo "7) Packaging results"

tar -czf restless_hn_material.tar.gz "$OUT"

echo
echo "Done."
echo
echo "Archive created:"
echo "restless_hn_material.tar.gz"

echo
echo "Preview:"
ls -lh restless_hn_material.tar.gz
