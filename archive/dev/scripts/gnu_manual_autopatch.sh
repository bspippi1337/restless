#!/usr/bin/env sh
set -eu

echo "== GNU manual pack patch (safe, backup, prompt commit) =="

need() { [ -e "$1" ] || { echo "Error: missing $1 (run from repo root)"; exit 1; }; }

need cmd/restless/main.go
need go.mod

ts="$(date +%s)"
backup_dir=".patch_backup_$ts"
mkdir -p "$backup_dir"
cp cmd/restless/main.go "$backup_dir/main.go"

echo "Backup saved in $backup_dir/"

# Ensure openapi dispatch exists + fallback to help when args missing
if ! grep -q 'case "openapi"' cmd/restless/main.go; then
  echo "Patching cmd/restless/main.go: add openapi routing..."
  tmp="cmd/restless/main.go.tmp"
  sed '/switch args\[0\] {/a\

\tcase "openapi":\
\t\tif len(args) == 1 {\
\t\t\t_ = entry.OpenAPI([]string{})\
\t\t\treturn\
\t\t}\
\t\tif err := entry.OpenAPI(args[1:]); err != nil {\
\t\t\tfmt.Println("openapi error:", err)\
\t\t\tos.Exit(1)\
\t\t}\
' cmd/restless/main.go > "$tmp"
  mv "$tmp" cmd/restless/main.go
else
  echo "openapi routing already present; ensuring fallback..."
  if ! grep -q 'len(args) == 1' cmd/restless/main.go; then
    tmp="cmd/restless/main.go.tmp"
    awk '
      {print}
      $0 ~ /case "openapi":/ && !done {
        print "        if len(args) == 1 {"
        print "            _ = entry.OpenAPI([]string{})"
        print "            return"
        print "        }"
        done=1
      }
    ' cmd/restless/main.go > "$tmp"
    mv "$tmp" cmd/restless/main.go
  fi
fi

echo "Formatting..."
go fmt ./... >/dev/null

echo "Build..."
go build ./cmd/restless

echo "Tests..."
go test ./... || echo "WARN: tests failed"

echo "Generate man pages..."
chmod +x scripts/build-manpages.sh 2>/dev/null || true
sh scripts/build-manpages.sh

printf "Commit changes? [y/N]: "
read ans
case "$ans" in
  y|Y)
    git add -A
    git commit -m "docs(manual): embed manuals + generate roff man pages; fix openapi help fallback"
    echo "Committed."
    ;;
  *)
    echo "No commit performed."
    echo "To rollback: cp $backup_dir/main.go cmd/restless/main.go"
    ;;
esac
