"use client";

import { Separator } from "@/components/ui/separator";
import { ChevronDown, LogOut, User } from "lucide-react";
import Image from "next/image";
import Link from "next/link";
import { signOut, useSession } from "next-auth/react";
import { useEffect, useRef, useState } from "react";
import { ThemeToggle } from "./theme-toggle";

export function NavbarApp() {
  const { data: session, status } = useSession();
  const [open, setOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement | null>(null);

  const name = session?.user?.name ?? "User";
  const email = session?.user?.email ?? "";
  const avatarSeed = encodeURIComponent(name);
  const avatarSrc =
    session?.user?.image ??
    `https://api.dicebear.com/7.x/initials/svg?seed=${avatarSeed}`;

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
        <div className="flex min-w-0 flex-1 items-center">
          <Link
            href="/dashboard"
            className="shrink-0 font-mono text-base font-semibold tracking-tight"
          >
            branchly
          </Link>
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
              {status === "loading" ? (
                <span className="h-8 w-8 shrink-0 rounded-full bg-gray-200 dark:bg-gray-800" />
              ) : (
                <Image
                  src={avatarSrc}
                  alt=""
                  width={32}
                  height={32}
                  className="rounded-full"
                  unoptimized={avatarSrc.includes("dicebear.com")}
                />
              )}
              <ChevronDown className="hidden h-4 w-4 text-gray-500 sm:block" />
            </button>
            {open ? (
              <div
                className="absolute right-0 mt-2 w-52 rounded-lg border border-gray-200 bg-background py-1 dark:border-gray-800"
                role="menu"
              >
                <div className="px-3 py-2">
                  <p className="truncate text-sm font-medium">{name}</p>
                  <p className="truncate text-xs text-gray-500 dark:text-gray-400">
                    {email}
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
                  onClick={() => {
                    setOpen(false);
                    void signOut({ callbackUrl: "/login" });
                  }}
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
