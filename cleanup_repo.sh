#!/usr/bin/env bash
set -euo pipefail

echo "🛠️ Start cleanup for Restless v2..."

# 1) Safety check
if [ ! -d ".git" ]; then
  echo "❌ Not a git repo"
  exit 1
fi

# 2) Ensure clean working tree
git status --porcelain | grep . && {
  echo "❌ Working tree not clean. Commit or stash first."
  exit 1
}

# 3) Normalize main branch
git checkout main || git checkout -b main

# 4) Allowed core folders
CORE_FOLDERS=("cmd" "internal" "go.mod" "go.sum" ".github" "CHANGELOG.md" "CONTRIBUTING.md" "LICENSE")

echo "🧹 Archiving non-core folders..."
for d in landing web lazyfyne restless-gui examples install docs brand-guidelines.pdf; do
  if [ -e "$d" ]; then
    mkdir -p legacy
    git mv "$d" legacy/
    echo "➡️ Moved $d → legacy/"
  fi
done

# 5) Updated .gitignore
cat > .gitignore <<EOF
# Build outputs
bin/
dist/
build/

# Go
*.test
*.exe
*.out

# OS
.DS_Store
Thumbs.db

# IDE
.vscode/
.idea/

# Snapshots & history
.snapshots/
.history/

# Legacy
legacy/
EOF

echo "📝 .gitignore updated."

# 6) Update README.md to v2 content
cat > README.md <<'EOF'
# Restless v2 ⚡

Terminal-first API Workbench.

## Install

```bash
go install github.com/bspippi1337/restless/cmd/restless@latest
