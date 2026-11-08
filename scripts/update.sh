#!/usr/bin/env bash
# Upgrade a deployed stack: pull code (optional), rebuild or pull images, recreate.
# Does not wipe Postgres/Redis volumes. Core re-runs forward migrations on boot.
# Usage: ./scripts/update.sh [--no-git] [--no-build] [--pull] [--watch]
set -euo pipefail
# shellcheck source=lib.sh
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"

NO_GIT=0
NO_BUILD=0
PULL=0
WATCH=0
for arg in "$@"; do
  case "$arg" in
    --no-git) NO_GIT=1 ;;
    --no-build) NO_BUILD=1 ;;
    --pull) PULL=1 ;;
    --watch) WATCH=1 ;;
    -h|--help)
      cat <<EOF
Usage: $(basename "$0") [--no-git] [--no-build] [--pull] [--watch]

  Update an existing BambuFarmer deploy via one command.

  (default)    git pull --ff-only, rebuild images, recreate containers
  --pull       Pull prebuilt registry images and recreate (no local build)
  --no-git     Skip git pull / fetch
  --no-build   Skip image rebuild (compose up only)
  --watch      Poll data/ops/update.requested (Admin UI "Request update")
               and apply updates automatically on the host

Env:
  ENGINE=podman|docker
  BF_GIT_REF=origin/dev   # unused today; pull is ff-only against upstream
  OPS_DIR=data/ops        # host path watched for UI update requests

Keep BambuFarmer in step with Farm Manager: after upgrading FM, run this
(or request an update from Admin while --watch is running).
EOF
      exit 0
      ;;
    *)
      echo "unknown arg: $arg" >&2
      exit 1
      ;;
  esac
done

OPS_HOST_DIR="${OPS_DIR:-data/ops}"
mkdir -p "$OPS_HOST_DIR"

bf_write_ops_status() {
  local state="$1"
  local message="${2:-}"
  local ver at
  ver="$(bf_version)"
  at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  # Minimal JSON escape for quotes/backslashes in message.
  message="${message//\\/\\\\}"
  message="${message//\"/\\\"}"
  printf '{"state":"%s","message":"%s","at":"%s","version":"%s"}\n' \
    "$state" "$message" "$at" "$ver" >"$OPS_HOST_DIR/update.status.json"
}

bf_do_update() {
  bf_require_env
  echo "==> BambuFarmer update (engine=$ENGINE)"
  bf_write_ops_status "running" "update started"

  if [[ "$PULL" -eq 1 ]]; then
    echo "==> Pulling images"
    bf_compose pull
  else
    if [[ "$NO_GIT" -eq 0 ]]; then
      if [[ -d .git ]]; then
        echo "==> Fetching git"
        git fetch --tags --prune
        if git rev-parse --abbrev-ref --symbolic-full-name '@{u}' >/dev/null 2>&1; then
          echo "==> Pulling (ff-only) $(git rev-parse --abbrev-ref HEAD)"
          git pull --ff-only
        else
          echo "warning: no upstream configured — skipped git pull (use --no-git to silence)" >&2
        fi
        echo "    at $(git rev-parse --short HEAD) $(git log -1 --pretty=%s)"
      else
        echo "warning: not a git checkout — skipped pull (pass --no-git to silence)" >&2
      fi
    fi
    if [[ "$NO_BUILD" -eq 0 ]]; then
      echo "==> Rebuilding images"
      bf_compose build
    fi
  fi

  echo "==> Recreating containers (volumes preserved)"
  bf_compose up -d --remove-orphans

  echo "==> Waiting for gateway + core"
  bf_health_wait "http://127.0.0.1:18080/healthz" 60 || true
  bf_health_wait "http://127.0.0.1:18081/healthz" 60 || true

  bf_write_ops_status "ok" "update complete"
  cat <<EOF

Update complete.

  Gateway: http://127.0.0.1:18080/healthz
  Version: curl -s http://127.0.0.1:18081/api/v1/version
  Admin:   System panel shows status after UI-requested updates

Schema: core applies forward migrations on boot (see docs/DATABASE.md).
Lab wipe/reseed only: make db-reset  (destructive)
Health: ./scripts/bf doctor
EOF
}

if [[ "$WATCH" -eq 1 ]]; then
  bf_require_env
  echo "==> Watching $OPS_HOST_DIR/update.requested (Ctrl+C to stop)"
  echo "    Admin UI: POST /api/v1/admin/system/update → drops this file"
  bf_write_ops_status "idle" "watcher running"
  while true; do
    if [[ -f "$OPS_HOST_DIR/update.requested" ]]; then
      echo "==> UI update request detected"
      rm -f "$OPS_HOST_DIR/update.requested"
      if ! bf_do_update; then
        bf_write_ops_status "error" "update failed — check logs"
        echo "error: update failed" >&2
      fi
    fi
    sleep 5
  done
fi

if bf_do_update; then
  exit 0
fi
bf_write_ops_status "error" "update failed — check logs"
exit 1
