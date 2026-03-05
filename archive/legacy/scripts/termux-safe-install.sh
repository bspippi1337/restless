#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail

STATE="$HOME/.restless_termux_state"
LOCK="$HOME/.restless_termux_lock"
LOG="$HOME/.restless_termux_log"

exec > >(tee -a "$LOG") 2>&1

step(){ echo "$1" > "$STATE"; }
current(){ [[ -f "$STATE" ]] && cat "$STATE" || echo INIT; }

retry(){
  local n=0 max=4
  until "$@"; do
    ((n++)) || true
    if [[ $n -ge $max ]]; then
      echo "❌ Failed after $max attempts: $*"
      exit 1
    fi
    echo "Retry $n/$max: $*"
    sleep 2
  done
}

if [[ -f "$LOCK" ]]; then
  echo "⚠️ Previous run detected."
  read -rp "Resume? [Y/n]: " r
  if [[ "${r:-Y}" =~ ^[Nn]$ ]]; then
    rm -f "$STATE" "$LOCK"
  fi
fi
touch "$LOCK"

if [[ "$(current)" == "INIT" ]]; then
  echo "== Termux update =="
  retry pkg update -y
  retry pkg upgrade -y || true
  step DEPS
fi

if [[ "$(current)" == "DEPS" ]]; then
  echo "== Installing deps =="
  retry pkg install -y git curl build-essential golang
  step CLONE
fi

if [[ "$(current)" == "CLONE" ]]; then
  echo "== Fetching repo =="
  if [[ ! -d "$HOME/restless" ]]; then
    retry git clone --depth=1 https://github.com/bspippi1337/restless.git "$HOME/restless"
  else
    cd "$HOME/restless"
    retry git pull --rebase || true
  fi
  step BUILD
fi

if [[ "$(current)" == "BUILD" ]]; then
  echo "== Building =="
  cd "$HOME/restless"
  retry go mod tidy
  mkdir -p "$HOME/go/bin"
  retry go build -o "$HOME/go/bin/restless" ./cmd/restless
  step DONE
fi

if [[ "$(current)" == "DONE" ]]; then
  echo "✅ Done. Try: $HOME/go/bin/restless --help"
  rm -f "$STATE" "$LOCK"
fi
