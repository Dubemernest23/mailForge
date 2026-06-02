# MailForge Project Analysis

## Overview

MailForge is currently a Go-based backend API project for an email campaign platform. The project is structured around a modular backend architecture, with dependency injection, a MySQL database layer, database migrations, request logging, and an HTTP server.

The codebase is still in an early foundation stage. The application bootstrap, configuration loading, database connection, HTTP server lifecycle, request logger, health route, and migration runner are implemented. The business modules for authentication, campaigns, and organization are present as package scaffolds, but they do not yet contain handlers, services, repositories, models, or route registrations.

The strongest implemented areas so far are:

- Application startup through `cmd/api/main.go`
- Dependency injection through Uber Fx
- Environment-based configuration loading
- MySQL connection setup through Bun
- Structured logging through Zap
- Chi router setup with middleware
- A `/health` endpoint
- Embedded SQL migrations
- A standalone migration command
- Makefile commands for common development workflows

## Current Project Structure

The current repository contains these main areas:

```text
mailForge/
|-- cmd/
|   |-- api/
|   |   +-- main.go
|   +-- migration/
|       +-- main.go
|-- internal/
|   |-- config/
|   |   +-- config.go
|   |-- database/
|   |   +-- database.go
|   |-- di/
|   |   +-- container.go
|   |-- middleware/
|   |   +-- logger.go
|   |-- migrations/
|   |   |-- ebed.go
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
|   |-- routes/
|   |   +-- router.go
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
+-- README.md
```

The README currently describes a broader architecture than what exists in the repository. For example, it mentions `internal/domain`, `internal/dto`, generic `repository` and `service` directories, subscriber handlers, and campaign handlers. The actual codebase is using an `internal/modules/...` structure instead, and most module files are currently placeholders.

## Application Entry Point

The API starts from:

```text
cmd/api/main.go
```

The current entry point does the following:

1. Loads environment variables from `.env` using `github.com/joho/godotenv`.
2. Logs a warning if `.env` is missing and falls back to system environment variables.
3. Creates an Uber Fx application.
4. Registers the dependency graph from `di.NewModules()`.
5. Invokes `server.StartServer`.
6. Runs the Fx application lifecycle.
7. Checks `app.Err()` after `app.Run()`.
8. Exits with status `1` on fatal application error, otherwise exits with status `0`.

This means application wiring is centralized in the dependency injection container rather than manually constructing every dependency in `main.go`.

There is one current working tree change in `cmd/api/main.go`: an older commented-out version of `main()` has been removed. The active runtime behavior is the explicit `fx.New(...); app.Run(); app.Err()` version.

## Dependency Injection

Dependency injection is implemented in:

```text
internal/di/container.go
```

The `NewModules()` function returns an Fx option bundle that provides and invokes the core application dependencies:

- `config.NewInitConfig`
- `provideLogger`
- `database.NewDatabase`
- `routes.NewRouter`
- `server.NewServer`
- `registerDBHooks`

The resulting dependency graph is:

```text
Config
  |-- Logger
  |-- Database
  +-- Server configuration

Logger
  |-- Database logging
  |-- Request middleware logging
  +-- Shutdown logging

Bun DB
  +-- Registered for lifecycle shutdown

Chi Router
  +-- Used by HTTP server

HTTP Server
  +-- Started through Fx lifecycle hook
```

The database connection is also registered with an Fx lifecycle `OnStop` hook. During shutdown, the application logs that the database connection is closing and calls `db.Close()`.

No feature modules are currently registered in the DI container. Auth, campaign, and organization package files exist, but their services, handlers, repositories, and routes are not wired into Fx yet.

## Configuration

Configuration is implemented in:

```text
internal/config/config.go
```

The project uses environment variables as the source of configuration. The `Config` struct groups settings into:

- `Server`
- `Database`
- `Jwt`
- `Email`
- `DB`

### Server Configuration

The server config contains:

```go
type ServerConfig struct {
    AppEnv  string
    AppPort string
    AppName string
}
```

Defaults:

