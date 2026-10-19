import 'package:flutter/material.dart';
import '../../constants/colors.dart';
import '../../core/di/service_locator.dart';
import '../../core/widgets/app_button.dart';
import '../../core/widgets/loading_indicator.dart';
import '../../domain/models/user.dart';
import '../../screens_owner/owner_dashboard.dart';
import '../auth/login_screen.dart';
import '../notifications/notifications_screen.dart';
import '../orders/order_history_screen.dart';
import 'saved_addresses_screen.dart';

class ProfileScreen extends StatefulWidget {
  const ProfileScreen({super.key});

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  User? _currentUser;
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _loadProfile();
  }

  Future<void> _loadProfile() async {
    setState(() => _isLoading = true);
    final res = await ServiceLocator.instance.authRepository.getCurrentUser();
    if (!mounted) return;
    if (res.isSuccess) {
      setState(() {
        _currentUser = res.data;
        _isLoading = false;
      });
    } else {
      setState(() => _isLoading = false);
    }
  }

  Future<void> _handleLogout() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Log Out'),
        content: const Text(
          'Are you sure you want to log out of your Aupa account?',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: const Text('Log Out', style: TextStyle(color: Colors.red)),
          ),
        ],
      ),
    );

    if (confirmed == true) {
      await ServiceLocator.instance.authRepository.logout();
      if (!mounted) return;
      Navigator.pushAndRemoveUntil(
        context,
        MaterialPageRoute(builder: (_) => const LoginScreen()),
        (route) => false,
      );
    }
  }

  void _showHelpSupportDialog() {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Aupa Support'),
        content: const Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('Need help with your order or account?'),
            SizedBox(height: 12),
            Row(
              children: [
                Icon(Icons.phone, size: 18, color: AppColors.primaryColor),
                SizedBox(width: 8),
                Text(
                  '+880 9612-400400',
                  style: TextStyle(fontWeight: FontWeight.w600),
                ),
              ],
            ),
            SizedBox(height: 8),
            Row(
              children: [
                Icon(Icons.email, size: 18, color: AppColors.primaryColor),
                SizedBox(width: 8),
                Text('support@aupa.com.bd'),
              ],
            ),
            SizedBox(height: 8),
            Row(
              children: [
                Icon(
                  Icons.location_on,
                  size: 18,
                  color: AppColors.primaryColor,
                ),
                SizedBox(width: 8),
                Expanded(child: Text('Gulshan-2, Dhaka 1212, Bangladesh')),
              ],
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Close'),
          ),
        ],
      ),
    );
  }

  void _showTermsDialog() {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Terms & Privacy'),
        content: const SingleChildScrollView(
          child: Text(
            'Aupa operates under Bangladesh consumer digital commerce guidelines. '
            'We value your privacy and never sell customer data to third-party ad networks. '
            'All payments are processed securely via SSLCommerz / bKash merchant gateway.',
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('Agree'),
          ),
        ],
      ),
    );
  }

  Widget _buildMenuItem({
    required IconData icon,
    required String title,
    String? subtitle,
    required VoidCallback onTap,
    Color? iconColor,
    Color? textColor,
  }) {
    return ListTile(
      leading: Container(
        padding: const EdgeInsets.all(8),
        decoration: BoxDecoration(
          color: (iconColor ?? AppColors.primaryColor).withValues(alpha: 0.1),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Icon(icon, color: iconColor ?? AppColors.primaryColor, size: 20),
      ),
      title: Text(
        title,
        style: TextStyle(
          fontWeight: FontWeight.w600,
          fontSize: 15,
          color: textColor ?? AppColors.textPrimary,
        ),
      ),
      subtitle: subtitle != null
          ? Text(
              subtitle,
              style: const TextStyle(
                fontSize: 12,
                color: AppColors.textSecondary,
              ),
            )
          : null,
      trailing: const Icon(
        Icons.chevron_right,
        color: AppColors.textSecondary,
        size: 20,
      ),
      onTap: onTap,
    );
  }

  @override
  Widget build(BuildContext context) {
    if (_isLoading) {
      return const Scaffold(
        backgroundColor: AppColors.background,
        body: Center(child: AppLoadingIndicator()),
      );
    }

    final user = _currentUser;

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        backgroundColor: AppColors.background,
        elevation: 0,
        title: const Text(
          'My Profile',
          style: TextStyle(
            color: AppColors.textPrimary,
            fontWeight: FontWeight.bold,
            fontSize: 20,
          ),
        ),
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.symmetric(horizontal: 16.0, vertical: 8.0),
        child: Column(
          children: [
            // User Header Card
            Container(
              padding: const EdgeInsets.all(16.0),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(16),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.04),
                    blurRadius: 10,
                    offset: const Offset(0, 4),
                  ),
                ],
              ),
              child: user != null
                  ? Row(
                      children: [
                        CircleAvatar(
                          radius: 32,
                          backgroundColor: AppColors.primaryColor,
                          child: Text(
                            user.fullName.isNotEmpty
                                ? user.fullName[0].toUpperCase()
                                : 'A',
                            style: const TextStyle(
                              fontSize: 26,
                              fontWeight: FontWeight.bold,
                              color: Colors.white,
                            ),
                          ),
                        ),
                        const SizedBox(width: 16),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  Flexible(
                                    child: Text(
                                      user.fullName,
                                      style: const TextStyle(
                                        fontSize: 18,
                                        fontWeight: FontWeight.bold,
                                        color: AppColors.textPrimary,
                                      ),
                                      overflow: TextOverflow.ellipsis,
                                    ),
                                  ),
                                  if (user.isVerified) ...[
                                    const SizedBox(width: 6),
                                    const Icon(
                                      Icons.verified,
                                      color: Colors.green,
                                      size: 18,
                                    ),
                                  ],
                                ],
                              ),
                              const SizedBox(height: 4),
                              Text(
                                user.email,
                                style: const TextStyle(
                                  fontSize: 13,
                                  color: AppColors.textSecondary,
                                ),
                                overflow: TextOverflow.ellipsis,
                              ),
                              if (user.phone != null) ...[
                                const SizedBox(height: 2),
                                Text(
                                  user.phone!,
                                  style: const TextStyle(
                                    fontSize: 12,
                                    color: AppColors.textSecondary,
                                  ),
                                ),
                              ],
                            ],
                          ),
                        ),
                      ],
                    )
                  : Column(
                      children: [
                        const Text(
                          'You are currently signed out',
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.w600,
                            color: AppColors.textPrimary,
                          ),
                        ),
                        const SizedBox(height: 12),
                        AppButton(
                          label: 'Sign In to Aupa',
                          onPressed: () {
                            Navigator.push(
                              context,
                              MaterialPageRoute(
                                builder: (_) => const LoginScreen(),
                              ),
                            ).then((_) => _loadProfile());
                          },
                        ),
                      ],
                    ),
            ),
            const SizedBox(height: 20),

            // Navigation Links Section
            Container(
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(16),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.04),
                    blurRadius: 10,
                    offset: const Offset(0, 4),
                  ),
                ],
              ),
              child: Column(
                children: [
                  _buildMenuItem(
                    icon: Icons.location_on_outlined,
                    title: 'Saved Delivery Addresses',
                    subtitle: 'Manage home, office and saved locations',
                    onTap: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (_) => const SavedAddressesScreen(),
                        ),
                      );
                    },
                  ),
                  const Divider(height: 1, indent: 64),
                  _buildMenuItem(
                    icon: Icons.receipt_long_outlined,
                    title: 'My Orders',
                    subtitle: 'Order history, invoices & live tracking',
                    onTap: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (_) => const OrderHistoryScreen(),
                        ),
                      );
                    },
                  ),
                  const Divider(height: 1, indent: 64),
                  _buildMenuItem(
                    icon: Icons.notifications_none,
                    title: 'Notifications',
                    subtitle: 'Delivery updates, discounts & vouchers',
                    onTap: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (_) => const NotificationsScreen(),
                        ),
                      );
                    },
                  ),
                  const Divider(height: 1, indent: 64),
                  _buildMenuItem(
                    icon: Icons.storefront_outlined,
                    title: 'Merchant & Owner Portal',
                    subtitle: 'Manage restaurant menu, branch & orders',
                    iconColor: Colors.deepPurple,
                    onTap: () {
                      Navigator.push(
                        context,
                        MaterialPageRoute(
                          builder: (_) => const OwnerDashboard(),
                        ),
                      );
                    },
                  ),
                ],
              ),
            ),
            const SizedBox(height: 20),

            // Support & Info Section
            Container(
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(16),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.04),
                    blurRadius: 10,
                    offset: const Offset(0, 4),
                  ),
                ],
              ),
              child: Column(
                children: [
                  _buildMenuItem(
                    icon: Icons.help_outline,
                    title: 'Help & Customer Support',
                    iconColor: Colors.teal,
                    onTap: _showHelpSupportDialog,
                  ),
                  const Divider(height: 1, indent: 64),
                  _buildMenuItem(
                    icon: Icons.privacy_tip_outlined,
                    title: 'Terms & Privacy Policy',
                    iconColor: Colors.blueGrey,
                    onTap: _showTermsDialog,
                  ),
                ],
              ),
            ),
            const SizedBox(height: 20),

            // Logout Button
            if (user != null)
              Container(
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(16),
                ),
                child: _buildMenuItem(
                  icon: Icons.logout,
                  title: 'Log Out',
                  iconColor: Colors.red,
                  textColor: Colors.red,
                  onTap: _handleLogout,
                ),
              ),
            const SizedBox(height: 32),
          ],
        ),
      ),
    );
  }
}
