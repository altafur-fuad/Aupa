# Antigravity Bootstrap Prompt — Aupa

Copy the prompt below into Google Antigravity after opening the Aupa repository as a Project/workspace. Root `AGENTS.md` is the always-on instruction file; this prompt tells the agent how to apply it safely to the current repository.

---

You are working on **Aupa**, a food-delivery application. First read `AGENTS.md`, `docs/PROJECT_STRUCTURE.md`, `docs/1. PRD.md`, `docs/3. ARCHITECTURE.md`, `docs/4. DATABASE.md`, `docs/5. API.md`, and `docs/11. Authentication & Security.md`.

## Objective

Audit the current repository and create a practical, phased implementation plan and repository layout based on those documents. The Flutter UI is already partially complete, so preserving existing screens, assets, routing and state management is a priority.

## Required steps

1. Inspect the full current tree, the Flutter entry point, `pubspec.yaml`, app routing/state management, tests, and `git status`. Do not modify files yet.
2. Report the existing project structure, what is already implemented, and the differences between current and target tree.
3. Propose a minimal migration plan. If the Flutter project currently lives at repository root, do not move it until you explain the required path/import/asset/platform changes. Never overwrite working UI to match a template.
4. Once the plan is clear, establish only the necessary foundation: root `AGENTS.md`, README, `.gitignore`, `.env.example`, API contract directory, Go API skeleton and health endpoint, and Supabase migration directory. Do not create hundreds of empty files.
5. Keep the backend architecture as follows: Go is the public API and custom-auth authority; Supabase is PostgreSQL only (do not use Supabase Auth); Rust is a private dispatch-scoring worker to add only when its interface is defined. Flutter calls the Go API, not privileged database credentials.
6. Do not implement payment-provider calls, SMS/OTP delivery, production migrations, or deploy anything until the provider/environment is selected and human approval is given. Use sandbox/local configuration only.
7. Add tests and documented run commands for each real slice. Run the available formatters, analyzers and tests; report any that could not run.
8. Finish with a file-by-file change summary, commands and actual results, unresolved decisions, security notes, and the next smallest step.

## Hard constraints

- Preserve the existing UI and assets.
- Never commit or expose real secrets.
- Never assume Aupa's custom JWT automatically populates Supabase `auth.uid()`.
- Never trust client-provided prices, roles, order/payment status or object ownership.
- Keep payment, order and delivery states separate; use idempotency and verified webhooks for payment outcomes.
- Do not perform destructive changes without explicit approval.

Begin with the audit and report. Do not start by generating all future files at once.
