enum AppEnvironment { mock, live }

class AppConfig {
  final AppEnvironment environment;
  final String apiBaseUrl;
  final Duration connectTimeout;
  final Duration receiveTimeout;
  final String currencyCode;
  final String currencySymbol;

  const AppConfig({
    required this.environment,
    required this.apiBaseUrl,
    this.connectTimeout = const Duration(seconds: 15),
    this.receiveTimeout = const Duration(seconds: 15),
    this.currencyCode = 'BDT',
    this.currencySymbol = '৳',
  });

  bool get isMock => environment == AppEnvironment.mock;

  static AppConfig? _current;

  static AppConfig get current {
    _current ??= const AppConfig(
      environment: AppEnvironment.mock,
      apiBaseUrl: 'https://api.aupa.local/v1',
    );
    return _current!;
  }

  static void initialize(AppConfig config) {
    _current = config;
  }
}
