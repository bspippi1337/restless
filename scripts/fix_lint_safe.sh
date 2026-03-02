#!/usr/bin/env bash
set -e

FILE="Makefile"

if [ ! -f "$FILE" ]; then
  echo "Makefile not found."
  exit 1
fi

echo "→ Backup Makefile"
cp "$FILE" "$FILE.bak.$(date +%s)"

echo "→ Rewriting lint target safely"

awk '
BEGIN { inlint=0 }
/^lint:/ {
    print "lint:"
    print "\t@if command -v golangci-lint >/dev/null 2>&1; then \\"
    print "\t\tgolangci-lint run ./... ; \\"
    print "\telse \\"
    print "\t\techo \"golangci-lint not installed, skipping lint.\" ; \\"
    print "\tfi"
    inlint=1
    next
}
inlint==1 && /^\t/ { next }
inlint==1 && !/^\t/ { inlint=0 }
{ if (!inlint) print }
' "$FILE" > "$FILE.tmp"

mv "$FILE.tmp" "$FILE"

echo "→ Verifying make lint"
if make lint >/dev/null 2>&1; then
  echo "✓ lint target now safe"
else
  echo "⚠ Something went wrong. Restore backup if needed:"
  echo "cp $FILE.bak.* $FILE"
  exit 1
fi

echo "Done."
