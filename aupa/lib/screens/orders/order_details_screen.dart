import 'package:flutter/material.dart';
import '../../constants/colors.dart';
import '../../core/di/service_locator.dart';
import '../../core/utils/currency_formatter.dart';
import '../../core/widgets/app_button.dart';
import '../../core/widgets/error_state_view.dart';
import '../../core/widgets/loading_indicator.dart';
import '../../domain/models/order.dart';
import 'delivery_tracking_screen.dart';

class OrderDetailsScreen extends StatefulWidget {
  final String orderId;

  const OrderDetailsScreen({super.key, required this.orderId});

  @override
  State<OrderDetailsScreen> createState() => _OrderDetailsScreenState();
}

class _OrderDetailsScreenState extends State<OrderDetailsScreen> {
  Order? _order;
  bool _isLoading = true;
  bool _isCancelling = false;
  String? _errorMessage;

  @override
  void initState() {
    super.initState();
    _loadOrder();
  }

  Future<void> _loadOrder() async {
    setState(() {
      _isLoading = true;
      _errorMessage = null;
    });

    final res = await ServiceLocator.instance.orderRepository.getOrderById(
      widget.orderId,
    );
    if (!mounted) return;

    res.when(
      success: (data) {
        setState(() {
          _order = data;
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

  Future<void> _handleCancelOrder() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        title: const Text('Cancel Order?'),
        content: const Text('Are you sure you want to cancel this order?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('No, Keep Order'),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(backgroundColor: AppColors.error),
            onPressed: () => Navigator.pop(context, true),
            child: const Text(
              'Yes, Cancel',
              style: TextStyle(color: Colors.white),
            ),
          ),
        ],
      ),
    );

    if (confirmed != true) return;

    setState(() => _isCancelling = true);
    final res = await ServiceLocator.instance.orderRepository.cancelOrder(
      widget.orderId,
      reason: 'Customer requested cancellation',
    );

    if (!mounted) return;
    setState(() => _isCancelling = false);

    res.when(
      success: (updated) {
        setState(() => _order = updated);
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Order has been cancelled.')),
        );
      },
      failure: (f) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(f.message), backgroundColor: AppColors.error),
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        title: Text(
          _order != null ? '#${_order!.orderNumber}' : 'Order Details',
        ),
        backgroundColor: Colors.transparent,
        elevation: 0,
        iconTheme: const IconThemeData(color: AppColors.textPrimary),
      ),
      body: _isLoading
          ? const Center(
              child: AppLoadingIndicator(message: 'Loading order...'),
            )
          : _errorMessage != null
          ? ErrorStateView(message: _errorMessage!, onRetry: _loadOrder)
          : _order == null
          ? const Center(child: Text('Order not found'))
          : ListView(
              padding: const EdgeInsets.all(20),
              children: [
                // Status Card
                Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: AppColors.primaryColor.withValues(alpha: 0.08),
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(
                      color: AppColors.primaryColor.withValues(alpha: 0.2),
                    ),
                  ),
                  child: Row(
                    children: [
                      const Icon(
                        Icons.delivery_dining,
                        size: 40,
                        color: AppColors.primaryColor,
                      ),
                      const SizedBox(width: 14),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              _order!.status.displayName,
                              style: const TextStyle(
                                fontSize: 18,
                                fontWeight: FontWeight.bold,
                                color: AppColors.primaryColor,
                              ),
                            ),
                            const SizedBox(height: 2),
                            Text(
                              _order!.status == OrderStatus.delivered
                                  ? 'Delivered safely.'
                                  : _order!.status == OrderStatus.cancelled
                                  ? 'Order was cancelled.'
                                  : 'Estimated arrival in ~25 mins',
                              style: const TextStyle(
                                color: AppColors.textSecondary,
                                fontSize: 13,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 24),

                // Restaurant Info
                const Text(
                  'Restaurant',
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                ),
                const SizedBox(height: 8),
                Text(
                  _order!.restaurantName,
                  style: const TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textPrimary,
                  ),
                ),
                const SizedBox(height: 20),

                // Order Items Breakdown
                const Text(
                  'Items Ordered',
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
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
                      for (final item in _order!.items) ...[
                        Padding(
                          padding: const EdgeInsets.symmetric(vertical: 6.0),
                          child: Row(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                '${item.quantity}x ',
                                style: const TextStyle(
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      item.name,
                                      style: const TextStyle(
                                        fontWeight: FontWeight.w500,
                                      ),
                                    ),
                                    if (item.selectedOptions.isNotEmpty)
                                      Text(
                                        item.selectedOptions.join(', '),
                                        style: const TextStyle(
                                          color: AppColors.textSecondary,
                                          fontSize: 12,
                                        ),
                                      ),
                                  ],
                                ),
                              ),
                              Text(
                                CurrencyFormatter.formatPoisha(
                                  item.totalPricePoisha,
                                ),
                                style: const TextStyle(
                                  fontWeight: FontWeight.bold,
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
                          Text(
                            CurrencyFormatter.formatPoisha(
                              _order!.amounts.subtotalPoisha,
                            ),
                          ),
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
                          Text(
                            CurrencyFormatter.formatPoisha(
                              _order!.amounts.deliveryFeePoisha,
                            ),
                          ),
                        ],
                      ),
                      const Divider(height: 20),
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          const Text(
                            'Total',
                            style: TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          Text(
                            CurrencyFormatter.formatPoisha(
                              _order!.amounts.totalPoisha,
                            ),
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
                const SizedBox(height: 24),

                // Delivery Address Card
                const Text(
                  'Delivery Address',
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                ),
                const SizedBox(height: 8),
                Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    border: Border.all(color: Colors.grey.shade300),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Row(
                    children: [
                      const Icon(
                        Icons.location_on,
                        color: AppColors.primaryColor,
                      ),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Text(
                          _order!.deliveryAddress.fullAddress,
                          style: const TextStyle(fontSize: 14),
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 32),

                // Action Buttons
                if (_order!.status != OrderStatus.delivered &&
                    _order!.status != OrderStatus.cancelled) ...[
                  AppButton(
                    label: 'Track Delivery Live',
                    icon: Icons.map,
                    onPressed: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (_) =>
                              DeliveryTrackingScreen(order: _order!),
                        ),
                      );
                    },
                  ),
                  const SizedBox(height: 12),
                ],

                if (_order!.status == OrderStatus.placed ||
                    _order!.status == OrderStatus.accepted) ...[
                  AppButton(
                    label: 'Cancel Order',
                    variant: AppButtonVariant.outline,
                    isLoading: _isCancelling,
                    onPressed: _handleCancelOrder,
                  ),
                ],
              ],
            ),
    );
  }
}
