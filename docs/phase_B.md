# MailForge — Phase B PDR: Authentication & Identity

> **Parent Document:** MailForge Product Definition & Requirements (PDR)
> **Phase:** B — Authentication & Identity
> **Status:** Ready to build
> **Rule:** This document breaks Phase B (PDR §7, §10) into buildable, PR-sized deliverables. Nothing here changes the parent PDR — it operationalizes it. If a conflict appears, the parent PDR wins.

---

## 1. Phase Goal (restated from parent PDR)

A user can register, log in, receive RS256 access and refresh tokens, and use those tokens to access protected routes. Token refresh and logout work correctly.

**Why this phase is gated:** every route built from Phase C onward sits behind auth. If auth is half-built, every later phase either builds on sand or has to stop and retrofit middleware and ownership checks. Auth ships complete, or it doesn't ship.

**No new migrations.** The `users` table already exists from Phase A models.

---

## 2. Deliverable Count & Sprint Structure

**Phase B = 8 deliverables, each its own PR, each its own small sprint.** They are strictly sequential — no deliverable starts until the one before it is merged into the target branch (`main`) with CI green. This is intentional: auth is exactly the kind of subsystem where reviewing 8 small, focused diffs catches more than reviewing one 2,000-line diff.

```
D1: Config & Key Loading
        │
        ▼
D2: Auth DTOs & Validation ──────┐
        │                        │
        ▼                        │
D3: Auth Repository              │
        │                        │
        ▼                        │
D4: JWT + Refresh Token Utility   │
        │                        │
        ▼                        │
D5: Auth Service (business logic)◄┘
        │
        ▼
D6: JWT Middleware
        │
        ▼
D7: Auth Handler + Routes
        │
        ▼
D8: Integration Test Suite + Docs Close-Out
```

D2 can technically be built in parallel with D1/D3 since it has no dependencies on either, but PRs still merge in the numbered order below to keep history linear and easy to bisect if something breaks.

| # | Deliverable | Branch name | Depends on |
|---|---|---|---|
| D1 | Config & RSA key loading | `Phase_B/Task_1-jwt-key-config` | Phase A |
| D2 | Auth DTOs & validation | `Phase_B/Task_2-auth-dtos` | Phase A |
| D3 | Auth repository | `Phase_B/Task_3-auth-repository` | D1 |
| D4 | JWT + refresh token utility | `Phase_B/Task_4-token-utility` | D1 |
| D5 | Auth service | `Phase_B/Task_5-auth-service` | D2, D3, D4 |
| D6 | JWT middleware | `Phase_B/Task_6-jwt-middleware` | D4 |
| D7 | Auth handler + routes | `Phase_B/Task_7-auth-routes` | D5, D6 |
| D8 | Integration suite + docs | `Phase_B/Task_8-integration-closeout` | D7 |

**Branch strategy:** unchanged from Phase A — `feature branch → PR → merge to main`. There is no `phase1` integration branch; every deliverable branches off `main` and merges directly back into `main` once CI is green.

---

## 3. Detailed Deliverable Specs

### D1 — Config & RSA Key Loading

**Goal:** The application can load the RSA key pair and JWT expiry settings from config at boot, via Fx, with no business logic attached yet.

**Files:**
- `internal/config/config.go` — **replace** the existing `JwtConfig` struct entirely, do not append to it
- `internal/config/jwt_keys.go` — loads and parses PEM files into `*rsa.PrivateKey` / `*rsa.PublicKey`
- `internal/di/` — Fx provider wiring the parsed keys as constructor-injectable values

**`JwtConfig` — before (HS256 remnant, to be deleted in full):**
```go
type JwtConfig struct {
    JwtSecret string   // HS256 remnant — must be deleted
    JwtExpiry string   // too vague — must be split into two
}
```

