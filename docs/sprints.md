# MailForge Sprints

This document contains the high-level sprint plan for MailForge. Each sprint will have its own detailed **PDR (Product Definition & Requirements)** document.

## Project Goal
Build a simple, easy-to-use email campaign tool for **normal people** and **small business owners** (e.g., tailors, shoemakers, individuals sending birthday invites, etc.). No complex organization/multi-tenancy.

## Sprint Structure

### Sprint 0: Foundation Polish & Cleanup
**Duration:** 3-5 days  
**Focus:** Clean up the current codebase and prepare a solid base.

**Key Tasks:**
- Remove Organization module completely
- Create proper Bun model structs for all tables
- Update architecture diagram and documentation
- Improve README and project setup
- Fix any remaining inconsistencies

**Deliverables:**
- Clean project structure
- Bun models in `internal/models/`
- Updated README with new architecture
- Working migrations and local setup

---

### Sprint 1: Authentication & User Management
**Duration:** 7-10 days  
**Focus:** Secure user accounts.

**Key Tasks:**
- User registration (email + password)
- Login with JWT
- Password hashing + basic profile
- Protected routes middleware
- Basic error handling & validation

**Deliverables:**
- Working Auth API endpoints
- JWT middleware
- Relevant tests
- PDR-Sprint-1.md completed

---

### Sprint 2: Contacts & Lists
**Duration:** 8-10 days  
**Focus:** Allow users to manage their email contacts.

**Key Tasks:**
- Create, list, update, delete Lists
- Add/remove subscribers to lists
- Email validation & deduplication
- Simple subscriber CRUD
- Basic import preparation (CSV later)

**Deliverables:**
- Full Contacts/List API
- Bun models fully utilized
- Tests for core operations

---

### Sprint 3: Campaigns – Core Management
**Duration:** 8-10 days  
**Focus:** Campaign creation and management.

**Key Tasks:**
- Create, edit, list, delete campaigns
- Subject, HTML content, plain text version
- Associate campaign with a list
- Campaign status (draft, scheduled, sent)
- Preview functionality

**Deliverables:**
- Campaign CRUD API
- Basic template support

---

### Sprint 4: Email Sending & Delivery Engine
**Duration:** 10-12 days  
**Focus:** Make sending emails actually work.

**Key Tasks:**
- Implement job queue (Redis recommended)
- Email worker implementation
- Provider abstraction layer (Resend + SMTP)
- Campaign sending logic
- Basic delivery status tracking

**Deliverables:**
- End-to-end campaign sending
- Worker system running

---

### Sprint 5: Polish, Tracking & MVP Release
**Duration:** 7-10 days  
**Focus:** Make the product reliable and usable.

**Key Tasks:**
- Delivery tracking (sent, failed, bounces)
- Unsubscribe handling
- Rate limiting & safety features
- Basic analytics endpoints
- Documentation & testing

**Deliverables:**
- Usable MVP
- Updated README with API examples
- Deployment ready

---

## Future Sprints (After MVP)
- Sprint 6: Scheduling & Templates gallery
- Sprint 7: Better analytics & reporting
- Sprint 8: CSV Import / Export
- Sprint 9: Webhooks & advanced tracking
- Sprint 10: UI (if building frontend) + polish

---

**Notes:**
- Every sprint will have its own PDR-*.md file with detailed requirements, definitions, and acceptance criteria.
- We will review and adjust scope before starting each sprint.
- Priority is on delivering working, tested features quickly.

---

**Current Status:** We are about to start **Sprint 0**.