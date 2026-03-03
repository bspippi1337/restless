#!/usr/bin/env sh
set -eu

echo "== Restless GNU CLI repair =="

FILE="cmd/restless/main.go"

if [ ! -f "$FILE" ]; then
  echo "Error: run from repo root."
  exit 1
fi

if grep -q 'case "openapi"' "$FILE"; then
  echo "openapi routing already present."
  exit 0
fi

BACKUP="$FILE.bak.$(date +%s)"
cp "$FILE" "$BACKUP"
echo "Backup: $BACKUP"

TMP="$FILE.tmp"

# Insert openapi case right after switch args[0] {
sed '/switch args\[0\] {/a\
\
    case "openapi":\
        if err := entry.OpenAPI(args[1:]); err != nil {\
            fmt.Println("openapi error:", err)\
            os.Exit(1)\
        }\
' "$FILE" > "$TMP"

mv "$TMP" "$FILE"

echo "Running go fmt..."
go fmt ./...

echo "Building..."
if ! go build ./cmd/restless; then
  echo "Build failed."
  echo "Restoring backup..."
  mv "$BACKUP" "$FILE"
  exit 1
fi

echo "Running tests..."
if ! go test ./...; then
  echo "Tests failed (non-fatal)."
fi

printf "Commit changes? [y/N]: "
read ans

case "$ans" in
  y|Y)
    git add "$FILE"
    git commit -m "fix(cli): add openapi routing"
    echo "Committed."
    ;;
  *)
    echo "No commit performed."
    ;;
esac

echo "Done."
