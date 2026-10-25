#!/bin/bash
# Starts Orla Numis in one container: the AI agent (on 127.0.0.1:8081, not
# published) and the app (on $PORT). Both run as PUID:PGID, the owner of the
# data folders, as Unraid and other NAS systems expect. The secrets they
# share are generated once and kept with the data, unless given as
# environment variables. When either process stops, the other is stopped
# too and the container exits, so Docker restarts it.
set -euo pipefail

# A setting left empty (Unraid passes those as empty variables) would hide
# the image's default: an empty one gets it back.
export PORT="${PORT:-8080}"
export DB_PATH="${DB_PATH:-/app/data/ancientcoins.db}"
export UPLOAD_DIR="${UPLOAD_DIR:-/app/uploads}"
export BACKUP_DIR="${BACKUP_DIR:-/app/backups}"
export GIN_MODE="${GIN_MODE:-release}"
export TRUSTED_PROXIES="${TRUSTED_PROXIES:-none}"

DATA_DIR="$(dirname "${DB_PATH}")"
mkdir -p "$DATA_DIR" "$UPLOAD_DIR" "$BACKUP_DIR"

run_as=()
if [ "$(id -u)" = "0" ]; then
  PUID="${PUID:-10001}"
  PGID="${PGID:-10001}"
  # The data folders become PUID:PGID's; only what is not theirs yet is
  # touched, so a large collection does not slow every start. A share that
  # does not allow it (some network folders) is left as it is.
  for dir in "$DATA_DIR" "$UPLOAD_DIR" "$BACKUP_DIR"; do
    find "$dir" \( ! -user "$PUID" -o ! -group "$PGID" \) -exec chown "$PUID:$PGID" {} + \
      || echo "orlanumis: could not make $dir belong to $PUID:$PGID; it must be writable by them" >&2
  done
  run_as=(setpriv --reuid="$PUID" --regid="$PGID" --clear-groups)
fi

# JWT_SECRET signs the sign-in tokens; AGENT_INTERNAL_SERVICE_TOKEN lets the
# app and the agent trust each other. Given ones win; otherwise each is made
# once and kept in the data folder, so sign-ins survive restarts and updates.
SECRETS="$DATA_DIR/.orlanumis-secrets"
saved_secret() {
  [ -f "$SECRETS" ] && sed -n "s/^$1=//p" "$SECRETS" | tail -n 1
  return 0
}
ensure_secret() {
  local name="$1" value
  if [ -n "${!name:-}" ]; then
    return
  fi
  value="$(saved_secret "$name")"
  if [ -z "$value" ]; then
    value="$(python3 -c 'import secrets; print(secrets.token_urlsafe(48))')"
    (umask 077 && printf '%s=%s\n' "$name" "$value" >> "$SECRETS")
  fi
  export "$name=$value"
}
ensure_secret JWT_SECRET
ensure_secret AGENT_INTERNAL_SERVICE_TOKEN
if [ -f "$SECRETS" ]; then
  chmod 600 "$SECRETS"
  if [ "$(id -u)" = "0" ]; then
    chown "$PUID:$PGID" "$SECRETS"
  fi
fi

# The agent reaches the app on the loopback interface; AI_AGENT=off leaves
# it out (the app then works without its AI features).
export AGENT_TRUSTED_OUTBOUND_ORIGINS="${AGENT_TRUSTED_OUTBOUND_ORIGINS:-http://127.0.0.1:${PORT}}"
export AGENT_ALLOW_LOCAL_OUTBOUND="${AGENT_ALLOW_LOCAL_OUTBOUND:-true}"
export AGENT_INTERNAL_CALLBACK_URL="${AGENT_INTERNAL_CALLBACK_URL:-http://127.0.0.1:${PORT}}"

pids=()
stop() {
  if [ "${#pids[@]}" -gt 0 ]; then
    kill -TERM "${pids[@]}" 2>/dev/null || true
  fi
}
trap stop TERM INT

if [ "${AI_AGENT:-on}" != "off" ]; then
  (cd /app/agent && exec "${run_as[@]}" env HOME=/tmp /app/agent/.venv/bin/uvicorn app.main:app --host 127.0.0.1 --port 8081) &
  pids+=("$!")
fi
(cd /app && exec "${run_as[@]}" env HOME=/tmp ./ancient-coins-api) &
pids+=("$!")

status=0
wait -n "${pids[@]}" || status=$?
stop
wait || true
exit "$status"