- `APP_ENV`: `development`
- `APP_PORT`: `3010`
- `APP_NAME`: `MailForge`

### Database Configuration

There are two database-related config structs:

```go
type DatabaseConfig struct {
    DSN string
}
```

and:

```go
type DBConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    Name     string
}
```

`DatabaseConfig.DSN` is used by the actual database connection code. It reads from:

```text
DB_DSN
```

The separate `DBConfig` reads host, port, user, password, and database name values:

- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`

Those values are currently used mainly for logging and Makefile/database helper intent, but the active database connection uses `DB_DSN`.

This is important because `.env.example` documents split database variables but does not document `DB_DSN`. Without `DB_DSN`, `database.NewDatabase()` will attempt to open MySQL with an empty DSN.

### JWT Configuration

JWT config contains:

```go
type JwtConfig struct {
    JwtSecret string
    JwtExpiry string
}
```

Defaults:

- `JWT_SECRET`: `your_jwt_secret`
- `JWT_EXPIRY`: `24h`

The `.env.example` file uses `JWT_EXPIRY_HOURS=24`, but the code reads `JWT_EXPIRY`. That mismatch should be resolved once authentication is implemented.

### Email Configuration

Email config contains:

```go
type EmailConfig struct {
    SmtpHost     string
    SmtpPort     int
    SmtpUser     string
    SmtpPassword string
    SmtpFrom     string
}
```

Defaults:

- `SMTP_HOST`: `smtp.gmail.com`
- `SMTP_PORT`: `587`
- `SMTP_USER`: `your-email@gmail.com`
- `SMTP_PASSWORD`: `your-app-password`
- `SMTP_FROM`: `noreply@mailforge.com`

The email settings are loaded but not yet used by an email sending service.

## Database Layer

The database connection is implemented in:

```text
internal/database/database.go
```

The project uses:

- `database/sql`
- MySQL driver: `github.com/go-sql-driver/mysql`
- Bun ORM: `github.com/uptrace/bun`
- Bun MySQL dialect: `github.com/uptrace/bun/dialect/mysqldialect`
- Bun debug query hook in non-production environments

The connection setup does the following:

1. Opens a MySQL connection using `cfg.Database.DSN`.
2. Configures the connection pool.
3. Pings the database to verify connectivity.
4. Wraps the `sql.DB` connection in Bun.
5. Adds Bun query debugging outside production.
6. Logs successful database connection metadata.

Connection pool settings:

- Max open connections: `25`
- Max idle connections: `10`
- Max connection lifetime: `5 minutes`
- Max idle time: `2 minutes`

The connection is closed through an Fx lifecycle hook in `internal/di/container.go`.

## Logging

Logging is implemented in:

```text
pkg/logger/logger.go
```

The project wraps Zap in a custom `Logger` type. The wrapper exposes:

- `Info`
- `Error`
- `Warn`
- `Debug`
- `Fatal`
- `With`
- `Sync`

The logger behaves differently depending on the environment:

- In `production`, it uses JSON logs at `InfoLevel`.
- In other environments, it uses colored console logs at `DebugLevel`.

The logger includes caller information and stack traces for errors.

There is a small text encoding issue in a comment where an em dash appears garbled in the source, but this does not affect runtime behavior.

## HTTP Server

The HTTP server is implemented in:

```text
internal/server/server.go
```

The `NewServer()` function creates an `http.Server` using the configured application port and the Chi router.

Configured server timeouts:

- Read timeout: `15 seconds`
- Write timeout: `15 seconds`
- Idle timeout: `60 seconds`

The server starts through an Fx lifecycle hook:

1. `net.Listen` binds to the configured address.
2. A startup message prints the localhost URL.
3. `srv.Serve(ln)` runs in a goroutine.
4. A shutdown signal watcher also runs in a goroutine.

Shutdown handling listens for:

- `SIGINT`
- `SIGTERM`

When a shutdown signal is received, the server attempts graceful shutdown with a `10 second` timeout.

One implementation detail to watch: the server has both an Fx `OnStop` shutdown hook and a separate signal watcher that calls `srv.Shutdown()` directly. This works as a basic graceful shutdown mechanism, but future cleanup should make sure shutdown is coordinated cleanly through Fx.

## Routing

Routing is implemented in:

```text
internal/routes/router.go
```

The current router uses Chi and registers:

- Chi request ID middleware
- Chi panic recovery middleware
- Custom structured request logging middleware
- `GET /health`

The only active API route today is:

```http
GET /health
```

It returns:

```json
{
  "status": "ok"
}
```

No authentication routes, campaign routes, list routes, subscriber routes, or organization routes are currently registered.

## Request Logging Middleware

Request logging is implemented in:

```text
internal/middleware/logger.go
```

The middleware wraps each response with Chi's `NewWrapResponseWriter`, allowing it to record:

- HTTP method
- URL path
- Query string
- Response status
- Latency in milliseconds
- Bytes written
- Remote IP
- Request ID

Logging level depends on response status:

- `500+`: error log
- `400-499`: warning log
- Everything else: info log

This gives the project useful observability from the beginning, even before the feature modules are fully built.

## Migration System

The migration command is implemented in:

```text
cmd/migration/main.go
```

The embedded migration files are exposed from:

```text
internal/migrations/ebed.go
```

The file name appears to be a typo and likely should be `embed.go`, but the code itself is valid. It embeds all SQL files in `internal/migrations`:

```go
//go:embed *.sql
var SQLMigrations embed.FS
```

The migration command accepts:

```text
up
down
status
```

Behavior:

- `up`: initializes migration tracking and applies pending migrations.
- `down`: rolls back the last migration group.
- `status`: prints each migration and whether it is pending or applied.

The migration command uses `config.NewInitConfig()` and connects using `cfg.Database.DSN`, so it also depends on `DB_DSN`.

## Database Schema Implemented By Migrations

The migration files define the planned data model for users, lists, subscribers, list membership, and campaigns.

### Users Table

Defined in:

```text
internal/migrations/000001_create_users_table.up.sql
```

The `users` table includes:

- Auto-increment numeric `id`
- Public UUID-style `public_id`
- Unique `email`
- `password_hash`
- Email verification fields
- User verification flag
- Role enum
- Status enum
- Last login timestamp
- Failed login attempt counter
- Created and updated timestamps

Roles:

- `user`
- `admin`
- `super_admin`

Statuses:

- `active`
- `suspended`
- `deleted`

Important constraints:

- Unique `public_id`
- Unique email
- Unique verification token

This table is designed to support authentication, email verification, roles, account state, and login security tracking.

### Lists Table

Defined in:

```text
internal/migrations/000002_create_lists_table.up.sql
```

The `lists` table represents email/contact lists owned by users.

Columns include:

- Auto-increment `id`
- Public UUID-style `public_id`
- `user_id`
- `name`
- Optional `description`
- Status
- Created and updated timestamps

Statuses:

- `active`
- `archived`

Important constraints:

- `user_id` references `users(id)`
- Deleting a user cascades to their lists
- A user cannot have two lists with the same name
- Index on `user_id`

### Subscribers Table

Defined in:

```text
internal/migrations/000003_create_subscribers_table.up.sql
```

The `subscribers` table represents individual contacts owned by a user.

Columns include:

- Auto-increment `id`
- Public UUID-style `public_id`
- `user_id`
- `email`
- Optional `name`
- Status
- Created and updated timestamps

Statuses:

- `active`
- `unsubscribed`
- `bounced`
- `complained`

Important constraints:

- `user_id` references `users(id)`
- Deleting a user cascades to their subscribers
- A user cannot have duplicate subscribers with the same email
- Index on `email`

There is a commented-out index on `user_id`. Since `user_id` is used for ownership lookup, it may be useful to restore that index later.

### List Subscribers Join Table

Migration files:

```text
internal/migrations/000004_create_list_subscribers_table.up.sql
internal/migrations/000004_create_list_subscribers_table.down.sql
```

The intended table appears to be `list_subscribers`, a many-to-many join table between lists and subscribers.

However, the current migration files are reversed:

- The `.up.sql` file drops `list_subscribers`.
- The `.down.sql` file creates `list_subscribers`.

That means running migrations upward will not create the join table. Rolling the migration down would create it, which is the opposite of the expected behavior.

The table definition currently stored in the down migration includes:

- `list_id`
- `subscriber_id`
- Membership status
- Created and updated timestamps
- Composite primary key on `(list_id, subscriber_id)`
- Foreign key to `lists(id)`
- Foreign key to `subscribers(id)`
- Index on `subscriber_id`

Membership statuses:

- `subscribed`
- `unsubscribed`

This should be corrected before relying on migrations in a real environment.

### Campaigns Table

Defined in:

```text
internal/migrations/000005_create_campaigns_table.up.sql
```

The `campaigns` table represents email campaigns owned by users and optionally associated with a list.

Columns include:

- Auto-increment `id`
- `user_id`
- Public UUID-style `public_id`
- Optional `list_id`
- `name`
- `subject`
- `preview_text`
- `body`
- Status
- Scheduled timestamp
- Started timestamp
- Completed timestamp
- Created and updated timestamps

Statuses:

- `draft`
- `scheduled`
- `sending`
- `sent`
- `cancelled`

Important constraints:

- `user_id` references `users(id)`
- Deleting a user cascades to their campaigns
- `list_id` references `lists(id)`
- Deleting a list sets `campaigns.list_id` to `NULL`
- A user cannot have two campaigns with the same name

Indexes:

- `user_id`
- `list_id`
- `status`
- `scheduled_at`

This schema supports drafts, scheduled campaigns, sending progress, and historical campaign records.

## Module Scaffolding

There are module directories for:

- `auth`
- `campaign`
- `organization`

Current files:

```text
internal/modules/auth/auth.handler.go
internal/modules/auth/auth.repo.go
internal/modules/auth/auth.service.go
internal/modules/campaign/campaign.handler.go
internal/modules/campaign/campaign.repo.go
internal/modules/campaign/campaign.service.go
internal/modules/organization/organization.repo.go
```

The campaign files currently contain only:

```go
package campaign
```

The organization file currently contains only:

```go
package organization
```

The auth files currently contain only:

```go
package user
```

The auth package name does not match the folder name. Go allows package names to differ from folder names, but this can become confusing. If this module is meant to be authentication, it should probably use `package auth`; if it is meant to be user/account functionality, the directory may need to be renamed.

No module currently exposes implemented repository methods, service methods, request/response DTOs, or HTTP handlers.

## Build And Development Tooling

The project includes:

- `go.mod`
- `go.sum`
- `.air.toml`
- `Makefile`
- `.env.example`
- `.gitignore`

### Go Module

The module name is:

```text
mailForgeApi
```

The project currently declares:

```text
go 1.26.1
```

This is unusual because Go 1.26.1 is newer than common stable Go versions available in many environments at the time this project appears to have been created. If local tooling does not support this version, builds/tests may fail until the Go version is aligned with the installed toolchain.

Major dependencies include:

- Chi router
- Go MySQL driver
- Godotenv
- Bun ORM
- Bun MySQL dialect
- Bun debug hook
- Uber Fx
- Uber Zap

Most dependencies are marked indirect in `go.mod`, even though several are used directly in code. Running `go mod tidy` with the intended Go toolchain would likely clean this up.

### Makefile

The Makefile provides commands for:

- `make help`
- `make dev`
- `make build`
- `make test`
- `make lint`
- `make clean`
- `make db-create`
- `make migrate-up`
- `make migrate-down`
- `make migrate-status`
- `make db-reset`

Notable current issues:

- The migration command paths in the Makefile use `./cmd/migrate/...`, but the actual directory is `cmd/migration`.
- The README startup command uses `go run ./cmd/mailforge`, but the actual API entry point is `cmd/api`.
- Some text in the Makefile appears garbled due to encoding issues.
- The `clean` target uses Unix-style `rm -rf`, which may not work in plain Windows PowerShell unless a compatible shell is used.

## README State

The README provides a useful high-level description of the intended project, but it is ahead of the actual implementation.

It currently says the project supports:

- Creating, updating, and managing campaigns
- Sending campaign-related requests through a REST API
- Authentication-protected endpoints

In the current codebase, those features are planned but not implemented as API behavior yet. Only the health endpoint is registered.

README mismatches:

- Startup command points to `./cmd/mailforge`, but actual entry point is `./cmd/api`.
- Project structure lists folders that do not exist.
- API overview lists routes that are not registered.
- Environment docs include split DB variables, but the code requires `DB_DSN`.
- README says auth support exists, but auth is currently only scaffolding.

## Current API Surface

The current working API surface is:

```http
GET /health
```

Expected response:

```json
{
  "status": "ok"
}
```

No other endpoints are currently registered.

## What Has Been Done So Far

So far, the project has completed the backend foundation work:

1. A Go module has been initialized.
2. Core dependencies have been added for routing, dependency injection, logging, MySQL, ORM access, environment loading, and migrations.
3. The application entry point has been created.
4. Environment variable loading has been implemented.
5. A configuration model has been created.
6. Uber Fx has been introduced for dependency injection and lifecycle management.
7. A Zap-based logger wrapper has been implemented.
8. A Bun/MySQL database connection has been implemented.
9. Database connection pooling has been configured.
10. Query debugging has been enabled for non-production environments.
11. Database shutdown cleanup has been registered.
12. An HTTP server has been implemented.
13. Server timeouts have been configured.
14. Graceful shutdown handling has been added.
15. A Chi router has been created.
16. Request ID middleware has been added.
17. Panic recovery middleware has been added.
18. Structured request logging middleware has been added.
19. A health check endpoint has been added.
20. SQL migrations have been added for core domain tables.
21. SQL migrations are embedded into the Go binary.
22. A migration CLI has been implemented.
23. Makefile commands have been added for development and database workflows.
24. Module directories have been created for auth, campaigns, and organization.

## What Is Not Implemented Yet

The following areas are not implemented yet:

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
- Request DTOs and response DTOs
- Domain models or Bun models
- Input validation
- Error response conventions
- Authentication-protected routes
- Tests
- OpenAPI/API documentation

## Important Issues To Fix Next

The most important current issues are:

1. Fix the `list_subscribers` migration direction.
   The up migration should create the table, and the down migration should drop it.

2. Align the database configuration.
   Either document and use `DB_DSN` everywhere, or build the DSN from `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, and `DB_NAME`.

