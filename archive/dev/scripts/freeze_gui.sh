#!/usr/bin/env bash
set -e

echo "=== Restless GUI Freeze Script ==="
echo

# 1. Sjekk at vi er i repo-root
if [ ! -f "go.mod" ]; then
  echo "Error: go.mod not found. Run this from repo root."
  exit 1
fi

# 2. Flytt GUI hvis den finnes
if [ -d "internal/gui" ]; then
  echo "→ Moving internal/gui to experiments/gui"
  mkdir -p experiments
  mv internal/gui experiments/gui
else
  echo "→ No internal/gui found (already moved?)"
fi

# 3. Fjern Fyne dependency fra go.mod
echo "→ Removing fyne dependency"
go mod edit -droprequire fyne.io/fyne/v2 2>/dev/null || true
go mod edit -droprequire fyne.io/systray 2>/dev/null || true

# 4. Rydd opp moduler
echo "→ Running go mod tidy"
go mod tidy

# 5. Sjekk at ingen filer fortsatt importerer fyne
echo "→ Checking for remaining fyne imports"
if grep -R "fyne.io/fyne" . --exclude-dir=experiments --exclude-dir=.git; then
  echo
  echo "WARNING: Remaining fyne imports detected above."
  echo "You must remove or refactor them."
  exit 1
else
  echo "✓ No active fyne imports found."
fi

# 6. Test build (uten GUI)
echo
echo "→ Testing build"
if go build ./cmd/restless; then
  echo "✓ CLI build successful"
else
  echo "Build failed. Investigate before committing."
  exit 1
fi

echo
echo "=== GUI successfully frozen ==="
echo "You can now commit:"
echo
echo "git add -A"
echo 'git commit -m "chore: freeze GUI layer to stabilize core"'
