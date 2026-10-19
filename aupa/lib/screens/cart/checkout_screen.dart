import 'dart:math';
import 'package:flutter/material.dart';
import '../../constants/colors.dart';
import '../../core/cart/cart_manager.dart';
import '../../core/di/service_locator.dart';
import '../../core/utils/currency_formatter.dart';
import '../../core/widgets/app_button.dart';
import '../../core/widgets/loading_indicator.dart';
import '../../domain/models/address.dart';
import '../../domain/models/order.dart';
import '../../domain/repositories/order_repository.dart';
import '../home_screen.dart';

class CheckoutScreen extends StatefulWidget {
  final int discountPoisha;

  const CheckoutScreen({super.key, this.discountPoisha = 0});

  @override
  State<CheckoutScreen> createState() => _CheckoutScreenState();
}

class _CheckoutScreenState extends State<CheckoutScreen> {
  List<Address> _addresses = [];
  Address? _selectedAddress;
  PaymentMethod _selectedPaymentMethod = PaymentMethod.cod;
  bool _isLoadingAddresses = true;
  bool _isSubmitting = false;

  @override
  void initState() {
    super.initState();
    _loadAddresses();
  }

  Future<void> _loadAddresses() async {
    final res = await ServiceLocator.instance.addressRepository.getAddresses();
    if (!mounted) return;
    if (res.isSuccess) {
      setState(() {
        _addresses = res.data;
        _selectedAddress =
            res.data.where((a) => a.isDefault).firstOrNull ??
            res.data.firstOrNull;
        _isLoadingAddresses = false;
      });
    } else {
      setState(() => _isLoadingAddresses = false);
    }
  }

