# MailForge — Phase C PDR: Lists & Subscribers

> **Parent Document:** MailForge Product Definition & Requirements (PDR)
> **Phase:** C — Lists & Subscribers
> **Status:** Ready to build
> **Rule:** This document breaks Phase C (PDR §10 Phase C) into buildable, PR-sized deliverables. Nothing here changes the parent PDR — it operationalizes it. If a conflict appears, the parent PDR wins.
> **Depends on:** Phase B (Auth & Identity) — complete, merged to `main`, CI green
> **Blocks:** Phase D (Campaigns) — campaigns target lists, so this phase must ship first
> **New migrations:** None — `lists`, `subscribers`, `list_subscribers` already exist as models from Phase A

---

## 1. Phase Goal (restated from parent PDR)

Authenticated users can create and manage email lists, add subscribers (manually or via CSV), and manage which subscribers belong to which lists. Every read and write is scoped to the authenticated user — a user must never be able to see, edit, or reference another user's list or subscriber, even indirectly.

**Why this phase is gated:** campaigns target lists. A campaign built against no real list/subscriber data is untestable in any meaningful way. This phase exists so that by the time Phase D starts, the campaign layer has real lists and real subscribers to attach to, instead of being built and tested against fixtures that don't reflect the actual data shape.

There's a second reason this phase matters more than its size suggests: it's where the **ownership-scoping pattern gets established for every module after it**. Auth (Phase B) didn't need this pattern — a user only ever acts on their own account. Lists and subscribers are the first resources where "does this row belong to the person asking for it?" becomes a real question, asked on every single endpoint. Get the pattern right here — `userID` explicit in every repository method signature, never inferred — and campaigns, send jobs, and analytics all just copy it. Get it wrong here and it has to be retrofitted everywhere downstream, which is a much more expensive fix.

**No new migrations.** `lists`, `subscribers`, `list_subscribers` already exist from Phase A models.

---

## 2. Deliverable Count & Sprint Structure

**Phase C = 10 deliverables, each its own PR, each its own small sprint.** List and Subscriber are independent resources up through their individual CRUD, so their DTO/repository/service pairs can be built in parallel branches — but they still merge in the numbered order below to keep history linear and bisectable. CSV import and list membership both depend on subscriber and/or list plumbing being merged first, since they compose those repositories rather than duplicating query logic.

```
D1: List DTOs & Validation ─────────┐
        │                           │
        ▼                           │
D2: List Repository                 │
        │                           │
        ▼                           │
D3: List Service + Handler ◄────────┘
        │
        │        D4: Subscriber DTOs & Validation ─────┐
        │                │                              │
        │                ▼                              │
        │        D5: Subscriber Repository               │
        │                │                              │
        │                ▼                              │
        │        D6: Subscriber Service + Handler ◄──────┘
        │                │
        │                ▼
        │        D7: CSV Bulk Import
        │                │
        ▼                ▼
D8: List Membership Endpoints
        │
        ▼
D9: Route Registration & Wiring
        │
        ▼
D10: Integration Test Suite + Docs Close-Out
```

| # | Deliverable | Branch name | Depends on |
|---|---|---|---|
| D1 | List DTOs & validation | `Phase_C/Task_1-list-dtos` | Phase B |
| D2 | List repository | `Phase_C/Task_2-list-repository` | Phase A |
| D3 | List service + handler | `Phase_C/Task_3-list-service-handler` | D1, D2 |
| D4 | Subscriber DTOs & validation | `Phase_C/Task_4-subscriber-dtos` | Phase B |
| D5 | Subscriber repository | `Phase_C/Task_5-subscriber-repository` | Phase A |
| D6 | Subscriber service + handler | `Phase_C/Task_6-subscriber-service-handler` | D4, D5 |
| D7 | CSV bulk import | `Phase_C/Task_7-csv-bulk-import` | D6 |
| D8 | List membership endpoints | `Phase_C/Task_8-list-membership` | D2, D5 |
| D9 | Route registration & wiring | `Phase_C/Task_9-routes-wiring` | D3, D6, D7, D8 |
| D10 | Integration suite + docs close-out | `Phase_C/Task_10-integration-closeout` | D9 |

