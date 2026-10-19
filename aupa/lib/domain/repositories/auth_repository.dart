import '../../core/network/result.dart';
import '../models/user.dart';

abstract class AuthRepository {
  Future<Result<User>> login({required String email, required String password});

  Future<Result<User>> register({
    required String fullName,
    required String email,
    required String password,
    String? phone,
  });

  Future<Result<bool>> verifyOtp({required String email, required String otp});

  Future<Result<bool>> resendOtp({required String email});

  Future<Result<bool>> requestPasswordReset({required String email});

  Future<Result<bool>> resetPassword({
    required String token,
    required String newPassword,
  });

  Future<Result<User?>> getCurrentUser();

  Future<Result<void>> logout();
}
