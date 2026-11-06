COMPOSE := docker compose -f deploy/docker-compose.yml
SERVICES := gateway core edge integrations analytics web

.PHONY: help up down build logs health tidy

help:
	@echo "Targets: up down build logs health tidy"

up:
	$(COMPOSE) up -d --build

down:
	$(COMPOSE) down

build:
	$(COMPOSE) build

logs:
	$(COMPOSE) logs -f --tail=100

health:
	@for p in 18080 18081 18082 18083 18084 13000; do \
	  echo -n "port $$p: "; curl -sf "http://127.0.0.1:$$p/healthz" && echo || echo FAIL; \
	done

# Requires local Go 1.23+
tidy:
	@for s in $(SERVICES); do (cd services/$$s && go mod tidy); done
	@for p in api config events httpserver; do (cd packages/$$p && go mod tidy); done
