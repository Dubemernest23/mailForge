# MailForge

MailForge is an early-stage email campaign API built in Go. The current codebase focuses on the backend foundation: application bootstrapping, dependency injection, configuration loading, database connectivity, request logging, HTTP server lifecycle management, a health route, and embedded SQL migrations.

The feature modules for authentication, campaigns, and organization are scaffolded, but their business logic and HTTP endpoints are not implemented yet.

## Tech Stack

- Go
- Chi for HTTP routing
- Uber Fx for dependency injection and application lifecycle
- Bun ORM for MySQL database access
- Zap for structured logging
- Godotenv for local environment loading
- Bun migrations with embedded SQL files

## Current Features

- Application startup from `cmd/api/main.go`
- Environment configuration from `.env` or system environment variables
- Fx dependency container in `internal/di`
- MySQL connection setup through Bun
- Connection pool configuration
- Bun query debugging in non-production environments
- Structured application logger
- Request logging middleware
- Chi router with request IDs and panic recovery
- `GET /health` endpoint
- Graceful HTTP server shutdown
- Embedded SQL migrations
- Standalone migration command in `cmd/migration`

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
- Request and response DTOs
- Domain/Bun model structs
- Tests

## Getting Started

1. Install Go.
2. Create a `.env` file from `.env.example`.
3. Make sure `DB_DSN` is configured, because the current database code connects through `DB_DSN`.
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

# Active database connection value used by the app
DB_DSN=root:yourpassword@tcp(localhost:3306)/mailforge_db?parseTime=true

# Additional database fields currently loaded into config
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=yourpassword
DB_NAME=mailforge_db

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

Note: `.env.example` currently documents some values that do not perfectly match the code yet, such as `JWT_EXPIRY_HOURS` instead of `JWT_EXPIRY`. The app currently reads `JWT_EXPIRY`.

## API

The only registered endpoint right now is:

```http
GET /health
```

Example response:

```json
{
  "status": "ok"
}
```

Campaign, auth, subscriber, list, and organization routes are not registered yet.

## Project Structure

```text
mailForge/
|-- cmd/
|   |-- api/
|   |   +-- main.go                 # API entry point and Fx bootstrap
|   +-- migration/
|       +-- main.go                 # Migration CLI
|-- internal/
|   |-- config/
|   |   +-- config.go               # Environment config loading
|   |-- database/
|   |   +-- database.go             # MySQL/Bun database setup
|   |-- di/
|   |   +-- container.go            # Fx dependency wiring
|   |-- middleware/
|   |   +-- logger.go               # Structured request logger
|   |-- migrations/
|   |   |-- ebed.go                 # Embedded SQL migration filesystem
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
|   |   |   |-- auth.handler.go      # Scaffold only
|   |   |   |-- auth.repo.go         # Scaffold only
|   |   |   +-- auth.service.go      # Scaffold only
|   |   |-- campaign/
|   |   |   |-- campaign.handler.go  # Scaffold only
|   |   |   |-- campaign.repo.go     # Scaffold only
|   |   |   +-- campaign.service.go  # Scaffold only
|   |   +-- organization/
|   |       +-- organization.repo.go # Scaffold only
|   |-- routes/
|   |   +-- router.go               # Chi router and health route
|   +-- server/
|       +-- server.go               # HTTP server lifecycle
|-- pkg/
|   +-- logger/
|       +-- logger.go               # Zap logger wrapper
|-- .air.toml
|-- .env.example
|-- .gitignore
|-- analysis.md
|-- go.mod
|-- go.sum
|-- Makefile
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

The router currently provides global middleware and the health endpoint only. Feature modules are present as folders, but they are not wired into the container or router.

## Database And Migrations

The project uses MySQL through Bun. SQL migrations are embedded from `internal/migrations`.

Run the migration command directly with:

```bash
go run ./cmd/migration up
go run ./cmd/migration down
go run ./cmd/migration status
```

The schema currently includes migrations for:

- `users`
- `lists`
- `subscribers`
- `list_subscribers`
- `campaigns`

Important current migration note: the `000004_create_list_subscribers_table` migration appears to be reversed. The `.up.sql` file currently drops the table, while the `.down.sql` file creates it. That should be fixed before relying on migrations in a real database.

## Makefile

The Makefile includes commands for development, build, tests, linting, cleanup, database creation, migrations, and database reset.

Common commands:

```bash
make dev
make build
make test
make lint
```

Current note: the migration targets in the Makefile reference `./cmd/migrate/...`, but the actual migration command lives at `./cmd/migration`. Use the direct `go run ./cmd/migration ...` commands until the Makefile is corrected.

## Development Status

MailForge currently has a solid backend skeleton, but not a complete product API. The next useful development steps are:

1. Fix migration/config/tooling mismatches.
2. Add Bun model structs.
3. Implement auth repository, service, handler, and routes.
4. Add JWT middleware.
5. Implement list and subscriber management.
6. Implement campaign CRUD.
7. Add tests.
