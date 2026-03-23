import { DashboardContent } from "./dashboard-content";
import { DashboardSkeleton } from "@/components/skeletons/dashboard-skeleton";
import { Suspense } from "react";

export default function DashboardPage() {
  return (
    <Suspense fallback={<DashboardSkeleton />}>
      <DashboardContent />
    </Suspense>
  );
}
