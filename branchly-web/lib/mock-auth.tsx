"use client";

import { mockUser } from "@/lib/mock-data";
import type { User } from "@/types";
import { createContext, useContext, useMemo, type ReactNode } from "react";

export type MockAuthContextValue = {
  user: User;
  isAuthenticated: boolean;
};

const MockAuthContext = createContext<MockAuthContextValue | null>(null);

export function MockAuthProvider({ children }: { children: ReactNode }) {
  const value = useMemo<MockAuthContextValue>(
    () => ({
      user: mockUser,
      isAuthenticated: true,
    }),
    []
  );
  return (
    <MockAuthContext.Provider value={value}>{children}</MockAuthContext.Provider>
  );
}

export function useMockAuth(): MockAuthContextValue {
  const ctx = useContext(MockAuthContext);
  if (!ctx) {
    return { user: mockUser, isAuthenticated: true };
  }
  return ctx;
}
