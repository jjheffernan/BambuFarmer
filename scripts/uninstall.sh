#!/usr/bin/env bash
# Tear down a BambuFarmer compose deploy.
# Usage: ./scripts/uninstall.sh --yes [--volumes] [--images] [--env]
set -euo pipefail
# shellcheck source=lib.sh
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"

YES=0
VOLUMES=0
IMAGES=0
REMOVE_ENV=0

for arg in "$@"; do
  case "$arg" in
    -y|--yes) YES=1 ;;
    --volumes) VOLUMES=1 ;;
    --images) IMAGES=1 ;;
    --env) REMOVE_ENV=1 ;;
    -h|--help)
      cat <<EOF
Usage: $(basename "$0") --yes [--volumes] [--images] [--env]

  Stop and remove the BambuFarmer Compose project.

  --yes        Required confirmation
  --volumes    Also delete named volumes (Postgres/Redis/MinIO data)
  --images     Remove local images tagged localhost/bambufarmer-*:dev
  --env        Delete repo-root .env (secrets) — irreversible

Without --volumes, database data is kept on disk for a later setup.

Env:
  ENGINE=podman|docker
EOF
      exit 0
      ;;
    *)
      echo "unknown arg: $arg" >&2
      exit 1
      ;;
  esac
done

if [[ "$YES" -ne 1 ]]; then
  echo "error: refusing to uninstall without --yes" >&2
  echo "  preview: $ENGINE compose -f deploy/compose.yml ps" >&2
  echo "  destroy: ./scripts/bf uninstall --yes" >&2
  echo "  wipe DB: ./scripts/bf uninstall --yes --volumes" >&2
  exit 1
fi

bf_require_engine

echo "==> BambuFarmer uninstall (engine=$ENGINE)"

if [[ -f .env ]]; then
  down_args=(down --remove-orphans)
  if [[ "$VOLUMES" -eq 1 ]]; then
    down_args+=(--volumes)
    echo "==> Stopping stack and deleting volumes"
  else
    echo "==> Stopping stack (volumes preserved)"
  fi
  bf_compose "${down_args[@]}"
else
  echo "warning: no .env — attempting compose down without env-file" >&2
  "$ENGINE" compose -f deploy/compose.yml down --remove-orphans ${VOLUMES:+--volumes} || true
fi

if [[ "$IMAGES" -eq 1 ]]; then
  echo "==> Removing localhost/bambufarmer-*:dev images"
  for s in gateway core edge integrations analytics web; do
    "$ENGINE" image rm -f "localhost/bambufarmer-${s}:dev" 2>/dev/null || true
  done
fi

if [[ "$REMOVE_ENV" -eq 1 ]]; then
  if [[ -f .env ]]; then
    rm -f .env
    echo "==> Removed .env"
  fi
fi

cat <<EOF

Uninstall complete.

  Re-install:  ./scripts/bf setup
  Doctor:      ./scripts/bf doctor
EOF
