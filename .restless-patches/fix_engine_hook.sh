#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail

TARGET=""

for candidate in \
  rootCmd \
  RootCmd \
  cliCmd \
  appCmd \
  mainCmd \
  restlessCmd
do
  if grep -R "${candidate}[[:space:]]*:=[[:space:]]*&cobra.Command" cmd internal . >/dev/null 2>&1; then
    TARGET="$candidate"
    break
  fi

  if grep -R "var[[:space:]]\+${candidate}[[:space:]]*=[[:space:]]*&cobra.Command" cmd internal . >/dev/null 2>&1; then
    TARGET="$candidate"
    break
  fi
done

if [ -z "$TARGET" ]; then
  echo "[!] Fant ikke command root automatisk."
  echo
  echo "Kjør:"
  echo "  grep -R 'cobra.Command' cmd internal"
  exit 1
fi

echo "[+] Fant command root: $TARGET"

sed -i \
  "s/rootCmd.AddCommand/${TARGET}.AddCommand/g" \
  cmd/restless/engine_wow.go

echo "[+] Patched engine_wow.go"

echo
echo "[+] Rebuild:"
echo "    go build -o build/restless ./cmd/restless"
