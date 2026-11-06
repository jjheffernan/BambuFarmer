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

## 2. Schema layout

```text
packages/db/
  schema/baseline.sql     # human-readable full schema (canonical docs copy)
  migrations/
    001_baseline.sql      # same end-state; applied by Migrate on boot
    005_*.sql …           # forward-only deltas after the squash
  seed/lab.sql            # lab defaults (presets, etc.)
```

- **Fresh install:** empty Postgres → `db.Migrate` applies `001_baseline` (+ any `005+`) → optional `seed/lab.sql`.
- **Forward changes:** add `005_….sql`, never edit the baseline in place on released branches.
- Core (and analytics, when sharing migrate) call `db.Migrate` on boot.

---

## 3. Lab reset and snapshots

Requires compose stack up (or at least `postgres` + ability to restart `core`).

```bash
make env
make up                 # if not already running
make db-reset           # wipe public schema, restart core (migrate), apply seed/lab.sql
make db-snapshot        # optional: pg_dump -Fc → deploy/snapshots/lab.dump (local artifact)
```

**Reconfigure base** = wipe volume/schema + `db-reset`. Prefer regenerating from `seed/lab.sql` over committing binary dumps. `deploy/snapshots/*.dump` is gitignored.

---

## 4. Related

- Deploy modes: [`DEPLOYMENT_MODES.md`](./DEPLOYMENT_MODES.md)
- Install wizard: [`INSTALL.md`](./INSTALL.md)
- Compose: [`../deploy/podman/README.md`](../deploy/podman/README.md)
- Design SoR: [`DESIGN_SPEC.md`](./DESIGN_SPEC.md)
