import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:aupa/core/theme/app_theme.dart';
import 'package:aupa/screens/orders/delivery_tracking_screen.dart';
import 'package:aupa/screens/orders/order_details_screen.dart';
import 'package:aupa/screens/orders/order_history_screen.dart';

void main() {
  Widget createWidget(Widget child) {
    return MaterialApp(theme: AppTheme.lightTheme, home: child);
  }

  group('Order & Delivery Tracking Screens', () {
    testWidgets('OrderHistoryScreen renders past orders', (tester) async {
      await tester.pumpWidget(createWidget(const OrderHistoryScreen()));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 300));
      await tester.pumpAndSettle();

      expect(find.text('My Orders'), findsOneWidget);
      expect(find.text('Burger Barn'), findsOneWidget);
      expect(find.text("Sultan's Dine"), findsOneWidget);
      expect(find.text('On the Way'), findsOneWidget);
      expect(find.text('Delivered'), findsOneWidget);
    });

    testWidgets('tapping an order navigates to OrderDetailsScreen', (
      tester,
    ) async {
      await tester.pumpWidget(createWidget(const OrderHistoryScreen()));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 300));
      await tester.pumpAndSettle();

      await tester.tap(find.text('Burger Barn'));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 300));
      await tester.pumpAndSettle();

      expect(find.byType(OrderDetailsScreen), findsOneWidget);
      expect(find.text('Items Ordered'), findsOneWidget);

      final trackButton = find.text('Track Delivery Live');
      await tester.scrollUntilVisible(
        trackButton,
        100,
        scrollable: find.byType(Scrollable),
      );
      expect(trackButton, findsOneWidget);
    });

    testWidgets('tapping Track Delivery Live opens DeliveryTrackingScreen', (
      tester,
    ) async {
      await tester.pumpWidget(
        createWidget(const OrderDetailsScreen(orderId: 'ord_1024')),
      );
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 300));
      await tester.pumpAndSettle();

      final trackButton = find.text('Track Delivery Live');
      await tester.scrollUntilVisible(
        trackButton,
        100,
        scrollable: find.byType(Scrollable),
      );
      await tester.pumpAndSettle();

      await tester.tap(trackButton);
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 300));
      await tester.pumpAndSettle();

      expect(find.byType(DeliveryTrackingScreen), findsOneWidget);
      expect(find.text('Live Delivery Tracking'), findsOneWidget);
      expect(find.text('Delivery Status'), findsOneWidget);

      final riderText = find.text('Tariqul Islam');
      await tester.scrollUntilVisible(
        riderText,
        100,
        scrollable: find.byType(Scrollable),
      );
      expect(riderText, findsOneWidget);
    });
  });
}
