import 'package:flutter/material.dart';
import '../../constants/colors.dart';

enum AppButtonVariant { primary, secondary, outline, text }

class AppButton extends StatelessWidget {
  final String label;
  final VoidCallback? onPressed;
  final bool isLoading;
  final bool isDisabled;
  final AppButtonVariant variant;
  final IconData? icon;
  final double? width;
  final double height;

  const AppButton({
    super.key,
    required this.label,
    required this.onPressed,
    this.isLoading = false,
    this.isDisabled = false,
    this.variant = AppButtonVariant.primary,
    this.icon,
    this.width,
    this.height = 50,
  });

  @override
  Widget build(BuildContext context) {
    final effectiveOnPressed = (isDisabled || isLoading) ? null : onPressed;

    Widget content;
    if (isLoading) {
      content = const SizedBox(
        height: 20,
        width: 20,
        child: CircularProgressIndicator(
          strokeWidth: 2.5,
          valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
        ),
      );
    } else if (icon != null) {
      content = Row(
        mainAxisAlignment: MainAxisAlignment.center,
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 18),
          const SizedBox(width: 8),
          Flexible(
            child: Text(label, overflow: TextOverflow.ellipsis, maxLines: 1),
          ),
        ],
      );
    } else {
      content = Text(label, overflow: TextOverflow.ellipsis, maxLines: 1);
    }

    final buttonStyle = ElevatedButton.styleFrom(
      padding: const EdgeInsets.symmetric(horizontal: 16),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
    );

    switch (variant) {
      case AppButtonVariant.primary:
        return SizedBox(
          width: width,
          height: height,
          child: ElevatedButton(
            onPressed: effectiveOnPressed,
            style: buttonStyle.copyWith(
              backgroundColor: WidgetStateProperty.resolveWith(
                (states) => states.contains(WidgetState.disabled)
                    ? AppColors.primaryColor.withValues(alpha: 0.5)
                    : AppColors.primaryColor,
              ),
              foregroundColor: const WidgetStatePropertyAll(Colors.white),
            ),
            child: content,
          ),
        );
      case AppButtonVariant.secondary:
        return SizedBox(
          width: width,
          height: height,
          child: ElevatedButton(
            onPressed: effectiveOnPressed,
            style: buttonStyle.copyWith(
              backgroundColor: const WidgetStatePropertyAll(
                AppColors.greyLight,
              ),
              foregroundColor: const WidgetStatePropertyAll(
                AppColors.textPrimary,
              ),
              elevation: const WidgetStatePropertyAll(0),
            ),
            child: content,
          ),
        );
      case AppButtonVariant.outline:
        return SizedBox(
          width: width,
          height: height,
          child: OutlinedButton(
            onPressed: effectiveOnPressed,
            style: OutlinedButton.styleFrom(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              side: const BorderSide(color: AppColors.primaryColor, width: 1.5),
              foregroundColor: AppColors.primaryColor,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
            ),
            child: content,
          ),
        );
      case AppButtonVariant.text:
        return TextButton(onPressed: effectiveOnPressed, child: content);
    }
  }
}
