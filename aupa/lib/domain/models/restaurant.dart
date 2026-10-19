class Restaurant {
  final String id;
  final String name;
  final String cuisine;
  final String area;
  final double rating;
  final int reviewCount;
  final int deliveryTimeMinutes;
  final int deliveryFeePoisha;
  final bool isOpen;
  final String? imageUrl;
  final String address;

  const Restaurant({
    required this.id,
    required this.name,
    required this.cuisine,
    required this.area,
    required this.rating,
    required this.reviewCount,
    required this.deliveryTimeMinutes,
    required this.deliveryFeePoisha,
    this.isOpen = true,
    this.imageUrl,
    required this.address,
  });

  factory Restaurant.fromJson(Map<String, dynamic> json) {
    return Restaurant(
      id: json['id'] as String? ?? '',
      name: json['name'] as String? ?? '',
      cuisine: json['cuisine'] as String? ?? '',
      area: json['area'] as String? ?? 'Banani',
      rating: (json['rating'] as num?)?.toDouble() ?? 4.5,
      reviewCount: json['reviewCount'] as int? ?? 100,
      deliveryTimeMinutes: json['deliveryTimeMinutes'] as int? ?? 30,
      deliveryFeePoisha: json['deliveryFeePoisha'] as int? ?? 5000,
      isOpen: json['isOpen'] as bool? ?? true,
      imageUrl: json['imageUrl'] as String?,
      address: json['address'] as String? ?? 'Dhaka, Bangladesh',
    );
  }

  Restaurant copyWith({
    String? id,
    String? name,
    String? cuisine,
    String? area,
    double? rating,
    int? reviewCount,
    int? deliveryTimeMinutes,
    int? deliveryFeePoisha,
    bool? isOpen,
    String? imageUrl,
    String? address,
  }) {
    return Restaurant(
      id: id ?? this.id,
      name: name ?? this.name,
      cuisine: cuisine ?? this.cuisine,
      area: area ?? this.area,
      rating: rating ?? this.rating,
      reviewCount: reviewCount ?? this.reviewCount,
      deliveryTimeMinutes: deliveryTimeMinutes ?? this.deliveryTimeMinutes,
      deliveryFeePoisha: deliveryFeePoisha ?? this.deliveryFeePoisha,
      isOpen: isOpen ?? this.isOpen,
      imageUrl: imageUrl ?? this.imageUrl,
      address: address ?? this.address,
    );
  }
}
