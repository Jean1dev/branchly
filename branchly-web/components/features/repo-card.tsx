import { RepoLanguageIcon } from "@/components/features/repo-language-icon";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { formatDate } from "@/lib/utils";
import type { Repository } from "@/types";

export function RepoCard({ repo }: { repo: Repository }) {
  return (
    <Card className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <div className="flex min-w-0 flex-1 items-start gap-3 sm:items-center">
        <RepoLanguageIcon
          key={repo.language ?? "none"}
          language={repo.language}
          size="md"
        />
        <div className="min-w-0 space-y-1">
          <p className="truncate font-mono text-sm font-semibold">
            {repo.fullName}
          </p>
          <p className="text-sm text-gray-500 dark:text-gray-400">
            {repo.language} · {repo.jobsCount} jobs · Last job{" "}
            {formatDate(repo.lastJobAt)}
          </p>
        </div>
      </div>
      <div className="flex shrink-0 gap-2">
        <Button variant="secondary" size="sm" href={`/repositories/${repo.id}`}>
          View
        </Button>
        <Button size="sm" href="/jobs/new">
          New task
        </Button>
      </div>
    </Card>
  );
}
