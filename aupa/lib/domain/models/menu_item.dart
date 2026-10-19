class OptionValue {
  final String id;
  final String name;
  final int additionalPricePoisha;
  final bool isDefault;

  const OptionValue({
    required this.id,
    required this.name,
    this.additionalPricePoisha = 0,
    this.isDefault = false,
  });
}

class MenuItemOption {
  final String id;
  final String name;
  final bool isRequired;
  final bool allowsMultiple;
  final List<OptionValue> values;

  const MenuItemOption({
    required this.id,
    required this.name,
    this.isRequired = false,
    this.allowsMultiple = false,
    required this.values,
  });
}

class MenuItem {
  final String id;
  final String restaurantId;
  final String name;
  final String description;
  final int pricePoisha;
  final String category;
  final bool isAvailable;
  final String? imageUrl;
  final List<MenuItemOption> options;

  const MenuItem({
    required this.id,
    required this.restaurantId,
    required this.name,
    required this.description,
    required this.pricePoisha,
    required this.category,
    this.isAvailable = true,
    this.imageUrl,
    this.options = const [],
  });

  factory MenuItem.fromJson(Map<String, dynamic> json) {
    var rawOptions = json['optionGroups'] ?? json['options'];
    List<MenuItemOption> parsedOptions = [];
    if (rawOptions is List) {
      parsedOptions = rawOptions.map((og) {
        var rawValues = og['items'] ?? og['values'];
        List<OptionValue> parsedValues = [];
        if (rawValues is List) {
          parsedValues = rawValues
              .map(
                (v) => OptionValue(
                  id: v['id'] as String? ?? '',
                  name: v['name'] as String? ?? '',
                  additionalPricePoisha:
                      v['priceModifierPoisha'] as int? ??
                      v['priceDeltaPoisha'] as int? ??
                      0,
                  isDefault: v['isDefault'] as bool? ?? false,
                ),
              )
              .toList();
        }
        return MenuItemOption(
          id: og['id'] as String? ?? '',
          name: og['name'] as String? ?? '',
          isRequired: og['isRequired'] as bool? ?? false,
          allowsMultiple: (og['maxSelection'] as int? ?? 1) > 1,
          values: parsedValues,
        );
      }).toList();
    }

    return MenuItem(
      id: json['id'] as String? ?? '',
      restaurantId: json['restaurantId'] as String? ?? '',
      name: json['name'] as String? ?? '',
      description: json['description'] as String? ?? '',
      pricePoisha: json['pricePoisha'] as int? ?? 0,
      category: json['category'] as String? ?? 'General',
      isAvailable: json['isAvailable'] as bool? ?? true,
      imageUrl: json['imageUrl'] as String?,
      options: parsedOptions,
    );
  }

  MenuItem copyWith({
    String? id,
    String? restaurantId,
    String? name,
    String? description,
    int? pricePoisha,
    String? category,
    bool? isAvailable,
    String? imageUrl,
    List<MenuItemOption>? options,
  }) {
    return MenuItem(
      id: id ?? this.id,
      restaurantId: restaurantId ?? this.restaurantId,
      name: name ?? this.name,
      description: description ?? this.description,
      pricePoisha: pricePoisha ?? this.pricePoisha,
      category: category ?? this.category,
      isAvailable: isAvailable ?? this.isAvailable,
      imageUrl: imageUrl ?? this.imageUrl,
      options: options ?? this.options,
    );
  }
}
