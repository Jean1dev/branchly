"use client";

import { PageHeader } from "@/components/layout/page-header";
import { unwrapApiData } from "@/lib/map-api";
import { Button } from "@/components/ui/button";
import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useMemo, useState } from "react";

type RepoRow = { id: string; full_name: string };

export function NewTaskForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const defaultRepoId = searchParams.get("repositoryId") ?? "";
  const [repos, setRepos] = useState<RepoRow[]>([]);
  const [selectedId, setSelectedId] = useState("");
  const [prompt, setPrompt] = useState("");
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    void fetch("/api/repositories")
      .then(async (r) => {
        if (!r.ok) return [] as RepoRow[];
        return unwrapApiData<RepoRow[]>(await r.json());
      })
      .then((data) => setRepos(Array.isArray(data) ? data : []))
      .catch(() => setRepos([]));
  }, []);

  const repositoryId = useMemo(() => {
    if (selectedId && repos.some((r) => r.id === selectedId)) {
      return selectedId;
    }
    if (defaultRepoId && repos.some((r) => r.id === defaultRepoId)) {
      return defaultRepoId;
    }
    return repos[0]?.id ?? "";
  }, [repos, selectedId, defaultRepoId]);

  const canSubmit =
    repositoryId.length > 0 && prompt.trim().length > 0 && !submitting;

  return (
    <>
      <PageHeader title="New task" />
      <form
        className="max-w-2xl space-y-6"
        onSubmit={(e) => {
          e.preventDefault();
          if (!canSubmit) return;
          setSubmitting(true);
          void fetch("/api/jobs", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
              repository_id: repositoryId,
              prompt: prompt.trim(),
            }),
          })
            .then(async (res) => {
              if (!res.ok) return;
              const j = unwrapApiData<{ id: string }>(await res.json());
              router.push(`/jobs/${j.id}`);
            })
            .finally(() => setSubmitting(false));
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
            onChange={(e) => setSelectedId(e.target.value)}
            disabled={repos.length === 0}
          >
            {repos.length === 0 ? (
              <option value="">No repositories connected</option>
            ) : (
              repos.map((r) => (
                <option key={r.id} value={r.id}>
                  {r.full_name}
                </option>
              ))
            )}
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
