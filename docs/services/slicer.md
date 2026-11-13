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

## 2. API

- `POST /api/v1/slice/jobs` body:
  - `name` (required)
  - `input_path` (required when `ORCA_BIN` is set; absolute or relative to `SLICE_WORK_DIR` / `ARTIFACTS_DIR`)
  - `output_path` (optional; default sibling `*.gcode.3mf`)
  - `printer_json` / `process_json` → `--load-settings` (`;`-joined)
  - `filament_json` → `--load-filaments`
- Empty `ORCA_BIN` returns `sliced: false` and does not invoke a slicer
- When `ORCA_BIN` is set, runs the CLI (`--arrange 1 --slice 0 --export-3mf …`) and returns `sliced: true` only if exit 0 and the output file exists
- `GET /api/v1/slice/status` — `orca_configured`, `work_dir`
- The Orca CLI is operator-supplied and is not bundled in the image

---

## 3. Non-goals (v1)

- Full browser plate editor  
- Bundling into core/edge images  

See DESIGN_SPEC §8.3.
