import { cn } from "@/lib/utils";
import type { HTMLAttributes, ReactNode } from "react";

export type BadgeProps = HTMLAttributes<HTMLSpanElement> & {
  variant?: "default" | "success" | "warning" | "error" | "muted";
  children: ReactNode;
};

const variantClasses: Record<NonNullable<BadgeProps["variant"]>, string> = {
  default:
    "border border-gray-200 bg-gray-100 text-gray-900 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-50",
  success:
    "border border-gray-300 bg-gray-200 text-gray-950 dark:border-gray-600 dark:bg-gray-800 dark:text-gray-50",
  warning:
    "border border-gray-200 bg-gray-50 text-gray-700 dark:border-gray-800 dark:bg-gray-950 dark:text-gray-300",
  error:
    "border border-gray-400 bg-gray-300 text-gray-950 dark:border-gray-600 dark:bg-gray-800 dark:text-gray-100",
  muted: "border border-gray-200 text-gray-500 dark:border-gray-800 dark:text-gray-400",
};

export function Badge({
  className,
  variant = "default",
  children,
  ...props
}: BadgeProps) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium",
        variantClasses[variant],
        className
      )}
      {...props}
    >
      {children}
    </span>
  );
}
