import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:aupa/core/widgets/app_button.dart';
import 'package:aupa/core/widgets/empty_state_view.dart';
import 'package:aupa/core/widgets/error_state_view.dart';
import 'package:aupa/core/widgets/loading_indicator.dart';

void main() {
  Widget wrapWithMaterial(Widget child) {
    return MaterialApp(home: Scaffold(body: child));
  }

  group('Shared Widgets', () {
    testWidgets('AppButton renders label and responds to tap', (tester) async {
      bool tapped = false;
      await tester.pumpWidget(
        wrapWithMaterial(
          AppButton(label: 'Submit Order', onPressed: () => tapped = true),
        ),
      );

      expect(find.text('Submit Order'), findsOneWidget);
      await tester.tap(find.byType(ElevatedButton));
      expect(tapped, isTrue);
    });

    testWidgets('AppButton displays loading spinner when isLoading is true', (
      tester,
    ) async {
      bool tapped = false;
      await tester.pumpWidget(
        wrapWithMaterial(
          AppButton(
            label: 'Submit Order',
            isLoading: true,
            onPressed: () => tapped = true,
          ),
        ),
      );

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(find.text('Submit Order'), findsNothing);
      await tester.tap(find.byType(ElevatedButton));
      expect(tapped, isFalse);
    });

    testWidgets('EmptyStateView displays title, message, and action button', (
      tester,
    ) async {
      bool actionTriggered = false;
      await tester.pumpWidget(
        wrapWithMaterial(
          EmptyStateView(
            title: 'Cart is empty',
            message: 'Explore our top restaurants and add delicious meals.',
            actionLabel: 'Browse Restaurants',
            onAction: () => actionTriggered = true,
          ),
        ),
      );

      expect(find.text('Cart is empty'), findsOneWidget);
      expect(
        find.text('Explore our top restaurants and add delicious meals.'),
        findsOneWidget,
      );
      expect(find.text('Browse Restaurants'), findsOneWidget);

      await tester.tap(find.text('Browse Restaurants'));
      expect(actionTriggered, isTrue);
    });

    testWidgets('ErrorStateView displays message, code, and retry button', (
      tester,
    ) async {
      bool retried = false;
      await tester.pumpWidget(
        wrapWithMaterial(
          ErrorStateView(
            message: 'Unable to connect to server.',
            errorCode: 'NETWORK_ERROR',
            onRetry: () => retried = true,
          ),
        ),
      );

      expect(find.text('Unable to connect to server.'), findsOneWidget);
      expect(find.text('Code: NETWORK_ERROR'), findsOneWidget);
      expect(find.text('Try Again'), findsOneWidget);

      await tester.tap(find.text('Try Again'));
      expect(retried, isTrue);
    });

    testWidgets('AppLoadingIndicator renders spinner and optional message', (
      tester,
    ) async {
      await tester.pumpWidget(
        wrapWithMaterial(
          const AppLoadingIndicator(message: 'Preparing your menu...'),
        ),
      );

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      expect(find.text('Preparing your menu...'), findsOneWidget);
    });
  });
}
