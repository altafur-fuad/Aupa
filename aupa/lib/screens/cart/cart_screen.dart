import 'package:flutter/material.dart';
import '../../constants/colors.dart';
import '../../core/cart/cart_manager.dart';
import '../../core/utils/currency_formatter.dart';
import '../../core/widgets/app_button.dart';
import '../../core/widgets/empty_state_view.dart';
import 'checkout_screen.dart';

class CartScreen extends StatefulWidget {
  const CartScreen({super.key});

  @override
  State<CartScreen> createState() => _CartScreenState();
}

class _CartScreenState extends State<CartScreen> {
  final TextEditingController _promoController = TextEditingController();
  int _discountPoisha = 0;
  String? _appliedPromo;

  @override
  void dispose() {
    _promoController.dispose();
    super.dispose();
  }

  void _applyPromoCode() {
    final code = _promoController.text.trim().toUpperCase();
    if (code == 'AUPA50') {
      setState(() {
        _discountPoisha = 5000; // ৳50
        _appliedPromo = code;
      });
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Promo AUPA50 applied! ৳50 discount added.'),
          backgroundColor: AppColors.success,
        ),
      );
    } else if (code.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please enter a promo code.')),
      );
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Invalid promo code. Try "AUPA50" for ৳50 off!'),
          backgroundColor: AppColors.error,
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return ListenableBuilder(
      listenable: CartManager.instance,
      builder: (context, _) {
        final cart = CartManager.instance;
        final subtotal = cart.subtotalPoisha;
        final deliveryFee = cart.deliveryFeePoisha;
        final discount = _discountPoisha > subtotal
            ? subtotal
            : _discountPoisha;
        final total = (subtotal + deliveryFee - discount).clamp(0, 999999999);

        return Scaffold(
          backgroundColor: AppColors.background,
          appBar: AppBar(
            title: const Text(
              'Your Cart',
              style: TextStyle(color: AppColors.textPrimary),
            ),
            backgroundColor: Colors.transparent,
            elevation: 0,
            iconTheme: const IconThemeData(color: AppColors.textPrimary),
            actions: [
              if (!cart.isEmpty)
                TextButton(
                  onPressed: () {
                    cart.clear();
                    setState(() {
                      _discountPoisha = 0;
                      _appliedPromo = null;
                    });
                  },
                  child: const Text(
                    'Clear',
                    style: TextStyle(color: AppColors.error),
                  ),
                ),
            ],
          ),
          body: cart.isEmpty
              ? EmptyStateView(
                  icon: Icons.shopping_basket_outlined,
                  title: 'Your cart is empty',
                  message:
                      'Looks like you haven\'t added any food yet. Browse top restaurants nearby!',
                  actionLabel: 'Find Food',
                  onAction: () => Navigator.pop(context),
                )
              : ListView(
                  padding: const EdgeInsets.all(20),
                  children: [
                    if (cart.restaurantName != null) ...[
                      Row(
                        children: [
                          const Icon(
                            Icons.storefront,
                            size: 20,
                            color: AppColors.primaryColor,
                          ),
                          const SizedBox(width: 8),
                          Text(
                            cart.restaurantName!,
                            style: const TextStyle(
                              fontSize: 18,
                              fontWeight: FontWeight.bold,
                              color: AppColors.textPrimary,
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 16),
                    ],

                    // Cart Items List
                    for (int i = 0; i < cart.items.length; i++) ...[
                      _buildCartItem(cart.items[i], i, cart),
                    ],

                    const SizedBox(height: 24),

                    // Promo Code Section
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 6,
                      ),
                      decoration: BoxDecoration(
                        color: AppColors.greyLight,
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Row(
                        children: [
                          const Icon(
                            Icons.local_offer_outlined,
                            color: AppColors.primaryColor,
                          ),
                          const SizedBox(width: 12),
                          Expanded(
                            child: TextField(
                              controller: _promoController,
                              decoration: InputDecoration(
                                hintText: _appliedPromo != null
                                    ? 'Promo: $_appliedPromo'
                                    : 'Promo Code (try AUPA50)',
                                border: InputBorder.none,
                                enabledBorder: InputBorder.none,
                                focusedBorder: InputBorder.none,
                                contentPadding: EdgeInsets.zero,
                              ),
                            ),
                          ),
                          TextButton(
                            onPressed: _applyPromoCode,
                            child: Text(
                              _appliedPromo != null ? 'Applied' : 'Apply',
                              style: TextStyle(
                                color: _appliedPromo != null
                                    ? AppColors.success
                                    : AppColors.primaryColor,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 24),

                    // Price Breakdown
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const Text(
                          'Subtotal',
                          style: TextStyle(color: AppColors.textSecondary),
                        ),
                        Text(
                          CurrencyFormatter.formatPoisha(subtotal),
                          style: const TextStyle(fontWeight: FontWeight.bold),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const Text(
                          'Delivery Fee',
                          style: TextStyle(color: AppColors.textSecondary),
                        ),
                        Text(
                          CurrencyFormatter.formatPoisha(deliveryFee),
                          style: const TextStyle(fontWeight: FontWeight.bold),
                        ),
                      ],
                    ),
                    if (discount > 0) ...[
                      const SizedBox(height: 8),
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          const Text(
                            'Promo Discount',
                            style: TextStyle(color: AppColors.success),
                          ),
                          Text(
                            '-${CurrencyFormatter.formatPoisha(discount)}',
                            style: const TextStyle(
                              fontWeight: FontWeight.bold,
                              color: AppColors.success,
                            ),
                          ),
                        ],
                      ),
                    ],
                    const Divider(height: 32),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const Text(
                          'Total',
                          style: TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        Text(
                          CurrencyFormatter.formatPoisha(total),
                          style: const TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                            color: AppColors.primaryColor,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 80),
                  ],
                ),
          bottomNavigationBar: cart.isEmpty
              ? null
              : SafeArea(
                  child: Padding(
                    padding: const EdgeInsets.all(20.0),
                    child: AppButton(
                      label:
                          'Proceed to Checkout (${CurrencyFormatter.formatPoisha(total)})',
                      onPressed: () {
                        Navigator.push(
                          context,
                          MaterialPageRoute(
                            builder: (_) =>
                                CheckoutScreen(discountPoisha: discount),
                          ),
                        );
                      },
                    ),
                  ),
                ),
        );
      },
    );
  }

  Widget _buildCartItem(CartItem item, int index, CartManager cart) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            width: 55,
            height: 55,
            decoration: BoxDecoration(
              color: AppColors.greyLight,
              borderRadius: BorderRadius.circular(10),
            ),
            child: const Icon(Icons.fastfood, color: AppColors.textSecondary),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  item.name,
                  style: const TextStyle(
                    fontWeight: FontWeight.bold,
                    fontSize: 15,
                  ),
                ),
                if (item.selectedOptionNames.isNotEmpty) ...[
                  const SizedBox(height: 2),
                  Text(
                    item.selectedOptionNames.join(', '),
                    style: const TextStyle(
                      color: AppColors.textSecondary,
                      fontSize: 12,
                    ),
                  ),
                ],
                const SizedBox(height: 6),
                Text(
                  CurrencyFormatter.formatPoisha(item.lineTotalPoisha),
                  style: const TextStyle(
                    color: AppColors.primaryColor,
                    fontWeight: FontWeight.bold,
                    fontSize: 14,
                  ),
                ),
              ],
            ),
          ),
          Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              IconButton(
                visualDensity: VisualDensity.compact,
                icon: const Icon(
                  Icons.remove_circle_outline,
                  size: 22,
                  color: AppColors.textSecondary,
                ),
                onPressed: () {
                  cart.updateQuantity(index, item.quantity - 1);
                },
              ),
              Text(
                '${item.quantity}',
                style: const TextStyle(
                  fontWeight: FontWeight.bold,
                  fontSize: 16,
                ),
              ),
              IconButton(
                visualDensity: VisualDensity.compact,
                icon: const Icon(
                  Icons.add_circle_outline,
                  size: 22,
                  color: AppColors.primaryColor,
                ),
                onPressed: () {
                  cart.updateQuantity(index, item.quantity + 1);
                },
              ),
            ],
          ),
        ],
      ),
    );
  }
}
