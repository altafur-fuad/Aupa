import 'package:flutter/material.dart';
import '../../constants/colors.dart';
import '../../core/widgets/empty_state_view.dart';

enum NotificationType { order, promo, system }

class AppNotificationItem {
  final String id;
  final String title;
  final String message;
  final DateTime time;
  final NotificationType type;
  bool isRead;

  AppNotificationItem({
    required this.id,
    required this.title,
    required this.message,
    required this.time,
    required this.type,
    this.isRead = false,
  });
}

class NotificationsScreen extends StatefulWidget {
  const NotificationsScreen({super.key});

  @override
  State<NotificationsScreen> createState() => _NotificationsScreenState();
}

class _NotificationsScreenState extends State<NotificationsScreen> {
  String _selectedFilter = 'All';

  final List<AppNotificationItem> _notifications = [
    AppNotificationItem(
      id: 'notif_1',
      title: 'Order Delivered! 🎉',
      message:
          'Your order from Sultan\'s Dine has been delivered. Enjoy your Kacchi Biryani!',
      time: DateTime.now().subtract(const Duration(minutes: 25)),
      type: NotificationType.order,
      isRead: false,
    ),
    AppNotificationItem(
      id: 'notif_2',
      title: 'Flat 50% OFF Weekend! 🍔',
      message:
          'Use voucher code AUPA50 on orders above ৳500 to enjoy delicious discounts.',
      time: DateTime.now().subtract(const Duration(hours: 3)),
      type: NotificationType.promo,
      isRead: false,
    ),
    AppNotificationItem(
      id: 'notif_3',
      title: 'Rider Assigned 🛵',
      message:
          'Rider Tariqul Islam is picking up your order from Chillox Banani.',
      time: DateTime.now().subtract(const Duration(hours: 5)),
      type: NotificationType.order,
      isRead: true,
    ),
    AppNotificationItem(
      id: 'notif_4',
      title: 'Free Delivery in Gulshan 🚀',
      message:
          'Zero delivery fee for the next 2 hours on all orders from selected restaurants.',
      time: DateTime.now().subtract(const Duration(days: 1)),
      type: NotificationType.promo,
      isRead: true,
    ),
    AppNotificationItem(
      id: 'notif_5',
      title: 'Account Security Notice',
      message: 'Your Aupa account password was successfully verified.',
      time: DateTime.now().subtract(const Duration(days: 2)),
      type: NotificationType.system,
      isRead: true,
    ),
  ];

  List<AppNotificationItem> get _filteredNotifications {
    if (_selectedFilter == 'Orders') {
      return _notifications
          .where((n) => n.type == NotificationType.order)
          .toList();
    }
    if (_selectedFilter == 'Offers') {
      return _notifications
          .where((n) => n.type == NotificationType.promo)
          .toList();
    }
    return _notifications;
  }

