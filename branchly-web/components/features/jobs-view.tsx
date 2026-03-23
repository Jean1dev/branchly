"use client";

import { StatusBadge } from "@/components/features/status-badge";
import { EmptyState } from "@/components/features/empty-state";
import { PageHeader } from "@/components/layout/page-header";
import { Button } from "@/components/ui/button";
import { mockJobs } from "@/lib/mock-data";
import { formatDurationMs, truncate } from "@/lib/utils";
import type { JobStatus } from "@/types";
import Link from "next/link";
import { usePathname, useRouter, useSearchParams } from "next/navigation";
import { useCallback, useMemo } from "react";

const PAGE_SIZE = 10;

const tabs: { id: "all" | JobStatus; label: string }[] = [
  { id: "all", label: "All" },
  { id: "running", label: "Running" },
  { id: "completed", label: "Completed" },
  { id: "failed", label: "Failed" },
];

export function JobsView() {
  const router = useRouter();
  const pathname = usePathname();
  const searchParams = useSearchParams();
  const statusParam = searchParams.get("status") ?? "all";
  const pageParam = Number(searchParams.get("page") ?? "1") || 1;

  const activeTab = tabs.some((t) => t.id === statusParam)
    ? (statusParam as (typeof tabs)[number]["id"])
    : "all";

  const filtered = useMemo(() => {
    if (activeTab === "all") return mockJobs;
    return mockJobs.filter((j) => j.status === activeTab);
  }, [activeTab]);

  const totalPages = Math.max(1, Math.ceil(filtered.length / PAGE_SIZE));
  const page = Math.min(Math.max(1, pageParam), totalPages);
  const slice = useMemo(() => {
    const start = (page - 1) * PAGE_SIZE;
    return filtered.slice(start, start + PAGE_SIZE);
  }, [filtered, page]);

  const setQuery = useCallback(
    (next: { status?: string; page?: number }) => {
      const p = new URLSearchParams(searchParams.toString());
      if (next.status !== undefined) {
        if (next.status === "all") p.delete("status");
        else p.set("status", next.status);
        p.set("page", "1");
      }
      if (next.page !== undefined) p.set("page", String(next.page));
      router.push(`${pathname}?${p.toString()}`);
    },
    [pathname, router, searchParams]
  );

  return (
    <>
      <PageHeader
        title="Jobs"
        actions={
          <Button href="/jobs/new">New task</Button>
        }
      />
      <div
        className="mb-6 flex flex-wrap gap-2 border-b border-gray-200 pb-4 dark:border-gray-800"
        role="tablist"
        aria-label="Job status filters"
      >
        {tabs.map((t) => {
          const selected = activeTab === t.id;
          return (
            <button
              key={t.id}
              type="button"
              role="tab"
              aria-selected={selected}
              className={
                selected
                  ? "rounded-md bg-gray-100 px-3 py-1.5 text-sm font-medium text-foreground dark:bg-gray-900"
                  : "rounded-md px-3 py-1.5 text-sm font-medium text-gray-500 transition-colors duration-150 hover:text-foreground dark:text-gray-400"
              }
              onClick={() => setQuery({ status: t.id })}
            >
              {t.label}
            </button>
          );
        })}
      </div>
      {filtered.length === 0 ? (
        <EmptyState
          title="No jobs match this filter"
          description="Try another status or create a new task."
          action={{ label: "New task", href: "/jobs/new" }}
        />
      ) : (
        <>
          <div className="overflow-x-auto rounded-lg border border-gray-200 dark:border-gray-800">
            <table className="w-full min-w-[720px] text-left text-sm" role="table">
              <thead className="border-b border-gray-200 bg-gray-50 dark:border-gray-800 dark:bg-gray-950">
                <tr>
                  <th scope="col" className="px-4 py-3 font-medium">
                    Repository
                  </th>
                  <th scope="col" className="px-4 py-3 font-medium">
                    Task
                  </th>
                  <th scope="col" className="px-4 py-3 font-medium">
                    Status
                  </th>
                  <th scope="col" className="px-4 py-3 font-medium">
                    Duration
                  </th>
                  <th scope="col" className="px-4 py-3 font-medium">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody>
                {slice.map((job) => (
                  <tr
                    key={job.id}
                    className="border-b border-gray-200 last:border-0 dark:border-gray-800"
                  >
                    <td className="px-4 py-3 font-mono text-xs">
                      {job.repositoryName}
                    </td>
                    <td className="max-w-[280px] px-4 py-3 text-gray-600 dark:text-gray-300">
                      {truncate(job.prompt, 72)}
                    </td>
                    <td className="px-4 py-3">
                      <StatusBadge status={job.status} />
                    </td>
                    <td className="whitespace-nowrap px-4 py-3 text-gray-500 dark:text-gray-400">
                      {formatDurationMs(job.createdAt, job.completedAt)}
                    </td>
                    <td className="px-4 py-3">
                      <Link
                        href={`/jobs/${job.id}`}
                        className="text-sm font-medium hover:underline"
                      >
                        View
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <div className="mt-6 flex flex-col items-start justify-between gap-4 sm:flex-row sm:items-center">
            <p className="text-sm text-gray-500 dark:text-gray-400">
              Showing {(page - 1) * PAGE_SIZE + 1}–
              {Math.min(page * PAGE_SIZE, filtered.length)} of {filtered.length}
            </p>
            <div className="flex gap-2">
              <Button
                type="button"
                variant="secondary"
                size="sm"
                disabled={page <= 1}
                onClick={() => setQuery({ page: page - 1 })}
              >
                Previous
              </Button>
              <Button
                type="button"
                variant="secondary"
                size="sm"
                disabled={page >= totalPages}
                onClick={() => setQuery({ page: page + 1 })}
              >
                Next
              </Button>
            </div>
          </div>
        </>
      )}
    </>
  );
}
