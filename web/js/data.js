// ============================================================
// AUPA INITIAL DATASET & MOCK CATALOG
// Fallback & seed data for restaurants, menus, and orders
// ============================================================

window.AupaData = {
  categories: [
    { id: 'all', name: 'All Cuisines', icon: '🍽️' },
    { id: 'biryani', name: 'Biryani & Kacchi', icon: '🥘' },
    { id: 'burgers', name: 'Burgers & Fast Food', icon: '🍔' },
    { id: 'pizza', name: 'Pizza & Pasta', icon: '🍕' },
    { id: 'bengali', name: 'Traditional Bengali', icon: '🍛' },
    { id: 'asian', name: 'Asian & Sushi', icon: '🍱' },
    { id: 'bakery', name: 'Cafe & Desserts', icon: '🍰' },
  ],

  restaurants: [
    {
      id: 'rest-1',
      name: 'Kacchi Bhai - Banani',
      category: 'biryani',
      rating: 4.8,
      reviewsCount: 1420,
      deliveryTime: '25-35 min',
      minOrder: 200,
      deliveryFee: 40,
      isFreeDelivery: false,
      isOpen: true,
      image: 'https://images.unsplash.com/photo-1563379091339-03b21ab4a4f8?w=700&auto=format&fit=crop&q=80',
      cuisines: ['Kacchi Biryani', 'Mutton', 'Bengali Delight'],
      badge: 'Popular Choice',
      menu: [
        { id: 'm1', name: 'Basmati Mutton Kacchi (Full)', price: 420, desc: 'Aromatic basmati rice cooked with tender mutton chunks and traditional spices, served with borhani and alu bukhara chutney.', tag: 'Bestseller' },
        { id: 'm2', name: 'Basmati Chicken Kacchi', price: 340, desc: 'Fragrant basmati rice layered with marinated chicken roast and golden potato.', tag: 'Popular' },
        { id: 'm3', name: 'Spicy Shahi Borhani (250ml)', price: 60, desc: 'Traditional rich yogurt beverage infused with mint, coriander, and black salt.', tag: 'Beverage' },
        { id: 'm4', name: 'Shahi Firni', price: 80, desc: 'Silky ground rice pudding sweetened with condensed milk and garnished with pistachios.', tag: 'Dessert' }
      ]
    },
    {
      id: 'rest-2',
      name: 'Takeout Burgers - Dhanmondi',
      category: 'burgers',
      rating: 4.7,
      reviewsCount: 980,
      deliveryTime: '20-30 min',
      minOrder: 150,
      deliveryFee: 30,
      isFreeDelivery: true,
      isOpen: true,
      image: 'https://images.unsplash.com/photo-1568901346375-23c9450c58cd?w=700&auto=format&fit=crop&q=80',
      cuisines: ['Burgers', 'Crispy Fries', 'Milkshakes'],
      badge: 'Free Delivery',
      menu: [
        { id: 'm5', name: 'Beef Double Gourmet Burger', price: 380, desc: 'Two char-grilled 100g beef patties with double cheddar, caramelized onions, and house special BBQ sauce.', tag: 'Chef Choice' },
        { id: 'm6', name: 'Crispy Zesty Chicken Supreme', price: 310, desc: 'Buttermilk fried chicken breast fillet topped with spicy mayo and crisp lettuce.', tag: 'Crunchy' },
        { id: 'm7', name: 'Loaded Cheesy Beef Fries', price: 220, desc: 'Crinkle-cut french fries drizzled with melted mozzarella, minced beef, and jalapeños.', tag: 'Snack' }
      ]
    },
    {
      id: 'rest-3',
      name: 'Pizza Guy - Gulshan 1',
      category: 'pizza',
      rating: 4.9,
      reviewsCount: 1650,
      deliveryTime: '30-40 min',
      minOrder: 350,
      deliveryFee: 50,
      isFreeDelivery: false,
      isOpen: true,
      image: 'https://images.unsplash.com/photo-1513104890138-7c749659a591?w=700&auto=format&fit=crop&q=80',
      cuisines: ['Artisan Pizza', 'Italian', 'Wood-Fired'],
      badge: 'Top Rated',
      menu: [
        { id: 'm8', name: 'Classic Pepperoni Overload (12")', price: 890, desc: 'Italian marinara base loaded with premium beef pepperoni and stretchy mozzarella cheese.', tag: 'Legendary' },
        { id: 'm9', name: 'Smoked BBQ Chicken Pizza (12")', price: 820, desc: 'Smoked pulled chicken, red bell peppers, sweet corn, and chipotle BBQ drizzle.', tag: 'Bestseller' },
        { id: 'm10', name: 'Creamy Garlic Mushroom Fettuccine', price: 460, desc: 'Fresh pasta tossed in velvety parmesan cream with sautéed shiitake and button mushrooms.', tag: 'Pasta' }
      ]
    },
    {
      id: 'rest-4',
      name: 'Tokyo Tokyo Sushi & Ramen',
      category: 'asian',
      rating: 4.6,
      reviewsCount: 520,
      deliveryTime: '35-45 min',
      minOrder: 300,
      deliveryFee: 50,
      isFreeDelivery: false,
      isOpen: true,
      image: 'https://images.unsplash.com/photo-1579871494447-9811cf80d66c?w=700&auto=format&fit=crop&q=80',
      cuisines: ['Japanese', 'Sushi', 'Ramen Noodles'],
      badge: 'Trending',
      menu: [
        { id: 'm11', name: 'Crispy Tempura Prawn Roll (8 Pcs)', price: 650, desc: 'Tiger prawn tempura, avocado, and cucumber rolled in nori and toasted sesame.', tag: 'Sushi' },
        { id: 'm12', name: 'Spicy Chicken Miso Ramen', price: 540, desc: 'Rich chicken bone broth with homemade springy noodles, ajitsuke tamago egg, and nori.', tag: 'Hot Bowl' }
      ]
    },
    {
      id: 'rest-5',
      name: 'Star Kabab & Restaurant',
      category: 'bengali',
      rating: 4.7,
      reviewsCount: 3100,
      deliveryTime: '20-30 min',
      minOrder: 100,
      deliveryFee: 30,
      isFreeDelivery: false,
      isOpen: true,
      image: 'https://images.unsplash.com/photo-1603894584373-5ac82b2ae398?w=700&auto=format&fit=crop&q=80',
      cuisines: ['Mutton Chaap', 'Grilled Chicken', 'Luchi & Naan'],
      badge: 'Historic Favorite',
      menu: [
        { id: 'm13', name: 'Special Mutton Chaap with Butter', price: 230, desc: 'Tender marinated mutton cutlet pan-fried in pure ghee with aromatic spices.', tag: 'Must Try' },
        { id: 'm14', name: 'Garlic Butter Naan', price: 60, desc: 'Fluffy tandoor-baked flatbread brushed with garlic-infused butter.', tag: 'Bread' },
        { id: 'm15', name: 'Signature Star Faluda', price: 160, desc: 'Layered vermicelli dessert with sweet milk, jelly cubes, basil seeds, and ice cream scoop.', tag: 'Sweet' }
      ]
    },
    {
      id: 'rest-6',
      name: 'Glazed & Baked Patisserie',
      category: 'bakery',
      rating: 4.8,
      reviewsCount: 840,
      deliveryTime: '15-25 min',
      minOrder: 150,
      deliveryFee: 35,
      isFreeDelivery: true,
      isOpen: true,
      image: 'https://images.unsplash.com/photo-1509440159596-0249088772ff?w=700&auto=format&fit=crop&q=80',
      cuisines: ['Artisan Bakery', 'Croissants', 'Cheesecakes'],
      badge: 'Free Delivery',
      menu: [
        { id: 'm16', name: 'Belgian Chocolate Truffle Cake (Slice)', price: 290, desc: 'Moist dark cocoa sponge layered with 70% Belgian chocolate ganache.', tag: 'Decadent' },
        { id: 'm17', name: 'Butter Croissant (2 Pcs)', price: 210, desc: 'Golden flaky French croissants made with cultured European butter.', tag: 'Fresh Baked' }
      ]
    }
  ],

  // Live order state for Admin / Restaurant Portal
  partnerOrders: [
    {
      id: 'AUP-8921',
      customerName: 'Tanvir Ahmed',
      customerPhone: '01712-889900',
      address: 'House 14, Road 7, Banani, Dhaka',
      items: [
        { name: 'Basmati Mutton Kacchi (Full)', qty: 2, price: 420 },
        { name: 'Spicy Shahi Borhani (250ml)', qty: 2, price: 60 }
      ],
      total: 960,
      paymentMethod: 'Cash on Delivery',
      status: 'pending',
      timestamp: 'Just now'
    },
    {
      id: 'AUP-8920',
      customerName: 'Sabrina Rahman',
      customerPhone: '01844-332211',
      address: 'Apt 4B, Concord Tower, Gulshan 2',
      items: [
        { name: 'Beef Double Gourmet Burger', qty: 1, price: 380 },
        { name: 'Loaded Cheesy Beef Fries', qty: 1, price: 220 }
      ],
      total: 650,
      paymentMethod: 'bKash Online',
      status: 'preparing',
      timestamp: '12 mins ago'
    }
  ]
};
