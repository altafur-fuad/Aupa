import '../config/app_config.dart';
import '../network/api_client.dart';
import '../../data/mock/mock_address_repository.dart';
import '../../data/mock/mock_auth_repository.dart';
import '../../data/mock/mock_order_repository.dart';
import '../../data/mock/mock_restaurant_repository.dart';
import '../../data/remote/remote_auth_repository.dart';
import '../../data/remote/remote_order_repository.dart';
import '../../data/remote/remote_restaurant_repository.dart';
import '../../domain/repositories/address_repository.dart';
import '../../domain/repositories/auth_repository.dart';
import '../../domain/repositories/order_repository.dart';
import '../../domain/repositories/restaurant_repository.dart';

class ServiceLocator {
  static final ServiceLocator _instance = ServiceLocator._internal();
  factory ServiceLocator() => _instance;
  ServiceLocator._internal() {
    initialize(environment: AppEnvironment.mock);
  }

  static ServiceLocator get instance => _instance;

  late ApiClient _apiClient;
  late AuthRepository _authRepository;
  late RestaurantRepository _restaurantRepository;
  late OrderRepository _orderRepository;
  late AddressRepository _addressRepository;

  ApiClient get apiClient => _apiClient;
  AuthRepository get authRepository => _authRepository;
  RestaurantRepository get restaurantRepository => _restaurantRepository;
  OrderRepository get orderRepository => _orderRepository;
  AddressRepository get addressRepository => _addressRepository;

  void initialize({AppEnvironment? environment}) {
    final env = environment ?? AppConfig.current.environment;
    _apiClient = ApiClient();

    if (env == AppEnvironment.live) {
      _authRepository = RemoteAuthRepository(_apiClient);
      _restaurantRepository = RemoteRestaurantRepository(_apiClient);
      _orderRepository = RemoteOrderRepository(_apiClient);
      _addressRepository = MockAddressRepository();
    } else {
      _authRepository = MockAuthRepository();
      _restaurantRepository = MockRestaurantRepository();
      _orderRepository = MockOrderRepository();
      _addressRepository = MockAddressRepository();
    }
  }

  void overrideWith({
    AuthRepository? authRepository,
    RestaurantRepository? restaurantRepository,
    OrderRepository? orderRepository,
    AddressRepository? addressRepository,
  }) {
    if (authRepository != null) {
      _authRepository = authRepository;
    }
    if (restaurantRepository != null) {
      _restaurantRepository = restaurantRepository;
    }
    if (orderRepository != null) {
      _orderRepository = orderRepository;
    }
    if (addressRepository != null) {
      _addressRepository = addressRepository;
    }
  }

  void reset() {
    initialize(environment: AppEnvironment.mock);
  }
}
