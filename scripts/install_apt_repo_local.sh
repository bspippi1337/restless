#!/usr/bin/env bash
set -euo pipefail

# Helper: serve the repo locally for quick tests
# Usage: ./scripts/install_apt_repo_local.sh
# Then in another shell: python3 -m http.server 8000 --directory .apt-repo
echo "To test locally:"
echo "  make aptrepo"
echo "  python3 -m http.server 8000 --directory .apt-repo"
echo "Then add to /etc/apt/sources.list.d/restless.list:"
echo "  deb [trusted=yes] http://127.0.0.1:8000 stable main"
