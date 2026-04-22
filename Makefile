ifneq (,$(wildcard .env))
  include .env
  export
endif

.PHONY: help dev build test clean lint db-create migrate-up migrate-down migrate-status db-reset

help:
	@echo "MailForge - Available commands:"
	@echo "  make dev             - Run app with hot reload (Air)"
	@echo "  make build           - Build the binary"
	@echo "  make test            - Run all tests"
	@echo "  make lint            - Run linter"
	@echo "  make clean           - Clean build artifacts"
	@echo ""
	@echo "Database Commands:"
	@echo "  make db-create       - Create the MySQL database"
	@echo "  make migrate-up      - Run all pending migrations"
	@echo "  make migrate-down    - Rollback last migration"
	@echo "  make migrate-status  - Show migration status"
	@echo "  make db-reset        - ⚠️  Reset database (drop + recreate + migrate)"

dev:
	air

build:
	go build -o bin/mailforge ./cmd/api/...

test:
	go test ./... -v

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/ tmp/
	go clean

# ── Database ────────────────────────────────────────────────

db-create:
	@echo "Creating database $(DB_NAME) on $(DB_HOST)..."
	mysql -h $(DB_HOST) -P $(DB_PORT) -u $(DB_USER) -p$(DB_PASSWORD) \
		-e "CREATE DATABASE IF NOT EXISTS $(DB_NAME) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
	@echo "✅ Done."

migrate-up:
	@echo "Running migrations..."
	go run ./cmd/migrate/... up

migrate-down:
	@echo "Rolling back last migration..."
	go run ./cmd/migrate/... down

migrate-status:
	go run ./cmd/migrate/... status

db-reset:
	@echo "⚠️  Resetting database $(DB_NAME) on $(DB_HOST)..."
	@read -p "Are you sure? [y/N] " confirm && [ "$$confirm" = "y" ] || exit 1
	mysql -h $(DB_HOST) -P $(DB_PORT) -u $(DB_USER) -p$(DB_PASSWORD) \
		-e "DROP DATABASE IF EXISTS $(DB_NAME);"
	@$(MAKE) db-create
	@$(MAKE) migrate-up
	@echo "✅ Database reset complete."