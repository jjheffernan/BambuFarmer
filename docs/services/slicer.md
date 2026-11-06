# Service shape: `bambufarmer-slicer` (post-MVP)

**Language:** Go + OrcaSlicer CLI binary  
**Container:** separate heavy image (not in default compose)  
**Role:** Async slice jobs → `.gcode.3mf` + metadata for core costing  

---

## 1. Responsibility

Wrap OrcaSlicer headless CLI; do not embed a slicer GUI. Isolate CPU/RAM from edge/core.

Typical invocation:

```bash
orca-slicer model.3mf \
  --load-settings "printer.json;process.json" \
  --load-filaments filament.json \
  --arrange 1 --slice 0 \
  --export-3mf out.gcode.3mf
```

Preset `process_json` / `printer_json` from core feed `--load-settings`.

---

## 2. API (future)

- `POST /api/v1/slice/jobs`  
- `GET /api/v1/slice/jobs/{id}`  
- Events: `bf.events.slice`  

---

## 3. Non-goals (v1)

- Full browser plate editor  
- Bundling into core/edge images  

See DESIGN_SPEC §8.3.
