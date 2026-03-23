import { RepositoriesContent } from "./repositories-content";
import { TableSkeleton } from "@/components/skeletons/table-skeleton";
import { Suspense } from "react";

export default function RepositoriesPage() {
  return (
    <Suspense fallback={<TableSkeleton rows={4} />}>
      <RepositoriesContent />
    </Suspense>
  );
}
