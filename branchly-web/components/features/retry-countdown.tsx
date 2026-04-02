"use client";

import { useEffect, useState } from "react";

function formatCountdown(ms: number): string {
  if (ms <= 0) return "Retrying soon…";
  const totalSecs = Math.ceil(ms / 1000);
  const mins = Math.floor(totalSecs / 60);
  const secs = totalSecs % 60;
  if (mins > 0) {
    return `in ${mins}m ${secs}s`;
  }
  return `in ${secs}s`;
}

export function RetryCountdown({ nextRetryAt }: { nextRetryAt: string }) {
  const [remaining, setRemaining] = useState(() => new Date(nextRetryAt).getTime() - Date.now());

  useEffect(() => {
    const id = setInterval(() => {
      const ms = new Date(nextRetryAt).getTime() - Date.now();
      setRemaining(ms);
    }, 1000);
    return () => clearInterval(id);
  }, [nextRetryAt]);

  return <span>{formatCountdown(remaining)}</span>;
}
