# BambuFarmer CLI

Operator-facing commands for bring-up, upgrades, health, and teardown.

```bash
./scripts/bf help
./scripts/bf setup
./scripts/bf doctor
./scripts/bf update
./scripts/bf uninstall --yes
```

Or via Make: `make setup`, `make update`, `make doctor`, `make uninstall`, `make cli-help`.

| Command | Script | Purpose |
|---------|--------|---------|
| `setup` | [`setup.sh`](./setup.sh) | First deploy / rebuild |
| `update` | [`update.sh`](./update.sh) | Upgrade without wiping volumes |
| `doctor` | [`doctor.sh`](./doctor.sh) | Preflight + HTTP health |
| `uninstall` | [`uninstall.sh`](./uninstall.sh) | Stop stack; optional wipe |
| `help` | [`help.sh`](./help.sh) | CLI help |
| — | [`bf`](./bf) | Dispatcher |
| — | [`lib.sh`](./lib.sh) | Shared helpers |

```bash
ENGINE=docker ./scripts/bf setup
./scripts/bf uninstall --yes --volumes   # delete Postgres/Redis data
./scripts/bf uninstall --yes --images --env
```

See [`docs/INSTALL.md`](../docs/INSTALL.md) and [`docs/DATABASE.md`](../docs/DATABASE.md).
