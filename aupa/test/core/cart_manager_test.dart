import 'package:flutter_test/flutter_test.dart';
import 'package:aupa/core/cart/cart_manager.dart';

void main() {
  late CartManager cart;

  setUp(() {
    cart = CartManager.instance;
    cart.clear();
  });

  group('CartManager', () {
    test('starts empty', () {
      expect(cart.isEmpty, isTrue);
      expect(cart.items, isEmpty);
      expect(cart.subtotalPoisha, 0);
      expect(cart.totalPoisha, 0);
    });

    test('addItem calculates subtotal and delivery fee correctly', () {
      cart.addItem(
        menuItemId: 'menu_01',
        name: 'Spicy BBQ Burger',
        basePricePoisha: 39000,
        quantity: 2,
        optionsAdditionalPoisha: 5000,
        restaurantId: 'rest_01',
        restaurantName: 'Burger Barn',
        deliveryFeePoisha: 6000,
      );

      expect(cart.items.length, 1);
      // Unit price = 39000 + 5000 = 44000
      // Line total = 44000 * 2 = 88000
      expect(cart.subtotalPoisha, 88000);
      expect(cart.deliveryFeePoisha, 6000);
      expect(cart.totalPoisha, 94000);
    });

    test('adding same item merges quantities', () {
      cart.addItem(
        menuItemId: 'menu_01',
        name: 'Spicy BBQ Burger',
        basePricePoisha: 39000,
        quantity: 1,
        selectedOptions: ['Mild'],
        restaurantId: 'rest_01',
      );

      cart.addItem(
        menuItemId: 'menu_01',
        name: 'Spicy BBQ Burger',
        basePricePoisha: 39000,
        quantity: 2,
        selectedOptions: ['Mild'],
        restaurantId: 'rest_01',
      );

      expect(cart.items.length, 1);
      expect(cart.items.first.quantity, 3);
    });

    test(
      'adding item from different restaurant resets cart to new restaurant',
      () {
        cart.addItem(
          menuItemId: 'menu_01',
          name: 'Burger',
          basePricePoisha: 30000,
          quantity: 1,
          restaurantId: 'rest_01',
        );

        expect(cart.items.length, 1);

        cart.addItem(
          menuItemId: 'menu_05',
          name: 'Biryani',
          basePricePoisha: 45000,
          quantity: 1,
          restaurantId: 'rest_02',
        );

        expect(cart.items.length, 1);
        expect(cart.restaurantId, 'rest_02');
        expect(cart.items.first.name, 'Biryani');
      },
    );

    test('updateQuantity modifies quantity or removes when 0', () {
      cart.addItem(
        menuItemId: 'menu_01',
        name: 'Burger',
        basePricePoisha: 30000,
        quantity: 1,
      );

      cart.updateQuantity(0, 5);
      expect(cart.items.first.quantity, 5);

      cart.updateQuantity(0, 0);
      expect(cart.isEmpty, isTrue);
    });
  });
}
