#!/usr/bin/env bash
set -euo pipefail

STRICT=0
if [[ "${1:-}" == "--strict" ]]; then
  STRICT=1
fi

# Colors (safe if no TTY)
if [[ -t 1 ]]; then
  RED=$'\033[31m'; GRN=$'\033[32m'; YEL=$'\033[33m'; BLU=$'\033[34m'; DIM=$'\033[2m'; RST=$'\033[0m'
else
  RED=""; GRN=""; YEL=""; BLU=""; DIM=""; RST=""
fi

PASS=0; WARN=0; FAIL=0
pass(){ PASS=$((PASS+1)); echo "${GRN}✔${RST} $*"; }
warn(){ WARN=$((WARN+1)); echo "${YEL}▲${RST} $*"; }
fail(){ FAIL=$((FAIL+1)); echo "${RED}✘${RST} $*"; }

need_repo_root() {
  [[ -d .git ]] || { fail "Not in repo root (missing .git)"; exit 2; }
  [[ -f go.mod ]] || { fail "Missing go.mod"; exit 2; }
}

run() { echo "${DIM}$*${RST}"; eval "$@"; }

check_git_clean() {
  local s
  s="$(git status --porcelain=v1 || true)"
  if [[ -z "$s" ]]; then
    pass "Git working tree clean"
  else
    warn "Git working tree has changes"
    echo "$s" | sed 's/^/  /'
  fi
}

check_upstream() {
  local head upstream ab
  head="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || true)"
  upstream="$(git rev-parse --abbrev-ref --symbolic-full-name @{u} 2>/dev/null || true)"
  if [[ -z "$upstream" ]]; then
    warn "No upstream tracking branch set for $head"
    return
  fi
  ab="$(git rev-list --left-right --count "$upstream"...HEAD 2>/dev/null || echo "0 0")"
  local behind ahead
  behind="$(echo "$ab" | awk '{print $1}')"
  ahead="$(echo "$ab" | awk '{print $2}')"
  if [[ "$ahead" == "0" && "$behind" == "0" ]]; then
    pass "Branch '$head' is up to date with '$upstream'"
  else
    warn "Branch '$head' diverged vs '$upstream' (ahead=$ahead behind=$behind)"
  fi
}

check_last_commit() {
  if git log -1 --oneline >/dev/null 2>&1; then
    pass "Last commit: $(git log -1 --oneline)"
  else
    fail "Cannot read git log"
  fi
}

check_tags() {
  local count
  count="$(git tag | wc -l | tr -d ' ')"
  if [[ "$count" -gt 0 ]]; then
    pass "Tags present ($count). Latest: $(git tag --sort=-creatordate | head -n 1)"
  else
    warn "No tags found"
  fi
}

check_workflows() {
  if [[ -d .github/workflows ]]; then
    local n
    n="$(ls -1 .github/workflows 2>/dev/null | wc -l | tr -d ' ')"
    if [[ "$n" -gt 0 ]]; then
      pass "GitHub workflows present ($n)"
      ls -1 .github/workflows | sed 's/^/  - /'
    else
      warn "No workflow files in .github/workflows"
    fi
  else
    warn "No .github/workflows directory"
  fi
}

check_go_env() {
  if command -v go >/dev/null 2>&1; then
    pass "Go: $(go version)"
  else
    fail "Go not found in PATH"
    return
  fi
  local gomod
  gomod="$(go env GOMOD 2>/dev/null || true)"
  if [[ "$gomod" == *"/go.mod" ]]; then
    pass "GOMOD: $gomod"
  else
    warn "GOMOD looks odd: $gomod"
  fi
}

check_build_test() {
  if go build ./cmd/restless >/dev/null 2>&1; then
    pass "go build ./cmd/restless OK"
  else
    fail "go build ./cmd/restless FAILED"
    return
  fi

  if go test ./... >/dev/null 2>&1; then
    pass "go test ./... OK"
  else
    fail "go test ./... FAILED"
  fi

  if go vet ./... >/dev/null 2>&1; then
    pass "go vet ./... OK"
  else
    warn "go vet ./... reported issues"
  fi
}

check_make_targets() {
  if [[ -f Makefile ]]; then
    pass "Makefile present"
    grep -E '^[a-zA-Z0-9_-]+:' Makefile | sed 's/:.*$//' | sed 's/^/  - /' || true
  else
    warn "No Makefile"
  fi
}

check_repro_flags() {
  # Correct grep order: -n before --
  local hits
  hits="$(grep -R -n -- "-trimpath" . 2>/dev/null || true)"

  if [[ -n "$hits" ]]; then
    pass "Repro flag '-trimpath' detected"
    echo "$hits" | head -n 10 | sed 's/^/  /'
    local make_has
    make_has="$(grep -R -n -- "BUILD_FLAGS := -trimpath" . 2>/dev/null || true)"
    if [[ -n "$make_has" ]]; then
      pass "BUILD_FLAGS includes -trimpath"
    else
      warn "Found -trimpath references, but not BUILD_FLAGS pattern (ok if handled elsewhere)"
    fi
  else
    warn "No '-trimpath' found anywhere"
  fi

  # Starship sanity: ensure release.yml exists if you expect CI releases
  if [[ -f .github/workflows/release.yml ]]; then
    pass "release.yml workflow present"
  else
    warn "No release.yml workflow"
  fi
}

check_gpg() {
  if command -v gpg >/dev/null 2>&1; then
    pass "GPG available"
    if gpg --list-secret-keys >/dev/null 2>&1; then
      pass "GPG secret keys accessible (signing possible)"
    else
      warn "GPG present, but no secret keys accessible (signing not configured)"
    fi
  else
    warn "GPG not available"
  fi
}

check_man_pages() {
  # Basic: if man pages installed, man -w should succeed.
  if command -v man >/dev/null 2>&1; then
    if man -w restless >/dev/null 2>&1; then
      pass "man page found: restless"
    else
      warn "man page not found: restless (ok if not installed on this machine)"
    fi
    if man -w restless-openapi >/dev/null 2>&1; then
      pass "man page found: restless-openapi"
    else
      warn "man page not found: restless-openapi"
    fi
  else
    warn "man command not available"
  fi
}

summary() {
  echo
  echo "${BLU}== Summary ==${RST}"
  echo "  PASS: $PASS"
  echo "  WARN: $WARN"
  echo "  FAIL: $FAIL"
  echo

  if [[ "$FAIL" -gt 0 ]]; then
    echo "${RED}Not ready.${RST} Fix FAIL items first."
    exit 1
  fi
  if [[ "$WARN" -gt 0 && "$STRICT" -eq 1 ]]; then
    echo "${YEL}Strict mode:${RST} warnings treated as failure."
    exit 2
  fi

  echo "${GRN}Starship-ready enough to proceed.${RST}"
  exit 0
}

main() {
  need_repo_root

  echo "== Starship Audit =="
  echo "Mode: $([[ "$STRICT" -eq 1 ]] && echo STRICT || echo NORMAL)"
  echo

  check_git_clean
  check_upstream
  check_last_commit
  check_tags
  check_workflows
  check_make_targets
  check_go_env
  check_build_test
  check_repro_flags
  check_gpg
  check_man_pages

  summary
}

main "$@"
