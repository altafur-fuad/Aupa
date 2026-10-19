import 'package:flutter_test/flutter_test.dart';
import 'package:aupa/data/mock/mock_auth_repository.dart';

void main() {
  late MockAuthRepository repo;

  setUp(() {
    repo = MockAuthRepository();
  });

  group('MockAuthRepository', () {
    test(
      'login with valid credentials succeeds and sets current user',
      () async {
        final res = await repo.login(
          email: 'test@aupa.com',
          password: 'password123',
        );
        expect(res.isSuccess, isTrue);
        expect(res.data.email, 'test@aupa.com');

        final currentUserRes = await repo.getCurrentUser();
        expect(currentUserRes.data, isNotNull);
        expect(currentUserRes.data!.email, 'test@aupa.com');
      },
    );

    test('login with invalid email fails validation', () async {
      final res = await repo.login(
        email: 'invalid-email',
        password: 'password123',
      );
      expect(res.isFailure, isTrue);
      expect(res.failure.code, 'VALIDATION_ERROR');
    });

    test('login with short password fails validation', () async {
      final res = await repo.login(email: 'test@aupa.com', password: '123');
      expect(res.isFailure, isTrue);
      expect(res.failure.code, 'VALIDATION_ERROR');
    });

    test('login with wrong password fails auth', () async {
      final res = await repo.login(
        email: 'test@aupa.com',
        password: 'wrongpassword',
      );
      expect(res.isFailure, isTrue);
      expect(res.failure.code, 'AUTH_ERROR');
    });

    test('register with valid details creates user', () async {
      final res = await repo.register(
        fullName: 'New User',
        email: 'new@aupa.com',
        password: 'password123',
      );
      expect(res.isSuccess, isTrue);
      expect(res.data.fullName, 'New User');
    });

    test('verifyOtp with valid 4 digit code succeeds', () async {
      await repo.login(email: 'test@aupa.com', password: 'password123');
      final res = await repo.verifyOtp(email: 'test@aupa.com', otp: '1234');
      expect(res.isSuccess, isTrue);
    });

    test('verifyOtp with 0000 fails', () async {
      final res = await repo.verifyOtp(email: 'test@aupa.com', otp: '0000');
      expect(res.isFailure, isTrue);
    });

    test('logout clears current user', () async {
      await repo.login(email: 'test@aupa.com', password: 'password123');
      await repo.logout();
      final current = await repo.getCurrentUser();
      expect(current.data, isNull);
    });
  });
}
