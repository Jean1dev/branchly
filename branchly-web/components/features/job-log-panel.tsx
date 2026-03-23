"use client";

import { LogTerminal } from "@/components/features/log-terminal";
import type { JobLog, JobStatus } from "@/types";

type JobLogPanelProps = {
  lines: JobLog[];
  status: JobStatus;
};

export function JobLogPanel({ lines, status }: JobLogPanelProps) {
  return (
    <LogTerminal
      lines={lines}
      stream={status === "running"}
      intervalMs={400}
    />
  );
}
