import { JobsData } from "./jobs-data";
import { TableSkeleton } from "@/components/skeletons/table-skeleton";
import { Suspense } from "react";

export default function JobsPage() {
  return (
    <Suspense fallback={<TableSkeleton rows={10} />}>
      <JobsData />
    </Suspense>
  );
}
