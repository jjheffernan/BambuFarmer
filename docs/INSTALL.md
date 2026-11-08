# Install wizard & first bring-up

## Quick path

```bash
./scripts/bf setup            # creates .env if missing, builds, starts stack
./scripts/bf doctor           # preflight + health
open http://127.0.0.1:18080   # gateway redirects to /setup until complete
```

## Updates (keep pace with Farm Manager)

Containers do not rebuild themselves. Use **one host command**, or the Admin UI plus a host watcher.

```bash
./scripts/bf update              # git pull --ff-only, rebuild, recreate (volumes kept)
make update                      # same
./scripts/bf update --pull       # pull prebuilt registry images + recreate (no local build)
./scripts/bf update --watch      # apply Admin → “Request update” automatically
```

**UI path:** Admin → System updates → **Request update** writes `data/ops/update.requested`. A host process running `./scripts/bf update --watch` applies it (status in `data/ops/update.status.json`).

After upgrading Farm Manager on the printer LAN, update BambuFarmer the same day so edge/FM API assumptions stay aligned. Forward DB migrations run automatically on core boot.

### Tear down

```bash
./scripts/bf uninstall --yes              # stop containers, keep volumes
./scripts/bf uninstall --yes --volumes    # also delete Postgres/Redis data
```

CLI overview: `./scripts/bf help` · [`scripts/README.md`](../scripts/README.md).

Complete the wizard (including admin password), then sign in at `/login` and use the fleet UI. Farm Manager / OIDC secrets stay in **`.env`**. Operator passwords are Argon2id hashes in Postgres `users` only. Set a real `SESSION_SECRET` before production.

**Password reset (LAN):** signed-in users change their own password via Admin → “Change my password” (`POST /api/v1/auth/password`). Admins can set a new password for any user (`POST /api/v1/admin/users/{id}/password`). There is no email-based forgot-password flow in LAN mode.

Lab DB reset / seed snapshot: [`DATABASE.md`](./DATABASE.md) (`make db-reset`).

## Wizard steps

| Step | Writes |
|------|--------|
| Welcome | — |
| Site | `sites.name`, timezone, admin username |
| Auth | Records `lan` / `oidc` intent — still set `AUTH_MODE` in `.env`. Finish step sets admin **password** (hashed into `users`; not stored on `install_state`). |
| Costing | `rate_cards` ($/hr, $/kg, kg/hr) |
| Farm Manager | URL + transport intent — set `FARM_MANAGER_*` in `.env` on the edge host |
| Integrations | Spoolman/Manyfold URL intent — tokens in `.env` |
| Finish | Marks `install_state.completed`, optional starter presets |

APIs:

- `GET /api/v1/setup` — status + rate card + version  
- `POST /api/v1/setup` — save draft (`complete: false`) or finish (`complete: true`)  
- `POST /api/v1/setup/reset` — re-open wizard (dev)  
- `GET /api/v1/version` — build version / commit  
- `GET /api/v1/admin/system` — version + update status (admin)  
- `POST /api/v1/admin/system/update` — request host-side update (admin; needs `--watch`)  

Gateway redirects HTML routes to `/setup` until completed. Bypass with `SKIP_SETUP_REDIRECT=true`.

## Auth modes during install

See [`DEPLOYMENT_MODES.md`](./DEPLOYMENT_MODES.md). Default is **LAN mode**. Choosing OIDC in the wizard does not enable SSO by itself — set `AUTH_MODE=oidc` and `OIDC_*` in `.env`.

## Releases

Published from git tags `vMAJOR.MINOR.PATCH` (see [`../CHANGELOG.md`](../CHANGELOG.md) and [`.github/workflows/release.yml`](../.github/workflows/release.yml)).

**Changelog hygiene:** every user/ops-visible change gets a bullet under `## [Unreleased]` in the same PR. On release:

1. Move Unreleased bullets into `## [X.Y.Z] — YYYY-MM-DD`
2. Set [`VERSION`](../VERSION) to `X.Y.Z`
3. Update compare links at the bottom of `CHANGELOG.md`
4. Tag and push:

```bash
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

The release workflow builds linux/amd64 service binaries (version ldflags), OCI images tagged `:v0.1.0` / `:0.1.0`, attaches checksums, and uses that changelog section as GitHub Release notes.
