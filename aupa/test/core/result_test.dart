import 'package:flutter_test/flutter_test.dart';
import 'package:aupa/core/config/app_config.dart';
import 'package:aupa/core/errors/failures.dart';
import 'package:aupa/core/network/result.dart';

void main() {
  group('AppConfig', () {
    test('default configuration is mock environment', () {
      expect(AppConfig.current.isMock, isTrue);
      expect(AppConfig.current.currencyCode, 'BDT');
      expect(AppConfig.current.currencySymbol, '৳');
    });
  });

  group('Result', () {
    test('success result returns data and isSuccess is true', () {
      const result = Result<String>.success('hello');
      expect(result.isSuccess, isTrue);
      expect(result.isFailure, isFalse);
      expect(result.data, 'hello');
      expect(() => result.failure, throwsStateError);
    });

    test('failure result returns failure and isFailure is true', () {
      const failure = AuthFailure('Invalid credentials');
      const result = Result<String>.failure(failure);
      expect(result.isSuccess, isFalse);
      expect(result.isFailure, isTrue);
      expect(result.failure, failure);
      expect(() => result.data, throwsStateError);
    });

    test('when pattern matching executes corresponding callback', () {
      const success = Result<int>.success(42);
      final val1 = success.when(
        success: (d) => 'got $d',
        failure: (f) => 'error',
      );
      expect(val1, 'got 42');

      const failed = Result<int>.failure(NetworkFailure());
      final val2 = failed.when(
        success: (d) => 'got $d',
        failure: (f) => f.code,
      );
      expect(val2, 'NETWORK_ERROR');
    });

    test('map transforms success value and propagates failure', () {
      const success = Result<int>.success(10);
      final mappedSuccess = success.map((d) => d * 2);
      expect(mappedSuccess.data, 20);

      const failed = Result<int>.failure(ServerFailure('error'));
      final mappedFailed = failed.map((d) => d * 2);
      expect(mappedFailed.isFailure, isTrue);
    });
  });
}
