#!/usr/bin/env bash
# Upgrade a deployed stack: pull code (optional), rebuild, rolling compose up.
# Does not wipe Postgres/Redis volumes. Core re-runs forward migrations on boot.
# Usage: ./scripts/update.sh [--no-git] [--no-build]
set -euo pipefail
# shellcheck source=lib.sh
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"

NO_GIT=0
NO_BUILD=0
for arg in "$@"; do
  case "$arg" in
    --no-git) NO_GIT=1 ;;
    --no-build) NO_BUILD=1 ;;
    -h|--help)
      cat <<EOF
Usage: $(basename "$0") [--no-git] [--no-build]

  Update an existing BambuFarmer deploy via CLI.

  --no-git     Skip git pull / fetch (use when deploying from release artifacts)
  --no-build   Skip image rebuild (compose up only)

Env:
  ENGINE=podman|docker
  BF_GIT_REF=origin/dev   # ref to merge/rebase onto when pulling (default: pull --ff-only)
EOF
      exit 0
      ;;
    *)
      echo "unknown arg: $arg" >&2
      exit 1
      ;;
  esac
done

bf_require_env

echo "==> BambuFarmer update (engine=$ENGINE)"

if [[ "$NO_GIT" -eq 0 ]]; then
  if [[ -d .git ]]; then
    echo "==> Fetching git"
    git fetch --tags --prune
    # Prefer fast-forward only so local edits are not clobbered silently.
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

echo "==> Recreating containers (volumes preserved)"
bf_compose up -d --remove-orphans

echo "==> Waiting for gateway + core"
bf_health_wait "http://127.0.0.1:18080/healthz" 60 || true
bf_health_wait "http://127.0.0.1:18081/healthz" 60 || true

cat <<EOF

Update complete.

  Gateway: http://127.0.0.1:18080/healthz
  Version: curl -s http://127.0.0.1:18081/api/v1/version

Schema: core applies forward migrations on boot (see docs/DATABASE.md).
Lab wipe/reseed only: make db-reset  (destructive)
Health: ./scripts/bf doctor
EOF
