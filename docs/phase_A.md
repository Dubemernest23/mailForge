# MailForge — Phase A: Foundation & Infrastructure

> **Phase:** A  
> **Status:** In Progress  
> **Owner:** Engineering Team  
> **Tech Lead:** LOL  
> **Product Lead:** Duby  
> **Estimated Duration:** 3–5 days  
> **Prerequisite:** None — this is the starting point  
> **Next Phase:** Phase B (Authentication & Identity)

---

## Why This Phase Exists

Before a single feature is built, the ground it stands on must be solid. Phase A is not glamorous work — there are no endpoints, no business logic, no user-facing features. But it is the most important phase in the entire project.

Here is why:

- **Broken migrations mean broken onboarding.** If `make migrate-up` fails, no developer can run the project. Every hour spent debugging setup is an hour not spent building.
- **Wrong package names cause compilation errors.** They are easy to fix now and cascading to fix later once other packages import them.
- **No Docker means "works on my machine."** One developer using MySQL 5.7 and another using 8.0 will produce different behaviour. Docker eliminates this class of problem entirely.
- **No CI means bugs reach main.** Every merged PR that breaks something costs more to fix than it would have cost to catch in 30 seconds of automated tests.
- **No domain models means no foundation.** Every repository, service, and handler in every future phase depends on these structs. Writing them correctly once — with proper Bun tags, nullability, and field types — means nothing above them gets retrofitted. Writing them wrong means cascading changes across the entire codebase.

Complete Phase A properly and every subsequent phase will be faster, safer, and more predictable. Cut corners here and you will pay for it in every phase that follows.

---

## Scope

### In Scope
- Fix all five legacy blocking issues in the existing codebase
- Docker Compose setup for MySQL, Redis, and MailHog
- Bun domain model structs for all current tables
- Makefile with all required dev workflow targets
- `.env.example` fully documented
- GitHub Actions CI pipeline
- RSA key generation utility (`make gen-keys`)
- Remove empty organisation scaffold
- README updated to reflect reality

### Out of Scope
- Any business logic
- Any HTTP routes beyond `/health`
- Auth, JWT, or middleware
- `send_jobs`, `tracking_events` models — these are added in their respective phases alongside the code that uses them, not speculatively

---

## Task Breakdown

---

### Task 1 — Fix the Five Legacy Blocking Issues

These are pre-existing issues in the codebase that will prevent compilation, migration, or correct behaviour. They must all be fixed in a single PR before anything else proceeds. One engineer owns this.

---

#### Issue 1: Inverted Migration for `list_subscribers`

**Problem:** The `list_subscribers` migration has its files swapped. The `.up.sql` file drops the table and the `.down.sql` file creates it. Running `migrate-up` will attempt to drop a table that doesn't exist yet. Running `migrate-down` will attempt to create a table that already exists.

**Why this matters:** Migrations must be deterministic and reversible. An inverted migration means a developer following the standard onboarding process will hit an error on their very first `make migrate-up`. It also means the migration history is semantically wrong — any automation that reads migration intent will be confused.

**Fix:** Swap the file contents. The `.up.sql` file should contain the `CREATE TABLE` statement. The `.down.sql` file should contain the `DROP TABLE IF EXISTS` statement.

**How to verify:** Run `make migrate-up` on a clean database. It should succeed with no errors. Then run `make migrate-down` and back up again. Both directions should work cleanly.

---

#### Issue 2: DB DSN Configuration Mismatch

**Problem:** The application code reads a single `DB_DSN` environment variable to construct the MySQL connection string. However, `.env.example` documents individual split variables (`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`) that the code never reads. The result is that every developer who follows `.env.example` will configure variables that have no effect, and the application will fail to connect to the database with a confusing error.

**Why this matters:** Configuration is the contract between the application and its environment. If the documented contract doesn't match the code, developers waste time debugging something that should be transparent. This is especially damaging for onboarding — it's the first thing a new developer does.

**Fix:** Remove the single `DB_DSN` approach. Update `config.go` to read the individual variables (`DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`) and construct the DSN programmatically:

```go
// In config.go
func (c *DatabaseConfig) DSN() string {
    return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4",
        c.User, c.Password, c.Host, c.Port, c.Name,
    )
}
```

**Why `parseTime=true`?** Without this, MySQL's `DATETIME` columns are returned as raw byte strings, not `time.Time` values. Bun cannot map them correctly. This flag must always be present.

