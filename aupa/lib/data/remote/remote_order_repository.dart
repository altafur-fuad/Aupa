import 'dart:math';
import '../../core/network/api_client.dart';
import '../../core/network/result.dart';
import '../../domain/models/order.dart';
import '../../domain/repositories/order_repository.dart';

class RemoteOrderRepository implements OrderRepository {
  final ApiClient _apiClient;

  RemoteOrderRepository(this._apiClient);

  @override
  Future<Result<Order>> createOrder(CreateOrderParams params) async {
    final idempotencyKey =
        params.idempotencyKey ??
        'idemp_${DateTime.now().millisecondsSinceEpoch}_${Random().nextInt(999999)}';

    String apiPaymentMethod = 'cash_on_delivery';
    switch (params.paymentMethod) {
      case PaymentMethod.bkash:
        apiPaymentMethod = 'bkash';
        break;
      case PaymentMethod.card:
        apiPaymentMethod = 'sslcommerz';
        break;
      case PaymentMethod.cod:
        apiPaymentMethod = 'cash_on_delivery';
        break;
    }

    final body = {
      'restaurantId': params.restaurantId,
      'deliveryAddressId': params.addressId,
      'paymentMethod': apiPaymentMethod,
      'items': params.items
          .map(
            (i) => {
              'itemId': i.menuItemId,
              'quantity': i.quantity,
              'optionItemIds': i.selectedOptions,
            },
          )
          .toList(),
    };

    final result = await _apiClient.post<Map<String, dynamic>>(
      '/orders',
      body: body,
      headers: {'Idempotency-Key': idempotencyKey},
    );

    return result.map((data) => Order.fromJson(data));
  }

  @override
  Future<Result<List<Order>>> getOrders() async {
    final result = await _apiClient.get<dynamic>('/orders');

    return result.map((data) {
      if (data is List) {
        return data
            .map((o) => Order.fromJson(o as Map<String, dynamic>))
            .toList();
      }
      return <Order>[];
    });
  }

  @override
  Future<Result<Order>> getOrderById(String orderId) async {
    final result = await _apiClient.get<Map<String, dynamic>>(
      '/orders/$orderId',
    );
    return result.map((data) => Order.fromJson(data));
  }

  @override
  Future<Result<Order>> cancelOrder(
    String orderId, {
    required String reason,
  }) async {
    final result = await _apiClient.post<Map<String, dynamic>>(
      '/orders/$orderId/cancel',
      body: {'reason': reason},
    );
    return result.map((data) => Order.fromJson(data));
  }
}
