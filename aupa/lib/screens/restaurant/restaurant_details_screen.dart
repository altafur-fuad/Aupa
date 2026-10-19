import 'package:flutter/material.dart';
import '../../constants/colors.dart';
import '../../core/di/service_locator.dart';
import '../../core/utils/currency_formatter.dart';
import '../../core/widgets/error_state_view.dart';
import '../../core/widgets/loading_indicator.dart';
import '../../domain/models/menu_item.dart';
import '../../domain/models/restaurant.dart';
import '../food/food_details_screen.dart';

class RestaurantDetailsScreen extends StatefulWidget {
  final Restaurant? restaurant;

  const RestaurantDetailsScreen({super.key, this.restaurant});

  @override
  State<RestaurantDetailsScreen> createState() =>
      _RestaurantDetailsScreenState();
}

class _RestaurantDetailsScreenState extends State<RestaurantDetailsScreen> {
  late final Restaurant _restaurant;
  List<MenuItem>? _menuItems;
  bool _isLoading = true;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _restaurant =
        widget.restaurant ??
        const Restaurant(
          id: 'rest_01',
          name: 'Burger Barn',
          cuisine: 'Burgers & Fast Food',
          area: 'Gulshan 2',
          address: 'House 14, Road 113, Gulshan 2, Dhaka',
          rating: 4.8,
          reviewCount: 2150,
          deliveryTimeMinutes: 25,
          deliveryFeePoisha: 6000,
        );
    _loadMenu();
  }

  Future<void> _loadMenu() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    final res = await ServiceLocator.instance.restaurantRepository
        .getRestaurantMenu(_restaurant.id);

    if (!mounted) return;

    res.when(
      success: (items) {
        setState(() {
          _menuItems = items;
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
      body: CustomScrollView(
        slivers: [
          SliverAppBar(
            expandedHeight: 200,
            pinned: true,
            backgroundColor: AppColors.primaryColor,
            foregroundColor: Colors.white,
            flexibleSpace: FlexibleSpaceBar(
              title: Text(
                _restaurant.name,
                style: const TextStyle(
                  fontWeight: FontWeight.bold,
                  fontSize: 18,
                ),
              ),
              background: Container(
                color: AppColors.primaryDark,
                child: Center(
                  child: Icon(
                    _restaurant.cuisine.contains('Pizza')
                        ? Icons.local_pizza
                        : _restaurant.cuisine.contains('Biryani')
                        ? Icons.rice_bowl
                        : Icons.fastfood,
                    size: 80,
                    color: Colors.white54,
                  ),
                ),
              ),
            ),
          ),
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.all(16.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      const Icon(Icons.star, color: Colors.amber, size: 20),
                      const SizedBox(width: 4),
                      Text(
                        '${_restaurant.rating} (${_restaurant.reviewCount}+ reviews)',
                        style: const TextStyle(fontWeight: FontWeight.bold),
                      ),
                      const Spacer(),
                      const Icon(
                        Icons.access_time,
                        color: AppColors.textSecondary,
                        size: 20,
                      ),
                      const SizedBox(width: 4),
                      Text(
                        '${_restaurant.deliveryTimeMinutes} mins',
                        style: const TextStyle(color: AppColors.textSecondary),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Text(
                    '${_restaurant.cuisine} • ${_restaurant.address}',
                    style: const TextStyle(
                      color: AppColors.textSecondary,
                      fontSize: 13,
                    ),
                  ),
                  const SizedBox(height: 24),
                  const Text(
                    'Menu',
                    style: TextStyle(
                      fontSize: 22,
                      fontWeight: FontWeight.bold,
                      color: AppColors.textPrimary,
                    ),
                  ),
                  const SizedBox(height: 16),
                  if (_isLoading) ...[
                    const AppSkeletonCard(height: 100),
                    const AppSkeletonCard(height: 100),
                    const AppSkeletonCard(height: 100),
                  ] else if (_errorMessage != null) ...[
                    ErrorStateView(message: _errorMessage!, onRetry: _loadMenu),
                  ] else if (_menuItems != null) ...[
                    for (final item in _menuItems!) ...[
                      _buildMenuItem(context, item),
                    ],
                  ],
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildMenuItem(BuildContext context, MenuItem item) {
    return GestureDetector(
      onTap: () {
        Navigator.push(
          context,
          MaterialPageRoute(
            builder: (_) => FoodDetailsScreen(
              menuItem: item,
              restaurantName: _restaurant.name,
              restaurantId: _restaurant.id,
            ),
          ),
        );
      },
      child: Container(
        margin: const EdgeInsets.only(bottom: 16),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
          border: Border.all(color: Colors.grey.shade200),
        ),
        child: Row(
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    item.name,
                    style: const TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    item.description,
                    style: const TextStyle(
                      color: AppColors.textSecondary,
                      fontSize: 12,
                    ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                  const SizedBox(height: 8),
                  Text(
                    CurrencyFormatter.formatPoisha(item.pricePoisha),
                    style: const TextStyle(
                      fontWeight: FontWeight.bold,
                      color: AppColors.primaryColor,
                    ),
                  ),
                ],
              ),
            ),
            const SizedBox(width: 16),
            Container(
              width: 80,
              height: 80,
              decoration: BoxDecoration(
                color: AppColors.greyLight,
                borderRadius: BorderRadius.circular(12),
              ),
              child: Icon(
                item.category == 'Drinks'
                    ? Icons.local_cafe
                    : item.category == 'Biryani'
                    ? Icons.rice_bowl
                    : Icons.lunch_dining,
                color: AppColors.textSecondary,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
