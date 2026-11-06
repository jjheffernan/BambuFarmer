# Podman is the preferred engine. Override with ENGINE=docker if needed.
ENGINE ?= podman
COMPOSE := $(ENGINE) compose -f deploy/compose.yml
SERVICES := gateway core edge integrations analytics web

.PHONY: help up down build logs health tidy images

help:
	@echo "Preferred engine: podman (ENGINE=docker for Docker)"
	@echo "Targets: up down build images logs health tidy"

up:
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
	@for p in api config db events httpserver; do (cd packages/$$p && go mod tidy); done
