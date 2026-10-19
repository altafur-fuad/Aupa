import 'package:flutter_test/flutter_test.dart';
import 'package:aupa/main.dart';
import 'package:aupa/screens/splash_screen.dart';

void main() {
  testWidgets('AupaApp renders SplashScreen and branding', (
    WidgetTester tester,
  ) async {
    await tester.pumpWidget(const AupaApp());

    // Verify SplashScreen is loaded
    expect(find.byType(SplashScreen), findsOneWidget);

    // Verify brand title and tagline
    expect(find.text('Aupa'), findsOneWidget);
    expect(find.text('Your Food Delivered Fast'), findsOneWidget);

    // Verify navigation buttons
    expect(find.text('Enter as User (Customer)'), findsOneWidget);
    expect(find.text('Enter as Restaurant Owner'), findsOneWidget);
  });
}
