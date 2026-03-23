"use client";

import { Button } from "@/components/ui/button";
import { useMockAuth } from "@/lib/mock-auth";
import { ChevronRight } from "lucide-react";
import Link from "next/link";

type NavbarMarketingProps = {
  transparent?: boolean;
};

export function NavbarMarketing({ transparent }: NavbarMarketingProps) {
  const { isAuthenticated } = useMockAuth();
  return (
    <header
      className={
        transparent
          ? "fixed left-0 right-0 top-0 z-40 h-14 border-b border-transparent bg-transparent"
          : "fixed left-0 right-0 top-0 z-40 h-14 border-b border-gray-200 bg-background dark:border-gray-800"
      }
    >
      <div className="mx-auto flex h-full max-w-[1200px] items-center justify-between px-4 sm:px-6">
        <Link
          href="/"
          className="font-mono text-base font-semibold tracking-tight"
        >
          branchly
        </Link>
        <div className="flex items-center gap-3">
          {isAuthenticated ? (
            <Button variant="ghost" size="sm" href="/dashboard">
              Go to dashboard
            </Button>
          ) : (
            <Button variant="ghost" size="sm" href="/login">
              Sign in
            </Button>
          )}
          <Button size="sm" href="/login" className="gap-1">
            Get started
            <ChevronRight className="h-4 w-4" aria-hidden />
          </Button>
        </div>
      </div>
    </header>
  );
}
