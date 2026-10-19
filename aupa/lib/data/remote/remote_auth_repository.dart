import '../../core/network/api_client.dart';
import '../../core/network/result.dart';
import '../../domain/models/user.dart';
import '../../domain/repositories/auth_repository.dart';

class RemoteAuthRepository implements AuthRepository {
  final ApiClient _apiClient;
  User? _cachedUser;

  RemoteAuthRepository(this._apiClient);

  @override
  Future<Result<User>> login({
    required String email,
    required String password,
  }) async {
    final result = await _apiClient.post<Map<String, dynamic>>(
      '/auth/login',
      body: {'email': email, 'password': password},
    );

    return result.map((data) {
      final user = User.fromJson(data['user'] as Map<String, dynamic>);
      _cachedUser = user;
      _apiClient.accessToken = data['accessToken'] as String?;
      _apiClient.refreshToken = data['refreshToken'] as String?;
      return user;
    });
  }

  @override
  Future<Result<User>> register({
    required String fullName,
    required String email,
    required String password,
    String? phone,
  }) async {
    final result = await _apiClient.post<Map<String, dynamic>>(
      '/auth/register',
      body: {
        'fullName': fullName,
        'email': email,
        'password': password,
        if (phone != null) 'phone': phone,
      },
    );

    return result.map((data) {
      final user = User.fromJson(data['user'] as Map<String, dynamic>);
      _cachedUser = user;
      _apiClient.accessToken = data['accessToken'] as String?;
      _apiClient.refreshToken = data['refreshToken'] as String?;
      return user;
    });
  }

  @override
  Future<Result<bool>> verifyOtp({
    required String email,
    required String otp,
  }) async {
    final result = await _apiClient.post<Map<String, dynamic>>(
      '/auth/verify-contact',
      body: {'email': email, 'otp': otp},
    );
    return result.map((data) => data['verified'] as bool? ?? true);
  }

  @override
  Future<Result<bool>> resendOtp({required String email}) async {
    // API mock returns success
    return Result.success(true);
  }

  @override
  Future<Result<bool>> requestPasswordReset({required String email}) async {
    return Result.success(true);
  }

  @override
  Future<Result<bool>> resetPassword({
    required String token,
    required String newPassword,
  }) async {
    return Result.success(true);
  }

  @override
  Future<Result<User?>> getCurrentUser() async {
    if (_cachedUser != null) {
      return Result.success(_cachedUser);
    }
    if (_apiClient.accessToken == null) {
      return Result.success(null);
    }

    final result = await _apiClient.get<Map<String, dynamic>>('/me');
    return result.map((data) {
      final user = User.fromJson(data);
      _cachedUser = user;
      return user;
    });
  }

  @override
  Future<Result<void>> logout() async {
    if (_apiClient.refreshToken != null) {
      await _apiClient.post<void>(
        '/auth/logout',
        body: {'refreshToken': _apiClient.refreshToken},
      );
    }
    _cachedUser = null;
    _apiClient.accessToken = null;
    _apiClient.refreshToken = null;
    return Result.success(null);
  }
}
