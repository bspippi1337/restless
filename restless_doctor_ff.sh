#!/usr/bin/env bash
set -euo pipefail

# ==============================================
#  RESTLESS DOCTOR FF
#  Surgeon + Scout (knots) + MacGyver + Agnostic
#  Diagnose first. Cure when asked. Leave a trail.
# ==============================================

TARGET="${TARGET:-example.com}"
MODE="diagnose"
NUKE_LOCAL="no"

if [[ "${1:-}" == "--repair" ]]; then MODE="repair"; shift; fi
if [[ "${1:-}" == "--nuke-local" ]]; then NUKE_LOCAL="yes"; shift; fi
if [[ "${1:-}" != "" ]]; then TARGET="$1"; fi

TS="$(date +%Y%m%d_%H%M%S)"
REPORT="doctor_ff_${TS}.log"
BACKUP_DIR="${HOME}/.restless-doctor-ff/backup_${TS}"

log() { echo -e "$*" | tee -a "$REPORT" >/dev/null; }
run() { log "\n$ $*"; ( "$@" 2>&1 || true ) | tee -a "$REPORT" >/dev/null; }

banner() {
  log "=============================================="
  log "           RESTLESS DOCTOR FF"
  log "=============================================="
  log "Target: ${TARGET}"
  log "Mode:   ${MODE}"
  log "Nuke local CAs: ${NUKE_LOCAL}"
  log "Report: ${REPORT}"
  log "=============================================="
}

preflight() {
  : > "$REPORT"
  banner

  log "\n==> System"
  run uname -a
  run date
  run id
  run pwd

  log "\n==> Network quick peek"
  run getent hosts "$TARGET"
  run ip -brief addr || true
  run ip route || true

  log "\n==> Tooling"
  run which curl
  run curl --version
  run which openssl
  run openssl version
  run openssl version -d
  run openssl version -a

  log "\n==> SSL env (often the hidden knot)"
  run env
}

diagnose_tls() {
  log "\n==> CA store presence"
  run ls -lah /etc/ssl/certs
  run ls -lh /etc/ssl/certs/ca-certificates.crt
  run stat /etc/ssl/certs/ca-certificates.crt

  log "\n==> Local CA injections (these can poison trust chains)"
  run ls -lah /usr/local/share/ca-certificates || true
  run find /usr/local/share/ca-certificates -maxdepth 1 -type f -name "*.crt" -print || true

  log "\n==> OpenSSL chain + verify (CApath)"
  run openssl s_client -connect "${TARGET}:443" -servername "${TARGET}" -CApath /etc/ssl/certs < /dev/null

  log "\n==> OpenSSL chain + verify (CAfile)"
  run openssl s_client -connect "${TARGET}:443" -servername "${TARGET}" -CAfile /etc/ssl/certs/ca-certificates.crt < /dev/null

  log "\n==> Show cert chain count"
  CERTS_SENT="$(openssl s_client -connect "${TARGET}:443" -servername "${TARGET}" -showcerts < /dev/null 2>/dev/null | grep -c "BEGIN CERTIFICATE" || true)"
  log "Certificates sent by server: ${CERTS_SENT}"

  log "\n==> Curl tests (default)"
  run curl -I "https://${TARGET}"

  log "\n==> Curl with explicit CAfile"
  run curl --cacert /etc/ssl/certs/ca-certificates.crt -I "https://${TARGET}"

  log "\n==> Curl with explicit CApath"
  run curl --capath /etc/ssl/certs -I "https://${TARGET}"

  log "\n==> Curl forced IPv4"
  run curl -4 -I "https://${TARGET}"

  log "\n==> Curl forced IPv6"
  run curl -6 -I "https://${TARGET}"

  log "\n==> Curl TLS caps"
  run curl --tls-max 1.2 -I "https://${TARGET}"
  run curl --tlsv1.3 -I "https://${TARGET}"

  log "\n==> Control group (if these fail too, it's system-wide)"
  for H in google.com cloudflare.com; do
    log "\n--- Control: ${H} ---"
    run curl -I "https://${H}"
  done
}

