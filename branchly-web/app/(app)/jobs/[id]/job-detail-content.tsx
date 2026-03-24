import { JobLogPanel } from "@/components/features/job-log-panel";
import { StatusBadge } from "@/components/features/status-badge";
import { Card } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { apiFetch } from "@/lib/api-client";
import {
  mapJob,
  mapJobLog,
  unwrapApiData,
  type ApiJob,
} from "@/lib/map-api";
import { formatDate, truncate } from "@/lib/utils";
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
    ? unwrapApiData<Array<{ id: string; full_name: string }>>(
        await repoRes.json()
      )
    : [];
  const repos = Array.isArray(reposParsed) ? reposParsed : [];
  const repoName = repos.find((r) => r.id === raw.repository_id)?.full_name;
  const job = mapJob(raw, repoName);
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
        <p className="font-mono text-sm text-gray-500 dark:text-gray-400">
          {job.repositoryName}
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
        </div>
      </div>
    </>
  );
}
