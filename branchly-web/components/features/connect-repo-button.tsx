"use client";

import { Button } from "@/components/ui/button";
import { unwrapApiData } from "@/lib/map-api";
import { useCallback, useState } from "react";

type GitHubRepoItem = {
  id: number;
  full_name: string;
  default_branch: string;
  language: string;
};

export function ConnectRepoButton() {
  const [open, setOpen] = useState(false);
  const [repos, setRepos] = useState<GitHubRepoItem[]>([]);
  const [sel, setSel] = useState("");
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetch("/api/repositories/github");
      if (!res.ok) return;
      const data = unwrapApiData<GitHubRepoItem[]>(await res.json());
      const list = Array.isArray(data) ? data : [];
      setRepos(list);
      setSel(list[0] ? String(list[0].id) : "");
    } finally {
      setLoading(false);
    }
  }, []);

  const onOpen = () => {
    setOpen(true);
    void load();
  };

  const connect = async () => {
    const r = repos.find((x) => String(x.id) === sel);
    if (!r) return;
    setSaving(true);
    try {
      const res = await fetch("/api/repositories", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          github_repo_id: r.id,
          full_name: r.full_name,
          default_branch: r.default_branch,
          language: r.language ?? "",
        }),
      });
      if (res.ok) {
        setOpen(false);
        window.location.assign("/repositories");
      }
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
        >
          <div className="w-full max-w-md rounded-lg border border-gray-200 bg-background p-6 shadow-lg dark:border-gray-800">
            <h2 id="connect-repo-title" className="text-lg font-semibold">
              Connect repository
            </h2>
            <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
              Choose a GitHub repository to link with Branchly.
            </p>
            {loading ? (
              <p className="mt-6 text-sm text-gray-500">Loading repositories…</p>
            ) : repos.length === 0 ? (
              <p className="mt-6 text-sm text-gray-500">
                No repositories found on your account.
              </p>
            ) : (
              <div className="mt-6 space-y-2">
                <label htmlFor="gh-repo" className="text-sm font-medium">
                  Repository
                </label>
                <select
                  id="gh-repo"
                  className="flex h-9 w-full rounded-md border border-gray-200 bg-transparent px-3 text-sm dark:border-gray-800"
                  value={sel}
                  onChange={(e) => setSel(e.target.value)}
                >
                  {repos.map((repo) => (
                    <option key={repo.id} value={String(repo.id)}>
                      {repo.full_name}
                    </option>
                  ))}
                </select>
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
                disabled={saving || !sel || repos.length === 0}
                onClick={() => void connect()}
              >
                Connect
              </Button>
            </div>
          </div>
        </div>
      ) : null}
    </>
  );
}