  void _markAllAsRead() {
    setState(() {
      for (final n in _notifications) {
        n.isRead = true;
      }
    });
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('All notifications marked as read.'),
        duration: Duration(seconds: 2),
      ),
    );
  }

  void _clearAll() {
    setState(() {
      _notifications.clear();
    });
  }

  String _formatTimeAgo(DateTime time) {
    final diff = DateTime.now().difference(time);
    if (diff.inMinutes < 60) {
      return '${diff.inMinutes}m ago';
    } else if (diff.inHours < 24) {
      return '${diff.inHours}h ago';
    } else {
      return '${diff.inDays}d ago';
    }
  }

  IconData _getIconForType(NotificationType type) {
    switch (type) {
      case NotificationType.order:
        return Icons.delivery_dining;
      case NotificationType.promo:
        return Icons.local_offer;
      case NotificationType.system:
        return Icons.notifications;
    }
  }

  Color _getColorForType(NotificationType type) {
    switch (type) {
      case NotificationType.order:
        return AppColors.primaryColor;
      case NotificationType.promo:
        return Colors.deepOrange;
      case NotificationType.system:
        return Colors.blueGrey;
    }
  }

  @override
  Widget build(BuildContext context) {
    final items = _filteredNotifications;

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        backgroundColor: AppColors.background,
        elevation: 0,
        title: const Text(
          'Notifications',
          style: TextStyle(
            color: AppColors.textPrimary,
            fontWeight: FontWeight.bold,
            fontSize: 20,
          ),
        ),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: AppColors.textPrimary),
          onPressed: () => Navigator.pop(context),
        ),
        actions: [
          if (_notifications.isNotEmpty) ...[
            TextButton(
              onPressed: _markAllAsRead,
              child: const Text(
                'Mark read',
                style: TextStyle(
                  color: AppColors.primaryColor,
                  fontWeight: FontWeight.w600,
                  fontSize: 14,
                ),
              ),
            ),
            IconButton(
              icon: const Icon(
                Icons.delete_outline,
                color: AppColors.textSecondary,
              ),
              tooltip: 'Clear All',
              onPressed: () {
                showDialog(
                  context: context,
                  builder: (ctx) => AlertDialog(
                    title: const Text('Clear all notifications?'),
                    content: const Text(
                      'Are you sure you want to delete all notifications?',
                    ),
                    actions: [
                      TextButton(
                        onPressed: () => Navigator.pop(ctx),
                        child: const Text('Cancel'),
                      ),
                      TextButton(
                        onPressed: () {
                          Navigator.pop(ctx);
                          _clearAll();
                        },
                        child: const Text(
                          'Clear',
                          style: TextStyle(color: Colors.red),
                        ),
                      ),
                    ],
                  ),
                );
              },
            ),
          ],
        ],
      ),
      body: _notifications.isEmpty
          ? const EmptyStateView(
              icon: Icons.notifications_off_outlined,
              title: 'No Notifications',
              message:
                  'You have no notifications right now. Updates and promotions will appear here.',
            )
          : Column(
              children: [
                // Filter chips
                Padding(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 16.0,
                    vertical: 8.0,
                  ),
                  child: Row(
                    children: ['All', 'Orders', 'Offers'].map((filter) {
                      final isSelected = _selectedFilter == filter;
                      return Padding(
                        padding: const EdgeInsets.only(right: 8.0),
                        child: ChoiceChip(
                          label: Text(filter),
                          selected: isSelected,
                          selectedColor: AppColors.primaryColor,
                          labelStyle: TextStyle(
                            color: isSelected
                                ? Colors.white
                                : AppColors.textPrimary,
                            fontWeight: isSelected
                                ? FontWeight.bold
                                : FontWeight.normal,
                          ),
                          backgroundColor: AppColors.greyLight,
                          onSelected: (_) {
                            setState(() {
                              _selectedFilter = filter;
                            });
                          },
                        ),
                      );
                    }).toList(),
                  ),
                ),
                Expanded(
                  child: items.isEmpty
                      ? EmptyStateView(
                          icon: Icons.filter_list_off,
                          title: 'No $_selectedFilter Notifications',
                          message:
                              'There are no notifications matching the selected filter.',
                        )
                      : ListView.separated(
                          padding: const EdgeInsets.all(16.0),
                          itemCount: items.length,
                          separatorBuilder: (_, __) =>
                              const SizedBox(height: 12),
                          itemBuilder: (context, index) {
                            final notif = items[index];
                            final iconColor = _getColorForType(notif.type);

                            return InkWell(
                              borderRadius: BorderRadius.circular(12),
                              onTap: () {
                                setState(() {
                                  notif.isRead = true;
                                });
                              },
                              child: Container(
                                padding: const EdgeInsets.all(14.0),
                                decoration: BoxDecoration(
                                  color: notif.isRead
                                      ? Colors.white
                                      : AppColors.primaryColor.withValues(
                                          alpha: 0.05,
                                        ),
                                  borderRadius: BorderRadius.circular(12),
                                  border: Border.all(
                                    color: notif.isRead
                                        ? AppColors.greyLight
                                        : AppColors.primaryColor.withValues(
                                            alpha: 0.3,
                                          ),
                                  ),
                                ),
                                child: Row(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Container(
                                      padding: const EdgeInsets.all(10),
                                      decoration: BoxDecoration(
                                        color: iconColor.withValues(
                                          alpha: 0.12,
                                        ),
                                        shape: BoxShape.circle,
                                      ),
                                      child: Icon(
                                        _getIconForType(notif.type),
                                        color: iconColor,
                                        size: 22,
                                      ),
                                    ),
                                    const SizedBox(width: 14),
                                    Expanded(
                                      child: Column(
                                        crossAxisAlignment:
                                            CrossAxisAlignment.start,
                                        children: [
                                          Row(
                                            mainAxisAlignment:
                                                MainAxisAlignment.spaceBetween,
                                            children: [
                                              Expanded(
                                                child: Text(
                                                  notif.title,
                                                  style: TextStyle(
                                                    fontWeight: notif.isRead
                                                        ? FontWeight.w600
                                                        : FontWeight.bold,
                                                    fontSize: 15,
                                                    color:
                                                        AppColors.textPrimary,
                                                  ),
                                                ),
                                              ),
                                              Text(
                                                _formatTimeAgo(notif.time),
                                                style: const TextStyle(
                                                  fontSize: 12,
                                                  color:
                                                      AppColors.textSecondary,
                                                ),
                                              ),
                                            ],
                                          ),
                                          const SizedBox(height: 6),
                                          Text(
                                            notif.message,
                                            style: const TextStyle(
                                              fontSize: 13,
                                              color: AppColors.textSecondary,
                                              height: 1.3,
                                            ),
                                          ),
                                        ],
                                      ),
                                    ),
                                    if (!notif.isRead) ...[
                                      const SizedBox(width: 8),
                                      Container(
                                        width: 8,
                                        height: 8,
                                        decoration: const BoxDecoration(
                                          color: AppColors.primaryColor,
                                          shape: BoxShape.circle,
                                        ),
                                      ),
                                    ],
                                  ],
                                ),
                              ),
                            );
                          },
                        ),
                ),
              ],
            ),
    );
  }
}
