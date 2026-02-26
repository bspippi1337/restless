#!/usr/bin/env bash
set -euo pipefail

echo "🧹 Restless repo deep clean starting..."

############################################
# Safety check
############################################
if [ ! -d ".git" ]; then
  echo "❌ Run this inside the repo root"
  exit 1
fi

git pull --rebase || true

############################################
# Create new structure
############################################
echo "📁 Creating clean folder structure..."
mkdir -p packaging/{debian,termux}
mkdir -p web
mkdir -p docs
mkdir -p dist
mkdir -p assets/screenshots || true

############################################
# Move website
############################################
if [ -f "index.html" ]; then
  echo "🌐 Moving website → /web"
  git mv index.html web/index.html 2>/dev/null || mv index.html web/
fi

############################################
# Move manual
############################################
if [ -f "MANUAL.md" ]; then
  echo "📚 Moving manual → /docs"
  git mv MANUAL.md docs/manual.md 2>/dev/null || mv MANUAL.md docs/manual.md
fi

############################################
# Move Termux mess
############################################
if [ -d "data/data/com.termux/files/usr/bin" ]; then
  echo "📦 Moving Termux packaging → /packaging/termux"
  mv data/data/com.termux/files/usr/bin/* packaging/termux/ 2>/dev/null || true
  rm -rf data
fi

############################################
# Remove compiled binaries
############################################
echo "💣 Removing binaries from git..."
rm -f restless || true
git rm -f restless 2>/dev/null || true

############################################
# Remove backup files
############################################
echo "🗑 Removing backup files..."
find . -type f -name "*.bak*" -delete

############################################
# Create proper .gitignore
############################################
echo "🧾 Writing proper .gitignore..."
cat > .gitignore << 'EOF'
# Binaries
/dist/
/build/
/bin/
restless

# Logs
*.log

# OS junk
.DS_Store
Thumbs.db

# Editors
.vscode/
.idea/

# Backup files
*.bak*
*~

# Packages / releases
*.deb
*.apk
*.tar.gz
*.zip

# Termux runtime junk
data/
EOF

############################################
# Create repo meta docs
############################################
echo "📄 Creating repo docs..."

cat > CONTRIBUTING.md << 'EOF'
# Contributing

PRs welcome. Keep it simple. Keep it fast. Keep it terminal-first.
EOF

cat > CHANGELOG.md << 'EOF'
# Changelog
All notable changes will be documented here.
EOF

cat > ARCHITECTURE.md << 'EOF'
# Architecture

cmd/restless -> CLI entrypoint
internal/     -> core logic
packaging/    -> distribution builds
web/          -> github pages demo
docs/         -> manuals and guides
EOF

cat > ROADMAP.md << 'EOF'
# Roadmap

- TUI autocomplete
- OpenAPI auto-import
- Plugin system
- Homebrew package
EOF

############################################
# Stage everything
############################################
echo "📦 Staging changes..."
git add -A

############################################
# Commit
############################################
git commit -m "repo: massive cleanup and restructure" || true

############################################
# Push
############################################
echo "🚀 Pushing..."
git push

echo "✅ Repo cleanup complete!"
echo "Your repository is now civilized."
