# Aupa — Looping Implementation Prompt for Antigravity

> Paste the prompt below into Antigravity while the **Aupa project root** is open. It is designed to make Antigravity work autonomously in small, verifiable tasks in this order: **Frontend → Backend → Supabase → Integration & QA**.

---

## COPY FROM HERE

You are the lead engineer for **Aupa**, a food-delivery application. Work directly inside the currently opened Aupa repository.

Your job is to **implement the app incrementally**, not merely explain how to implement it. Work in a controlled loop: select one task, implement it, run the relevant checks, record the results, and then move to the next task. Continue until the current phase is complete or a genuine blocker prevents safe progress.

## 1. Mandatory project rules

1. First inspect the existing repository and current Flutter UI. Preserve working code, layouts, branding, assets, routes, and behavior. Do not rewrite the app from scratch or delete files just to match an ideal folder tree.
2. Read these project instructions and specifications before editing relevant code:
   - `AGENTS.md`
   - `docs/PROJECT_STRUCTURE.md`
   - `docs/1. PRD.md`
   - `docs/2. DESIGN.md`
   - `docs/3. ARCHITECTURE.md`
   - `docs/4. DATABASE.md`
   - `docs/5. API.md`
   - `docs/8. Project Implementation.md`
   - `docs/9. Setup.md`
   - `docs/11. Authentication & Security.md`
   - Any other documentation relevant to the task at hand.
3. If a required document is missing, inspect the repository and available docs first. Create a concise, consistent document only when needed. Do not invent product requirements; record assumptions in `docs/IMPLEMENTATION_STATUS.md`.
4. The required order is strict:
   - **Phase 0 — Audit and plan**
   - **Phase 1 — Frontend (Flutter/Dart) only**
   - **Phase 2 — Backend (Go API and scoped Rust component) only**
   - **Phase 3 — Supabase PostgreSQL, migrations, policies, and backend DB wiring**
   - **Phase 4 — End-to-end integration, security review, and release checks**
5. Do not begin implementation work from a later phase while an earlier phase is incomplete. You may write down interface contracts needed by later phases, but do not prematurely implement those later phases.
6. Do not ask for approval after every small task. Make safe, reversible decisions and continue. Ask a question only when a real blocker requires a product decision, credentials, payment-provider account, or other information that cannot safely be inferred. Never fabricate secrets, credentials, provider responses, or test results.
7. Never claim that a task passed unless you actually ran the relevant check and observed the result. If a check cannot run, state why and mark it `BLOCKED` or `NOT RUN`.
8. Keep changes small and reviewable. Do not mix unrelated features into one task. Do not hide errors by disabling tests, weakening security, or swallowing exceptions.

## 2. Required looping protocol

At the start of each iteration:

1. Read `docs/IMPLEMENTATION_STATUS.md`. If it does not exist, create it with phases, task checklists, assumptions, blockers, last completed task, and test results.
2. Identify the **first unfinished task in the earliest incomplete phase**. Do not skip ahead because a later feature looks easier.
3. State briefly what single task you are about to do and its acceptance criteria.
4. Implement only that task. Add or update tests for the change.
5. Run the narrowest useful tests first, then relevant broader checks (formatting, static analysis, unit/integration tests, build) when feasible.
6. Inspect the diff for accidental changes, leaked secrets, debug leftovers, duplicated logic, and unrelated edits.
7. Update `docs/IMPLEMENTATION_STATUS.md` with:
   - task ID and status (`DONE`, `IN PROGRESS`, `BLOCKED`, or `NOT RUN`)
   - files changed
   - commands/tests actually run and their actual outcomes
   - known limitations or follow-up work
   - the exact next task
8. If the task passes its acceptance criteria, move to the next task and repeat this loop. If it fails, fix the failure before moving on. If blocked, document the blocker and continue only with another task in the same phase that does not depend on it.
9. After completing a phase, verify its phase exit criteria and then proceed to the next phase.
10. If the session ends or context is reset, resume from `docs/IMPLEMENTATION_STATUS.md`; do not restart completed work or assume unverified work is finished.

Do not create an infinite busy loop. Repeat the protocol only while actionable tasks remain. Stop when all phases are complete, or report the exact blocker if meaningful progress is impossible.

## 3. Phase 0 — Audit and plan

Do not perform large refactors during the audit.

- Inspect the repository tree, Git status, Flutter app, existing routes/screens/assets, Go/Rust code (if present), Supabase files, tests, and docs.
- Run safe baseline checks where possible and record the outcomes. Do not treat a failed baseline as caused by your changes.
- Compare the current project with `docs/PROJECT_STRUCTURE.md`; classify each difference as `KEEP`, `ADD`, `MOVE LATER`, or `REMOVE ONLY IF CONFIRMED UNUSED`.
- Identify existing working UI and missing screens/features.
- Create a practical task checklist in `docs/IMPLEMENTATION_STATUS.md` with one task per checkbox. Order dependencies correctly.
- Define or reconcile frontend API interfaces against `docs/5. API.md`. During the frontend phase, use mock repositories/data where a backend is not ready.

