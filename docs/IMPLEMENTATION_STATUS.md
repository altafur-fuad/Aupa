# Aupa Implementation Status

> Antigravity updates this file after every task. Never mark work complete without actual verification.

## Current phase
- [x] Phase 0 — Audit and plan
- [x] Phase 1 — Frontend (Flutter/Dart)
- [x] Phase 2 — Backend (Go API / scoped Rust worker)
- [x] Phase 3 — Supabase PostgreSQL and backend DB wiring
- [ ] Phase 4 — Integration, security review and release QA

## Baseline audit
- Date/time: 2026-10-19T11:40:00+06:00
- Current git status: Initialized Git repository on `main` branch with remote `https://github.com/altafur-fuad/Aupa.git`.
- Flutter baseline checks:
  - Flutter 3.35.2 • channel stable • Dart 3.9.0 available on system.
  - `flutter analyze`: 8 info issues (`.withOpacity` deprecated; suggests `.withValues()`). 0 errors.
  - `flutter test`: 1 failed (stock counter test in `test/widget_test.dart` failing against `AupaApp` splash screen).
  - `dart format`: 10 files need formatting.
- Go baseline checks: `go` not currently in system PATH (`CommandNotFoundException`). Documented for Phase 2.
- Rust baseline checks: `cargo` not currently in system PATH (`CommandNotFoundException`). Scoped dispatch worker deferred per architecture rules.
- Database/migration baseline: Supabase PostgreSQL migrations not yet started; targeted for Phase 3.
- Existing UI/screens preserved:
  - `SplashScreen`: Brand, title, customer & owner entry paths.
  - `LoginScreen`: Email, password, forgot password, sign up button.
  - `SignupScreen`: Full name, email, password, navigates to OTP.
  - `OtpScreen`: 4-digit code verification, resend code.
  - `HomeScreen`: Delivery address bar, search input, horizontal categories, restaurant cards, bottom navigation.
  - `RestaurantDetailsScreen`: Hero header, ratings, reviews, delivery ETA, menu items.
  - `FoodDetailsScreen`: Item details, customization options, quantity selector, add-to-cart button.
  - `CartScreen`: Items list, promo code row, subtotal, delivery fee, total, proceed to checkout button.
  - `CheckoutScreen`: Delivery address selection, payment method selection, place order button, order confirmation dialog.
  - `OwnerDashboard`: Multi-branch dropdown, sales metrics, incoming orders with accept/reject.
  - `MenuManager`: Category sections, dish cards, prices, availability toggle.
- Differences against `docs/PROJECT_STRUCTURE.md`:
  - Flutter app is currently located at `aupa/`. Status: `KEEP` (preserve working app, avoid risky moves until needed).
  - Empty directory `aupa/New folder`. Status: `REMOVE ONLY IF CONFIRMED UNUSED`.

## Assumptions and decisions
| ID | Assumption/decision | Reason | Needs user confirmation? | Status |
|---|---|---|---|---|
| A01 | Keep Flutter app in `aupa/` | Prevents breaking native configs, assets, and project files | No (per AGENTS.md rule 3) | ACTIVE |
| A02 | Launch currency is BDT (Bangladeshi Taka), poisha minor units | PRD section 5 & API specs standard | No | ACTIVE |
| A03 | Frontend mock repository pattern with explicit mock/remote switch | Phase 1 requires working UI without backend dependency | No | ACTIVE |
| A04 | Replace template counter test with Aupa smoke & widget tests | Existing test `widget_test.dart` tests counter that doesn't exist | No | ACTIVE |

## Task queue

### Phase 0 — Audit and plan
- [x] P0-01 Inspect repository and preserve existing UI (acceptance criteria: current tree and protected UI recorded)
- [x] P0-02 Read relevant project docs and reconcile assumptions (acceptance criteria: gaps and assumptions recorded)
- [x] P0-03 Create ordered task backlog and baseline test report (acceptance criteria: each task has acceptance criteria)

