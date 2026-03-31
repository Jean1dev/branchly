"use client";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { ProviderLogo } from "@/components/ui/provider-logo";
import { PageHeader } from "@/components/layout/page-header";
import { parseApiErrorMessage, unwrapApiData, type ApiIntegration } from "@/lib/map-api";
import { useCallback, useEffect, useState } from "react";

type IntegrationInfo = {
  id: string;
  provider: "github" | "gitlab" | "azure-devops";
  tokenType: "oauth" | "pat";
  connectedAt: string;
};

export default function IntegrationsPage() {
  const [integrations, setIntegrations] = useState<IntegrationInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [gitlabModalOpen, setGitlabModalOpen] = useState(false);
  const [azureModalOpen, setAzureModalOpen] = useState(false);
  const [disconnectingId, setDisconnectingId] = useState<string | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetch("/api/integrations");
      if (!res.ok) return;
      const json: unknown = await res.json();
      const data = unwrapApiData<ApiIntegration[]>(json);
      const list: IntegrationInfo[] = (Array.isArray(data) ? data : []).map((i) => ({
        id: i.id,
        provider: i.provider === "gitlab" ? "gitlab" : i.provider === "azure-devops" ? "azure-devops" : "github",
        tokenType: i.token_type === "pat" ? "pat" : "oauth",
        connectedAt: i.connected_at,
      }));
      // GitHub always first
      list.sort((a, b) => (a.provider === "github" ? -1 : b.provider === "github" ? 1 : 0));
      setIntegrations(list);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { void load(); }, [load]);

  const githubIntegration = integrations.find((i) => i.provider === "github");
  const gitlabIntegration = integrations.find((i) => i.provider === "gitlab");
  const azureIntegration = integrations.find((i) => i.provider === "azure-devops");

  async function handleDisconnect(id: string) {
    setDisconnectingId(id);
    try {
      await fetch(`/api/integrations/${id}`, { method: "DELETE" });
      await load();
    } finally {
      setDisconnectingId(null);
    }
  }

  return (
    <>
      <PageHeader
        title="Git integrations"
        description="Connect your repositories from any Git provider."
      />
      {loading ? (
        <div className="space-y-4">
          {[1, 2].map((i) => (
            <div key={i} className="h-24 animate-pulse rounded-lg bg-gray-100 dark:bg-gray-900" />
          ))}
        </div>
      ) : (
        <div className="max-w-2xl space-y-4">
          {/* GitHub card */}
          <Card className="flex flex-wrap items-center justify-between gap-4 p-6">
            <div className="flex items-center gap-3">
              <ProviderLogo provider="github" size={24} />
              <div>
                <p className="font-medium">GitHub</p>
                {githubIntegration ? (
                  <p className="text-sm text-gray-500 dark:text-gray-400">
                    Connected via OAuth
                  </p>
                ) : (
                  <p className="text-sm text-gray-500 dark:text-gray-400">Not connected</p>
                )}
              </div>
            </div>
            <div className="flex items-center gap-3">
              {githubIntegration && <Badge variant="success">Connected</Badge>}
              <div title="GitHub is required for sign in">
                <Button variant="secondary" disabled size="sm">
                  Disconnect
                </Button>
              </div>
            </div>
          </Card>

          {/* GitLab card */}
          <Card className="flex flex-wrap items-center justify-between gap-4 p-6">
            <div className="flex items-center gap-3">
              <ProviderLogo provider="gitlab" size={24} />
              <div>
                <p className="font-medium">GitLab</p>
                {gitlabIntegration ? (
                  <p className="text-sm text-gray-500 dark:text-gray-400">
                    Connected via Personal Access Token
                  </p>
                ) : (
                  <p className="text-sm text-gray-500 dark:text-gray-400">Not connected</p>
                )}
              </div>
            </div>
            <div className="flex items-center gap-3">
              {gitlabIntegration && <Badge variant="success">Connected</Badge>}
              {gitlabIntegration ? (
                <Button
                  variant="secondary"
                  size="sm"
                  disabled={disconnectingId === gitlabIntegration.id}
                  onClick={() => void handleDisconnect(gitlabIntegration.id)}
                >
                  {disconnectingId === gitlabIntegration.id ? "Disconnecting…" : "Disconnect"}
                </Button>
              ) : (
                <Button
                  size="sm"
                  onClick={() => setGitlabModalOpen(true)}
                >
                  Connect GitLab →
                </Button>
              )}
            </div>
          </Card>
          {/* Azure DevOps card */}
          <Card className="flex flex-wrap items-center justify-between gap-4 p-6">
            <div className="flex items-center gap-3">
              <ProviderLogo provider="azure-devops" size={24} />
              <div>
                <p className="font-medium">Azure DevOps</p>
                {azureIntegration ? (
                  <p className="text-sm text-gray-500 dark:text-gray-400">
                    Connected via Personal Access Token
                  </p>
                ) : (
                  <p className="text-sm text-gray-500 dark:text-gray-400">Not connected</p>
                )}
              </div>
            </div>
            <div className="flex items-center gap-3">
              {azureIntegration && <Badge variant="success">Connected</Badge>}
              {azureIntegration ? (
                <Button
                  variant="secondary"
                  size="sm"
                  disabled={disconnectingId === azureIntegration.id}
                  onClick={() => void handleDisconnect(azureIntegration.id)}
                >
                  {disconnectingId === azureIntegration.id ? "Disconnecting…" : "Disconnect"}
                </Button>
              ) : (
                <Button
                  size="sm"
                  onClick={() => setAzureModalOpen(true)}
                >
                  Connect Azure DevOps →
                </Button>
              )}
            </div>
          </Card>
        </div>
      )}

      {gitlabModalOpen && (
        <ConnectGitLabModal
          onClose={() => setGitlabModalOpen(false)}
          onSuccess={() => { setGitlabModalOpen(false); void load(); }}
        />
      )}
      {azureModalOpen && (
        <ConnectAzureDevOpsModal
          onClose={() => setAzureModalOpen(false)}
          onSuccess={() => { setAzureModalOpen(false); void load(); }}
        />
      )}
    </>
  );
}

