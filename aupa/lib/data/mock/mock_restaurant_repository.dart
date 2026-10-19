import '../../core/errors/failures.dart';
import '../../core/network/result.dart';
import '../../domain/models/menu_item.dart';
import '../../domain/models/restaurant.dart';
import '../../domain/repositories/restaurant_repository.dart';

class MockRestaurantRepository implements RestaurantRepository {
  final List<Restaurant> _restaurants = const [
    Restaurant(
      id: 'rest_01',
      name: 'Burger Barn',
      cuisine: 'Burgers & Fast Food',
      area: 'Gulshan 2',
      address: 'House 14, Road 113, Gulshan 2, Dhaka',
      rating: 4.8,
      reviewCount: 2150,
      deliveryTimeMinutes: 25,
      deliveryFeePoisha: 6000, // ৳60
      isOpen: true,
    ),
    Restaurant(
      id: 'rest_02',
      name: "Sultan's Dine",
      cuisine: 'Biryani & Kacchi',
      area: 'Banani',
      address: 'Navana Tower, Road 11, Banani, Dhaka',
      rating: 4.9,
      reviewCount: 4890,
      deliveryTimeMinutes: 30,
      deliveryFeePoisha: 5000, // ৳50
      isOpen: true,
    ),
    Restaurant(
      id: 'rest_03',
      name: 'Takeout',
      cuisine: 'American & Gourmet Burgers',
      area: 'Dhanmondi',
      address: 'Concord Royal Court, Dhanmondi 27, Dhaka',
      rating: 4.7,
      reviewCount: 3120,
      deliveryTimeMinutes: 20,
      deliveryFeePoisha: 6000, // ৳60
      isOpen: true,
    ),
    Restaurant(
      id: 'rest_04',
      name: 'Pizza Guy',
      cuisine: 'Artisan Pizza & Italian',
      area: 'Uttara',
      address: 'Plot 28, Sector 3, Uttara, Dhaka',
      rating: 4.8,
      reviewCount: 1680,
      deliveryTimeMinutes: 35,
      deliveryFeePoisha: 8000, // ৳80
      isOpen: true,
    ),
    Restaurant(
      id: 'rest_05',
      name: 'Kacchi Bhai',
      cuisine: 'Traditional Bengali & Mughlai',
      area: 'Mirpur',
      address: 'Mirpur 10 Roundabout, Dhaka',
      rating: 4.6,
      reviewCount: 2450,
      deliveryTimeMinutes: 40,
      deliveryFeePoisha: 7000, // ৳70
      isOpen: false, // currently closed to test closed status
    ),
  ];

