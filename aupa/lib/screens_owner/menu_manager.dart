import 'package:flutter/material.dart';
import '../../constants/colors.dart';

class MenuManager extends StatelessWidget {
  const MenuManager({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        title: const Text('Manage Menu'),
        backgroundColor: AppColors.primaryColor,
        foregroundColor: Colors.white,
        actions: [IconButton(icon: const Icon(Icons.add), onPressed: () {})],
      ),
      body: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          const Text(
            'Burgers',
            style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 16),
          _buildMenuEditorItem('Classic Cheese Burger', '\$14.99', true),
          _buildMenuEditorItem('Spicy BBQ Burger', '\$16.99', true),
          const SizedBox(height: 24),
          const Text(
            'Sides',
            style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 16),
          _buildMenuEditorItem('Large Fries', '\$5.99', false), // out of stock
        ],
      ),
    );
  }

  Widget _buildMenuEditorItem(String title, String price, bool isAvailable) {
    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade300),
      ),
      child: Row(
        children: [
          Container(
            width: 60,
            height: 60,
            decoration: BoxDecoration(
              color: AppColors.greyLight,
              borderRadius: BorderRadius.circular(8),
            ),
            child: const Icon(Icons.fastfood, color: AppColors.textSecondary),
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: const TextStyle(
                    fontWeight: FontWeight.bold,
                    fontSize: 16,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  price,
                  style: const TextStyle(
                    color: AppColors.primaryColor,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
          ),
          Switch(
            value: isAvailable,
            activeThumbColor: AppColors.primaryColor,
            onChanged: (val) {},
          ),
        ],
      ),
    );
  }
}
