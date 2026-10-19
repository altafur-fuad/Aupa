import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:aupa/core/theme/app_theme.dart';
import 'package:aupa/screens/home_screen.dart';
import 'package:aupa/screens/restaurant/restaurant_details_screen.dart';

void main() {
  Widget createHomeScreen() {
    return MaterialApp(theme: AppTheme.lightTheme, home: const HomeScreen());
  }

  Future<void> pumpUntilLoaded(WidgetTester tester) async {
    await tester.pump();
    await tester.pump(const Duration(milliseconds: 300));
    await tester.pumpAndSettle();
  }

  group('HomeScreen Discovery & Search', () {
    testWidgets(
      'renders location, search bar, categories and loads restaurant cards',
      (tester) async {
        await tester.pumpWidget(createHomeScreen());
        await pumpUntilLoaded(tester);

        // Location header
        expect(find.text('Banani, Dhaka'), findsOneWidget);

        // Search field
        expect(find.text('Search for food, restaurants...'), findsOneWidget);

        // Categories
        expect(find.text('All'), findsOneWidget);
        expect(find.text('Burgers'), findsOneWidget);
        expect(find.text('Biryani'), findsOneWidget);

        // Loaded restaurants
        expect(find.text('Burger Barn'), findsOneWidget);
        expect(find.text("Sultan's Dine"), findsOneWidget);
      },
    );

    testWidgets('filtering by category displays matching restaurants', (
      tester,
    ) async {
      await tester.pumpWidget(createHomeScreen());
      await pumpUntilLoaded(tester);

      // Tap on Biryani category
      await tester.tap(find.text('Biryani'));
      await pumpUntilLoaded(tester);

      expect(find.text("Sultan's Dine"), findsOneWidget);
      expect(find.text('Burger Barn'), findsNothing);
    });

    testWidgets('typing in search bar filters restaurants dynamically', (
      tester,
    ) async {
      await tester.pumpWidget(createHomeScreen());
      await pumpUntilLoaded(tester);

      final searchField = find.byType(TextField).first;
      await tester.enterText(searchField, 'Pizza');
      await pumpUntilLoaded(tester);

      expect(find.text('Pizza Guy'), findsOneWidget);
      expect(find.text('Burger Barn'), findsNothing);
    });

    testWidgets(
      'tapping a restaurant card navigates to RestaurantDetailsScreen',
      (tester) async {
        await tester.pumpWidget(createHomeScreen());
        await pumpUntilLoaded(tester);

        final cardFinder = find.text('Burger Barn');
        expect(cardFinder, findsOneWidget);

        await tester.ensureVisible(cardFinder);
        await tester.pumpAndSettle();

        await tester.tap(cardFinder);
        await pumpUntilLoaded(tester);

        expect(find.byType(RestaurantDetailsScreen), findsOneWidget);
      },
    );

    testWidgets('tapping bottom navigation tabs switches screens', (
      tester,
    ) async {
      await tester.pumpWidget(createHomeScreen());
      await pumpUntilLoaded(tester);

      // Tap on 'Search' tab in bottom nav
      await tester.tap(find.text('Search'));
      await pumpUntilLoaded(tester);

      expect(find.text('Search & Explore'), findsOneWidget);
      expect(find.text('Popular Searches'), findsOneWidget);

      // Tap on 'Orders' tab in bottom nav
      await tester.tap(find.text('Orders'));
      await pumpUntilLoaded(tester);

      expect(find.text('My Orders'), findsOneWidget);

      // Tap on 'Profile' tab in bottom nav
      await tester.tap(find.text('Profile'));
      await pumpUntilLoaded(tester);

      expect(find.text('My Profile'), findsOneWidget);

      // Tap back on 'Home' tab
      await tester.tap(find.text('Home'));
      await pumpUntilLoaded(tester);

      expect(find.text('Popular Near You'), findsOneWidget);
    });

    testWidgets('tapping notification icon navigates to NotificationsScreen', (
      tester,
    ) async {
      await tester.pumpWidget(createHomeScreen());
      await pumpUntilLoaded(tester);

      await tester.tap(find.byIcon(Icons.notifications_outlined));
      await pumpUntilLoaded(tester);

      expect(find.text('Notifications'), findsOneWidget);
    });

    testWidgets('tapping location header navigates to SavedAddressesScreen', (
      tester,
    ) async {
      await tester.pumpWidget(createHomeScreen());
      await pumpUntilLoaded(tester);

      await tester.tap(find.text('Banani, Dhaka'));
      await pumpUntilLoaded(tester);

      expect(find.text('Saved Addresses'), findsOneWidget);
    });
  });
}