### Phase 1 — Frontend (Flutter/Dart only)
- [x] P1-01 Fix baseline widget test and fix deprecated `.withOpacity` calls across screens (acceptance criteria: `flutter test` passes, `flutter analyze` has 0 warnings/infos).
- [x] P1-02 Establish core architecture in `aupa/lib/core`: environment config (mock/remote switch), failure/error models, currency/money formatters, and repository interfaces.
- [x] P1-03 Implement domain models & mock repositories for Auth, Restaurants, Menus, Cart, Orders, and Addresses with realistic BDT data.
- [x] P1-04 Enhance App Theme, design system tokens, and shared UI states (loading skeleton, empty state, error/retry widget, custom button).
- [x] P1-05 Upgrade Auth Flow (Sign-in, Sign-up, OTP verification, Forgot/Reset Password) with form validation, stateful feedback, and mock auth repository integration.
- [x] P1-06 Upgrade Restaurant Discovery & Search (Filter by category, search bar with debounce, dynamic restaurant cards, empty search state).
- [x] P1-07 Upgrade Menu & Item Customization (Live item options/add-on selection, dynamic price calculation, add to cart with feedback).
- [x] P1-08 Implement robust Client Cart State & Checkout Flow (Quantity management, item removal, live order summary, delivery address selection, payment method selection [COD / Online Pay], idempotent submit simulation).
- [x] P1-09 Implement Order Confirmation, Order History, Order Details, and Live Delivery Tracking UI with status stepper and ETA.
- [x] P1-10 Implement Customer Profile, Saved Addresses management (Add/Edit/Delete address), and Notifications screen.
- [x] P1-11 Integrate bottom navigation tabs (Home, Search, Orders, Profile) and verify complete customer journey end-to-end with mock data.
- [x] P1-12 Run full Phase 1 verification: `dart format`, `flutter analyze`, and comprehensive widget & unit tests for all customer flows.

### Phase 2 — Backend (Go API and scoped Rust component)
- [x] P2-01 Setup Go module `services/api-go`, configuration package, structured logging (`slog`), health/readiness endpoints (`/healthz`, `/readyz`), and graceful shutdown HTTP server.
- [x] P2-02 Implement HTTP middleware suite: Request ID (`X-Request-Id`), recovery handler, timeouts, CORS, and standardized success/error JSON response envelopes.
- [x] P2-03 Implement Custom Authentication Service: Password hashing (Argon2id/Bcrypt), JWT access token generation/claims verification, and cryptographically secure refresh token rotation & revocation with in-memory test doubles.
- [x] P2-04 Implement Role and Resource Ownership Authorization middleware (customer, merchant, rider, admin).
- [x] P2-05 Implement Restaurant & Menu Domain Services and endpoints (`GET /v1/restaurants`, `GET /v1/restaurants/{id}`, `GET /v1/restaurants/{id}/menu`).
- [x] P2-06 Implement Server-side Cart & Price Calculation: `/v1/checkout/quote` with poisha integer calculations, minimum order bounds, and promo verification.
- [x] P2-07 Implement Order Lifecycle & State Machine: Order creation with idempotency key enforcement, state transition validator (`placed` -> `delivered`/`cancelled`), history, and details endpoints.
- [x] P2-08 Implement Payment Provider Abstraction & Webhook Verification: Payment attempt orchestration with test double adapter and verified webhook handler.
- [x] P2-09 Implement Delivery Tracking & Notification Interfaces with timeline progression simulation.
- [x] P2-10 Scoped Rust component audit & decision: Explicitly verify if dispatch worker needs implementation or record formal deferral per AGENTS.md rules.

