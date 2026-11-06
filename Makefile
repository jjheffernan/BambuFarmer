# Podman is the preferred engine. Override with ENGINE=docker if needed.
ENGINE ?= podman
# --env-file .env: interpolate ${VAR} from repo-root secrets file
COMPOSE := $(ENGINE) compose --env-file .env -f deploy/compose.yml
SERVICES := gateway core edge integrations analytics web

VERSION ?= $(shell tr -d '[:space:]' < VERSION 2>/dev/null || echo 0.1.0-dev)

.PHONY: help env up down build logs health tidy images release-tag

help:
	@echo "Preferred engine: podman (ENGINE=docker for Docker)"
	@echo "Targets: env up down build images logs health tidy release-tag"
	@echo "Secrets: cp .env.example .env  (see docs/DEPLOYMENT_MODES.md)"
	@echo "Install wizard: docs/INSTALL.md  ·  Version: $(VERSION)"

# Create .env from the template if missing (does not overwrite).
env:
	@test -f .env || cp .env.example .env
	@echo "Using .env (edit secrets before production). Modes: docs/DEPLOYMENT_MODES.md"

up: env
	$(COMPOSE) up -d --build

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
