# MailForge — Product Definition & Requirements (PDR)

> **Document Owner:** Product & Engineering Lead  
> **Status:** Approved — Source of Truth  
> **Last Updated:** July 2026  
> **Rule:** Every phase, every PR, every decision traces back to this document. If something in the codebase contradicts this PDR, the PDR wins — update the code, not the PDR. If a product decision changes, update this document first, then the code.

---

## Table of Contents

1. [Product Definition](#1-product-definition)
2. [Target Users](#2-target-users)
3. [Architecture Overview](#3-architecture-overview)
4. [Technology Stack](#4-technology-stack)
5. [Locked Architectural Decisions](#5-locked-architectural-decisions)
6. [Domain Model](#6-domain-model)
7. [Authentication & Identity](#7-authentication--identity)
8. [Admin System](#8-admin-system)
9. [Feature Scope](#9-feature-scope)
10. [Phase Plan](#10-phase-plan)
11. [API Reference by Phase](#11-api-reference-by-phase)
12. [Infrastructure & DevOps](#12-infrastructure--devops)
13. [CI/CD Pipeline](#13-cicd-pipeline)
14. [Testing Strategy](#14-testing-strategy)
15. [Cross-Cutting Standards](#15-cross-cutting-standards)
16. [Environment Variables Reference](#16-environment-variables-reference)
17. [Decision Log](#17-decision-log)

---

## 1. Product Definition

MailForge is a **self-serve email campaign REST API** built for individuals and small business owners. It gives users the ability to manage subscribers, build email lists, create and send campaigns, track engagement, and review analytics — all through a clean, well-documented API.

**One user. One account. No organisations. No teams.**

Every resource in the system (lists, subscribers, campaigns) is owned by exactly one user. There is no sharing, delegation, or multi-tenancy at the user level. Complexity that serves enterprise teams is deliberately excluded from scope.

**What MailForge is:**
- A campaign management and delivery engine
- An engagement tracking system (opens, clicks, unsubscribes)
- A lightweight analytics layer per campaign
- An operationally observable platform via an admin system

**What MailForge is not:**
- A drag-and-drop email builder (no frontend in MVP)
- A CRM or contact intelligence tool
- A multi-tenant SaaS with org hierarchies
- A transactional email service (it sends campaigns, not individual triggered emails to app users)

---

## 2. Target Users

### End Users (API Consumers)
People who register and use the platform to send campaigns:

- **Individuals** — sending event invites, personal announcements, reminders
- **Small business owners** — tailors, bakers, service providers sending product updates, promotions, seasonal offers
- **Freelancers and creators** — communicating with their audience on a budget

Their common trait: they have a list of people who want to hear from them, and they need a reliable way to send to that list without paying for an enterprise tool.

### Platform Operators (Admin Users)
Internal team members who oversee the platform:

- **Super Admin** — full platform visibility, manages moderators and their permissions
- **Moderator** — supervised operator, works within permissions granted by super admin

---

## 3. Architecture Overview

MailForge is built around an **event-driven, job-queue architecture**. HTTP handlers are intentionally thin: they validate input, persist state to MySQL, and enqueue work. They never do heavy lifting inline. Workers run in a separate goroutine pool, consuming jobs from Redis and doing the actual email delivery.

```
┌─────────────────────────────────────────────────────────────────┐
│                        Client Apps                              │
│                  Web / Mobile / API Consumers                   │
└───────────────────────────┬─────────────────────────────────────┘
                            │ HTTPS + RS256 JWT
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│                      MailForge API                              │
│                  Go + Chi Router + Uber Fx                      │
│                                                                 │
│  ┌─────────────┐  ┌──────────────────┐  ┌──────────────────┐   │
│  │ Auth Module │  │ Contacts Module  │  │ Campaign Module  │   │
│  │  Register   │  │ Lists/Subscribers│  │  Create/Manage   │   │
│  │  Login      │  │  CSV Import      │  │  Send/Schedule   │   │
│  │  Refresh    │  │                  │  │                  │   │
│  └─────────────┘  └──────────────────┘  └──────────────────┘   │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    Admin Module                          │   │
│  │         Super Admin + Moderator + Permissions            │   │
│  └──────────────────────────────────────────────────────────┘   │
└───────────────────────────┬─────────────────────────────────────┘
                            │
              ┌─────────────┴───────────────┐
              ▼                             ▼
┌─────────────────────┐         ┌───────────────────────┐
│   MySQL (Primary)   │         │   Redis (Queue)       │
│   Bun ORM           │         │   asynq library       │
│   Persistent data   │         │   Job queue + DLQ     │
│   Analytics source  │         │   Refresh tokens      │
│   Audit trail       │         │   Scheduled jobs      │
└─────────────────────┘         └────────────┬──────────┘
                                             │
                                             ▼
                                ┌───────────────────────┐
                                │    Worker Pool        │
                                │  Goroutine pool (Fx)  │
                                │  Consumes asynq jobs  │
                                └────────────┬──────────┘
                                             │
                              ┌──────────────┼──────────────┐
                              ▼              ▼              ▼
                        ┌──────────┐  ┌──────────┐  ┌──────────┐
                        │ Success  │  │ Retry    │  │  DLQ     │
                        │ → MySQL  │  │ (≤ 3×)   │  │ (Redis + │
                        │ updated  │  │          │  │  MySQL)  │
                        └──────────┘  └──────────┘  └──────────┘
                                             │
                                             ▼
                              ┌───────────────────────────┐
                              │      Email Provider       │
                              │  EmailProvider interface  │
                              │                           │
                              │  SMTPProvider (default)   │
                              │  ResendProvider (later)   │
                              └───────────────────────────┘
```

**Why this architecture?**

Sending an email campaign to thousands of subscribers cannot happen synchronously inside an HTTP request. If it did, the request would time out, retries would be impossible, failures would be invisible, and the API would be unusable. Decoupling the HTTP layer from delivery via a job queue means:

- The API responds in milliseconds regardless of list size
- Failures are retried automatically without user intervention
- Every job is observable — the admin can see what's pending, processing, failed, or in the DLQ
- The delivery engine can scale independently of the API

---

## 4. Technology Stack

| Layer | Technology | Why |
|---|---|---|
| Language | Go (1.22+) | Performance, concurrency model, excellent stdlib |
| HTTP Router | Chi | Lightweight, idiomatic, composable middleware |
| Dependency Injection | Uber Fx | Structured DI, clean lifecycle management |
| ORM | Bun | Thin, fast, idiomatic Go, good MySQL support |
| Primary Database | MySQL 8.0 | Relational integrity, widely understood, already in stack |
| Job Queue | Redis + asynq | Native scheduling, retry policies, DLQ, built-in inspector UI |
| Auth Tokens | RS256 JWT | Asymmetric signing — private key signs, public key verifies |
| Password Hashing | bcrypt (cost 12) | Industry baseline for security vs performance |
| Logging | Zap | Structured, performant |
| Email Delivery | go-mail + provider interface | TLS handling, clean API, swappable |
| Testing | Go testing + testify | Standard, no magic |
| Containerisation | Docker + Docker Compose | Consistent local and CI environments |
| CI | GitHub Actions | Free for public repos, 2,000 min/month free for private |

---

## 5. Locked Architectural Decisions

These decisions are final for MVP. They may be revisited after MVP ships. Any proposal to change them requires updating this PDR first.

### 5.1 MySQL is the persistent record, Redis is the queue

MySQL holds all business data: users, lists, subscribers, campaigns, send job records (for analytics and audit), tracking events. Nothing that needs to survive a Redis restart lives only in Redis.

Redis holds the active job queue (via asynq), scheduled job references, and refresh tokens. These are transient or reconstructable.

**The send_jobs table in MySQL is not the queue — it is the audit log.** When a campaign is sent, we write one row per subscriber to `send_jobs` (status: pending). We then enqueue the job in Redis. The worker updates the MySQL row when it completes. This gives us analytics (how many delivered, failed, bounced) without relying on Redis for historical data.

### 5.2 EmailProvider is an interface

```go
type EmailProvider interface {
    Send(ctx context.Context, msg Message) error
    Name() string
}
```

Concrete implementations live in `internal/providers/`. The worker receives an `EmailProvider` injected by Fx. Swapping providers is a config change (`EMAIL_PROVIDER=resend`), not a code change. We ship `SMTPProvider` in Phase E. `ResendProvider` and others are added later without modifying worker logic.

### 5.3 asynq for the job queue

asynq is a Go library built on Redis. It gives us:
- Job enqueueing and processing
- Configurable retry with backoff
- Built-in DLQ (failed queue)
- Native scheduling (delayed/scheduled jobs — this is how scheduled campaigns work)
- An optional web UI (`asynqmon`) for job inspection

This is why scheduling comes at low additional cost — asynq handles "process this job at 8:00 AM tomorrow" natively.

### 5.4 RS256 for JWT auth

RS256 uses an asymmetric key pair: a private key to sign tokens (never leaves the server) and a public key to verify them (can be distributed freely). This is more secure than HS256 (symmetric, single secret) because:
- The signing key and verification key are different
- Future services can verify tokens without being trusted with the signing key
- Key rotation is possible without invalidating all tokens immediately

**Access token:** expires in 1 hour — short-lived, limits blast radius of a leaked token  
**Refresh token:** expires in 7 days — stored in Redis, used only to issue new access tokens  
**Refresh token rotation:** every time a refresh token is used, it is invalidated and a new one is issued

### 5.5 Admin is a role, not a separate user model

Admin users are regular users with an elevated `role` field. There is no separate admin table. The roles are:

```
user          — regular platform user
moderator     — platform operator with configurable permissions
super_admin   — full platform access, manages moderators
```

Role is stored on the `users` table. Moderator permissions are stored in a separate `moderator_permissions` table to allow fine-grained control.

### 5.6 Scheduling is in scope

A campaign can be sent immediately or scheduled for a future datetime. The handler validates the request, writes the campaign to MySQL, and enqueues a scheduled asynq job for the target time. At the scheduled time, asynq triggers the worker, which runs the normal send flow.

### 5.7 Soft deletes for subscribers

Subscribers are never hard-deleted. A "deleted" subscriber has `status = unsubscribed`. This preserves historical analytics: open/click/bounce counts for past campaigns remain accurate. Hard-deleting a subscriber would corrupt these numbers.

---

## 6. Domain Model

All tables live in MySQL. Struct tags use Bun conventions. Every table has `created_at` and `updated_at`. Every public-facing resource has a `public_id` (UUID) — internal `id` (auto-increment) is never exposed in API responses.

### 6.1 users

```sql
CREATE TABLE users (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    public_id     VARCHAR(36) NOT NULL UNIQUE,
    username          VARCHAR(255) NOT NULL UNIQUE,
    email         VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role          ENUM('user', 'moderator', 'super_admin') NOT NULL DEFAULT 'user',
    status        ENUM('active', 'suspended') NOT NULL DEFAULT 'active',
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

**Why `role` here?** Admin is a property of a user, not a separate entity. A user is promoted to moderator or super_admin by a super_admin. Keeping it in one table avoids joins on every authentication check.

**Why `status`?** Admins need to be able to suspend a user without deleting them (preserves their history).

### 6.2 moderator_permissions

```sql
CREATE TABLE moderator_permissions (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    public_id     VARCHAR(36) NOT NULL UNIQUE,
    moderator_id  BIGINT UNSIGNED NOT NULL,
    permission    VARCHAR(100) NOT NULL,
    granted_by    BIGINT UNSIGNED NOT NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (moderator_id) REFERENCES users(id),
    FOREIGN KEY (granted_by)   REFERENCES users(id),
    UNIQUE KEY uq_mod_permission (moderator_id, permission)
);
```

**Why a separate table?** Permissions are many-to-one (one moderator can have multiple permissions). Storing them as a comma-separated string or JSON column would make querying and revoking individual permissions messy. A row per permission is clean, auditable, and easy to query.

**Example permissions:** `view_users`, `suspend_users`, `view_campaigns`, `view_queue`, `replay_jobs`

### 6.3 lists

```sql
CREATE TABLE lists (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    public_id   VARCHAR(36) NOT NULL UNIQUE,
    user_id     BIGINT UNSIGNED NOT NULL,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    status      ENUM('active', 'archived') NOT NULL DEFAULT 'active',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### 6.4 subscribers

```sql
CREATE TABLE subscribers (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    public_id   VARCHAR(36) NOT NULL UNIQUE,
    user_id     BIGINT UNSIGNED NOT NULL,
    name        VARCHAR(255) NOT NULL,
    email       VARCHAR(255) NOT NULL,
    status      ENUM('subscribed', 'unsubscribed') NOT NULL DEFAULT 'subscribed',
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id),
    UNIQUE KEY uq_user_email (user_id, email)
);
```

**Why `UNIQUE KEY uq_user_email`?** A user cannot have the same email address appear twice in their subscriber pool. The unique constraint is scoped to `user_id` — two different users can have the same subscriber email, but one user cannot duplicate it.

### 6.5 list_subscribers

```sql
CREATE TABLE list_subscribers (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    list_id       BIGINT UNSIGNED NOT NULL,
    subscriber_id BIGINT UNSIGNED NOT NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (list_id)       REFERENCES lists(id),
    FOREIGN KEY (subscriber_id) REFERENCES subscribers(id),
    UNIQUE KEY uq_list_subscriber (list_id, subscriber_id)
);
```

### 6.6 campaigns

```sql
CREATE TABLE campaigns (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    public_id     VARCHAR(36) NOT NULL UNIQUE,
    user_id       BIGINT UNSIGNED NOT NULL,
    list_id       BIGINT UNSIGNED,
    name          VARCHAR(255) NOT NULL,
    subject       VARCHAR(500) NOT NULL,
    preview_text  VARCHAR(500),
    html_body     LONGTEXT,
    plain_body    LONGTEXT,
    status        ENUM('draft', 'scheduled', 'sending', 'sent', 'cancelled') NOT NULL DEFAULT 'draft',
    scheduled_at  DATETIME NULL,
    sent_at       DATETIME NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (list_id) REFERENCES lists(id)
);
```

**Campaign state machine:**
```
draft → sending     (immediate send: user hits send now)
draft → scheduled   (scheduled send: user sets a future time)
scheduled → sending (asynq fires the job at scheduled_at)
sending → sent      (all jobs delivered or exhausted)
draft → cancelled   (user cancels before sending)
```

**Why `html_body` AND `plain_body`?** SMTP supports multipart/alternative emails — HTML version for email clients that render it, plain text fallback for clients that do not. We store both and send both. Never send HTML-only.

### 6.7 send_jobs

```sql
CREATE TABLE send_jobs (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    public_id     VARCHAR(36) NOT NULL UNIQUE,
    campaign_id   BIGINT UNSIGNED NOT NULL,
    subscriber_id BIGINT UNSIGNED NOT NULL,
    status        ENUM('pending', 'processing', 'delivered', 'failed') NOT NULL DEFAULT 'pending',
    attempts      TINYINT UNSIGNED NOT NULL DEFAULT 0,
    last_error    TEXT,
    processed_at  DATETIME NULL,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    FOREIGN KEY (campaign_id)   REFERENCES campaigns(id),
    FOREIGN KEY (subscriber_id) REFERENCES subscribers(id)
);
```

**Why this table if Redis/asynq is the queue?** This is the **audit log and analytics source**, not the queue. The queue tells workers what to do next. This table tells the analytics layer how many emails were sent, delivered, failed, and bounced for a given campaign. It persists forever. Redis does not.

### 6.8 tracking_events

```sql
CREATE TABLE tracking_events (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    public_id     VARCHAR(36) NOT NULL UNIQUE,
    campaign_id   BIGINT UNSIGNED NOT NULL,
    subscriber_id BIGINT UNSIGNED NOT NULL,
    event_type    ENUM('open', 'click', 'unsubscribe') NOT NULL,
    metadata      JSON,
    ip_address    VARCHAR(45),
    user_agent    TEXT,
    created_at    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (campaign_id)   REFERENCES campaigns(id),
    FOREIGN KEY (subscriber_id) REFERENCES subscribers(id)
);
```

**Why `metadata JSON`?** Click events need to store the original URL that was clicked. Open events need nothing extra. Unsubscribe events may store the list context. JSON is flexible enough to handle per-event-type payloads without a separate column per type.

---

## 7. Authentication & Identity

### 7.1 Overview

MailForge uses **RS256 JWT** with a two-token model:

| Token | Algorithm | Expiry | Purpose |
|---|---|---|---|
| Access token | RS256 | 1 hour | Authenticate API requests |
| Refresh token | Opaque (UUID) | 7 days | Obtain new access tokens |

**Why RS256 over HS256?**

HS256 uses a single secret for both signing and verification — any service that can verify a token can also forge one. RS256 uses a private key (signing, never leaves the auth service) and a public key (verification only, safe to distribute). This is the correct approach for a system that may eventually have multiple services or expose a public JWKS endpoint.

**Why two tokens?**

Access tokens are short-lived (1 hour) to limit the damage window if one is intercepted. But asking a user to log in every hour is terrible UX. The refresh token (7 days) solves this: the client silently exchanges it for a new access token. If the refresh token is compromised, it can be revoked from Redis without touching access token signing keys.

### 7.2 Token Payload

**Access token claims:**
```json
{
  "sub": "user-public-uuid",
  "role": "user",
  "iat": 1234567890,
  "exp": 1234571490
}
```

**Why `role` in the token?** So middleware can enforce role-based access control without a database query on every request. The role is validated once at login and embedded in the token.

### 7.3 Refresh Token Storage

Refresh tokens are stored in Redis with a key pattern of `refresh:<token_uuid>` and a TTL of 7 days.

**Why Redis, not MySQL?** Refresh tokens are session data. They are high-write (created on every login, deleted on logout or rotation), short-lived, and do not need relational queries. Redis with TTL handles this perfectly — expired tokens self-delete, no cleanup job needed.

**Refresh token rotation:** Every time a refresh token is used, the old token is deleted from Redis and a new one is issued. This means a refresh token can only be used once. If an attacker steals and uses a refresh token before the legitimate user does, the legitimate user's next attempt will fail (the token is gone), alerting them that their session was compromised.

### 7.4 Key Management

```
/keys/
  private.pem    — RSA private key (never committed to git, loaded from env or file mount)
  public.pem     — RSA public key (can be committed, used for verification)
```

In development, keys are generated once and added to `.env`. In production, they are injected as environment variables or mounted as files. The Makefile will include a `make gen-keys` command to generate a fresh key pair.

### 7.5 Auth Endpoints

```
POST /auth/register     — create account, return access + refresh tokens
POST /auth/login        — verify credentials, return access + refresh tokens
POST /auth/refresh      — exchange refresh token for new access + refresh tokens
POST /auth/logout       — revoke refresh token from Redis
```

### 7.6 Middleware

```
JWTMiddleware       — validates access token on every protected route, injects userID + role into context
AdminMiddleware     — checks role = moderator | super_admin (used on /admin/* routes)
SuperAdminMiddleware — checks role = super_admin only
```

---

## 8. Admin System

### 8.1 Two-Tier Model

**Super Admin**
- Created directly in the database (seeded at deployment)
- Can view everything across the platform: all users, all campaigns, queue state, delivery stats
- Can promote a user to moderator
- Can grant or revoke individual permissions from any moderator
- Can suspend any user (except other super admins)
- Sees a log of all moderator actions

**Moderator**
- Is a regular user promoted by a super admin
- Has only the permissions explicitly granted to them
- All their actions are timestamped and visible to super admins
- Cannot grant permissions to other users

### 8.2 Permission Catalogue

| Permission Key | What it allows |
|---|---|
| `view_users` | Read all user records |
| `suspend_users` | Suspend or reactivate user accounts |
| `view_campaigns` | Read all campaigns across all users |
| `view_queue` | View the asynq job queue and DLQ |
| `replay_jobs` | Re-enqueue failed jobs from the DLQ |
| `view_analytics` | Platform-wide delivery and engagement stats |

### 8.3 Admin Endpoints

All admin routes are under `/admin/` and require a valid access token with role `moderator` or `super_admin`. Super-admin-only routes are noted.

```
# User management
GET    /admin/users                          — list all users (paginated)
GET    /admin/users/:userId                  — get user details
PATCH  /admin/users/:userId/suspend          — suspend user [requires: suspend_users]
PATCH  /admin/users/:userId/reactivate       — reactivate user [requires: suspend_users]

# Moderator management (super admin only)
POST   /admin/moderators                     — promote user to moderator [super_admin]
DELETE /admin/moderators/:userId             — demote moderator to user [super_admin]
POST   /admin/moderators/:userId/permissions — grant permission [super_admin]
DELETE /admin/moderators/:userId/permissions/:permission — revoke permission [super_admin]
GET    /admin/moderators/:userId/permissions — list moderator's permissions [super_admin]

# Campaign oversight
GET    /admin/campaigns                      — list all campaigns [requires: view_campaigns]
GET    /admin/campaigns/:campaignId          — campaign detail [requires: view_campaigns]

# Queue oversight
GET    /admin/queue/stats                    — queue stats (pending, active, failed) [requires: view_queue]
GET    /admin/queue/failed                   — list failed/DLQ jobs [requires: view_queue]
POST   /admin/queue/failed/:jobId/replay     — re-enqueue failed job [requires: replay_jobs]

# Platform analytics
GET    /admin/analytics/overview             — platform-wide delivery stats [requires: view_analytics]
```

---

## 9. Feature Scope

### In MVP

| Feature | Notes |
|---|---|
| User registration and login | RS256 JWT, access + refresh tokens |
| Email list management | Create, read, update, archive |
| Subscriber management | Manual add, CSV bulk import, soft delete |
| Campaign CRUD | Draft state, full editing before send |
| Immediate send | Enqueues jobs for all subscribed contacts |
| Scheduled send | Uses asynq native scheduling |
| Email delivery via SMTP | Via `EmailProvider` interface, SMTP implementation |
| Retry logic | Up to 3 retries per job, exponential backoff |
| Dead letter queue | Failed jobs after 3 retries move to DLQ |
| Open tracking | 1×1 pixel, first open per subscriber per campaign |
| Click tracking | Link rewriting + redirect, records click event |
| Unsubscribe handling | Signed token link, flips subscriber status |
| Per-campaign analytics | Sent, delivered, opens, clicks, bounces, unsubscribes, rates |
| Admin — super admin | Full platform oversight |
| Admin — moderator | Permission-gated operator access |
| Docker Compose | MySQL + Redis + MailHog for local dev |
| GitHub Actions CI | Tests + lint on every PR |

### Explicitly Deferred (Post-MVP)

| Feature | Reason deferred |
|---|---|
| Resend / SES / Mailgun providers | Interface is built; implementation is a later sprint |
| Frontend / dashboard | API-first MVP |
| Webhook system | Post-MVP integration feature |
| CSV export | Nice to have, not critical path |
| Template gallery | UX enhancement, not core |
| Advanced analytics (time-series) | Requires aggregation tables, deferred for scale |
| Email verification on register | Not in PRD; add post-MVP if spam becomes a problem |
| Rate limiting per user | Useful, but out of scope for early-stage build |

---

## 10. Phase Plan

Each phase has a single clear goal. No phase begins until the previous phase's acceptance criteria are fully met and all tests pass. Each phase will have its own dedicated PDR section reviewed before work starts.

---

### Phase A — Foundation & Infrastructure

**Goal:** The project boots cleanly. Every developer can clone, run `make dev`, and have a working local environment with MySQL and Redis running in Docker. All five legacy blocking issues are fixed. The domain models exist. CI runs on every push.

**Why first:** Nothing else can be built on a broken foundation. If migrations are wrong, models are wrong, or the project doesn't boot, every phase after this is blocked or building on sand. Get this right once and it pays dividends for the entire build.

**Tasks:**
- Fix all five legacy blocking issues (inverted migration, DSN config, Makefile path, wrong package name, invalid go version)
- Set up Docker Compose (MySQL 8.0, Redis 7, MailHog)
- Write `Makefile` targets: `dev`, `migrate-up`, `migrate-down`, `test`, `gen-keys`, `lint`
- Create `internal/models/` with all Bun model structs: `User`, `List`, `Subscriber`, `ListSubscriber`, `Campaign`
- Note: `SendJob` and `TrackingEvent` models added in their respective phases — not speculatively
- Remove empty `internal/modules/organization/` scaffold
- Write `docker-compose.yml` for local dev
- Write `.env.example` with all variables documented
- Configure GitHub Actions CI workflow (lint + test on push and PR)
- Update README: actual entry point, structure, env vars, how to run

**Deliverables:**
- `docker-compose.yml`
- `internal/models/*.go` (user, list, subscriber, list_subscriber, campaign)
- `.github/workflows/ci.yml`
- Updated `README.md`
- Updated `.env.example`
- Fixed Makefile

**Acceptance Criteria:**
- `docker compose up -d` starts MySQL and Redis without errors
- `make migrate-up` runs all migrations cleanly against the Dockerised MySQL
- `go build ./...` produces no errors
- `go test ./...` passes (even if no tests yet — must not error)
- No references to `organization` remain in active code
- CI workflow runs green on a push to a test branch

---

### Phase B — Authentication & Identity

**Goal:** A user can register, log in, receive RS256 access and refresh tokens, and use those tokens to access protected routes. Token refresh and logout work correctly.

**Why second:** Every subsequent endpoint in the system is protected. Building lists, campaigns, or anything else without auth means every route is open, and you will spend the next five phases retrofitting middleware and ownership checks. Auth is the gate — build the gate before building anything behind it.

**New migrations:** None (users table exists from Phase A models)

**Tasks:**
1. Generate RSA key pair (`make gen-keys` → writes `keys/private.pem` and `keys/public.pem`)
2. Load keys from config in Fx bootstrap
3. Build auth DTOs: `RegisterRequest`, `LoginRequest`, `RefreshRequest`, `AuthResponse`
4. Build auth repository: `FindByEmail`, `CreateUser`, `FindByID`
5. Build auth service: `Register` (bcrypt hash, create user, sign tokens), `Login` (verify hash, sign tokens), `Refresh` (validate refresh token in Redis, rotate), `Logout` (delete refresh token from Redis)
6. Build JWT middleware: validates access token, injects `userID` and `role` into request context
7. Build auth handler: thin, calls service, returns response DTO
8. Register routes in router (auth routes are public; all other routes require JWTMiddleware)
9. Write unit tests for all service methods
10. Write integration tests for all four endpoints

**Acceptance Criteria:**
- `POST /auth/register` creates a user and returns `access_token` + `refresh_token`
- `POST /auth/login` with correct credentials returns tokens; with wrong credentials returns 401
- `POST /auth/refresh` with valid refresh token returns new token pair and invalidates old refresh token
- `POST /auth/logout` invalidates the refresh token
- Accessing a protected route without a token returns 401
- Accessing a protected route with an expired token returns 401
- All service unit tests pass
- All handler integration tests pass

---

### Phase C — Lists & Subscribers

**Goal:** Authenticated users can create and manage email lists, add subscribers (manually or via CSV), and manage list membership.

**Why third:** Campaigns target lists. A campaign without a list is incomplete. We need real lists and subscribers before the campaign layer has any data to work with.

**Tasks:**
1. Build List: repository, service, handler, DTOs for full CRUD
2. Build Subscriber: repository, service, handler, DTOs for full CRUD + CSV import
3. Build list membership endpoints (add/remove subscriber from list)
4. CSV import: parse in service layer, validate rows, batch insert in a transaction, return summary (inserted, skipped, invalid)
5. All queries scoped to authenticated user's `user_id` at repository signature level
6. Write unit tests for all service methods
7. Write integration tests for all endpoints

**Key implementation rules:**
- Repository methods must accept `userID` as a parameter — they must not infer it from context. This is explicit and auditable.
- Subscriber deletion is a status flip, never a `DELETE` statement
- Duplicate email within a user's subscriber pool returns `409 Conflict`
- CSV import returns `202 Accepted` + a summary object (not a job — the import is fast enough to be synchronous unless list > 10,000 rows, which we defer)

**Acceptance Criteria:**
- Full list CRUD works, all responses scoped to the authenticated user
- Full subscriber CRUD works
- CSV import with 1,000 rows completes and returns a summary
- A subscriber can be in multiple lists
- Accessing another user's list returns 403
- All tests pass

---

### Phase D — Campaigns (CRUD)

**Goal:** Authenticated users can create, read, update, and delete campaigns in draft state.

**Why before send:** The send pipeline is the most complex part of the system. Separating campaign storage from delivery means we can validate the full campaign lifecycle (draft creation, editing, readiness checks) before introducing async complexity. Never mix persistence with delivery.

**Tasks:**
1. Build Campaign: repository, service, handler, DTOs for full CRUD
2. Enforce state machine: only `draft` campaigns can be edited or deleted
3. Validate that `list_id` on a campaign belongs to the authenticated user
4. Write unit tests for all service methods
5. Write integration tests for all endpoints

**Acceptance Criteria:**
- Campaign can be created in draft state
- Campaign can be updated while in draft state
- Campaign cannot be updated or deleted once it is in `sending` or `sent` state
- Assigning another user's list to a campaign returns 403
- All tests pass

---

### Phase E — Send Engine

**Goal:** User can trigger an immediate send. The system enqueues one job per subscriber, workers deliver via SMTP, retries happen automatically, failures land in the DLQ. MySQL is updated throughout as the audit log.

**Why event-driven:** See Architecture Overview section 3. Summary: HTTP handlers must never block on email delivery.

**New migrations:** `send_jobs` table

**New code:**
- `SendJob` model in `internal/models/`
- `internal/providers/` package with `EmailProvider` interface
- `internal/providers/smtp.go` — `SMTPProvider` implementation
- `internal/workers/` package — asynq worker setup, job handlers
- Fx lifecycle hook to start/stop the worker pool

**Send flow in detail:**
```
POST /campaigns/:id/send
    │
    ├── Validate: campaign exists, belongs to user, status = draft
    ├── Validate: campaign has a list_id
    ├── Validate: campaign has subject and at least one body (html or plain)
    ├── Query: all subscribers in list where status = subscribed
    ├── If zero subscribers → return 422 (nothing to send)
    ├── Insert: one send_job row per subscriber in MySQL (status = pending)
    ├── Enqueue: one asynq job per subscriber in Redis
    ├── Update: campaign status → sending
    └── Return: 202 Accepted + { job_count: N }

Worker (per job):
    ├── Mark send_job status → processing in MySQL
    ├── Rewrite links in HTML body (for click tracking)
    ├── Inject tracking pixel into HTML body
    ├── Inject unsubscribe link into body
    ├── Send via EmailProvider (SMTP)
    ├── On success:
    │   ├── Mark send_job status → delivered
    │   └── If all jobs done → update campaign status → sent
    └── On failure:
        ├── asynq handles retry (up to 3×, exponential backoff)
        └── After 3 failures → asynq moves to failed queue (DLQ in Redis)
            └── Mark send_job status → failed in MySQL
```

**Why do link rewriting and pixel injection at send time (not enqueue time)?** Because the subscriber-specific tracking tokens must be generated per-subscriber. The worker already has subscriber context. Doing it at enqueue time would require passing large bodies through Redis, which is wasteful. Rewrite at delivery time.

**Acceptance Criteria:**
- `POST /campaigns/:id/send` returns 202 and a job count
- Campaign status changes to `sending`
- Workers deliver emails to MailHog in the test environment
- Failed jobs (simulated by pointing at a bad SMTP host) retry up to 3 times
- After 3 failures, job lands in DLQ and `send_job.status = failed` in MySQL
- When all jobs complete, campaign status changes to `sent`
- All tests pass

---

### Phase F — Scheduled Sending

**Goal:** A user can schedule a campaign to send at a future datetime. At the scheduled time, the system automatically triggers the send flow without user action.

**Why separate from Phase E:** Scheduled sending is a variant of the send flow, not a modification of it. Keeping it separate means Phase E is clean and testable before we add the scheduling layer on top.

**Tasks:**
1. Add `scheduled_at` handling to the campaign send endpoint (accept optional `scheduled_at` in request body)
2. If `scheduled_at` is present and in the future: set campaign status to `scheduled`, enqueue asynq job with the future time
3. The scheduled asynq job, when fired, runs the exact same send logic as Phase E
4. Validate: `scheduled_at` must be in the future (not past)
5. A scheduled campaign can be cancelled (status → `cancelled`) before the scheduled time
6. Write tests for scheduling, cancellation, and the "fires at right time" flow

**Acceptance Criteria:**
- Sending with a future `scheduled_at` sets status to `scheduled`
- Sending without `scheduled_at` works exactly as Phase E (immediate)
- Cancelling a scheduled campaign prevents delivery
- A scheduled campaign transitions to `sending` at the correct time
- All tests pass

---

### Phase G — Tracking

**Goal:** Opens, clicks, and unsubscribes are recorded. Unsubscribes flip the subscriber's status. These endpoints are public (no JWT) — they are called by email clients, not the user.

**New migrations:** `tracking_events` table

**New code:**
- `TrackingEvent` model
- Tracking token generation utilities (HMAC-signed, base64url-encoded)
- Tracking handler (public routes, no JWT middleware)

**Endpoints:**
```
GET /t/open/:token       — serve 1×1 transparent GIF, record open event
GET /t/click/:token      — record click event, redirect to original URL
GET /t/unsubscribe/:token — flip subscriber status, return confirmation page
```

**Why no JWT here?** These URLs are embedded in emails delivered to subscribers. Subscribers are not MailForge users. They do not have JWT tokens. Authentication is via the signed token embedded in the URL at send time.

**Open pixel behaviour:** Always return the GIF, even if the token is invalid or already seen. Never 404. A broken image in an email is a terrible user experience for the subscriber.

**Deduplication for opens:** Only record the first open per subscriber per campaign. Use a unique index on `(campaign_id, subscriber_id, event_type)` with `INSERT IGNORE` for opens.

**Acceptance Criteria:**
- `/t/open/:token` always returns a valid GIF
- Open events are recorded once per subscriber per campaign
- `/t/click/:token` records the click and redirects to the correct URL
- `/t/unsubscribe/:token` sets subscriber status to `unsubscribed` and returns a confirmation
- All events appear in `tracking_events` table
- All tests pass

---

### Phase H — Analytics

**Goal:** Return flat aggregate numbers per campaign. The user can see how their campaign performed.

**Endpoint:**
```
GET /campaigns/:campaignId/analytics
```

**Response:**
```json
{
  "campaign_id": "uuid",
  "sent": 1000,
  "delivered": 980,
  "opens": 430,
  "open_rate": 43.9,
  "clicks": 210,
  "click_rate": 21.4,
  "bounces": 20,
  "unsubscribes": 15
}
```

**Source of truth for each number:**
- `sent` → COUNT of send_jobs for campaign
- `delivered` → send_jobs WHERE status = delivered
- `bounces` → send_jobs WHERE status = failed
- `opens` → COUNT DISTINCT subscriber_id in tracking_events WHERE event_type = open
- `clicks` → COUNT DISTINCT subscriber_id WHERE event_type = click
- `unsubscribes` → COUNT WHERE event_type = unsubscribe
- Rates are calculated in the service layer, not the DB, rounded to one decimal place

**Why not a pre-aggregated table?** For MVP campaign volumes these queries are fast. Pre-aggregating now is premature optimisation. We note that if campaign volume grows significantly, a materialized summary table becomes necessary — but not yet.

**Acceptance Criteria:**
- Analytics endpoint returns correct numbers derived from real send_jobs and tracking_events data
- Rate calculations are correct
- User can only retrieve analytics for their own campaigns
- All tests pass

---

### Phase I — Admin Panel

**Goal:** Super admins can oversee the entire platform. Moderators can perform permitted actions. All admin actions are traceable.

**Tasks:**
1. Add `AdminMiddleware` (role check: moderator or super_admin)
2. Add `SuperAdminMiddleware` (role check: super_admin only)
3. Add permission checking middleware (checks `moderator_permissions` table for non-super-admins)
4. Build all admin endpoints as listed in Section 8.3
5. Seed script for creating the first super admin user
6. Write tests for all permission-gated scenarios

**Acceptance Criteria:**
- A super admin can view all users, all campaigns, queue stats
- A super admin can promote a user to moderator
- A super admin can grant/revoke individual permissions from a moderator
- A moderator with `view_users` permission can list users; without it, receives 403
- A moderator cannot grant permissions to anyone
- A moderator's actions are visible to super admin
- All tests pass

---

## 11. API Reference by Phase

### Public Routes (No Auth)
```
POST /auth/register
POST /auth/login
POST /auth/refresh
POST /auth/logout
GET  /t/open/:token
GET  /t/click/:token
GET  /t/unsubscribe/:token
GET  /health
```

### Protected Routes (JWT Required)
```
# Lists
POST   /lists
GET    /lists
GET    /lists/:listId
PUT    /lists/:listId
DELETE /lists/:listId
GET    /lists/:listId/subscribers
POST   /lists/:listId/subscribers/:subscriberId
DELETE /lists/:listId/subscribers/:subscriberId

# Subscribers
POST   /subscribers
GET    /subscribers
GET    /subscribers/:subscriberId
PUT    /subscribers/:subscriberId
DELETE /subscribers/:subscriberId
POST   /subscribers/bulk

# Campaigns
POST   /campaigns
GET    /campaigns
GET    /campaigns/:campaignId
PUT    /campaigns/:campaignId
DELETE /campaigns/:campaignId
POST   /campaigns/:campaignId/send
GET    /campaigns/:campaignId/analytics
```

### Admin Routes (JWT + Admin Role Required)
```
GET    /admin/users
GET    /admin/users/:userId
PATCH  /admin/users/:userId/suspend
PATCH  /admin/users/:userId/reactivate
POST   /admin/moderators
DELETE /admin/moderators/:userId
POST   /admin/moderators/:userId/permissions
DELETE /admin/moderators/:userId/permissions/:permission
GET    /admin/moderators/:userId/permissions
GET    /admin/campaigns
GET    /admin/campaigns/:campaignId
GET    /admin/queue/stats
GET    /admin/queue/failed
POST   /admin/queue/failed/:jobId/replay
GET    /admin/analytics/overview
```

---

## 12. Infrastructure & DevOps

### 12.1 Docker Compose (Local Dev)

```yaml
# docker-compose.yml
services:
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: secret
      MYSQL_DATABASE: mailforge
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql

  mysql_test:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: secret
      MYSQL_DATABASE: mailforge_test
    ports:
      - "3307:3306"

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

  mailhog:
    image: mailhog/mailhog
    ports:
      - "1025:1025"   # SMTP trap
      - "8025:8025"   # Web UI to view caught emails

volumes:
  mysql_data:
```

**Why MailHog?** During development and testing, we do not want to send real emails. MailHog is a fake SMTP server that catches all outgoing email and displays it in a browser UI at `localhost:8025`. Point `SMTP_HOST=localhost`, `SMTP_PORT=1025` in your `.env` and every email your workers send appears there instantly.

### 12.2 Package Structure

```
mailforge/
├── cmd/
│   ├── api/            — main entry point (main.go)
│   └── migration/      — migration CLI
├── internal/
│   ├── models/         — Bun model structs (user.go, list.go, etc.)
│   ├── modules/
│   │   ├── auth/       — handler, service, repository, dto
│   │   ├── list/       — handler, service, repository, dto
│   │   ├── subscriber/ — handler, service, repository, dto
│   │   ├── campaign/   — handler, service, repository, dto
│   │   ├── tracking/   — handler, service, repository
│   │   ├── analytics/  — handler, service, repository
│   │   └── admin/      — handler, service, repository, dto
│   ├── middleware/     — JWT, admin, permission middleware
│   ├── providers/      — EmailProvider interface + implementations
│   ├── workers/        — asynq worker setup and job handlers
│   ├── routes/         — route registration only
│   └── di/             — Uber Fx wiring
├── pkg/
│   └── logger/         — Zap logger utility
├── migrations/         — SQL migration files
├── keys/               — RSA key pair (private.pem gitignored, public.pem committed)
├── docker-compose.yml
├── Makefile
├── .env.example
└── README.md
```

---

## 13. CI/CD Pipeline

### 13.1 GitHub Actions

**Cost:** GitHub Actions is free for public repositories (unlimited minutes). For private repositories, GitHub provides 2,000 minutes/month on the free plan — more than sufficient for a small team at this stage.

**File:** `.github/workflows/ci.yml`

**Pipeline runs on:** every `push` to any branch, every `pull_request` targeting `main`.

**Pipeline steps:**
1. Checkout code
2. Set up Go (version from `go.mod`)
3. Start MySQL and Redis via Docker Compose (service containers)
4. Run database migrations against the test DB
5. Run `go vet ./...`
6. Run `golangci-lint` (static analysis)
7. Run `go test ./... -race -cover` (all tests with race detector)

**Why the race detector?** Go's `-race` flag instruments the binary to detect data races at runtime. Since MailForge uses goroutines (worker pool, asynq, concurrent request handling), a data race is a real risk. The race detector catches them in CI before they reach production.

**Why lint in CI?** Code review catches logic errors. Lint catches style drift, unused variables, shadowed errors, and other subtle bugs that reviewers miss. Automating it removes the burden from reviewers.

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: ["**"]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      mysql:
        image: mysql:8.0
        env:
          MYSQL_ROOT_PASSWORD: secret
          MYSQL_DATABASE: mailforge_test
        ports:
          - 3306:3306
        options: --health-cmd="mysqladmin ping" --health-interval=10s --health-timeout=5s --health-retries=3

      redis:
        image: redis:7-alpine
        ports:
          - 6379:6379

    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version-file: go.mod

      - name: Install dependencies
        run: go mod download

      - name: Run migrations
        run: make migrate-up
        env:
          DB_HOST: 127.0.0.1
          DB_PORT: 3306
          DB_USER: root
          DB_PASSWORD: secret
          DB_NAME: mailforge_test

      - name: Lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: latest

      - name: Test
        run: go test ./... -race -cover
        env:
          DB_HOST: 127.0.0.1
          DB_PORT: 3306
          DB_USER: root
          DB_PASSWORD: secret
          DB_NAME: mailforge_test
          REDIS_URL: redis://localhost:6379
```

---

## 14. Testing Strategy

### 14.1 Philosophy

**Test against reality, not assumptions.** Mocking the database means you are testing your mock, not your code. Bugs in SQL queries, ORM tag mistakes, transaction logic, and constraint violations only surface against a real database. We use a real MySQL test database and a real Redis instance in all tests.

**Tests are not optional.** Phase A and the domain model work (no business logic) are the only areas that may ship without tests. From Phase B onward, every service method has a unit test and every handler has an integration test. A PR that adds a feature without its tests will not be merged.

### 14.2 Test Types

| Type | What it tests | Tools |
|---|---|---|
| Unit | Service layer business logic (with real DB if needed) | `testing`, `testify` |
| Integration | Full HTTP handler → service → repository → DB round-trip | `net/http/httptest`, `testify` |
| Worker | Job enqueue → worker processes → DB updated | asynq test helpers, real Redis |

### 14.3 Test Database

A separate `mailforge_test` database (Dockerised on port 3307 locally, 3306 in CI) is used for all tests. The test suite runs migrations on boot and truncates relevant tables between test cases.

### 14.4 Coverage Targets

| Phase | Minimum Coverage |
|---|---|
| Phase B (Auth) | 90% of service layer |
| Phase C onwards | 80% of service layer, 100% of happy + error paths for handlers |

---

## 15. Cross-Cutting Standards

These standards apply to every line of code in every phase. Establish them in Phase B and never deviate.

### 15.1 Error Response Format

All errors return the same JSON shape, always:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "email is required",
    "details": {}
  }
}
```

HTTP status mapping:
- `400` — validation error (bad input)
- `401` — unauthenticated (missing or invalid token)
- `403` — unauthorised (valid token, wrong ownership or role)
- `404` — resource not found
  - `409` — conflict (e.g. duplicate email)
- `422` — unprocessable (e.g. trying to edit a sent campaign)
- `500` — internal error (never expose stack traces or internal error messages)

### 15.2 DTO Convention

Request and response structs are defined separately from domain models.

- Location: `internal/modules/<module>/<module>.dto.go`
- Domain models (Bun structs) **never** appear in HTTP responses
- Response DTOs always use `public_id`, never the internal auto-increment `id`
- Every response DTO has explicit JSON tags — never rely on Go's default field name lowercasing

### 15.3 Ownership Scoping

Every repository method that reads user-owned data must accept `userID` as an explicit parameter. This is not the caller's responsibility to remember — it is enforced at the method signature:

```go
// Correct
func (r *ListRepository) FindAll(ctx context.Context, userID uint64) ([]models.List, error)

// Wrong — userID inferred somewhere inside, invisible to the caller
func (r *ListRepository) FindAll(ctx context.Context) ([]models.List, error)
```

### 15.4 Never Expose Internal IDs

The `id` column (auto-increment integer) is an internal database detail. It must never appear in any API response, URL parameter, or token payload. Use `public_id` (UUID) everywhere externally.

### 15.5 Package Naming

Packages are named after their directory. `internal/modules/auth/` is `package auth`. No `package user`, `package main` in non-main packages, or generic names like `package utils`.

### 15.6 Context Propagation

All repository and service methods accept `context.Context` as their first parameter. This enables request cancellation, deadline propagation, and clean shutdown behaviour. Never call a database or Redis operation without a context.

---

## 16. Environment Variables Reference

```env
# Application
APP_ENV=development
APP_PORT=3010
APP_NAME=MailForge

# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=secret
DB_NAME=mailforge

# Redis
REDIS_URL=redis://localhost:6379

# JWT (RS256)
JWT_PRIVATE_KEY_PATH=./keys/private.pem
JWT_PUBLIC_KEY_PATH=./keys/public.pem
JWT_ACCESS_EXPIRY=1h
JWT_REFRESH_EXPIRY=168h

# SMTP (default provider)
SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_USER=
SMTP_PASSWORD=
SMTP_FROM=noreply@mailforge.com
SMTP_TLS=false

# Email provider selection
EMAIL_PROVIDER=smtp

# Workers
WORKER_POOL_SIZE=5

# App base URL (used for tracking links)
APP_BASE_URL=http://localhost:3010
```

---

## 17. Decision Log

| Phase | Decision | Reason |
|---|---|---|
| Foundation | Redis (asynq) for job queue instead of MySQL SKIP LOCKED | asynq provides native scheduling, retry policies, DLQ, and a monitoring UI out of the box. MySQL queue would require re-implementing all of these manually. Operational simplicity favoured. |
| Foundation | MySQL remains primary store for all persistent data including send_jobs audit records | Redis is ephemeral. Business data (campaign delivery history, analytics) must survive a Redis restart. MySQL is the source of truth; Redis is the processing layer. |
| Foundation | EmailProvider interface from day one, SMTP first | We want to add Resend, SES, or Mailgun later without modifying worker logic. An interface costs almost nothing to define now and saves significant refactoring later. |
| Foundation | Scheduling in scope | asynq natively supports scheduled/delayed jobs. The feature has real user value (Grace's use case). Cost is minimal given the tooling. |
| Foundation | Admin as two-tier (super_admin + moderator with permissions) | A single all-powerful admin account is a security risk and doesn't scale to a small ops team. Least-privilege access per operator is the correct pattern. |
| Auth | RS256 JWT instead of HS256 | Asymmetric signing: private key signs (never leaves server), public key verifies (safe to distribute). More secure than a shared secret and positions the system for multi-service verification in the future. |
| Auth | Access token 1 hour, refresh token 7 days with rotation | Short access token limits breach window. Refresh token rotation means each token can only be used once — a stolen and used refresh token is detectable because the legitimate holder's next attempt will fail. |
| Auth | Refresh tokens stored in Redis | Session data with TTL. Redis handles expiry automatically, is fast, and doesn't need relational queries. Avoids polluting MySQL with high-write session rows. |
| Subscribers | Soft delete only | Hard-deleting a subscriber corrupts open/click/bounce counts for historical campaigns. Preserve the row, flip the status. |
| All | Public UUIDs, never internal IDs | Auto-increment integers are predictable (enumerable). UUIDs are safe in URLs, tokens, and external references. Internal IDs are a database implementation detail. |
| Infrastructure | Docker Compose for local dev | Every developer gets an identical MySQL, Redis, and MailHog environment. Zero manual setup, no "works on my machine." CI uses the same images. |
| CI | GitHub Actions | Free tier (2,000 min/month for private repos, unlimited for public) is sufficient for the team at this stage. No cost. Tight GitHub integration with PRs and branch protection. |
| Testing | No database mocking | Mocking the DB layer tests the mock, not the code. SQL bugs, ORM tag errors, and constraint violations only surface against a real database. We test against reality. |