**Why `charset=utf8mb4`?** `utf8` in MySQL is actually a 3-byte encoding that cannot handle emoji or certain Unicode characters. `utf8mb4` is proper 4-byte UTF-8. Campaign bodies will contain arbitrary user content — we must support the full Unicode range.

**How to verify:** Set the individual variables in `.env`, start the application, and confirm it connects to MySQL and the `/health` endpoint returns 200.

---

#### Issue 3: Makefile References Wrong Migration Path

**Problem:** The Makefile has a target that references `cmd/migrate` but the actual directory in the project is `cmd/migration`. Running `make migrate-up` will fail with a "no such file or directory" error.

**Why this matters:** The Makefile is the team's single interface for common operations. If it's broken, developers work around it, run commands manually, and diverge from each other. The Makefile must be trusted.

**Fix:** Find every reference to `cmd/migrate` in the Makefile and change it to `cmd/migration`.

**How to verify:** `make migrate-up` runs without a path error.

---

#### Issue 4: Wrong Package Name in Auth Module

**Problem:** The files inside `internal/modules/auth/` declare `package user` at the top. The directory is `auth`, and Go convention requires the package name to match the directory name. Any code that imports this package using the `auth` path will get `user` as the package identifier, which is confusing and inconsistent.

**Why this matters:** Go package naming is a contract. When you write `import "mailforge/internal/modules/auth"`, you expect to use `auth.Something`, not `user.Something`. Mismatches confuse IDEs, confuse developers reading the code, and can cause subtle bugs when two packages both expose a `user` identifier.

**Fix:** Change `package user` to `package auth` in every file inside `internal/modules/auth/`.

**How to verify:** `go build ./...` produces no package name errors. The import path resolves correctly.

---

#### Issue 5: Invalid Go Version in `go.mod`

**Problem:** `go.mod` declares `go 1.26.1`, which does not exist. Go versions follow a predictable pattern: 1.21, 1.22, 1.23, etc. A non-existent version declaration can cause toolchain errors depending on the Go toolchain version installed on the developer's machine.

**Why this matters:** `go.mod` is authoritative. If the declared version doesn't match any real Go release, different developers may get different toolchain resolution behaviour, and CI may fail on Go version checks.

**Fix:** Change the Go version in `go.mod` to match the actual installed toolchain. Check with `go version` in the terminal and align `go.mod` to it. Use `1.22.x` or `1.23.x` depending on what's installed. Then run `go mod tidy` to clean up.

**How to verify:** `go mod tidy` runs without errors. `go build ./...` succeeds.

---

### Task 2 — Remove the Organisation Scaffold

**Problem:** There is an empty `internal/modules/organization/` directory in the project. MailForge has no organisation model, no org tables, and no org functionality. It is not in the PRD. Leaving it creates confusion — future developers (or future us) will wonder if it should be wired up.

**Why this matters:** Dead code and dead directories are noise. They suggest incomplete features, confuse onboarding, and sometimes get partially wired up by mistake. Remove it now while the codebase is small.

**Fix:** Delete `internal/modules/organization/` entirely. Grep the codebase for any remaining references (`organization`, `Organisation`) and remove them — particularly in the Fx DI container and the router if they were ever referenced there.

**How to verify:** `grep -r "organization" ./internal` returns no results. `go build ./...` still succeeds.

---

### Task 3 — Docker Compose

**Why Docker Compose?** Every developer needs MySQL, Redis, and a fake SMTP server to run MailForge locally. Without Docker, each developer installs and configures these independently — different versions, different defaults, different OS-specific behaviours. The result is "works on my machine" bugs that are invisible in code review and unpredictable in CI. Docker Compose gives every developer an identical environment that starts with a single command.

**What we're running:**

| Service | Image | Purpose | Local Port |
|---|---|---|---|
| `mysql` | `mysql:8.0` | Primary database (dev) | 3306 |
| `mysql_test` | `mysql:8.0` | Isolated test database | 3307 |
| `redis` | `redis:7-alpine` | Job queue + refresh token store | 6379 |
| `mailhog` | `mailhog/mailhog` | Fake SMTP — catches emails locally | 1025 (SMTP), 8025 (UI) |

**Why a separate test database?** Running tests against the dev database is dangerous — a test that truncates a table will wipe your dev data. A dedicated `mailforge_test` database means tests can create, modify, and delete data freely without affecting the working dev environment.

**Why MailHog?** During development, we do not want to send real emails to real inboxes. MailHog acts as an SMTP server that catches everything and displays it in a browser UI at `http://localhost:8025`. Point your SMTP config at `localhost:1025` and every email your workers send appears there instantly. No real emails leave the machine.

