"use client";

import { JobCostCard } from "@/components/features/job-cost-card";
import { JobLogsLive } from "@/components/features/job-logs-live";
import { LogTerminal } from "@/components/features/log-terminal";
import { StatusBadge } from "@/components/features/status-badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { mapJob, mapJobLog, unwrapApiData, type ApiJob } from "@/lib/map-api";
import { formatDate, formatDurationMs } from "@/lib/utils";
import type { Job, JobLog } from "@/types";
import { useRouter } from "next/navigation";
import { useRef, useState } from "react";

// ──────────────────────────────────────────────────────────────────────────────
// Types
// ──────────────────────────────────────────────────────────────────────────────

type TurnProps = {
  job: Job;
  initialLogs: JobLog[];
  isLatest: boolean;
};

type ThreadProps = {
  initialJobs: Job[];
  /** Pre-fetched logs for the latest job (avoids an extra client-side request). */
  initialLastJobLogs: JobLog[];
};

// ──────────────────────────────────────────────────────────────────────────────
// Single turn
// ──────────────────────────────────────────────────────────────────────────────

function ThreadTurn({ job, initialLogs, isLatest }: TurnProps) {
  const [logsOpen, setLogsOpen] = useState(isLatest);
  const isLive = job.status === "running" || job.status === "pending";

  return (
    <div className="group relative flex gap-4">
      {/* Turn spine */}
      <div className="flex flex-col items-center">
        <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full border border-gray-200 bg-white text-xs font-semibold text-gray-500 dark:border-gray-800 dark:bg-gray-950">
          {job.threadPosition + 1}
        </div>
        <div className="mt-1 w-px flex-1 bg-gray-200 dark:bg-gray-800 group-last:hidden" />
      </div>

      {/* Turn body */}
      <div className="mb-8 min-w-0 flex-1">
        {/* User prompt bubble */}
        <div className="mb-3 inline-block max-w-prose rounded-2xl rounded-tl-sm bg-gray-100 px-4 py-2.5 text-sm dark:bg-gray-900">
          <p className="whitespace-pre-wrap break-words">{job.prompt}</p>
        </div>

        {/* Agent result card */}
        <div className="rounded-lg border border-gray-200 dark:border-gray-800">
          {/* Header row: status + branch */}
          <div className="flex flex-wrap items-center gap-3 px-4 py-3">
            <StatusBadge status={job.status} />
            {job.branchName ? (
              <span className="font-mono text-xs text-gray-500 dark:text-gray-400">
                {job.branchName}
              </span>
            ) : null}
          </div>

          {/* Meta row: dates, duration, error */}
          <div className="flex flex-wrap items-center gap-x-6 gap-y-1 border-t border-gray-200 px-4 py-2 text-xs text-gray-500 dark:border-gray-800 dark:text-gray-400">
            <span>{formatDate(job.createdAt)}</span>
            {job.completedAt ? (
              <span>Duration: {formatDurationMs(job.createdAt, job.completedAt)}</span>
            ) : null}
            {job.lastError ? (
              <span className="text-red-500 dark:text-red-400" title={job.lastError}>
                Error: {job.lastError.slice(0, 100)}{job.lastError.length > 100 ? "…" : ""}
              </span>
            ) : null}
          </div>

          {/* Pull request row */}
          <div className="border-t border-gray-200 px-4 py-3 dark:border-gray-800">
            <p className="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
              Pull request
            </p>
            {job.prUrl ? (
              <a
                href={job.prUrl}
                target="_blank"
                rel="noreferrer"
                className="mt-1 inline-block text-sm font-medium hover:underline"
              >
                Open pull request ↗
              </a>
            ) : (
              <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                {job.status === "completed" || job.status === "failed"
                  ? "Not generated"
                  : "Pending…"}
              </p>
            )}
          </div>

          {/* Log toggle */}
          <div className="border-t border-gray-200 dark:border-gray-800">
            <button
              type="button"
              onClick={() => setLogsOpen((v) => !v)}
              className="flex w-full items-center gap-2 px-4 py-2 text-left text-xs font-medium text-gray-500 transition-colors hover:bg-gray-50 hover:text-foreground dark:text-gray-400 dark:hover:bg-gray-900"
            >
              <span
                className={`transition-transform ${logsOpen ? "rotate-90" : ""}`}
                aria-hidden
              >
                ▶
              </span>
              {logsOpen ? "Hide" : "Show"} logs
            </button>

            {logsOpen ? (
              <div className="px-4 pb-4">
                {isLive ? (
                  <JobLogsLive key={job.id} jobId={job.id} initialLines={initialLogs} />
                ) : (
                  <LogTerminal lines={initialLogs} stream={false} />
                )}
              </div>
            ) : null}
          </div>
        </div>

        {/* Cost card — full detail, shown once available */}
        {job.cost ? (
          <div className="mt-3">
            <JobCostCard cost={job.cost} />
          </div>
        ) : job.status === "completed" || job.status === "failed" ? (
          <Card className="mt-3 animate-pulse space-y-4 p-6">
            <div className="h-3 w-1/2 rounded bg-gray-200 dark:bg-gray-700" />
            <div className="h-7 w-1/3 rounded bg-gray-200 dark:bg-gray-700" />
            <Separator />
            <div className="grid grid-cols-2 gap-3">
              {[...Array(4)].map((_, i) => (
                <div key={i} className="space-y-1">
                  <div className="h-2.5 w-2/3 rounded bg-gray-200 dark:bg-gray-700" />
                  <div className="h-4 w-1/2 rounded bg-gray-200 dark:bg-gray-700" />
                </div>
              ))}
            </div>
          </Card>
        ) : null}
      </div>
    </div>
  );
}

