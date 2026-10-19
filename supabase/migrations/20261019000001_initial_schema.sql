-- ============================================================================
-- Aupa Food Delivery — Initial PostgreSQL Schema Migration
-- Migration: 20261019000001_initial_schema.sql
-- Description: Core tables, money poisha integrity, state check constraints,
--              transactional outbox, idempotency, and foreign key relations.
-- ============================================================================

-- Ensure pgcrypto extension is enabled for UUID generation
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ----------------------------------------------------------------------------
-- 1. Identity & Custom Authentication
-- Note: Aupa Go API manages custom authentication; Supabase Auth is NOT used.
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS public.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL CHECK (length(name) >= 2 AND length(name) <= 120),
    email TEXT,
    phone TEXT,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'customer' CHECK (role IN ('customer', 'merchant', 'rider', 'admin')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    email_verified BOOLEAN NOT NULL DEFAULT false,
    phone_verified BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now()),
    CONSTRAINT chk_email_or_phone CHECK (email IS NOT NULL OR phone IS NOT NULL)
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_users_lower_email
    ON public.users (lower(email)) WHERE email IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_users_phone
    ON public.users (phone) WHERE phone IS NOT NULL;

-- Auth Sessions / Refresh Tokens with Replay Attack Family Tracking
CREATE TABLE IF NOT EXISTS public.refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL UNIQUE,
    family_id UUID NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now())
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_family
    ON public.refresh_tokens (family_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_active
    ON public.refresh_tokens (user_id, expires_at) WHERE revoked = false;

-- ----------------------------------------------------------------------------
-- 2. Customer Saved Addresses
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS public.addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    label TEXT NOT NULL CHECK (length(label) >= 1 AND length(label) <= 50),
    street TEXT NOT NULL,
    area TEXT NOT NULL,
    city TEXT NOT NULL DEFAULT 'Dhaka',
    delivery_instructions TEXT,
    latitude DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    longitude DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now())
);

CREATE INDEX IF NOT EXISTS idx_addresses_user
    ON public.addresses (user_id);

-- ----------------------------------------------------------------------------
-- 3. Restaurant & Menu Catalog
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS public.restaurants (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    slug TEXT NOT NULL UNIQUE,
    description TEXT,
    cuisine TEXT NOT NULL,
    image_url TEXT,
    rating DOUBLE PRECISION NOT NULL DEFAULT 0.0 CHECK (rating >= 0.0 AND rating <= 5.0),
    review_count INT NOT NULL DEFAULT 0 CHECK (review_count >= 0),
    delivery_eta TEXT NOT NULL,
    minimum_order_poisha BIGINT NOT NULL DEFAULT 0 CHECK (minimum_order_poisha >= 0),
    is_open BOOLEAN NOT NULL DEFAULT true,
    latitude DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    longitude DOUBLE PRECISION NOT NULL DEFAULT 0.0,
    zone_id TEXT NOT NULL DEFAULT 'dhaka_central',
    created_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now())
);

CREATE INDEX IF NOT EXISTS idx_restaurants_zone_open
    ON public.restaurants (zone_id, is_open);

CREATE TABLE IF NOT EXISTS public.menu_categories (
    id TEXT PRIMARY KEY,
    restaurant_id TEXT NOT NULL REFERENCES public.restaurants(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    display_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now())
);

CREATE INDEX IF NOT EXISTS idx_menu_categories_rest
    ON public.menu_categories (restaurant_id, display_order);

CREATE TABLE IF NOT EXISTS public.menu_items (
    id TEXT PRIMARY KEY,
    restaurant_id TEXT NOT NULL REFERENCES public.restaurants(id) ON DELETE CASCADE,
    category_id TEXT NOT NULL REFERENCES public.menu_categories(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    price_poisha BIGINT NOT NULL CHECK (price_poisha >= 0),
    image_url TEXT,
    is_available BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now())
);

CREATE INDEX IF NOT EXISTS idx_menu_items_rest_cat
    ON public.menu_items (restaurant_id, category_id, is_available);

CREATE TABLE IF NOT EXISTS public.menu_item_options (
    id TEXT PRIMARY KEY,
    item_id TEXT NOT NULL REFERENCES public.menu_items(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    price_delta_poisha BIGINT NOT NULL DEFAULT 0 CHECK (price_delta_poisha >= 0),
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now())
);

CREATE INDEX IF NOT EXISTS idx_menu_item_options_item
    ON public.menu_item_options (item_id);

