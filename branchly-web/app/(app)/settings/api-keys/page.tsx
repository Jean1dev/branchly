"use client";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { PageHeader } from "@/components/layout/page-header";
import { parseApiErrorMessage, unwrapApiData } from "@/lib/map-api";
import type { APIKeyInfo, APIKeyProvider } from "@/types";
import { useCallback, useEffect, useState } from "react";

type ProviderConfig = {
  id: APIKeyProvider;
  label: string;
  usedBy: string;
  prefix: string;
  placeholder: string;
  docsURL: string;
  docsLabel: string;
  icon: string;
};

const PROVIDERS: ProviderConfig[] = [
  {
    id: "anthropic",
    label: "Anthropic",
    usedBy: "Claude Code",
    prefix: "sk-ant-",
    placeholder: "sk-ant-api03-…",
    docsURL: "https://console.anthropic.com/settings/keys",
    docsLabel: "Get your key at console.anthropic.com",
    icon: "A",
  },
  {
    id: "google",
    label: "Google AI",
    usedBy: "Gemini CLI",
    prefix: "AIza",
    placeholder: "AIzaSy…",
    docsURL: "https://aistudio.google.com/app/apikey",
    docsLabel: "Get your key at aistudio.google.com",
    icon: "G",
  },
  {
    id: "openai",
    label: "OpenAI",
    usedBy: "GPT Codex",
    prefix: "sk-",
    placeholder: "sk-…",
    docsURL: "https://platform.openai.com/api-keys",
    docsLabel: "Get your key at platform.openai.com",
    icon: "O",
  },
];

export default function APIKeysPage() {
  const [keys, setKeys] = useState<APIKeyInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [modalProvider, setModalProvider] = useState<ProviderConfig | null>(null);
  const [removingProvider, setRemovingProvider] = useState<APIKeyProvider | null>(null);
  const [confirmRemove, setConfirmRemove] = useState<APIKeyProvider | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetch("/api/settings/api-keys");
      if (!res.ok) return;
      const json: unknown = await res.json();
      const data = unwrapApiData<APIKeyInfo[]>(json);
      setKeys(Array.isArray(data) ? data : []);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { void load(); }, [load]);

  async function handleRemove(provider: APIKeyProvider) {
    setRemovingProvider(provider);
    try {
      await fetch(`/api/settings/api-keys/${provider}`, { method: "DELETE" });
      await load();
    } finally {
      setRemovingProvider(null);
      setConfirmRemove(null);
    }
  }

  return (
    <>
      <PageHeader
        title="API keys"
        description="Use your own API keys. Branchly will use them instead of the shared keys."
      />
      {loading ? (
        <div className="max-w-2xl space-y-4">
          {[1, 2].map((i) => (
            <div key={i} className="h-24 animate-pulse rounded-lg bg-gray-100 dark:bg-gray-900" />
          ))}
        </div>
      ) : (
        <div className="max-w-2xl space-y-4">
          {PROVIDERS.map((provider) => {
            const configured = keys.find((k) => k.provider === provider.id);
            return (
              <Card key={provider.id} className="p-6">
                <div className="flex flex-wrap items-center justify-between gap-4">
                  <div className="flex items-center gap-3">
                    <div className="flex h-9 w-9 items-center justify-center rounded-lg border border-gray-200 bg-gray-50 font-semibold text-gray-700 dark:border-gray-800 dark:bg-gray-900 dark:text-gray-300">
                      {provider.icon}
                    </div>
                    <div>
                      <p className="font-medium">{provider.label}</p>
                      <p className="text-sm text-gray-500 dark:text-gray-400">
                        Used by: {provider.usedBy}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center gap-3">
                    {configured ? (
                      <>
                        <Badge variant="success">Configured</Badge>
                        <span className="font-mono text-sm text-gray-500 dark:text-gray-400">
                          ············{configured.key_hint}
                        </span>
                        <Button
                          variant="secondary"
                          size="sm"
                          onClick={() => setModalProvider(provider)}
                        >
                          Update
                        </Button>
                        {confirmRemove === provider.id ? (
                          <div className="flex items-center gap-2">
                            <span className="text-xs text-gray-500">Remove?</span>
                            <Button
                              variant="secondary"
                              size="sm"
                              disabled={removingProvider === provider.id}
                              onClick={() => void handleRemove(provider.id)}
                            >
                              {removingProvider === provider.id ? "Removing…" : "Yes"}
                            </Button>
                            <Button
                              variant="secondary"
                              size="sm"
                              onClick={() => setConfirmRemove(null)}
                            >
                              Cancel
                            </Button>
                          </div>
                        ) : (
                          <Button
                            variant="secondary"
                            size="sm"
                            onClick={() => setConfirmRemove(provider.id)}
                          >
                            Remove
                          </Button>
                        )}
                      </>
                    ) : (
                      <>
                        <Badge variant="muted">Using shared key</Badge>
                        <Button size="sm" onClick={() => setModalProvider(provider)}>
                          Add key
                        </Button>
                      </>
                    )}
                  </div>
                </div>
              </Card>
            );
          })}
        </div>
      )}

      {modalProvider && (
        <APIKeyModal
          provider={modalProvider}
          isUpdate={!!keys.find((k) => k.provider === modalProvider.id)}
          onClose={() => setModalProvider(null)}
          onSuccess={() => {
            setModalProvider(null);
            void load();
          }}
        />
      )}
    </>
  );
}

