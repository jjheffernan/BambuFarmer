# Testing layout

Unit tests live in a **module-local `tests/` package**, not beside production `.go` files.

```text
packages/config/
  config.go
  tests/
    config_test.go      # package tests
services/core/
  main.go
  internal/…
  tests/
    config_test.go
    httpx_test.go
```

## Rules

1. **No colocated `*_test.go`** next to implementation (keeps source trees scannable).
2. **`package tests`** black-box imports the code under test (exported API only).
3. **`go test ./...`** from each module still discovers `tests/` automatically.
4. Testable logic in `package main` belongs in **`internal/…`** so `tests/` can import it.
5. Fixtures stay in repo-root **`fixtures/`** (or `testdata/` when Go’s convention fits).

## Force-fail canary

`packages/config/tests` includes `TestForceFailCanary`. Normal runs skip it.  
CI sets `FORCE_TEST_FAIL=1` and asserts the suite fails (harness smoke).

```bash
FORCE_TEST_FAIL=1 go test -count=1 -run TestForceFailCanary ./packages/config/...
```
