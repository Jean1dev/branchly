"use client";

import { Button } from "@/components/ui/button";
import { GitHubStars } from "@/components/features/github-stars";
import { ChevronRight } from "lucide-react";
import Link from "next/link";
import { useSession } from "next-auth/react";

const GITHUB_URL = "https://github.com/Jean1dev/branchly";

type NavbarMarketingProps = {
  transparent?: boolean;
};

function GitHubIcon({ className }: { className?: string }) {
  return (
    <svg
      className={className}
      viewBox="0 0 24 24"
      fill="currentColor"
      aria-hidden
    >
      <path d="M12 2C6.477 2 2 6.477 2 12c0 4.42 2.865 8.166 6.839 9.489.5.092.682-.217.682-.482 0-.237-.008-.866-.013-1.7-2.782.603-3.369-1.342-3.369-1.342-.454-1.155-1.11-1.462-1.11-1.462-.908-.62.069-.608.069-.608 1.003.07 1.531 1.03 1.531 1.03.892 1.529 2.341 1.087 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.11-4.555-4.943 0-1.091.39-1.984 1.029-2.683-.103-.253-.446-1.27.098-2.647 0 0 .84-.269 2.75 1.025A9.578 9.578 0 0112 6.836c.85.004 1.705.114 2.504.336 1.909-1.294 2.747-1.025 2.747-1.025.546 1.377.202 2.394.1 2.647.64.699 1.028 1.592 1.028 2.683 0 3.842-2.339 4.687-4.566 4.935.359.309.678.919.678 1.852 0 1.336-.012 2.415-.012 2.741 0 .267.18.578.688.48C19.138 20.163 22 16.418 22 12c0-5.523-4.477-10-10-10z" />
    </svg>
  );
}

export function NavbarMarketing({ transparent }: NavbarMarketingProps) {
  const { status } = useSession();
  const isAuthenticated = status === "authenticated";

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

        <div className="flex items-center gap-2">
          <Link
            href={GITHUB_URL}
            target="_blank"
            rel="noopener noreferrer"
            className="flex items-center gap-1.5 rounded-lg border border-gray-200 bg-gray-50 px-3 py-1.5 font-mono text-xs text-gray-600 transition-colors hover:border-gray-300 hover:text-gray-900 dark:border-gray-800 dark:bg-gray-950 dark:text-gray-400 dark:hover:border-gray-700 dark:hover:text-gray-200"
          >
            <GitHubIcon className="h-3.5 w-3.5" />
            GitHub
          </Link>
          <GitHubStars />
          {isAuthenticated ? (
            <Button variant="ghost" size="sm" href="/dashboard">
              Dashboard
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
