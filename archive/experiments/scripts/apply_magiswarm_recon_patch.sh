
#!/usr/bin/env bash
set -euo pipefail

echo "Applying magiswarm recon patch..."

FILE="internal/cli/root.go"
if [ ! -f "$FILE" ]; then
  echo "ERROR: $FILE not found"
  exit 1
fi

# Remove duplicates if any
sed -i '/NewMagiswarmCmd()/d' "$FILE"

# Insert once before first 'return cmd'
perl -0777 -i -pe 's/\n(\s*)return cmd/\n\1cmd.AddCommand(NewMagiswarmCmd())\n\1return cmd/' "$FILE"

gofmt -w internal/cli internal/core >/dev/null 2>&1 || true

echo "Done."
echo "Rebuild:"
echo "  make build"
echo "Run:"
echo "  ./build/restless magiswarm https://api.github.com"
