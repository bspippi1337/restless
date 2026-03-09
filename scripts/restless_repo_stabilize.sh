#!/usr/bin/env bash

# restless_repo_stabilize.sh
# Purpose:
# - back up the current dirty state
# - clean runtime artifacts and temporary pack directories
# - add a sane .gitignore
# - stage only the intentional Restless changes
# - optionally commit with --commit
#
# Usage:
#   bash restless_repo_stabilize.sh
#   bash restless_repo_stabilize.sh --commit

SCRIPT_NAME="$(basename "$0")"
REPO_ROOT="$(git rev-parse --show-toplevel 2>/dev/null)"
COMMIT_MODE=0
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
BACKUP_DIR=""
BACKUP_TGZ=""

for arg in "$@"; do
  case "$arg" in
    --commit) COMMIT_MODE=1 ;;
  esac
done

say() {
  printf '%s\n' "$*"
}

die() {
  printf 'ERROR: %s\n' "$*" >&2
  exit 1
}

require_repo() {
  [ -n "$REPO_ROOT" ] || die "Run this inside the restless git repository."
  cd "$REPO_ROOT" || die "Unable to cd into repo root."
}

make_backup() {
  mkdir -p ".backup" || die "Unable to create .backup directory."
  BACKUP_DIR=".backup/repo-stabilize-$TIMESTAMP"
  mkdir -p "$BACKUP_DIR" || die "Unable to create backup directory."

  say "== collecting changed files for backup =="
  git status --porcelain > "$BACKUP_DIR/status.txt"

  CHANGED_LIST="$BACKUP_DIR/changed-files.txt"
  : > "$CHANGED_LIST"

  while IFS= read -r line; do
    [ -n "$line" ] || continue
    path="${line:3}"
    printf '%s\n' "$path" >> "$CHANGED_LIST"
  done < "$BACKUP_DIR/status.txt"

  BACKUP_TGZ="$BACKUP_DIR/dirty-state.tgz"
  python3 - "$CHANGED_LIST" "$BACKUP_TGZ" <<'PY'
import os, sys, tarfile

changed_list = sys.argv[1]
backup_tgz = sys.argv[2]

paths = []
with open(changed_list, "r", encoding="utf-8") as f:
    for line in f:
        p = line.rstrip("\n")
        if p and os.path.lexists(p):
            paths.append(p)

with tarfile.open(backup_tgz, "w:gz") as tar:
    for p in paths:
        if os.path.lexists(p):
            tar.add(p)
PY

  [ -f "$BACKUP_TGZ" ] || die "Backup archive was not created."
  say "Backup archive: $BACKUP_TGZ"
}

write_gitignore() {
  say "== updating .gitignore =="
  touch .gitignore || die "Unable to create .gitignore."

  add_ignore() {
    pattern="$1"
    grep -Fqx "$pattern" .gitignore || printf '%s\n' "$pattern" >> .gitignore
  }

  add_ignore "build/"
  add_ignore "*.svg"
  add_ignore "*.dot"
  add_ignore "*.tmp"
  add_ignore "*.log"
  add_ignore "api.github.com.svg"
  add_ignore "api.github.com.svg.dot"
  add_ignore "api.dot"
  add_ignore "api.dot\\necho dot"
  add_ignore "blckswan_v8_fix/"
  add_ignore "restless_blckswan_engine/"
  add_ignore ".backup/"
}

clean_runtime_junk() {
  say "== removing runtime artifacts and temp packs =="
  rm -f "api.github.com.svg" "api.github.com.svg.dot" "api.dot" "api.dot
echo dot"
  rm -rf "blckswan_v8_fix" "restless_blckswan_engine"
}

show_engine_inventory() {
  say "== current engine files =="
  find internal/engine -maxdepth 1 -type f | sort
}

stage_intentional_changes() {
  say "== staging intentional changes =="

  # stage current tracked modifications/deletions
  git add -u -- cmd/restless internal/engine .gitignore

  # stage compatibility files if present
  [ -f internal/engine/compat.go ] && git add internal/engine/compat.go
  [ -f internal/engine/runtime_compat.go ] && git add internal/engine/runtime_compat.go

  # stage scripts you explicitly created, but only if they still exist
  [ -f scripts/engine_upgrade.sh ] && git add scripts/engine_upgrade.sh
  [ -f scripts/polish_engine.sh ] && git add scripts/polish_engine.sh
  [ -f scripts/upgrade_restless.sh ] && git add scripts/upgrade_restless.sh
}

sanity_report() {
  say
  say "== staged summary =="
  git diff --cached --stat

  say
  say "== remaining unstaged changes =="
  git status --short
}

maybe_commit() {
  [ "$COMMIT_MODE" -eq 1 ] || return 0

  say
  say "== creating commit =="
  git commit -m "blckswan engine: stabilize repo state, clean artifacts, add compatibility layer"
}

main() {
  require_repo

  say "== restless repo stabilizer =="
  say "Repo root: $REPO_ROOT"

  make_backup
  write_gitignore
  clean_runtime_junk
  show_engine_inventory
  stage_intentional_changes
  sanity_report
  maybe_commit

  say
  say "Done."
  say "Backup saved at: $BACKUP_TGZ"
  say "Next:"
  say "  make clean && make build"
  say "  ./build/restless api.github.com"
}

main
