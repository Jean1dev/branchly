"use client";

import { Button } from "@/components/ui/button";
import { useRouter } from "next/navigation";
import { useState } from "react";

export function RetryButton({ jobId }: { jobId: string }) {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  async function handleRetry() {
    setLoading(true);
    setError(null);
    try {
      const res = await fetch(`/api/jobs/${jobId}/retry`, { method: "POST" });
      if (res.status === 409) {
        setError("This job cannot be retried.");
        return;
      }
      if (!res.ok) {
        setError("Failed to retry job. Please try again.");
        return;
      }
      router.refresh();
    } catch {
      setError("Failed to retry job. Please try again.");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="space-y-2">
      <Button
        type="button"
        variant="secondary"
        size="sm"
        disabled={loading}
        onClick={handleRetry}
        className="w-full"
      >
        {loading ? "Retrying…" : "↺ Retry this job"}
      </Button>
      {error ? (
        <p className="text-xs text-red-500">{error}</p>
      ) : null}
    </div>
  );
}
