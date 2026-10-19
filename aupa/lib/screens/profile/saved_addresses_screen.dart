import 'package:flutter/material.dart';
import '../../constants/colors.dart';
import '../../core/di/service_locator.dart';
import '../../core/widgets/app_button.dart';
import '../../core/widgets/empty_state_view.dart';
import '../../core/widgets/loading_indicator.dart';
import '../../domain/models/address.dart';

class SavedAddressesScreen extends StatefulWidget {
  const SavedAddressesScreen({super.key});

  @override
  State<SavedAddressesScreen> createState() => _SavedAddressesScreenState();
}

class _SavedAddressesScreenState extends State<SavedAddressesScreen> {
  List<Address>? _addresses;
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _loadAddresses();
  }

  Future<void> _loadAddresses() async {
    setState(() => _isLoading = true);
    final res = await ServiceLocator.instance.addressRepository.getAddresses();
    if (!mounted) return;
    if (res.isSuccess) {
      setState(() {
        _addresses = res.data;
        _isLoading = false;
      });
    } else {
      setState(() => _isLoading = false);
    }
  }

  Future<void> _deleteAddress(String id) async {
    await ServiceLocator.instance.addressRepository.deleteAddress(id);
    _loadAddresses();
  }

  Future<void> _setDefaultAddress(String id) async {
    await ServiceLocator.instance.addressRepository.setDefaultAddress(id);
    _loadAddresses();
  }

  void _showAddAddressDialog() {
    final streetController = TextEditingController();
    final areaController = TextEditingController(text: 'Gulshan');
    AddressLabel selectedLabel = AddressLabel.home;
    bool isDefault = false;

    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (context) {
        return StatefulBuilder(
          builder: (context, setModalState) {
            return Padding(
              padding: EdgeInsets.only(
                left: 20,
                right: 20,
                top: 20,
                bottom: MediaQuery.of(context).viewInsets.bottom + 20,
              ),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const Text(
                    'Add New Address',
                    style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
                  ),
                  const SizedBox(height: 16),
                  TextField(
                    controller: streetController,
                    decoration: const InputDecoration(
                      labelText: 'Street / Building / Flat',
                      hintText: 'e.g. House 12, Road 4',
                    ),
                  ),
                  const SizedBox(height: 12),
                  TextField(
                    controller: areaController,
                    decoration: const InputDecoration(
                      labelText: 'Area / Neighborhood',
                      hintText: 'e.g. Banani, Gulshan, Dhanmondi',
                    ),
                  ),
                  const SizedBox(height: 16),
                  const Text(
                    'Label',
                    style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
                  ),
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      ChoiceChip(
                        label: const Text('Home'),
                        selected: selectedLabel == AddressLabel.home,
                        onSelected: (_) => setModalState(
                          () => selectedLabel = AddressLabel.home,
                        ),
                      ),
                      const SizedBox(width: 8),
                      ChoiceChip(
                        label: const Text('Work'),
                        selected: selectedLabel == AddressLabel.work,
                        onSelected: (_) => setModalState(
                          () => selectedLabel = AddressLabel.work,
                        ),
                      ),
                      const SizedBox(width: 8),
                      ChoiceChip(
                        label: const Text('Other'),
                        selected: selectedLabel == AddressLabel.other,
                        onSelected: (_) => setModalState(
                          () => selectedLabel = AddressLabel.other,
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 12),
                  CheckboxListTile(
                    value: isDefault,
                    contentPadding: EdgeInsets.zero,
                    activeColor: AppColors.primaryColor,
                    title: const Text(
                      'Set as default delivery address',
                      style: TextStyle(fontSize: 14),
                    ),
                    onChanged: (val) =>
                        setModalState(() => isDefault = val ?? false),
                  ),
                  const SizedBox(height: 16),
                  AppButton(
                    label: 'Save Address',
                    onPressed: () async {
                      if (streetController.text.trim().isEmpty) return;
                      final newAddr = Address(
                        id: '',
                        label: selectedLabel,
                        street: streetController.text.trim(),
                        area: areaController.text.trim().isEmpty
                            ? 'Dhaka'
                            : areaController.text.trim(),
                        city: 'Dhaka',
                        isDefault: isDefault,
                      );
                      await ServiceLocator.instance.addressRepository
                          .addAddress(newAddr);
                      if (!context.mounted) return;
                      Navigator.pop(context);
                      _loadAddresses();
                    },
                  ),
                ],
              ),
            );
          },
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        title: const Text('Saved Addresses'),
        backgroundColor: Colors.transparent,
        elevation: 0,
        iconTheme: const IconThemeData(color: AppColors.textPrimary),
        actions: [
          IconButton(
            icon: const Icon(Icons.add_location_alt_outlined),
            onPressed: _showAddAddressDialog,
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: AppLoadingIndicator())
          : _addresses == null || _addresses!.isEmpty
          ? EmptyStateView(
              icon: Icons.location_off_outlined,
              title: 'No addresses saved',
              message: 'Add your home or office address for faster checkout.',
              actionLabel: 'Add Address',
              onAction: _showAddAddressDialog,
            )
          : ListView.builder(
              padding: const EdgeInsets.all(16),
              itemCount: _addresses!.length,
              itemBuilder: (context, index) {
                final addr = _addresses![index];
                return Container(
                  margin: const EdgeInsets.only(bottom: 14),
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    borderRadius: BorderRadius.circular(16),
                    border: Border.all(
                      color: addr.isDefault
                          ? AppColors.primaryColor
                          : Colors.grey.shade200,
                      width: addr.isDefault ? 1.5 : 1,
                    ),
                  ),
                  child: Row(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Icon(
                        addr.label == AddressLabel.home
                            ? Icons.home
                            : addr.label == AddressLabel.work
                            ? Icons.work
                            : Icons.location_on,
                        color: AppColors.primaryColor,
                      ),
                      const SizedBox(width: 14),
                      Expanded(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              children: [
                                Text(
                                  addr.displayLabel,
                                  style: const TextStyle(
                                    fontWeight: FontWeight.bold,
                                    fontSize: 16,
                                  ),
                                ),
                                if (addr.isDefault) ...[
                                  const SizedBox(width: 8),
                                  Container(
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 8,
                                      vertical: 2,
                                    ),
                                    decoration: BoxDecoration(
                                      color: AppColors.primaryColor.withValues(
                                        alpha: 0.1,
                                      ),
                                      borderRadius: BorderRadius.circular(6),
                                    ),
                                    child: const Text(
                                      'Default',
                                      style: TextStyle(
                                        color: AppColors.primaryColor,
                                        fontWeight: FontWeight.bold,
                                        fontSize: 11,
                                      ),
                                    ),
                                  ),
                                ],
                              ],
                            ),
                            const SizedBox(height: 4),
                            Text(
                              addr.fullAddress,
                              style: const TextStyle(
                                color: AppColors.textSecondary,
                                fontSize: 13,
                              ),
                            ),
                            const SizedBox(height: 10),
                            Row(
                              children: [
                                if (!addr.isDefault)
                                  TextButton(
                                    onPressed: () =>
                                        _setDefaultAddress(addr.id),
                                    style: TextButton.styleFrom(
                                      padding: EdgeInsets.zero,
                                      minimumSize: Size.zero,
                                      tapTargetSize:
                                          MaterialTapTargetSize.shrinkWrap,
                                    ),
                                    child: const Text(
                                      'Set as Default',
                                      style: TextStyle(fontSize: 12),
                                    ),
                                  ),
                                const Spacer(),
                                IconButton(
                                  icon: const Icon(
                                    Icons.delete_outline,
                                    size: 20,
                                    color: AppColors.textSecondary,
                                  ),
                                  onPressed: () => _deleteAddress(addr.id),
                                ),
                              ],
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                );
              },
            ),
      floatingActionButton: FloatingActionButton.extended(
        backgroundColor: AppColors.primaryColor,
        icon: const Icon(Icons.add, color: Colors.white),
        label: const Text(
          'Add New Address',
          style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold),
        ),
        onPressed: _showAddAddressDialog,
      ),
    );
  }
}