backup_everything() {
  log "\n==> Backup (so we can undo surgery)"
  run mkdir -p "$BACKUP_DIR"
  run mkdir -p "${BACKUP_DIR}/etc_ssl" "${BACKUP_DIR}/local_ca"

  # Copy instead of tar first so it works even if tar isn't there.
  run sudo cp -a /etc/ssl "${BACKUP_DIR}/etc_ssl/" || true
  run sudo cp -a /usr/local/share/ca-certificates "${BACKUP_DIR}/local_ca/" || true

  # Make a single tarball too (handy).
  run tar -czf "${BACKUP_DIR}.tar.gz" -C "$(dirname "$BACKUP_DIR")" "$(basename "$BACKUP_DIR")" || true

  log "Backup saved to:"
  log "  ${BACKUP_DIR}"
  log "  ${BACKUP_DIR}.tar.gz"
}

repair_safely() {
  backup_everything

  log "\n==> Surgery step 1: remove local CA injections (optional)"
  if [[ "$NUKE_LOCAL" == "yes" ]]; then
    log "NUKE_LOCAL=yes: moving local .crt files out (after backup)."
    run sudo mkdir -p /usr/local/share/ca-certificates/_doctor_ff_quarantine
    # Move only .crt, keep dirs.
    run bash -lc 'shopt -s nullglob; for f in /usr/local/share/ca-certificates/*.crt; do sudo mv "$f" /usr/local/share/ca-certificates/_doctor_ff_quarantine/; done'
  else
    log "NUKE_LOCAL=no: leaving /usr/local/share/ca-certificates alone."
  fi

  log "\n==> Surgery step 2: reinstall trust + crypto stack"
  if command -v apt >/dev/null 2>&1; then
    run sudo apt update -y
    run sudo apt install --reinstall -y ca-certificates openssl libssl3
  else
    log "apt not found. Skipping package reinstall."
  fi

  log "\n==> Surgery step 3: rebuild trust store"
  run sudo update-ca-certificates --fresh

  log "\n==> Surgery step 4: rehash (sometimes the missing knot)"
  if command -v c_rehash >/dev/null 2>&1; then
    run sudo c_rehash /etc/ssl/certs
  else
    log "c_rehash not found (ok on some systems)."
  fi

  log "\n==> Surgery step 5: ensure curl/openssl are pointed at the right trust store (session-only)"
  export SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt
  export SSL_CERT_DIR=/etc/ssl/certs
  log "Set env (this shell only):"
  log "  SSL_CERT_FILE=${SSL_CERT_FILE}"
  log "  SSL_CERT_DIR=${SSL_CERT_DIR}"

  log "\n==> Surgery step 6: re-test target + controls"
  diagnose_tls

  log "\n==> If still failing: print likely causes"
  log "- If target fails but google/cloudflare work: upstream edge chain issue."
  log "- If everything fails: OpenSSL policy/provider config or middlebox/DPI."
  log "- If only IPv6 fails: IPv6 path/MTU/DPI issues."
  log "- If only TLS1.3 fails: DPI that breaks TLS1.3; use --tls-max 1.2 as workaround."
}

final_words() {
  log "\n=============================================="
  log "Doctor FF finished."
  log "Report: ${REPORT}"
  if [[ "$MODE" == "repair" ]]; then
    log "Backup: ${BACKUP_DIR}.tar.gz"
    log "To undo: restore /etc/ssl + /usr/local/share/ca-certificates from backup."
  fi
  log "=============================================="
}

main() {
  preflight

  case "$MODE" in
    diagnose)
      diagnose_tls
      ;;
    repair)
      repair_safely
      ;;
    *)
      log "Unknown mode: $MODE"
      exit 2
      ;;
  esac

  final_words
}

main
