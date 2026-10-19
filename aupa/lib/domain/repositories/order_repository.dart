import '../../core/network/result.dart';
import '../models/order.dart';

class CreateOrderItemParam {
  final String menuItemId;
  final String name;
  final int quantity;
  final int unitPricePoisha;
  final List<String> selectedOptions;

  const CreateOrderItemParam({
    required this.menuItemId,
    required this.name,
    required this.quantity,
    required this.unitPricePoisha,
    this.selectedOptions = const [],
  });
}

class CreateOrderParams {
  final String restaurantId;
  final String restaurantName;
  final String addressId;
  final PaymentMethod paymentMethod;
  final List<CreateOrderItemParam> items;
  final String? idempotencyKey;

  const CreateOrderParams({
    required this.restaurantId,
    required this.restaurantName,
    required this.addressId,
    required this.paymentMethod,
    required this.items,
    this.idempotencyKey,
  });
}

abstract class OrderRepository {
  Future<Result<Order>> createOrder(CreateOrderParams params);

  Future<Result<List<Order>>> getOrders();

  Future<Result<Order>> getOrderById(String orderId);

  Future<Result<Order>> cancelOrder(String orderId, {required String reason});
}
