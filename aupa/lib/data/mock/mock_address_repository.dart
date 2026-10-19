import '../../core/errors/failures.dart';
import '../../core/network/result.dart';
import '../../domain/models/address.dart';
import '../../domain/repositories/address_repository.dart';

class MockAddressRepository implements AddressRepository {
  final List<Address> _addresses = [
    const Address(
      id: 'addr_01',
      label: AddressLabel.home,
      street: 'House 45, Road 7, Block C',
      area: 'Banani',
      city: 'Dhaka',
      postalCode: '1213',
      notes: 'Ring bell 4B',
      isDefault: true,
      latitude: 23.7937,
      longitude: 90.4066,
    ),
    const Address(
      id: 'addr_02',
      label: AddressLabel.work,
      street: 'Crystal Palace, 3rd Floor, Gulshan South Ave',
      area: 'Gulshan 1',
      city: 'Dhaka',
      postalCode: '1212',
      notes: 'Reception desk dropoff',
      isDefault: false,
      latitude: 23.7785,
      longitude: 90.4172,
    ),
  ];

  @override
  Future<Result<List<Address>>> getAddresses() async {
    await Future.delayed(const Duration(milliseconds: 100));
    return Result.success(List.unmodifiable(_addresses));
  }

  @override
  Future<Result<Address>> getAddressById(String id) async {
    await Future.delayed(const Duration(milliseconds: 100));
    final match = _addresses.where((a) => a.id == id).firstOrNull;
    if (match == null) {
      return const Result.failure(NotFoundFailure('Address not found.'));
    }
    return Result.success(match);
  }

  @override
  Future<Result<Address>> getDefaultAddress() async {
    await Future.delayed(const Duration(milliseconds: 100));
    final defaultAddr = _addresses.where((a) => a.isDefault).firstOrNull;
    if (defaultAddr != null) {
      return Result.success(defaultAddr);
    }
    if (_addresses.isNotEmpty) {
      return Result.success(_addresses.first);
    }
    return const Result.failure(NotFoundFailure('No delivery address saved.'));
  }

  @override
  Future<Result<Address>> addAddress(Address address) async {
    await Future.delayed(const Duration(milliseconds: 150));

    final newId = address.id.isEmpty
        ? 'addr_${DateTime.now().millisecondsSinceEpoch}'
        : address.id;

    if (address.isDefault) {
      for (int i = 0; i < _addresses.length; i++) {
        _addresses[i] = _addresses[i].copyWith(isDefault: false);
      }
    }

    final newAddress = address.copyWith(id: newId);
    _addresses.add(newAddress);
    return Result.success(newAddress);
  }

  @override
  Future<Result<Address>> updateAddress(Address address) async {
    await Future.delayed(const Duration(milliseconds: 150));
    final index = _addresses.indexWhere((a) => a.id == address.id);
    if (index == -1) {
      return const Result.failure(
        NotFoundFailure('Address not found to update.'),
      );
    }

    if (address.isDefault) {
      for (int i = 0; i < _addresses.length; i++) {
        _addresses[i] = _addresses[i].copyWith(isDefault: false);
      }
    }

    _addresses[index] = address;
    return Result.success(address);
  }

  @override
  Future<Result<void>> deleteAddress(String id) async {
    await Future.delayed(const Duration(milliseconds: 150));
    final index = _addresses.indexWhere((a) => a.id == id);
    if (index == -1) {
      return const Result.failure(
        NotFoundFailure('Address not found to delete.'),
      );
    }

    final wasDefault = _addresses[index].isDefault;
    _addresses.removeAt(index);

    if (wasDefault && _addresses.isNotEmpty) {
      _addresses[0] = _addresses[0].copyWith(isDefault: true);
    }

    return const Result.success(null);
  }

  @override
  Future<Result<Address>> setDefaultAddress(String id) async {
    await Future.delayed(const Duration(milliseconds: 100));
    final index = _addresses.indexWhere((a) => a.id == id);
    if (index == -1) {
      return const Result.failure(NotFoundFailure('Address not found.'));
    }

    for (int i = 0; i < _addresses.length; i++) {
      _addresses[i] = _addresses[i].copyWith(isDefault: i == index);
    }

    return Result.success(_addresses[index]);
  }
}
