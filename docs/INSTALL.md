# Install wizard & first bring-up

## Quick path

```bash
./scripts/setup.sh            # creates .env if missing, builds, starts stack
# or: make setup
open http://127.0.0.1:18080   # gateway redirects to /setup until complete
```

Upgrade later without wiping data:

```bash
./scripts/update.sh           # git pull --ff-only, rebuild, recreate containers
# or: make update
```

Scripts: [`scripts/README.md`](../scripts/README.md).

Complete the wizard, then use the fleet UI. Secrets (FM password, OIDC client secret, etc.) stay in **`.env`** — the wizard only stores non-secret site defaults in Postgres.

Lab DB reset / seed snapshot: [`DATABASE.md`](./DATABASE.md) (`make db-reset`).

## Wizard steps

| Step | Writes |
|------|--------|
| Welcome | — |
| Site | `sites.name`, timezone, admin username intent |
| Auth | Records `lan` / `oidc` intent — still set `AUTH_MODE` in `.env` and restart gateway |
| Costing | `rate_cards` ($/hr, $/kg, kg/hr) |
| Farm Manager | URL + transport intent — set `FARM_MANAGER_*` in `.env` on the edge host |
| Integrations | Spoolman/Manyfold URL intent — tokens in `.env` |
| Finish | Marks `install_state.completed`, optional starter presets |

APIs:

- `GET /api/v1/setup` — status + rate card + version  
- `POST /api/v1/setup` — save draft (`complete: false`) or finish (`complete: true`)  
- `POST /api/v1/setup/reset` — re-open wizard (dev)  
- `GET /api/v1/version` — build version / commit  

Gateway redirects HTML routes to `/setup` until completed. Bypass with `SKIP_SETUP_REDIRECT=true`.

## Auth modes during install

See [`DEPLOYMENT_MODES.md`](./DEPLOYMENT_MODES.md). Default is **LAN mode**. Choosing OIDC in the wizard does not enable SSO by itself — set `AUTH_MODE=oidc` and `OIDC_*` in `.env`.

## Releases

Published from git tags `vMAJOR.MINOR.PATCH` (see [`../CHANGELOG.md`](../CHANGELOG.md) and [`.github/workflows/release.yml`](../.github/workflows/release.yml)).

```bash
# After merging to main and updating CHANGELOG + VERSION:
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

The release workflow builds linux/amd64 service binaries (version ldflags), OCI images tagged `:v0.1.0` / `:0.1.0`, and attaches checksums to the GitHub Release.
