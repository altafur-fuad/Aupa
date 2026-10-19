# Aupa — Project Documentation and Antigravity Pack

This folder contains the documentation and root-level agent instructions for **Aupa**, a food-delivery app using Flutter/Dart, a Go API, Supabase PostgreSQL, and an optional private Rust dispatch worker. Supabase Auth is not part of the chosen plan.

## Start here

1. Put `AGENTS.md` at the root of your actual Aupa repository. Antigravity can discover root `AGENTS.md` files as persistent project instructions. See the official [Antigravity Rules documentation](https://www.antigravity.google/docs/rules/).
2. Copy the numbered files inside `docs/` into your project's `docs/` folder.
3. Read `docs/PROJECT_STRUCTURE.md` for the proposed folder tree and what each folder is for.
4. Open the Aupa project in Antigravity and use `docs/ANTIGRAVITY_BOOTSTRAP_PROMPT.md` as your first task prompt.
5. Let Antigravity audit the existing repository before it moves or replaces anything. Since the UI is already partly done, preserve it and migrate gradually.

## Main architecture

- `apps/customer_flutter/`: customer-facing Flutter app (adapt to your current location rather than moving blindly).
- `services/api-go/`: public API, custom auth, business rules, checkout/order/payment orchestration.
- `services/dispatch-rust/`: optional private dispatch scoring worker; defer until there is a defined need.
- `supabase/migrations/`: PostgreSQL schema history.
- `api/openapi/openapi.yaml`: API contract.
- `docs/`: product, design, architecture, database, API, security, testing and implementation docs.

**Security note:** Aupa's own JWT is not automatically a Supabase Auth session. In the initial plan Flutter talks only to Go; the server authenticates requests and enforces authorization. Do not put database credentials or Supabase service-role keys into the app.

This pack includes planning documents and agent instructions; it is not a completed or deployed application. Confirm provider choices and test environments before implementing payments, OTP delivery or production deployment.
