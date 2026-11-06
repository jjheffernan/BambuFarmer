# Farm Manager lab fixtures

Drop **redacted** JSON captures from your working Farm Manager Server here for contract tests.

Suggested files (see [`docs/services/edge-farm-manager.md`](../../docs/services/edge-farm-manager.md) §7):

| File | Source |
|------|--------|
| `captain.json` | `GET /captain` |
| `devices_idle.json` | `GET /devices` while idle |
| `devices_running.json` | `GET /devices` while printing |
| `tags.json` | `GET /tags` |

Do **not** commit passwords, JWTs, or private keys. Redact serial numbers if required by policy.

Record your server `release` string in a `VERSION.txt` next to the fixtures.
