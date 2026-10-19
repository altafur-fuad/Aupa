# Aupa — Project Structure and Folder Responsibilities

**Purpose:** this is the canonical proposed monorepo layout for Aupa, a Flutter food-delivery application with a Go API, a future/private Rust dispatch worker, and Supabase PostgreSQL without Supabase Auth.

> Important: inspect the existing repository before applying this tree. If the Flutter UI already exists, preserve it and migrate gradually. This document describes the target structure; it does not mean every future directory must be created empty on day one.

## 1. Architecture at a glance

- **Flutter/Dart** renders the app and calls Aupa's HTTPS API.
- **Go** is the primary backend. It owns custom authentication, authorization, API contracts, catalog/cart/checkout, order state transitions, payment orchestration, webhooks and notifications.
- **Supabase PostgreSQL** is the database. Schema migrations are versioned in Git.
- **Rust** is a private worker for dispatch scoring/assignment proposals only, introduced after the core order flow works. Go revalidates every proposal and owns the final write.
- **Supabase Auth is not used.** Aupa's Go API issues and validates Aupa tokens.

Keep the first deployment simple: one Go API, one PostgreSQL database, and optionally one Rust worker when required. Begin with a transactional outbox in PostgreSQL rather than adopting a broker prematurely.

## 2. Canonical target tree