// ──────────────────────────────────────────────────────────────────────────────
// Follow-up form
// ──────────────────────────────────────────────────────────────────────────────

type FollowupFormProps = {
  parentJob: Job;
  onJobCreated: (job: Job) => void;
};

function FollowupForm({ parentJob, onJobCreated }: FollowupFormProps) {
  const [prompt, setPrompt] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const trimmed = prompt.trim();
    if (!trimmed) return;

    setLoading(true);
    setError(null);
    try {
      const res = await fetch("/api/jobs", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          repository_id: parentJob.repositoryId,
          prompt: trimmed,
          agent_type: parentJob.agentType,
          parent_job_id: parentJob.id,
        }),
      });
      const json = await res.json();
      if (!res.ok) {
        const msg =
          json?.error?.message ??
          (res.status === 429
            ? "Too many active jobs — wait for one to finish."
            : "Failed to start job.");
        setError(msg);
        return;
      }
      const newJob = mapJob(
        unwrapApiData<ApiJob>(json),
        parentJob.repositoryName,
        parentJob.repositoryProvider
      );
      onJobCreated(newJob);
      setPrompt("");
    } catch {
      setError("Something went wrong. Please try again.");
    } finally {
      setLoading(false);
    }
  }

  function handleKeyDown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
      void handleSubmit(e as unknown as React.FormEvent);
    }
  }

  return (
    <div className="flex gap-4">
      {/* Align with turns */}
      <div className="w-7 shrink-0" />
      <form onSubmit={handleSubmit} className="flex-1 space-y-2">
        <textarea
          ref={textareaRef}
          value={prompt}
          onChange={(e) => setPrompt(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Continue this task… (⌘Enter to send)"
          rows={3}
          disabled={loading}
          className="w-full resize-none rounded-lg border border-gray-200 bg-white px-4 py-3 text-sm placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-black/10 disabled:opacity-50 dark:border-gray-800 dark:bg-gray-950 dark:placeholder-gray-600 dark:focus:ring-white/10"
        />
        {error ? (
          <p className="text-xs text-red-500">{error}</p>
        ) : null}
        <div className="flex justify-end">
          <Button
            type="submit"
            disabled={loading || !prompt.trim()}
            size="sm"
          >
            {loading ? "Starting…" : "Send follow-up"}
          </Button>
        </div>
      </form>
    </div>
  );
}

// ──────────────────────────────────────────────────────────────────────────────
// Thread root
// ──────────────────────────────────────────────────────────────────────────────

export function JobThread({ initialJobs, initialLastJobLogs }: ThreadProps) {
  const router = useRouter();
  const [jobs, setJobs] = useState<Job[]>(initialJobs);

  const latestJob = jobs[jobs.length - 1];
  const canContinue =
    latestJob.status === "completed" || latestJob.status === "failed";

  function handleJobCreated(newJob: Job) {
    setJobs((prev) => [...prev, newJob]);
    // Refresh page data in the background so the server component re-validates.
    router.refresh();
  }

  return (
    <div>
      {jobs.map((job, idx) => {
        const isLatest = idx === jobs.length - 1;
        return (
          <ThreadTurn
            key={job.id}
            job={job}
            initialLogs={isLatest ? initialLastJobLogs : []}
            isLatest={isLatest}
          />
        );
      })}

      {canContinue ? (
        <FollowupForm parentJob={latestJob} onJobCreated={handleJobCreated} />
      ) : null}
    </div>
  );
}
