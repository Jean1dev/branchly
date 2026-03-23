import { cn } from "@/lib/utils";
import type { HTMLAttributes, ReactNode } from "react";

export type CardProps = HTMLAttributes<HTMLDivElement> & {
  children: ReactNode;
};

export function Card({ className, children, ...props }: CardProps) {
  return (
    <div
      className={cn(
        "rounded-lg border border-gray-200 bg-background p-6 dark:border-gray-800",
        className
      )}
      {...props}
    >
      {children}
    </div>
  );
}
