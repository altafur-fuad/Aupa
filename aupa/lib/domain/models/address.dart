enum AddressLabel { home, work, other }

class Address {
  final String id;
  final AddressLabel label;
  final String street;
  final String area;
  final String city;
  final String? postalCode;
  final String? notes;
  final bool isDefault;
  final double? latitude;
  final double? longitude;

  const Address({
    required this.id,
    required this.label,
    required this.street,
    required this.area,
    this.city = 'Dhaka',
    this.postalCode,
    this.notes,
    this.isDefault = false,
    this.latitude,
    this.longitude,
  });

  String get fullAddress => '$street, $area, $city';

  String get displayLabel {
    switch (label) {
      case AddressLabel.home:
        return 'Home';
      case AddressLabel.work:
        return 'Work';
      case AddressLabel.other:
        return 'Other';
    }
  }

  Address copyWith({
    String? id,
    AddressLabel? label,
    String? street,
    String? area,
    String? city,
    String? postalCode,
    String? notes,
    bool? isDefault,
    double? latitude,
    double? longitude,
  }) {
    return Address(
      id: id ?? this.id,
      label: label ?? this.label,
      street: street ?? this.street,
      area: area ?? this.area,
      city: city ?? this.city,
      postalCode: postalCode ?? this.postalCode,
      notes: notes ?? this.notes,
      isDefault: isDefault ?? this.isDefault,
      latitude: latitude ?? this.latitude,
      longitude: longitude ?? this.longitude,
    );
  }
}
