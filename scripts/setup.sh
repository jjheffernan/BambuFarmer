#!/usr/bin/env bash
# First-time (or rebuild) deploy: ensure .env, build images, start stack.
# Usage: ./scripts/setup.sh [--no-build] [--seed]
set -euo pipefail
# shellcheck source=lib.sh
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"

NO_BUILD=0
SEED=0
for arg in "$@"; do
  case "$arg" in
    --no-build) NO_BUILD=1 ;;
    --seed) SEED=1 ;;
    -h|--help)
      cat <<EOF
Usage: $(basename "$0") [--no-build] [--seed]

  First bring-up for BambuFarmer (Podman/Docker Compose).

  --no-build   Skip image build (use existing images)
  --seed       After healthy, apply packages/db/seed/lab.sql (lab only)

Env:
  ENGINE=podman|docker   (default: podman, falls back to docker)
EOF
      exit 0
      ;;
    *)
      echo "unknown arg: $arg" >&2
      exit 1
      ;;
  esac
done

echo "==> BambuFarmer setup (engine=$ENGINE)"
bf_ensure_env

if grep -q 'SESSION_SECRET=change-me' .env 2>/dev/null || grep -q 'POSTGRES_PASSWORD=change-me' .env 2>/dev/null; then
  echo "warning: .env still has change-me placeholders — set SESSION_SECRET and POSTGRES_PASSWORD" >&2
fi

if [[ "$NO_BUILD" -eq 0 ]]; then
  echo "==> Building images"
  bf_compose build
fi

echo "==> Starting stack"
bf_compose up -d

echo "==> Waiting for gateway"
bf_health_wait "http://127.0.0.1:18080/healthz" 60 || true

if [[ "$SEED" -eq 1 ]]; then
  echo "==> Waiting for core (migrate)"
  bf_health_wait "http://127.0.0.1:18081/healthz" 60 || true
  echo "==> Applying lab seed"
  bf_compose exec -T postgres psql -U "${POSTGRES_USER:-bambufarmer}" -d "${POSTGRES_DB:-bambufarmer}" \
    -v ON_ERROR_STOP=1 < packages/db/seed/lab.sql
fi

cat <<EOF

Setup complete.

  UI / wizard:  http://127.0.0.1:18080/setup
  Gateway API:  http://127.0.0.1:18080/healthz

Next:
  1. Edit .env for Farm Manager / OIDC / integrations as needed
  2. Complete the install wizard in the browser
  3. Later upgrades:  ./scripts/bf update
  4. Health check:    ./scripts/bf doctor

Docs: docs/INSTALL.md · docs/DEPLOYMENT_MODES.md · docs/DATABASE.md
     CLI: ./scripts/bf help
EOF
