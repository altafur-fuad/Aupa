import '../config/app_config.dart';

class CurrencyFormatter {
  static const int poishaPerBdt = 100;

  /// Converts poisha (integer minor units) to BDT (double).
  static double poishaToBdt(int poisha) {
    return poisha / poishaPerBdt;
  }

  /// Converts BDT (double) to poisha (integer minor units).
  static int bdtToPoisha(double bdt) {
    return (bdt * poishaPerBdt).round();
  }

  /// Formats an amount given in poisha into a display string.
  /// Example: 25000 poisha -> "৳250" or "৳250.00"
  static String formatPoisha(
    int poisha, {
    bool includeSymbol = true,
    bool showDecimals = false,
    String? symbol,
  }) {
    final effectiveSymbol = symbol ?? AppConfig.current.currencySymbol;
    final bdt = poishaToBdt(poisha);
    final formattedNumber = showDecimals
        ? bdt.toStringAsFixed(2)
        : (bdt % 1 == 0 ? bdt.toInt().toString() : bdt.toStringAsFixed(2));

    if (includeSymbol) {
      return '$effectiveSymbol$formattedNumber';
    }
    return formattedNumber;
  }

  /// Formats an amount given in standard BDT currency.
  static String formatBdt(
    double amount, {
    bool includeSymbol = true,
    bool showDecimals = false,
    String? symbol,
  }) {
    return formatPoisha(
      bdtToPoisha(amount),
      includeSymbol: includeSymbol,
      showDecimals: showDecimals,
      symbol: symbol,
    );
  }
}