**Branch strategy:** unchanged from Phase B — `feature branch → PR → merge to main`. No `phaseC` integration branch; every deliverable branches off `main` and merges directly back once CI is green.

---

## 3. Detailed Deliverable Specs

### D1 — List DTOs & Validation

**Goal:** Request/response shapes for lists exist and validate input, with zero DB logic attached.

**Files:**
- `internal/modules/list/list.dto.go` — `CreateListRequest`, `UpdateListRequest`, `ListResponse`

**Implementation notes:**
- Per PDR §15.2: DTOs are separate from domain models, explicit JSON tags always, `public_id` only — never internal `id`.
- `CreateListRequest`: `name` (required, min 1 char), `description` (optional).
- `UpdateListRequest`: same fields, all optional — a partial update shouldn't force the caller to resend everything.
- `ListResponse` includes `status` (`active`/`archived`) so a client can distinguish an archived list from a deleted one without a second call.

**Tests:** Unit tests on validation logic only (valid input passes, each invalid case fails with the right message).

**Acceptance criteria:**
- All list DTOs validate correctly on valid and invalid input
- No route, service, or repository code in this PR

**PR title:** `feat(list): add list request/response DTOs with validation`

---

### D2 — List Repository

**Goal:** Data access methods for the `lists` table exist and are tested against the real test DB.

**Files:**
- `internal/modules/list/list.repository.go` — `Create`, `FindAll(ctx, userID)`, `FindByID(ctx, userID, publicID)`, `Update`, `Archive`

**Implementation notes:**
- Per PDR §15.3 and §15.6: every method takes `context.Context` first, and every user-scoped method takes `userID` as an explicit parameter — never inferred inside the method. This is what turns an ownership bug into something visible in a method signature during review, instead of something buried in a `WHERE` clause that has to be hunted for.
- Uses the `List` Bun model from Phase A — no schema changes.
- `Archive` is an `UPDATE ... SET status = 'archived'`, never a `DELETE` — the row must survive so campaigns that once targeted it still resolve historically.

**Tests:** Integration tests against the real `mailforge_test` DB (per PDR §14.1 — no mocking). Cover: create succeeds, `FindAll` returns only the calling user's lists, `FindByID` returns a clean "not found" case for another user's list, `Archive` correctly flips status without deleting the row.

**Acceptance criteria:**
- All five methods work against real MySQL
- `FindByID` and `FindAll` never return another user's row, proven by a dedicated test, not just incidentally covered later

**PR title:** `feat(list): add list repository (Create, FindAll, FindByID, Update, Archive)`

---

### D3 — List Service + Handler

**Goal:** Full list CRUD is callable over HTTP, composing D1 (DTOs) and D2 (repository).

**Files:**
- `internal/modules/list/list.service.go`
- `internal/modules/list/list.handler.go`

**Implementation notes:**
- Service enforces: archiving a list is a status flip, never a row delete — the repository already guarantees this, but the service is where the state transition is decided, so this is the file where that intent lives.
- Handler returns **403**, not 404, when a list exists but belongs to a different user. 404 would tell an attacker whether the resource exists at all — 403 doesn't leak that.
- Handlers stay thin per architecture overview: validate via DTO, call service, map result/error to response, nothing else.

**Tests:** One handler smoke test per route to confirm routing and DTO serialization wire correctly. Full behavioral coverage lives in the service unit tests, not duplicated here — see §5.

**Acceptance criteria:**
- `POST/GET/PUT/DELETE /lists` and `GET /lists/:listId` all work end-to-end
- Cross-user access to a list returns `403`

**PR title:** `feat(list): implement list service and handler (full CRUD)`

---

### D4 — Subscriber DTOs & Validation

