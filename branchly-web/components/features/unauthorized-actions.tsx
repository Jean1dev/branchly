"use client";

import { Button } from "@/components/ui/button";
import { signOut } from "next-auth/react";

type UnauthorizedActionsProps = {
  isSignedIn: boolean;
};

export function UnauthorizedActions({ isSignedIn }: UnauthorizedActionsProps) {
  if (!isSignedIn) {
    return (
      <Button href="/login" size="lg">
        Sign in
      </Button>
    );
  }

  return (
    <div className="flex flex-col gap-3 sm:flex-row">
      <Button
        type="button"
        variant="secondary"
        size="lg"
        onClick={() => void signOut({ callbackUrl: "/login" })}
      >
        Sign out and sign in again
      </Button>
      <Button href="/login" variant="secondary" size="lg">
        Back to login
      </Button>
    </div>
  );
}
