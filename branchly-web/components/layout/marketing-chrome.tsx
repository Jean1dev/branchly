"use client";

import type { ReactNode } from "react";
import { usePathname } from "next/navigation";
import { NavbarMarketing } from "./navbar-marketing";

export function MarketingChrome({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const transparent = pathname === "/";
  return (
    <>
      <NavbarMarketing transparent={transparent} />
      {children}
    </>
  );
}
