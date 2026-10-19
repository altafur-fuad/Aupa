import 'package:flutter_test/flutter_test.dart';
import 'package:aupa/data/mock/mock_restaurant_repository.dart';

void main() {
  late MockRestaurantRepository repo;

  setUp(() {
    repo = MockRestaurantRepository();
  });

  group('MockRestaurantRepository', () {
    test('getRestaurants returns list of restaurants', () async {
      final res = await repo.getRestaurants();
      expect(res.isSuccess, isTrue);
      expect(res.data, isNotEmpty);
      expect(res.data.any((r) => r.name == 'Burger Barn'), isTrue);
    });

    test('getRestaurants filters by query string', () async {
      final res = await repo.getRestaurants(query: 'sultan');
      expect(res.isSuccess, isTrue);
      expect(res.data.length, 1);
      expect(res.data.first.name, "Sultan's Dine");
    });

    test('getRestaurants filters by category/cuisine', () async {
      final res = await repo.getRestaurants(category: 'pizza');
      expect(res.isSuccess, isTrue);
      expect(res.data.length, 1);
      expect(res.data.first.name, 'Pizza Guy');
    });

    test('getRestaurantById returns restaurant or not found', () async {
      final found = await repo.getRestaurantById('rest_01');
      expect(found.isSuccess, isTrue);
      expect(found.data.name, 'Burger Barn');

      final notFound = await repo.getRestaurantById('unknown_id');
      expect(notFound.isFailure, isTrue);
      expect(notFound.failure.code, 'NOT_FOUND');
    });

    test('getRestaurantMenu returns menu items for restaurant', () async {
      final res = await repo.getRestaurantMenu('rest_01');
      expect(res.isSuccess, isTrue);
      expect(res.data, isNotEmpty);
      expect(res.data.first.name, 'Spicy BBQ Burger');
      expect(res.data.first.options, isNotEmpty);
    });
  });
}
