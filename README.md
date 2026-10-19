# Aupa — Food Delivery Platform

Aupa is a modern food-delivery application for Bangladesh, built with a high-performance stack:
- **Mobile Client:** Flutter + Dart (`aupa/`) with custom UI, dynamic cart/quote calculation in poisha (integer minor units; 100 poisha = ৳1 BDT), and clean switch between Mock and Live API clients.
- **Backend API:** Go REST API (`services/api-go/`) under `/v1`, featuring custom authentication with short-lived access JWTs and rotating refresh tokens with replay attack detection, server-authoritative checkout calculation, idempotent order placement, pluggable payment provider abstractions, and HMAC-SHA256 verified webhook reconciliation.
- **Database:** Supabase-hosted PostgreSQL with versioned SQL migrations (`supabase/migrations/`), defense-in-depth Row Level Security (RLS) policies, and realistic development seed data (`supabase/seed.sql`).
- **Dispatch Worker:** Scoped Rust delivery-dispatch worker formally specified in `docs/RUST_DISPATCH_WORKER_SPEC.md` (deferred per project rules until dedicated optimization algorithm is needed).

---

## 1. Quick Start

### Frontend (Flutter Client)
```bash
cd aupa
flutter pub get
flutter run
```
To run tests and static analysis:
```bash
dart format --set-exit-if-changed .
flutter analyze
flutter test
```

### Backend (Go API)
```bash
cd services/api-go
go run ./cmd/api
```
To run test suites and verification:
```bash
go fmt ./...
go vet ./...
go test -count=1 ./...
go build ./cmd/api
```

---

## 2. Database Operations (Supabase PostgreSQL)

### Run Migrations
```bash
# Using Supabase CLI:
supabase db push

# Or using psql directly:
psql "$DATABASE_URL" -f supabase/migrations/20261019000001_initial_schema.sql
psql "$DATABASE_URL" -f supabase/migrations/20261019000002_rls_and_policies.sql
```

### Apply Development Seed Data
```bash
psql "$DATABASE_URL" -f supabase/seed.sql
```

### Backup and Restore
```bash
# Backup database (custom format archive):
pg_dump -Fc -v --no-owner --no-acl "$DATABASE_URL" > backup_aupa_$(date +%Y%m%d).dump

# Restore database:
pg_restore -v --clean --no-owner --no-acl -d "$DATABASE_URL" backup_aupa.dump
```

---

## 3. Security Principles

- **No Supabase Auth:** Authentication is implemented exclusively by Aupa's Go backend.
- **Strict Money Representation:** All monetary values are integer minor units (poisha). Floating-point arithmetic for money is strictly prohibited.
- **Secrets Isolation:** `.env` files and credentials are strictly ignored in Git. Only `.env.example` with safe placeholder values is committed.
- **Object-Level Authorization:** Access control is strictly enforced on the server. Customers cannot view, modify, pay, or track orders owned by other users.

