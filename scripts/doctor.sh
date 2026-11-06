#!/usr/bin/env bash
# Preflight + runtime health checks for a BambuFarmer deploy.
# Usage: ./scripts/doctor.sh
set -euo pipefail
# shellcheck source=lib.sh
source "$(cd "$(dirname "$0")" && pwd)/lib.sh"

for arg in "$@"; do
  case "$arg" in
    -h|--help)
      cat <<EOF
Usage: $(basename "$0")

  Check host tools, .env, compose config, containers, and HTTP health.

  Exit 0 if no FAIL checks (WARN is allowed).
  Exit 1 if any FAIL checks.

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

DOCTOR_FAILS=0
DOCTOR_WARNS=0

echo "==> BambuFarmer doctor  v$(bf_version)"
echo

echo "-- Host"
if [[ -n "${ENGINE}" ]]; then
  bf_pass "container engine: $ENGINE ($(command -v "$ENGINE"))"
  if "$ENGINE" compose version >/dev/null 2>&1; then
    bf_pass "compose plugin available"
  else
    bf_fail "compose plugin missing for $ENGINE"
  fi
else
  bf_fail "neither podman nor docker found (set ENGINE=…)"
fi

if command -v curl >/dev/null 2>&1; then
  bf_pass "curl available"
else
  bf_fail "curl not found (needed for health checks)"
fi

if command -v git >/dev/null 2>&1; then
  bf_pass "git available"
else
  bf_warn "git not found (update pull skipped)"
fi

echo
echo "-- Repository"
if [[ -d .git ]]; then
  bf_pass "git checkout at $(git rev-parse --short HEAD 2>/dev/null || echo '?')"
  if [[ -n "$(git status --porcelain 2>/dev/null)" ]]; then
    bf_warn "working tree has uncommitted changes"
  else
    bf_pass "working tree clean"
  fi
  if git rev-parse --abbrev-ref --symbolic-full-name '@{u}' >/dev/null 2>&1; then
    # Count commits behind upstream if fetch not required
    behind="$(git rev-list --count 'HEAD..@{u}' 2>/dev/null || echo 0)"
    if [[ "${behind}" != "0" ]]; then
      bf_warn "branch is ${behind} commit(s) behind upstream (run: ./scripts/bf update)"
    else
      bf_pass "branch tracking upstream"
    fi
  else
    bf_warn "no git upstream configured"
  fi
else
  bf_warn "not a git checkout (release tarball / copied tree)"
fi

if [[ -f VERSION ]]; then
  bf_pass "VERSION=$(bf_version)"
else
  bf_warn "VERSION file missing"
fi

echo
echo "-- Configuration"
if [[ -f .env ]]; then
  bf_pass ".env present"
  if grep -Eq 'SESSION_SECRET=change-me|POSTGRES_PASSWORD=change-me' .env 2>/dev/null; then
    bf_warn ".env still has change-me placeholders"
  else
    bf_pass ".env placeholders look customized"
  fi
  auth="$(grep -E '^AUTH_MODE=' .env 2>/dev/null | head -1 | cut -d= -f2- || true)"
  auth="${auth:-lan}"
  bf_pass "AUTH_MODE=${auth}"
  if [[ "$auth" == "lan" ]] || [[ "$auth" == "local_insecure" ]]; then
    bf_warn "AUTH_MODE=${auth} — do not expose gateway on the public internet"
  fi
  if [[ "$auth" == "oidc" ]]; then
    for k in OIDC_ISSUER OIDC_CLIENT_ID OIDC_CLIENT_SECRET OIDC_REDIRECT_URL; do
      if grep -Eq "^${k}=.+" .env && ! grep -Eq "^${k}=$" .env; then
        bf_pass "${k} set"
      else
        bf_fail "${k} required when AUTH_MODE=oidc"
      fi
    done
  fi
else
  bf_fail ".env missing — run: ./scripts/bf setup"
fi

if [[ -f .env.example ]]; then
  bf_pass ".env.example present"
else
  bf_warn ".env.example missing"
fi

echo
echo "-- Compose"
if [[ -n "${ENGINE}" ]]; then
  if [[ -f .env ]]; then
    if bf_compose_config config -q >/dev/null 2>&1; then
      bf_pass "compose config validates"
    else
      bf_fail "compose config invalid (run: $ENGINE compose -f deploy/compose.yml config)"
    fi
  else
    bf_warn "skipping compose config (no .env)"
  fi

  if [[ -f .env ]] && bf_compose_config ps >/dev/null 2>&1; then
    running="$(bf_compose_config ps --status running -q 2>/dev/null | wc -l | tr -d ' ')"
    if [[ "${running}" -gt 0 ]]; then
      bf_pass "${running} container(s) running"
    else
      bf_warn "no containers running (run: ./scripts/bf setup)"
    fi
  fi
fi

echo
echo "-- HTTP health"
declare -a ENDPOINTS=(
  "gateway|http://127.0.0.1:18080/healthz"
  "core|http://127.0.0.1:18081/healthz"
  "edge|http://127.0.0.1:18082/healthz"
  "integrations|http://127.0.0.1:18083/healthz"
  "analytics|http://127.0.0.1:18084/healthz"
  "web|http://127.0.0.1:13000/healthz"
)
for row in "${ENDPOINTS[@]}"; do
  name="${row%%|*}"
  url="${row#*|}"
  if curl -sf --max-time 2 "$url" >/dev/null 2>&1; then
    bf_pass "$name $url"
  else
    bf_warn "$name not healthy at $url"
  fi
done

if curl -sf --max-time 2 "http://127.0.0.1:18081/api/v1/version" >/dev/null 2>&1; then
  ver_json="$(curl -sf --max-time 2 "http://127.0.0.1:18081/api/v1/version" || true)"
  bf_pass "core /api/v1/version → ${ver_json}"
fi

echo
if [[ "$DOCTOR_FAILS" -gt 0 ]]; then
  echo "Doctor: ${DOCTOR_FAILS} fail(s), ${DOCTOR_WARNS} warn(s)"
  exit 1
fi
echo "Doctor: ok (${DOCTOR_WARNS} warn(s))"
exit 0
