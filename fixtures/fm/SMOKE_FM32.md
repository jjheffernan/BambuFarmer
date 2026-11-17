# Live FM 3.2 contract smoke

Use when a Farm Manager Server (~3.2) is reachable on the LAN. Fixtures under this directory are CI/dev stand-ins (`captain.json` release `03.02.00.00`).

## Prep

```bash
# .env — never commit secrets
FARM_MANAGER_URL=http://<fm-lan-ip>:8888
FARM_MANAGER_USER=...
FARM_MANAGER_PASSWORD=...
FARM_MANAGER_TRANSPORT=fm_native
FARM_MANAGER_MIN_VERSION=03.00.00.00
# leave EDGE_FIXTURE_DIR unset / unused when URL is set
```

## Checklist

- [ ] `GET /captain` → dotted `release` ≥ min; edge `GET /api/v1/edge/status` shows `compatible: true`
- [ ] Device list populates Fleet; online/IDLE/RUNNING track via poll or MQTT
- [ ] Upload `.gcode.3mf` via Artifacts or Slice → Create job → dispatch → printer starts
- [ ] Reassign / maintenance: mark a printer maintenance → scheduler skips it
- [ ] If 3.x auth/headers drift: fix only `services/edge/app/farmmanager` (+ tests)

## Fixture smoke (no LAN)

With `FARM_MANAGER_URL` empty and `EDGE_FIXTURE_DIR=/fixtures/fm`:

```bash
go test ./services/edge/app/farmmanager/ ./services/edge/app/bambu/fixtureplane/ ./packages/api/ -count=1
```

Expect captain-shaped `03.02.00.00` and version compare helpers to pass.
