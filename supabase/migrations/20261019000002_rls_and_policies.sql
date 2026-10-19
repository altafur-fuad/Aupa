-- ============================================================================
-- Aupa Food Delivery — Row Level Security (RLS) & Access Policy Migration
-- Migration: 20261019000002_rls_and_policies.sql
-- Description: Defense-in-depth RLS configuration.
--
-- CRITICAL SECURITY NOTICE:
-- As documented in docs/4. DATABASE.md and AGENTS.md:
-- 1. Flutter clients communicate EXCLUSIVELY with Aupa's Go backend API.
-- 2. Direct mobile access to PostgreSQL/Supabase Data API is NOT permitted.
-- 3. Aupa's Go API issues custom HMAC/Argon2 JWT tokens which do NOT automatically
--    populate Supabase Auth's `auth.uid()`.
-- 4. Authoritative authorization and object-level permissions are strictly enforced
--    in the Go application layer.
-- 5. These RLS policies provide defense-in-depth for direct database queries and
--    support session variables ('app.current_user_id') when scoped connections are used.
-- ============================================================================

-- Helper function to extract user ID from custom session variable or request claim
CREATE OR REPLACE FUNCTION public.current_app_user_id()
RETURNS UUID AS $$
BEGIN
    RETURN COALESCE(
        NULLIF(current_setting('request.jwt.claim.sub', true), ''),
        NULLIF(current_setting('app.current_user_id', true), '')
    )::UUID;
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- ----------------------------------------------------------------------------
-- 1. Enable RLS on All Public Tables (Default Deny)
-- ----------------------------------------------------------------------------
ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.refresh_tokens ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.addresses ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.restaurants ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.menu_categories ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.menu_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.menu_item_options ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.orders ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.order_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.order_status_history ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.payment_attempts ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.webhook_audits ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.idempotency_records ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.outbox_events ENABLE ROW LEVEL SECURITY;

-- ----------------------------------------------------------------------------
-- 2. Public / Anonymous Catalog Read Policies
-- Public read access to active restaurant listings and menus
-- ----------------------------------------------------------------------------
CREATE POLICY "Public read access for restaurants"
    ON public.restaurants FOR SELECT
    USING (true);

CREATE POLICY "Public read access for menu categories"
    ON public.menu_categories FOR SELECT
    USING (is_active = true);

CREATE POLICY "Public read access for menu items"
    ON public.menu_items FOR SELECT
    USING (is_available = true);

CREATE POLICY "Public read access for menu item options"
    ON public.menu_item_options FOR SELECT
    USING (true);

-- ----------------------------------------------------------------------------
-- 3. Customer Data Scoping Policies (Defense in Depth)
-- ----------------------------------------------------------------------------
-- Addresses: Customers can only manage their own addresses
CREATE POLICY "Customer manage own addresses"
    ON public.addresses FOR ALL
    USING (user_id = public.current_app_user_id())
    WITH CHECK (user_id = public.current_app_user_id());

-- Orders: Customers can only view their own orders
CREATE POLICY "Customer view own orders"
    ON public.orders FOR SELECT
    USING (customer_id = public.current_app_user_id());

-- Order Items: Customers can only view items of their own orders
CREATE POLICY "Customer view own order items"
    ON public.order_items FOR SELECT
    USING (
        EXISTS (
            SELECT 1 FROM public.orders
            WHERE orders.id = order_items.order_id
            AND orders.customer_id = public.current_app_user_id()
        )
    );

-- ----------------------------------------------------------------------------
-- 4. Backend Service Role Unrestricted Access
-- Note: Aupa Go API connects via database credentials or service_role
-- ----------------------------------------------------------------------------
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'service_role') THEN
        CREATE POLICY "Service role full access on users" ON public.users FOR ALL TO service_role USING (true) WITH CHECK (true);
        CREATE POLICY "Service role full access on refresh_tokens" ON public.refresh_tokens FOR ALL TO service_role USING (true) WITH CHECK (true);
        CREATE POLICY "Service role full access on addresses" ON public.addresses FOR ALL TO service_role USING (true) WITH CHECK (true);
        CREATE POLICY "Service role full access on restaurants" ON public.restaurants FOR ALL TO service_role USING (true) WITH CHECK (true);
        CREATE POLICY "Service role full access on menu_categories" ON public.menu_categories FOR ALL TO service_role USING (true) WITH CHECK (true);
        CREATE POLICY "Service role full access on menu_items" ON public.menu_items FOR ALL TO service_role USING (true) WITH CHECK (true);
        CREATE POLICY "Service role full access on menu_item_options" ON public.menu_item_options FOR ALL TO service_role USING (true) WITH CHECK (true);
        CREATE POLICY "Service role full access on orders" ON public.orders FOR ALL TO service_role USING (true) WITH CHECK (true);
        CREATE POLICY "Service role full access on order_items" ON public.order_items FOR ALL TO service_role USING (true) WITH CHECK (true);
        CREATE POLICY "Service role full access on order_status_history" ON public.order_status_history FOR ALL TO service_role USING (true) WITH CHECK (true);
        CREATE POLICY "Service role full access on payment_attempts" ON public.payment_attempts FOR ALL TO service_role USING (true) WITH CHECK (true);
        CREATE POLICY "Service role full access on webhook_audits" ON public.webhook_audits FOR ALL TO service_role USING (true) WITH CHECK (true);
        CREATE POLICY "Service role full access on idempotency_records" ON public.idempotency_records FOR ALL TO service_role USING (true) WITH CHECK (true);
        CREATE POLICY "Service role full access on outbox_events" ON public.outbox_events FOR ALL TO service_role USING (true) WITH CHECK (true);
    END IF;
END $$;
