import '../../core/errors/failures.dart';
import '../../core/network/api_client.dart';
import '../../core/network/result.dart';
import '../../domain/models/menu_item.dart';
import '../../domain/models/restaurant.dart';
import '../../domain/repositories/restaurant_repository.dart';

class RemoteRestaurantRepository implements RestaurantRepository {
  final ApiClient _apiClient;

  RemoteRestaurantRepository(this._apiClient);

  @override
  Future<Result<List<Restaurant>>> getRestaurants({
    String? query,
    String? category,
  }) async {
    final queryParams = <String, String>{};
    if (query != null && query.isNotEmpty) queryParams['q'] = query;
    if (category != null && category.isNotEmpty && category != 'All') {
      queryParams['category'] = category;
    }

    final queryString = queryParams.isNotEmpty
        ? '?${queryParams.entries.map((e) => '${Uri.encodeComponent(e.key)}=${Uri.encodeComponent(e.value)}').join('&')}'
        : '';

    final result = await _apiClient.get<dynamic>('/restaurants$queryString');

    return result.map((data) {
      if (data is List) {
        return data
            .map((r) => Restaurant.fromJson(r as Map<String, dynamic>))
            .toList();
      }
      return <Restaurant>[];
    });
  }

  @override
  Future<Result<Restaurant>> getRestaurantById(String id) async {
    final result = await _apiClient.get<Map<String, dynamic>>(
      '/restaurants/$id',
    );
    return result.map((data) => Restaurant.fromJson(data));
  }

  @override
  Future<Result<List<MenuItem>>> getRestaurantMenu(String restaurantId) async {
    final result = await _apiClient.get<dynamic>(
      '/restaurants/$restaurantId/menu',
    );

    return result.map((data) {
      if (data is List) {
        return data
            .map((m) => MenuItem.fromJson(m as Map<String, dynamic>))
            .toList();
      }
      return <MenuItem>[];
    });
  }

  @override
  Future<Result<MenuItem>> getMenuItemById(String menuItemId) async {
    // Falls back to menu search
    return Result.failure(
      ServerFailure('Method not directly supported on remote API'),
    );
  }
}
