import { ConnectRepoButton } from "@/components/features/connect-repo-button";
import { EmptyState } from "@/components/features/empty-state";
import { RepoCard } from "@/components/features/repo-card";
import { PageHeader } from "@/components/layout/page-header";
import { mockRepositories } from "@/lib/mock-data";
import { delay } from "@/lib/utils";
import { FolderGit2 } from "lucide-react";

export async function RepositoriesContent() {
  await delay(350);
  return (
    <>
      <PageHeader
        title="Repositories"
        actions={<ConnectRepoButton />}
      />
      {mockRepositories.length === 0 ? (
        <EmptyState
          title="No repositories connected"
          description="Link a repository to describe tasks and open pull requests."
          action={{ label: "Connect repository", href: "/repositories" }}
          icon={<FolderGit2 className="h-8 w-8" />}
        />
      ) : (
        <ul className="space-y-4">
          {mockRepositories.map((r) => (
            <li key={r.id}>
              <RepoCard repo={r} />
            </li>
          ))}
        </ul>
      )}
    </>
  );
}
