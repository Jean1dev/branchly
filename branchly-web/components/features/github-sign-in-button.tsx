"use client";

import { GitHubMark } from "@/components/icons/github-mark";
import { Button } from "@/components/ui/button";
import { signIn } from "next-auth/react";

export function GitHubSignInButton() {
  return (
    <Button
      type="button"
      className="mt-8 w-full gap-2"
      size="lg"
      onClick={() => signIn("github", { callbackUrl: "/dashboard" })}
    >
      <GitHubMark className="h-5 w-5" />
      Continue with GitHub
    </Button>
  );
}
