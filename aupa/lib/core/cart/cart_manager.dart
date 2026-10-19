import 'package:flutter/foundation.dart';

class CartItem {
  final String menuItemId;
  final String name;
  final int basePricePoisha;
  final int quantity;
  final List<String> selectedOptionNames;
  final int optionsAdditionalPoisha;
  final String? restaurantId;
  final String? restaurantName;

  const CartItem({
    required this.menuItemId,
    required this.name,
    required this.basePricePoisha,
    required this.quantity,
    this.selectedOptionNames = const [],
    this.optionsAdditionalPoisha = 0,
    this.restaurantId,
    this.restaurantName,
  });

  int get unitPricePoisha => basePricePoisha + optionsAdditionalPoisha;
  int get lineTotalPoisha => unitPricePoisha * quantity;

  CartItem copyWith({
    String? menuItemId,
    String? name,
    int? basePricePoisha,
    int? quantity,
    List<String>? selectedOptionNames,
    int? optionsAdditionalPoisha,
    String? restaurantId,
    String? restaurantName,
  }) {
    return CartItem(
      menuItemId: menuItemId ?? this.menuItemId,
      name: name ?? this.name,
      basePricePoisha: basePricePoisha ?? this.basePricePoisha,
      quantity: quantity ?? this.quantity,
      selectedOptionNames: selectedOptionNames ?? this.selectedOptionNames,
      optionsAdditionalPoisha:
          optionsAdditionalPoisha ?? this.optionsAdditionalPoisha,
      restaurantId: restaurantId ?? this.restaurantId,
      restaurantName: restaurantName ?? this.restaurantName,
    );
  }
}

class CartManager extends ChangeNotifier {
  static final CartManager _instance = CartManager._internal();
  factory CartManager() => _instance;
  CartManager._internal();

  static CartManager get instance => _instance;

  final List<CartItem> _items = [];
  String? _restaurantId;
  String? _restaurantName;
  int _deliveryFeePoisha = 6000; // default ৳60

  List<CartItem> get items => List.unmodifiable(_items);
  String? get restaurantId => _restaurantId;
  String? get restaurantName => _restaurantName;
  int get deliveryFeePoisha => _items.isEmpty ? 0 : _deliveryFeePoisha;

  int get subtotalPoisha =>
      _items.fold(0, (sum, item) => sum + item.lineTotalPoisha);

  int get totalPoisha => subtotalPoisha + deliveryFeePoisha;

  bool get isEmpty => _items.isEmpty;
  int get totalItemCount => _items.fold(0, (sum, item) => sum + item.quantity);

  void addItem({
    required String menuItemId,
    required String name,
    required int basePricePoisha,
    required int quantity,
    List<String> selectedOptions = const [],
    int optionsAdditionalPoisha = 0,
    String? restaurantId,
    String? restaurantName,
    int? deliveryFeePoisha,
  }) {
    // If cart has items from a different restaurant, reset cart for new restaurant
    if (_restaurantId != null &&
        restaurantId != null &&
        _restaurantId != restaurantId) {
      _items.clear();
    }

    if (restaurantId != null) _restaurantId = restaurantId;
    if (restaurantName != null) _restaurantName = restaurantName;
    if (deliveryFeePoisha != null) _deliveryFeePoisha = deliveryFeePoisha;

    // Check if an identical item exists (same id and same options)
    final existingIndex = _items.indexWhere(
      (item) =>
          item.menuItemId == menuItemId &&
          listEquals(item.selectedOptionNames, selectedOptions),
    );

    if (existingIndex != -1) {
      final existing = _items[existingIndex];
      _items[existingIndex] = existing.copyWith(
        quantity: existing.quantity + quantity,
      );
    } else {
      _items.add(
        CartItem(
          menuItemId: menuItemId,
          name: name,
          basePricePoisha: basePricePoisha,
          quantity: quantity,
          selectedOptionNames: selectedOptions,
          optionsAdditionalPoisha: optionsAdditionalPoisha,
          restaurantId: _restaurantId,
          restaurantName: _restaurantName,
        ),
      );
    }

    notifyListeners();
  }

  void updateQuantity(int index, int newQuantity) {
    if (index < 0 || index >= _items.length) return;
    if (newQuantity <= 0) {
      removeItem(index);
    } else {
      _items[index] = _items[index].copyWith(quantity: newQuantity);
      notifyListeners();
    }
  }

  void removeItem(int index) {
    if (index < 0 || index >= _items.length) return;
    _items.removeAt(index);
    if (_items.isEmpty) {
      _restaurantId = null;
      _restaurantName = null;
    }
    notifyListeners();
  }

  void clear() {
    _items.clear();
    _restaurantId = null;
    _restaurantName = null;
    notifyListeners();
  }
}
