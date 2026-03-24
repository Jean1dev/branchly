"use client";

import { Button } from "@/components/ui/button";
import { signOut } from "next-auth/react";

export function DisconnectAccountButton() {
  return (
    <Button
      type="button"
      variant="destructive"
      onClick={() => void signOut({ callbackUrl: "/login" })}
    >
      Disconnect account
    </Button>
  );
}
