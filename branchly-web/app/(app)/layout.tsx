import { NavbarApp } from "@/components/layout/navbar-app";
import { PageLayout } from "@/components/layout/page-layout";
import { Sidebar } from "@/components/layout/sidebar";
import type { ReactNode } from "react";

export const dynamic = "force-dynamic";

export default function AppGroupLayout({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-screen bg-background">
      <NavbarApp />
      <Sidebar />
      <main className="ml-14 min-h-screen pt-14 md:ml-[220px]">
        <div className="p-8">
          <PageLayout>{children}</PageLayout>
        </div>
      </main>
    </div>
  );
}
