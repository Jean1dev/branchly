import { cn } from "@/lib/utils";
import type { ReactNode } from "react";

type PageLayoutProps = {
  children: ReactNode;
  className?: string;
};

export function PageLayout({ children, className }: PageLayoutProps) {
  return (
    <div className={cn("mx-auto w-full max-w-[1200px] px-4 sm:px-0", className)}>
      {children}
    </div>
  );
}
