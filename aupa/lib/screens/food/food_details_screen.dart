import 'package:flutter/material.dart';
import '../../constants/colors.dart';
import '../../core/cart/cart_manager.dart';
import '../../core/utils/currency_formatter.dart';
import '../../domain/models/menu_item.dart';

class FoodDetailsScreen extends StatefulWidget {
  final MenuItem? menuItem;
  final String? restaurantName;
  final String? restaurantId;

  const FoodDetailsScreen({
    super.key,
    this.menuItem,
    this.restaurantName,
    this.restaurantId,
  });

  @override
  State<FoodDetailsScreen> createState() => _FoodDetailsScreenState();
}

class _FoodDetailsScreenState extends State<FoodDetailsScreen> {
  late final MenuItem _item;
  int _quantity = 1;

  // Track single-choice selections: optionId -> optionValueId
  final Map<String, String> _singleSelections = {};

  // Track multi-choice selections: optionId -> Set<optionValueId>
  final Map<String, Set<String>> _multiSelections = {};

  @override
  void initState() {
    super.initState();
    _item =
        widget.menuItem ??
        const MenuItem(
          id: 'menu_01',
          restaurantId: 'rest_01',
          name: 'Spicy BBQ Burger',
          description:
              'Spicy beef patty, jalapenos, BBQ sauce, crispy onions, melted cheese on a toasted brioche bun.',
          pricePoisha: 39000,
          category: 'Burgers',
          options: [
            MenuItemOption(
              id: 'opt_spicy',
              name: 'Spice Level',
              isRequired: true,
              allowsMultiple: false,
              values: [
                OptionValue(id: 'v_mild', name: 'Mild', isDefault: true),
                OptionValue(id: 'v_med', name: 'Medium'),
                OptionValue(
                  id: 'v_hot',
                  name: 'Extra Hot (+৳20)',
                  additionalPricePoisha: 2000,
                ),
              ],
            ),
            MenuItemOption(
              id: 'opt_custom',
              name: 'Customize Your Order',
              isRequired: false,
              allowsMultiple: true,
              values: [
                OptionValue(
                  id: 'v_cheese',
                  name: 'Extra Cheese',
                  additionalPricePoisha: 5000,
                ),
                OptionValue(
                  id: 'v_bacon',
                  name: 'Add Bacon',
                  additionalPricePoisha: 8000,
                ),
                OptionValue(
                  id: 'v_onion',
                  name: 'No Onions',
                  additionalPricePoisha: 0,
                ),
              ],
            ),
          ],
        );

    // Initialize defaults for options
    for (final opt in _item.options) {
      if (!opt.allowsMultiple) {
        final defVal =
            opt.values.where((v) => v.isDefault).firstOrNull ??
            opt.values.firstOrNull;
        if (defVal != null) {
          _singleSelections[opt.id] = defVal.id;
        }
      } else {
        _multiSelections[opt.id] = {};
      }
    }
  }

  int get _optionsAdditionalPricePoisha {
    int total = 0;
    for (final opt in _item.options) {
      if (!opt.allowsMultiple) {
        final selectedValId = _singleSelections[opt.id];
        final val = opt.values.where((v) => v.id == selectedValId).firstOrNull;
        if (val != null) total += val.additionalPricePoisha;
      } else {
        final selectedSet = _multiSelections[opt.id] ?? {};
        for (final val in opt.values) {
          if (selectedSet.contains(val.id)) {
            total += val.additionalPricePoisha;
          }
        }
      }
    }
    return total;
  }

  int get _unitPricePoisha => _item.pricePoisha + _optionsAdditionalPricePoisha;

  int get _totalPricePoisha => _unitPricePoisha * _quantity;

  List<String> get _selectedOptionLabels {
    final labels = <String>[];
    for (final opt in _item.options) {
      if (!opt.allowsMultiple) {
        final selectedValId = _singleSelections[opt.id];
        final val = opt.values.where((v) => v.id == selectedValId).firstOrNull;
        if (val != null) labels.add(val.name);
      } else {
        final selectedSet = _multiSelections[opt.id] ?? {};
        for (final val in opt.values) {
          if (selectedSet.contains(val.id)) {
            labels.add(val.name);
          }
        }
      }
    }
    return labels;
  }

