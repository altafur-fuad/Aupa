# Aupa — Antigravity Repository Instructions

These instructions apply to the Aupa food-delivery project. Follow them for every task in this repository.

## 1. Product and stack

- Product name: **Aupa**.
- Mobile client: Flutter + Dart.
- Main backend: Go; it owns the public API, custom authentication, business rules and authoritative order/payment transitions.
- Database: Supabase-hosted PostgreSQL. **Do not use Supabase Auth.** Authentication is implemented by Aupa's Go backend.
- Rust: a private delivery-dispatch/optimization worker, introduced only when its job is defined and useful. Do not create Rust just to have a Rust folder.
- API style: versioned REST under `/v1`, described in `api/openapi/openapi.yaml`.

## 2. Read before editing

1. Inspect the current repository and `git status` before making changes.
2. Read `docs/PROJECT_STRUCTURE.md` for the intended layout.
3. For product behavior, read `docs/1. PRD.md`; for service boundaries read `docs/3. ARCHITECTURE.md`; for SQL read `docs/4. DATABASE.md`; for HTTP contracts read `docs/5. API.md`; for security read `docs/11. Authentication & Security.md`.
4. Treat this file and the numbered documents as project context. If they conflict with existing working code, report the conflict and propose a migration instead of silently breaking the application.

## 3. Protect the existing Flutter UI

The Flutter UI may already exist. Inspect it first. Do not replace `lib/`, rewrite screens, change the state-management library, or run a destructive scaffold command just to match the proposed tree. Preserve working screens and assets. Propose or carry out only a minimal, reversible migration, and keep the app runnable after each step.

If the repository is currently a single Flutter app at the root, adapt the target structure to the actual repository first. Do not move the app to `apps/customer_flutter/` until imports, assets, native platform folders, tests, and build configuration have a migration plan.

## 4. Non-negotiable security rules

- Never commit `.env`, production credentials, database URLs/passwords, signing keys, service-role keys, OTPs, or real user data. Commit `.env.example` with placeholders only.
- Never place database credentials, Supabase secret/service-role keys, or JWT signing keys in Flutter.
- Flutter calls Aupa's Go API. Do not make the client depend on privileged direct PostgreSQL access.
- A custom Aupa JWT is **not automatically a Supabase Auth identity**. Never assume it populates Supabase `auth.uid()` without an explicitly designed and tested supported JWT trust configuration.
- Validate JWT signature, algorithm allow-list, issuer, audience, expiry and key ID. Use short-lived access tokens and revocable/rotated refresh sessions. Store tokens only with platform-backed secure storage; never log them.
- Hash passwords with a suitable password-hashing algorithm (prefer Argon2id where supported and properly configured); never encrypt passwords reversibly or store plain text.
- Enforce authentication, role permissions and object-level ownership in backend use cases. A customer must not access another customer's addresses/orders; a merchant must not mutate another merchant's catalog/orders; a rider must not access unassigned deliveries.
- Use parameterized SQL, server-side price calculations, database constraints, bounded request sizes/timeouts, rate limits and safe error responses.
- Never trust client-supplied price, total, role, payment status, delivery status or ownership.
- Payment success is established by a verified provider webhook or server-to-server reconciliation, not a mobile redirect.
- Use idempotency and validated state transitions for checkout, payment operations, provider webhooks and delivery assignment.
- Do not run destructive production SQL, access production secrets, deploy, issue real OTPs, refund money, or trigger real payment actions without explicit human approval.

## 5. Architecture boundaries

- **Go API**: public REST API, custom auth, users, restaurants/menus, cart validation, checkout/order orchestration, payment-provider adapters, webhook verification, notifications, role checks and audit trail.
- **Rust worker (later phase)**: private dispatch scoring / assignment proposals only. Go validates every proposal and performs authoritative state transitions. Rust must not own authentication, pricing, payment state or arbitrary order transitions.
- **PostgreSQL**: single source of truth. Keep schema changes in `supabase/migrations/`; make migrations reviewable and preferably backward compatible.
- **Events**: use a transactional outbox with retries and idempotency before adding an external message broker. Do not add Redis/Kafka or extra services without a measured need.
- Keep transactions short. Never hold a DB transaction open while calling a payment, SMS, email, map or push provider.

## 6. Implementation style

- Make the smallest coherent change that satisfies the current task. Do not build the entire future platform in one pass.
- Follow existing patterns and dependencies before introducing a new library. Do not change Flutter state management or routing without explaining why.
- Keep handlers/controllers thin; put business policy in explicit service/use-case/domain code.
- Use UTC timestamps internally. Use integer minor units for money (e.g. poisha for BDT) and never floating-point for amounts. Keep currency configurable if markets may change.
- Keep order, payment and delivery statuses separate. Enforce allowed transitions and write audit/event history for important transitions.
- Update API docs, migrations, tests and engineering docs whenever contracts or behavior change.
- Add loading, empty, error and retry states in UI work; preserve accessibility and localization.
- Do not create many empty placeholder files. Create a directory when its first real file or a clear deliverable is implemented.

## 7. Required verification

Run the checks relevant to the changed code and report the exact results. Do not claim tests passed if they were not run.

- Flutter: `dart format --set-exit-if-changed .`, `flutter analyze`, `flutter test` from the Flutter app directory.
- Go: `gofmt` on changed Go files, `go vet ./...`, `go test ./...` from `services/api-go/`.
- Rust: `cargo fmt --check`, `cargo clippy --all-targets --all-features -- -D warnings`, `cargo test` from `services/dispatch-rust/` once that service exists.
- Database: validate migrations against a disposable/local Supabase database and test constraints, tenant/owner scoping, rollback or forward-fix strategy, and duplicate requests.
- Security: include negative authorization tests, refresh-token revocation tests, idempotency tests and webhook-signature tests where relevant.

## 8. Working sequence

Implement in small, reviewable stages:

1. Audit the repository and preserve the existing UI.
2. Agree the folder layout and establish documented run/test commands.
3. Create local configuration examples and health checks.
4. Add PostgreSQL migrations and safe database access.
5. Implement custom authentication/session lifecycle and authorization.
6. Add restaurant/menu, cart, checkout and order lifecycle.
7. Integrate one selected payment provider in sandbox mode, including webhook verification and reconciliation.
8. Add notifications and delivery tracking.
9. Add the Rust dispatch worker only when the dispatch rules and input/output contract are ready.
10. Harden monitoring, deployment, backups, recovery and release checks.

At the end of each task, summarize changed files, commands/tests run, known gaps, and the next smallest step. Ask before destructive changes or unclear product decisions that affect money, identity, privacy or order lifecycle.
