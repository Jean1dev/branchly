"use client";

import { PageHeader } from "@/components/layout/page-header";
import { Button } from "@/components/ui/button";
import { mockRepositories } from "@/lib/mock-data";
import { useRouter, useSearchParams } from "next/navigation";
import { useMemo, useState } from "react";

export function NewTaskForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const defaultRepoId = searchParams.get("repositoryId") ?? "";
  const validDefault = useMemo(
    () => mockRepositories.some((r) => r.id === defaultRepoId),
    [defaultRepoId]
  );
  const [repositoryId, setRepositoryId] = useState(
    validDefault ? defaultRepoId : mockRepositories[0]?.id ?? ""
  );
  const [prompt, setPrompt] = useState("");
  const canSubmit = repositoryId.length > 0 && prompt.trim().length > 0;

  return (
    <>
      <PageHeader title="New task" />
      <form
        className="max-w-2xl space-y-6"
        onSubmit={(e) => {
          e.preventDefault();
          if (!canSubmit) return;
          router.push("/jobs/job_02");
        }}
      >
        <div className="space-y-2">
          <label htmlFor="repository" className="text-sm font-medium">
            Repository
          </label>
          <select
            id="repository"
            className="flex h-9 w-full max-w-md rounded-md border border-gray-200 bg-transparent px-3 text-sm transition-colors duration-150 focus-visible:outline-none focus-visible:ring-[3px] focus-visible:ring-black/10 dark:border-gray-800 dark:focus-visible:ring-white/[0.12]"
            value={repositoryId}
            onChange={(e) => setRepositoryId(e.target.value)}
          >
            {mockRepositories.map((r) => (
              <option key={r.id} value={r.id}>
                {r.fullName}
              </option>
            ))}
          </select>
        </div>
        <div className="space-y-2">
          <label htmlFor="prompt" className="text-sm font-medium">
            Describe what you want to build
          </label>
          <textarea
            id="prompt"
            rows={8}
            placeholder="Example: Add rate limiting to the authentication endpoints using a sliding window algorithm"
            className="w-full rounded-md border border-gray-200 bg-transparent px-3 py-2 text-sm placeholder:text-gray-500 focus-visible:outline-none focus-visible:ring-[3px] focus-visible:ring-black/10 dark:border-gray-800 dark:placeholder:text-gray-400 dark:focus-visible:ring-white/[0.12]"
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
          />
        </div>
        <Button type="submit" size="lg" disabled={!canSubmit} className="gap-1">
          Run task
          <span aria-hidden>→</span>
        </Button>
      </form>
    </>
  );
}
