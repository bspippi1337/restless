#!/usr/bin/env bash
set +e

README="${1:-README.md}"
BIN="${BIN:-./build/restless}"
LOGDIR="${LOGDIR:-./readme-test-logs}"
TIMEOUT="${TIMEOUT:-20}"

mkdir -p "$LOGDIR"
TS="$(date +%Y%m%d-%H%M%S)"
COMMANDS="$LOGDIR/readme-commands-$TS.txt"
RESULTS="$LOGDIR/results-$TS.log"
SUMMARY="$LOGDIR/summary-$TS.txt"

[ -f "$README" ] || { echo "[!] Fant ikke $README"; exit 1; }

[ -x "$BIN" ] || make build

extract_commands() {
  {
    # 1. Kommandoer fra kodeblokker
    awk '
      /^```/ { fence=!fence; next }
      fence { print }
    ' "$README"

    # 2. Hele linjer som ser ut som kommandoer
    grep -E '(^|[[:space:]])(\$ |restless |\.\/build\/restless |go run \./cmd/restless|make )' "$README"

    # 3. Inline `restless ...`
    grep -oE '`[^`]*restless[^`]*`' "$README" | tr -d '`'
  } |
  sed \
    -e 's/^[[:space:]]*//' \
    -e 's/^\$[[:space:]]*//' \
    -e 's/^>[[:space:]]*//' \
    -e 's/#.*$//' \
    -e "s#^restless#$BIN#" \
    -e "s#^\./build/restless#$BIN#" |
  grep -Ev '^[[:space:]]*$' |
  grep -E '(^|[[:space:]])('"$(printf '%s' "$BIN" | sed 's/[.[\*^$()+?{}|]/\\&/g')"')([[:space:]]|$)|^go run \./cmd/restless|^make([[:space:]]|$)' |
  grep -Ev '(install|docker|release|sudo|rm -rf|mkfs|dd |reboot|shutdown|curl.*\|.*sh|wget.*\|.*sh)' |
  awk '!seen[$0]++'
}

extract_commands > "$COMMANDS"

pass=0
fail=0
skip=0
total=0

echo "[*] Fant $(wc -l < "$COMMANDS") aktuelle kommandoer"
echo "[*] Kommandoer: $COMMANDS"
echo

: > "$RESULTS"

while IFS= read -r cmd; do
  total=$((total+1))
  echo "================================================================" | tee -a "$RESULTS"
  echo "[$total] $cmd" | tee -a "$RESULTS"

  out="$LOGDIR/cmd-$TS-$total.out"
  err="$LOGDIR/cmd-$TS-$total.err"

  timeout "$TIMEOUT" bash -lc "$cmd" >"$out" 2>"$err"
  rc=$?

  if [ "$rc" -eq 0 ]; then
    echo "[PASS] rc=0" | tee -a "$RESULTS"
    pass=$((pass+1))
  elif [ "$rc" -eq 124 ]; then
    echo "[FAIL] timeout ${TIMEOUT}s" | tee -a "$RESULTS"
    fail=$((fail+1))
  else
    echo "[FAIL] rc=$rc" | tee -a "$RESULTS"
    fail=$((fail+1))
  fi

  echo "--- stdout ---" >> "$RESULTS"
  sed -n '1,120p' "$out" >> "$RESULTS"
  echo "--- stderr ---" >> "$RESULTS"
  sed -n '1,120p' "$err" >> "$RESULTS"
  echo >> "$RESULTS"
done < "$COMMANDS"

{
  echo "README unicorn summary"
  echo "======================"
  echo "Total: $total"
  echo "Pass:  $pass"
  echo "Fail:  $fail"
  echo "Skip:  $skip"
  echo
  echo "Commands: $COMMANDS"
  echo "Results:  $RESULTS"
} | tee "$SUMMARY"

echo
echo "[*] Ferdig:"
echo "    $SUMMARY"
echo "    $RESULTS"

[ "$fail" -eq 0 ]
