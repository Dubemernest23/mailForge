# MailForge Project Analysis

## Overview

MailForge is a Go backend API foundation for an email campaign platform. The project is still early-stage, but the core infrastructure is now in better shape: application startup, dependency injection, configuration, database connectivity, migrations, logging, routing, uniform HTTP errors, and focused tests are all present.

The application currently exposes only the health endpoint. Auth, campaign, and organization modules are scaffolded, but they do not yet contain working repositories, services, handlers, or route registrations.

## What Is Implemented

- API entry point in `cmd/api/main.go`
- Migration CLI in `cmd/migration/main.go`
- Environment config loading in `internal/config`
- Database DSN building from split `DB_*` values, with optional `DB_DSN` override
- MySQL/Bun database connection setup
- Connection pool tuning
- Bun query debugging outside production
- Uber Fx dependency wiring
- HTTP server lifecycle setup
- Graceful server shutdown handling
- Zap-based application logger
- Structured request logging middleware
- Custom panic recovery middleware
- Shared HTTP status constants
- Uniform JSON error response helpers
- Chi router with request IDs
- `GET /health`
- JSON `404 route_not_found`
- JSON `405 method_not_allowed`
- Embedded SQL migrations
- Makefile commands for build, test, and migrations
- Tests for config and router behavior

## Current Project Structure

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
|   |   +-- organization/
|   |       +-- organization.repo.go
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
|-- go.mod
|-- go.sum
|-- Makefile
|-- README.md
+-- walkthrough.md
```

## Application Flow

The API starts from:

```text
cmd/api/main.go
```

Startup flow:

1. `.env` is loaded with `godotenv`.
2. The Fx app is created.
3. `di.NewModules()` registers providers.
4. `server.StartServer` is invoked.
5. Fx starts the HTTP server lifecycle.

The current dependency graph provides:

- `config.NewInitConfig`
- `logger.New`
- `database.NewDatabase`
- `routes.NewRouter`
- `server.NewServer`
- database shutdown hook

## Configuration

Configuration lives in:

```text
internal/config/config.go
```

The config groups settings into:

- `Server`
- `Database`
- `DB`
- `Jwt`
- `Email`

The database configuration now supports both modes:

1. If `DB_DSN` is set, it is used directly.
2. If `DB_DSN` is empty, the app builds a MySQL DSN from:
   - `DB_HOST`
   - `DB_PORT`
   - `DB_USER`
   - `DB_PASSWORD`
   - `DB_NAME`
   - `DB_CHARSET`

This fixes the previous mismatch where `.env.example` documented split database values but the application only used `DB_DSN`.

`JWT_EXPIRY` is now documented consistently in `.env.example`; the older `JWT_EXPIRY_HOURS` mismatch has been removed.

## Database Layer

Database setup lives in:

```text
internal/database/database.go
```

The code:

- Opens a MySQL connection.
- Verifies connectivity with `Ping`.
- Wraps the connection with Bun.
- Sets connection pool limits.
- Enables Bun debug query logging outside production.

Connection settings:

- Max open connections: `25`
- Max idle connections: `10`
- Max connection lifetime: `5 minutes`
- Max idle time: `2 minutes`

The database is closed during Fx shutdown.

## Routing And Error Handling

Routing lives in:

```text
internal/routes/router.go
```

Current route:

```http
GET /health
```

Health response:

```json
{
  "status": "ok"
}
```

The router now uses:

- Chi request ID middleware
- Structured request logger
- Custom panic recoverer
- JSON not-found handler
- JSON method-not-allowed handler

Shared HTTP constants live in:

```text
internal/constants/http_status_code.go
```

Uniform response/error helpers live in:

```text
internal/response/error.go
```

Error response shape:

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

The response package also provides `AppError`, `NewAppError`, `WrapAppError`, and `HandleError` for future handlers/services.

## Migrations

Migrations are embedded from:

```text
internal/migrations/embed.go
```

The migration CLI supports:

```bash
go run ./cmd/migration up
go run ./cmd/migration down
go run ./cmd/migration status
```

The schema migrations currently cover:

- `users`
- `lists`
- `subscribers`
- `list_subscribers`
- `campaigns`

The previous `list_subscribers` issue has been fixed. The up migration now creates the join table and the down migration drops it.

The migration embed file has also been corrected to `embed.go`, replacing the earlier typo-like `ebed.go` filename.

## Module Scaffolding

Module folders exist for:

- `auth`
- `campaign`
- `organization`

Auth package naming has been corrected. Files under `internal/modules/auth` now use:

```go
package auth
```

Campaign files use:

```go
package campaign
```

These modules are still scaffolds only. They are not wired into Fx or the router yet.

## Tooling

The Makefile now points at the real migration command:

```bash
go run ./cmd/migration up
go run ./cmd/migration down
go run ./cmd/migration status
```

Common commands:

```bash
make build
make test
make migrate-up
make migrate-down
make migrate-status
```

The Makefile text has also been cleaned to plain ASCII.

`go.mod` now separates direct dependencies from indirect dependencies after `go mod tidy`.

The local Go toolchain reports:

```text
go version go1.26.1 windows/amd64
```

That matches the `go 1.26.1` directive in `go.mod`.

## Tests

Tests now exist for:

- Explicit `DB_DSN` override
- DSN construction from split database fields
- Fallback behavior for invalid integer environment values
- `GET /health`
- Uniform JSON 404 response
- Uniform JSON 405 response

Test files:

```text
internal/config/config_test.go
internal/routes/router_test.go
```

## Fixed Issues From The Previous Analysis

The previous analysis listed several important issues. Current status:

1. `list_subscribers` migration direction: fixed.
2. Database config mismatch: fixed.
3. Makefile migration paths: fixed.
4. README command paths and architecture: updated.
5. Auth package name mismatch: fixed.
6. No tests: partially fixed with foundation tests.
7. Go version uncertainty: resolved locally; installed Go is `go1.26.1`.

The remaining larger item is actual module implementation. Auth, campaigns, lists, subscribers, and organization behavior are still not built.

## What Is Not Implemented Yet

- User registration
- User login
- Password hashing
- JWT creation
- JWT validation middleware
- Email verification flow
- Campaign CRUD endpoints
- Campaign repository methods
- Campaign service logic
- Campaign sending logic
- Subscriber CRUD endpoints
- List CRUD endpoints
- List/subscriber membership management
- Organization logic
- Email provider integration
- SMTP sending
- Feature DTOs
- Bun model structs
- Input validation for feature routes
- Auth-protected routes
- OpenAPI/API documentation

## Suggested Next Things To Do

1. Add Bun model structs for `users`, `lists`, `subscribers`, `list_subscribers`, and `campaigns`.
2. Implement auth repository methods for creating users and finding users by email/public ID.
3. Implement password hashing and password verification.
4. Implement auth service methods for registration and login.
5. Add auth handler DTOs and routes.
6. Add JWT creation and JWT middleware.
7. Wire auth dependencies into `internal/di/container.go`.
8. Add list and subscriber repositories/services/handlers.
9. Add campaign CRUD repository/service/handler.
10. Add tests per module as each feature becomes real.

## Summary

MailForge now has a cleaner and more reliable backend foundation than before. The important infrastructure issues from the previous analysis have been addressed: migrations are corrected, config is aligned with `.env.example`, Makefile paths are fixed, auth package naming is consistent, direct dependencies are tidied, and foundation tests exist.

The next major milestone is implementing the first real product module, with auth being the natural starting point because later list, subscriber, and campaign endpoints will need authenticated ownership.
