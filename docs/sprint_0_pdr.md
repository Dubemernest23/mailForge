# PDR - Sprint 0: Foundation Polish & Cleanup

## 1. Sprint Overview
**Sprint Name:** Foundation Polish & Cleanup  
**Sprint Number:** 0  
**Duration:** 3-5 days  
**Status:** Ready to Start  
**Project Lead:** duby

**Goal:**  
Clean up the existing codebase to align with the simplified product vision (no Organizations, no GraphQL). Create a solid, consistent foundation with proper models so future sprints can move quickly and smoothly.

## 2. Product Context (Reminder)
MailForge is a simple email campaign tool for normal individuals and small business owners (birthday invites, customer notifications, discounts, etc.). We are removing organization/multi-tenancy complexity to keep the product lightweight and user-friendly.

## 3. Scope & Key Objectives

### In Scope
- Remove all Organization-related code and references
- Create proper Bun model structs for all database tables
- Update project documentation (README, architecture)
- Ensure migrations work cleanly with new models
- Improve consistency in project structure
- Update DI container and router if needed after cleanup
- Create a clean architecture diagram (text + recommendation for visual)

### Out of Scope
- New business features (Auth, Campaigns, etc.)
- Frontend work
- Email sending logic
- Tests beyond foundation level

## 4. Detailed Tasks

1. **Code Cleanup**
   - Delete or comment out `internal/modules/organization/`
   - Remove any references to Organization in DI container, router, models, etc.
   - Update package names and imports if affected

2. **Bun Models**
   - Create `internal/models/` directory
   - Add model files:
     - `user.go`
     - `list.go`
     - `subscriber.go`
     - `campaign.go`
     - (Optionally `list_subscriber.go` for join table)
   - Models should include proper Bun tags, timestamps, relationships where appropriate

3. **Database & Migrations**
   - Verify all migrations run cleanly
   - Ensure models match migration schema exactly
   - Test basic CRUD operations manually if needed

4. **Documentation Updates**
   - Update `README.md` with new simplified architecture
   - Update `sprints.md` if needed
   - Add architecture diagram (text version + PNG recommendation)

5. **Project Structure Polish**
   - Ensure clean folder layout
   - Update any scaffolded files that reference removed modules
   - Run `go mod tidy`

6. **Testing & Validation**
   - Run full test suite
   - Run migrations up/down
   - Confirm app starts without errors

## 5. Definitions & Terminology

- **Bun Model**: Struct with `bun:"table:..."` tag and field tags for columns
- **Single User Mode**: Each user has their own lists and campaigns (no org scoping)
- **Foundation**: All infrastructure needed before building real features

## 6. Deliverables

### Must Have
- `internal/models/` with all required Bun models
- Organization module fully removed
- Updated `README.md` with correct architecture description
- Working `make dev`, `make migrate-up`, and `go test ./...`
- New architecture diagram (at least text version in README)

### Nice to Have
- Basic relationship methods in models (e.g., User has Lists)
- Updated `analysis.md` or `walkthrough.md` if present

## 7. Acceptance Criteria
- App starts successfully (`go run ./cmd/api`)
- Migrations run without errors
- No compilation errors after cleanup
- No remaining references to "organization" in active code
- Models are ready for use in Sprint 1 (Auth)

## 8. Risks & Dependencies
- Risk: Breaking existing migrations → Mitigate by testing thoroughly
- Dependency: None (builds directly on current foundation)

## 9. Next Steps After This Sprint
- Kick off Sprint 1: Authentication
- Create PDR-Sprint-1.md

---

**Approval:**  
Ready for execution.  

**Project Lead Note:**  
This sprint is mostly cleanup and setup. Completing it well will make the next sprints much faster and less frustrating. Let's keep things calm and deliver a clean base.