function APIKeyModal({
  provider,
  isUpdate,
  onClose,
  onSuccess,
}: {
  provider: ProviderConfig;
  isUpdate: boolean;
  onClose: () => void;
  onSuccess: () => void;
}) {
  const [key, setKey] = useState("");
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    function onKey(e: KeyboardEvent) {
      if (e.key === "Escape") onClose();
    }
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, [onClose]);

  function validatePrefix(value: string): string | null {
    if (!value.startsWith(provider.prefix)) {
      return `Key must start with "${provider.prefix}"`;
    }
    return null;
  }

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const prefixErr = validatePrefix(key.trim());
    if (prefixErr) {
      setError(prefixErr);
      return;
    }
    setSaving(true);
    setError(null);
    try {
      const res = await fetch(`/api/settings/api-keys/${provider.id}`, {
        method: "PUT",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ key: key.trim() }),
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

  const prefixError = key.length > 0 ? validatePrefix(key) : null;

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="api-key-modal-title"
      onClick={onClose}
    >
      <div
        className="w-full max-w-md rounded-lg border border-gray-200 bg-background p-6 shadow-lg dark:border-gray-800"
        onClick={(e) => e.stopPropagation()}
      >
        <h2 id="api-key-modal-title" className="text-lg font-semibold">
          {isUpdate ? "Update" : "Add"} {provider.label} API key
        </h2>
        <p className="mt-2 text-sm text-gray-500 dark:text-gray-400">
          {provider.docsLabel}
        </p>
        <form onSubmit={(e) => void handleSubmit(e)} className="mt-5 space-y-4">
          <div className="space-y-1">
            <label htmlFor="api-key-input" className="text-sm font-medium">
              API Key
            </label>
            <Input
              id="api-key-input"
              type="password"
              placeholder={provider.placeholder}
              value={key}
              onChange={(e) => setKey(e.target.value)}
              autoComplete="off"
              required
            />
            {prefixError && (
              <p className="text-xs text-red-600 dark:text-red-400">{prefixError}</p>
            )}
          </div>
          <p className="text-xs text-gray-500 dark:text-gray-400">
            Your key is encrypted and never shared or logged.
          </p>
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
            <Button
              type="submit"
              disabled={saving || !key.trim() || !!prefixError}
            >
              {saving ? "Saving…" : "Save key"}
            </Button>
          </div>
        </form>
      </div>
    </div>
  );
}