**Create `docker-compose.yml` in the project root:**

```yaml
version: "3.9"

services:
  mysql:
    image: mysql:8.0
    container_name: mailforge_mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: secret
      MYSQL_DATABASE: mailforge
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

  mysql_test:
    image: mysql:8.0
    container_name: mailforge_mysql_test
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: secret
      MYSQL_DATABASE: mailforge_test
    ports:
      - "3307:3306"
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: mailforge_redis
    restart: unless-stopped
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  mailhog:
    image: mailhog/mailhog
    container_name: mailforge_mailhog
    restart: unless-stopped
    ports:
      - "1025:1025"
      - "8025:8025"

volumes:
  mysql_data:
```

**How to verify:** `docker compose up -d` starts all four containers without errors. `docker compose ps` shows all four as healthy. `http://localhost:8025` opens the MailHog UI.

---

### Task 4 — Domain Models

**Why now?** Every repository, service, and handler in every future phase imports these structs. They are the shared language of the entire codebase. Writing them correctly in Phase A means nothing above them gets retrofitted. The earlier a bad field type or missing Bun tag is caught, the cheaper it is to fix.

**Where they live:** `internal/models/` — one file per model.

**General rules for every model:**
- Use `bun:"table:..."` to declare the table name explicitly — never rely on Bun's default pluralisation
- Use `bun:"column:..."` where the field name doesn't exactly match the column name
- Use `time.Time` for non-nullable timestamps; `sql.NullTime` for nullable timestamps
- Every model has `ID`, `PublicID`, `CreatedAt`, `UpdatedAt`
- No JSON presentation logic in models — that belongs in DTOs
- No HTTP-specific concerns in models

---

#### `internal/models/user.go`

```go
package models

import "time"

type User struct {
    bun.BaseModel `bun:"table:users,alias:u"`

    ID           uint64    `bun:"id,pk,autoincrement"`
    PublicID     string    `bun:"public_id,notnull"`
    Name         string    `bun:"name,notnull"`
    Email        string    `bun:"email,notnull"`
    PasswordHash string    `bun:"password_hash,notnull"`
    Role         string    `bun:"role,notnull,default:'user'"`
    Status       string    `bun:"status,notnull,default:'active'"`
    CreatedAt    time.Time `bun:"created_at,notnull"`
    UpdatedAt    time.Time `bun:"updated_at,notnull"`
}
```

**Why `Role` as a `string` and not a Go `type`?** We will define a typed constant set in the auth module (`const RoleUser = "user"` etc.) but the model itself stays as `string` to remain a pure data struct with no business logic dependencies. The service layer enforces role validity.

---

#### `internal/models/list.go`

```go
package models

import "time"

type List struct {
    bun.BaseModel `bun:"table:lists,alias:l"`

    ID          uint64    `bun:"id,pk,autoincrement"`
    PublicID    string    `bun:"public_id,notnull"`
    UserID      uint64    `bun:"user_id,notnull"`
    Name        string    `bun:"name,notnull"`
    Description string    `bun:"description"`
    Status      string    `bun:"status,notnull,default:'active'"`
    CreatedAt   time.Time `bun:"created_at,notnull"`
    UpdatedAt   time.Time `bun:"updated_at,notnull"`

    // Relation (populated only when explicitly joined — never auto-loaded)
    User *User `bun:"rel:belongs-to,join:user_id=id"`
}
```

**Why declare the relation but not auto-load?** Bun does not auto-load relations by default — you must call `.Relation("User")` explicitly. Declaring the relation struct tag allows us to load it when we need it (e.g., admin views) without paying the JOIN cost on every query.

---

#### `internal/models/subscriber.go`

```go
package models

import "time"

type Subscriber struct {
    bun.BaseModel `bun:"table:subscribers,alias:s"`

    ID        uint64    `bun:"id,pk,autoincrement"`
    PublicID  string    `bun:"public_id,notnull"`
    UserID    uint64    `bun:"user_id,notnull"`
    Name      string    `bun:"name,notnull"`
    Email     string    `bun:"email,notnull"`
    Status    string    `bun:"status,notnull,default:'subscribed'"`
    CreatedAt time.Time `bun:"created_at,notnull"`
    UpdatedAt time.Time `bun:"updated_at,notnull"`

    User  *User  `bun:"rel:belongs-to,join:user_id=id"`
    Lists []List `bun:"m2m:list_subscribers,join:Subscriber=List"`
}
```

