#!/usr/bin/env bash
# Print CLI help, or help for a subcommand.
# Usage: ./scripts/help.sh [command]
set -euo pipefail
# shellcheck source=lib.sh
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"

cmd="${1:-}"

usage_main() {
  local ver
  ver="$(bf_version)"
  cat <<EOF
BambuFarmer CLI  v${ver}

Usage:
  ./scripts/bf <command> [options]
  make <command>

Deploy commands:
  setup       First bring-up (.env, build, start stack)
  update      Upgrade in place (git/rebuild or --pull; --watch for Admin UI)
  uninstall   Stop stack; optionally remove volumes/images/.env
  doctor      Run preflight + health checks
  help        Show this help (or: help <command>)

Make-only / related:
  db-reset    Wipe Postgres schema, remigrate, lab seed (destructive)
  db-snapshot pg_dump lab DB to deploy/snapshots/lab.dump

Environment:
  ENGINE=podman|docker   Container engine (default: podman, else docker)

Docs:
  docs/INSTALL.md  docs/DEPLOYMENT_MODES.md  docs/DATABASE.md  scripts/README.md

Examples:
  ./scripts/bf setup
  ./scripts/bf doctor
  ./scripts/bf update
  ./scripts/bf uninstall --yes --volumes
  ./scripts/bf help doctor
EOF
}

case "$cmd" in
  ""|-h|--help|help)
    usage_main
    ;;
  setup|update|uninstall|doctor)
    exec "$(cd "$(dirname "$0")" && pwd)/${cmd}.sh" --help
    ;;
  *)
    echo "unknown command: $cmd" >&2
    echo "Run: ./scripts/bf help" >&2
    exit 1
    ;;
esac
