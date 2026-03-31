"use client";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { ProviderBadge } from "@/components/ui/provider-badge";
import { ProviderLogo } from "@/components/ui/provider-logo";
import { RepoLanguageIcon } from "@/components/features/repo-language-icon";
import { parseApiErrorMessage, unwrapApiData, type ApiIntegration, type ApiProviderRepo } from "@/lib/map-api";
import { cn } from "@/lib/utils";
import type { GitProvider } from "@/types";
import { useCallback, useEffect, useMemo, useState } from "react";

type Step = "provider" | "repos";

type IntegrationInfo = {
  id: string;
  provider: GitProvider;
};

export function ConnectRepoButton() {
  const [open, setOpen] = useState(false);
  const [step, setStep] = useState<Step>("provider");
  const [integrations, setIntegrations] = useState<IntegrationInfo[]>([]);
  const [selectedIntegration, setSelectedIntegration] = useState<IntegrationInfo | null>(null);
  const [repos, setRepos] = useState<ApiProviderRepo[]>([]);
  const [sel, setSel] = useState("");
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [loadError, setLoadError] = useState<string | null>(null);
  const [connectError, setConnectError] = useState<string | null>(null);

  const loadIntegrations = useCallback(async () => {
    try {
      const res = await fetch("/api/integrations");
      if (!res.ok) return;
      const json: unknown = await res.json();
      const data = unwrapApiData<ApiIntegration[]>(json);
      const list: IntegrationInfo[] = (Array.isArray(data) ? data : []).map((i) => ({
        id: i.id,
        provider: i.provider === "gitlab" ? "gitlab" : i.provider === "azure-devops" ? "azure-devops" : "github",
      }));
      list.sort((a, b) => (a.provider === "github" ? -1 : b.provider === "github" ? 1 : 0));
      setIntegrations(list);
    } catch { /* silent */ }
  }, []);

  const loadRepos = useCallback(async (integrationId: string) => {
    setLoading(true);
    setLoadError(null);
    setRepos([]);
    setSel("");
    try {
      const res = await fetch(`/api/integrations/${integrationId}/repositories`);
      const json: unknown = await res.json();
      if (!res.ok) {
        setLoadError(parseApiErrorMessage(json, res.status));
        return;
      }
      const data = unwrapApiData<ApiProviderRepo[]>(json);
      const list = Array.isArray(data) ? data : [];
      setRepos(list);
      setSel(list[0]?.external_id ?? "");
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
    if (filteredRepos.some((r) => r.external_id === sel)) return;
    setSel(filteredRepos[0]?.external_id ?? "");
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
    setStep("provider");
    setSelectedIntegration(null);
    setConnectError(null);
    setQuery("");
    void loadIntegrations();
  };

  const onSelectProvider = (integ: IntegrationInfo) => {
    setSelectedIntegration(integ);
    setStep("repos");
    void loadRepos(integ.id);
  };

  const connect = async () => {
    const r = filteredRepos.find((x) => x.external_id === sel);
    if (!r || !selectedIntegration) return;
    setSaving(true);
    setConnectError(null);
    try {
      const res = await fetch("/api/repositories", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          integration_id: selectedIntegration.id,
          external_id: r.external_id,
          full_name: r.full_name,
          clone_url: r.clone_url,
          default_branch: r.default_branch || "main",
          language: r.language ?? "",
          provider: r.provider,
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

  const providerDisplayName: Record<GitProvider, string> = { github: "GitHub", gitlab: "GitLab", "azure-devops": "Azure DevOps" };

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
            {step === "provider" ? (
              <>
                <h2 id="connect-repo-title" className="text-lg font-semibold">
                  Connect a repository
                </h2>
                <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
                  Select the Git provider to browse your repositories.
                </p>
                <div className="mt-6 grid grid-cols-2 gap-3">
                  {(["github", "gitlab", "azure-devops"] as GitProvider[]).map((provider) => {
                    const integ = integrations.find((i) => i.provider === provider);
                    const connected = !!integ;
                    return (
                      <button
                        key={provider}
                        type="button"
                        className={cn(
                          "flex flex-col items-center gap-3 rounded-lg border p-5 text-sm transition-colors",
                          connected
                            ? "border-gray-200 hover:border-gray-300 hover:bg-gray-50 dark:border-gray-800 dark:hover:border-gray-700 dark:hover:bg-gray-900"
                            : "border-dashed border-gray-300 opacity-60 dark:border-gray-700"
                        )}
                        onClick={() => {
                          if (!connected) {
                            window.location.assign("/settings/integrations");
                          } else {
                            onSelectProvider(integ);
                          }
                        }}
                      >
                        <ProviderLogo provider={provider} size={28} />
                        <span className="font-medium">{providerDisplayName[provider]}</span>
                        {!connected && (
                          <span className="text-xs text-gray-500">Not connected</span>
                        )}
                      </button>
                    );
                  })}
                </div>
                <div className="mt-6 flex justify-end">
                  <Button type="button" variant="secondary" onClick={() => setOpen(false)}>
                    Cancel
                  </Button>
                </div>
              </>
            ) : (
              <>
                <div className="flex items-center gap-2">
                  <button
                    type="button"
                    className="text-sm text-gray-500 hover:text-foreground"
                    onClick={() => setStep("provider")}
                  >
                    ← Back
                  </button>
                  <h2 id="connect-repo-title" className="text-lg font-semibold">
                    {selectedIntegration && (
                      <span className="flex items-center gap-2">
                        <ProviderBadge provider={selectedIntegration.provider} />
                        repositories
                      </span>
                    )}
                  </h2>
                </div>
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
                    No repositories available to connect.
                  </p>
                ) : (
                  <div className="mt-4 space-y-4">
                    <div className="space-y-2">
                      <label htmlFor="repo-search" className="text-sm font-medium">
                        Search
                      </label>
                      <Input
                        id="repo-search"
                        type="search"
                        placeholder="Filter by name…"
                        value={query}
                        onChange={(e) => setQuery(e.target.value)}
                        autoComplete="off"
                      />
                    </div>
                    <div className="space-y-2">
                      <div className="text-sm font-medium" id="repo-label">
                        Repository
                      </div>
                      <ul
                        role="listbox"
                        aria-labelledby="repo-label"
                        className="max-h-60 overflow-y-auto rounded-md border border-gray-200 bg-background p-1 dark:border-gray-800"
                      >
                        {filteredRepos.map((repo) => {
                          const selected = sel === repo.external_id;
                          return (
                            <li key={repo.external_id}>
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
                                onClick={() => setSel(repo.external_id)}
                              >
                                <RepoLanguageIcon
                                  language={repo.language}
                                  size="sm"
                                />
                                <span className="min-w-0 flex-1 truncate font-mono text-xs">
                                  {repo.full_name}
                                </span>
                                {repo.default_branch && (
                                  <span className="shrink-0 text-xs text-gray-400">
                                    {repo.default_branch}
                                  </span>
                                )}
                              </button>
                            </li>
                          );
                        })}
                      </ul>
                    </div>
                    {filteredRepos.length === 0 && repos.length > 0 ? (
                      <p className="text-sm text-gray-500">No repositories match your search.</p>
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
              </>
            )}
          </div>
        </div>
      ) : null}
    </>
  );
}
