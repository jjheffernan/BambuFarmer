## Summary

<!-- 1–3 bullets: why this change exists -->

-

## Type

- [ ] feat
- [ ] fix
- [ ] refactor
- [ ] docs
- [ ] test
- [ ] chore / deploy / ci

## Services / areas

- [ ] gateway
- [ ] core
- [ ] edge
- [ ] integrations
- [ ] analytics
- [ ] web
- [ ] packages
- [ ] deploy / scripts
- [ ] docs

## Changelog

- [ ] Updated [`CHANGELOG.md`](../CHANGELOG.md) **Unreleased** (or N/A: docs-only / no user-visible change)

## Test plan

- [ ] `go test` for touched modules
- [ ] `./scripts/bf doctor` (if stack running)
- [ ] UI / API path exercised if web or public API changed

## Risks / notes

<!-- AUTH_MODE, secrets, FM/LAN, DB migrations, breaking API — or "none" -->

Do not commit `.env`. Prefer conventional commits: `feat(scope): …`.
