import { JobCostCard } from "@/components/features/job-cost-card";
import { JobLogPanel } from "@/components/features/job-log-panel";
import { StatusBadge } from "@/components/features/status-badge";
import { Card } from "@/components/ui/card";
import { ProviderBadge } from "@/components/ui/provider-badge";
import { Separator } from "@/components/ui/separator";
import { apiFetch } from "@/lib/api-client";
import {
  mapJob,
  mapJobLog,
  unwrapApiData,
  type ApiJob,
  type ApiRepository,
} from "@/lib/map-api";
import { formatDate, truncate } from "@/lib/utils";
import { AGENTS } from "@/types";
import Link from "next/link";
import { notFound } from "next/navigation";

export async function JobDetailContent({ id }: { id: string }) {
  const res = await apiFetch(`/jobs/${encodeURIComponent(id)}`);
  if (res.status === 404) {
    notFound();
  }
  if (!res.ok) {
    notFound();
  }
  const raw = unwrapApiData<ApiJob>(await res.json());
  const repoRes = await apiFetch("/repositories");
  const reposParsed = repoRes.ok
    ? unwrapApiData<ApiRepository[]>(await repoRes.json())
    : [];
  const repos = Array.isArray(reposParsed) ? reposParsed : [];
  const matchedRepo = repos.find((r) => r.id === raw.repository_id);
  const job = mapJob(raw, matchedRepo?.full_name, matchedRepo?.provider);
  const logLines = (raw.logs ?? []).map(mapJobLog);

  return (
    <>
      <nav className="mb-6 text-sm text-gray-500 dark:text-gray-400" aria-label="Breadcrumb">
        <ol className="flex flex-wrap items-center gap-2">
          <li>
            <Link href="/jobs" className="hover:text-foreground">
              Jobs
            </Link>
          </li>
          <li aria-hidden>/</li>
          <li className="font-mono text-foreground">{job.id}</li>
        </ol>
      </nav>
      <header className="mb-8 space-y-3">
        <p className="flex items-center gap-1.5 font-mono text-sm text-gray-500 dark:text-gray-400">
          <ProviderBadge provider={job.repositoryProvider} />
          <span>{job.repositoryName}</span>
        </p>
        <div className="flex flex-wrap items-center gap-3">
          <h1 className="text-xl font-semibold tracking-tight md:text-2xl">
            {truncate(job.prompt, 100)}
          </h1>
          <StatusBadge status={job.status} />
        </div>
      </header>
      <div className="grid gap-8 lg:grid-cols-5 lg:items-start">
        <div className="lg:col-span-3">
          <h2 className="mb-3 text-sm font-medium text-gray-500 dark:text-gray-400">
            Log output
          </h2>
          <JobLogPanel jobId={job.id} lines={logLines} status={job.status} />
        </div>
        <div className="lg:col-span-2">
          <Card className="space-y-4 p-6">
            <div>
              <p className="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                Agent
              </p>
              {(() => {
                const agent = AGENTS.find((a) => a.id === job.agentType) ?? AGENTS[0];
                return (
                  <p className="mt-1 text-sm font-medium">
                    {agent.name}{" "}
                    <span className="font-normal text-gray-500 dark:text-gray-400">
                      · {agent.provider}
                    </span>
                  </p>
                );
              })()}
            </div>
            <Separator />
            <div>
              <p className="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                Status
              </p>
              <p className="mt-1 text-sm font-medium capitalize">{job.status}</p>
            </div>
            <Separator />
            <div>
              <p className="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                Branch
              </p>
              <p className="mt-1 font-mono text-sm">{job.branchName}</p>
            </div>
            <Separator />
            <div>
              <p className="text-xs font-medium uppercase tracking-wide text-gray-500 dark:text-gray-400">
                Pull request
              </p>
              {job.prUrl ? (
                <a
                  href={job.prUrl}
                  className="mt-1 inline-block text-sm font-medium hover:underline"
                  target="_blank"
                  rel="noreferrer"
                >
                  Open pull request
                </a>
              ) : (
                <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                  Not available yet
                </p>
              )}
            </div>
            <Separator />
            <div>
              <p className="text-xs text-gray-500 dark:text-gray-400">Created</p>
              <p className="mt-1 text-sm">{formatDate(job.createdAt)}</p>
            </div>
            {job.completedAt ? (
              <div>
                <p className="text-xs text-gray-500 dark:text-gray-400">Completed</p>
                <p className="mt-1 text-sm">{formatDate(job.completedAt)}</p>
              </div>
            ) : null}
          </Card>

          {job.cost ? (
            <JobCostCard cost={job.cost} />
          ) : job.status === "completed" || job.status === "failed" ? (
            <Card className="animate-pulse space-y-4 p-6">
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
    </>
  );
}
