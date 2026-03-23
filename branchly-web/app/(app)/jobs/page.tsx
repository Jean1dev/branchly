import { JobsView } from "@/components/features/jobs-view";
import { TableSkeleton } from "@/components/skeletons/table-skeleton";
import { Suspense } from "react";

export default function JobsPage() {
  return (
    <Suspense fallback={<TableSkeleton rows={10} />}>
      <JobsView />
    </Suspense>
  );
}
