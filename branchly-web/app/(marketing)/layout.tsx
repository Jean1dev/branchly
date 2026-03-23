import { MarketingChrome } from "@/components/layout/marketing-chrome";
import type { ReactNode } from "react";

export default function MarketingLayout({ children }: { children: ReactNode }) {
  return <MarketingChrome>{children}</MarketingChrome>;
}
