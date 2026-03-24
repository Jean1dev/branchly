"use client";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { RepoLanguageIcon } from "@/components/features/repo-language-icon";
import {
  parseApiErrorMessage,
  unwrapApiData,
} from "@/lib/map-api";
import { cn } from "@/lib/utils";
import { useCallback, useEffect, useMemo, useState } from "react";

type GitHubRepoItem = {
  id: number;
  full_name: string;
  default_branch: string;
  language: string;
  private?: boolean;
};

export function ConnectRepoButton() {
  const [open, setOpen] = useState(false);
  const [repos, setRepos] = useState<GitHubRepoItem[]>([]);
  const [sel, setSel] = useState("");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [connectError, setConnectError] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    setLoadError(null);
    try {
      const res = await fetch("/api/repositories/github");
      const json: unknown = await res.json();
      if (!res.ok) {
        setLoadError(parseApiErrorMessage(json, res.status));
        setRepos([]);
        setSel("");
        return;
      }
      const data = unwrapApiData<GitHubRepoItem[]>(json);
      const list = Array.isArray(data) ? data : [];
      setRepos(list);
      setSel(list[0] ? String(list[0].id) : "");
    } finally {
      setLoading(false);
    }
  }, []);

  const filteredRepos = useMemo(() => {
    const q = query.trim().toLowerCase();
    if (!q) return repos;
    return repos.filter((r) => r.full_name.toLowerCase().includes(q));
  }, [repos, query]);

  useEffect(() => {
    if (filteredRepos.length === 0) return;
    if (filteredRepos.some((r) => String(r.id) === sel)) return;
    setSel(String(filteredRepos[0].id));
  }, [filteredRepos, sel]);

  useEffect(() => {
    if (!open) return;
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") setOpen(false);
    }
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, [open]);

  const onOpen = () => {
    setOpen(true);
    setConnectError(null);
    setQuery("");
    void load();
  };

  const connect = async () => {
    const r = filteredRepos.find((x) => String(x.id) === sel);
    if (!r) return;
    setSaving(true);
    setConnectError(null);
    try {
      const res = await fetch("/api/repositories", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          github_repo_id: r.id,
          full_name: r.full_name,
          default_branch: r.default_branch || "main",
          language: r.language ?? "",
        }),
      });
      const json: unknown = await res.json();
      if (!res.ok) {
        setConnectError(parseApiErrorMessage(json, res.status));
        return;
      }
      setOpen(false);
      window.location.assign("/repositories");
    } finally {
      setSaving(false);
    }
  };

  return (
    <>
      <Button type="button" variant="secondary" onClick={onOpen}>
        Connect repository
      </Button>
      {open ? (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
          role="dialog"
          aria-modal="true"
          aria-labelledby="connect-repo-title"
          onClick={() => setOpen(false)}
        >
          <div
            className="w-full max-w-md rounded-lg border border-gray-200 bg-background p-6 shadow-lg dark:border-gray-800"
            onClick={(e) => e.stopPropagation()}
          >
            <h2 id="connect-repo-title" className="text-lg font-semibold">
              Connect a GitHub repository
            </h2>
            <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
              Repositories you can push to and that are not already linked appear
              here. Sign in with GitHub using the{" "}
              <code className="rounded bg-gray-100 px-1 font-mono text-xs dark:bg-gray-900">
                repo
              </code>{" "}
              scope so we can list them.
            </p>
            {loadError ? (
              <p
                className="mt-4 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800 dark:border-red-900 dark:bg-red-950/40 dark:text-red-200"
                role="alert"
              >
                {loadError}
              </p>
            ) : null}
            {connectError ? (
              <p
                className="mt-4 rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800 dark:border-red-900 dark:bg-red-950/40 dark:text-red-200"
                role="alert"
              >
                {connectError}
              </p>
            ) : null}
            {loading ? (
              <p className="mt-6 text-sm text-gray-500">Loading repositories…</p>
            ) : loadError ? null : repos.length === 0 ? (
              <p className="mt-6 text-sm text-gray-500">
                No repositories available to connect. You may have linked them
                all, or none grant push access.
              </p>
            ) : (
              <div className="mt-6 space-y-4">
                <div className="space-y-2">
                  <label htmlFor="gh-repo-search" className="text-sm font-medium">
                    Search
                  </label>
                  <Input
                    id="gh-repo-search"
                    type="search"
                    placeholder="Filter by name…"
                    value={query}
                    onChange={(e) => setQuery(e.target.value)}
                    autoComplete="off"
                  />
                </div>
                <div className="space-y-2">
                  <div className="text-sm font-medium" id="gh-repo-label">
                    Repository
                  </div>
                  <ul
                    role="listbox"
                    aria-labelledby="gh-repo-label"
                    className="max-h-60 overflow-y-auto rounded-md border border-gray-200 bg-background p-1 dark:border-gray-800"
                  >
                    {filteredRepos.map((repo) => {
                      const id = String(repo.id);
                      const selected = sel === id;
                      return (
                        <li key={repo.id}>
                          <button
                            type="button"
                            role="option"
                            aria-selected={selected}
                            className={cn(
                              "flex w-full items-center gap-3 rounded-md px-3 py-2 text-left text-sm text-foreground transition-colors",
                              selected
                                ? "bg-gray-100 dark:bg-gray-900"
                                : "hover:bg-gray-100 dark:hover:bg-gray-800"
                            )}
                            onClick={() => setSel(id)}
                          >
                            <RepoLanguageIcon
                              key={`${repo.id}-${repo.language ?? ""}`}
                              language={repo.language}
                              size="sm"
                            />
                            <span className="min-w-0 flex-1 truncate font-mono">
                              {repo.full_name}
                            </span>
                            {repo.private ? (
                              <span className="shrink-0 text-xs text-gray-500 dark:text-gray-400">
                                Private
                              </span>
                            ) : null}
                          </button>
                        </li>
                      );
                    })}
                  </ul>
                </div>
                {filteredRepos.length === 0 && repos.length > 0 ? (
                  <p className="text-sm text-gray-500">
                    No repositories match your search.
                  </p>
                ) : null}
              </div>
            )}
            <div className="mt-6 flex justify-end gap-2">
              <Button
                type="button"
                variant="secondary"
                onClick={() => setOpen(false)}
              >
                Cancel
              </Button>
              <Button
                type="button"
                disabled={
                  saving ||
                  loading ||
                  !!loadError ||
                  filteredRepos.length === 0 ||
                  !sel
                }
                onClick={() => void connect()}
              >
                {saving ? "Connecting…" : "Connect"}
              </Button>
            </div>
          </div>
        </div>
      ) : null}
    </>
  );
}
