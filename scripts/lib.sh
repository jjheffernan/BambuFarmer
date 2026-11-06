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
    echo "error: need podman or docker (set ENGINE=…)" >&2
    exit 1
  fi
fi

COMPOSE=("$ENGINE" compose --env-file .env -f deploy/compose.yml)

bf_require_env() {
  if [[ ! -f .env ]]; then
    echo "error: .env missing — run scripts/setup.sh first (or: cp .env.example .env)" >&2
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
  bf_require_env
  "${COMPOSE[@]}" "$@"
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
