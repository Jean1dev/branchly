"use client";

import { SessionProvider, signOut, useSession } from "next-auth/react";
import type { ReactNode } from "react";
import { useEffect } from "react";

function SessionWatcher() {
  const { data: session, update } = useSession();

  useEffect(() => {
    if (!session) return;

    // Force sign out when token refresh has permanently failed
    if (session.error === "RefreshAccessTokenError") {
      signOut({ callbackUrl: "/login" });
      return;
    }

    const expiry = session.internalTokenExpiry;
    if (!expiry) return;

    const refreshBuffer = 5 * 60 * 1000; // refresh 5 min before expiry
    const msUntilRefresh = expiry - refreshBuffer - Date.now();

    if (msUntilRefresh <= 0) {
      // Already expired or within buffer — refresh immediately
      update();
      return;
    }

    const timer = setTimeout(() => update(), msUntilRefresh);
    return () => clearTimeout(timer);
  }, [session, update]);

  return null;
}

export function AuthSessionProvider({ children }: { children: ReactNode }) {
  return (
    <SessionProvider>
      <SessionWatcher />
      {children}
    </SessionProvider>
  );
}