```text
Aupa/
├── AGENTS.md                         # Always-on repository instructions for Antigravity
├── README.md                         # Human setup, run, test and deploy overview
├── .gitignore                        # Build output, local env files, IDE state, secrets
├── .env.example                      # Placeholder variable names only; no real secrets
├── Makefile                          # Common local check/test targets (optional initially)
├── docker-compose.yml                # Optional local app services; never production secrets
│
├── apps/
│   └── customer_flutter/             # Existing customer-facing Flutter application
│       ├── pubspec.yaml
│       ├── analysis_options.yaml
│       ├── android/                  # Flutter-generated Android project
│       ├── ios/                      # Flutter-generated iOS project
│       ├── assets/
│       │   ├── images/
│       │   ├── icons/
│       │   └── fonts/
│       ├── l10n/                     # Localization resources
│       ├── lib/
│       │   ├── main.dart             # Minimal entry point
│       │   ├── bootstrap.dart        # App initialization and error hooks
│       │   ├── app/
│       │   │   ├── app.dart          # MaterialApp/CupertinoApp composition
│       │   │   ├── router.dart       # Route definitions and auth gates
│       │   │   └── theme/            # Design tokens, themes, typography
│       │   ├── core/
│       │   │   ├── config/           # Compile-time config; no secrets
│       │   │   ├── errors/           # Typed app/API errors
│       │   │   ├── network/          # HTTP client, interceptors, API base URL
│       │   │   ├── secure_storage/   # Platform-backed token storage adapter
│       │   │   ├── utils/            # Small, reusable utilities
│       │   │   └── widgets/          # Truly shared UI primitives
│       │   ├── features/
│       │   │   ├── auth/
│       │   │   │   ├── data/         # API data sources, DTOs, repository impls
│       │   │   │   ├── domain/       # Entities, repository contracts, use cases
│       │   │   │   └── presentation/ # Login/register/verification/reset screens
│       │   │   ├── home/
│       │   │   ├── restaurants/
│       │   │   ├── menu/
│       │   │   ├── search/
│       │   │   ├── cart/
│       │   │   ├── checkout/
│       │   │   ├── orders/
│       │   │   ├── delivery_tracking/
│       │   │   ├── addresses/
│       │   │   ├── profile/
│       │   │   └── notifications/
│       │   └── shared/               # Only code shared across multiple features
│       ├── test/                     # Unit and widget tests
│       └── integration_test/         # Key end-to-end client flows
│
├── services/
│   ├── api-go/                       # Main public backend
│   │   ├── go.mod
│   │   ├── go.sum
│   │   ├── Dockerfile
│   │   ├── cmd/
│   │   │   ├── api/main.go           # HTTP API entry point
│   │   │   └── worker/main.go        # Background outbox/provider worker; add when needed
│   │   ├── internal/
│   │   │   ├── app/                  # Dependency wiring and startup/shutdown
│   │   │   ├── config/               # Env/config parsing and validation
│   │   │   ├── httpapi/
│   │   │   │   ├── handlers/v1/      # Versioned HTTP handlers/DTO mapping
│   │   │   │   ├── middleware/       # Auth, request ID, limits, recovery, logging
│   │   │   │   └── response/         # Standard success/error envelopes
│   │   │   ├── auth/                 # Passwords, access JWT, refresh sessions, OTP
│   │   │   ├── users/                # Profile and account lifecycle
│   │   │   ├── restaurants/          # Restaurant and opening-hours use cases
│   │   │   ├── menus/                # Categories, items, options, availability
│   │   │   ├── cart/                 # Server-side cart/price validation
│   │   │   ├── orders/               # Checkout, order transitions, order history
│   │   │   ├── payments/             # Provider-independent payment use cases
│   │   │   ├── delivery/             # Delivery assignment and status policy
│   │   │   ├── notifications/        # Notification orchestration and retries
│   │   │   ├── database/postgres/    # Connection pool and transaction helpers
│   │   │   ├── events/outbox/        # Transactional outbox publisher/consumer
│   │   │   ├── integrations/         # payment, SMS/email, push, maps adapters
│   │   │   ├── security/             # Authorization policy, rate-limit helpers
│   │   │   └── observability/        # Structured logging, metrics, tracing hooks
│   │   ├── db/queries/               # Reviewed SQL query files for Go codegen/use
│   │   ├── internal/dbgen/           # Generated DB code if sqlc is selected
│   │   └── tests/integration/        # Database and HTTP integration tests
│   │
│   └── dispatch-rust/                # Private worker; defer until dispatch is defined
│       ├── Cargo.toml                # Rust workspace manifest
│       ├── Cargo.lock
│       ├── Dockerfile
│       ├── crates/
│       │   ├── dispatch-core/        # Pure scoring/eligibility logic, no network
│       │   │   ├── Cargo.toml
│       │   │   └── src/lib.rs
│       │   └── dispatch-worker/       # Worker process and adapters
│       │       ├── Cargo.toml
│       │       └── src/main.rs
│       └── tests/                    # Algorithm/worker tests
│
├── api/
│   └── openapi/openapi.yaml          # Versioned public API contract
│
├── supabase/
│   ├── config.toml                   # Supabase CLI local development config
│   ├── migrations/                   # Ordered SQL schema migrations; source-controlled
│   ├── seed.sql                      # Fake local/dev data only
│   └── tests/                        # SQL/database policy tests as appropriate
│
├── infra/
│   ├── docker/                       # Build/runtime container definitions if shared
│   ├── deploy/                       # Environment-specific deployment manifests
│   └── monitoring/                   # Dashboards/alerts/configuration templates
│
├── scripts/                          # Safe developer scripts; no secrets
│
├── docs/
│   ├── 1. PRD.md
│   ├── 2. DESIGN.md
│   ├── 3. ARCHITECTURE.md
│   ├── 4. DATABASE.md
│   ├── 5. API.md
│   ├── 6. AGENTS.md
│   ├── 7. Readme.md
│   ├── 8. Project Implementation.md
│   ├── 9. Setup.md
│   ├── 10. Feature Development.md
│   ├── 11. Authentication & Security.md
│   ├── 12. MCP Connection.md
│   ├── 13. Integrations.md
│   ├── 14. Testing & QA.md
│   ├── 15. Debugging.md
│   ├── PROJECT_STRUCTURE.md          # This file
│   └── ANTIGRAVITY_BOOTSTRAP_PROMPT.md# Copy/paste prompt to start the agent
│
└── .github/workflows/ci.yml          # Lint/test/validation CI; add with first code
```

## 3. What to create now vs later

### Create during the initial setup

- Root `AGENTS.md`, `README.md`, `.gitignore`, `.env.example`.
- The existing Flutter project and its current screens/assets, preserved as-is.
- `services/api-go/` with a minimal health endpoint, validated configuration, graceful shutdown and tests.
- `supabase/migrations/` and a first baseline schema only after reviewing `docs/4. DATABASE.md`.
- `api/openapi/openapi.yaml` with the first implemented routes.
- CI checks for whichever languages already have real code.

### Add when the feature requires it

- `cmd/worker/` and outbox processors when an asynchronous task needs durable retries.
- Payment/SMS/push/maps adapters only after the provider is selected.
- `services/dispatch-rust/` only when dispatch scoring has a defined contract and testable input/output. Until then, document the intended boundary but do not deploy an empty Rust service.
- `apps/merchant_flutter/`, `apps/rider_flutter/`, or an admin web app only when the product scope explicitly calls for separate clients. These can be added later without forcing the customer app to contain every workflow.
- Redis or an external message broker only after measurements show PostgreSQL outbox polling is insufficient.

## 4. Flutter feature convention

Use feature-first organization. A feature may use `data/`, `domain/`, and `presentation/` where this actually clarifies responsibilities. Do not split every tiny widget into layers just to follow a template. Keep current routing, state management and design system unless there is a concrete reason to change them.

