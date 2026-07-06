ifneq (,$(wildcard .env))
  include .env
  export
endif

.PHONY: help dev build test clean lint tidy \
        docker-up docker-down \
        db-create migrate-up migrate-down migrate-status db-reset \
        gen-keys

help:
	@echo ""
	@echo "MailForge — Available commands:"
	@echo ""
	@echo "  Development:"
	@echo "    make dev             - Run app with hot reload (Air)"
	@echo "    make build           - Build the binary"
	@echo "    make test            - Run all tests (with race detector)"
	@echo "    make lint            - Run linter"
	@echo "    make clean           - Clean build artifacts"
	@echo "    make tidy            - Tidy Go modules"
	@echo ""
	@echo "  Docker:"
	@echo "    make docker-up       - Start MySQL, Redis, and MailHog containers"
	@echo "    make docker-down     - Stop and remove containers"
	@echo ""
	@echo "  Database:"
	@echo "    make db-create       - Create the MySQL database"
	@echo "    make migrate-up      - Run all pending migrations"
	@echo "    make migrate-down    - Rollback last migration"
	@echo "    make migrate-status  - Show migration status"
	@echo "    make db-reset        - Reset database (drop + recreate + migrate)"
	@echo ""
	@echo "  Keys:"
	@echo "    make gen-keys        - Generate RSA key pair for JWT signing (Phase B)"
	@echo ""

dev:
	air

build:
	go build -o bin/mailforge ./cmd/api

# -race:    enables the Go race detector — catches concurrent data access bugs
# -cover:   prints test coverage per package
# -count=1: disables result caching — always runs tests fresh against real DB/Redis
test:
	go test ./... -race -cover -count=1

lint:
	golangci-lint run ./...

# Removes compiled binary and the bin/ directory
clean:
	go clean
	rm -rf bin/

tidy:
	go mod tidy

# ─── Docker ──────────────────────────────────────────────────────────────────

docker-up:
	docker compose up -d

docker-down:
	docker compose down

# ─── Database ────────────────────────────────────────────────────────────────

db-create:
	@echo "Creating database $(DB_NAME) on $(DB_HOST)..."
	mysql -h $(DB_HOST) -P $(DB_PORT) -u $(DB_USER) -p$(DB_PASSWORD) \
		-e "CREATE DATABASE IF NOT EXISTS $(DB_NAME) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
	@echo "Done."

migrate-up:
	@echo "Running migrations..."
	go run ./cmd/migration up

migrate-down:
	@echo "Rolling back last migration..."
	go run ./cmd/migration down

migrate-status:
	go run ./cmd/migration status

db-reset:
	@echo "Resetting database $(DB_NAME) on $(DB_HOST)..."
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ] || exit 1
	mysql -h $(DB_HOST) -P $(DB_PORT) -u $(DB_USER) -p$(DB_PASSWORD) \
		-e "DROP DATABASE IF EXISTS $(DB_NAME);"
	@$(MAKE) db-create
	@$(MAKE) migrate-up
	@echo "Database reset complete."

# ─── Keys ────────────────────────────────────────────────────────────────────

# Generates an RSA-2048 key pair for RS256 JWT signing.
# private.pem signs tokens — never commit this file.
# public.pem verifies tokens — safe to commit.
# Used in Phase B when we wire up authentication.
gen-keys:
	@mkdir -p keys
	@openssl genrsa -out keys/private.pem 2048
	@openssl rsa -in keys/private.pem -pubout -out keys/public.pem
	@echo ""
	@echo "Keys generated:"
	@echo "  keys/private.pem — NEVER commit this file"
	@echo "  keys/public.pem  — safe to commit"
	@echo ""