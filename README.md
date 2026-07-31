# MailForge

MailForge is a Go backend for an email campaign platform. The current project state has the backend foundation in place plus the Phase B authentication slice: registration, login, refresh-token rotation, logout, RS256 JWT signing/verification, Redis-backed refresh tokens, MySQL persistence, and Swagger/OpenAPI documentation.

## Tech Stack

- Go 1.26.1
- Chi for HTTP routing
- Uber Fx for dependency injection and lifecycle wiring
- Bun ORM with MySQL
- Redis for refresh-token storage
- RS256 JWT access tokens
- Zap for structured logging
- Godotenv for local environment loading
- Swag plus Swagger UI for API documentation
- Docker Compose for local MySQL, test MySQL, Redis, and MailHog

## Current Features

- API startup from `cmd/api/main.go`
- Fx dependency graph in `internal/di`
- Environment configuration from `.env` or system variables
- MySQL setup through Bun
- Redis client setup
- Embedded SQL migrations and migration runner
- Request logging, request IDs, panic recovery, JSON 404/405 responses
- `GET /health`
- Auth endpoints:
  - `POST /auth/register`
  - `POST /auth/login`
  - `POST /auth/refresh`
  - `POST /auth/logout`
- Auth validation DTOs and shared JSON error responses
- Password hashing with bcrypt
- RS256 access token generation and middleware verification
- Refresh-token issue, rotation, and revocation through Redis
- Swagger UI in non-production environments at `/swagger/index.html`
- OpenAPI JSON at `/swagger/doc.json`

## Not Implemented Yet

These areas are still planned or scaffolded:

- Campaign HTTP endpoints
- Subscriber and list HTTP endpoints
- Organization/moderator behavior
- Email sending workflow
- Background workers
- Tracking endpoints
- Production deployment configuration

## Getting Started

1. Install Go 1.26.1 or a compatible local toolchain.
2. Create a `.env` file from `.env.example`.
3. Start local services:

```bash
make docker-up
```

4. Generate local JWT keys:

```bash
make gen-keys
```

5. Run migrations:

```bash
make migrate-up
```

6. Run the API:

```bash
go run ./cmd/api
```

By default, the API runs on `http://localhost:3010`.

## Environment Variables

The example local environment is in `.env.example`.

```env
APP_ENV=development
APP_PORT=3010
APP_NAME=MailForge

DB_HOST=localhost
DB_PORT=3308
DB_USER=root
DB_PASSWORD=secret
DB_NAME=mailforge

REDIS_URL=redis://localhost:6379

JWT_PRIVATE_KEY_PATH=keys/private.pem
JWT_PUBLIC_KEY_PATH=keys/public.pem
JWT_ACCESS_EXPIRY=1h
JWT_REFRESH_EXPIRY=7d

SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_USER=
SMTP_PASSWORD=
SMTP_FROM=noreply@mailforge.com

EMAIL_PROVIDER=smtp
WORKER_POOL_SIZE=5
DEV_APP_BASE_URL=http://localhost:3010
PROD_APP_BASE_URL=https://mailforge.tech
```

The Docker Compose file exposes:

- MySQL app database on port `3308`
- MySQL test database on port `3307`
- Redis on port `6379`
- MailHog SMTP on port `1025`
- MailHog UI on `http://localhost:8025`

## API

Health check:

```http
GET /health
```

Auth:

```http
POST /auth/register
POST /auth/login
POST /auth/refresh
POST /auth/logout
```

Swagger:

```http
GET /swagger/index.html
GET /swagger/doc.json
```

Swagger is only mounted when `APP_ENV` is not `production`.

## Project Structure

```text
mailForge/
|-- cmd/
|   |-- api/                  # API entrypoint
|   |-- migration/            # Migration CLI
|   +-- redischeck/           # Redis connectivity helper
|-- docs/                     # Product and phase planning docs
|-- internal/
|   |-- config/               # App config, expiry parsing, JWT key loading
|   |-- database/             # Bun/MySQL setup
|   |-- di/                   # Fx providers and module wiring
|   |-- docs/                 # Generated Swagger/OpenAPI files
|   |-- middleware/           # Logging, recovery, JWT middleware
|   |-- migrations/           # Embedded SQL migrations and runner
|   |-- models/               # Bun model structs
|   |-- modules/
|   |   |-- auth/             # Auth DTOs, handler, service, repository
|   |   +-- campaign/         # Campaign scaffold
|   |-- redisclient/          # Redis client provider
|   |-- routes/               # Chi router and route registration
|   |-- server/               # HTTP server lifecycle
|   |-- shared/
|   |   |-- apperrors/        # Shared service/domain errors
|   |   |-- constants/        # HTTP status constants
|   |   |-- http/             # Request decode/validate helpers
|   |   +-- response/         # JSON response and error helpers
|   +-- testUtils/            # DB, Redis, key, and fixture test helpers
|-- pkg/
|   |-- logger/               # Logger wrapper
|   +-- token/                # JWT and refresh-token helpers
|-- docker-compose.yml
|-- Makefile
|-- go.mod
+-- README.md
```

## Runtime Flow

```text
cmd/api/main.go
  -> loads .env
  -> creates Fx app
  -> registers internal/di.NewModules()
  -> invokes server.StartServer
```

The dependency graph currently provides config, logger, database, Redis, RSA keys, refresh-token manager, auth repository/service/handler, router, server, and shutdown hooks.

## Swagger Docs

Swagger is generated from the comments in `cmd/api/main.go` and the auth handler annotations in `internal/modules/auth/auth.handler.go`.

Regenerate docs after changing handler annotations or DTO shapes:

```bash
make swag
```

Check that committed docs match generated output:

```bash
make swag-check
```

Run the API and confirm Swagger in a browser:

```text
http://localhost:3010/swagger/index.html
```

The OpenAPI JSON should be available at:

```text
http://localhost:3010/swagger/doc.json
```

## Database And Migrations

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

Current migrations cover users, lists, subscribers, list subscribers, campaigns, send jobs, tracking events, and moderator permissions.

## Makefile

Common commands:

```bash
make docker-up
make docker-down
make gen-keys
make dev
make build
make test
make lint
make tidy
make migrate-up
make migrate-down
make migrate-status
make swag
make swag-check
```

## Tests

Run all tests:

```bash
make test
```

The full test suite expects Docker services to be running, including MySQL test database on port `3307` and Redis on port `6379`.

Current coverage includes config parsing, JWT key loading, JWT generation/validation, refresh-token behavior, middleware behavior, shared responses, auth DTO validation, auth service/repository behavior, auth route integration, and router behavior.

## PR Confirmation Checklist

Before opening or merging the PR, confirm:

```bash
make docker-up
make gen-keys
make migrate-up
make swag-check
make test
make build
```

Then start the API:

```bash
go run ./cmd/api
```

Manual checks:

- Visit `http://localhost:3010/health` and confirm `{"status":"ok"}`.
- Visit `http://localhost:3010/swagger/index.html` and confirm the Swagger UI loads.
- Open `http://localhost:3010/swagger/doc.json` and confirm it includes `openapi: 3.0.0`.
- In Swagger UI, confirm the auth endpoints are listed under the `auth` tag.
- Use `POST /auth/register` with a valid body and confirm a `201` response with `access_token`, `refresh_token`, and `expires_in`.
- Use `POST /auth/login` with the same credentials and confirm a `200` response.
- Use `POST /auth/refresh` with the refresh token and confirm a new token pair.
- Use `POST /auth/logout` and confirm `204 No Content`.
