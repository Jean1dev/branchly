"use client";

import { JobLogsLive } from "@/components/features/job-logs-live";
import { LogTerminal } from "@/components/features/log-terminal";
import type { JobLog, JobStatus } from "@/types";

type JobLogPanelProps = {
  jobId: string;
  lines: JobLog[];
  status: JobStatus;
};

export function JobLogPanel({ jobId, lines, status }: JobLogPanelProps) {
  const live = status === "running" || status === "pending";
  if (live) {
    return <JobLogsLive jobId={jobId} />;
  }
  return <LogTerminal lines={lines} stream={false} />;
}
