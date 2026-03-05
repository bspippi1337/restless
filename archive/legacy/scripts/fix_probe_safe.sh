#!/usr/bin/env bash
set -e

FILE="internal/ui/cli/probe.go"

if [ ! -f "$FILE" ]; then
  echo "File not found: $FILE"
  exit 1
fi

echo "→ Creating backup"
cp "$FILE" "$FILE.bak.$(date +%s)"

echo "→ Rewriting broken string sections safely"

awk '
BEGIN { skip=0 }
/func detectSpec\(/ { skip=1; print; next }
/^}/ && skip==1 { 
    print "        // quick fingerprint"
    print "        s := string(b)"
    print "        if strings.Contains(s, \"\\\"openapi\\\"\") || strings.Contains(s, \"\\\"swagger\\\"\") {"
    print "                return u, true"
    print "        }"
    print "}"
    skip=0
    next
}
/idx := strings.Index\(b,/ {
    print "        idx := strings.Index(b, \"\\\"paths\\\"\")"
    next
}
/strings.Split\(limit,/ {
    print "        for _, tok := range strings.Split(limit, \"\\\"\") {"
    next
}
{ print }
' "$FILE" > "$FILE.tmp"

mv "$FILE.tmp" "$FILE"

echo "→ Verifying build for this package"
if go test ./internal/ui/cli >/dev/null 2>&1; then
  echo "✓ probe.go fixed successfully"
else
  echo "⚠ Build still failing. Restore backup if needed:"
  echo "cp $FILE.bak.* $FILE"
  exit 1
fi

echo "Done."
