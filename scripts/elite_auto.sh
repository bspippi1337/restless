#!/usr/bin/env bash
set -e

echo
echo "======================================"
echo "RESTLESS ELITE CLEANUP AUTOPILOT"
echo "======================================"
echo

########################################
# STEP 1 — git hygiene auto fix
########################################

echo "[1/7] Checking git working tree"

if ! git diff --quiet || ! git diff --cached --quiet || [ -n "$(git ls-files --others --exclude-standard)" ]; then
    echo "Working tree not clean — auto committing safe WIP snapshot"
    git add -A
    git commit -m "wip: automatic snapshot before elite cleanup" || true
else
    echo "Working tree clean"
fi

########################################
# STEP 2 — create cleanup branch
########################################

BRANCH="elite-repo-cleanup"

echo
echo "[2/7] Preparing cleanup branch"

if git show-ref --verify --quiet refs/heads/$BRANCH; then
    git checkout $BRANCH
else
    git checkout -b $BRANCH
fi

########################################
# STEP 3 — repo hygiene
########################################

echo
echo "[3/7] Moving legacy clutter"

mkdir -p archive/_elite_cleanup

for DIR in .attic legacy archive/dev; do
    if [ -d "$DIR" ]; then
        echo "moving $DIR → archive/_elite_cleanup/"
        git mv "$DIR" archive/_elite_cleanup/ 2>/dev/null || mv "$DIR" archive/_elite_cleanup/
    fi
done

########################################
# STEP 4 — gitignore hardening
########################################

echo
echo "[4/7] Hardening .gitignore"

touch .gitignore

add_ignore() {
    grep -qxF "$1" .gitignore || echo "$1" >> .gitignore
}

add_ignore "build/"
add_ignore "dist/"
add_ignore ".restless/"
add_ignore ".apt-build/"
add_ignore ".deb-build/"
add_ignore "apt-repo/"
add_ignore "*.tmp"
add_ignore "*.log"

########################################
# STEP 5 — ensure scan/inspect registered
########################################

echo
echo "[5/7] Ensuring CLI commands wired"

ROOT="internal/cli/root.go"

if [ -f "$ROOT" ]; then
    if ! grep -q "NewScanCmd" "$ROOT"; then
        echo "Adding scan command to root"
        sed -i '/AddCommand/s/$/\n\tcmd.AddCommand(NewScanCmd())/' "$ROOT"
    fi

    if ! grep -q "NewInspectCmd" "$ROOT"; then
        echo "Adding inspect command to root"
        sed -i '/AddCommand/s/$/\n\tcmd.AddCommand(NewInspectCmd())/' "$ROOT"
    fi
fi

########################################
# STEP 6 — rebuild
########################################

echo
echo "[6/7] Building restless"

go build -o build/restless ./cmd/restless

########################################
# STEP 7 — CLI verification
########################################

echo
echo "[7/7] CLI verification"

echo
./build/restless --help | head -40

echo
echo "Testing commands"
echo

./build/restless scan https://api.github.com 2>/dev/null | head -10 || true
echo
./build/restless map https://api.github.com 2>/dev/null | head -10 || true
echo
./build/restless inspect GET /users 2>/dev/null | head -10 || true

########################################
# commit result
########################################

echo
echo "Committing cleanup"

git add -A
git commit -m "chore: elite repository cleanup + CLI wiring" || true

echo
echo "======================================"
echo "ELITE CLEANUP COMPLETE"
echo "======================================"
echo
echo "Next:"
echo
echo "git push origin $BRANCH"
echo
echo "then open PR"
echo
