import { ConnectRepoButton } from "@/components/features/connect-repo-button";
import { EmptyState } from "@/components/features/empty-state";
import { RepoCard } from "@/components/features/repo-card";
import { PageHeader } from "@/components/layout/page-header";
import { apiFetch } from "@/lib/api-client";
import {
  mapRepository,
  unwrapApiData,
  type ApiRepository,
} from "@/lib/map-api";
import { FolderGit2 } from "lucide-react";

export async function RepositoriesContent() {
  const res = await apiFetch("/repositories");
  if (!res.ok) {
    return (
      <>
        <PageHeader
          title="Repositories"
          actions={<ConnectRepoButton />}
        />
        <p className="text-sm text-gray-500 dark:text-gray-400">
          Could not load repositories.
        </p>
      </>
    );
  }
  const raw = unwrapApiData<ApiRepository[]>(await res.json());
  const list = (Array.isArray(raw) ? raw : []).map(mapRepository);

  return (
    <>
      <PageHeader
        title="Repositories"
        actions={<ConnectRepoButton />}
      />
      {list.length === 0 ? (
        <EmptyState
          title="No repositories connected"
          description="Link a repository to describe tasks and open pull requests."
          action={{ label: "Connect repository", href: "/repositories" }}
          icon={<FolderGit2 className="h-8 w-8" />}
        />
      ) : (
        <ul className="space-y-4">
          {list.map((r) => (
            <li key={r.id}>
              <RepoCard repo={r} />
            </li>
          ))}
        </ul>
      )}
    </>
  );
}
