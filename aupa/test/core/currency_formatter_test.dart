import 'package:flutter_test/flutter_test.dart';
import 'package:aupa/core/utils/currency_formatter.dart';

void main() {
  group('CurrencyFormatter', () {
    test('poishaToBdt converts poisha minor units correctly', () {
      expect(CurrencyFormatter.poishaToBdt(100), 1.0);
      expect(CurrencyFormatter.poishaToBdt(25000), 250.0);
      expect(CurrencyFormatter.poishaToBdt(99), 0.99);
    });

    test('bdtToPoisha converts BDT to integer poisha', () {
      expect(CurrencyFormatter.bdtToPoisha(1.0), 100);
      expect(CurrencyFormatter.bdtToPoisha(250.0), 25000);
      expect(CurrencyFormatter.bdtToPoisha(0.99), 99);
    });

    test(
      'formatPoisha formats integer BDT without decimals when showDecimals is false',
      () {
        expect(CurrencyFormatter.formatPoisha(25000), '৳250');
        expect(
          CurrencyFormatter.formatPoisha(25000, includeSymbol: false),
          '250',
        );
      },
    );

    test(
      'formatPoisha includes decimals when fractional or showDecimals is true',
      () {
        expect(CurrencyFormatter.formatPoisha(25050), '৳250.50');
        expect(
          CurrencyFormatter.formatPoisha(25000, showDecimals: true),
          '৳250.00',
        );
      },
    );

    test('formatBdt delegates correctly', () {
      expect(CurrencyFormatter.formatBdt(150.0), '৳150');
      expect(CurrencyFormatter.formatBdt(150.50), '৳150.50');
    });
  });
}
