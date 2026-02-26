#!/usr/bin/env bash
set -euo pipefail

REPO_PATH="/home/pippi/dev/restless"
BRANCH="main"

echo "======================================"
echo "   RESTLESS PUSH LOCAL -> GITHUB"
echo "======================================"
echo
echo "Repo path: $REPO_PATH"
echo "Branch:    $BRANCH"
echo

cd "$REPO_PATH"

echo "==> 1. Status"
git status

echo
echo "==> 2. Add all files"
git add -A

echo
echo "==> 3. Commit (if needed)"
git commit -m "🔥 Sync: restore working local version as source of truth" || echo "No changes to commit"

echo
echo "==> 4. Ensure correct branch"
git checkout -B "$BRANCH"

echo
echo "==> 5. Push FORCE (this overwrites remote)"
read -p "About to FORCE PUSH to origin/$BRANCH. Continue? (yes/no) " CONFIRM
if [[ "$CONFIRM" != "yes" ]]; then
  echo "Aborted."
  exit 1
fi

git push origin "$BRANCH" --force

echo
echo "======================================"
echo "         DONE."
echo "GitHub now matches your local version."
echo "======================================"