**Goal:** Request/response shapes for subscribers exist and validate input, with zero DB logic attached.

**Files:**
- `internal/modules/subscriber/subscriber.dto.go` — `CreateSubscriberRequest`, `UpdateSubscriberRequest`, `SubscriberResponse`

**Implementation notes:**
- `CreateSubscriberRequest`: `name` (required), `email` (required, valid email format).
- `SubscriberResponse` includes `status` (`subscribed`/`unsubscribed`) — same reasoning as the list DTO, the client needs to see state without a second call.

**Tests:** Unit tests on validation logic only.

**Acceptance criteria:**
- All subscriber DTOs validate correctly on valid and invalid input
- No route, service, or repository code in this PR

**PR title:** `feat(subscriber): add subscriber request/response DTOs with validation`

---

### D5 — Subscriber Repository

**Goal:** Data access methods for the `subscribers` table exist and are tested against the real test DB.

**Files:**
- `internal/modules/subscriber/subscriber.repository.go` — `Create`, `FindAll(ctx, userID)`, `FindByID(ctx, userID, publicID)`, `Update`, `Unsubscribe`

**Implementation notes:**
- Same explicit-`userID` repository pattern as D2.
- Schema constraint to respect: `UNIQUE KEY uq_user_email (user_id, email)`. A duplicate-email insert should surface a distinguishable error the service layer can map to `409 Conflict` — not a raw MySQL constraint error leaking through.
- `Unsubscribe` is a status flip to `unsubscribed`, never a `DELETE` — a hard-deleted subscriber corrupts historical open/click/bounce counts for any campaign they were once part of. The row has to survive forever; only the status changes.

**Tests:** Integration tests against real MySQL. Cover: create succeeds, duplicate email within the same user fails at the DB constraint level, `FindAll`/`FindByID` scoped correctly to `userID`, `Unsubscribe` flips status without deleting the row.

**Acceptance criteria:**
- All five methods work against real MySQL
- Duplicate email attempt surfaces a distinguishable error, exercised by a dedicated test

**PR title:** `feat(subscriber): add subscriber repository (Create, FindAll, FindByID, Update, Unsubscribe)`

---

### D6 — Subscriber Service + Handler

**Goal:** Full single-subscriber CRUD is callable over HTTP, composing D4 (DTOs) and D5 (repository).

**Files:**
- `internal/modules/subscriber/subscriber.service.go`
- `internal/modules/subscriber/subscriber.handler.go`

**Implementation notes:**
- Duplicate email on create maps to `409 Conflict` per PDR §15.1 error format.
- `DELETE /subscribers/:subscriberId` calls `Unsubscribe`, not a hard delete — same soft-delete reasoning as D5, decided at the service layer.
- Cross-user access returns `403`, same as list handler.

**Tests:** One handler smoke test per route. Full behavioral coverage lives in service unit tests — see §5.

**Acceptance criteria:**
- `POST/GET/PUT/DELETE /subscribers` and `GET /subscribers/:subscriberId` all work end-to-end
- Cross-user access to a subscriber returns `403`

**PR title:** `feat(subscriber): implement subscriber service and handler (single CRUD)`

---

### D7 — CSV Bulk Import

**Goal:** `POST /subscribers/bulk` accepts a CSV file, validates and inserts rows in a batch, and returns a summary — without ever failing the whole request over one bad row.

**Files:**
- `internal/modules/subscriber/subscriber.service.go` — add `BulkImport`
- `internal/modules/subscriber/subscriber.dto.go` — add `BulkImportResponse`

**Implementation notes:**
- **Locked decision: skip-and-report.** A single invalid or duplicate row is skipped and counted, not treated as a reason to fail the entire batch. This is why the response is a summary object with three counters instead of a single success/fail boolean — the shape of the response only makes sense once the semantics are fixed, which is why this had to be settled before writing any code here.
- Parse in the service layer. Validate each row (name + email present, valid email format). Insert in a **single batched transaction** — batching is for insert performance, not for all-or-nothing semantics; a row failing validation is filtered out *before* the transaction, it never causes a rollback of the rows that were valid.
- Response shape:
  ```json
  { "inserted": 940, "skipped_duplicates": 45, "invalid_rows": 15 }
  ```
