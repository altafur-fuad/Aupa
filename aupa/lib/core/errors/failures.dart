class FieldError {
  final String field;
  final String code;
  final String? message;

  const FieldError({required this.field, required this.code, this.message});

  @override
  String toString() =>
      'FieldError(field: $field, code: $code, message: $message)';
}

abstract class AppFailure {
  final String message;
  final String? code;

  const AppFailure(this.message, {this.code});

  @override
  String toString() => '$runtimeType(code: $code, message: $message)';
}

class ServerFailure extends AppFailure {
  final int? statusCode;

  const ServerFailure(
    super.message, {
    super.code = 'SERVER_ERROR',
    this.statusCode,
  });
}

class NetworkFailure extends AppFailure {
  const NetworkFailure([
    super.message = 'Please check your internet connection.',
    String? code,
  ]) : super(code: code ?? 'NETWORK_ERROR');
}

class AuthFailure extends AppFailure {
  const AuthFailure(super.message, {super.code = 'AUTH_ERROR'});
}

class ValidationFailure extends AppFailure {
  final List<FieldError> fieldErrors;

  const ValidationFailure(
    super.message, {
    super.code = 'VALIDATION_ERROR',
    this.fieldErrors = const [],
  });
}

class NotFoundFailure extends AppFailure {
  const NotFoundFailure(super.message, {super.code = 'NOT_FOUND'});
}

class ConflictFailure extends AppFailure {
  const ConflictFailure(super.message, {super.code = 'CONFLICT'});
}

class UnknownFailure extends AppFailure {
  const UnknownFailure([
    super.message = 'An unexpected error occurred. Please try again.',
    String? code,
  ]) : super(code: code ?? 'UNKNOWN_ERROR');
}
