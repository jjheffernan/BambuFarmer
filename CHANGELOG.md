# Changelog

All notable changes to BambuFarmer are documented here.  
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).  
Versioning follows [SemVer](https://semver.org/).

When you land a user-visible or ops-visible change, add a bullet under **[Unreleased]** in the same PR/commit.  
On release: move that section to a new `## [X.Y.Z] — YYYY-MM-DD`, bump [`VERSION`](VERSION), tag `vX.Y.Z`.

## [Unreleased]

### Added
- Jobs page: create form (name, artifact, printer SN, copies, estimates) + list with state/cost/error and 10s refresh
- Build-out roadmap + lean plugin docs; optional `commerce` compose profile (health stub)
- Reserved `farm.KindDirect` for a future LAN printer adapter behind `ControlPlane`
- LAN local accounts: Postgres `users`, Argon2id passwords, HMAC `bf_session` cookies
- Setup wizard creates first admin (password on Finish); `/login` + admin user CRUD
- Password change (self) and admin password reset APIs + Admin UI
- Stack updates: `./scripts/bf update` / `make update`, `--pull`, Admin “Request update” + `--watch`
- Operator CLI `./scripts/bf` (`setup`, `update`, `doctor`, `uninstall`, `help`)
- Tailwind CSS toolchain for web (`make web-css`, embedded `/static/css/app.css`)
- Shared API errors (`packages/api.Error`) with JSON `{error,code}` via `httpserver.WriteError`
- Postgres baseline migration + lab seed / `make db-reset`
- Spoolman and Manyfold list clients (optional live URLs)
- Module-local `tests/` packages and FORCE_TEST_FAIL canary
- GitHub issue/PR templates; CI changelog shape check

### Changed
- Go toolchain raised to 1.26 (CI, Containerfiles, modules) for current dependency minimums
- Dependencies: `pgx/v5` 5.12.0, `go-redis/v9` 9.23.0, `golang.org/x/crypto` 0.58.0
- Schema/seed defined in Go (`packages/db`); `*.sql` gitignored (no committed SQL dumps)
- Restyled fleet UI and install wizard with shared `bf-*` Tailwind components
- Relocated layout: `core/internal`, `gateway/auth`, `web/setup`, `edge`/`integrations` `app/`
- Safer web JSON fetches (`bfJSON`) so non-JSON API errors are readable
- After install, gateway requires login for UI/API (except setup/login/health)

### Fixed
- CI compose config check seeds `.env` from `.env.example`
- Analytics `/api/v1/analytics/counters` served from current image (was 404 → browser JSON parse error)

## [0.1.0] — 2026-11-06

### Added
- Initial modular Go services: gateway, core, edge, integrations, analytics, web
- Podman-first compose, `.env.example`, `AUTH_MODE=lan` + OIDC stubs
- Farm Manager `fm_native` adapter scaffold + printer state machine
- Versioned filament/print presets in Postgres
- First-run install wizard (`/setup`) with site, auth, costing, FM, and integrations steps
- Redis Streams bus, gateway SSE, edge fixture mode, jobs/rate cards, analytics consumer
- Release workflow (tagged `v*` → GitHub Release + module images)
- Deployment modes doc (LAN vs OIDC)

[Unreleased]: https://github.com/jjheffernan/BambuFarmer/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/jjheffernan/BambuFarmer/releases/tag/v0.1.0
