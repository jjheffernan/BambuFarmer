# Deployment modes: LAN vs URL’d / OIDC

BambuFarmer supports two **operator auth** modes and several **network placement** patterns. They are related but not the same thing: you can run containers on a private LAN while still choosing OIDC later, and you must keep the **edge** on the printer LAN in every pattern.

Canonical secrets template: [`.env.example`](../.env.example) → copy to **`.env`** (never commit `.env`).

---

## 1. Quick comparison

| | **LAN mode (default)** | **OIDC / URL’d deploy** |
|--|------------------------|-------------------------|
| `AUTH_MODE` | `lan` (alias: `local_insecure`) | `oidc` |
| Who authenticates operators | Local username/password in BambuFarmer | External IdP (Authentik, Keycloak, Google Workspace, …) |
| Public URL / reverse proxy | **Not required** — and **must not** be exposed to the public internet | **Required** — HTTPS URL + redirect URIs |
| IdP / OIDC_* secrets | Unused | Required for SSO (`OIDC_*`); **code exchange still stubbed in gateway** |
| Gateway bind | Private LAN / VPN / localhost only | Behind TLS reverse proxy (or gateway TLS later) |
| Boot behavior | Loud warning: insecure LAN mode | No LAN warning; fail closed if OIDC misconfigured |
| Default for | Lab, shop floor, air‑gapped / no IdP farms | Multi-user orgs, remote access, anything with a real hostname |
| Farm Manager edge | Still LAN-local to printers | Still LAN-local to printers (unchanged) |

**Rule:** Never put `AUTH_MODE=lan` on a public ingress. If operators reach the UI from outside the farm network, use `AUTH_MODE=oidc` (or VPN + LAN mode, never naked internet + LAN mode).

---

## 2. What “LAN mode” means (and what it does not)

### Means

- Operator identity is **local** to BambuFarmer (`AUTH_MODE=lan`).
- No OIDC client registration, no IdP dependency.
- Intended for a **trusted private network** (printer VLAN, shop LAN, WireGuard/Tailscale to that LAN).
- Gateway logs a boot warning and local login stays available.

### Does not mean

- It does **not** change Farm Manager transport (`FARM_MANAGER_TRANSPORT` stays `fm_native` / `cdp`).
- It does **not** move edge off the printer LAN — edge secrets and FM reachability are always LAN concerns.
- It does **not** make the stack “secure enough” for the public internet (Argon2id local users + session cookies still assume a private network boundary).

Alias: `AUTH_MODE=local_insecure` ≡ `lan`.

---

## 3. What “OIDC / URL’d” means

- Operators sign in via your IdP (`AUTH_MODE=oidc`).
- Stack is reachable at a **real HTTPS URL** (or internal URL your IdP can redirect to).
- Session/JWT still issued by **gateway**; core/edge do not speak OIDC.
- Set all `OIDC_*` values in `.env` (see template). Leave them empty in LAN mode.

Use this whenever:

- The UI/API has a DNS name or reverse proxy on a broader network, or  
- Org policy requires SSO / MFA / centralized offboarding.

---

## 4. Network placement (orthogonal to auth mode)

```text
                    ┌─────────────────────────────────────┐
                    │  Printer / Farm Manager LAN         │
                    │  FM Server :8888 / MQTT :1883       │
                    │  bambufarmer-edge  ← secrets here   │
                    └──────────────▲──────────────────────┘
                                   │ Redis / internal net
┌──────────────────────────────────┴──────────────────────┐
│  Ops network (may be same LAN)                          │
│  gateway · core · web · analytics · integrations · DB   │
└──────────────────────────────────▲──────────────────────┘
                                   │
                    Operators (LAN mode: same net / VPN only)
                    Operators (OIDC: HTTPS URL → gateway)
```

| Profile | Auth | Gateway exposure | Edge host |
|---------|------|------------------|-----------|
| **Dev / lab** | `lan` | `127.0.0.1` or LAN IP | Same machine or LAN |
| **Shop floor (no IdP)** | `lan` | LAN / VPN bind only | Printer VLAN (or routable to FM) |
| **Shop + remote SSO** | `oidc` | Reverse proxy + HTTPS | Still printer VLAN |
| **Split site (later)** | usually `oidc` | Cloud/DMZ gateway | Edge appliance on each farm LAN |

Edge credentials (`FARM_MANAGER_*`, `EDGE_CERT_DIR`, `EDGE_TOKEN`) **never** belong in the browser or in a public web tier. Keep them in `.env` / secret mounts on the edge host only.

---

## 5. Secrets inventory

| Secret / sensitive | Used by | LAN mode | OIDC mode |
|--------------------|---------|----------|-----------|
| `SESSION_SECRET` | gateway | **Required** | **Required** |
| `POSTGRES_PASSWORD` / `DATABASE_URL` | postgres, core, gateway, analytics | Required | Required |
| `FARM_MANAGER_PASSWORD` | edge | If FM live | If FM live |
| `EDGE_TOKEN` | reserved (unused in Go today) | Optional | Optional |
| `OIDC_CLIENT_SECRET` | gateway | Unused | **Required** |
| `SPOOLMAN_API_TOKEN` | integrations | Optional | Optional |
| `MANYFOLD_CLIENT_SECRET` / `MANYFOLD_TOKEN` | integrations | Optional | Optional |
| `OBJECT_STORE_*` / MinIO root | minio (profile) | Optional | Optional |
| `EDGE_CERT_DIR` materials (`*.pem`) | edge | If mTLS | If mTLS |

Store all of the above in **`.env`** (or a sealed secret store that injects the same keys). Commit only **`.env.example`** with placeholders.

---

## 6. Bring-up checklist

First visit hits the **install wizard** at `/setup` (see [`INSTALL.md`](./INSTALL.md)) until `install_state.completed`.

### LAN mode (default)

1. `cp .env.example .env`
2. Set `AUTH_MODE=lan`, strong `SESSION_SECRET`, DB password, and FM fields if talking to a real server.
3. Confirm gateway ports are not published to the public internet.
4. `make up` → complete `/setup` → fleet UI on LAN/localhost only.

### OIDC / URL’d

1. Same `.env` copy; set `AUTH_MODE=oidc` and all `OIDC_*`.
2. Register redirect URI with IdP (`OIDC_REDIRECT_URL`).
3. Terminate TLS at reverse proxy (or future gateway TLS); proxy to gateway.
4. Keep edge + `.env` FM secrets on the farm LAN host.

---

## 7. Related docs

- Design: [`DESIGN_SPEC.md`](./DESIGN_SPEC.md) §10 (identity) · §13 (12-factor)
- Database / snapshots: [`DATABASE.md`](./DATABASE.md) (Postgres SoR; Appwrite rejected)
- Gateway shape: [`services/gateway.md`](./services/gateway.md)
- Edge / FM: [`services/edge.md`](./services/edge.md), [`services/edge-farm-manager.md`](./services/edge-farm-manager.md)
- Compose / Podman: [`../deploy/podman/README.md`](../deploy/podman/README.md)
