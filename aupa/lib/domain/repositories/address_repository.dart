import '../../core/network/result.dart';
import '../models/address.dart';

abstract class AddressRepository {
  Future<Result<List<Address>>> getAddresses();

  Future<Result<Address>> getAddressById(String id);

  Future<Result<Address>> getDefaultAddress();

  Future<Result<Address>> addAddress(Address address);

  Future<Result<Address>> updateAddress(Address address);

  Future<Result<void>> deleteAddress(String id);

  Future<Result<Address>> setDefaultAddress(String id);
}