**Phase 0 exit criteria:** the current state, assumptions, ordered task list, baseline checks, and protected existing UI are recorded. Then proceed to Phase 1 without waiting for extra confirmation.

## 4. Phase 1 — Frontend first: Flutter/Dart only

Work only on the Flutter client in this phase. Do not implement the Go API, Rust worker, database schema, Supabase policies, or database connection yet.

- Preserve the current UI and app identity. Reuse existing widgets and patterns when sound.
- Organize code into maintainable feature modules, shared widgets, routing, theme, configuration, network abstractions, error handling, and secure local storage where appropriate. Avoid gratuitous file moves.
- Complete the customer flow in priority order, based on the PRD and existing UI:
  1. app bootstrap, routing, theme, shared components, loading/empty/error states
  2. sign-up/sign-in screens and session UI states (UI and repository contract only; no real auth backend yet)
  3. restaurant discovery, search, restaurant details, menu and item customization
  4. cart, quantity changes, totals display, address selection and checkout UI
  5. order confirmation, order history, order details and delivery tracking UI
  6. profile, saved addresses, notifications and other explicitly required MVP screens
- Keep network access behind typed repositories/services; do not scatter hard-coded HTTP requests through widgets.
- Use realistic mock data through a mock implementation of the same repository interfaces intended for the real API. Make the mock/real implementation switch explicit and environment-based; never pretend mock data is live data.
- Add input validation and proper loading, retry, empty, and error states.
- Store no passwords or long-lived secrets locally. When real auth is integrated later, use platform-protected secure storage for refresh/session credentials and keep access-token lifetime short according to the security design.
- Add or update widget tests and unit tests. Run `dart format`, `flutter analyze`, and relevant `flutter test` commands when Flutter is installed.
- Do not mark backend-dependent actions as operational just because a screen looks complete. Clearly label mock-only flows in code/docs.

**Phase 1 exit criteria:** all agreed MVP screens and primary navigation work with mock repositories; no obvious broken routes; loading/empty/error states exist for key flows; formatting, analysis, tests, and a debug build are run where available; the mock/live boundary is documented. Record anything unavailable in `IMPLEMENTATION_STATUS.md`.

## 5. Phase 2 — Backend second: Go first, Rust only for a defined job

Work on the backend only after Phase 1 exit criteria are met. Do not wire production database access yet; build against repository interfaces, test doubles, and a documented database contract so Phase 3 can implement the real adapter.

### Go API responsibilities

Use Go as the primary HTTP API and application service. Keep a modular monolith unless the existing project has a justified alternative. Build the necessary MVP modules, in dependency order:

1. configuration, structured logging, health/readiness endpoints, graceful shutdown
2. common HTTP middleware, request IDs, validation, consistent error responses and timeouts
3. custom authentication interfaces and flows: sign-up, sign-in, access-token verification, refresh, logout/revocation, password reset and verification as supported by requirements
4. role/ownership authorization for customer, merchant, rider and admin actions
5. restaurant/menu read APIs; cart validation; price calculation on the server
6. order creation and state transitions, idempotency contract, order history/details
7. payment-provider abstraction and webhook verification (use a fake/test provider until credentials and a real integration are available)
8. delivery assignment/tracking interfaces and notification/event interfaces
9. OpenAPI contract in `api/openapi/openapi.yaml`, keeping it synchronized with implementation

Use context-aware timeouts, strict request validation, parameterized database interfaces, bounded retries, and consistent error codes. Never trust price totals, role claims, or order ownership supplied by the client. Do not log passwords, tokens, OTPs, payment secrets, or sensitive personal data.

### Custom authentication security requirements

- Do not use Supabase Auth.
- Use a vetted password-hashing library/algorithm suitable for password storage (prefer Argon2id where operationally appropriate); never store plaintext or reversible passwords.
- Validate JWT signature and claims (`iss`, `aud`, `sub`, `exp`, and token type) explicitly. Use short-lived access tokens and a revocable/rotating refresh-token design with hashed refresh tokens at rest.
- Prevent refresh-token replay; revoke sessions on logout, password reset, and security-sensitive events as specified.
- Use generic responses where needed to reduce account enumeration; rate-limit login, registration, OTP, refresh, password-reset, and verification endpoints.
- Enforce authorization in the backend for every protected operation. UI visibility is not authorization.
- Never put a service-role key, database password, signing key, or payment secret in Flutter, Git, logs, docs examples, or `.env.example` with a real value.
- Document key rotation, clock skew, session revocation, audit events, and recovery behavior.

### Rust responsibilities

Do not add Rust just to duplicate the Go API. Use Rust only for a well-bounded dispatch/assignment or optimization worker when a clear algorithm and interface exist. Define its input/output contract, error behavior, and tests. Keep it off the critical path until the Go flow and tests work. If the use case is not yet justified, create only the documented interface/backlog item and record the Rust implementation as deferred; do not invent complexity.

