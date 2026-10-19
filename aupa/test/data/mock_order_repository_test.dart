import 'package:flutter_test/flutter_test.dart';
import 'package:aupa/data/mock/mock_order_repository.dart';
import 'package:aupa/domain/models/order.dart';
import 'package:aupa/domain/repositories/order_repository.dart';

void main() {
  late MockOrderRepository repo;

  setUp(() {
    repo = MockOrderRepository();
  });

  group('MockOrderRepository', () {
    test('getOrders returns initial orders', () async {
      final res = await repo.getOrders();
      expect(res.isSuccess, isTrue);
      expect(res.data.length, 2);
    });

    test('createOrder calculates totals and prepends to orders', () async {
      final params = CreateOrderParams(
        restaurantId: 'rest_01',
        restaurantName: 'Burger Barn',
        addressId: 'addr_01',
        paymentMethod: PaymentMethod.bkash,
        items: const [
          CreateOrderItemParam(
            menuItemId: 'menu_01',
            name: 'Spicy BBQ Burger',
            quantity: 2,
            unitPricePoisha: 39000,
          ),
        ],
      );

      final res = await repo.createOrder(params);
      expect(res.isSuccess, isTrue);
      expect(res.data.amounts.subtotalPoisha, 78000);
      expect(res.data.amounts.deliveryFeePoisha, 6000);
      expect(res.data.amounts.totalPoisha, 84000);
      expect(res.data.paymentStatus, PaymentStatus.paid);

      final allOrders = await repo.getOrders();
      expect(allOrders.data.first.id, res.data.id);
    });

    test('createOrder with empty items fails validation', () async {
      const params = CreateOrderParams(
        restaurantId: 'rest_01',
        restaurantName: 'Burger Barn',
        addressId: 'addr_01',
        paymentMethod: PaymentMethod.cod,
        items: [],
      );

      final res = await repo.createOrder(params);
      expect(res.isFailure, isTrue);
      expect(res.failure.code, 'VALIDATION_ERROR');
    });

    test(
      'cancelOrder updates status to cancelled or fails if already en route',
      () async {
        // ord_1024 is currently enRoute in mock data, should conflict
        final res = await repo.cancelOrder('ord_1024', reason: 'Changed mind');
        expect(res.isFailure, isTrue);
        expect(res.failure.code, 'CONFLICT');
      },
    );
  });
}
