import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:aupa/core/cart/cart_manager.dart';
import 'package:aupa/core/theme/app_theme.dart';
import 'package:aupa/domain/models/menu_item.dart';
import 'package:aupa/screens/food/food_details_screen.dart';

void main() {
  setUp(() {
    CartManager.instance.clear();
  });

  const testItem = MenuItem(
    id: 'test_m1',
    restaurantId: 'rest_01',
    name: 'Gourmet Beef Burger',
    description: 'Juicy Angus patty with artisanal cheddar.',
    pricePoisha: 40000, // ৳400
    category: 'Burgers',
    options: [
      MenuItemOption(
        id: 'opt_cheese',
        name: 'Extras',
        allowsMultiple: true,
        values: [
          OptionValue(
            id: 'v_extra_cheese',
            name: 'Extra Cheddar',
            additionalPricePoisha: 5000,
          ), // +৳50
        ],
      ),
    ],
  );

  Widget createWidget(Widget child) {
    return MaterialApp(theme: AppTheme.lightTheme, home: child);
  }

  group('FoodDetailsScreen', () {
    testWidgets('renders item details and initial price', (tester) async {
      await tester.pumpWidget(
        createWidget(const FoodDetailsScreen(menuItem: testItem)),
      );
      await tester.pumpAndSettle();

      expect(find.text('Gourmet Beef Burger'), findsOneWidget);
      expect(
        find.text('Juicy Angus patty with artisanal cheddar.'),
        findsOneWidget,
      );
      expect(find.text('Add to Cart - ৳400'), findsOneWidget);
    });

    testWidgets('selecting option updates price dynamically', (tester) async {
      await tester.pumpWidget(
        createWidget(const FoodDetailsScreen(menuItem: testItem)),
      );
      await tester.pumpAndSettle();

      // Check the Extra Cheddar option
      await tester.tap(find.text('Extra Cheddar'));
      await tester.pumpAndSettle();

      // Price should now be ৳450
      expect(find.text('Add to Cart - ৳450'), findsOneWidget);
    });

    testWidgets('increasing quantity updates total button price', (
      tester,
    ) async {
      await tester.pumpWidget(
        createWidget(const FoodDetailsScreen(menuItem: testItem)),
      );
      await tester.pumpAndSettle();

      // Tap + button
      await tester.tap(find.byIcon(Icons.add));
      await tester.pumpAndSettle();

      // 400 * 2 = 800
      expect(find.text('Add to Cart - ৳800'), findsOneWidget);
    });

    testWidgets('tapping Add to Cart adds item to CartManager', (tester) async {
      await tester.pumpWidget(
        createWidget(const FoodDetailsScreen(menuItem: testItem)),
      );
      await tester.pumpAndSettle();

      await tester.tap(find.text('Add to Cart - ৳400'));
      await tester.pump();

      expect(CartManager.instance.items.length, 1);
      expect(CartManager.instance.items.first.name, 'Gourmet Beef Burger');
      expect(CartManager.instance.items.first.quantity, 1);
    });
  });
}