-- ----------------------------------------------------------------------------
-- 4. Commerce & Order Lifecycle
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS public.orders (
    id TEXT PRIMARY KEY,
    customer_id UUID NOT NULL REFERENCES public.users(id) ON DELETE RESTRICT,
    restaurant_id TEXT NOT NULL REFERENCES public.restaurants(id) ON DELETE RESTRICT,
    delivery_address_id UUID REFERENCES public.addresses(id) ON DELETE SET NULL,
    delivery_address_snapshot JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL CHECK (status IN ('placed', 'accepted', 'preparing', 'ready', 'out_for_delivery', 'delivered', 'cancelled')),
    payment_status TEXT NOT NULL CHECK (payment_status IN ('unpaid', 'pending', 'paid', 'failed', 'refunded')),
    payment_method TEXT NOT NULL CHECK (payment_method IN ('cash_on_delivery', 'bkash', 'sslcommerz')),
    subtotal_poisha BIGINT NOT NULL CHECK (subtotal_poisha >= 0),
    delivery_fee_poisha BIGINT NOT NULL CHECK (delivery_fee_poisha >= 0),
    discount_poisha BIGINT NOT NULL DEFAULT 0 CHECK (discount_poisha >= 0),
    vat_poisha BIGINT NOT NULL CHECK (vat_poisha >= 0),
    total_amount_poisha BIGINT NOT NULL CHECK (total_amount_poisha >= 0),
    currency TEXT NOT NULL DEFAULT 'BDT',
    special_instructions TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now())
);

CREATE INDEX IF NOT EXISTS idx_orders_customer_created
    ON public.orders (customer_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_restaurant_status
    ON public.orders (restaurant_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_orders_status_created
    ON public.orders (status, created_at DESC);

CREATE TABLE IF NOT EXISTS public.order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id TEXT NOT NULL REFERENCES public.orders(id) ON DELETE CASCADE,
    item_id TEXT NOT NULL,
    name TEXT NOT NULL,
    unit_price_poisha BIGINT NOT NULL CHECK (unit_price_poisha >= 0),
    quantity INT NOT NULL CHECK (quantity > 0),
    total_poisha BIGINT NOT NULL CHECK (total_poisha >= 0),
    selected_options_snapshot JSONB DEFAULT '[]'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_order_items_order
    ON public.order_items (order_id);

CREATE TABLE IF NOT EXISTS public.order_status_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id TEXT NOT NULL REFERENCES public.orders(id) ON DELETE CASCADE,
    from_status TEXT,
    to_status TEXT NOT NULL,
    reason TEXT,
    actor_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now())
);

CREATE INDEX IF NOT EXISTS idx_order_status_history_order
    ON public.order_status_history (order_id, created_at ASC);

-- ----------------------------------------------------------------------------
-- 5. Payments & Webhook Auditing
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS public.payment_attempts (
    id TEXT PRIMARY KEY,
    order_id TEXT NOT NULL REFERENCES public.orders(id) ON DELETE CASCADE,
    customer_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    provider TEXT NOT NULL,
    amount_poisha BIGINT NOT NULL CHECK (amount_poisha >= 0),
    currency TEXT NOT NULL DEFAULT 'BDT',
    status TEXT NOT NULL CHECK (status IN ('pending', 'succeeded', 'failed', 'cancelled')),
    transaction_id TEXT,
    redirect_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now()),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now())
);

CREATE INDEX IF NOT EXISTS idx_payment_attempts_order
    ON public.payment_attempts (order_id, created_at DESC);

CREATE TABLE IF NOT EXISTS public.webhook_audits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider TEXT NOT NULL,
    event_id TEXT NOT NULL,
    order_id TEXT,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now()),
    CONSTRAINT uq_provider_event UNIQUE (provider, event_id)
);

CREATE INDEX IF NOT EXISTS idx_webhook_audits_order
    ON public.webhook_audits (order_id);

-- ----------------------------------------------------------------------------
-- 6. Orchestration: Idempotency & Transactional Outbox
-- ----------------------------------------------------------------------------
CREATE TABLE IF NOT EXISTS public.idempotency_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key TEXT NOT NULL,
    customer_id UUID NOT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    request_hash TEXT NOT NULL,
    order_id TEXT REFERENCES public.orders(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now()),
    CONSTRAINT uq_idempotency_key_customer UNIQUE (key, customer_id)
);

CREATE TABLE IF NOT EXISTS public.outbox_events (
    id TEXT PRIMARY KEY,
    aggregate_type TEXT NOT NULL,
    aggregate_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    retry_count INT NOT NULL DEFAULT 0,
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now())
);

CREATE INDEX IF NOT EXISTS idx_outbox_events_unprocessed
    ON public.outbox_events (created_at ASC) WHERE processed_at IS NULL;
