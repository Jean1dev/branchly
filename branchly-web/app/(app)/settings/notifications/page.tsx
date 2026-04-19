"use client";

import { Card } from "@/components/ui/card";
import { PageHeader } from "@/components/layout/page-header";
import { unwrapApiData } from "@/lib/map-api";
import { useCallback, useEffect, useRef, useState } from "react";

interface EmailPrefs {
  enabled: boolean;
  on_job_completed: boolean;
  on_job_failed: boolean;
  on_pr_opened: boolean;
}

interface NotificationPrefs {
  email: EmailPrefs;
}

function Toggle({
  checked,
  onChange,
  disabled,
  id,
}: {
  checked: boolean;
  onChange: (v: boolean) => void;
  disabled?: boolean;
  id: string;
}) {
  return (
    <button
      id={id}
      role="switch"
      aria-checked={checked}
      disabled={disabled}
      onClick={() => onChange(!checked)}
      className={[
        "relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus-visible:ring-2 focus-visible:ring-gray-900 dark:focus-visible:ring-gray-100",
        checked
          ? "bg-gray-900 dark:bg-gray-100"
          : "bg-gray-200 dark:bg-gray-700",
        disabled ? "cursor-not-allowed opacity-50" : "",
      ].join(" ")}
    >
      <span
        className={[
          "pointer-events-none inline-block h-5 w-5 rounded-full bg-white shadow-lg ring-0 transition duration-200 ease-in-out dark:bg-gray-900",
          checked ? "translate-x-5" : "translate-x-0",
        ].join(" ")}
      />
    </button>
  );
}

function ToggleRow({
  id,
  label,
  description,
  checked,
  onChange,
  disabled,
}: {
  id: string;
  label: string;
  description?: string;
  checked: boolean;
  onChange: (v: boolean) => void;
  disabled?: boolean;
}) {
  return (
    <div className="flex items-center justify-between gap-4 py-3">
      <div>
        <label htmlFor={id} className="text-sm font-medium cursor-pointer">
          {label}
        </label>
        {description && (
          <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
            {description}
          </p>
        )}
      </div>
      <Toggle id={id} checked={checked} onChange={onChange} disabled={disabled} />
    </div>
  );
}

export default function NotificationsPage() {
  const [prefs, setPrefs] = useState<NotificationPrefs | null>(null);
  const [loading, setLoading] = useState(true);
  const [savedAt, setSavedAt] = useState<number | null>(null);
  const saveTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  const load = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetch("/api/user/notification-preferences");
      if (!res.ok) return;
      const json: unknown = await res.json();
      const data = unwrapApiData<{ notification_preferences: NotificationPrefs }>(json);
      if (data?.notification_preferences) {
        setPrefs(data.notification_preferences);
      }
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const save = useCallback(async (patch: Partial<EmailPrefs>) => {
    try {
      const res = await fetch("/api/user/notification-preferences", {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(patch),
      });
      if (res.ok) {
        setSavedAt(Date.now());
        if (saveTimer.current) clearTimeout(saveTimer.current);
        saveTimer.current = setTimeout(() => setSavedAt(null), 2500);
      }
    } catch {
      // silent — load will resync on next mount
    }
  }, []);

  function update(patch: Partial<EmailPrefs>) {
    if (!prefs) return;
    const next: NotificationPrefs = {
      email: { ...prefs.email, ...patch },
    };
    setPrefs(next);
    void save(patch);
  }

  if (loading) {
    return (
      <>
        <PageHeader title="Notifications" />
        <div className="max-w-2xl space-y-4">
          {[1, 2].map((i) => (
            <div
              key={i}
              className="h-24 animate-pulse rounded-lg bg-gray-100 dark:bg-gray-900"
            />
          ))}
        </div>
      </>
    );
  }

  if (!prefs) {
    return (
      <>
        <PageHeader title="Notifications" />
        <p className="text-sm text-gray-500 dark:text-gray-400">
          Could not load notification preferences.
        </p>
      </>
    );
  }

  const emailDisabled = !prefs.email.enabled;

  return (
    <>
      <PageHeader
        title="Notifications"
        description="Choose which email notifications you want to receive."
      />
      <div className="max-w-2xl space-y-6">
        <Card className="p-6">
          <div className="flex items-center justify-between mb-4">
            <div>
              <p className="font-semibold text-sm">Email notifications</p>
              <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
                Master switch — disabling this stops all email notifications.
              </p>
            </div>
            <Toggle
              id="email-enabled"
              checked={prefs.email.enabled}
              onChange={(v) => update({ enabled: v })}
            />
          </div>

          <div
            className={[
              "border-t border-gray-100 dark:border-gray-800 divide-y divide-gray-100 dark:divide-gray-800",
              emailDisabled ? "opacity-50" : "",
            ].join(" ")}
          >
            <ToggleRow
              id="on_job_completed"
              label="Job completed successfully"
              description="Get notified when a job finishes without errors."
              checked={prefs.email.on_job_completed}
              onChange={(v) => update({ on_job_completed: v })}
              disabled={emailDisabled}
            />
            <ToggleRow
              id="on_job_failed"
              label="Job failed"
              description="Get notified when a job encounters an error."
              checked={prefs.email.on_job_failed}
              onChange={(v) => update({ on_job_failed: v })}
              disabled={emailDisabled}
            />
            <ToggleRow
              id="on_pr_opened"
              label="Pull request opened"
              description="Get notified when a pull request is opened on GitHub or GitLab."
              checked={prefs.email.on_pr_opened}
              onChange={(v) => update({ on_pr_opened: v })}
              disabled={emailDisabled}
            />
          </div>
        </Card>

        {savedAt && (
          <p className="text-xs text-gray-500 dark:text-gray-400">
            Preferences saved.
          </p>
        )}
      </div>
    </>
  );
}
