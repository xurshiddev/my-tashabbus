.DEFAULT_GOAL := help

API_DIR := apps/api
BOT_DIR := apps/bot
ADMIN_DIR := apps/admin
MINIAPP_DIR := apps/miniapp
COMPOSE := docker compose -f docker-compose.dev.yml
DATABASE_URL ?= postgres://my_tashabbus:my_tashabbus_password@localhost:5432/my_tashabbus?sslmode=disable

.PHONY: help check-tools dev down logs dev-api test-api build-api lint-api dev-bot test-bot build-bot lint-bot dev-admin build-admin lint-admin dev-miniapp build-miniapp lint-miniapp test build lint validate-local validate-code validate-docker bind-telegram-admin sqlc-generate migrate-up migrate-down migrate-create docker-config docker-up docker-down docker-logs

define require_tool
	@command -v $(1) >/dev/null 2>&1 || (echo "Required command '$(1)' was not found. Install it and retry this target." && exit 127)
endef

help:
	@awk 'BEGIN {FS = ":.*##"; printf "My Tashabbus commands:\n"} /^[a-zA-Z0-9_-]+:.*##/ {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

check-tools: ## Check required local development tools.
	./scripts/check-tools.sh

dev: docker-up ## Start the development Docker stack.

down: docker-down ## Stop the development Docker stack.

logs: docker-logs ## Tail Docker logs.

dev-api: ## Run the API locally.
	$(call require_tool,go)
	cd $(API_DIR) && go run ./cmd/api

test-api: ## Run API tests.
	$(call require_tool,go)
	cd $(API_DIR) && go test ./...

build-api: ## Build the API.
	$(call require_tool,go)
	cd $(API_DIR) && go build ./...

lint-api: ## Run API lint checks.
	$(call require_tool,go)
	cd $(API_DIR) && go vet ./...

dev-bot: ## Run the Telegram bot locally.
	$(call require_tool,go)
	cd $(BOT_DIR) && go run ./cmd/bot

test-bot: ## Run bot tests.
	$(call require_tool,go)
	cd $(BOT_DIR) && go test ./...

build-bot: ## Build the bot.
	$(call require_tool,go)
	cd $(BOT_DIR) && go build ./...

lint-bot: ## Run bot lint checks.
	$(call require_tool,go)
	cd $(BOT_DIR) && go vet ./...

dev-admin: ## Run the admin dashboard locally.
	$(call require_tool,npm)
	cd $(ADMIN_DIR) && npm run dev

build-admin: ## Build the admin dashboard.
	$(call require_tool,npm)
	cd $(ADMIN_DIR) && npm run build

lint-admin: ## Run admin TypeScript checks.
	$(call require_tool,npm)
	cd $(ADMIN_DIR) && npm run lint

dev-miniapp: ## Run the Telegram Mini App locally.
	$(call require_tool,npm)
	cd $(MINIAPP_DIR) && npm run dev

build-miniapp: ## Build the Telegram Mini App.
	$(call require_tool,npm)
	cd $(MINIAPP_DIR) && npm run build

lint-miniapp: ## Run Mini App TypeScript checks.
	$(call require_tool,npm)
	cd $(MINIAPP_DIR) && npm run lint

test: test-api test-bot ## Run all tests.

build: build-api build-bot build-admin build-miniapp ## Build all apps.

lint: lint-api lint-bot lint-admin lint-miniapp ## Run all lint checks.

validate-local: check-tools validate-code validate-docker ## Run all local validation checks.

validate-code: ## Run code validation for Go and frontend apps.
	$(call require_tool,go)
	$(call require_tool,npm)
	@if command -v sqlc >/dev/null 2>&1; then cd $(API_DIR) && sqlc generate; else echo "sqlc not found; skipping sqlc generate check"; fi
	cd $(API_DIR) && go mod tidy && go test ./... && go build ./...
	cd $(BOT_DIR) && go mod tidy && go test ./... && go build ./...
	cd $(ADMIN_DIR) && npm install && npm run build && npm run lint
	cd $(MINIAPP_DIR) && npm install && npm run build && npm run lint

sqlc-generate: ## Generate sqlc code for API queries.
	$(call require_tool,sqlc)
	cd $(API_DIR) && sqlc generate

validate-docker: ## Validate Docker Compose and API health.
	$(call require_tool,docker)
	$(call require_tool,curl)
	$(COMPOSE) config
	$(COMPOSE) up -d postgres api
	@trap '$(COMPOSE) down' EXIT; \
	for attempt in 1 2 3 4 5 6 7 8 9 10; do \
		if curl -fsS http://localhost:8080/health; then \
			echo ""; \
			exit 0; \
		fi; \
		echo "Waiting for API health check... ($$attempt/10)"; \
		sleep 2; \
	done; \
	echo "API health check failed"; \
	$(COMPOSE) logs api; \
	exit 1

bind-telegram-admin: ## Bind local admin Telegram identity for real Mini App testing.
	./scripts/bind-telegram-admin.sh

migrate-up: ## Apply database migrations.
	$(call require_tool,migrate)
	migrate -path $(API_DIR)/migrations -database "$(DATABASE_URL)" up

migrate-down: ## Roll back one database migration.
	$(call require_tool,migrate)
	migrate -path $(API_DIR)/migrations -database "$(DATABASE_URL)" down 1

migrate-create: ## Create a new migration with name=some_name.
	$(call require_tool,migrate)
	@test -n "$(name)" || (echo "Usage: make migrate-create name=some_name" && exit 1)
	migrate create -ext sql -dir $(API_DIR)/migrations -seq $(name)

docker-config: ## Validate Docker Compose configuration.
	$(call require_tool,docker)
	$(COMPOSE) config

docker-up: ## Start postgres and API with Docker Compose.
	$(call require_tool,docker)
	$(COMPOSE) up -d postgres api

docker-down: ## Stop Docker Compose services.
	$(call require_tool,docker)
	$(COMPOSE) down

docker-logs: ## Tail Docker Compose logs.
	$(call require_tool,docker)
	$(COMPOSE) logs -f
