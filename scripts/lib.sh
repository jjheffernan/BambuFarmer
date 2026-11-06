#!/usr/bin/env bash
# Shared helpers for BambuFarmer CLI scripts. Source from repo scripts only.
set -euo pipefail

_SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$_SCRIPT_DIR/.." && pwd)"
cd "$ROOT"

ENGINE="${ENGINE:-podman}"
if ! command -v "$ENGINE" >/dev/null 2>&1; then
  if command -v docker >/dev/null 2>&1; then
    ENGINE=docker
  else
    ENGINE=""
  fi
fi

COMPOSE_FILE="deploy/compose.yml"

bf_require_engine() {
  if [[ -z "${ENGINE}" ]]; then
    echo "error: need podman or docker (set ENGINE=…)" >&2
    exit 1
  fi
}

bf_require_env() {
  if [[ ! -f .env ]]; then
    echo "error: .env missing — run: ./scripts/bf setup" >&2
    exit 1
  fi
}

bf_ensure_env() {
  if [[ ! -f .env ]]; then
    cp .env.example .env
    echo "Created .env from .env.example — edit secrets before production."
  fi
}

bf_compose() {
  bf_require_engine
  bf_require_env
  if [[ -f .env ]]; then
    "$ENGINE" compose --env-file .env -f "$COMPOSE_FILE" "$@"
  else
    "$ENGINE" compose -f "$COMPOSE_FILE" "$@"
  fi
}

# Like bf_compose but allows missing .env (doctor / early uninstall).
bf_compose_config() {
  bf_require_engine
  if [[ -f .env ]]; then
    "$ENGINE" compose --env-file .env -f "$COMPOSE_FILE" "$@"
  else
    "$ENGINE" compose -f "$COMPOSE_FILE" "$@"
  fi
}

bf_health_wait() {
  local url="${1:-http://127.0.0.1:18080/healthz}"
  local tries="${2:-45}"
  local i=0
  while (( i < tries )); do
    if curl -sf "$url" >/dev/null 2>&1; then
      return 0
    fi
    i=$((i + 1))
    sleep 1
  done
  echo "warning: timed out waiting for $url" >&2
  return 1
}

bf_version() {
  if [[ -f VERSION ]]; then
    tr -d '[:space:]' < VERSION
  else
    echo "0.0.0-unknown"
  fi
}

bf_pass() { printf '  PASS  %s\n' "$*"; }
bf_warn() { printf '  WARN  %s\n' "$*"; DOCTOR_WARNS=$((DOCTOR_WARNS + 1)); }
bf_fail() { printf '  FAIL  %s\n' "$*"; DOCTOR_FAILS=$((DOCTOR_FAILS + 1)); }