- Returns `202 Accepted`. This is synchronous-but-fast (up to ~10k rows per PDR §10 Phase C notes), not an asynq job — reaching for the queue here would be solving a problem this phase doesn't have.

**Tests:** Integration tests: one clean batch (all rows valid), one mixed batch (valid + duplicate + malformed rows) asserting the three counters are exactly correct. This is the one genuinely novel piece of logic in the phase, so it gets real integration coverage, not just a service-level unit test.

**Acceptance criteria:**
- A 1,000-row CSV with a mix of valid, duplicate, and invalid rows completes in one request and returns correct counts
- No single bad row causes the whole import to fail

**PR title:** `feat(subscriber): add CSV bulk import with skip-and-report semantics`

---

### D8 — List Membership Endpoints

**Goal:** Subscribers can be added to and removed from lists, composing the List (D2) and Subscriber (D5) repositories.

**Files:**
- `internal/modules/list/list.service.go` — add `AddSubscriber`, `RemoveSubscriber`, `ListSubscribers`
- `internal/modules/list/list.handler.go` — add the three membership routes

**Implementation notes:**
- `POST /lists/:listId/subscribers/:subscriberId`, `DELETE /lists/:listId/subscribers/:subscriberId`, `GET /lists/:listId/subscribers`
- Both `listId` and `subscriberId` must be checked against `userID` **independently**. Without this, a user could reference a subscriber ID they don't own and attach it to a list they do own — ownership scoping on only one side of the relationship isn't enough.
- Adding a subscriber already on the list should be idempotent (no error), not a `409` — the join table's `UNIQUE KEY uq_list_subscriber` exists to prevent duplicate rows, not to punish a repeat request.

**Tests:** One integration test: add subscriber to list, confirm via `GET`, remove, confirm gone. This is the one place two resources interact through a join table, so it's worth proving the round-trip works through the full stack once — see §5.

**Acceptance criteria:**
- A subscriber can be added to and removed from a list
- A subscriber can belong to multiple lists simultaneously
- Referencing another user's list or subscriber in any membership call returns `403`

**PR title:** `feat(list): add list membership endpoints (add/remove/list subscribers)`

---

### D9 — Route Registration & Wiring

**Goal:** All list, subscriber, and membership routes are registered and reachable behind `JWTMiddleware`.

**Files:**
- `internal/routes/router.go`

**Implementation notes:**
- Standard Fx/Chi wiring, same pattern as Phase B's D7. All Phase C routes sit inside the protected router group that Phase B's D7 already set up — this PR shouldn't need to touch router-group middleware wiring itself, only add routes inside the existing group.

**Tests:** None new — this is wiring, exercised by every integration test in D3/D6/D7/D8 once routed.

**Acceptance criteria:**
- Every Phase C endpoint listed in the parent PDR's §11 API reference resolves through the router and enforces `JWTMiddleware`

**PR title:** `feat(routes): register list, subscriber, and membership routes`

---

### D10 — Integration Test Suite + Docs Close-Out

**Goal:** Confirm the whole phase holds together as a system, not just as individually-tested parts, and leave the repo in a state where Phase D can start immediately.

**Files:**
- `internal/modules/list/list_integration_test.go`, `internal/modules/subscriber/subscriber_integration_test.go` (fill any remaining gaps — e.g. a sequential flow like create list → add subscriber → CSV import more subscribers → remove one → confirm list count, as one continuous test)
- `README.md` — document the list, subscriber, and membership endpoints
- `.env.example` — no new vars expected this phase, confirm nothing was missed