- `data/`: remote data sources, request/response DTOs, repository implementations and mapping.
- `domain/`: app-specific entities, repository contracts and use cases (only when useful).
- `presentation/`: pages/screens, components and state/controller classes using the project's chosen state manager.
- `core/`: cross-cutting network/error/configuration/security facilities, not business-specific UI.

The Flutter client should call Aupa's Go API over HTTPS. In the MVP, it should not query Supabase tables directly.

## 5. Go service convention

- Keep Go module imports under `internal/` unless a package is intentionally a public library.
- `cmd/api/main.go` wires configuration, dependencies and the server; do not place business logic in `main.go`.
- HTTP handlers validate/parse input, call a use case, map the result to a stable response and return. Business rules belong in service/domain code.
- Repositories own persistence queries and transaction boundaries. Use parameterized SQL.
- Keep provider-specific code behind interfaces/adapters so provider changes do not affect order policy.
- Separate order status, payment status and delivery status. State changes must be validated and auditable.
- Use integer minor units for amounts, never floating-point. If the first market is Bangladesh, document BDT as the launch currency but keep currency semantics explicit.

## 6. Rust worker convention

Rust is not a second public API and not the owner of core business rules. Its primary task is deterministic dispatch scoring or candidate ranking from a defined, limited input. Start with pure, unit-testable logic in `dispatch-core`; keep I/O in `dispatch-worker`.

The worker may claim explicitly queued/outbox jobs and emit a proposal. Go rechecks current order state, rider availability, permissions and constraints before writing the final assignment. The Rust worker must not edit identity, user, order pricing, payment or arbitrary operational records. Give its database credential only the minimum permissions required for its actual interface.

## 7. Database and custom authentication guardrails

- Put schema changes in ordered `supabase/migrations/*.sql` files; do not make undocumented manual production changes.
- Keep seeds synthetic. Never load real production customer/address/payment data into development fixtures.
- Do not assume that Aupa's custom JWT automatically makes Supabase `auth.uid()` work. That function and RLS behavior depend on Supabase's supported JWT/auth context.
- In the initial architecture, Go authenticates the request and enforces object-level authorization; backend services connect to PostgreSQL using server-only credentials/roles. RLS can be defense in depth, but do not claim it protects queries executed with a privileged server role that bypasses it.
- If direct client-to-Supabase access is considered later, design and test an explicit supported external JWT trust configuration and policy claims before enabling it.
- Never ship a database password or Supabase secret/service-role key to Flutter.

## 8. Naming and repository rules

- Directories and code packages: lowercase, descriptive names with underscores only where conventional for that ecosystem (e.g. `delivery_tracking`).
- Go source: `gofmt`; Rust source: `rustfmt`; Dart source: `dart format`.
- Use `*_test.go`, Rust `#[test]`/integration tests, Dart `*_test.dart`.
- Name migration files with a sortable timestamp/sequence and a concise snake_case description, e.g. `202610190001_create_identity_tables.sql`.
- Never rename a public API field/endpoint, database column, or event type silently. Add a compatibility/migration note.
- Do not check in generated build folders such as `.dart_tool/`, `build/`, `target/`, `vendor/` unless a specific toolchain requires it.

## 9. Antigravity task procedure

For each request, Antigravity should:

1. Read root `AGENTS.md` and the relevant docs.
2. Inspect existing files and git changes before creating/editing.
3. State assumptions and the smallest implementation plan.
4. Implement one coherent slice, preserving existing UI.
5. Add/update tests, API docs, migration/docs as applicable.
6. Run available checks and report exact outcomes and unrun checks.
7. Highlight security, migration and provider configuration risks; ask before destructive or production operations.

Do not create all the tree's placeholder files in one giant change. Build a running vertical slice, test it, then extend it.

## 10. Authoritative references

- Antigravity Rules and instruction discovery: https://www.antigravity.google/docs/rules/
- Antigravity Projects: https://www.antigravity.google/docs/projects/
- Flutter app architecture: https://docs.flutter.dev/app-architecture
- Go modules and project structure: https://go.dev/doc/modules/layout
- Rust Cargo workspaces: https://doc.rust-lang.org/book/ch14-03-cargo-workspaces.html
- Supabase Row Level Security: https://supabase.com/docs/guides/database/postgres/row-level-security
- Supabase JWTs: https://supabase.com/docs/guides/auth/jwts
- OWASP Password Storage Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html
- OWASP JSON Web Token Cheat Sheet: https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html