3. Fix Makefile migration paths.
   The Makefile currently references `cmd/migrate`, but the actual folder is `cmd/migration`.

4. Fix README command paths and architecture description.
   The README should match the current `cmd/api` entry point and `internal/modules` structure.

5. Decide the auth package name.
   The `internal/modules/auth` files currently use `package user`, which should be clarified.

6. Add actual module implementations.
   The route layer currently has no auth, campaign, list, or subscriber endpoints.

7. Add tests.
   There are currently no `_test.go` files in the repository.

8. Review the Go version.
   `go.mod` declares `go 1.26.1`, which may not match available Go toolchains.

## Suggested Next Development Order

A practical next sequence would be:

1. Fix migration/config/tooling mismatches.
2. Add Bun model structs for users, lists, subscribers, list subscribers, and campaigns.
3. Implement auth repository and service.
4. Implement registration and login endpoints.
5. Add JWT middleware.
6. Implement list and subscriber management.
7. Implement campaign CRUD.
8. Add campaign scheduling/sending later, once campaign storage is stable.
9. Add tests around config, repositories, services, and handlers.

That order keeps the foundation stable before building higher-level campaign behavior.

## Summary

MailForge currently has a solid backend skeleton. The project can be described as an early-stage Go API foundation for an email campaign system. The infrastructure pieces are mostly in place: configuration, logging, dependency injection, database connection, HTTP server, request middleware, health checking, and migrations.

The feature layer is not yet built. Auth, campaigns, subscribers, lists, and organization behavior are represented mainly by database migrations and empty module files. The next major milestone is to turn those module placeholders into real repositories, services, handlers, and route registrations.
