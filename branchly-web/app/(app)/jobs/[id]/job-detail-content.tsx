import { JobThread } from "@/components/features/job-thread";
import { ProviderBadge } from "@/components/ui/provider-badge";
import { apiFetch } from "@/lib/api-client";
import {
  mapJob,
  mapJobLog,
  unwrapApiData,
  type ApiJob,
  type ApiRepository,
} from "@/lib/map-api";
import { truncate } from "@/lib/utils";
import Link from "next/link";
import { notFound } from "next/navigation";

export async function JobDetailContent({ id }: { id: string }) {
  // ── 1. Fetch the full thread ──────────────────────────────────────────────
  const threadRes = await apiFetch(`/jobs/${encodeURIComponent(id)}/thread`);
  if (threadRes.status === 404) notFound();
  if (!threadRes.ok) notFound();

  const rawThread = unwrapApiData<ApiJob[]>(await threadRes.json());
  const threadJobs = Array.isArray(rawThread) ? rawThread : [rawThread as ApiJob];

  // ── 2. Fetch latest job with logs (GET /jobs/:id returns tail logs) ───────
  const latestRaw = threadJobs[threadJobs.length - 1];
  const latestRes = await apiFetch(`/jobs/${encodeURIComponent(latestRaw.id)}`);
  const latestFull = latestRes.ok
    ? unwrapApiData<ApiJob>(await latestRes.json())
    : latestRaw;

  // ── 3. Repository name / provider for display ─────────────────────────────
  const repoRes = await apiFetch("/repositories");
  const reposParsed = repoRes.ok
    ? unwrapApiData<ApiRepository[]>(await repoRes.json())
    : [];
  const repos = Array.isArray(reposParsed) ? reposParsed : [];
  const matchedRepo = repos.find((r) => r.id === latestRaw.repository_id);

  // ── 4. Map to domain types ────────────────────────────────────────────────
  const jobs = threadJobs.map((j) =>
    mapJob(j, matchedRepo?.full_name, matchedRepo?.provider)
  );
  const initialLastJobLogs = (latestFull.logs ?? []).map(mapJobLog);

  const rootJob = jobs[0];

  return (
    <>
      {/* Breadcrumb */}
      <nav className="mb-6 text-sm text-gray-500 dark:text-gray-400" aria-label="Breadcrumb">
        <ol className="flex flex-wrap items-center gap-2">
          <li>
            <Link href="/jobs" className="hover:text-foreground">
              Jobs
            </Link>
          </li>
          <li aria-hidden>/</li>
          <li className="font-mono text-foreground">{rootJob.id}</li>
        </ol>
      </nav>

      {/* Header */}
      <header className="mb-8 space-y-2">
        <p className="flex items-center gap-1.5 font-mono text-sm text-gray-500 dark:text-gray-400">
          <ProviderBadge provider={rootJob.repositoryProvider} />
          <span>{rootJob.repositoryName}</span>
          {jobs.length > 1 ? (
            <span className="ml-2 rounded-full bg-gray-100 px-2 py-0.5 text-xs font-medium text-gray-600 dark:bg-gray-900 dark:text-gray-400">
              {jobs.length} turns
            </span>
          ) : null}
        </p>
        <h1 className="text-xl font-semibold tracking-tight md:text-2xl">
          {truncate(rootJob.prompt, 100)}
        </h1>
      </header>

      {/* Thread */}
      <JobThread initialJobs={jobs} initialLastJobLogs={initialLastJobLogs} />
    </>
  );
}
