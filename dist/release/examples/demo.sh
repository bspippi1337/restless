#!/data/data/com.termux/files/usr/bin/sh
set -e
echo "== Restless demo =="
./restless --help | sed -n "1,80p" || true
echo
echo "[discover] openai.com"
./restless discover openai.com || true
