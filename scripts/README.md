# CLI deploy scripts

Podman-first helpers for bring-up and upgrades. From the repo root:

```bash
./scripts/setup.sh          # create .env if needed, build, up
./scripts/update.sh         # git pull (ff-only), rebuild, up (keeps volumes)
```

Or via Make:

```bash
make setup
make update
```

| Script | Purpose |
|--------|---------|
| [`setup.sh`](./setup.sh) | First deploy / full rebuild |
| [`update.sh`](./update.sh) | Routine upgrade without wiping DB |
| [`lib.sh`](./lib.sh) | Shared `ENGINE` / compose helpers |

`ENGINE=docker ./scripts/setup.sh` if Podman is unavailable.

See [`docs/INSTALL.md`](../docs/INSTALL.md) and [`docs/DATABASE.md`](../docs/DATABASE.md).
