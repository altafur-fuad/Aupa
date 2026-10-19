import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:aupa/core/cart/cart_manager.dart';
import 'package:aupa/core/theme/app_theme.dart';
import 'package:aupa/screens/cart/cart_screen.dart';
import 'package:aupa/screens/cart/checkout_screen.dart';

void main() {
  setUp(() {
    CartManager.instance.clear();
  });

  Widget createWidget(Widget child) {
    return MaterialApp(theme: AppTheme.lightTheme, home: child);
  }

  group('Cart & Checkout Flow', () {
    testWidgets('CartScreen shows EmptyStateView when cart is empty', (
      tester,
    ) async {
      await tester.pumpWidget(createWidget(const CartScreen()));
      await tester.pumpAndSettle();

      expect(find.text('Your cart is empty'), findsOneWidget);
      expect(find.text('Find Food'), findsOneWidget);
    });

    testWidgets('CartScreen renders items and applies promo code', (
      tester,
    ) async {
      CartManager.instance.addItem(
        menuItemId: 'menu_01',
        name: 'Spicy BBQ Burger',
        basePricePoisha: 39000,
        quantity: 1,
        restaurantName: 'Burger Barn',
        deliveryFeePoisha: 6000,
      );

      await tester.pumpWidget(createWidget(const CartScreen()));
      await tester.pumpAndSettle();

      expect(find.text('Spicy BBQ Burger'), findsOneWidget);
      expect(find.text('Burger Barn'), findsOneWidget);

      // Enter promo code
      final promoField = find.byType(TextField);
      await tester.enterText(promoField, 'AUPA50');
      await tester.tap(find.text('Apply'));
      await tester.pumpAndSettle();

      expect(find.text('Promo Discount'), findsOneWidget);
      expect(find.text('-৳50'), findsOneWidget);
    });

    testWidgets(
      'CheckoutScreen displays addresses, payment methods, and places order',
      (tester) async {
        CartManager.instance.addItem(
          menuItemId: 'menu_01',
          name: 'Spicy BBQ Burger',
          basePricePoisha: 39000,
          quantity: 1,
          restaurantName: 'Burger Barn',
          deliveryFeePoisha: 6000,
        );

        await tester.pumpWidget(createWidget(const CheckoutScreen()));
        await tester.pump();
        await tester.pump(const Duration(milliseconds: 300));
        await tester.pumpAndSettle();

        expect(find.text('Delivery Address'), findsOneWidget);
        expect(find.text('Payment Method'), findsOneWidget);
        expect(find.text('Cash on Delivery'), findsOneWidget);
        expect(find.text('bKash Online'), findsOneWidget);

        // Tap Place Order button
        final placeOrderButton = find.widgetWithText(
          ElevatedButton,
          'Place Order - ৳450',
        );
        expect(placeOrderButton, findsOneWidget);

        await tester.tap(placeOrderButton);
        await tester.pump();
        await tester.pump(const Duration(milliseconds: 400));
        await tester.pumpAndSettle();

        expect(find.text('Order Placed!'), findsOneWidget);
        expect(find.text('Back to Home'), findsOneWidget);
        expect(CartManager.instance.isEmpty, isTrue);
      },
    );
  });
}
