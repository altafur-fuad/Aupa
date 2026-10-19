import '../../core/errors/failures.dart';
import '../../core/network/result.dart';
import '../../domain/models/user.dart';
import '../../domain/repositories/auth_repository.dart';

class MockAuthRepository implements AuthRepository {
  User? _currentUser;

  MockAuthRepository({User? initialUser}) : _currentUser = initialUser;

  @override
  Future<Result<User>> login({
    required String email,
    required String password,
  }) async {
    await Future.delayed(const Duration(milliseconds: 200));

    final normalizedEmail = email.trim().toLowerCase();
    if (normalizedEmail.isEmpty || !normalizedEmail.contains('@')) {
      return const Result.failure(
        ValidationFailure(
          'Please enter a valid email address.',
          fieldErrors: [FieldError(field: 'email', code: 'INVALID_FORMAT')],
        ),
      );
    }

    if (password.length < 6) {
      return const Result.failure(
        ValidationFailure(
          'Password must be at least 6 characters.',
          fieldErrors: [FieldError(field: 'password', code: 'TOO_SHORT')],
        ),
      );
    }

    if (password == 'wrongpassword') {
      return const Result.failure(AuthFailure('Invalid email or password.'));
    }

    final user = User(
      id: 'usr_mock_123',
      fullName: normalizedEmail.startsWith('owner')
          ? 'Restaurant Manager'
          : 'Rahim Ahmed',
      email: normalizedEmail,
      phone: '+8801712345678',
      isVerified: true,
      createdAt: DateTime.now().subtract(const Duration(days: 30)),
    );

    _currentUser = user;
    return Result.success(user);
  }

  @override
  Future<Result<User>> register({
    required String fullName,
    required String email,
    required String password,
    String? phone,
  }) async {
    await Future.delayed(const Duration(milliseconds: 200));

    final trimmedName = fullName.trim();
    final normalizedEmail = email.trim().toLowerCase();

    if (trimmedName.isEmpty) {
      return const Result.failure(
        ValidationFailure(
          'Full name cannot be empty.',
          fieldErrors: [FieldError(field: 'fullName', code: 'REQUIRED')],
        ),
      );
    }

    if (!normalizedEmail.contains('@')) {
      return const Result.failure(
        ValidationFailure(
          'Please provide a valid email.',
          fieldErrors: [FieldError(field: 'email', code: 'INVALID_FORMAT')],
        ),
      );
    }

    if (password.length < 6) {
      return const Result.failure(
        ValidationFailure(
          'Password must be at least 6 characters long.',
          fieldErrors: [FieldError(field: 'password', code: 'TOO_SHORT')],
        ),
      );
    }

    final user = User(
      id: 'usr_${DateTime.now().millisecondsSinceEpoch}',
      fullName: trimmedName,
      email: normalizedEmail,
      phone: phone ?? '+8801700000000',
      isVerified: false,
      createdAt: DateTime.now(),
    );

    _currentUser = user;
    return Result.success(user);
  }

  @override
  Future<Result<bool>> verifyOtp({
    required String email,
    required String otp,
  }) async {
    await Future.delayed(const Duration(milliseconds: 200));

    if (otp == '0000' || otp.length != 4) {
      return const Result.failure(
        AuthFailure(
          'Invalid or expired OTP code. Please try again.',
          code: 'INVALID_OTP',
        ),
      );
    }

    if (_currentUser != null) {
      _currentUser = _currentUser!.copyWith(isVerified: true);
    }

    return const Result.success(true);
  }

  @override
  Future<Result<bool>> resendOtp({required String email}) async {
    await Future.delayed(const Duration(milliseconds: 150));
    return const Result.success(true);
  }

  @override
  Future<Result<bool>> requestPasswordReset({required String email}) async {
    await Future.delayed(const Duration(milliseconds: 200));
    return const Result.success(true);
  }

  @override
  Future<Result<bool>> resetPassword({
    required String token,
    required String newPassword,
  }) async {
    await Future.delayed(const Duration(milliseconds: 200));
    if (newPassword.length < 6) {
      return const Result.failure(
        ValidationFailure('New password must be at least 6 characters.'),
      );
    }
    return const Result.success(true);
  }

  @override
  Future<Result<User?>> getCurrentUser() async {
    return Result.success(_currentUser);
  }

  @override
  Future<Result<void>> logout() async {
    _currentUser = null;
    return const Result.success(null);
  }
}
