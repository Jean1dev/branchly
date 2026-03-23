import { RepoDetailContent } from "./repo-detail-content";
import { TableSkeleton } from "@/components/skeletons/table-skeleton";
import { Suspense } from "react";

type PageProps = { params: Promise<{ id: string }> };

export default async function RepositoryDetailPage({ params }: PageProps) {
  const { id } = await params;
  return (
    <Suspense fallback={<TableSkeleton rows={5} />}>
      <RepoDetailContent id={id} />
    </Suspense>
  );
}
