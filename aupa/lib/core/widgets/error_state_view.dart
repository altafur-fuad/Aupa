import 'package:flutter/material.dart';
import '../../constants/colors.dart';
import 'app_button.dart';

class ErrorStateView extends StatelessWidget {
  final String title;
  final String message;
  final String? errorCode;
  final VoidCallback? onRetry;
  final String retryLabel;

  const ErrorStateView({
    super.key,
    this.title = 'Something went wrong',
    required this.message,
    this.errorCode,
    this.onRetry,
    this.retryLabel = 'Try Again',
  });

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32.0),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                color: AppColors.error.withValues(alpha: 0.1),
                shape: BoxShape.circle,
              ),
              child: const Icon(
                Icons.error_outline,
                size: 56,
                color: AppColors.error,
              ),
            ),
            const SizedBox(height: 24),
            Text(
              title,
              textAlign: TextAlign.center,
              style: const TextStyle(
                fontSize: 20,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              message,
              textAlign: TextAlign.center,
              style: const TextStyle(
                fontSize: 14,
                color: AppColors.textSecondary,
                height: 1.4,
              ),
            ),
            if (errorCode != null) ...[
              const SizedBox(height: 6),
              Text(
                'Code: $errorCode',
                style: const TextStyle(
                  fontSize: 12,
                  color: AppColors.textMuted,
                  fontFamily: 'monospace',
                ),
              ),
            ],
            if (onRetry != null) ...[
              const SizedBox(height: 24),
              AppButton(
                label: retryLabel,
                onPressed: onRetry,
                width: 200,
                height: 46,
                icon: Icons.refresh,
              ),
            ],
          ],
        ),
      ),
    );
  }
}
