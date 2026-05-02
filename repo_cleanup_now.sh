#!/usr/bin/env bash
set -euo pipefail

TS="$(date +%Y%m%d-%H%M%S)"
ATTIC="_attic/cleanup-$TS"

mkdir -p "$ATTIC"

say(){ printf '[*] %s\n' "$*"; }

stash_file(){
  local f="$1"
  [ -e "$f" ] || return 0
  mkdir -p "$ATTIC/$(dirname "$f")"
  git ls-files --error-unmatch "$f" >/dev/null 2>&1 && git rm -q -- "$f" || mv -- "$f" "$ATTIC/$f"
}

say "Rydder repo-rot. Flytter søppel til $ATTIC"

# 1. Fjern genererte SVG/DOT/scans fra rot
find . -maxdepth 1 -type f \( \
  -name '*.svg' -o \
  -name '*.svg.dot' -o \
  -name '*.dot' \
\) -print0 | while IFS= read -r -d '' f; do
  stash_file "${f#./}"
done

# 2. Fjern kjente test/output artefakter fra rot
for f in \
  resultat.txt \
  current.json \
  order.json \
  request.har \
  SHA \
  MAKEFILE_APPEND.mk \
  Makefile.deb_patch \
  fix_restless_doc_contract.sh \
  restless_doc_stress.sh \
  restless_readme_unicorn_test.sh
do
  stash_file "$f"
done

# 3. Fjern genererte testmapper
for d in doc-stress-* readme-test-logs test_logs; do
  for x in $d; do
    [ -e "$x" ] || continue
    git ls-files --error-unmatch "$x" >/dev/null 2>&1 && git rm -rq -- "$x" || mv -- "$x" "$ATTIC/$x"
  done
done

# 4. CHANGELOG: behold Markdown, arkiver ChangeLog hvis begge finnes
if [ -f CHANGELOG.md ] && [ -f ChangeLog ]; then
  stash_file "ChangeLog"
fi

# 5. Behold ekte assets, men flytt løse grafikk-output fra assets hvis åpenbart generert
find assets -type f \( -name '*.svg.dot' -o -name '*.tmp' -o -name '*.bak' \) -print0 2>/dev/null | while IFS= read -r -d '' f; do
  stash_file "$f"
done

# 6. Sørg for at fremtidig søppel ikke kommer tilbake
touch .gitignore
add_ignore(){
  grep -qxF "$1" .gitignore || echo "$1" >> .gitignore
}

add_ignore '_attic/'
add_ignore 'doc-stress-*/'
add_ignore 'readme-test-logs/'
add_ignore 'test_logs/'
add_ignore '*.svg.dot'
add_ignore '*.dot'
add_ignore 'current.json'
add_ignore 'order.json'
add_ignore 'request.har'
add_ignore 'resultat.txt'
add_ignore 'MAKEFILE_APPEND.mk'
add_ignore 'Makefile.deb_patch'
add_ignore 'fix_restless_doc_contract.sh'
add_ignore 'restless_doc_stress.sh'
add_ignore 'restless_readme_unicorn_test.sh'

# 7. Lag profesjonell rotliste
cat > docs/REPO_HYGIENE.md <<'MD'
# Repository hygiene

The repository root should contain source, packaging, documentation, and project metadata only.

Generated files belong outside git, especially:

- scan output
- `.svg`, `.dot`, and `.svg.dot` API maps
- temporary JSON/HAR files
- local test reports
- one-off patch scripts
- local stress-test logs

Use `examples/`, `docs/`, `assets/`, `scripts/`, `tools/`, `cmd/`, `internal/`, and `pkg/` for intentional project files.
MD

say "Formatter Go"
go fmt ./... >/dev/null || true

say "Status etter rydding:"
git status --short

echo
say "Neste steg hvis dette ser riktig ut:"
echo "git add -A"
echo "git commit -m 'chore: clean generated artifacts from repository root'"