**`JwtConfig` — after (this PR's actual output):**
```go
type JwtConfig struct {
    PrivateKeyPath  string
    PublicKeyPath   string
    AccessExpiry    string  // maps to JWT_ACCESS_EXPIRY
    RefreshExpiry   string  // maps to JWT_REFRESH_EXPIRY
}
```

**Implementation notes:**
- This is a **replacement, not an addition**. `JwtSecret` and `JwtExpiry` are deleted along with the env vars they read (`JWT_SECRET`, `JWT_EXPIRY`) — both are HS256-era and have no place in an RS256 system.
- `.env` and `.env.example` already carry the correct RS256 variables from Phase A (`JWT_PRIVATE_KEY_PATH`, `JWT_PUBLIC_KEY_PATH`, `JWT_ACCESS_EXPIRY`, `JWT_REFRESH_EXPIRY`) — this PR wires them up, it doesn't introduce them.
- Fail fast at boot if keys are missing or malformed — don't let the app start in a state where auth will panic on first request.
- Reuse `make gen-keys` (already exists from Phase A) — no new keygen tooling needed.

**Tests:** Unit tests only — valid key pair loads correctly; missing file errors; malformed PEM errors.

**Acceptance criteria:**
- App boots successfully with valid keys present
- App fails to boot with a clear error if keys are missing/invalid
- No behavior beyond loading — this PR touches no HTTP routes

**PR title:** `feat(config): load RSA key pair and JWT expiry settings via Fx`

---

### D2 — Auth DTOs & Validation

**Goal:** Request/response shapes for auth exist and validate input, with zero DB or token logic attached.

**Files:**
- `internal/modules/auth/auth.dto.go` — `RegisterRequest`, `LoginRequest`, `RefreshRequest`, `LogoutRequest`, `AuthResponse`

**Implementation notes:**
- Per PDR §15.2: DTOs are separate from domain models, explicit JSON tags always, `public_id` only — never internal `id`.
- The `users` table carries a `username` column (product decision, already in schema) — `RegisterRequest` must include it explicitly. Fields are:

```
RegisterRequest:
  - username  string  (required, min 3 chars, alphanumeric + underscores)
  - email     string  (required, valid email format)
  - password  string  (required, min 8 chars)
```

- `AuthResponse` fields, fully specified:

```
AuthResponse:
  access_token  string   — signed RS256 JWT
  refresh_token string   — opaque UUID
  expires_in    int      — seconds until access token expires (e.g. 3600)
```

  `expires_in` is an integer count of seconds, not a timestamp or duration string — this lets clients schedule a proactive refresh without parsing the JWT payload themselves.

**Tests:** Unit tests on validation logic only (valid input passes, each invalid case fails with the right message).

**Acceptance criteria:**
- All 4 request DTOs validate correctly on valid and invalid input
- No route, service, or repository code in this PR

**PR title:** `feat(auth): add auth request/response DTOs with validation`

---

### D3 — Auth Repository

**Goal:** Data access methods for the `users` table exist and are tested against the real test DB.

**Files:**
- `internal/modules/auth/auth.repository.go` — `FindByEmail`, `CreateUser`, `FindByID`, `UpdateLastLogin`, `IncrementFailedAttempts`

**Implementation notes:**
- Per PDR §15.3 and §15.6: every method takes `context.Context` first, and any user-scoped method takes explicit parameters — no inference.
- Uses the `User` Bun model from Phase A — no schema changes.
- `CreateUser` assumes password is already hashed by the caller (service layer owns bcrypt, not the repository).
- The `users` table has two columns that only get written at login time — `last_login_at` and `failed_login_attempts`. Leaving them unwritten makes them dead weight in the schema, so this deliverable adds the two methods that write them:

```
UpdateLastLogin(ctx, userID)          — sets last_login_at = NOW(), resets failed_login_attempts = 0
IncrementFailedAttempts(ctx, userID)  — increments failed_login_attempts by 1
```

**Tests:** Integration tests against the real `mailforge_test` DB (per PDR §14.1 — no mocking). Cover: create succeeds, duplicate email fails at DB constraint level, `FindByEmail`/`FindByID` return correct rows and a clean "not found" case, `UpdateLastLogin` correctly sets the timestamp and resets the counter, `IncrementFailedAttempts` correctly increments from any starting value.

**Acceptance criteria:**
- All five methods work against real MySQL
- Duplicate email attempt surfaces a distinguishable error the service layer can map to `409 Conflict`
- `UpdateLastLogin` and `IncrementFailedAttempts` are exercised by dedicated tests, not just incidentally covered by D5's later service tests

**PR title:** `feat(auth): add auth repository (FindByEmail, CreateUser, FindByID, UpdateLastLogin, IncrementFailedAttempts)`

---

### D4 — JWT + Refresh Token Utility

**Goal:** A standalone token utility can sign/verify RS256 access tokens and issue/rotate opaque refresh tokens in Redis, independent of the auth service.

**Files:**
- `pkg/tokens/jwt.go` — `GenerateAccessToken(userID, role)`, `VerifyAccessToken(token)`
- `pkg/tokens/refresh.go` — `IssueRefreshToken(ctx, userID)`, `ValidateAndRotate(ctx, token)`, `RevokeRefreshToken(ctx, token)`

**Implementation notes:**
- Access token claims exactly per PDR §7.2: `sub`, `role`, `iat`, `exp`.
- Refresh tokens are opaque UUIDs stored in Redis as `refresh:<token_uuid>` with a 7-day TTL (PDR §7.3) — never JWTs themselves.
- Rotation must use **`GETDEL`** specifically — not a `MULTI`/`EXEC` transaction, and not a read-then-delete as two separate round trips. A transaction without `WATCH` still leaves a race window between the read and the delete; `GETDEL` is a single atomic command (available since Redis 6.2, confirmed present on our `redis:7-alpine` image), so it's the only correct choice here, not one of several acceptable options.

  **Rotation flow, exactly:**
  ```
  1. GETDEL refresh:<token>        — atomically read and delete in one command
  2. If nil → token doesn't exist or was already used → reject
  3. If found → extract userID from the stored Redis value → use it to issue new token pair
  4. SET refresh:<new_token> <userID> EX 604800
  ```

**Tests:** Unit tests with a real Redis instance (per PDR §14.1). Cover: valid token round-trip, expired token rejected, tampered JWT rejected, refresh token single-use enforced (second use after rotation fails via `GETDEL` returning nil).

**Acceptance criteria:**
- Access tokens sign and verify correctly with the RSA key pair from D1
- Refresh tokens are single-use — reuse after rotation fails
- Expired tokens (both types) are rejected

**PR title:** `feat(tokens): add RS256 JWT signing and Redis-backed refresh token rotation`

---

### D5 — Auth Service (business logic)

**Goal:** `Register`, `Login`, `Refresh`, `Logout` exist as service methods, composing D2 (validation), D3 (repository), and D4 (tokens) — the first PR where real auth behavior comes together.

**Files:**
- `internal/modules/auth/auth.service.go`

**Implementation notes:**
- `Register`: validate → bcrypt hash (cost 12, per PDR §4) → `CreateUser` → issue token pair. `UpdateLastLogin` is **not** called here — the user has registered, not logged in.
- `Login`, exactly:

```
Login:
  1. FindByEmail → if not found → generic 401 (never reveal whether email or password was wrong)
  2. bcrypt.Compare → if mismatch → IncrementFailedAttempts → generic 401
  3. On success → UpdateLastLogin → issue token pair → return AuthResponse
```

- `Refresh`: validate + rotate via D4's `GETDEL` flow, issue new access token
- `Logout`: revoke refresh token via D4
- Duplicate email on register maps to `409 Conflict` per PDR §15.1 error format

**Tests:** Unit tests targeting **90% coverage of the service layer**, per PDR §14.4's Phase B target specifically. Cover every happy and error path: successful register, duplicate email, successful login, wrong password, wrong email, successful refresh, expired/reused refresh token, logout.

**Acceptance criteria:**
- All four service methods behave exactly as specified in PDR §7
- 90%+ test coverage on this file, verified locally with `go test ./internal/modules/auth/... -cover`

**PR title:** `feat(auth): implement auth service (register, login, refresh, logout)`

---

### D6 — JWT Middleware

**Goal:** A Chi middleware validates the access token on protected routes and injects `userID` + `role` into request context.

**Files:**
- `internal/middleware/jwt.go` — `JWTMiddleware(verifier)`

**Implementation notes:**
- Missing or malformed `Authorization` header → `401`
- Expired or invalid signature → `401`
- On success, inject `userID` and `role` into `r.Context()` using unexported context key types (avoid collisions, standard Go practice)
- This PR does **not** wire the middleware into the router yet — that's D7. Keep this PR focused purely on the middleware unit in isolation.

**Tests:** Unit tests with `httptest` — valid token passes through to next handler with context populated; missing/expired/malformed tokens all return 401 with the correct error body shape from PDR §15.1.

**Acceptance criteria:**
- Valid token → request proceeds, context has `userID` and `role`
- Any invalid case → `401` with standard error JSON shape

**PR title:** `feat(middleware): add JWT access token validation middleware`

---

### D7 — Auth Handler + Routes

**Goal:** The four public auth endpoints exist and are wired into the router. This is the first PR where the feature is actually callable over HTTP end-to-end.

**Files:**
- `internal/modules/auth/auth.handler.go`
- `internal/routes/` — register `/auth/register`, `/auth/login`, `/auth/refresh`, `/auth/logout` as public; apply `JWTMiddleware` (D6) as router-group middleware for all future protected routes (even though no protected business routes exist yet — the grouping goes in now so Phase C doesn't have to touch router wiring again)

**Implementation notes:**
- Handlers stay thin per architecture overview — validate via DTO, call service, map result/error to response, nothing else.
- Error mapping follows PDR §15.1 exactly (401 for bad credentials, 409 for duplicate email, etc.)

**Tests:** Integration tests via `net/http/httptest`, full round-trip through handler → service → repository → real DB/Redis (per PDR §14.1 — no mocking). Cover every endpoint's happy path and every documented error case from the parent PDR's Phase B acceptance criteria.

**Acceptance criteria (mirrors parent PDR §10 Phase B exactly):**
- `POST /auth/register` creates a user, returns tokens
- `POST /auth/login` — correct creds → tokens; wrong creds → 401
- `POST /auth/refresh` — valid refresh token → new pair + old one invalidated
- `POST /auth/logout` — invalidates refresh token
- Protected route without token → 401
- Protected route with expired token → 401

**PR title:** `feat(auth): wire auth handler and routes (register, login, refresh, logout)`

---

### D8 — Integration Test Suite + Docs Close-Out

**Goal:** Confirm the whole phase holds together as a system, not just as individually-tested parts, and leave the repo in a state where Phase C can start immediately.

**Files:**
- `internal/modules/auth/auth_integration_test.go` (if not already sufficiently covered by D7 — this PR fills any remaining gaps, e.g. sequential flows like register → login → refresh → logout as one continuous test)
- `README.md` — document the four auth endpoints, how to run `make gen-keys`, and how auth fits into the request lifecycle
- `.env.example` — confirm all JWT vars from PDR §16 are present and documented

**Implementation notes:**
- This is a checkpoint PR, not new feature work. If D1–D7 were built correctly, this PR should mostly be additive tests plus documentation, not bug fixes. If real bugs surface here, that's useful signal that an earlier PR's tests had a gap.
- Run the full suite with `-race` locally before pushing, matching CI's exact command from PDR §13.1.

**Tests:** Full `go test ./... -race -cover` run, confirming no regressions across the whole repo, not just the auth module.

**Acceptance criteria:**
- All Phase B acceptance criteria from parent PDR §10 pass together, in sequence, in one test run
- `go test ./... -race -cover` is green
- CI is green on the PR
- README accurately describes how to authenticate against the running API

**PR title:** `test(auth): full auth flow integration coverage + Phase B docs`

---

## 4. PR & Merge Discipline

- **One deliverable = one PR = one merge.** No bundling two deliverables into one PR, even if they feel small — the point is reviewability and the ability to `git bisect` cleanly if something in auth breaks later.
- **Don't start Dn+1 until Dn is merged and CI is green** on the target branch (`main`). Building D5 against an unmerged D3 branch risks rebasing pain and review confusion about what's actually being reviewed.
- **Every PR description should state:** which deliverable it is (D1–D8), what it depends on, and which specific parent-PDR acceptance criteria (§7, §10) it satisfies — reviewers should be able to check the PR against the PDR line by line.
- **Commit hygiene:** within a single deliverable's branch, still split commits logically (e.g. D5 might be "add Register method", "add Login method", "add Refresh + Logout methods", "add service unit tests") — same pattern you used on the Phase A legacy fixes.

---

## 5. Testing & Coverage Recap (per parent PDR §14)

- **No mocking the DB or Redis** — every test in D3, D4, D5, D7, D8 runs against the real Dockerised `mailforge_test` MySQL and real Redis.
- **Coverage target for Phase B specifically: 90% of the service layer** (D5) — this is a higher bar than the 80% default from Phase C onward, because auth is the highest-blast-radius module in the system.
- **No PR ships without tests** from D2 onward — only D1 (pure config plumbing) is borderline-exemptable, but even it should carry basic unit tests per this doc.

---

## 6. Phase B Definition of Done

Phase B is complete when, in one continuous run:

- [ ] `POST /auth/register` creates a user and returns `access_token` + `refresh_token`
- [ ] `POST /auth/login` returns tokens on correct credentials, `401` on incorrect
- [ ] `POST /auth/refresh` rotates the refresh token and issues a new pair
- [ ] `POST /auth/logout` revokes the refresh token
- [ ] A protected route without a token returns `401`
- [ ] A protected route with an expired token returns `401`
- [ ] Service layer test coverage ≥ 90%
- [ ] `go test ./... -race -cover` passes locally and in CI
- [ ] All 8 deliverable PRs merged into `main` in order
- [ ] README and `.env.example` reflect the finished auth system

Once every box is checked, Phase C (Lists & Subscribers) can begin — it depends on JWT middleware and authenticated `userID` context being reliable, which is exactly what this phase delivers.