"use client";

import { Separator } from "@/components/ui/separator";
import { useMockAuth } from "@/lib/mock-auth";
import { ChevronDown, LogOut, User } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { ThemeToggle } from "./theme-toggle";

const links = [
  { href: "/dashboard", label: "Dashboard" },
  { href: "/repositories", label: "Repositories" },
  { href: "/jobs", label: "Jobs" },
];

export function NavbarApp() {
  const pathname = usePathname();
  const { user } = useMockAuth();
  const [open, setOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    function onDoc(e: MouseEvent) {
      if (!menuRef.current?.contains(e.target as Node)) setOpen(false);
    }
    document.addEventListener("mousedown", onDoc);
    return () => document.removeEventListener("mousedown", onDoc);
  }, []);

  return (
    <header className="fixed left-0 right-0 top-0 z-40 h-14 border-b border-gray-200 bg-background dark:border-gray-800">
      <div className="flex h-full items-center justify-between gap-4 px-4 lg:px-6">
        <div className="flex min-w-0 flex-1 items-center gap-6">
          <Link
            href="/dashboard"
            className="shrink-0 font-mono text-base font-semibold tracking-tight"
          >
            branchly
          </Link>
          <nav
            className="hidden items-center gap-1 md:flex"
            aria-label="Primary"
          >
            {links.map(({ href, label }) => {
              const active =
                href === "/dashboard"
                  ? pathname === "/dashboard"
                  : pathname.startsWith(href);
              return (
                <Link
                  key={href}
                  href={href}
                  className={
                    active
                      ? "rounded-md bg-gray-100 px-3 py-1.5 text-sm font-medium text-foreground dark:bg-gray-900"
                      : "rounded-md px-3 py-1.5 text-sm font-medium text-gray-500 transition-colors duration-150 hover:text-foreground dark:text-gray-400"
                  }
                >
                  {label}
                </Link>
              );
            })}
          </nav>
        </div>
        <div className="flex shrink-0 items-center gap-2">
          <ThemeToggle />
          <div className="relative" ref={menuRef}>
            <button
              type="button"
              className="flex items-center gap-2 rounded-md p-1 transition-colors duration-150 hover:bg-gray-100 dark:hover:bg-gray-900"
              aria-expanded={open}
              aria-haspopup="menu"
              aria-label="User menu"
              onClick={() => setOpen((o) => !o)}
            >
              <Image
                src={user.avatar}
                alt=""
                width={32}
                height={32}
                className="rounded-full"
              />
              <ChevronDown className="hidden h-4 w-4 text-gray-500 sm:block" />
            </button>
            {open ? (
              <div
                className="absolute right-0 mt-2 w-52 rounded-lg border border-gray-200 bg-background py-1 dark:border-gray-800"
                role="menu"
              >
                <div className="px-3 py-2">
                  <p className="truncate text-sm font-medium">{user.name}</p>
                  <p className="truncate text-xs text-gray-500 dark:text-gray-400">
                    {user.email}
                  </p>
                </div>
                <Separator className="my-1" />
                <Link
                  href="/settings"
                  className="flex items-center gap-2 px-3 py-2 text-sm hover:bg-gray-100 dark:hover:bg-gray-900"
                  role="menuitem"
                  onClick={() => setOpen(false)}
                >
                  <User className="h-4 w-4" />
                  Settings
                </Link>
                <button
                  type="button"
                  className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm hover:bg-gray-100 dark:hover:bg-gray-900"
                  role="menuitem"
                  onClick={() => setOpen(false)}
                >
                  <LogOut className="h-4 w-4" />
                  Sign out
                </button>
              </div>
            ) : null}
          </div>
        </div>
      </div>
    </header>
  );
}
