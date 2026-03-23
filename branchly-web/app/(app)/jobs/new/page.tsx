import { NewTaskForm } from "./new-task-form";
import { Skeleton } from "@/components/skeletons/skeleton";
import { Suspense } from "react";

function FormFallback() {
  return (
    <div className="space-y-6">
      <Skeleton className="h-8 w-48" />
      <Skeleton className="h-9 w-full max-w-md" />
      <Skeleton className="h-40 w-full max-w-2xl" />
      <Skeleton className="h-11 w-40" />
    </div>
  );
}

export default function NewJobPage() {
  return (
    <Suspense fallback={<FormFallback />}>
      <NewTaskForm />
    </Suspense>
  );
}