  final Map<String, List<MenuItem>> _menus = {
    'rest_01': const [
      MenuItem(
        id: 'menu_01',
        restaurantId: 'rest_01',
        name: 'Spicy BBQ Burger',
        description:
            'Juicy grilled beef patty, melted cheddar cheese, caramelized onions, jalapeños, and signature smoky BBQ sauce on a toasted brioche bun.',
        pricePoisha: 39000, // ৳390
        category: 'Burgers',
        isAvailable: true,
        options: [
          MenuItemOption(
            id: 'opt_spicy',
            name: 'Spice Level',
            isRequired: true,
            allowsMultiple: false,
            values: [
              OptionValue(id: 'v_mild', name: 'Mild', isDefault: true),
              OptionValue(id: 'v_med', name: 'Medium'),
              OptionValue(
                id: 'v_hot',
                name: 'Extra Hot (+৳20)',
                additionalPricePoisha: 2000,
              ),
            ],
          ),
          MenuItemOption(
            id: 'opt_addons',
            name: 'Custom Add-ons',
            isRequired: false,
            allowsMultiple: true,
            values: [
              OptionValue(
                id: 'v_cheese',
                name: 'Extra Melted Cheese',
                additionalPricePoisha: 5000,
              ),
              OptionValue(
                id: 'v_bacon',
                name: 'Crispy Beef Bacon',
                additionalPricePoisha: 8000,
              ),
              OptionValue(
                id: 'v_patty',
                name: 'Extra Beef Patty',
                additionalPricePoisha: 14000,
              ),
            ],
          ),
        ],
      ),
      MenuItem(
        id: 'menu_02',
        restaurantId: 'rest_01',
        name: 'Classic Cheese Burger',
        description:
            'Prime beef patty, American cheddar slice, crisp lettuce, tomato, pickles, and house secret sauce.',
        pricePoisha: 32000, // ৳320
        category: 'Burgers',
        isAvailable: true,
        options: [
          MenuItemOption(
            id: 'opt_cheese_type',
            name: 'Cheese Choice',
            isRequired: true,
            allowsMultiple: false,
            values: [
              OptionValue(
                id: 'v_cheddar',
                name: 'Sharp Cheddar',
                isDefault: true,
              ),
              OptionValue(
                id: 'v_swiss',
                name: 'Swiss Cheese (+৳30)',
                additionalPricePoisha: 3000,
              ),
            ],
          ),
        ],
      ),
      MenuItem(
        id: 'menu_03',
        restaurantId: 'rest_01',
        name: 'Large Crispy Fries',
        description:
            'Golden seasoned potato fries served with garlic mayo dipping sauce.',
        pricePoisha: 14000, // ৳140
        category: 'Sides',
        isAvailable: true,
      ),
      MenuItem(
        id: 'menu_04',
        restaurantId: 'rest_01',
        name: 'Choco Lava Cake',
        description: 'Warm chocolate cake with a molten chocolate center.',
        pricePoisha: 18000, // ৳180
        category: 'Dessert',
        isAvailable: true,
      ),
    ],
    'rest_02': const [
      MenuItem(
        id: 'menu_05',
        restaurantId: 'rest_02',
        name: 'Kacchi Khadok Platter',
        description:
            'Fragrant chinigura rice mutton kacchi biryani, alu bokhara, egg, jorda, and refreshing borhani.',
        pricePoisha: 48000, // ৳480
        category: 'Biryani',
        isAvailable: true,
      ),
      MenuItem(
        id: 'menu_06',
        restaurantId: 'rest_02',
        name: 'Shahi Borhani 250ml',
        description:
            'Traditional spiced yogurt drink crafted with mint and secret herbs.',
        pricePoisha: 8000, // ৳80
        category: 'Drinks',
        isAvailable: true,
      ),
    ],
  };

  @override
  Future<Result<List<Restaurant>>> getRestaurants({
    String? query,
    String? category,
  }) async {
    await Future.delayed(const Duration(milliseconds: 150));

    var filtered = _restaurants.toList();

    if (query != null && query.trim().isNotEmpty) {
      final q = query.trim().toLowerCase();
      filtered = filtered.where((r) {
        return r.name.toLowerCase().contains(q) ||
            r.cuisine.toLowerCase().contains(q) ||
            r.area.toLowerCase().contains(q);
      }).toList();
    }

    if (category != null &&
        category.trim().isNotEmpty &&
        category.toLowerCase() != 'all') {
      final cat = category.trim().toLowerCase();
      filtered = filtered.where((r) {
        return r.cuisine.toLowerCase().contains(cat);
      }).toList();
    }

    return Result.success(filtered);
  }

  @override
  Future<Result<Restaurant>> getRestaurantById(String id) async {
    await Future.delayed(const Duration(milliseconds: 100));

    final match = _restaurants.where((r) => r.id == id).firstOrNull;
    if (match == null) {
      return const Result.failure(NotFoundFailure('Restaurant not found.'));
    }
    return Result.success(match);
  }

  @override
  Future<Result<List<MenuItem>>> getRestaurantMenu(String restaurantId) async {
    await Future.delayed(const Duration(milliseconds: 150));

    final menu = _menus[restaurantId];
    if (menu != null) {
      return Result.success(menu);
    }

    // Default to sample menu if specific restaurant menu not configured
    final defaultMenu = _menus['rest_01']!;
    return Result.success(defaultMenu);
  }

  @override
  Future<Result<MenuItem>> getMenuItemById(String menuItemId) async {
    await Future.delayed(const Duration(milliseconds: 100));

    for (final menuList in _menus.values) {
      final item = menuList.where((m) => m.id == menuItemId).firstOrNull;
      if (item != null) return Result.success(item);
    }
    return const Result.failure(NotFoundFailure('Menu item not found.'));
  }
}
