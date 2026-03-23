import { StatusBadge } from "@/components/features/status-badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { formatDate, truncate } from "@/lib/utils";
import type { Job } from "@/types";
import Link from "next/link";

export function JobCard({ job }: { job: Job }) {
  return (
    <Card className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div className="min-w-0 flex-1 space-y-2">
        <div className="flex flex-wrap items-center gap-2">
          <Link
            href={`/jobs/${job.id}`}
            className="font-mono text-sm font-medium hover:underline"
          >
            {job.id}
          </Link>
          <StatusBadge status={job.status} />
        </div>
        <p className="text-sm text-gray-500 dark:text-gray-400">
          {job.repositoryName}
        </p>
        <p className="text-sm">{truncate(job.prompt, 120)}</p>
        <p className="text-xs text-gray-500 dark:text-gray-400">
          {formatDate(job.createdAt)}
        </p>
      </div>
      <Button variant="secondary" size="sm" href={`/jobs/${job.id}`}>
        View
      </Button>
    </Card>
  );
}
