import '../../core/network/result.dart';
import '../models/menu_item.dart';
import '../models/restaurant.dart';

abstract class RestaurantRepository {
  Future<Result<List<Restaurant>>> getRestaurants({
    String? query,
    String? category,
  });

  Future<Result<Restaurant>> getRestaurantById(String id);

  Future<Result<List<MenuItem>>> getRestaurantMenu(String restaurantId);

  Future<Result<MenuItem>> getMenuItemById(String menuItemId);
}
