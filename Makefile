# Makefile
.PHONY: help run build test migrate-up migrate-down migrate-create migrate-version migrate-force migrate-force-go docker-up docker-down

ENV_FILE := env/.env

ifneq (,$(wildcard $(ENV_FILE)))
    include $(ENV_FILE)
    export
endif

# Variable Setting
APP_NAME := gamers-api
BUILD_DIR := ./bin
MIGRATIONS_PATH ?= ./migrations

# Local DB URL
DB_URL := mysql://$(LOCAL_DB_USER):$(LOCAL_DB_PASSWORD)@tcp($(LOCAL_DB_HOST):$(LOCAL_DB_PORT))/$(LOCAL_DB_NAME)

# ========================================
# Basic Command
# ========================================

help: ## Show this help message
	@echo "GAMERS Backend - Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

run: ## Run the application
	@echo "🚀 Starting GAMERS API..."
	RUN_MIGRATIONS=true go run ./cmd/server.go

run-no-migrate: ## Run without migrations
	@echo "🚀 Starting GAMERS API (no migrations)..."
	go run ./cmd/server.go

build: ## Build the application
	@echo "🔨 Building $(APP_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server.go
	@echo "✅ Build complete: $(BUILD_DIR)/$(APP_NAME)"

test: ## Run tests
	@echo "🧪 Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage
	@echo "🧪 Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

clean: ## Clean build artifacts
	@echo "🧹 Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	go clean

# ========================================
# Migration 명령어
# ========================================

migrate-up: ## Run all pending migrations
	@echo "🔄 Running migrations..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up

migrate-down: ## Rollback last migration
	@echo "⏪ Rolling back last migration..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down 1

migrate-down-all: ## Rollback all migrations
	@echo "⏪ Rolling back all migrations..."
	@read -p "Are you sure? [y/N] " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down -all; \
	fi

migrate-create: ## Create new migration (usage: make migrate-create name=create_users_table)
	@if [ -z "$(name)" ]; then \
		echo "❌ Error: name parameter is required"; \
		echo "Usage: make migrate-create name=create_users_table"; \
		exit 1; \
	fi
	@echo "📝 Creating migration: $(name)"
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)

migrate-version: ## Show current migration version
	@migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" version

migrate-force: ## Force set migration version using migrate CLI (usage: make migrate-force version=1)
	@if [ -z "$(version)" ]; then \
		echo "❌ Error: version parameter is required"; \
		echo "Usage: make migrate-force version=1"; \
		exit 1; \
	fi
	@echo "⚠️  Forcing migration version to $(version)..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" force $(version)

migrate-force-go: ## Force set migration version using Go code (usage: make migrate-force-go version=1)
	@if [ -z "$(version)" ]; then \
		echo "❌ Error: version parameter is required"; \
		echo "Usage: make migrate-force-go version=1"; \
		exit 1; \
	fi
	@echo "⚠️  Forcing migration version to $(version) using Go..."
	go run scripts/force-migration.go $(version)

# ========================================
# Docker Command
# ========================================

docker-build: ## Build Docker image
	@echo "🐳 Building Docker image..."
	docker compose -f ./docker/docker-compose.yaml build

docker-up: ## Start Docker containers
	@echo "🐳 Starting Docker containers..."
	docker compose -f ./docker/docker-compose.yaml up -d

docker-down: ## Stop Docker containers
	@echo "🐳 Stopping Docker containers..."
	docker compose -f ./docker/docker-compose.yaml down

docker-logs: ## Show Docker logs
	docker compose -f ./docker/docker-compose.yaml logs -f app

docker-restart: docker down docker-up ## Restart Docker containers

migrate-up-docker: ## Run migrations in Docker environment
	@echo "🔄 Running migrations (Docker)..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up

migrate-down-docker: ## Rollback migration in Docker environment
	@echo "⏪ Rolling back migration (Docker)..."
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down 1

migrate-version-docker: ## Show migration version in Docker
	@migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" version

# ========================================
# Development Tools
# ========================================

swagger: ## Generate swagger documentation
	@echo "📚 Generating Swagger docs..."
	swag init

deps: ## Download dependencies
	@echo "📦 Downloading dependencies..."
	go mod download
	go mod tidy

redis-dump: ## Show all Redis keys
	@echo "🔍 Redis keys:"
	@docker exec gamers-redis redis-cli --scan --pattern "*" | while read key; do \
		value=$$(docker exec gamers-redis redis-cli GET "$$key"); \
		ttl=$$(docker exec gamers-redis redis-cli TTL "$$key"); \
		echo "Key: $$key"; \
		echo "Value: $$value"; \
		echo "TTL: $$ttl seconds"; \
		echo "---"; \
	done

redis-clear: ## Clear Redis database
	@echo "🗑️  Clearing Redis..."
	@docker exec gamers-redis redis-cli FLUSHDB
	@echo "✅ Redis cleared"

# ========================================
# Deployment Commands
# ========================================

docker-login-ghcr: ## Login to GitHub Container Registry
	@echo "🔐 Logging in to GitHub Container Registry..."
	@read -p "GitHub Username: " username; \
	read -sp "GitHub Token: " token; \
	echo; \
	echo $$token | docker login ghcr.io -u $$username --password-stdin

docker-push: ## Build and push Docker image to GHCR
	@echo "🚀 Building and pushing Docker image..."
	@if [ -z "$(TAG)" ]; then \
		echo "Usage: make docker-push TAG=v1.0.0"; \
		exit 1; \
	fi
	docker build -f docker/Dockerfile -t ghcr.io/$(shell git config --get remote.origin.url | sed 's/.*://;s/.git$$//' | tr '[:upper:]' '[:lower:]'):$(TAG) .
	docker push ghcr.io/$(shell git config --get remote.origin.url | sed 's/.*://;s/.git$$//' | tr '[:upper:]' '[:lower:]'):$(TAG)

docker-pull: ## Pull Docker image from GHCR
	@echo "⬇️  Pulling Docker image from GHCR..."
	@if [ -z "$(TAG)" ]; then TAG=latest; fi
	docker pull ghcr.io/$(shell git config --get remote.origin.url | sed 's/.*://;s/.git$$//' | tr '[:upper:]' '[:lower:]'):$(TAG)

check-deploy: ## Check deployment prerequisites
	@echo "🔍 Checking deployment prerequisites..."
	@command -v docker >/dev/null 2>&1 || { echo "❌ Docker is not installed"; exit 1; }
	@command -v git >/dev/null 2>&1 || { echo "❌ Git is not installed"; exit 1; }
	@echo "✅ All prerequisites met"

# ========================================
# Integration Command
# ========================================

setup: deps migrate-up ## Setup project (download deps + run migrations)
	@echo "✅ Project setup complete"

dev: docker-up run ## Start development environment (Docker + App)

rebuild: clean build ## Clean and rebuild

all: clean deps build test ## Run all tasks

.DEFAULT_GOAL := help