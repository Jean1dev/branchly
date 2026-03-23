import { cn } from "@/lib/utils";
import type { HTMLAttributes } from "react";

export function Separator({
  className,
  ...props
}: HTMLAttributes<HTMLHRElement>) {
  return (
    <hr
      className={cn("h-px w-full border-0 bg-gray-200 dark:bg-gray-800", className)}
      {...props}
    />
  );
}