---

#### `internal/models/list_subscriber.go`

```go
package models

import "time"

type ListSubscriber struct {
    bun.BaseModel `bun:"table:list_subscribers,alias:ls"`

    ID           uint64    `bun:"id,pk,autoincrement"`
    ListID       uint64    `bun:"list_id,notnull"`
    SubscriberID uint64    `bun:"subscriber_id,notnull"`
    CreatedAt    time.Time `bun:"created_at,notnull"`

    List       *List       `bun:"rel:belongs-to,join:list_id=id"`
    Subscriber *Subscriber `bun:"rel:belongs-to,join:subscriber_id=id"`
}
```

**Why does this model exist when it's just a join table?** Bun requires an explicit model struct for many-to-many join tables when you want to query the relationship directly (e.g., "get all subscribers in list X"). Without it, you can only query through the owning models. We will need direct join table access frequently.

---

#### `internal/models/campaign.go`

```go
package models

import (
    "database/sql"
    "time"
)

type Campaign struct {
    bun.BaseModel `bun:"table:campaigns,alias:c"`

    ID          uint64       `bun:"id,pk,autoincrement"`
    PublicID    string       `bun:"public_id,notnull"`
    UserID      uint64       `bun:"user_id,notnull"`
    ListID      sql.NullInt64 `bun:"list_id"`
    Name        string       `bun:"name,notnull"`
    Subject     string       `bun:"subject,notnull"`
    PreviewText string       `bun:"preview_text"`
    HtmlBody    string       `bun:"html_body"`
    PlainBody   string       `bun:"plain_body"`
    Status      string       `bun:"status,notnull,default:'draft'"`
    ScheduledAt sql.NullTime `bun:"scheduled_at"`
    SentAt      sql.NullTime `bun:"sent_at"`
    CreatedAt   time.Time    `bun:"created_at,notnull"`
    UpdatedAt   time.Time    `bun:"updated_at,notnull"`

    User *User `bun:"rel:belongs-to,join:user_id=id"`
    List *List `bun:"rel:belongs-to,join:list_id=id"`
}
```

**Why `sql.NullInt64` for `ListID`?** A campaign can exist in draft state without a list assigned yet — the user might set the list later before sending. A nullable foreign key models this correctly. `sql.NullInt64` is the idiomatic Go type for nullable integers that map to database columns; it has a `Valid` boolean field that tells you whether the value is present.

**Why `sql.NullTime` for `ScheduledAt` and `SentAt`?** These are only set at specific points in the campaign lifecycle. A new draft has no scheduled time and no sent time. Nullable columns + `sql.NullTime` models this accurately.

---

### Task 5 — Makefile

**Why a Makefile?** The Makefile is the team's shared command vocabulary. Instead of everyone remembering long `go` commands with specific flags and environment variables, the Makefile provides short, memorable, consistent targets. It documents how to operate the project.

**Add or update the following targets in `Makefile`:**

```makefile
.PHONY: dev build test lint migrate-up migrate-down gen-keys docker-up docker-down

# Start the API server
dev:
	go run ./cmd/api

# Build the binary
build:
	go build -o bin/mailforge ./cmd/api

# Run all tests with race detector and coverage
test:
	go test ./... -race -cover -count=1

# Run the linter
lint:
	golangci-lint run ./...

# Run migrations UP against the dev database
migrate-up:
	go run ./cmd/migration up

# Run migrations DOWN (one step)
migrate-down:
	go run ./cmd/migration down

# Start Docker services
docker-up:
	docker compose up -d

# Stop Docker services
docker-down:
	docker compose down

# Generate RSA key pair for JWT signing
gen-keys:
	@mkdir -p keys
	@openssl genrsa -out keys/private.pem 2048
	@openssl rsa -in keys/private.pem -pubout -out keys/public.pem
	@echo "Keys generated: keys/private.pem (private) and keys/public.pem (public)"
	@echo "IMPORTANT: Never commit keys/private.pem to git"

# Tidy Go modules
tidy:
	go mod tidy
```

**Why `-count=1` on tests?** By default, Go caches test results. If the code hasn't changed, `go test` returns the cached result instead of running the tests again. `-count=1` disables caching and forces a fresh run every time. This is important because tests that hit a real database may produce different results even if the code is unchanged.

**Why `gen-keys` in the Makefile?** RSA key generation requires `openssl`, which every developer has but whose exact flags nobody remembers. Encoding it in the Makefile means one command, same result, every time.

---

### Task 6 — `.env.example`

