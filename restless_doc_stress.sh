#!/usr/bin/env bash
set -u

ROOT="$(pwd)"
TS="$(date +%Y%m%d-%H%M%S)"
OUT="$ROOT/doc-stress-$TS"
BIN="${BIN:-$ROOT/build/restless}"
TIMEOUT="${TIMEOUT:-20}"

mkdir -p "$OUT"

log(){ printf '[*] %s\n' "$*"; }
warn(){ printf '[!] %s\n' "$*" | tee -a "$OUT/warnings.log" >&2; }

log "Restless public-doc stress"
git rev-parse --short HEAD > "$OUT/commit.txt" 2>/dev/null || echo "no-git" > "$OUT/commit.txt"
log "Commit: $(cat "$OUT/commit.txt")"

if [ ! -x "$BIN" ]; then
  log "Bygger binary..."
  mkdir -p build
  if ! go build -o "$BIN" ./cmd/restless >"$OUT/build.log" 2>&1; then
    warn "Build feilet. Se $OUT/build.log"
    exit 1
  fi
fi

"$BIN" --help > "$OUT/root-help.txt" 2>&1 || true
"$BIN" help > "$OUT/help.txt" 2>&1 || true
"$BIN" version > "$OUT/version.txt" 2>&1 || true

log "Finner publiserte docs, ekskluderer archive/legacy/scratch/dev-artifacts..."
find . \
  -path './.git' -prune -o \
  -path './archive' -prune -o \
  -path './.attic' -prune -o \
  -path './scratch' -prune -o \
  -path './_dev' -prune -o \
  -path './test_logs' -prune -o \
  -path './doc-stress-*' -prune -o \
  -path './vendor' -prune -o \
  -path './node_modules' -prune -o \
  -type f \( -name '*.md' -o -name '*.txt' -o -name '*.rst' -o -name '*.adoc' \) \
  -print | sort > "$OUT/docs.txt"

python3 - "$OUT/docs.txt" "$OUT/raw-commands.txt" <<'PY'
import re, sys, pathlib

docs_file, out_file = sys.argv[1], sys.argv[2]
docs = [pathlib.Path(x.strip()) for x in open(docs_file) if x.strip()]
cmds, seen = [], set()

def add(path, line_no, cmd):
    cmd = cmd.strip()
    if not cmd:
        return
    cmd = re.sub(r'^\s*(\$|#|❯|➜|>)\s*', '', cmd).strip()
    cmd = cmd.replace('./build/restless ', 'restless ')
    cmd = cmd.replace('./restless ', 'restless ')
    cmd = cmd.replace('go run ./cmd/restless ', 'restless ')

    if not cmd.startswith("restless "):
        return

    # Ikke kjør placeholders som shell-redirection.
    replacements = {
        "<id>": "demo",
        "<ID>": "demo",
        "<url>": "https://example.com",
        "<URL>": "https://example.com",
        "<target>": "https://example.com",
        "<host>": "example.com",
        "<path>": ".",
        "<file>": "README.md",
        "<command>": "true",
    }
    for k, v in replacements.items():
        cmd = cmd.replace(k, v)

    deny = [
        "restless openapi ",   # legacy only unless implemented later
        "restless profile ",   # legacy only unless implemented later
        "restless prefs ",     # legacy only unless implemented later
        "restless session",    # legacy only unless implemented later
    ]
    if any(cmd.startswith(x) for x in deny):
        return

    dangerous = [" rm ", " rm -", " mkfs", " dd ", " fastboot ", " adb ", " heimdall "]
    if any(x in f" {cmd} " for x in dangerous):
        return

    key = (str(path), line_no, cmd)
    if key not in seen:
        seen.add(key)
        cmds.append(key)

for p in docs:
    try:
        text = p.read_text(errors="ignore")
    except Exception:
        continue

    for m in re.finditer(r"```[a-zA-Z0-9_-]*\n(.*?)```", text, re.S):
        start = text[:m.start()].count("\n") + 1
        for i, line in enumerate(m.group(1).splitlines(), start):
            add(p, i, line)

    for ln, line in enumerate(text.splitlines(), 1):
        for m in re.finditer(r"`([^`\n]*restless [^`\n]*)`", line):
            add(p, ln, m.group(1))
        add(p, ln, line)

with open(out_file, "w") as f:
    for path, line, cmd in cmds:
        f.write(f"{path}:{line}\t{cmd}\n")
PY

log "Henter help-kommandoer fra CLI..."
{
  "$BIN" --help 2>/dev/null || true
  "$BIN" help 2>/dev/null || true
} | awk '
  /^[[:space:]]+[a-zA-Z0-9][a-zA-Z0-9_-]+[[:space:]]/ {
    cmd=$1
    if (cmd !~ /^(help|completion)$/) print "CLI_HELP\trestless " cmd " --help"
  }
' | sort -u > "$OUT/help-commands.txt"

cat "$OUT/raw-commands.txt" "$OUT/help-commands.txt" \
  | awk -F '\t' 'NF==2 && !seen[$2]++ {print}' \
  > "$OUT/test-commands.txt"

TOTAL="$(wc -l < "$OUT/test-commands.txt" | tr -d ' ')"
log "Fant $TOTAL testbare kommandoer"

PASS=0
FAIL=0
: > "$OUT/results.tsv"
: > "$OUT/failures.md"

while IFS=$'\t' read -r SRC CMD; do
  [ -n "${CMD:-}" ] || continue
  ID="$(printf '%s' "$CMD" | sha1sum | awk '{print substr($1,1,10)}')"
  LOG="$OUT/cmd-$ID.log"
  TESTCMD="${CMD/restless/$BIN}"

  set +e
  timeout "$TIMEOUT" bash -lc "$TESTCMD" >"$LOG" 2>&1
  RC=$?
  set -e

  if [ "$RC" -eq 0 ]; then
    STATUS="PASS"; PASS=$((PASS+1))
  elif [ "$RC" -eq 124 ]; then
    STATUS="TIMEOUT"; FAIL=$((FAIL+1))
  else
    STATUS="FAIL"; FAIL=$((FAIL+1))
  fi

  printf '%s\t%s\t%s\t%s\n' "$STATUS" "$RC" "$SRC" "$CMD" >> "$OUT/results.tsv"

  if [ "$STATUS" != "PASS" ]; then
    {
      printf '\n## %s rc=%s\n\n' "$STATUS" "$RC"
      printf 'Source: `%s`\n\n' "$SRC"
      printf 'Command:\n\n```bash\n%s\n```\n\n' "$CMD"
      printf 'Output:\n\n```text\n'
      tail -120 "$LOG"
      printf '\n```\n'
    } >> "$OUT/failures.md"
  fi
done < "$OUT/test-commands.txt"

cat > "$OUT/summary.md" <<EOF2
# Restless public-doc stress report

Commit: \`$(cat "$OUT/commit.txt")\`

- Testbare kommandoer: $TOTAL
- PASS: $PASS
- FAIL/TIMEOUT: $FAIL

Legacy docs under \`archive/\` are excluded from public contract tests.
EOF2

cat "$OUT/summary.md"
