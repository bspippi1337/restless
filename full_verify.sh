cat > full_verify.sh <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

APT_URL="https://bspippi1337.github.io/restless"
GPG_URL="${APT_URL}/restless.gpg"
BREW_TAP="bspippi1337/restless"
APP="restless"

echo "========================================="
echo "🐳 Testing APT in clean Debian container"
echo "========================================="

docker run --rm -i debian:stable-slim bash <<APT_TEST
set -e
apt update >/dev/null
apt install -y curl gnupg ca-certificates >/dev/null

curl -fsSL ${GPG_URL} | gpg --dearmor -o /usr/share/keyrings/restless.gpg
echo "deb [signed-by=/usr/share/keyrings/restless.gpg] ${APT_URL} stable main" > /etc/apt/sources.list.d/restless.list

apt update >/dev/null
apt install -y ${APP}

${APP} --version || true

echo "✓ APT install OK"
APT_TEST

echo
echo "========================================="
echo "🍺 Testing Homebrew in clean container"
echo "========================================="

docker run --rm -i homebrew/brew:latest bash <<BREW_TEST
set -e

brew tap ${BREW_TAP}
brew install ${APP}

${APP} --version || true

echo "✓ Brew install OK"
BREW_TEST

echo
echo "========================================="
echo "🎉 FULL DISTRIBUTION TEST PASSED"
echo "========================================="
EOF

chmod +x full_verify.sh
