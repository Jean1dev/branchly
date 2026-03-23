import { cn } from "@/lib/utils";
import Link from "next/link";
import type { ButtonHTMLAttributes, ReactNode } from "react";

export type ButtonProps = ButtonHTMLAttributes<HTMLButtonElement> & {
  variant?: "primary" | "secondary" | "ghost" | "destructive";
  size?: "sm" | "md" | "lg";
  children: ReactNode;
  href?: string;
};

const variantClasses: Record<NonNullable<ButtonProps["variant"]>, string> = {
  primary:
    "bg-black text-white hover:bg-gray-900 dark:bg-white dark:text-black dark:hover:bg-gray-100 border border-black dark:border-white",
  secondary:
    "border border-gray-200 bg-transparent text-foreground hover:bg-gray-100 dark:border-gray-800 dark:hover:bg-gray-900",
  ghost:
    "border border-transparent bg-transparent text-foreground hover:bg-gray-100 dark:hover:bg-gray-900",
  destructive:
    "border border-red-600 text-red-600 bg-transparent hover:bg-red-50 dark:hover:bg-red-950/30",
};

const sizeClasses: Record<NonNullable<ButtonProps["size"]>, string> = {
  sm: "h-8 px-3 text-sm rounded-md",
  md: "h-9 px-4 text-sm rounded-md",
  lg: "h-11 px-6 text-base rounded-md",
};

export function Button({
  className,
  variant = "primary",
  size = "md",
  type = "button",
  href,
  children,
  ...props
}: ButtonProps) {
  const classes = cn(
    "inline-flex items-center justify-center gap-2 font-medium transition-colors duration-150 focus-visible:outline-none focus-visible:ring-[3px] focus-visible:ring-black/10 dark:focus-visible:ring-white/[0.12] disabled:pointer-events-none disabled:opacity-50",
    variantClasses[variant],
    sizeClasses[size],
    className
  );
  if (href) {
    return (
      <Link href={href} className={classes}>
        {children}
      </Link>
    );
  }
  return (
    <button type={type} className={classes} {...props}>
      {children}
    </button>
  );
}
