"use client";

import { Button } from "@/components/ui/button";
import { useEffect } from "react";

export default function AppError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <div className="mx-auto flex min-h-[40vh] max-w-lg flex-col justify-center gap-6 py-12">
      <div className="space-y-2">
        <h1 className="text-xl font-semibold">Something went wrong</h1>
        <p className="text-sm text-gray-600 dark:text-gray-400">
          {error.message || "An unexpected error occurred while loading this page."}
        </p>
      </div>
      <div className="flex flex-wrap gap-3">
        <Button type="button" onClick={reset}>
          Try again
        </Button>
        <Button href="/dashboard" variant="secondary">
          Go to dashboard
        </Button>
      </div>
    </div>
  );
}
