import 'address.dart';

enum OrderStatus {
  placed,
  accepted,
  preparing,
  ready,
  enRoute,
  delivered,
  cancelled;

  String get displayName {
    switch (this) {
      case OrderStatus.placed:
        return 'Order Placed';
      case OrderStatus.accepted:
        return 'Confirmed';
      case OrderStatus.preparing:
        return 'Preparing Food';
      case OrderStatus.ready:
        return 'Ready for Pickup';
      case OrderStatus.enRoute:
        return 'On the Way';
      case OrderStatus.delivered:
        return 'Delivered';
      case OrderStatus.cancelled:
        return 'Cancelled';
    }
  }
}

enum PaymentStatus { pending, paid, failed, notRequired }

enum PaymentMethod {
  cod('Cash on Delivery'),
  bkash('bKash'),
  card('Credit/Debit Card');

  final String displayName;
  const PaymentMethod(this.displayName);
}

class OrderAmounts {
  final int subtotalPoisha;
  final int deliveryFeePoisha;
  final int discountPoisha;
  final int taxPoisha;
  final int totalPoisha;

  const OrderAmounts({
    required this.subtotalPoisha,
    required this.deliveryFeePoisha,
    this.discountPoisha = 0,
    this.taxPoisha = 0,
    required this.totalPoisha,
  });
}

class OrderItem {
  final String menuItemId;
  final String name;
  final int quantity;
  final int unitPricePoisha;
  final List<String> selectedOptions;
  final int totalPricePoisha;

  const OrderItem({
    required this.menuItemId,
    required this.name,
    required this.quantity,
    required this.unitPricePoisha,
    this.selectedOptions = const [],
    required this.totalPricePoisha,
  });
}

class Order {
  final String id;
  final String orderNumber;
  final String restaurantId;
  final String restaurantName;
  final OrderStatus status;
  final PaymentStatus paymentStatus;
  final PaymentMethod paymentMethod;
  final OrderAmounts amounts;
  final List<OrderItem> items;
  final Address deliveryAddress;
  final DateTime createdAt;
  final DateTime? estimatedDeliveryTime;

  const Order({
    required this.id,
    required this.orderNumber,
    required this.restaurantId,
    required this.restaurantName,
    required this.status,
    required this.paymentStatus,
    required this.paymentMethod,
    required this.amounts,
    required this.items,
    required this.deliveryAddress,
    required this.createdAt,
    this.estimatedDeliveryTime,
  });

  factory Order.fromJson(Map<String, dynamic> json) {
    OrderStatus parsedStatus = OrderStatus.placed;
    switch (json['status']) {
      case 'accepted':
        parsedStatus = OrderStatus.accepted;
        break;
      case 'preparing':
        parsedStatus = OrderStatus.preparing;
        break;
      case 'ready':
        parsedStatus = OrderStatus.ready;
        break;
      case 'out_for_delivery':
      case 'enRoute':
        parsedStatus = OrderStatus.enRoute;
        break;
      case 'delivered':
        parsedStatus = OrderStatus.delivered;
        break;
      case 'cancelled':
        parsedStatus = OrderStatus.cancelled;
        break;
      default:
        parsedStatus = OrderStatus.placed;
    }

    PaymentStatus parsedPaymentStatus = PaymentStatus.pending;
    switch (json['paymentStatus']) {
      case 'paid':
        parsedPaymentStatus = PaymentStatus.paid;
        break;
      case 'failed':
        parsedPaymentStatus = PaymentStatus.failed;
        break;
      case 'not_required':
        parsedPaymentStatus = PaymentStatus.notRequired;
        break;
      default:
        parsedPaymentStatus = PaymentStatus.pending;
    }

    PaymentMethod parsedPaymentMethod = PaymentMethod.cod;
    switch (json['paymentMethod']) {
      case 'bkash':
        parsedPaymentMethod = PaymentMethod.bkash;
        break;
      case 'card':
      case 'sslcommerz':
        parsedPaymentMethod = PaymentMethod.card;
        break;
      default:
        parsedPaymentMethod = PaymentMethod.cod;
    }

    var rawItems = json['items'];
    List<OrderItem> parsedItems = [];
    if (rawItems is List) {
      parsedItems = rawItems.map((it) {
        List<String> opts = [];
        if (it['selectedOptionNames'] is List) {
          opts = (it['selectedOptionNames'] as List)
              .map((o) => o.toString())
              .toList();
        }
        return OrderItem(
          menuItemId: it['itemId'] as String? ?? it['id'] as String? ?? '',
          name: it['name'] as String? ?? '',
          quantity: it['quantity'] as int? ?? 1,
          unitPricePoisha: it['unitPricePoisha'] as int? ?? 0,
          selectedOptions: opts,
          totalPricePoisha: it['totalPoisha'] as int? ?? 0,
        );
      }).toList();
    }

    final amounts = OrderAmounts(
      subtotalPoisha: json['subtotalPoisha'] as int? ?? 0,
      deliveryFeePoisha: json['deliveryFeePoisha'] as int? ?? 0,
      discountPoisha: json['discountPoisha'] as int? ?? 0,
      taxPoisha: json['vatPoisha'] as int? ?? 0,
      totalPoisha:
          json['totalAmountPoisha'] as int? ?? json['totalPoisha'] as int? ?? 0,
    );

    return Order(
      id: json['id'] as String? ?? '',
      orderNumber:
          json['orderNumber'] as String? ??
          'AUPA-${(json['id'] as String? ?? '').toUpperCase()}',
      restaurantId: json['restaurantId'] as String? ?? '',
      restaurantName: json['restaurantName'] as String? ?? "Sultan's Dine",
      status: parsedStatus,
      paymentStatus: parsedPaymentStatus,
      paymentMethod: parsedPaymentMethod,
      amounts: amounts,
      items: parsedItems,
      deliveryAddress: const Address(
        id: 'addr_default',
        label: AddressLabel.home,
        street: 'Banani',
        area: 'Banani',
        city: 'Dhaka',
      ),
      createdAt: json['createdAt'] != null
          ? DateTime.tryParse(json['createdAt'] as String) ?? DateTime.now()
          : DateTime.now(),
      estimatedDeliveryTime: DateTime.now().add(const Duration(minutes: 30)),
    );
  }

  Order copyWith({
    String? id,
    String? orderNumber,
    String? restaurantId,
    String? restaurantName,
    OrderStatus? status,
    PaymentStatus? paymentStatus,
    PaymentMethod? paymentMethod,
    OrderAmounts? amounts,
    List<OrderItem>? items,
    Address? deliveryAddress,
    DateTime? createdAt,
    DateTime? estimatedDeliveryTime,
  }) {
    return Order(
      id: id ?? this.id,
      orderNumber: orderNumber ?? this.orderNumber,
      restaurantId: restaurantId ?? this.restaurantId,
      restaurantName: restaurantName ?? this.restaurantName,
      status: status ?? this.status,
      paymentStatus: paymentStatus ?? this.paymentStatus,
      paymentMethod: paymentMethod ?? this.paymentMethod,
      amounts: amounts ?? this.amounts,
      items: items ?? this.items,
      deliveryAddress: deliveryAddress ?? this.deliveryAddress,
      createdAt: createdAt ?? this.createdAt,
      estimatedDeliveryTime:
          estimatedDeliveryTime ?? this.estimatedDeliveryTime,
    );
  }
}
