"use client";

import { RepoLanguageIcon } from "@/components/features/repo-language-icon";
import { PageHeader } from "@/components/layout/page-header";
import { type ApiRepository, unwrapApiData } from "@/lib/map-api";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { ChevronDown } from "lucide-react";
import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useMemo, useRef, useState } from "react";

type RepoRow = Pick<ApiRepository, "id" | "full_name" | "language">;

export function NewTaskForm() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const defaultRepoId = searchParams.get("repositoryId") ?? "";
  const [repos, setRepos] = useState<RepoRow[]>([]);
  const [selectedId, setSelectedId] = useState("");
  const [repoMenuOpen, setRepoMenuOpen] = useState(false);
  const repoPickerRef = useRef<HTMLDivElement>(null);
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

  useEffect(() => {
    if (!repoMenuOpen) return;
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") setRepoMenuOpen(false);
    }
    function onPointer(e: MouseEvent) {
      const el = repoPickerRef.current;
      if (el && !el.contains(e.target as Node)) setRepoMenuOpen(false);
    }
    document.addEventListener("keydown", onKey);
    document.addEventListener("mousedown", onPointer);
    return () => {
      document.removeEventListener("keydown", onKey);
      document.removeEventListener("mousedown", onPointer);
    };
  }, [repoMenuOpen]);

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

  const selectedRepo = repos.find((r) => r.id === repositoryId);

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
          <div id="repository-label" className="text-sm font-medium">
            Repository
          </div>
          <div ref={repoPickerRef} className="relative max-w-md">
            <button
              type="button"
              id="repository"
              aria-haspopup="listbox"
              aria-expanded={repoMenuOpen}
              aria-labelledby="repository-label repository"
              disabled={repos.length === 0}
              className={cn(
                "flex h-9 w-full items-center justify-between gap-2 rounded-md border border-gray-200 bg-transparent px-3 text-left text-sm text-foreground transition-colors duration-150 focus-visible:outline-none focus-visible:ring-[3px] focus-visible:ring-black/10 disabled:cursor-not-allowed disabled:opacity-50 dark:border-gray-800 dark:focus-visible:ring-white/[0.12]"
              )}
              onClick={() => setRepoMenuOpen((o) => !o)}
            >
              <span className="flex min-w-0 flex-1 items-center gap-2">
                {selectedRepo ? (
                  <>
                    <RepoLanguageIcon
                      language={selectedRepo.language}
                      size="sm"
                    />
                    <span className="truncate font-mono">
                      {selectedRepo.full_name}
                    </span>
                  </>
                ) : (
                  <span className="truncate text-gray-500 dark:text-gray-400">
                    {repos.length === 0
                      ? "No repositories connected"
                      : "Select a repository"}
                  </span>
                )}
              </span>
              <ChevronDown
                className={cn(
                  "h-4 w-4 shrink-0 text-gray-500 transition-transform dark:text-gray-400",
                  repoMenuOpen && "rotate-180"
                )}
                aria-hidden
              />
            </button>
            {repoMenuOpen && repos.length > 0 ? (
              <ul
                role="listbox"
                aria-labelledby="repository-label"
                className="absolute z-50 mt-1 max-h-60 w-full overflow-y-auto rounded-md border border-gray-200 bg-background p-1 shadow-lg dark:border-gray-800"
              >
                {repos.map((r) => {
                  const picked = r.id === repositoryId;
                  return (
                    <li key={r.id} role="presentation">
                      <button
                        type="button"
                        role="option"
                        aria-selected={picked}
                        className={cn(
                          "flex w-full items-center gap-3 rounded-md px-3 py-2 text-left text-sm text-foreground transition-colors",
                          picked
                            ? "bg-gray-100 dark:bg-gray-900"
                            : "hover:bg-gray-100 dark:hover:bg-gray-800"
                        )}
                        onClick={() => {
                          setSelectedId(r.id);
                          setRepoMenuOpen(false);
                        }}
                      >
                        <RepoLanguageIcon
                          language={r.language}
                          size="sm"
                        />
                        <span className="min-w-0 flex-1 truncate font-mono">
                          {r.full_name}
                        </span>
                      </button>
                    </li>
                  );
                })}
              </ul>
            ) : null}
          </div>
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
