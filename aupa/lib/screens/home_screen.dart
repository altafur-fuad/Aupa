import 'package:flutter/material.dart';
import '../constants/colors.dart';
import '../core/di/service_locator.dart';
import '../core/utils/currency_formatter.dart';
import '../core/widgets/empty_state_view.dart';
import '../core/widgets/error_state_view.dart';
import '../core/widgets/loading_indicator.dart';
import '../domain/models/address.dart';
import '../domain/models/restaurant.dart';
import 'cart/cart_screen.dart';
import 'notifications/notifications_screen.dart';
import 'orders/order_history_screen.dart';
import 'profile/profile_screen.dart';
import 'profile/saved_addresses_screen.dart';
import 'restaurant/restaurant_details_screen.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  final TextEditingController _searchController = TextEditingController();
  String _selectedCategory = 'All';
  Address? _currentAddress;
  List<Restaurant>? _restaurants;
  bool _isLoading = true;
  String? _errorMessage;
  int _currentNavIndex = 0;

  final List<Map<String, dynamic>> _categories = const [
    {'name': 'All', 'icon': Icons.restaurant},
    {'name': 'Burgers', 'icon': Icons.lunch_dining},
    {'name': 'Biryani', 'icon': Icons.rice_bowl},
    {'name': 'Pizza', 'icon': Icons.local_pizza},
    {'name': 'Sides', 'icon': Icons.fastfood},
    {'name': 'Drinks', 'icon': Icons.local_cafe},
    {'name': 'Dessert', 'icon': Icons.icecream},
  ];

  @override
  void initState() {
    super.initState();
    _loadInitialData();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _loadInitialData() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    final addressRes = await ServiceLocator.instance.addressRepository
        .getDefaultAddress();
    if (addressRes.isSuccess) {
      _currentAddress = addressRes.data;
    }

    await _fetchRestaurants();
  }

  Future<void> _fetchRestaurants() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    final query = _searchController.text.trim();
    final cat = _selectedCategory == 'All' ? null : _selectedCategory;

    final result = await ServiceLocator.instance.restaurantRepository
        .getRestaurants(query: query.isEmpty ? null : query, category: cat);

    if (!mounted) return;

    result.when(
      success: (data) {
        setState(() {
          _restaurants = data;
          _isLoading = false;
        });
      },
      failure: (failure) {
        setState(() {
          _errorMessage = failure.message;
          _isLoading = false;
        });
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: _currentNavIndex == 0
          ? _buildHomeAppBar()
          : _currentNavIndex == 1
          ? AppBar(
              backgroundColor: AppColors.background,
              elevation: 0,
              title: const Text(
                'Search & Explore',
                style: TextStyle(
                  color: AppColors.textPrimary,
                  fontWeight: FontWeight.bold,
                  fontSize: 20,
                ),
              ),
            )
          : null,
      body: _buildCurrentBody(),
      bottomNavigationBar: BottomNavigationBar(
        currentIndex: _currentNavIndex,
        onTap: (index) {
          setState(() => _currentNavIndex = index);
        },
        selectedItemColor: AppColors.primaryColor,
        unselectedItemColor: AppColors.textSecondary,
        showUnselectedLabels: true,
        type: BottomNavigationBarType.fixed,
        items: const [
          BottomNavigationBarItem(icon: Icon(Icons.home), label: 'Home'),
          BottomNavigationBarItem(icon: Icon(Icons.search), label: 'Search'),
          BottomNavigationBarItem(
            icon: Icon(Icons.receipt_long),
            label: 'Orders',
          ),
          BottomNavigationBarItem(icon: Icon(Icons.person), label: 'Profile'),
        ],
      ),
    );
  }

  PreferredSizeWidget _buildHomeAppBar() {
    return AppBar(
      backgroundColor: AppColors.background,
      elevation: 0,
      title: InkWell(
        onTap: () async {
          await Navigator.push(
            context,
            MaterialPageRoute(builder: (_) => const SavedAddressesScreen()),
          );
          _loadInitialData();
        },
        child: Row(
          children: [
            const Icon(Icons.location_on, color: AppColors.primaryColor),
            const SizedBox(width: 8),
            Expanded(
              child: Text(
                _currentAddress != null
                    ? '${_currentAddress!.area}, ${_currentAddress!.city}'
                    : 'Dhaka, Bangladesh',
                style: const TextStyle(
                  color: AppColors.textPrimary,
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                ),
                overflow: TextOverflow.ellipsis,
              ),
            ),
            const Icon(Icons.keyboard_arrow_down, color: AppColors.textPrimary),
          ],
        ),
      ),
      actions: [
        IconButton(
          icon: const Icon(
            Icons.shopping_cart_outlined,
            color: AppColors.textPrimary,
          ),
          onPressed: () {
            Navigator.push(
              context,
              MaterialPageRoute(builder: (_) => const CartScreen()),
            );
          },
        ),
        IconButton(
          icon: const Icon(
            Icons.notifications_outlined,
            color: AppColors.textPrimary,
          ),
          onPressed: () {
            Navigator.push(
              context,
              MaterialPageRoute(builder: (_) => const NotificationsScreen()),
            );
          },
        ),
      ],
    );
  }

  Widget _buildCurrentBody() {
    switch (_currentNavIndex) {
      case 0:
        return _buildHomeTab();
      case 1:
        return _buildSearchTab();
      case 2:
        return const OrderHistoryScreen();
      case 3:
        return const ProfileScreen();
      default:
        return _buildHomeTab();
    }
  }

  Widget _buildHomeTab() {
    return RefreshIndicator(
      color: AppColors.primaryColor,
      onRefresh: _fetchRestaurants,
      child: SingleChildScrollView(
        physics: const AlwaysScrollableScrollPhysics(),
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Search Bar
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              decoration: BoxDecoration(
                color: AppColors.greyLight,
                borderRadius: BorderRadius.circular(12),
              ),
              child: TextField(
                controller: _searchController,
                onChanged: (val) {
                  _fetchRestaurants();
                },
                decoration: InputDecoration(
                  icon: const Icon(
                    Icons.search,
                    color: AppColors.textSecondary,
                  ),
                  hintText: 'Search for food, restaurants...',
                  border: InputBorder.none,
                  enabledBorder: InputBorder.none,
                  focusedBorder: InputBorder.none,
                  hintStyle: const TextStyle(color: AppColors.textSecondary),
                  suffixIcon: _searchController.text.isNotEmpty
                      ? IconButton(
                          icon: const Icon(
                            Icons.clear,
                            size: 20,
                            color: AppColors.textSecondary,
                          ),
                          onPressed: () {
                            _searchController.clear();
                            _fetchRestaurants();
                          },
                        )
                      : null,
                ),
              ),
            ),
            const SizedBox(height: 24),

            // Categories
            const Text(
              'Categories',
              style: TextStyle(
                fontSize: 20,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
            const SizedBox(height: 16),
            SizedBox(
              height: 95,
              child: ListView.builder(
                scrollDirection: Axis.horizontal,
                itemCount: _categories.length,
                itemBuilder: (context, index) {
                  final cat = _categories[index];
                  final isSelected = _selectedCategory == cat['name'];
                  return _buildCategoryItem(
                    cat['icon'] as IconData,
                    cat['name'] as String,
                    isSelected,
                  );
                },
              ),
            ),
            const SizedBox(height: 24),

            // Popular Near You
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text(
                  'Popular Near You',
                  style: TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textPrimary,
                  ),
                ),
                TextButton(
                  onPressed: () {
                    setState(() {
                      _selectedCategory = 'All';
                      _searchController.clear();
                    });
                    _fetchRestaurants();
                  },
                  child: const Text(
                    'See All',
                    style: TextStyle(
                      color: AppColors.primaryColor,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 12),

            // Restaurant List / State
            if (_isLoading) ...[
              const AppSkeletonCard(height: 220),
              const SizedBox(height: 16),
              const AppSkeletonCard(height: 220),
            ] else if (_errorMessage != null) ...[
              ErrorStateView(
                message: _errorMessage!,
                onRetry: _fetchRestaurants,
              ),
            ] else if (_restaurants == null || _restaurants!.isEmpty) ...[
              EmptyStateView(
                icon: Icons.search_off,
                title: 'No restaurants found',
                message:
                    'Try adjusting your search keywords or clearing the category filter.',
                actionLabel: 'Reset Filters',
                onAction: () {
                  setState(() {
                    _selectedCategory = 'All';
                    _searchController.clear();
                  });
                  _fetchRestaurants();
                },
              ),
            ] else ...[
              for (final restaurant in _restaurants!) ...[
                _buildRestaurantCard(context, restaurant),
                const SizedBox(height: 16),
              ],
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildSearchTab() {
    final popularSearches = [
      'Biryani',
      'Burgers',
      'Pizza',
      'Kacchi',
      'Platter',
      'Drinks',
    ];

    return Padding(
      padding: const EdgeInsets.all(16.0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            decoration: BoxDecoration(
              color: AppColors.greyLight,
              borderRadius: BorderRadius.circular(12),
            ),
            child: TextField(
              controller: _searchController,
              onChanged: (_) => _fetchRestaurants(),
              decoration: InputDecoration(
                icon: const Icon(Icons.search, color: AppColors.textSecondary),
                hintText: 'Search food, dishes or restaurants...',
                border: InputBorder.none,
                enabledBorder: InputBorder.none,
                focusedBorder: InputBorder.none,
                hintStyle: const TextStyle(color: AppColors.textSecondary),
                suffixIcon: _searchController.text.isNotEmpty
                    ? IconButton(
                        icon: const Icon(
                          Icons.clear,
                          size: 20,
                          color: AppColors.textSecondary,
                        ),
                        onPressed: () {
                          _searchController.clear();
                          _fetchRestaurants();
                        },
                      )
                    : null,
              ),
            ),
          ),
          const SizedBox(height: 16),
          const Text(
            'Popular Searches',
            style: TextStyle(
              fontSize: 14,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: popularSearches.map((term) {
              final isSelected =
                  _searchController.text.toLowerCase() == term.toLowerCase();
              return ActionChip(
                label: Text(term),
                backgroundColor: isSelected
                    ? AppColors.primaryColor
                    : AppColors.greyLight,
                labelStyle: TextStyle(
                  color: isSelected ? Colors.white : AppColors.textPrimary,
                  fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
                  fontSize: 12,
                ),
                onPressed: () {
                  _searchController.text = term;
                  _fetchRestaurants();
                },
              );
            }).toList(),
          ),
          const SizedBox(height: 16),
          Expanded(
            child: _isLoading
                ? const Center(child: AppLoadingIndicator())
                : _restaurants == null || _restaurants!.isEmpty
                ? EmptyStateView(
                    icon: Icons.search_off,
                    title: 'No restaurants found',
                    message:
                        'We could not find anything matching "${_searchController.text.trim()}".',
                  )
                : ListView.separated(
                    itemCount: _restaurants!.length,
                    separatorBuilder: (_, __) => const SizedBox(height: 12),
                    itemBuilder: (context, index) {
                      return _buildRestaurantCard(
                        context,
                        _restaurants![index],
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }

  Widget _buildCategoryItem(IconData icon, String title, bool isSelected) {
    return GestureDetector(
      onTap: () {
        setState(() {
          _selectedCategory = title;
        });
        _fetchRestaurants();
      },
      child: Padding(
        padding: const EdgeInsets.only(right: 16.0),
        child: Column(
          children: [
            Container(
              padding: const EdgeInsets.all(14),
              decoration: BoxDecoration(
                color: isSelected
                    ? AppColors.primaryColor
                    : AppColors.primaryColor.withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(16),
                border: isSelected
                    ? Border.all(color: AppColors.primaryDark, width: 1.5)
                    : null,
              ),
              child: Icon(
                icon,
                color: isSelected ? Colors.white : AppColors.primaryColor,
                size: 28,
              ),
            ),
            const SizedBox(height: 6),
            Text(
              title,
              style: TextStyle(
                color: isSelected
                    ? AppColors.primaryColor
                    : AppColors.textPrimary,
                fontWeight: isSelected ? FontWeight.bold : FontWeight.w500,
                fontSize: 12,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildRestaurantCard(BuildContext context, Restaurant restaurant) {
    final feeText = restaurant.deliveryFeePoisha == 0
        ? 'Free Delivery'
        : '${CurrencyFormatter.formatPoisha(restaurant.deliveryFeePoisha)} delivery';

    return GestureDetector(
      onTap: () {
        Navigator.push(
          context,
          MaterialPageRoute(
            builder: (_) => RestaurantDetailsScreen(restaurant: restaurant),
          ),
        );
      },
      child: Container(
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.05),
              blurRadius: 10,
              offset: const Offset(0, 5),
            ),
          ],
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Stack(
              children: [
                Container(
                  height: 150,
                  decoration: const BoxDecoration(
                    color: AppColors.greyLight,
                    borderRadius: BorderRadius.vertical(
                      top: Radius.circular(16),
                    ),
                  ),
                  child: Center(
                    child: Icon(
                      restaurant.cuisine.contains('Pizza')
                          ? Icons.local_pizza
                          : restaurant.cuisine.contains('Biryani')
                          ? Icons.rice_bowl
                          : Icons.fastfood,
                      size: 50,
                      color: AppColors.textSecondary,
                    ),
                  ),
                ),
                if (!restaurant.isOpen) ...[
                  Positioned(
                    top: 12,
                    right: 12,
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 10,
                        vertical: 4,
                      ),
                      decoration: BoxDecoration(
                        color: Colors.black87,
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: const Text(
                        'Closed',
                        style: TextStyle(
                          color: Colors.white,
                          fontSize: 12,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
                  ),
                ],
              ],
            ),
            Padding(
              padding: const EdgeInsets.all(16.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Expanded(
                        child: Text(
                          restaurant.name,
                          style: const TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                            color: AppColors.textPrimary,
                          ),
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                      const Icon(
                        Icons.favorite_border,
                        color: AppColors.textSecondary,
                      ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  Text(
                    '${restaurant.cuisine} • ${restaurant.area}',
                    style: const TextStyle(
                      color: AppColors.textSecondary,
                      fontSize: 13,
                    ),
                  ),
                  const SizedBox(height: 10),
                  Row(
                    children: [
                      const Icon(Icons.star, color: Colors.amber, size: 16),
                      const SizedBox(width: 4),
                      Text(
                        restaurant.rating.toString(),
                        style: const TextStyle(
                          fontWeight: FontWeight.bold,
                          fontSize: 13,
                        ),
                      ),
                      const SizedBox(width: 4),
                      Text(
                        '(${restaurant.reviewCount}+)',
                        style: const TextStyle(
                          color: AppColors.textSecondary,
                          fontSize: 12,
                        ),
                      ),
                      const SizedBox(width: 16),
                      const Icon(
                        Icons.access_time,
                        color: AppColors.textSecondary,
                        size: 16,
                      ),
                      const SizedBox(width: 4),
                      Text(
                        '${restaurant.deliveryTimeMinutes} mins',
                        style: const TextStyle(
                          color: AppColors.textSecondary,
                          fontSize: 13,
                        ),
                      ),
                      const Spacer(),
                      Text(
                        feeText,
                        style: const TextStyle(
                          color: AppColors.primaryColor,
                          fontWeight: FontWeight.w600,
                          fontSize: 12,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
