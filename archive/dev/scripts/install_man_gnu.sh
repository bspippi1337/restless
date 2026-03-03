#!/usr/bin/env sh
set -eu

PREFIX="${PREFIX:-/usr/local}"
MANDIR="$PREFIX/share/man/man1"
DIST="dist/man"
BIN="./restless"

echo "== Restless manpage generator + installer =="

if [ ! -x "$BIN" ]; then
    echo "Building binary..."
    go build -o restless ./cmd/restless
fi

mkdir -p "$DIST"

echo "Generating man pages..."

"$BIN" --man > "$DIST/restless.1"
"$BIN" help openapi > "$DIST/restless-openapi.1"

# Ensure roff header exists
for f in "$DIST"/*.1; do
    if ! grep -q "^\.TH" "$f"; then
        tmp="$f.tmp"
        {
            echo ".TH RESTLESS 1"
            echo ".SH NAME"
            echo "restless - API probing and contract validation tool"
            echo ".SH DESCRIPTION"
            cat "$f"
        } > "$tmp"
        mv "$tmp" "$f"
    fi
done

echo "Installing to $MANDIR"

if [ ! -w "$MANDIR" ]; then
    echo "No write access to $MANDIR"
    echo "Re-run with sudo or set PREFIX=\$HOME/.local"
    exit 1
fi

mkdir -p "$MANDIR"
cp "$DIST"/*.1 "$MANDIR"

if command -v mandb >/dev/null 2>&1; then
    mandb >/dev/null 2>&1 || true
fi

echo "Installed:"
ls -1 "$MANDIR"/restless*.1

echo "Done."
