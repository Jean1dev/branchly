"use client";

import { cn } from "@/lib/utils";
import {
  FolderGit2,
  LayoutDashboard,
  ListChecks,
  Settings,
} from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";

const main = [
  { href: "/dashboard", label: "Overview", icon: LayoutDashboard },
  { href: "/repositories", label: "Repositories", icon: FolderGit2 },
  { href: "/jobs", label: "Jobs", icon: ListChecks },
];

export function Sidebar() {
  const pathname = usePathname();
  return (
    <aside
      className="fixed bottom-0 left-0 top-14 z-30 flex w-14 flex-col border-r border-gray-200 bg-background md:w-[220px] dark:border-gray-800"
      aria-label="Main navigation"
    >
      <nav className="flex flex-1 flex-col gap-1 p-2 md:p-3">
        {main.map(({ href, label, icon: Icon }) => {
          const active =
            href === "/dashboard"
              ? pathname === "/dashboard"
              : pathname.startsWith(href);
          return (
            <Link
              key={href}
              href={href}
              className={cn(
                "flex items-center gap-3 rounded-md px-2 py-2 text-sm font-medium text-gray-600 transition-colors duration-150 hover:bg-gray-100 hover:text-foreground dark:text-gray-400 dark:hover:bg-gray-900 dark:hover:text-gray-50",
                active &&
                  "bg-gray-100 text-foreground dark:bg-gray-900 dark:text-gray-50"
              )}
            >
              <Icon className="h-5 w-5 shrink-0" aria-hidden />
              <span className="hidden md:inline">{label}</span>
            </Link>
          );
        })}
      </nav>
      <div className="border-t border-gray-200 p-2 dark:border-gray-800 md:p-3">
        <Link
          href="/settings"
          className={cn(
            "flex items-center gap-3 rounded-md px-2 py-2 text-sm font-medium text-gray-600 transition-colors duration-150 hover:bg-gray-100 hover:text-foreground dark:text-gray-400 dark:hover:bg-gray-900 dark:hover:text-gray-50",
            pathname.startsWith("/settings") &&
              "bg-gray-100 text-foreground dark:bg-gray-900 dark:text-gray-50"
          )}
        >
          <Settings className="h-5 w-5 shrink-0" aria-hidden />
          <span className="hidden md:inline">Settings</span>
        </Link>
      </div>
    </aside>
  );
}