**Why a documented `.env.example`?** This file is the contract between the application and its environment. Every variable the app reads must be documented here with a safe example value and a comment explaining what it does. A developer cloning the project for the first time should be able to run `cp .env.example .env` and be 90% of the way to a working setup.

```env
# ─── Application ────────────────────────────────────────────────────────────
APP_ENV=development
APP_PORT=3010
APP_NAME=MailForge

# ─── Database ────────────────────────────────────────────────────────────────
# These are combined in config.go to build the MySQL DSN.
# Run `make docker-up` to start a MySQL container with these defaults.
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=secret
DB_NAME=mailforge

# ─── Redis ───────────────────────────────────────────────────────────────────
# Used for the job queue (asynq) and refresh token storage.
# Run `make docker-up` to start a Redis container.
REDIS_URL=redis://localhost:6379

# ─── JWT (RS256) ─────────────────────────────────────────────────────────────
# Paths to RSA key files. Run `make gen-keys` to generate a fresh pair.
# NEVER commit keys/private.pem to git.
JWT_PRIVATE_KEY_PATH=./keys/private.pem
JWT_PUBLIC_KEY_PATH=./keys/public.pem
JWT_ACCESS_EXPIRY=1h
JWT_REFRESH_EXPIRY=168h

# ─── SMTP ────────────────────────────────────────────────────────────────────
# In development, point this at MailHog (started by `make docker-up`).
# MailHog catches all emails — view them at http://localhost:8025
SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_USER=
SMTP_PASSWORD=
SMTP_FROM=noreply@mailforge.com
SMTP_TLS=false

# ─── Email Provider ──────────────────────────────────────────────────────────
# Which EmailProvider implementation to use. Options: smtp
# (resend, ses, mailgun will be added in later phases)
EMAIL_PROVIDER=smtp

# ─── Workers ─────────────────────────────────────────────────────────────────
# Number of concurrent email delivery workers.
WORKER_POOL_SIZE=5

# ─── App Base URL ────────────────────────────────────────────────────────────
# Used to construct tracking links embedded in campaign emails.
APP_BASE_URL=http://localhost:3010
```

---

### Task 7 — GitHub Actions CI

**Why CI?** CI (Continuous Integration) means every push to any branch automatically runs the tests. This is the safety net that catches bugs before they reach `main`. Without it, broken code can be merged without anyone noticing until it causes a problem in production or blocks another developer.

**Cost:** GitHub Actions is free for public repositories. For private repositories, GitHub provides 2,000 minutes/month on the free plan. For a project at this stage with a small team, this is more than enough. No credit card required.

**What the pipeline does:**
1. Starts MySQL and Redis as service containers (same images as Docker Compose)
2. Runs database migrations against the test DB
3. Lints the code (`golangci-lint`)
4. Runs all tests with the race detector enabled

**Why use service containers in CI instead of mocking?** Because we test against real infrastructure. Our tests hit real MySQL and real Redis. The CI environment must match. Service containers in GitHub Actions are Docker containers that start alongside the job — they are available to the test runner at `localhost` on their mapped ports.

**Create `.github/workflows/ci.yml`:**

```yaml
name: CI

on:
  push:
    branches: ["**"]
  pull_request:
    branches: [main]

jobs:
  test:
    name: Lint & Test
    runs-on: ubuntu-latest

    services:
      mysql:
        image: mysql:8.0
        env:
          MYSQL_ROOT_PASSWORD: secret
          MYSQL_DATABASE: mailforge_test
        ports:
          - 3306:3306
        options: >-
          --health-cmd="mysqladmin ping -h localhost"
          --health-interval=10s
          --health-timeout=5s
          --health-retries=5

      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379
        options: >-
          --health-cmd="redis-cli ping"
          --health-interval=10s
          --health-timeout=5s
          --health-retries=5

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod
          cache: true

      - name: Install dependencies
        run: go mod download

      - name: Run migrations
        run: go run ./cmd/migration up
        env:
          DB_HOST: 127.0.0.1
          DB_PORT: 3306
          DB_USER: root
          DB_PASSWORD: secret
          DB_NAME: mailforge_test
          APP_ENV: test

      - name: Lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest

      - name: Test
        run: go test ./... -race -cover -count=1
        env:
          APP_ENV: test
          DB_HOST: 127.0.0.1
          DB_PORT: 3306
          DB_USER: root
          DB_PASSWORD: secret
          DB_NAME: mailforge_test
          REDIS_URL: redis://localhost:6379
          JWT_ACCESS_EXPIRY: 1h
          JWT_REFRESH_EXPIRY: 168h
          APP_BASE_URL: http://localhost:3010
```

