import { cn } from "@/lib/utils";
import type { InputHTMLAttributes } from "react";

export type InputProps = InputHTMLAttributes<HTMLInputElement>;

export function Input({ className, type = "text", ...props }: InputProps) {
  return (
    <input
      type={type}
      className={cn(
        "flex h-9 w-full rounded-md border border-gray-200 bg-transparent px-3 py-2 text-sm text-foreground placeholder:text-gray-500 transition-colors duration-150 focus-visible:outline-none focus-visible:ring-[3px] focus-visible:ring-black/10 dark:border-gray-800 dark:placeholder:text-gray-400 dark:focus-visible:ring-white/[0.12]",
        className
      )}
      {...props}
    />
  );
}
