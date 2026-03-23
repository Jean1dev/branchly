import { JobLogPanel } from "@/components/features/job-log-panel";
import { StatusBadge } from "@/components/features/status-badge";
import { Card } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { getJobById, mockJobLogs } from "@/lib/mock-data";
import { delay, formatDate, truncate } from "@/lib/utils";
import Link from "next/link";
import { notFound } from "next/navigation";

export async function JobDetailContent({ id }: { id: string }) {
  await delay(280);
  const job = getJobById(id);
  if (!job) {
    notFound();
  }

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
          <JobLogPanel lines={mockJobLogs} status={job.status} />
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
