import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:aupa/core/di/service_locator.dart';
import 'package:aupa/core/theme/app_theme.dart';
import 'package:aupa/data/mock/mock_address_repository.dart';
import 'package:aupa/data/mock/mock_auth_repository.dart';
import 'package:aupa/data/mock/mock_order_repository.dart';
import 'package:aupa/data/mock/mock_restaurant_repository.dart';
import 'package:aupa/domain/models/user.dart';
import 'package:aupa/screens/notifications/notifications_screen.dart';
import 'package:aupa/screens/profile/profile_screen.dart';
import 'package:aupa/screens/profile/saved_addresses_screen.dart';

void main() {
  Widget createWidget(Widget child) {
    return MaterialApp(theme: AppTheme.lightTheme, home: child);
  }

  setUp(() {
    ServiceLocator.instance.reset();
    ServiceLocator.instance.overrideWith(
      authRepository: MockAuthRepository(
        initialUser: User(
          id: 'test_user_1',
          fullName: 'Rahim Ahmed',
          email: 'rahim@example.com',
          phone: '+8801712345678',
          isVerified: true,
          createdAt: DateTime.now(),
        ),
      ),
      restaurantRepository: MockRestaurantRepository(),
      orderRepository: MockOrderRepository(),
      addressRepository: MockAddressRepository(),
    );
  });

  group('SavedAddressesScreen', () {
    testWidgets('renders existing addresses and allows opening add dialog', (
      tester,
    ) async {
      await tester.pumpWidget(createWidget(const SavedAddressesScreen()));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 300));
      await tester.pumpAndSettle();

      expect(find.text('Saved Addresses'), findsOneWidget);
      expect(find.text('Home'), findsOneWidget);
      expect(
        find.text('House 45, Road 7, Block C, Banani, Dhaka'),
        findsOneWidget,
      );

      final addButton = find.text('Add New Address');
      expect(addButton, findsOneWidget);

      await tester.tap(addButton);
      await tester.pumpAndSettle();

      expect(find.text('Add New Address'), findsNWidgets(2)); // Title & button
      expect(find.text('Street / Building / Flat'), findsOneWidget);
      expect(find.text('Area / Neighborhood'), findsOneWidget);
    });
  });

  group('NotificationsScreen', () {
    testWidgets('renders notifications and supports filtering and clearing', (
      tester,
    ) async {
      await tester.pumpWidget(createWidget(const NotificationsScreen()));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 100));
      await tester.pumpAndSettle();

      expect(find.text('Notifications'), findsOneWidget);
      expect(find.text('Order Delivered! 🎉'), findsOneWidget);
      expect(find.text('Flat 50% OFF Weekend! 🍔'), findsOneWidget);

      // Filter by Offers
      await tester.tap(find.text('Offers'));
      await tester.pumpAndSettle();

      expect(find.text('Flat 50% OFF Weekend! 🍔'), findsOneWidget);
      expect(find.text('Order Delivered! 🎉'), findsNothing);

      // Filter by All
      await tester.tap(find.text('All'));
      await tester.pumpAndSettle();

      // Tap Mark read
      await tester.tap(find.text('Mark read'));
      await tester.pumpAndSettle();
      expect(find.text('All notifications marked as read.'), findsOneWidget);

      // Tap Clear All
      await tester.tap(find.byIcon(Icons.delete_outline));
      await tester.pumpAndSettle();

      expect(find.text('Clear all notifications?'), findsOneWidget);
      await tester.tap(find.text('Clear'));
      await tester.pumpAndSettle();

      expect(find.text('No Notifications'), findsOneWidget);
    });
  });

  group('ProfileScreen', () {
    testWidgets('renders user profile and menu navigation options', (
      tester,
    ) async {
      await tester.pumpWidget(createWidget(const ProfileScreen()));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 300));
      await tester.pumpAndSettle();

      expect(find.text('My Profile'), findsOneWidget);
      expect(find.text('Rahim Ahmed'), findsOneWidget);
      expect(find.text('rahim@example.com'), findsOneWidget);
      expect(find.text('+8801712345678'), findsOneWidget);

      expect(find.text('Saved Delivery Addresses'), findsOneWidget);
      expect(find.text('My Orders'), findsOneWidget);
      expect(find.text('Notifications'), findsOneWidget);
      expect(find.text('Merchant & Owner Portal'), findsOneWidget);
    });

    testWidgets('support dialog displays help contact info', (tester) async {
      await tester.pumpWidget(createWidget(const ProfileScreen()));
      await tester.pump();
      await tester.pump(const Duration(milliseconds: 300));
      await tester.pumpAndSettle();

      final supportOption = find.text('Help & Customer Support');
      await tester.scrollUntilVisible(
        supportOption,
        100,
        scrollable: find.byType(Scrollable),
      );
      await tester.tap(supportOption);
      await tester.pumpAndSettle();

      expect(find.text('Aupa Support'), findsOneWidget);
      expect(find.text('+880 9612-400400'), findsOneWidget);
      expect(find.text('support@aupa.com.bd'), findsOneWidget);
    });
  });
}
