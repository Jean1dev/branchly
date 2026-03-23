import { EmptyState } from "@/components/features/empty-state";
import { JobCard } from "@/components/features/job-card";
import { PageHeader } from "@/components/layout/page-header";
import { Button } from "@/components/ui/button";
import {
  getJobsByRepositoryId,
  getRepositoryById,
} from "@/lib/mock-data";
import { delay } from "@/lib/utils";
import Link from "next/link";
import { notFound } from "next/navigation";

export async function RepoDetailContent({ id }: { id: string }) {
  await delay(300);
  const repo = getRepositoryById(id);
  if (!repo) {
    notFound();
  }
  const jobs = getJobsByRepositoryId(repo.id);

  return (
    <>
      <nav className="mb-6 text-sm text-gray-500 dark:text-gray-400" aria-label="Breadcrumb">
        <ol className="flex flex-wrap items-center gap-2">
          <li>
            <Link href="/repositories" className="hover:text-foreground">
              Repositories
            </Link>
          </li>
          <li aria-hidden>/</li>
          <li className="font-mono text-foreground">{repo.fullName}</li>
        </ol>
      </nav>
      <PageHeader
        title={repo.fullName}
        description={`Default branch ${repo.defaultBranch} · ${repo.language}`}
        actions={
          <Button href={`/jobs/new?repositoryId=${repo.id}`}>
            New task for this repository
          </Button>
        }
      />
      <section>
        <h2 className="text-lg font-semibold">Jobs</h2>
        {jobs.length === 0 ? (
          <div className="mt-6">
            <EmptyState
              title="No jobs for this repository"
              description="Start a task to let Branchly implement changes and open a PR."
              action={{
                label: "New task",
                href: `/jobs/new?repositoryId=${repo.id}`,
              }}
            />
          </div>
        ) : (
          <ul className="mt-4 space-y-4">
            {jobs.map((job) => (
              <li key={job.id}>
                <JobCard job={job} />
              </li>
            ))}
          </ul>
        )}
      </section>
    </>
  );
}