function ConnectAzureDevOpsModal({
  onClose,
  onSuccess,
}: {
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [pat, setPat] = useState("");
  const [orgURL, setOrgURL] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") onClose();
    }
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, [onClose]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError(null);
    try {
      const res = await fetch("/api/integrations/azure-devops", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ pat, org_url: orgURL.trim() }),
      });
      const json: unknown = await res.json();
      if (!res.ok) {
        setError(parseApiErrorMessage(json, res.status));
        return;
      }
      onSuccess();
    } finally {
      setSaving(false);
    }
  }

  const createTokenURL =
    "https://dev.azure.com/_usersSettings/tokens";

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="connect-azure-title"
      onClick={onClose}
    >
      <div
        className="w-full max-w-md rounded-lg border border-gray-200 bg-background p-6 shadow-lg dark:border-gray-800"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 id="connect-azure-title" className="text-lg font-semibold">
          Connect Azure DevOps
        </h2>
        <p className="mt-3 text-sm text-gray-500 dark:text-gray-400">
          Create a Personal Access Token in Azure DevOps with the following scope:
        </p>
        <ul className="mt-2 space-y-1 text-sm text-gray-700 dark:text-gray-300">
          {["Code (Read & Write)"].map((s) => (
            <li key={s} className="flex items-center gap-2">
              <span className="text-gray-400">•</span>
              <code className="font-mono text-xs">{s}</code>
            </li>
          ))}
        </ul>
        <a
          href={createTokenURL}
          target="_blank"
          rel="noreferrer"
          className="mt-4 inline-flex items-center gap-1 text-sm font-medium hover:underline"
        >
          Create token on Azure DevOps →
        </a>
        <form onSubmit={(e) => void handleSubmit(e)} className="mt-5 space-y-4">
          <div className="space-y-1">
            <label htmlFor="azure-org-url" className="text-sm font-medium">
              Organization URL
            </label>
            <Input
              id="azure-org-url"
              type="url"
              placeholder="https://dev.azure.com/my-organization"
              value={orgURL}
              onChange={(e) => setOrgURL(e.target.value)}
              autoComplete="off"
              required
            />
          </div>
          <div className="space-y-1">
            <label htmlFor="azure-pat" className="text-sm font-medium">
              Personal Access Token
            </label>
            <Input
              id="azure-pat"
              type="password"
              placeholder="••••••••••••••••••••••••••••••••••••••••••"
              value={pat}
              onChange={(e) => setPat(e.target.value)}
              autoComplete="off"
              required
            />
          </div>
          {error && (
            <p
              className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800 dark:border-red-900 dark:bg-red-950/40 dark:text-red-200"
              role="alert"
            >
              {error}
            </p>
          )}
          <div className="flex justify-end gap-2">
            <Button type="button" variant="secondary" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={saving || !pat.trim() || !orgURL.trim()}>
              {saving ? "Connecting…" : "Connect Azure DevOps"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}

