# Farm Manager lab fixtures

Drop **redacted** JSON captures from your working Farm Manager Server here for contract tests.

Suggested files (see [`docs/services/edge-farm-manager.md`](../../docs/services/edge-farm-manager.md) §7):

| File | Source |
|------|--------|
| `captain.json` | `GET /captain` |
| `devices_idle.json` | `GET /devices` while idle |
| `devices_running.json` | `GET /devices` while printing |
| `tags.json` | `GET /tags` |

Synthetic redacted fixtures are checked in (`captain.json`, `devices_idle.json`, `devices_running.json`, `tags.json`) for contract tests. Captain release is shaped like FM **3.2** (`03.02.00.00`). Replace with captures from your lab when available.

Live LAN checklist: [`SMOKE_FM32.md`](SMOKE_FM32.md) and [`docs/services/edge-farm-manager.md`](../../docs/services/edge-farm-manager.md) (Live FM 3.2 contract smoke).

Do **not** commit passwords, JWTs, or private keys. Redact serial numbers if required by policy.

Record your server `release` string in `VERSION.txt`.
