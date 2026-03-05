#!/usr/bin/env bash

LOGDIR="logs"
mkdir -p "$LOGDIR"

LOGFILE="$LOGDIR/run_$(date +%Y%m%d_%H%M%S).log"

echo "Logging to $LOGFILE"

exec > >(tee -a "$LOGFILE") 2>&1

run() {
  echo
  echo ">>> $*"
  "$@"
  RC=$?
  echo "<<< exit code: $RC"
  return 0
}