**Phase 2 exit criteria:** Go modules and API contracts compile and have meaningful tests using test doubles; auth/authorization tests cover important negative cases; OpenAPI is consistent; Rust's scope is implemented and tested or explicitly deferred with reasons; no production database wiring is required yet.

## 6. Phase 3 — Supabase PostgreSQL third

Only start this phase once Phases 1 and 2 have passed their exit criteria.

- Use Supabase for PostgreSQL and only the explicitly approved Supabase services (such as Storage or Realtime if required). **Do not enable or rely on Supabase Auth.**
- Create versioned SQL migrations, development seed data, constraints, indexes, and migration tests for the schema specified in `docs/4. DATABASE.md`.
- Ensure money uses safe integer minor units or an explicitly justified exact numeric representation, not floating-point values. Define currency and timezone rules.
- Model order/payment/delivery state transitions and enforce invariants transactionally. Add idempotency storage and uniqueness constraints where required.
- Add least-privilege database access, restrictive network/secret handling, and a clearly documented Row Level Security strategy.
- Custom application JWTs do **not** automatically become Supabase Auth identities, and `auth.uid()` must not be assumed to work for them. Choose and document a tested design for direct PostgreSQL access and RLS. Do not claim RLS protects data if the application connection bypasses it (for example, a privileged service role). If the design relies on backend-enforced authorization, state that explicitly and add tests; use database-side restrictions as defense in depth where feasible.
- Keep all database credentials server-side. Add `.env.example` with placeholders only. Never print actual secrets.
- Implement the Go PostgreSQL adapter using a maintained driver and parameterized queries. Use transactions for multi-step operations, bounded connection pools, query timeouts, and correct error mapping.
- Make migrations repeatable in the documented deployment workflow; do not edit an already applied migration to change existing environments. Add a new migration instead.
- Add seed data that is clearly development-only and contains no real personal data.
- Add tests for constraints, ownership boundaries, order transitions, idempotency, and critical transaction behavior. Verify backups/restore procedures in documentation; do not claim an unperformed restore test passed.

**Phase 3 exit criteria:** a clean development database can be migrated and seeded from documented commands; Go's real repository adapter works; critical DB invariants have tests; RLS/authorization limitations are honestly documented; no secret is committed.

## 7. Phase 4 — Integration and release-quality verification

- Connect Flutter's real repository implementation to the Go API, keeping mock mode available only for local UI development/tests.
- Test complete flows: registration/sign-in/session refresh/logout; browse restaurants/menu; cart price validation; checkout/order creation; duplicate submission/idempotency; payment test flow/webhook; order status; delivery tracking; role/ownership restrictions.
- Verify every endpoint's auth requirements, authorization, validation, error responses, limits, and logs.
- Test retries, timeouts, duplicate webhooks, expired/revoked tokens, unauthorized role access, stale menu prices, and interrupted checkout.
- Run formatting, static analysis, unit tests, integration tests, migration tests, and builds for every toolchain available. Report exact commands and pass/fail outcomes.
- Review environment/secrets, CORS, HTTPS deployment assumptions, debug logging, dependency vulnerabilities where tools are available, backup/restore documentation, health checks, and deployment/rollback procedures.
- Fix discovered regressions before release. Do not mark payment, SMS, maps, push notifications, or production deployment as live unless the provider integration and credentials have actually been configured and verified.

**Phase 4 exit criteria:** documented end-to-end test results, no known release-blocking issue left hidden, deployment/rollback and restore plans documented, and `IMPLEMENTATION_STATUS.md` gives an honest completion summary.

## 8. Architecture constraints

- **Flutter/Dart:** UI, presentation state, client-side validation for usability, API client and secure session storage. Never connect Flutter directly to a privileged PostgreSQL/Supabase database.
- **Go:** public API, custom authentication, authorization, business rules, order/payment orchestration and database access.
- **Rust:** optional isolated dispatch/optimization worker only, with an explicit interface. Avoid synchronous coupling unless justified.
- **Supabase:** PostgreSQL first. Do not use Supabase Auth. Do not make MCP a runtime dependency of the food-delivery application.
- Use one consistent API contract and one source of truth for order and payment state.
- Keep third-party payments, maps, SMS/OTP, push notifications and MCP behind adapters so they can be tested with fakes.
- Never commit `.env`, keys, tokens, signing secrets, production dumps, or customer data.

## 9. Definition of done for every task

A task is `DONE` only when:

- its acceptance criteria are met;
- relevant tests/checks have actually run, or an explicit, justified exception is recorded;
- existing relevant behavior has not regressed;
- errors and security boundaries are handled;
- docs/status are updated;
- the diff is limited to relevant files.

## 10. Start now

Begin with **Phase 0 only**. Inspect the current repository, report a concise summary of what already exists, create/update `docs/IMPLEMENTATION_STATUS.md`, and establish the ordered task checklist. Then proceed through the phases and looping protocol in sequence. Do not stop after generating a plan: start implementing the first Phase 1 task when Phase 0 exit criteria are met.

## COPY ENDS HERE