### Phase 3 — Supabase PostgreSQL
- [x] P3-01 Design versioned SQL migrations in `supabase/migrations/` for core tables: `users`, `refresh_tokens`, `restaurants`, `menu_items`, `menu_categories`, `addresses`, `orders`, `order_items`, `payment_attempts`, `idempotency_records`, `outbox_events`.
- [x] P3-02 Implement financial constraints (integer poisha check >= 0), state machine transition check constraints, unique indexes (idempotency, refresh tokens, active cart), and foreign keys with referential actions.
- [x] P3-03 Define Row Level Security (RLS) policies and least-privilege database roles with documented custom-auth boundary (backend-enforced authorization with defense-in-depth DB policies).
- [x] P3-04 Create reproducible development seed data (`supabase/seed.sql`) containing realistic Dhaka restaurants, menus, sample customer, and address records with zero production/real user data.
- [x] P3-05 Implement PostgreSQL repository adapter in Go (`internal/db/postgres`) using `pgx` with parameterized queries, connection pooling, and query timeouts.
- [x] P3-06 Wire PostgreSQL repositories into Go API `cmd/api/main.go` and verify integration tests against repository interfaces.

### Phase 4 — Integration and QA
- [x] P4-01 Implement Flutter HTTP API Client & Remote Repository Adapter (`aupa/lib/data/remote/`) connecting to Aupa Go API (`/v1`).
- [x] P4-02 End-to-end critical flow verification: Auth, Discovery, Server-authoritative quote, Idempotent order creation, Payment checkout, and Delivery tracking.
- [ ] P4-03 Security review & edge-case testing: Negative authorization tests, rate limits, token replay detection, forged webhook signature rejection, expired token handling, input bounds validation.
- [ ] P4-04 Production deployment readiness: Run formatting, static analysis, unit & widget tests across Flutter and Go toolchains; document backup/restore, migration rollback/forward strategy, and environment configuration review.

## Iteration log
### Latest task
- Task ID: P4-02
- Status: DONE
- Goal and acceptance criteria: End-to-end critical flow verification: Auth, Discovery, Server-authoritative quote, Idempotent order creation, Payment checkout, and Delivery tracking.
- Files changed: `services/api-go/internal/httpapi/router.go`, `services/api-go/cmd/api/main.go`, `services/api-go/internal/httpapi/e2e_test.go`.
- Commands actually run:
  - `go test -v -count=1 ./internal/httpapi`: 11/11 subtests passed (`TestEndToEndCriticalCustomerFlow`).
  - `go test -count=1 ./...`: 100% pass across all packages in `services/api-go`.
  - `go vet ./...`: 0 issues found.
  - `flutter test`: 73/73 tests passed.
- Actual outcomes: Extracted reusable HTTP router factory `NewRouter` in `services/api-go/internal/httpapi/router.go` with repository dependency injection and fallback test doubles. Implemented end-to-end HTTP integration test suite (`e2e_test.go`) verifying all 11 critical stages: Health probes, user registration, authenticated `/v1/me` profile retrieval, JWT refresh with replay attack protection, restaurant & menu catalog discovery, server-authoritative quote with promo discount in poisha, idempotent order creation with replay vs conflict handling, payment checkout initiation, HMAC-SHA256 verified webhook reconciliation with order transition to accepted, and delivery tracking progression with courier coordinates.
- Known limitations: None.
- Next task: P4-03 Security review & edge-case testing: Negative authorization tests, rate limits, token replay detection, forged webhook signature rejection, expired token handling, input bounds validation.

## Blockers
| Blocker | Impact | What is needed | Safe independent work available? | Status |
|---|---|---|---|---|
| None for Phase 1 | None | Flutter SDK 3.35.2 is ready and functional | Yes, full Flutter client development | RESOLVED |

## Final release checklist
- [ ] Flutter formatting, analysis, tests, and build verified
- [ ] Go formatting, tests, and build verified
- [ ] Rust tests/build verified or Rust deferral documented
- [ ] Database migrations/seed/test verified
- [ ] Authentication and authorization negative tests passed
- [ ] End-to-end critical flows tested
- [ ] Secrets scan / configuration review performed where tooling is available
- [ ] Backup/restore and rollback instructions documented; execution status stated honestly
- [ ] No release-blocking known issues hidden
