import { RepoCard } from "@/components/features/repo-card";
import { StatusBadge } from "@/components/features/status-badge";
import { EmptyState } from "@/components/features/empty-state";
import { PageHeader } from "@/components/layout/page-header";
import { Card } from "@/components/ui/card";
import { apiFetch } from "@/lib/api-client";
import {
  jobRepoNameMap,
  mapJob,
  mapRepository,
  unwrapApiData,
  type ApiJob,
  type ApiRepository,
} from "@/lib/map-api";
import { formatDate, truncate } from "@/lib/utils";
import type { Job } from "@/types";
import { FolderGit2 } from "lucide-react";
import Link from "next/link";

function sortByCreatedDesc(jobs: Job[]): Job[] {
  return [...jobs].sort(
    (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime()
  );
}

export async function DashboardContent() {
  const [jobsRes, reposRes] = await Promise.all([
    apiFetch("/jobs"),
    apiFetch("/repositories"),
  ]);

  const jobsParsed = jobsRes.ok
    ? unwrapApiData<ApiJob[]>(await jobsRes.json())
    : [];
  const reposParsed = reposRes.ok
    ? unwrapApiData<ApiRepository[]>(await reposRes.json())
    : [];
  const jobsRaw: ApiJob[] = Array.isArray(jobsParsed) ? jobsParsed : [];
  const reposRaw: ApiRepository[] = Array.isArray(reposParsed)
    ? reposParsed
    : [];
  const names = jobRepoNameMap(reposRaw);
  const jobs = jobsRaw.map((j) => mapJob(j, names[j.repository_id]));
  const repositories = reposRaw.map(mapRepository);

  const total = jobs.length;
  const completed = jobs.filter((j) => j.status === "completed").length;
  const failed = jobs.filter((j) => j.status === "failed").length;
  const repoCount = repositories.length;
  const recent = sortByCreatedDesc(jobs).slice(0, 5);

  return (
    <>
      <PageHeader title="Overview" />
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {[
          { label: "Total Jobs", value: total },
          { label: "Completed", value: completed },
          { label: "Failed", value: failed },
          { label: "Repositories Connected", value: repoCount },
        ].map((m) => (
          <Card key={m.label} className="p-6">
            <p className="text-sm text-gray-500 dark:text-gray-400">{m.label}</p>
            <p className="mt-2 text-3xl font-semibold tabular-nums">{m.value}</p>
          </Card>
        ))}
      </div>
      <section className="mt-12">
        <h2 className="text-lg font-semibold">Recent jobs</h2>
        {recent.length === 0 ? (
          <div className="mt-6">
            <EmptyState
              title="No jobs yet"
              description="Run a task from a repository to see it here."
              action={{ label: "New task", href: "/jobs/new" }}
            />
          </div>
        ) : (
          <div className="mt-4 overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-800">
            <table className="w-full min-w-[640px] text-left text-sm" role="table">
              <thead className="border-b border-gray-200 bg-gray-50 dark:border-gray-800 dark:bg-gray-950">
                <tr>
                  <th scope="col" className="px-4 py-3 font-medium">
                    Repository
                  </th>
                  <th scope="col" className="px-4 py-3 font-medium">
                    Task
                  </th>
                  <th scope="col" className="px-4 py-3 font-medium">
                    Status
                  </th>
                  <th scope="col" className="px-4 py-3 font-medium">
                    Date
                  </th>
                  <th scope="col" className="px-4 py-3 font-medium">
                    PR
                  </th>
                </tr>
              </thead>
              <tbody>
                {recent.map((job) => (
                  <tr
                    key={job.id}
                    className="border-b border-gray-200 last:border-0 dark:border-gray-800"
                  >
                    <td className="px-4 py-3 font-mono text-xs">
                      <Link
                        href={`/repositories/${job.repositoryId}`}
                        className="hover:underline"
                      >
                        {job.repositoryName}
                      </Link>
                    </td>
                    <td className="max-w-[240px] px-4 py-3 text-gray-600 dark:text-gray-300">
                      {truncate(job.prompt, 60)}
                    </td>
                    <td className="px-4 py-3">
                      <StatusBadge status={job.status} />
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-gray-500 dark:text-gray-400">
                      {formatDate(job.createdAt)}
                    </td>
                    <td className="px-4 py-3">
                      {job.prUrl ? (
                        <a
                          href={job.prUrl}
                          className="text-sm font-medium hover:underline"
                          target="_blank"
                          rel="noreferrer"
                        >
                          Open
                        </a>
                      ) : (
                        <span className="text-gray-400">—</span>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
      <section className="mt-12">
        <div className="flex items-center justify-between gap-4">
          <h2 className="text-lg font-semibold">Repositories</h2>
          <Link
            href="/repositories"
            className="text-sm font-medium text-gray-500 hover:text-foreground dark:text-gray-400"
          >
            View all
          </Link>
        </div>
        {repositories.length === 0 ? (
          <div className="mt-6">
            <EmptyState
              title="No repositories"
              description="Connect a GitHub repository to start shipping tasks."
              action={{ label: "Connect repository", href: "/repositories" }}
              icon={<FolderGit2 className="h-8 w-8" />}
            />
          </div>
        ) : (
          <ul className="mt-4 space-y-3">
            {repositories.map((r) => (
              <li key={r.id}>
                <RepoCard repo={r} />
              </li>
            ))}
          </ul>
        )}
      </section>
    </>
  );
}
