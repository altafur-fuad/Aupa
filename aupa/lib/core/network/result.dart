import '../errors/failures.dart';

class Result<T> {
  final T? _data;
  final AppFailure? _failure;

  const Result.success(T data) : _data = data, _failure = null;

  const Result.failure(AppFailure failure) : _data = null, _failure = failure;

  bool get isSuccess => _failure == null;
  bool get isFailure => _failure != null;

  T get data {
    if (_failure != null) {
      throw StateError('Cannot get data from a failed Result: $_failure');
    }
    return _data as T;
  }

  AppFailure get failure {
    if (_failure == null) {
      throw StateError('Cannot get failure from a successful Result');
    }
    return _failure;
  }

  R when<R>({
    required R Function(T data) success,
    required R Function(AppFailure failure) failure,
  }) {
    if (isSuccess) {
      return success(data);
    } else {
      return failure(this.failure);
    }
  }

  Result<R> map<R>(R Function(T data) transform) {
    if (isSuccess) {
      return Result.success(transform(data));
    }
    return Result.failure(failure);
  }
}