  void _showAddressSelector() {
    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (context) {
        return Padding(
          padding: const EdgeInsets.all(20.0),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text(
                'Choose Delivery Address',
                style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 16),
              for (final addr in _addresses) ...[
                ListTile(
                  leading: Icon(
                    addr.label == AddressLabel.home ? Icons.home : Icons.work,
                    color: AppColors.primaryColor,
                  ),
                  title: Text(
                    addr.displayLabel,
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                  subtitle: Text(addr.fullAddress),
                  trailing: _selectedAddress?.id == addr.id
                      ? const Icon(
                          Icons.check_circle,
                          color: AppColors.primaryColor,
                        )
                      : null,
                  onTap: () {
                    setState(() => _selectedAddress = addr);
                    Navigator.pop(context);
                  },
                ),
              ],
            ],
          ),
        );
      },
    );
  }

  Future<void> _handlePlaceOrder() async {
    final cart = CartManager.instance;
    if (cart.isEmpty) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('Cart is empty.')));
      return;
    }

    if (_selectedAddress == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please select a delivery address.')),
      );
      return;
    }

    setState(() => _isSubmitting = true);

    // Generate high-entropy idempotency key
    final idempotencyKey =
        'idem_${DateTime.now().millisecondsSinceEpoch}_${Random().nextInt(999999)}';

    final orderItems = cart.items
        .map(
          (ci) => CreateOrderItemParam(
            menuItemId: ci.menuItemId,
            name: ci.name,
            quantity: ci.quantity,
            unitPricePoisha: ci.unitPricePoisha,
            selectedOptions: ci.selectedOptionNames,
          ),
        )
        .toList();

    final params = CreateOrderParams(
      restaurantId: cart.restaurantId ?? 'rest_01',
      restaurantName: cart.restaurantName ?? 'Burger Barn',
      addressId: _selectedAddress!.id,
      paymentMethod: _selectedPaymentMethod,
      items: orderItems,
      idempotencyKey: idempotencyKey,
    );

    final result = await ServiceLocator.instance.orderRepository.createOrder(
      params,
    );

    if (!mounted) return;
    setState(() => _isSubmitting = false);

    result.when(
      success: (createdOrder) {
        cart.clear();
        _showSuccessDialog(context, createdOrder);
      },
      failure: (failure) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(failure.message),
            backgroundColor: AppColors.error,
          ),
        );
      },
    );
  }

  void _showSuccessDialog(BuildContext context, Order order) {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        content: Padding(
          padding: const EdgeInsets.all(16.0),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              const Icon(
                Icons.check_circle,
                color: AppColors.primaryColor,
                size: 80,
              ),
              const SizedBox(height: 20),
              const Text(
                'Order Placed!',
                style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
              ),
              const SizedBox(height: 8),
              Text(
                'Order #${order.orderNumber}\nYour food is being prepared and will be delivered soon.',
                textAlign: TextAlign.center,
                style: const TextStyle(
                  color: AppColors.textSecondary,
                  height: 1.4,
                ),
              ),
              const SizedBox(height: 16),
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: AppColors.greyLight,
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text(
                      'Total Amount:',
                      style: TextStyle(color: AppColors.textSecondary),
                    ),
                    Text(
                      CurrencyFormatter.formatPoisha(order.amounts.totalPoisha),
                      style: const TextStyle(
                        fontWeight: FontWeight.bold,
                        color: AppColors.primaryColor,
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 24),
              AppButton(
                label: 'Back to Home',
                onPressed: () {
                  Navigator.pushAndRemoveUntil(
                    context,
                    MaterialPageRoute(builder: (_) => const HomeScreen()),
                    (route) => false,
                  );
                },
              ),
            ],
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final cart = CartManager.instance;
    final subtotal = cart.subtotalPoisha;
    final deliveryFee = cart.deliveryFeePoisha;
    final discount = widget.discountPoisha > subtotal
        ? subtotal
        : widget.discountPoisha;
    final total = (subtotal + deliveryFee - discount).clamp(0, 999999999);

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        title: const Text(
          'Checkout',
          style: TextStyle(color: AppColors.textPrimary),
        ),
        backgroundColor: Colors.transparent,
        elevation: 0,
        iconTheme: const IconThemeData(color: AppColors.textPrimary),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Delivery Address Section
            const Text(
              'Delivery Address',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 12),
            if (_isLoadingAddresses) ...[
              const AppSkeletonCard(height: 70),
            ] else ...[
              InkWell(
                onTap: _showAddressSelector,
                borderRadius: BorderRadius.circular(12),
                child: Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    border: Border.all(color: AppColors.primaryColor),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Row(
                    children: [
                      const Icon(
                        Icons.location_on,
                        color: AppColors.primaryColor,
                      ),
                      const SizedBox(width: 14),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              _selectedAddress?.displayLabel ??
                                  'Select Address',
                              style: const TextStyle(
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                            const SizedBox(height: 2),
                            Text(
                              _selectedAddress?.fullAddress ??
                                  'No address selected',
                              style: const TextStyle(
                                color: AppColors.textSecondary,
                                fontSize: 13,
                              ),
                              maxLines: 2,
                              overflow: TextOverflow.ellipsis,
                            ),
                          ],
                        ),
                      ),
                      const Icon(
                        Icons.keyboard_arrow_down,
                        color: AppColors.textSecondary,
                      ),
                    ],
                  ),
                ),
              ),
            ],

            const SizedBox(height: 28),

            // Payment Method Section
            const Text(
              'Payment Method',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 12),
            _buildPaymentOption(
              PaymentMethod.cod,
              'Cash on Delivery',
              Icons.money,
              'Pay with cash upon food arrival',
            ),
            const SizedBox(height: 10),
            _buildPaymentOption(
              PaymentMethod.bkash,
              'bKash Online',
              Icons.account_balance_wallet,
              'Fast & secure mobile payment',
            ),
            const SizedBox(height: 10),
            _buildPaymentOption(
              PaymentMethod.card,
              'Credit / Debit Card',
              Icons.credit_card,
              'Visa, Mastercard, Amex',
            ),

            const SizedBox(height: 28),

            // Order Summary
            const Text(
              'Order Summary',
              style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 12),
            Container(
              padding: const EdgeInsets.all(16),
              decoration: BoxDecoration(
                color: AppColors.greyLight,
                borderRadius: BorderRadius.circular(12),
              ),
              child: Column(
                children: [
                  for (final item in cart.items) ...[
                    Padding(
                      padding: const EdgeInsets.symmetric(vertical: 4.0),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Expanded(
                            child: Text(
                              '${item.quantity}x ${item.name}',
                              style: const TextStyle(fontSize: 14),
                              overflow: TextOverflow.ellipsis,
                            ),
                          ),
                          Text(
                            CurrencyFormatter.formatPoisha(
                              item.lineTotalPoisha,
                            ),
                            style: const TextStyle(
                              fontWeight: FontWeight.w600,
                              fontSize: 14,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                  const Divider(height: 20),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      const Text(
                        'Subtotal',
                        style: TextStyle(color: AppColors.textSecondary),
                      ),
                      Text(CurrencyFormatter.formatPoisha(subtotal)),
                    ],
                  ),
                  const SizedBox(height: 6),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      const Text(
                        'Delivery Fee',
                        style: TextStyle(color: AppColors.textSecondary),
                      ),
                      Text(CurrencyFormatter.formatPoisha(deliveryFee)),
                    ],
                  ),
                  if (discount > 0) ...[
                    const SizedBox(height: 6),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const Text(
                          'Discount',
                          style: TextStyle(color: AppColors.success),
                        ),
                        Text(
                          '-${CurrencyFormatter.formatPoisha(discount)}',
                          style: const TextStyle(color: AppColors.success),
                        ),
                      ],
                    ),
                  ],
                  const Divider(height: 20),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      const Text(
                        'Total to Pay',
                        style: TextStyle(
                          fontSize: 16,
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
                ],
              ),
            ),
            const SizedBox(height: 80),
          ],
        ),
      ),
      bottomNavigationBar: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(20.0),
          child: AppButton(
            label: 'Place Order - ${CurrencyFormatter.formatPoisha(total)}',
            isLoading: _isSubmitting,
            onPressed: _handlePlaceOrder,
          ),
        ),
      ),
    );
  }

  Widget _buildPaymentOption(
    PaymentMethod method,
    String title,
    IconData icon,
    String subtitle,
  ) {
    final isSelected = _selectedPaymentMethod == method;
    return InkWell(
      onTap: () => setState(() => _selectedPaymentMethod = method),
      borderRadius: BorderRadius.circular(12),
      child: Container(
        padding: const EdgeInsets.all(14),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: isSelected ? AppColors.primaryColor : Colors.grey.shade300,
            width: isSelected ? 1.5 : 1,
          ),
        ),
        child: Row(
          children: [
            Icon(
              icon,
              color: isSelected
                  ? AppColors.primaryColor
                  : AppColors.textSecondary,
            ),
            const SizedBox(width: 14),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: const TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 14,
                    ),
                  ),
                  Text(
                    subtitle,
                    style: const TextStyle(
                      color: AppColors.textSecondary,
                      fontSize: 12,
                    ),
                  ),
                ],
              ),
            ),
            Icon(
              isSelected ? Icons.radio_button_checked : Icons.radio_button_off,
              color: isSelected ? AppColors.primaryColor : Colors.grey.shade400,
            ),
          ],
        ),
      ),
    );
  }
}
