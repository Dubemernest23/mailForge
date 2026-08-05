# MailForge — Phase C PDR: Lists & Subscribers

> **Status:** Active
> **Depends on:** Phase B (Auth & Identity) — complete, merged to `main`, CI green
> **Blocks:** Phase D (Campaigns) — campaigns target lists, so this phase must ship first
> **New migrations:** None — `lists`, `subscribers`, `list_subscribers` already exist as models from Phase A
> **Coverage target:** 70% baseline (service layer). No auth-style carve-out here — this is not the auth module.

---

## Goal

Authenticated users can create and manage email lists, add subscribers (manually or via CSV), and manage which subscribers belong to which lists. Every read and write is scoped to the authenticated user — a user must never be able to see, edit, or reference another user's list or subscriber, even indirectly.

## Why this phase, and why now

Campaigns target lists. A campaign built against no real list/subscriber data is untestable in any meaningful way. This phase exists so that by the time Phase D starts, the campaign layer has real lists and real subscribers to attach to, instead of being built and tested against fixtures that don't reflect the actual data shape.

The other reason this phase matters more than its size suggests: it's where the **ownership-scoping pattern gets established for every module after it**. Auth (Phase B) didn't need this pattern — a user only ever acts on their own account. Lists and subscribers are the first resources where "does this row belong to the person asking for it?" becomes a real question, asked on every single endpoint. Get the pattern right here — `userID` explicit in every repository method signature, never inferred — and campaigns, send jobs, and analytics all just copy it. Get it wrong here and it has to be retrofitted everywhere downstream, which is a much more expensive fix.

---

## Deliverable Sequence

### D1 — List DTOs + repository

- `CreateListRequest`, `UpdateListRequest`, `ListResponse` → `internal/modules/list/list.dto.go`
- Repository methods: `Create`, `FindAll(ctx, userID)`, `FindByID(ctx, userID, publicID)`, `Update`, `Archive`
- **Rule:** every method takes `userID` as an explicit parameter. Never infer it from context inside the repository. This is what turns an ownership bug into something visible in a method signature during code review, instead of something buried in a `WHERE` clause that has to be hunted for.

### D2 — List service + handler

- Service enforces: archiving a list is a status flip (`active → archived`), never a row delete.
- Handler returns **403**, not 404, when a list exists but belongs to a different user. 404 would tell an attacker whether the resource exists at all — 403 doesn't leak that.

### D3 — Subscriber DTOs + repository

- `CreateSubscriberRequest`, `UpdateSubscriberRequest`, `SubscriberResponse`
- Same explicit-`userID` repository pattern as D1.
- Schema constraint to respect: `UNIQUE KEY uq_user_email (user_id, email)`. A duplicate-email insert should map to the `ErrDuplicate` sentinel → `409 Conflict` in the handler, not a raw MySQL constraint error leaking through.

### D4 — Subscriber service + handler (single add / update / get)

- Delete is a status flip to `unsubscribed`, never a `DELETE` statement. Same reasoning as the PDR's soft-delete rule elsewhere: a hard-deleted subscriber corrupts historical open/click/bounce counts for any campaign they were once part of. The row has to survive forever; only the status changes.

### D5 — CSV bulk import

- `POST /subscribers/bulk`
- Parse in the service layer. Validate each row (name + email present, valid email format). Insert in a **single batched transaction**. Return a summary:
  ```json
  { "inserted": 940, "skipped_duplicates": 45, "invalid_rows": 15 }
  ```
- Returns `202 Accepted`. This is synchronous-but-fast (up to ~10k rows), not an asynq job — don't reach for the queue here, that's solving a problem this phase doesn't have.
- **Failure semantics (confirm before building):** a single invalid or duplicate row should be *skipped and reported*, not fail the whole batch. That's the entire reason the summary object has three separate counters instead of a single success/fail boolean.

### D6 — List membership endpoints

- `POST /lists/:listId/subscribers/:subscriberId`
- `DELETE /lists/:listId/subscribers/:subscriberId`
- `GET /lists/:listId/subscribers`
- Both `listId` and `subscriberId` must be checked against `userID` independently. Without this, a user could reference a subscriber ID they don't own and attach it to a list they do own — ownership scoping on only one side of the relationship isn't enough.

### D7 — Route registration + wiring

Standard Fx/Chi wiring into `internal/routes/router.go`, same pattern as Phase B.

---

## Testing Strategy (scoped down from PDR default)

The PDR's default is unit tests for every service method and integration tests for every handler. For this phase, that's more than the risk profile justifies — most of these endpoints are structurally identical CRUD with a `userID` filter. Full exhaustive integration coverage on every route is expensive to write and mostly redundant once the pattern is proven once or twice.

**What actually gets full coverage: the service layer.** This is where the business rules live — soft delete, duplicate handling, CSV row validation, batch transaction behavior. Every service method gets a real unit test against the test DB, no shortcuts, no mocking.

**What gets targeted integration coverage — not exhaustive:**

1. **Cross-user access, once per resource type.** One test hitting another user's list, one hitting another user's subscriber, asserting `403`. This is the test that actually matters this phase — it's the one that proves the ownership-scoping pattern works end-to-end through the handler, not just in isolation in the service test. Don't repeat this per-endpoint; if the pattern holds for one read and one write path per resource, it holds for the rest, because they share the same repository method shape.
2. **CSV import happy path + one failure mode.** One test with a clean batch, one test with a mixed batch (some valid, some duplicate, some invalid) asserting the summary counts are correct. This is the one genuinely novel piece of logic in the phase — worth a real integration test, not just a service-level one.
3. **List membership add/remove round-trip.** One test: add subscriber to list, confirm via `GET`, remove, confirm gone. This is the one place two resources interact, so it's worth proving the join actually works through the full stack once.

Everything else (basic create/read/update on lists and subscribers individually) is covered by the service-layer unit tests plus one handler smoke test per module to confirm routing and DTO serialization are wired correctly — not a full matrix of every status code on every endpoint.

**Target:** 70% baseline coverage, service layer weighted toward real coverage rather than the integration suite chasing the same number.

---

## Acceptance Criteria

- Full list CRUD works, all responses scoped to the authenticated user
- Full subscriber CRUD works
- CSV import with 1,000 rows completes and returns a correct summary
- A subscriber can belong to multiple lists
- Accessing another user's list or subscriber returns `403` (proven via integration test, not just service-level assumption)
- Service layer at 70%+ coverage; targeted integration tests pass (cross-user access, CSV import, list membership round-trip)

---

## Open Decision Before D5

CSV row-failure semantics: **skip-and-report** (recommended, matches the summary object shape) vs. **fail-whole-batch**. This has to be locked before D5 starts — it changes both the service contract and the response DTO.


gofmt -l . 
go vet ./...
golangci-lint run ./...
make swag && git status internal/docs/
go build ./...
go test ./... -cover -p 1