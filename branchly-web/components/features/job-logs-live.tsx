"use client";

import { LogTerminal } from "@/components/features/log-terminal";
import { mapJobLog } from "@/lib/map-api";
import type { JobLog } from "@/types";
import { useEffect, useState } from "react";

type JobLogsLiveProps = {
  jobId: string;
};

export function JobLogsLive({ jobId }: JobLogsLiveProps) {
  const [lines, setLines] = useState<JobLog[]>([]);

  useEffect(() => {
    const es = new EventSource(`/api/jobs/${encodeURIComponent(jobId)}/logs`);

    es.onmessage = (ev) => {
      try {
        const p = JSON.parse(ev.data) as {
          timestamp: string;
          level: string;
          message: string;
        };
        setLines((prev) => [...prev, mapJobLog(p)]);
      } catch {
        return;
      }
    };

    es.addEventListener("done", () => {
      es.close();
    });

    es.onerror = () => {
      es.close();
    };

    return () => es.close();
  }, [jobId]);

  return <LogTerminal lines={lines} stream={false} />;
}
