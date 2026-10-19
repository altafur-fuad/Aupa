import '../../core/errors/failures.dart';
import '../../core/network/result.dart';
import '../../domain/models/address.dart';
import '../../domain/models/order.dart';
import '../../domain/repositories/order_repository.dart';

class MockOrderRepository implements OrderRepository {
  final List<Order> _orders = [
    Order(
      id: 'ord_1024',
      orderNumber: 'AUPA-2026-001024',
      restaurantId: 'rest_01',
      restaurantName: 'Burger Barn',
      status: OrderStatus.enRoute,
      paymentStatus: PaymentStatus.paid,
      paymentMethod: PaymentMethod.bkash,
      amounts: const OrderAmounts(
        subtotalPoisha: 78000, // ৳780
        deliveryFeePoisha: 6000, // ৳60
        discountPoisha: 0,
        taxPoisha: 0,
        totalPoisha: 84000, // ৳840
      ),
      items: const [
        OrderItem(
          menuItemId: 'menu_01',
          name: 'Spicy BBQ Burger',
          quantity: 2,
          unitPricePoisha: 39000,
          selectedOptions: ['Mild', 'Extra Melted Cheese'],
          totalPricePoisha: 78000,
        ),
      ],
      deliveryAddress: const Address(
        id: 'addr_01',
        label: AddressLabel.home,
        street: 'House 45, Road 7, Block C',
        area: 'Banani',
        city: 'Dhaka',
        isDefault: true,
      ),
      createdAt: DateTime.now().subtract(const Duration(minutes: 15)),
      estimatedDeliveryTime: DateTime.now().add(const Duration(minutes: 10)),
    ),
    Order(
      id: 'ord_1023',
      orderNumber: 'AUPA-2026-001023',
      restaurantId: 'rest_02',
      restaurantName: "Sultan's Dine",
      status: OrderStatus.delivered,
      paymentStatus: PaymentStatus.paid,
      paymentMethod: PaymentMethod.cod,
      amounts: const OrderAmounts(
        subtotalPoisha: 56000, // ৳560
        deliveryFeePoisha: 5000, // ৳50
        discountPoisha: 0,
        taxPoisha: 0,
        totalPoisha: 61000, // ৳610
      ),
      items: const [
        OrderItem(
          menuItemId: 'menu_05',
          name: 'Kacchi Khadok Platter',
          quantity: 1,
          unitPricePoisha: 48000,
          totalPricePoisha: 48000,
        ),
        OrderItem(
          menuItemId: 'menu_06',
          name: 'Shahi Borhani 250ml',
          quantity: 1,
          unitPricePoisha: 8000,
          totalPricePoisha: 8000,
        ),
      ],
      deliveryAddress: const Address(
        id: 'addr_01',
        label: AddressLabel.home,
        street: 'House 45, Road 7, Block C',
        area: 'Banani',
        city: 'Dhaka',
        isDefault: true,
      ),
      createdAt: DateTime.now().subtract(const Duration(days: 2)),
    ),
  ];

  @override
  Future<Result<Order>> createOrder(CreateOrderParams params) async {
    await Future.delayed(const Duration(milliseconds: 300));

    if (params.items.isEmpty) {
      return const Result.failure(
        ValidationFailure('Order must contain at least one item.'),
      );
    }

    int subtotal = 0;
    final orderItems = <OrderItem>[];

    for (final item in params.items) {
      final lineTotal = item.unitPricePoisha * item.quantity;
      subtotal += lineTotal;
      orderItems.add(
        OrderItem(
          menuItemId: item.menuItemId,
          name: item.name,
          quantity: item.quantity,
          unitPricePoisha: item.unitPricePoisha,
          selectedOptions: item.selectedOptions,
          totalPricePoisha: lineTotal,
        ),
      );
    }

    const deliveryFee = 6000; // ৳60
    final total = subtotal + deliveryFee;

    final order = Order(
      id: 'ord_${DateTime.now().millisecondsSinceEpoch}',
      orderNumber: 'AUPA-2026-00${_orders.length + 1025}',
      restaurantId: params.restaurantId,
      restaurantName: params.restaurantName,
      status: OrderStatus.placed,
      paymentStatus: params.paymentMethod == PaymentMethod.cod
          ? PaymentStatus.notRequired
          : PaymentStatus.paid,
      paymentMethod: params.paymentMethod,
      amounts: OrderAmounts(
        subtotalPoisha: subtotal,
        deliveryFeePoisha: deliveryFee,
        totalPoisha: total,
      ),
      items: orderItems,
      deliveryAddress: const Address(
        id: 'addr_01',
        label: AddressLabel.home,
        street: 'House 45, Road 7, Block C',
        area: 'Banani',
        city: 'Dhaka',
        isDefault: true,
      ),
      createdAt: DateTime.now(),
      estimatedDeliveryTime: DateTime.now().add(const Duration(minutes: 30)),
    );

    _orders.insert(0, order);
    return Result.success(order);
  }

  @override
  Future<Result<List<Order>>> getOrders() async {
    await Future.delayed(const Duration(milliseconds: 150));
    return Result.success(List.unmodifiable(_orders));
  }

  @override
  Future<Result<Order>> getOrderById(String orderId) async {
    await Future.delayed(const Duration(milliseconds: 100));
    final match = _orders.where((o) => o.id == orderId).firstOrNull;
    if (match == null) {
      return const Result.failure(NotFoundFailure('Order not found.'));
    }
    return Result.success(match);
  }

  @override
  Future<Result<Order>> cancelOrder(
    String orderId, {
    required String reason,
  }) async {
    await Future.delayed(const Duration(milliseconds: 200));
    final index = _orders.indexWhere((o) => o.id == orderId);
    if (index == -1) {
      return const Result.failure(NotFoundFailure('Order not found.'));
    }

    final current = _orders[index];
    if (current.status == OrderStatus.enRoute ||
        current.status == OrderStatus.delivered) {
      return const Result.failure(
        ConflictFailure(
          'Cannot cancel an order that is already on the way or delivered.',
        ),
      );
    }

    final updated = current.copyWith(status: OrderStatus.cancelled);
    _orders[index] = updated;
    return Result.success(updated);
  }
}
