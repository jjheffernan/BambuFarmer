# Podman is the preferred engine. Override with ENGINE=docker if needed.
ENGINE ?= podman
# --env-file .env: interpolate ${VAR} from repo-root secrets file
COMPOSE := $(ENGINE) compose --env-file .env -f deploy/compose.yml
SERVICES := gateway core edge integrations analytics web

VERSION ?= $(shell tr -d '[:space:]' < VERSION 2>/dev/null || echo 0.1.0-dev)

POSTGRES_USER ?= bambufarmer
POSTGRES_DB ?= bambufarmer

.PHONY: help env up down build logs health tidy images release-tag db-reset db-snapshot setup update

help:
	@echo "Preferred engine: podman (ENGINE=docker for Docker)"
	@echo "Targets: env up down build images logs health tidy release-tag"
	@echo "         setup update          (CLI deploy — scripts/setup.sh, scripts/update.sh)"
	@echo "         db-reset db-snapshot  (see docs/DATABASE.md)"
	@echo "Secrets: cp .env.example .env  (see docs/DEPLOYMENT_MODES.md)"
	@echo "Install wizard: docs/INSTALL.md  ·  Version: $(VERSION)"

# Create .env from the template if missing (does not overwrite).
env:
	@test -f .env || cp .env.example .env
	@echo "Using .env (edit secrets before production). Modes: docs/DEPLOYMENT_MODES.md"

up: env
	$(COMPOSE) up -d --build

# First bring-up / rebuild via CLI (see scripts/README.md).
setup:
	ENGINE=$(ENGINE) ./scripts/setup.sh

# Routine upgrade: git pull --ff-only, rebuild, up (keeps volumes).
update:
	ENGINE=$(ENGINE) ./scripts/update.sh

down:
	$(COMPOSE) down

build:
	$(COMPOSE) build

# Build each module image independently (no fat multi-binary image).
images:
	@for s in $(SERVICES); do \
	  echo "==> localhost/bambufarmer-$$s:dev"; \
	  $(ENGINE) build -t localhost/bambufarmer-$$s:dev -f services/$$s/Containerfile .; \
	done

logs:
	$(COMPOSE) logs -f --tail=100

health:
	@for p in 18080 18081 18082 18083 18084 13000; do \
	  printf "port %s: " "$$p"; curl -sf "http://127.0.0.1:$$p/healthz" && echo || echo FAIL; \
	done

# Requires local Go 1.23+
tidy:
	@for s in $(SERVICES); do (cd services/$$s && go mod tidy); done
	@for p in api config db events farm httpserver printerstate version; do (cd packages/$$p && go mod tidy); done

# Create an annotated tag from VERSION (does not push).
release-tag:
	@test -f VERSION
	@echo "Tagging v$$(tr -d '[:space:]' < VERSION) — update CHANGELOG first, then: git push origin v$$(tr -d '[:space:]' < VERSION)"
	git tag -a "v$$(tr -d '[:space:]' < VERSION)" -m "v$$(tr -d '[:space:]' < VERSION)"

# Wipe public schema, restart core (runs Migrate), apply lab seed.
# Requires: compose postgres (+ core) running. See docs/DATABASE.md.
db-reset: env
	@echo "Resetting Postgres schema $(POSTGRES_DB)…"
	$(COMPOSE) exec -T postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 -c \
		"DROP SCHEMA public CASCADE; CREATE SCHEMA public; GRANT ALL ON SCHEMA public TO $(POSTGRES_USER); GRANT ALL ON SCHEMA public TO public;"
	$(COMPOSE) restart core
	@echo "Waiting for core migrate…"
	@i=0; while [ $$i -lt 30 ]; do \
	  if curl -sf http://127.0.0.1:18081/healthz >/dev/null 2>&1; then break; fi; \
	  i=$$((i+1)); sleep 1; \
	done
	@echo "Applying seed/lab.sql…"
	$(COMPOSE) exec -T postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -v ON_ERROR_STOP=1 < packages/db/seed/lab.sql
	@echo "db-reset complete"

# Optional local custom-format dump (gitignored). Prefer seed/lab.sql as source of truth.
db-snapshot: env
	@mkdir -p deploy/snapshots
	$(COMPOSE) exec -T postgres pg_dump -U $(POSTGRES_USER) -d $(POSTGRES_DB) -Fc > deploy/snapshots/lab.dump
	@echo "Wrote deploy/snapshots/lab.dump"
