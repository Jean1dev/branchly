import { JobDetailContent } from "./job-detail-content";
import { TableSkeleton } from "@/components/skeletons/table-skeleton";
import { Suspense } from "react";

type PageProps = { params: Promise<{ id: string }> };

export default async function JobDetailPage({ params }: PageProps) {
  const { id } = await params;
  return (
    <Suspense fallback={<TableSkeleton rows={8} />}>
      <JobDetailContent id={id} />
    </Suspense>
  );
}
