# Database: Postgres + snapshots (not Appwrite)

BambuFarmer’s system of record is **Postgres 16+**. Event fan-out stays on **Redis Streams**. This doc locks that choice and describes lab reset / baseline schema.

---

## 1. Verdict: keep Postgres; do not bundle Appwrite

Self-hosted [Appwrite](https://appwrite.io/) was considered as a “mature BaaS + easy snapshot” alternative to raw SQL migrations. **It is not adopted** as the farm OLTP store.

| Concern | Why Appwrite loses |
|---------|-------------------|
| Domain fit | Jobs, rate cards, printer projections, and costing are relational with transactions. Core already uses pgx SQL ([`packages/db`](../packages/db)). |
| Ops weight | Self-hosted Appwrite still runs MariaDB/Mongo + Redis + many workers/volumes. Snapshots are dumps + volume tars—not lighter than Postgres. |
| Migrations | Collection/attribute changes still need versioning. Appwrite’s `migrate` upgrades *Appwrite*, not farm schema. |
| Boundaries | Gateway owns `AUTH_MODE=lan\|oidc`; optional MinIO covers objects. Appwrite Auth/Storage would duplicate those. |

Spoolman and Manyfold remain **optional external** URLs ([`docs/services/integrations.md`](./services/integrations.md)). They are not a substitute for farm OLTP.

---

## 2. Schema layout (Go, not committed `.sql`)

Schema and lab seed are **Go string constants** under [`packages/db`](../packages/db). Raw `.sql` files are gitignored so dumps and one-off scripts never land in git.

```text
packages/db/
  migrations.go              # ordered list of migration names
  migration_001_baseline.go  # baseline DDL const
  migration_002_users.go     # users table const
  seed_lab.go                # LabSeedSQL const
  migrate.go                 # Migrate() applies orderedMigrations
  cmd/seedlab/               # CLI: migrate + SeedLab via DATABASE_URL
```

- **Fresh install:** empty Postgres → `db.Migrate` on core/gateway boot → optional `go run ./packages/db/cmd/seedlab` or `make db-reset`.
- **Forward changes:** add `migration_00N_*.go` + register in `orderedMigrations` (keep stable `Name` like `003_foo.sql` for `schema_migrations` compatibility).
- Do **not** commit `.sql` dumps, scratch scripts, or `pg_dump` text. Use `make db-snapshot` for local binary dumps only (`deploy/snapshots/*.dump`, gitignored).

---

## 3. Lab reset and snapshots

Requires compose stack up (or at least `postgres` + ability to restart `core`). Local Go 1.26+ recommended for seed CLI.

```bash
make env
make up                 # if not already running
make db-reset           # wipe public schema, restart core (migrate), seedlab
make db-snapshot        # optional: pg_dump -Fc → deploy/snapshots/lab.dump (local artifact)
```

**Reconfigure base** = wipe volume/schema + `db-reset`. Prefer regenerating from `LabSeedSQL` over committing dumps.

---

## 4. Related

- Deploy modes: [`DEPLOYMENT_MODES.md`](./DEPLOYMENT_MODES.md)
- Install wizard: [`INSTALL.md`](./INSTALL.md)
- Compose: [`../deploy/podman/README.md`](../deploy/podman/README.md)
- Design SoR: [`DESIGN_SPEC.md`](./DESIGN_SPEC.md)
