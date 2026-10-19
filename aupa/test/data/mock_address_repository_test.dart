import 'package:flutter_test/flutter_test.dart';
import 'package:aupa/data/mock/mock_address_repository.dart';
import 'package:aupa/domain/models/address.dart';

void main() {
  late MockAddressRepository repo;

  setUp(() {
    repo = MockAddressRepository();
  });

  group('MockAddressRepository', () {
    test('getAddresses returns saved addresses', () async {
      final res = await repo.getAddresses();
      expect(res.isSuccess, isTrue);
      expect(res.data.length, 2);
    });

    test('getDefaultAddress returns default address', () async {
      final res = await repo.getDefaultAddress();
      expect(res.isSuccess, isTrue);
      expect(res.data.isDefault, isTrue);
      expect(res.data.area, 'Banani');
    });

    test('addAddress appends and sets as default when requested', () async {
      const newAddr = Address(
        id: '',
        label: AddressLabel.other,
        street: 'Road 11',
        area: 'Uttara',
        isDefault: true,
      );

      final res = await repo.addAddress(newAddr);
      expect(res.isSuccess, isTrue);
      expect(res.data.isDefault, isTrue);

      final defaultRes = await repo.getDefaultAddress();
      expect(defaultRes.data.id, res.data.id);
    });

    test('deleteAddress removes address', () async {
      final res = await repo.deleteAddress('addr_02');
      expect(res.isSuccess, isTrue);

      final all = await repo.getAddresses();
      expect(all.data.any((a) => a.id == 'addr_02'), isFalse);
    });
  });
}
