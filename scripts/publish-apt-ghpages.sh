#!/usr/bin/env bash
# scripts/publish-apt-ghpages.sh
set -euo pipefail

need() { command -v "$1" >/dev/null 2>&1 || { echo "Missing dependency: $1" >&2; exit 2; }; }

need reprepro
need git

if [[ "$#" -lt 1 ]]; then
  echo "Usage: $0 path/to/*.deb" >&2
  exit 2
fi

APT_REPO_DIR="${APT_REPO_DIR:-apt-repo}"
APT_ORIGIN="${APT_ORIGIN:-YourProject}"
APT_LABEL="${APT_LABEL:-YourProject APT}"
APT_CODENAME="${APT_CODENAME:-stable}"
APT_ARCHS="${APT_ARCHS:-amd64 arm64 all}"
APT_COMPONENT="${APT_COMPONENT:-main}"
APT_DESCRIPTION="${APT_DESCRIPTION:-APT repo for ${APT_ORIGIN}}"
APT_SIGN_KEY="${APT_SIGN_KEY:-}"   # optional
APT_BRANCH="${APT_BRANCH:-gh-pages}"
APT_PUSH="${APT_PUSH:-1}"

mkdir -p "${APT_REPO_DIR}/conf"

cat > "${APT_REPO_DIR}/conf/distributions" <<EOF
Origin: ${APT_ORIGIN}
Label: ${APT_LABEL}
Codename: ${APT_CODENAME}
Architectures: ${APT_ARCHS}
Components: ${APT_COMPONENT}
Description: ${APT_DESCRIPTION}
EOF

if [[ -n "${APT_SIGN_KEY}" ]]; then
  need gpg
  echo "SignWith: ${APT_SIGN_KEY}" >> "${APT_REPO_DIR}/conf/distributions"
fi

cat > "${APT_REPO_DIR}/conf/options" <<EOF
verbose
basedir ${APT_REPO_DIR}
EOF

echo "==> Including .deb packages into APT repo..."
for deb in "$@"; do
  [[ -f "$deb" ]] || { echo "Not found: $deb" >&2; exit 2; }
  echo "   + $deb"
  reprepro -b "${APT_REPO_DIR}" includedeb "${APT_CODENAME}" "$deb"
done

touch "${APT_REPO_DIR}/.nojekyll"

echo "==> Repo updated at ${APT_REPO_DIR}/ (dists/ + pool/)"

if [[ "${APT_PUSH}" != "1" ]]; then
  echo "==> APT_PUSH=0, skipping push."
  exit 0
fi

echo "==> Publishing to branch '${APT_BRANCH}'..."

origin_url="$(git remote get-url origin)"
tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

(
  cd "$tmpdir"
  git init -q
  git remote add origin "$origin_url"

  if git ls-remote --exit-code --heads origin "${APT_BRANCH}" >/dev/null 2>&1; then
    git fetch -q origin "${APT_BRANCH}"
    git checkout -q -B "${APT_BRANCH}" "origin/${APT_BRANCH}"
  else
    git checkout -q --orphan "${APT_BRANCH}"
    rm -rf ./* 2>/dev/null || true
  fi

  rm -rf ./* 2>/dev/null || true
  cp -a "../${APT_REPO_DIR}/." .

  git add -A
  if git diff --cached --quiet; then
    echo "==> Nothing new to publish."
    exit 0
  fi

  git commit -q -m "Publish APT repo (${APT_CODENAME})"
  git push -q origin "${APT_BRANCH}:${APT_BRANCH}"
)

echo "✅ Published to '${APT_BRANCH}'."