**Why `cache: true` on the Go setup step?** Go module downloads are slow. The cache stores the module download directory between runs. Subsequent CI runs skip the download step for unchanged dependencies, cutting pipeline time significantly.

**Why `go-version-file: go.mod`?** Instead of hardcoding a Go version string in the CI file, we tell the action to read it from `go.mod`. When we upgrade Go, we update `go.mod` once and CI picks it up automatically. No drift.

---

### Task 8 — `.gitignore` Additions

Make sure these entries exist in `.gitignore`:

```gitignore
# Environment
.env
.env.local

# RSA keys — private key must never be committed
keys/private.pem

# Binary output
bin/

# IDE
.idea/
.vscode/
*.swp

# OS
.DS_Store
```

**Why explicitly gitignore `keys/private.pem` but not `keys/public.pem`?** The public key is safe to commit — it can only verify tokens, not sign them. The private key is the signing secret. If it leaks, an attacker can forge any JWT for any user. Explicitly ignoring it in `.gitignore` adds a safety net against an accidental `git add .`.

---

## Deliverables Checklist

The following must all be true before Phase A is considered complete and Phase B can begin.

| # | Deliverable | How to Verify |
|---|---|---|
| 1 | All five legacy blocking issues fixed | `go build ./...` produces no errors |
| 2 | `go.mod` has a valid Go version | `go mod tidy` succeeds |
| 3 | Organisation scaffold removed | `grep -r "organization" ./internal` returns nothing |
| 4 | `docker-compose.yml` present and working | `docker compose up -d` starts all 4 containers healthy |
| 5 | `make migrate-up` works against Dockerised MySQL | Runs with no errors on a clean database |
| 6 | `make migrate-down` works | Reverses migrations cleanly |
| 7 | All domain models exist in `internal/models/` | Files: user.go, list.go, subscriber.go, list_subscriber.go, campaign.go |
| 8 | Models compile with correct Bun tags | `go build ./...` passes |
| 9 | `.env.example` documents every variable | All variables in reference section are documented |
| 10 | Makefile targets all work | `make dev`, `make test`, `make lint`, `make gen-keys`, `make docker-up` all run |
| 11 | `make gen-keys` produces key pair | `keys/private.pem` and `keys/public.pem` exist after running |
| 12 | `keys/private.pem` is gitignored | `git status` does not show private.pem as a tracked file |
| 13 | `.github/workflows/ci.yml` exists | File is present and valid YAML |
| 14 | CI passes on push | GitHub Actions run shows green on a test branch push |
| 15 | `README.md` reflects reality | Documents actual entry point, structure, env vars, how to run |

---

## Acceptance Criteria (Phase Complete When)

- `git clone` → `cp .env.example .env` → `make docker-up` → `make migrate-up` → `make dev` works end-to-end for a developer who has never seen the project before
- `go build ./...` — no errors
- `go test ./...` — passes (no tests yet, but must not error on compilation)
- CI pipeline on GitHub Actions runs green
- No compilation warnings, no dead imports, no references to removed modules

---

## What Phase B Will Build On

When Phase A is complete, Phase B (Authentication & Identity) will:
- Import `models.User` from `internal/models/user.go`
- Expect MySQL to be running and migrated (via Docker Compose)
- Expect Redis to be running (for refresh token storage)
- Use the RSA key pair generated by `make gen-keys`
- Run its tests against the `mailforge_test` database
- Have its PR validated automatically by the CI pipeline

Every decision made in Phase A directly enables Phase B to move faster and with confidence. Do not rush Phase A.

---

## Notes for the Team

- Do not add `send_jobs`, `dead_letter_jobs`, or `tracking_events` models in this phase. They will be added in Phase E and Phase G respectively, alongside the code that uses them. Adding them now would be speculative and they would have no callers, no tests, and no migration yet.

- If you discover additional issues in the codebase during Phase A work, document them and bring them to the tech lead before fixing them. Do not silently expand scope — if something needs fixing, it gets a task, an owner, and a decision.

- The `keys/` directory should exist in the repo (so `gen-keys` has somewhere to write) but `private.pem` must never be committed. Add a `keys/.gitkeep` to track the directory without tracking its contents.

- Every task in this phase should be done in a single PR if possible. The goal is one clean, reviewable unit of work that moves the project from "broken skeleton" to "solid foundation."