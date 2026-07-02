# MailForge

MailForge is an early-stage email campaign API built in Go. The codebase currently focuses on the backend foundation: application bootstrapping, dependency injection, configuration loading, MySQL connectivity, request logging, uniform HTTP error responses, server lifecycle management, a health route, embedded SQL migrations, and focused tests around the foundation.

The feature modules for authentication, campaigns, and organization are scaffolded, but their business logic and HTTP endpoints are not implemented yet.

## Tech Stack

- Go 1.26.1
- Chi for HTTP routing
- Uber Fx for dependency injection and application lifecycle
- Bun ORM for MySQL database access
- Zap for structured logging
- Godotenv for local environment loading
- Bun migrations with embedded SQL files

## Current Features

- Application startup from `cmd/api/main.go`
- Environment configuration from `.env` or system environment variables
- Database DSN building from `DB_*` fields, with optional `DB_DSN` override
- Fx dependency container in `internal/di`
- MySQL connection setup through Bun
- Connection pool configuration
- Bun query debugging in non-production environments
- Structured application logger
- Request logging middleware
- Custom panic recovery middleware
- Shared HTTP status constants in `internal/constants`
- Uniform JSON error helpers in `internal/response`
- Chi router with request IDs and JSON 404/405 handlers
- `GET /health` endpoint
- Graceful HTTP server shutdown
- Embedded SQL migrations
- Standalone migration command in `cmd/migration`
- Focused tests for config and router behavior

## Not Implemented Yet

These areas are planned or scaffolded, but not currently functional:

- User registration and login
- Password hashing
- JWT creation and validation
- Protected routes
- Campaign CRUD endpoints
- Campaign sending
- Subscriber and list endpoints
- Organization behavior
- Email provider or SMTP sending service
- Request and response DTOs for feature modules
- Domain/Bun model structs

## Getting Started

1. Install Go 1.26.1 or a compatible local toolchain.
2. Create a `.env` file from `.env.example`.
3. Configure the MySQL database values.
4. Run the API:

```bash
go run ./cmd/api
```

By default, the app uses `APP_PORT=3010` if no port is provided.

## Environment Variables

The current code reads these environment variables:

```env
# Application
APP_ENV=development
APP_PORT=3010
APP_NAME=MailForge

# MySQL Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=yourpassword
DB_NAME=mailforge_db
DB_CHARSET=utf8mb4

# Optional DSN override. If set, this takes precedence over the DB_* fields above.
DB_DSN=root:yourpassword@tcp(localhost:3306)/mailforge_db?charset=utf8mb4&parseTime=true

# JWT settings loaded by config, not yet used by implemented auth routes
JWT_SECRET=supersecretkeychangethisinproduction
JWT_EXPIRY=24h

# Email settings loaded by config, not yet used by an email sender
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@mailforge.com
```

## API

The only registered business endpoint right now is:

```http
GET /health
```

Example response:

```json
{
  "status": "ok"
}
```

Unknown routes and unsupported methods return a uniform JSON error response:

```json
{
  "success": false,
  "error": {
    "code": "route_not_found",
    "message": "route not found"
  },
  "request_id": "generated-request-id"
}
```

Campaign, auth, subscriber, list, and organization routes are not registered yet.

## Project Structure

```text
mailForge/
|-- cmd/
|   |-- api/
|   |   +-- main.go
|   +-- migration/
|       +-- main.go
|-- internal/
|   |-- config/
|   |   |-- config.go
|   |   +-- config_test.go
|   |-- constants/
|   |   +-- http_status_code.go
|   |-- database/
|   |   +-- database.go
|   |-- di/
|   |   +-- container.go
|   |-- middleware/
|   |   |-- logger.go
|   |   +-- recoverer.go
|   |-- migrations/
|   |   |-- embed.go
|   |   |-- 000001_create_users_table.up.sql
|   |   |-- 000001_create_users_table.down.sql
|   |   |-- 000002_create_lists_table.up.sql
|   |   |-- 000002_create_lists_table.down.sql
|   |   |-- 000003_create_subscribers_table.up.sql
|   |   |-- 000003_create_subscribers_table.down.sql
|   |   |-- 000004_create_list_subscribers_table.up.sql
|   |   |-- 000004_create_list_subscribers_table.down.sql
|   |   |-- 000005_create_campaigns_table.up.sql
|   |   +-- 000005_create_campaigns_table.down.sql
|   |-- modules/
|   |   |-- auth/
|   |   |   |-- auth.handler.go
|   |   |   |-- auth.repo.go
|   |   |   +-- auth.service.go
|   |   |-- campaign/
|   |   |   |-- campaign.handler.go
|   |   |   |-- campaign.repo.go
|   |   |   +-- campaign.service.go
|   |-- response/
|   |   +-- error.go
|   |-- routes/
|   |   |-- router.go
|   |   +-- router_test.go
|   +-- server/
|       +-- server.go
|-- pkg/
|   +-- logger/
|       +-- logger.go
|-- .air.toml
|-- .env.example
|-- .gitignore
|-- analysis.md
|-- go.mod
|-- go.sum
|-- Makefile
|-- walkthrough.md
+-- README.md
```

## Architecture

The current runtime flow is:

```text
cmd/api/main.go
  -> loads .env
  -> creates Fx app
  -> registers internal/di.NewModules()
  -> invokes server.StartServer
```

The dependency graph currently provides:

- `config.NewInitConfig`
- `logger.New`
- `database.NewDatabase`
- `routes.NewRouter`
- `server.NewServer`
- database shutdown hook

The router currently provides global middleware, the health endpoint, uniform JSON not-found and method-not-allowed handlers, and custom panic recovery. Feature modules are present as folders, but they are not wired into the container or router.

## Database And Migrations

The project uses MySQL through Bun. SQL migrations are embedded from `internal/migrations`.

Run migrations with:

```bash
go run ./cmd/migration up
go run ./cmd/migration down
go run ./cmd/migration status
```

Or use the Makefile:

```bash
make migrate-up
make migrate-down
make migrate-status
```

The schema currently includes migrations for:

- `users`
- `lists`
- `subscribers`
- `list_subscribers`
- `campaigns`

The `list_subscribers` migration direction has been corrected: the up migration creates the join table, and the down migration drops it.

## Makefile

Common commands:

```bash
make dev
make build
make test
make lint
make migrate-up
make migrate-down
make migrate-status
```

The migration targets point to the current `cmd/migration` command.

## Tests

Run all tests with:

```bash
go test ./...
```

Current test coverage includes:

- Config DSN override behavior
- Config DSN construction from split DB fields
- Integer fallback handling for invalid config values
- Health route response
- Uniform JSON 404 response
- Uniform JSON 405 response

## Development Status

MailForge has a stronger backend foundation now, but it is not a complete product API yet. The next useful development steps are:

1. Add Bun model structs for the migrated tables.
2. Implement auth repository, service, handler, and routes.
3. Add password hashing and JWT creation.
4. Add JWT middleware for protected routes.
5. Implement list and subscriber management.
6. Implement campaign CRUD.
7. Add service and handler tests as each module becomes functional.