function ConnectGitLabModal({
  onClose,
  onSuccess,
}: {
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [pat, setPat] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") onClose();
    }
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, [onClose]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setSaving(true);
    setError(null);
    try {
      const res = await fetch("/api/integrations/gitlab", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ pat }),
      });
      const json: unknown = await res.json();
      if (!res.ok) {
        setError(parseApiErrorMessage(json, res.status));
        return;
      }
      onSuccess();
    } finally {
      setSaving(false);
    }
  }

  const createTokenURL =
    "https://gitlab.com/-/user_settings/personal_access_tokens?name=Branchly&scopes=read_user,read_api,read_repository,write_repository";

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="connect-gitlab-title"
      onClick={onClose}
    >
      <div
        className="w-full max-w-md rounded-lg border border-gray-200 bg-background p-6 shadow-lg dark:border-gray-800"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 id="connect-gitlab-title" className="text-lg font-semibold">
          Connect GitLab
        </h2>
        <p className="mt-3 text-sm text-gray-500 dark:text-gray-400">
          Create a Personal Access Token in GitLab with the following scopes:
        </p>
        <ul className="mt-2 space-y-1 text-sm text-gray-700 dark:text-gray-300">
          {["read_user", "read_api", "read_repository", "write_repository"].map((s) => (
            <li key={s} className="flex items-center gap-2">
              <span className="text-gray-400">•</span>
              <code className="font-mono text-xs">{s}</code>
            </li>
          ))}
        </ul>
        <a
          href={createTokenURL}
          target="_blank"
          rel="noreferrer"
          className="mt-4 inline-flex items-center gap-1 text-sm font-medium hover:underline"
        >
          Create token on GitLab →
        </a>
        <form onSubmit={(e) => void handleSubmit(e)} className="mt-5 space-y-4">
          <div className="space-y-1">
            <label htmlFor="gitlab-pat" className="text-sm font-medium">
              Personal Access Token
            </label>
            <Input
              id="gitlab-pat"
              type="password"
              placeholder="glpat-xxxxxxxxxxxxxxxxxxxx"
              value={pat}
              onChange={(e) => setPat(e.target.value)}
              autoComplete="off"
              required
            />
          </div>
          {error && (
            <p
              className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800 dark:border-red-900 dark:bg-red-950/40 dark:text-red-200"
              role="alert"
            >
              {error}
            </p>
          )}
          <div className="flex justify-end gap-2">
            <Button type="button" variant="secondary" onClick={onClose}>
              Cancel
            </Button>
            <Button type="submit" disabled={saving || !pat.trim()}>
              {saving ? "Connecting…" : "Connect GitLab"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
