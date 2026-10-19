class User {
  final String id;
  final String fullName;
  final String email;
  final String? phone;
  final bool isVerified;
  final DateTime createdAt;

  const User({
    required this.id,
    required this.fullName,
    required this.email,
    this.phone,
    this.isVerified = true,
    required this.createdAt,
  });

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      id: json['id'] as String? ?? '',
      fullName: json['fullName'] as String? ?? json['name'] as String? ?? '',
      email: json['email'] as String? ?? '',
      phone: json['phone'] as String?,
      isVerified:
          json['isVerified'] as bool? ?? json['emailVerified'] as bool? ?? true,
      createdAt: json['createdAt'] != null
          ? DateTime.tryParse(json['createdAt'] as String) ?? DateTime.now()
          : DateTime.now(),
    );
  }

  User copyWith({
    String? id,
    String? fullName,
    String? email,
    String? phone,
    bool? isVerified,
    DateTime? createdAt,
  }) {
    return User(
      id: id ?? this.id,
      fullName: fullName ?? this.fullName,
      email: email ?? this.email,
      phone: phone ?? this.phone,
      isVerified: isVerified ?? this.isVerified,
      createdAt: createdAt ?? this.createdAt,
    );
  }
}
