import 'dart:convert';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:aupa/core/errors/failures.dart';
import 'package:aupa/core/network/api_client.dart';
import 'package:aupa/data/remote/remote_auth_repository.dart';
import 'package:aupa/data/remote/remote_order_repository.dart';
import 'package:aupa/data/remote/remote_restaurant_repository.dart';
import 'package:aupa/domain/models/order.dart';
import 'package:aupa/domain/repositories/order_repository.dart';

void main() {
  group('ApiClient Tests', () {
    test('successful 200 response unwraps data envelope', () async {
      final mockClient = MockClient((request) async {
        return http.Response(
          jsonEncode({
            'data': {'id': '123', 'name': 'Test Item'},
            'meta': {'requestId': 'req_abc'},
          }),
          200,
          headers: {'content-type': 'application/json'},
        );
      });

      final apiClient = ApiClient(client: mockClient);
      final result = await apiClient.get<Map<String, dynamic>>('/test');

      expect(result.isSuccess, isTrue);
      expect(result.data['id'], '123');
      expect(result.data['name'], 'Test Item');
    });

    test(
      '400 bad request returns ValidationFailure with field errors',
      () async {
        final mockClient = MockClient((request) async {
          return http.Response(
            jsonEncode({
              'error': {
                'code': 'VALIDATION_FAILED',
                'message': 'Invalid input data',
                'fields': [
                  {
                    'field': 'email',
                    'code': 'INVALID_EMAIL',
                    'message': 'Invalid email format',
                  },
                ],
              },
            }),
            400,
            headers: {'content-type': 'application/json'},
          );
        });

        final apiClient = ApiClient(client: mockClient);
        final result = await apiClient.post<dynamic>('/test');

        expect(result.isFailure, isTrue);
        expect(result.failure, isA<ValidationFailure>());
        final failure = result.failure as ValidationFailure;
        expect(failure.fieldErrors.length, 1);
        expect(failure.fieldErrors[0].field, 'email');
      },
    );

    test('401 unauthorized returns AuthFailure', () async {
      final mockClient = MockClient((request) async {
        return http.Response(
          jsonEncode({
            'error': {'code': 'UNAUTHORIZED', 'message': 'Token expired'},
          }),
          401,
          headers: {'content-type': 'application/json'},
        );
      });

      final apiClient = ApiClient(client: mockClient);
      final result = await apiClient.get<dynamic>('/protected');

      expect(result.isFailure, isTrue);
      expect(result.failure, isA<AuthFailure>());
      expect(result.failure.message, 'Token expired');
    });
  });

  group('RemoteAuthRepository Tests', () {
    test('login stores tokens and returns parsed User', () async {
      final mockClient = MockClient((request) async {
        expect(request.url.path, endsWith('/auth/login'));
        final body = jsonDecode(request.body);
        expect(body['email'], 'user@example.com');

        return http.Response(
          jsonEncode({
            'data': {
              'accessToken': 'jwt_access_123',
              'refreshToken': 'jwt_refresh_456',
              'user': {
                'id': 'u_01',
                'fullName': 'Sultan User',
                'email': 'user@example.com',
                'isVerified': true,
              },
            },
          }),
          200,
          headers: {'content-type': 'application/json'},
        );
      });

      final apiClient = ApiClient(client: mockClient);
      final repo = RemoteAuthRepository(apiClient);

      final result = await repo.login(
        email: 'user@example.com',
        password: 'Password123!',
      );

      expect(result.isSuccess, isTrue);
      expect(result.data.id, 'u_01');
      expect(result.data.fullName, 'Sultan User');
      expect(apiClient.accessToken, 'jwt_access_123');
      expect(apiClient.refreshToken, 'jwt_refresh_456');
    });
  });

  group('RemoteRestaurantRepository Tests', () {
    test(
      'getRestaurants sends category query and parses restaurant list',
      () async {
        final mockClient = MockClient((request) async {
          expect(request.url.path, endsWith('/restaurants'));
          expect(request.url.queryParameters['category'], 'Biryani');

          return http.Response(
            jsonEncode({
              'data': [
                {
                  'id': 'rest_01',
                  'name': "Sultan's Dine",
                  'cuisine': 'Biryani',
                  'rating': 4.9,
                  'deliveryFeePoisha': 5000,
                },
              ],
            }),
            200,
            headers: {'content-type': 'application/json'},
          );
        });

        final apiClient = ApiClient(client: mockClient);
        final repo = RemoteRestaurantRepository(apiClient);

        final result = await repo.getRestaurants(category: 'Biryani');

        expect(result.isSuccess, isTrue);
        expect(result.data.length, 1);
        expect(result.data[0].name, "Sultan's Dine");
        expect(result.data[0].deliveryFeePoisha, 5000);
      },
    );
  });

  group('RemoteOrderRepository Tests', () {
    test(
      'createOrder sends Idempotency-Key header and creates Order',
      () async {
        String? sentIdempKey;
        final mockClient = MockClient((request) async {
          expect(request.url.path, endsWith('/orders'));
          sentIdempKey = request.headers['Idempotency-Key'];

          return http.Response(
            jsonEncode({
              'data': {
                'id': 'ord_live_99',
                'orderNumber': 'AUPA-ORD-99',
                'restaurantId': 'rest_01',
                'restaurantName': "Sultan's Dine",
                'status': 'placed',
                'paymentStatus': 'pending',
                'paymentMethod': 'bkash',
                'subtotalPoisha': 48000,
                'deliveryFeePoisha': 5000,
                'totalAmountPoisha': 53000,
                'items': [
                  {
                    'itemId': 'item_01',
                    'name': 'Kacchi Biryani',
                    'quantity': 1,
                    'unitPricePoisha': 48000,
                    'totalPoisha': 48000,
                  },
                ],
              },
            }),
            201,
            headers: {'content-type': 'application/json'},
          );
        });

        final apiClient = ApiClient(client: mockClient);
        final repo = RemoteOrderRepository(apiClient);

        final result = await repo.createOrder(
          const CreateOrderParams(
            restaurantId: 'rest_01',
            restaurantName: "Sultan's Dine",
            addressId: 'addr_1',
            paymentMethod: PaymentMethod.bkash,
            items: [
              CreateOrderItemParam(
                menuItemId: 'item_01',
                name: 'Kacchi Biryani',
                quantity: 1,
                unitPricePoisha: 48000,
              ),
            ],
          ),
        );

        expect(result.isSuccess, isTrue);
        expect(sentIdempKey, isNotNull);
        expect(result.data.id, 'ord_live_99');
        expect(result.data.status, OrderStatus.placed);
        expect(result.data.amounts.totalPoisha, 53000);
      },
    );
  });
}
