-- ============================================================================
-- Aupa Food Delivery — Development Seed Data
-- File: supabase/seed.sql
-- Description: Deterministic development fixture data for local/testing database.
-- NOTICE: Contains zero real customer data or production secrets.
-- Default test password for seeded users is: Password123!
-- ============================================================================

-- 1. Development Users (bcrypt hash for "Password123!")
INSERT INTO public.users (id, name, email, phone, password_hash, role, is_active, email_verified)
VALUES
    ('00000000-0000-0000-0000-000000000001', 'Tanvir Ahmed', 'tanvir@example.com', '+8801711000001', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'customer', true, true),
    ('00000000-0000-0000-0000-000000000002', 'Sultan Merchant', 'merchant@sultansdine.com', '+8801711000002', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'merchant', true, true),
    ('00000000-0000-0000-0000-000000000003', 'Rafiqul Islam', 'rider@aupa.dev', '+8801711000003', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'rider', true, true),
    ('00000000-0000-0000-0000-000000000004', 'System Admin', 'admin@aupa.dev', '+8801711000004', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin', true, true)
ON CONFLICT (id) DO NOTHING;

-- 2. Customer Saved Addresses
INSERT INTO public.addresses (id, user_id, label, street, area, city, delivery_instructions, latitude, longitude, is_default)
VALUES
    ('10000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000001', 'Home', 'House 42, Road 11', 'Banani', 'Dhaka', 'Ring calling bell twice, 4th floor A4', 23.7937, 90.4066, true),
    ('10000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'Office', 'Crystal Palace, Level 8, Gulshan 1', 'Gulshan', 'Dhaka', 'Leave at reception with security', 23.7781, 90.4162, false)
ON CONFLICT (id) DO NOTHING;

-- 3. Restaurants
INSERT INTO public.restaurants (id, name, slug, description, cuisine, image_url, rating, review_count, delivery_eta, minimum_order_poisha, is_open, latitude, longitude, zone_id)
VALUES
    ('rest_01', 'Sultan''s Dine', 'sultans-dine', 'Authentic traditional Kacchi Biryani cooked with pure ghee and fragrant spices.', 'Biryani', 'https://images.unsplash.com/photo-1563379091339-03b21ab4a4f8?w=800', 4.9, 1250, '25-35 min', 20000, true, 23.7937, 90.4066, 'dhaka_banani'),
    ('rest_02', 'Takeout Burgers', 'takeout-burgers', 'Juicy smashed beef patties, house-made brioche buns and signature secret sauce.', 'Burgers', 'https://images.unsplash.com/photo-1568901346375-23c9450c58cd?w=800', 4.7, 890, '20-30 min', 15000, true, 23.7942, 90.4055, 'dhaka_banani'),
    ('rest_03', 'Star Kabab & Restaurant', 'star-kabab', 'Iconic Dhaka eatery renowned for tender mutton kacchi, chicken tikka and naan.', 'Bangladeshi', 'https://images.unsplash.com/photo-1603894584373-5ac82b2ae398?w=800', 4.8, 2100, '30-40 min', 10000, true, 23.7540, 90.3780, 'dhaka_central'),
    ('rest_04', 'Pizza Guy', 'pizza-guy', 'Artisanal Neapolitan wood-fired pizzas with fresh mozzarella and spicy beef pepperoni.', 'Pizza', 'https://images.unsplash.com/photo-1513104890138-7c749659a591?w=800', 4.6, 640, '25-35 min', 25000, true, 23.7950, 90.4070, 'dhaka_banani')
ON CONFLICT (id) DO NOTHING;

-- 4. Menu Categories
INSERT INTO public.menu_categories (id, restaurant_id, name, display_order, is_active)
VALUES
    ('cat_01', 'rest_01', 'Kacchi Platters', 1, true),
    ('cat_02', 'rest_01', 'Sides & Drinks', 2, true),
    ('cat_03', 'rest_02', 'Burgers', 1, true),
    ('cat_04', 'rest_02', 'Fries & Shakes', 2, true),
    ('cat_05', 'rest_03', 'Kabab & Naan', 1, true),
    ('cat_06', 'rest_04', 'Classic Pizzas', 1, true)
ON CONFLICT (id) DO NOTHING;

-- 5. Menu Items (Prices in Poisha: 100 Poisha = 1 BDT)
INSERT INTO public.menu_items (id, restaurant_id, category_id, name, description, price_poisha, image_url, is_available)
VALUES
    ('item_01', 'rest_01', 'cat_01', 'Kacchi Biryani (Bashmati)', 'Tender marinated mutton with long grain fragrant bashmati rice, spiced aloo and egg.', 48000, 'https://images.unsplash.com/photo-1563379091339-03b21ab4a4f8?w=800', true),
    ('item_02', 'rest_01', 'cat_01', 'Kacchi Biryani with Borhani & Jali Kebab', 'Complete feast featuring mutton kacchi, fresh borhani and traditional jali kebab.', 62000, 'https://images.unsplash.com/photo-1563379091339-03b21ab4a4f8?w=800', true),
    ('item_03', 'rest_01', 'cat_02', 'Special Borhani (250ml)', 'Refreshing sweet and spicy yogurt drink seasoned with mustard, black salt and mint.', 8000, 'https://images.unsplash.com/photo-1556881286-fc6915169721?w=800', true),
    ('item_04', 'rest_01', 'cat_02', 'Jali Kebab (1 pc)', 'Spiced minced beef patty dipped in beaten egg and pan fried until golden brown.', 6000, 'https://images.unsplash.com/photo-1603894584373-5ac82b2ae398?w=800', true),
    ('item_05', 'rest_02', 'cat_03', 'Beef Cheese Delight', 'Single prime beef patty, American cheddar slice, caramelized onions and spicy mayo.', 32000, 'https://images.unsplash.com/photo-1568901346375-23c9450c58cd?w=800', true),
    ('item_06', 'rest_02', 'cat_03', 'Crispy Chicken Burger', 'Crispy fried chicken breast fillet, pickled jalapenos, crunchy slaw and house sauce.', 29000, 'https://images.unsplash.com/photo-1568901346375-23c9450c58cd?w=800', true),
    ('item_07', 'rest_04', 'cat_06', 'Beef Pepperoni Pizza (12")', 'Wood-fired crust topped with rich tomato sauce, melted mozzarella and halal beef pepperoni.', 85000, 'https://images.unsplash.com/photo-1513104890138-7c749659a591?w=800', true)
ON CONFLICT (id) DO NOTHING;

-- 6. Menu Item Customization Options (Price delta in Poisha)
INSERT INTO public.menu_item_options (id, item_id, name, price_delta_poisha, is_default)
VALUES
    ('opt_01', 'item_01', 'Extra Spiced Aloo (1 pc)', 3000, false),
    ('opt_02', 'item_01', 'Extra Mutton Piece', 18000, false),
    ('opt_03', 'item_05', 'Add Extra Cheddar Slice', 5000, false),
    ('opt_04', 'item_05', 'Double Beef Patty', 14000, false),
    ('opt_05', 'item_07', 'Extra Mozzarella Cheese', 12000, false),
    ('opt_06', 'item_07', 'Thin Crust', 0, true)
ON CONFLICT (id) DO NOTHING;
