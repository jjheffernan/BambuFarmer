# Podman is the preferred engine. Override with ENGINE=docker if needed.
ENGINE ?= podman
# --env-file .env: interpolate ${VAR} from repo-root secrets file
COMPOSE := $(ENGINE) compose --env-file .env -f deploy/compose.yml
SERVICES := gateway core edge integrations analytics web

VERSION ?= $(shell tr -d '[:space:]' < VERSION 2>/dev/null || echo 0.1.0-dev)

POSTGRES_USER ?= bambufarmer
POSTGRES_DB ?= bambufarmer

.PHONY: help env up down build logs health tidy images release-tag db-reset db-snapshot \
	setup update doctor uninstall cli-help web-css

cli-help:
	@ENGINE=$(ENGINE) ./scripts/bf help

help: cli-help
	@echo ""
	@echo "Additional Make targets: env up down build images logs health tidy web-css"
	@echo "                         db-reset db-snapshot release-tag"
	@echo "Uninstall via Make: make uninstall YES=1 [VOLUMES=1] [IMAGES=1] [ENV=1]"
	@echo "Version: $(VERSION)  ·  ENGINE=$(ENGINE)"

# Create .env from the template if missing (does not overwrite).
env:
	@test -f .env || cp .env.example .env
	@echo "Using .env (edit secrets before production). Modes: docs/DEPLOYMENT_MODES.md"

up: env
	$(COMPOSE) up -d --build

# First bring-up / rebuild via CLI (see scripts/README.md).
setup:
	ENGINE=$(ENGINE) ./scripts/bf setup

# Routine upgrade: git pull --ff-only, rebuild, up (keeps volumes).
update:
	ENGINE=$(ENGINE) ./scripts/bf update

doctor:
	ENGINE=$(ENGINE) ./scripts/bf doctor

# Requires: make uninstall YES=1   (and optionally VOLUMES=1 IMAGES=1 ENV=1)
uninstall:
	@args="--yes"; \
	  [ "$(YES)" = "1" ] || { echo "Refusing: make uninstall YES=1 [VOLUMES=1] [IMAGES=1] [ENV=1]"; exit 1; }; \
	  [ "$(VOLUMES)" = "1" ] && args="$$args --volumes"; \
	  [ "$(IMAGES)" = "1" ] && args="$$args --images"; \
	  [ "$(ENV)" = "1" ] && args="$$args --env"; \
	  ENGINE=$(ENGINE) ./scripts/bf uninstall $$args

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

# Compile Tailwind → services/web/static/css/app.css (committed; go:embed).
# Requires Node 20+ locally. Source: static/css/input.css
web-css:
	cd services/web && npm install && npm run build:css

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
	@echo "Applying lab seed (Go: packages/db)…"
	@DATABASE_URL="$${DATABASE_URL:-postgres://$(POSTGRES_USER):bambufarmer@127.0.0.1:5432/$(POSTGRES_DB)?sslmode=disable}" \
		go run ./packages/db/cmd/seedlab
	@echo "db-reset complete"

# Optional local custom-format dump (gitignored). Schema/seed source of truth is Go under packages/db.
db-snapshot: env
	@mkdir -p deploy/snapshots
	$(COMPOSE) exec -T postgres pg_dump -U $(POSTGRES_USER) -d $(POSTGRES_DB) -Fc > deploy/snapshots/lab.dump
	@echo "Wrote deploy/snapshots/lab.dump"