  void _handleAddToCart() {
    CartManager.instance.addItem(
      menuItemId: _item.id,
      name: _item.name,
      basePricePoisha: _item.pricePoisha,
      quantity: _quantity,
      selectedOptions: _selectedOptionLabels,
      optionsAdditionalPoisha: _optionsAdditionalPricePoisha,
      restaurantId: widget.restaurantId ?? _item.restaurantId,
      restaurantName: widget.restaurantName ?? 'Burger Barn',
    );

    Navigator.pop(context);
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(
          'Added to Cart! (${CurrencyFormatter.formatPoisha(_totalPricePoisha)})',
        ),
        backgroundColor: AppColors.success,
        duration: const Duration(seconds: 2),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        iconTheme: const IconThemeData(color: AppColors.textPrimary),
        actions: [
          IconButton(icon: const Icon(Icons.favorite_border), onPressed: () {}),
        ],
      ),
      body: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              height: 240,
              width: double.infinity,
              color: AppColors.greyLight,
              child: Center(
                child: Icon(
                  _item.category == 'Drinks'
                      ? Icons.local_cafe
                      : _item.category == 'Biryani'
                      ? Icons.rice_bowl
                      : Icons.lunch_dining,
                  size: 110,
                  color: AppColors.textSecondary,
                ),
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(24.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Expanded(
                        child: Text(
                          _item.name,
                          style: const TextStyle(
                            fontSize: 24,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ),
                      const SizedBox(width: 16),
                      Text(
                        CurrencyFormatter.formatPoisha(_unitPricePoisha),
                        style: const TextStyle(
                          fontSize: 24,
                          fontWeight: FontWeight.bold,
                          color: AppColors.primaryColor,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  Text(
                    _item.description,
                    style: const TextStyle(
                      color: AppColors.textSecondary,
                      height: 1.5,
                      fontSize: 14,
                    ),
                  ),
                  const SizedBox(height: 28),

                  // Dynamic Options
                  for (final option in _item.options) ...[
                    Text(
                      option.name,
                      style: const TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(height: 12),
                    if (!option.allowsMultiple) ...[
                      Wrap(
                        spacing: 8,
                        runSpacing: 8,
                        children: [
                          for (final val in option.values) ...[
                            ChoiceChip(
                              label: Text(
                                val.additionalPricePoisha > 0
                                    ? '${val.name} (+${CurrencyFormatter.formatPoisha(val.additionalPricePoisha)})'
                                    : val.name,
                              ),
                              selected: _singleSelections[option.id] == val.id,
                              selectedColor: AppColors.primaryColor.withValues(
                                alpha: 0.15,
                              ),
                              side: BorderSide(
                                color: _singleSelections[option.id] == val.id
                                    ? AppColors.primaryColor
                                    : Colors.grey.shade300,
                              ),
                              labelStyle: TextStyle(
                                color: _singleSelections[option.id] == val.id
                                    ? AppColors.primaryColor
                                    : AppColors.textPrimary,
                                fontWeight:
                                    _singleSelections[option.id] == val.id
                                    ? FontWeight.bold
                                    : FontWeight.normal,
                              ),
                              onSelected: (selected) {
                                if (selected) {
                                  setState(() {
                                    _singleSelections[option.id] = val.id;
                                  });
                                }
                              },
                            ),
                          ],
                        ],
                      ),
                    ] else ...[
                      // Multi choice
                      for (final val in option.values) ...[
                        CheckboxListTile(
                          value: (_multiSelections[option.id] ?? {}).contains(
                            val.id,
                          ),
                          activeColor: AppColors.primaryColor,
                          contentPadding: EdgeInsets.zero,
                          title: Text(val.name),
                          secondary: Text(
                            val.additionalPricePoisha > 0
                                ? '+${CurrencyFormatter.formatPoisha(val.additionalPricePoisha)}'
                                : 'Free',
                            style: const TextStyle(
                              color: AppColors.textSecondary,
                              fontSize: 13,
                            ),
                          ),
                          onChanged: (checked) {
                            setState(() {
                              final set = _multiSelections[option.id] ?? {};
                              if (checked == true) {
                                set.add(val.id);
                              } else {
                                set.remove(val.id);
                              }
                              _multiSelections[option.id] = set;
                            });
                          },
                        ),
                      ],
                    ],
                    const SizedBox(height: 20),
                  ],
                ],
              ),
            ),
          ],
        ),
      ),
      bottomNavigationBar: Container(
        padding: const EdgeInsets.all(24),
        decoration: BoxDecoration(
          color: Colors.white,
          boxShadow: [
            BoxShadow(
              color: Colors.black.withValues(alpha: 0.05),
              blurRadius: 10,
              offset: const Offset(0, -5),
            ),
          ],
        ),
        child: SafeArea(
          child: Row(
            children: [
              // Quantity stepper
              Container(
                decoration: BoxDecoration(
                  border: Border.all(color: Colors.grey.shade300),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Row(
                  children: [
                    IconButton(
                      icon: const Icon(Icons.remove),
                      onPressed: () {
                        if (_quantity > 1) {
                          setState(() => _quantity--);
                        }
                      },
                    ),
                    Text(
                      '$_quantity',
                      style: const TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    IconButton(
                      icon: const Icon(Icons.add),
                      onPressed: () {
                        setState(() => _quantity++);
                      },
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 16),

              // Add to cart button
              Expanded(
                child: SizedBox(
                  height: 50,
                  child: ElevatedButton(
                    style: ElevatedButton.styleFrom(
                      backgroundColor: AppColors.primaryColor,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                    ),
                    onPressed: _handleAddToCart,
                    child: Text(
                      'Add to Cart - ${CurrencyFormatter.formatPoisha(_totalPricePoisha)}',
                      style: const TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
                      ),
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
