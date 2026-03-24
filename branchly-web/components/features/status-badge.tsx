import { Badge } from "@/components/ui/badge";
import type { JobStatus } from "@/types";

const labels: Record<JobStatus, string> = {
  pending: "Pending",
  completed: "Completed",
  running: "Running",
  failed: "Failed",
};

const variants: Record<
  JobStatus,
  "success" | "warning" | "error" | "default"
> = {
  pending: "default",
  completed: "success",
  running: "warning",
  failed: "error",
};

export function StatusBadge({ status }: { status: JobStatus }) {
  return <Badge variant={variants[status]}>{labels[status]}</Badge>;
}
