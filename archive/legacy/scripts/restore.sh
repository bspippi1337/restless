#!/usr/bin/env sh
set -eu

echo "== GNU help/man routing repair (robust edition) =="

FILE="cmd/restless/main.go"

[ -f "$FILE" ] || { echo "Run from repo root"; exit 1; }

backup=".help_patch_backup_$(date +%s)"
cp "$FILE" "$backup"
echo "Backup saved: $backup"

# Replace entire main() safely
awk '
BEGIN { skip=0 }
/^func main\(\)/ {
    skip=1
    print "func main() {"
    print "    args := os.Args[1:]"
    print ""
    print "    if len(args) == 0 {"
    print "        printHelp()"
    print "        return"
    print "    }"
    print ""
    print "    // global help"
    print "    if args[0] == \"help\" {"
    print "        _ = entry.Help(args[1:])"
    print "        return"
    print "    }"
    print ""
    print "    for _, a := range args {"
    print "        if a == \"--man\" {"
    print "            _ = entry.Man(args)"
    print "            return"
    print "        }"
    print "        if a == \"--help\" || a == \"-h\" {"
    print "            _ = entry.Help(args)"
    print "            return"
    print "        }"
    print "    }"
    print ""
    print "    switch args[0] {"
    print "    case \"openapi\":"
    print "        if len(args) == 1 {"
    print "            _ = entry.OpenAPI([]string{})"
    print "            return"
    print "        }"
    print "        if err := entry.OpenAPI(args[1:]); err != nil {"
    print "            fmt.Println(\"openapi error:\", err)"
    print "            os.Exit(1)"
    print "        }"
    print "    case \"smart\", \"simulate\":"
    print "        if err := entry.Smart(args[1:]); err != nil {"
    print "            os.Exit(1)"
    print "        }"
    print "    case \"probe\":"
    print "        if err := entry.Normal(args[1:]); err != nil {"
    print "            os.Exit(1)"
    print "        }"
    print "    default:"
    print "        if err := entry.Normal(args); err != nil {"
    print "            os.Exit(1)"
    print "        }"
    print "    }"
    print "}"
    next
}
skip && /^}/ { skip=0; next }
skip { next }
{ print }
' "$FILE" > "$FILE.tmp"

mv "$FILE.tmp" "$FILE"

echo "Formatting..."
go fmt ./...

echo "Building..."
if ! go build ./cmd/restless; then
    echo "Build failed. Restoring backup."
    mv "$backup" "$FILE"
    exit 1
fi

echo "Running tests..."
go test ./... || echo "WARN: tests failed"

printf "Commit changes? [y/N]: "
read ans
case "$ans" in
  y|Y)
    git add "$FILE"
    git commit -m "feat(cli): full GNU help/man routing"
    echo "Committed."
    ;;
  *)
    echo "No commit."
    echo "Restore with: cp $backup $FILE"
    ;;
esac

echo "Done."