**Implementation notes:**
- This is a checkpoint PR, not new feature work. If D1–D9 were built correctly, this PR should mostly be additive tests plus documentation, not bug fixes. If real bugs surface here, that's useful signal that an earlier PR's tests had a gap.
- Run the full suite with `-race` locally before pushing, matching CI's exact command from PDR §13.1.

**Tests:** Full `go test ./... -race -cover` run, confirming no regressions across the whole repo.

**Acceptance criteria:**
- All Phase C acceptance criteria from parent PDR §10 pass together, in sequence, in one test run
- `go test ./... -race -cover` is green
- CI is green on the PR
- README accurately describes list, subscriber, and membership endpoints

**PR title:** `test(lists-subscribers): full Phase C integration coverage + docs`

---

## 4. PR & Merge Discipline

- **One deliverable = one PR = one merge.** No bundling two deliverables into one PR, even if they feel small — the point is reviewability and the ability to `git bisect` cleanly if something breaks later.
- **Don't start Dn+1 until its dependencies are merged and CI is green** on `main`. D1–D3 (list) and D4–D6 (subscriber) can be developed in parallel since they don't share files, but D7 (CSV) needs D6 merged, and D8 (membership) needs both D2 and D5 merged.
- **Every PR description should state:** which deliverable it is (D1–D10), what it depends on, and which specific parent-PDR acceptance criteria (§10 Phase C) it satisfies — reviewers should be able to check the PR against the PDR line by line.
- **Commit hygiene:** within a single deliverable's branch, still split commits logically, same pattern used in Phase B.

---

## 5. Testing & Coverage Recap (scoped down from Phase B's default)

The parent PDR's default is unit tests for every service method and integration tests for every handler. For this phase, that's more than the risk profile justifies — most of these endpoints are structurally identical CRUD with a `userID` filter, and Phase B already proved the ownership-scoping pattern works. Full exhaustive integration coverage on every route here would mostly be re-proving the same pattern.

**What gets full coverage: the service layer.** This is where the business rules live — soft delete, duplicate handling, CSV row validation, batch transaction behavior. Every service method (D3, D6, D7, D8) gets a real unit test against the test DB. No mocking the DB or Redis, per PDR §14.1 — same rule as Phase B.

**What gets targeted integration coverage — not exhaustive:**

1. **Cross-user access, once per resource type** (D3, D6). One test hitting another user's list, one hitting another user's subscriber, asserting `403`. This is the test that actually matters this phase — it proves the ownership-scoping pattern works end-to-end through the handler, not just in isolation in the service test. Not repeated per-endpoint; if it holds for one read and one write path per resource, it holds for the rest, because they share the same repository method shape.
2. **CSV import happy path + one mixed-failure batch** (D7). Already specified in D7's own test section above.
3. **List membership add/remove round-trip** (D8). Already specified in D8's own test section above.

Everything else — basic create/read/update on lists and subscribers individually — is covered by service-layer unit tests plus one handler smoke test per module (D3, D6) to confirm routing and DTO serialization are wired correctly, not a full matrix of every status code on every endpoint.

**Coverage target:** 70% baseline (service layer), no auth-style carve-out — this is not the auth module.

---

## 6. Phase C Definition of Done

Phase C is complete when, in one continuous run:

- [ ] Full list CRUD works, all responses scoped to the authenticated user
- [ ] Full subscriber CRUD works
- [ ] CSV import with 1,000 mixed-validity rows completes and returns a correct summary (skip-and-report, not fail-whole-batch)
- [ ] A subscriber can belong to multiple lists
- [ ] Accessing another user's list or subscriber returns `403`, proven by integration test
- [ ] Service layer test coverage ≥ 70%
- [ ] `go test ./... -race -cover` passes locally and in CI
- [ ] All 10 deliverable PRs merged into `main` in order
- [ ] README reflects the finished list/subscriber/membership system

Once every box is checked, Phase D (Campaigns) can begin — it depends on real lists and subscribers existing, which is exactly what this phase delivers.