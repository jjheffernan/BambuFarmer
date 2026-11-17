#!/usr/bin/env bash
# Upgrade a deployed stack: pull code (optional), rebuild or pull images, recreate.
# Does not wipe Postgres/Redis volumes. Core re-runs forward migrations on boot.
# Usage: ./scripts/update.sh [--no-git] [--no-build] [--pull] [--watch] [--components a,b]
set -euo pipefail
# shellcheck source=lib.sh
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"

NO_GIT=0
NO_BUILD=0
PULL=0
WATCH=0
COMPONENTS_CSV=""
for arg in "$@"; do
  case "$arg" in
    --no-git) NO_GIT=1 ;;
    --no-build) NO_BUILD=1 ;;
    --pull) PULL=1 ;;
    --watch) WATCH=1 ;;
    --components=*) COMPONENTS_CSV="${arg#--components=}" ;;
    -h|--help)
      cat <<EOF
Usage: $(basename "$0") [--no-git] [--no-build] [--pull] [--watch] [--components=a,b]

  Update an existing BambuFarmer deploy via one command.

  (default)    git pull --ff-only, rebuild images, recreate containers
  --pull       Pull prebuilt registry images and recreate (no local build)
  --no-git     Skip git pull / fetch
  --no-build   Skip image rebuild (compose up only)
  --components Comma-separated compose services (gateway,core,edge,web,…).
               Empty = full stack. Plugins need matching compose profiles.
  --watch      Poll data/ops/update.requested (Admin UI update buttons)
               and apply updates automatically on the host

Env:
  ENGINE=podman|docker
  OPS_DIR=data/ops

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

# Plugin services live behind compose profiles — always enable so targeted updates work.
BF_UPDATE_PROFILES=(--profile commerce --profile slicer --profile studio-bridge)

bf_write_ops_status() {
  local state="$1"
  local message="${2:-}"
  local comps="${3:-}"
  local ver at
  ver="$(bf_version)"
  at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  message="${message//\\/\\\\}"
  message="${message//\"/\\\"}"
  # components as JSON array string (already quoted elements) or empty
  if [[ -n "$comps" ]]; then
    printf '{"state":"%s","message":"%s","at":"%s","version":"%s","components":%s}\n' \
      "$state" "$message" "$at" "$ver" "$comps" >"$OPS_HOST_DIR/update.status.json"
  else
    printf '{"state":"%s","message":"%s","at":"%s","version":"%s"}\n' \
      "$state" "$message" "$at" "$ver" >"$OPS_HOST_DIR/update.status.json"
  fi
}

bf_components_json_array() {
  local csv="$1"
  if [[ -z "$csv" ]]; then
    echo '["all"]'
    return
  fi
  python3 - "$csv" <<'PY'
import json, sys
parts = [p.strip() for p in sys.argv[1].split(",") if p.strip()]
print(json.dumps(parts))
PY
}

bf_parse_request_file() {
  # Sets: REQ_COMPONENTS_CSV, REQ_PULL, REQ_NO_GIT from update.requested JSON.
  REQ_COMPONENTS_CSV=""
  REQ_PULL=0
  REQ_NO_GIT=0
  local path="$OPS_HOST_DIR/update.requested"
  [[ -f "$path" ]] || return 0
  eval "$(python3 - "$path" <<'PY'
import json, sys
try:
    d = json.load(open(sys.argv[1]))
except Exception:
    d = {}
comps = d.get("components") or []
if isinstance(comps, str):
    comps = [c.strip() for c in comps.split(",") if c.strip()]
comps = [c for c in comps if c and c.lower() != "all"]
print("REQ_COMPONENTS_CSV=%r" % (",".join(comps),))
print("REQ_PULL=%d" % (1 if d.get("pull") else 0))
print("REQ_NO_GIT=%d" % (1 if d.get("no_git") else 0))
PY
)"
}

bf_do_update() {
  local comps_csv="${1:-$COMPONENTS_CSV}"
  bf_require_env
  echo "==> BambuFarmer update (engine=$ENGINE)"
  local comps_json
  comps_json="$(bf_components_json_array "$comps_csv")"
  bf_write_ops_status "running" "update started" "$comps_json"

  # Git once for any update (needed so targeted rebuilds pick up new code).
  if [[ "$PULL" -eq 0 && "$NO_GIT" -eq 0 ]]; then
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

  local -a services=()
  if [[ -n "$comps_csv" ]]; then
    IFS=',' read -r -a services <<<"$comps_csv"
  fi

  if [[ "$PULL" -eq 1 ]]; then
    echo "==> Pulling images ${services[*]:-all}"
    if ((${#services[@]})); then
      bf_compose "${BF_UPDATE_PROFILES[@]}" pull "${services[@]}"
    else
      bf_compose "${BF_UPDATE_PROFILES[@]}" pull
    fi
  else
    if [[ "$NO_BUILD" -eq 0 ]]; then
      echo "==> Rebuilding images ${services[*]:-all}"
      if ((${#services[@]})); then
        bf_compose "${BF_UPDATE_PROFILES[@]}" build "${services[@]}"
      else
        bf_compose "${BF_UPDATE_PROFILES[@]}" build
      fi
    fi
  fi

  echo "==> Recreating containers (volumes preserved) ${services[*]:-all}"
  if ((${#services[@]})); then
    bf_compose "${BF_UPDATE_PROFILES[@]}" up -d --no-deps --remove-orphans "${services[@]}"
  else
    bf_compose "${BF_UPDATE_PROFILES[@]}" up -d --remove-orphans
  fi

  echo "==> Waiting for gateway + core"
  bf_health_wait "http://127.0.0.1:18080/healthz" 60 || true
  bf_health_wait "http://127.0.0.1:18081/healthz" 60 || true

  bf_write_ops_status "ok" "update complete" "$comps_json"
  cat <<EOF

Update complete (${services[*]:-full stack}).

  Gateway: http://127.0.0.1:18080/healthz
  Version: curl -s http://127.0.0.1:18081/api/v1/version
  Admin:   System updates panel shows per-component status

Schema: core applies forward migrations on boot (see docs/DATABASE.md).
Health: ./scripts/bf doctor
EOF
}

if [[ "$WATCH" -eq 1 ]]; then
  bf_require_env
  echo "==> Watching $OPS_HOST_DIR/update.requested (Ctrl+C to stop)"
  echo "    Admin UI: POST /api/v1/admin/system/update → drops this file"
  bf_write_ops_status "idle" "watcher running" '[]'
  while true; do
    if [[ -f "$OPS_HOST_DIR/update.requested" ]]; then
      echo "==> UI update request detected"
      bf_parse_request_file
      rm -f "$OPS_HOST_DIR/update.requested"
      # Request flags override for this run only.
      _save_pull=$PULL
      _save_nogit=$NO_GIT
      [[ "$REQ_PULL" -eq 1 ]] && PULL=1
      [[ "$REQ_NO_GIT" -eq 1 ]] && NO_GIT=1
      if ! bf_do_update "$REQ_COMPONENTS_CSV"; then
        bf_write_ops_status "error" "update failed — check logs" "$(bf_components_json_array "$REQ_COMPONENTS_CSV")"
        echo "error: update failed" >&2
      fi
      PULL=$_save_pull
      NO_GIT=$_save_nogit
    fi
    sleep 5
  done
fi

if bf_do_update "$COMPONENTS_CSV"; then
  exit 0
fi
bf_write_ops_status "error" "update failed — check logs" "$(bf_components_json_array "$COMPONENTS_CSV")"
exit 1
