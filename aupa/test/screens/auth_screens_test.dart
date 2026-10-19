import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:aupa/core/theme/app_theme.dart';
import 'package:aupa/screens/auth/login_screen.dart';
import 'package:aupa/screens/auth/otp_screen.dart';
import 'package:aupa/screens/auth/signup_screen.dart';
import 'package:aupa/screens/home_screen.dart';

void main() {
  Widget createTestWidget(Widget child) {
    return MaterialApp(theme: AppTheme.lightTheme, home: child);
  }

  group('Auth Flow Screens', () {
    testWidgets('LoginScreen renders inputs and validates empty credentials', (
      tester,
    ) async {
      await tester.pumpWidget(createTestWidget(const LoginScreen()));

      expect(find.text('Login to Aupa'), findsOneWidget);
      expect(find.text('Login'), findsOneWidget);

      // Clear the default demo email
      final emailField = find.widgetWithText(TextFormField, 'Email');
      await tester.enterText(emailField, '');

      // Tap Login to trigger validation
      await tester.tap(find.widgetWithText(ElevatedButton, 'Login'));
      await tester.pumpAndSettle();

      expect(find.text('Email is required'), findsOneWidget);
    });

    testWidgets(
      'LoginScreen successfully logs in and navigates to HomeScreen',
      (tester) async {
        await tester.pumpWidget(createTestWidget(const LoginScreen()));

        final emailField = find.widgetWithText(TextFormField, 'Email');
        await tester.enterText(emailField, 'test@aupa.com');

        final passwordField = find.widgetWithText(TextFormField, 'Password');
        await tester.enterText(passwordField, 'password123');

        await tester.tap(find.widgetWithText(ElevatedButton, 'Login'));
        await tester.pump();
        await tester.pump(const Duration(milliseconds: 300));
        await tester.pumpAndSettle();

        expect(find.byType(HomeScreen), findsOneWidget);
      },
    );

    testWidgets('SignupScreen renders inputs and validates empty name', (
      tester,
    ) async {
      await tester.pumpWidget(createTestWidget(const SignupScreen()));

      expect(find.text('Create Account'), findsOneWidget);
      expect(find.text('Sign Up'), findsOneWidget);

      await tester.tap(find.widgetWithText(ElevatedButton, 'Sign Up'));
      await tester.pumpAndSettle();

      expect(find.text('Full name is required'), findsOneWidget);
    });

    testWidgets('OtpScreen renders 4 digit boxes and verify button', (
      tester,
    ) async {
      await tester.pumpWidget(
        createTestWidget(const OtpScreen(email: 'user@aupa.com')),
      );

      expect(find.text('Verify Your Number'), findsOneWidget);
      expect(find.text('Verify'), findsOneWidget);
      expect(find.byType(TextField), findsNWidgets(4));
    });
  });
}
