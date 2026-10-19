import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'package:http/http.dart' as http;
import '../config/app_config.dart';
import '../errors/failures.dart';
import 'result.dart';

class ApiClient {
  final http.Client _client;
  String? accessToken;
  String? refreshToken;

  ApiClient({http.Client? client}) : _client = client ?? http.Client();

  String get _baseUrl => AppConfig.current.apiBaseUrl;

  Future<Result<T>> get<T>(
    String path, {
    Map<String, String>? headers,
    T Function(dynamic data)? fromJson,
  }) async {
    return _send<T>('GET', path, headers: headers, fromJson: fromJson);
  }

  Future<Result<T>> post<T>(
    String path, {
    dynamic body,
    Map<String, String>? headers,
    T Function(dynamic data)? fromJson,
  }) async {
    return _send<T>(
      'POST',
      path,
      body: body,
      headers: headers,
      fromJson: fromJson,
    );
  }

  Future<Result<T>> delete<T>(
    String path, {
    Map<String, String>? headers,
    T Function(dynamic data)? fromJson,
  }) async {
    return _send<T>('DELETE', path, headers: headers, fromJson: fromJson);
  }

  Future<Result<T>> _send<T>(
    String method,
    String path, {
    dynamic body,
    Map<String, String>? headers,
    T Function(dynamic data)? fromJson,
  }) async {
    try {
      final uri = Uri.parse('$_baseUrl$path');
      final requestHeaders = {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
        'X-Request-Id': 'req_${DateTime.now().millisecondsSinceEpoch}',
        if (accessToken != null && accessToken!.isNotEmpty)
          'Authorization': 'Bearer $accessToken',
        ...?headers,
      };

      http.Response response;
      final timeout = AppConfig.current.connectTimeout;

      switch (method) {
        case 'POST':
          final payload = body != null ? jsonEncode(body) : null;
          response = await _client
              .post(uri, headers: requestHeaders, body: payload)
              .timeout(timeout);
          break;
        case 'DELETE':
          response = await _client
              .delete(uri, headers: requestHeaders)
              .timeout(timeout);
          break;
        case 'GET':
        default:
          response = await _client
              .get(uri, headers: requestHeaders)
              .timeout(timeout);
          break;
      }

      return _handleResponse<T>(response, fromJson);
    } on SocketException {
      return Result.failure(
        const NetworkFailure(
          'Unable to connect to Aupa server. Please check your internet connection.',
        ),
      );
    } on TimeoutException {
      return Result.failure(
        const NetworkFailure('Server request timed out. Please try again.'),
      );
    } catch (e) {
      return Result.failure(ServerFailure(e.toString()));
    }
  }

  Result<T> _handleResponse<T>(
    http.Response response,
    T Function(dynamic data)? fromJson,
  ) {
    dynamic decoded;
    try {
      decoded = response.body.isNotEmpty ? jsonDecode(response.body) : null;
    } catch (_) {
      decoded = null;
    }

    if (response.statusCode >= 200 && response.statusCode < 300) {
      dynamic dataPayload = decoded;
      if (decoded is Map<String, dynamic> && decoded.containsKey('data')) {
        dataPayload = decoded['data'];
      }

      if (fromJson == null) {
        return Result.success(dataPayload as T);
      }
      return Result.success(fromJson(dataPayload));
    }

    // Parse Error Envelope: { "error": { "code": "...", "message": "...", "fields": [...] } }
    String message = 'An unexpected error occurred.';
    String code = 'UNKNOWN_ERROR';
    List<FieldError> fieldErrors = [];

    if (decoded is Map<String, dynamic> &&
        decoded['error'] is Map<String, dynamic>) {
      final errMap = decoded['error'] as Map<String, dynamic>;
      message = errMap['message'] as String? ?? message;
      code = errMap['code'] as String? ?? code;

      if (errMap['fields'] is List) {
        fieldErrors = (errMap['fields'] as List)
            .map(
              (f) => FieldError(
                field: f['field'] as String? ?? '',
                code: f['code'] as String? ?? '',
                message: f['message'] as String?,
              ),
            )
            .toList();
      }
    }

    switch (response.statusCode) {
      case 400:
        return Result.failure(
          ValidationFailure(message, code: code, fieldErrors: fieldErrors),
        );
      case 401:
        return Result.failure(AuthFailure(message, code: code));
      case 403:
        return Result.failure(
          AuthFailure('Access forbidden: $message', code: 'FORBIDDEN'),
        );
      case 404:
        return Result.failure(NotFoundFailure(message, code: code));
      case 409:
        return Result.failure(ConflictFailure(message, code: code));
      default:
        return Result.failure(
          ServerFailure(message, code: code, statusCode: response.statusCode),
        );
    }
  }

  void close() {
    _client.close();
  }
}
