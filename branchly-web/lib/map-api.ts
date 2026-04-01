import type { AgentType, GitIntegration, GitProvider, Job, JobCost, JobLog, JobLogLevel, JobStatus, ProviderRepo, Repository } from "@/types";

export function unwrapApiData<T>(json: unknown): T {
  if (
    json !== null &&
    typeof json === "object" &&
    !Array.isArray(json) &&
    "data" in json
  ) {
    return (json as { data: T }).data;
  }
  return json as T;
}

export function parseApiErrorMessage(json: unknown, status: number): string {
  if (json !== null && typeof json === "object" && "error" in json) {
    const err = (json as { error?: { message?: string } }).error;
    if (typeof err?.message === "string" && err.message.length > 0) {
      return err.message;
    }
  }
  return `Request failed (${status})`;
}

export type ApiRepository = {
  id: string;
  integration_id: string;
  provider: string;
  external_id: string;
  full_name: string;
  clone_url: string;
  default_branch: string;
  language: string;
  connected_at: string;
};

export type ApiIntegration = {
  id: string;
  provider: string;
  token_type: string;
  connected_at: string;
};

export type ApiProviderRepo = {
  external_id: string;
  full_name: string;
  clone_url: string;
  default_branch: string;
  language: string;
  provider: string;
};

export type ApiJobCost = {
  input_tokens: number;
  output_tokens: number;
  total_tokens: number;
  estimated_usd: number;
  model_used: string;
  duration_secs: number;
  is_estimate: boolean;
};

export type ApiJob = {
  id: string;
  repository_id: string;
  prompt: string;
  status: string;
  agent_type?: string;
  branch_name: string;
  pr_url?: string;
  created_at: string;
  updated_at: string;
  completed_at?: string | null;
  logs?: Array<{ timestamp: string; level: string; message: string }>;
  cost?: ApiJobCost | null;
};

function mapJobStatus(s: string): JobStatus {
  if (
    s === "pending" ||
    s === "running" ||
    s === "completed" ||
    s === "failed"
  ) {
    return s;
  }
  return "pending";
}

function mapLogLevel(l: string): JobLogLevel {
  if (l === "info" || l === "success" || l === "warning" || l === "error") {
    return l;
  }
  return "info";
}

function mapGitProvider(s: string | undefined): GitProvider {
  if (s === "gitlab") return "gitlab";
  if (s === "azure-devops") return "azure-devops";
  return "github";
}

export function formatLogTimestamp(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso;
  return new Intl.DateTimeFormat("en-US", {
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
    hour12: false,
  }).format(d);
}

export function mapJobLog(e: {
  timestamp: string;
  level: string;
  message: string;
}): JobLog {
  return {
    timestamp: formatLogTimestamp(e.timestamp),
    level: mapLogLevel(e.level),
    message: e.message,
  };
}

export function mapIntegration(r: ApiIntegration): GitIntegration {
  return {
    id: r.id,
    provider: mapGitProvider(r.provider),
    tokenType: r.token_type === "pat" ? "pat" : "oauth",
    connectedAt: r.connected_at,
  };
}

export function mapRepository(r: ApiRepository): Repository {
  return {
    id: r.id,
    integrationId: r.integration_id ?? "",
    provider: mapGitProvider(r.provider),
    externalId: r.external_id ?? "",
    fullName: r.full_name,
    cloneUrl: r.clone_url ?? "",
    defaultBranch: r.default_branch,
    language: r.language || "—",
    lastJobAt: r.connected_at,
    jobsCount: 0,
  };
}

export function mapProviderRepo(r: ApiProviderRepo): ProviderRepo {
  return {
    externalId: r.external_id,
    fullName: r.full_name,
    cloneUrl: r.clone_url,
    defaultBranch: r.default_branch,
    language: r.language || "—",
    provider: mapGitProvider(r.provider),
  };
}

function mapJobCost(c: ApiJobCost): JobCost {
  return {
    inputTokens: c.input_tokens,
    outputTokens: c.output_tokens,
    totalTokens: c.total_tokens,
    estimatedUSD: c.estimated_usd,
    modelUsed: c.model_used,
    durationSecs: c.duration_secs,
    isEstimate: c.is_estimate,
  };
}

function mapAgentType(s: string | undefined): AgentType {
  if (s === "gemini") return "gemini";
  return "claude-code";
}

export function mapJob(
  j: ApiJob,
  repositoryName: string | undefined,
  repositoryProvider?: string | undefined
): Job {
  return {
    id: j.id,
    repositoryId: j.repository_id,
    repositoryName: repositoryName ?? j.repository_id,
    repositoryProvider: mapGitProvider(repositoryProvider),
    prompt: j.prompt,
    status: mapJobStatus(j.status),
    agentType: mapAgentType(j.agent_type),
    branchName: j.branch_name,
    prUrl: j.pr_url ?? null,
    createdAt: j.created_at,
    completedAt: j.completed_at ?? null,
    cost: j.cost ? mapJobCost(j.cost) : null,
  };
}

export function jobRepoNameMap(
  repos: ApiRepository[]
): Record<string, string> {
  const list = Array.isArray(repos) ? repos : [];
  return Object.fromEntries(
    list.map((r) => [r.id, r.full_name] as const)
  );
}

export function jobRepoProviderMap(
  repos: ApiRepository[]
): Record<string, GitProvider> {
  const list = Array.isArray(repos) ? repos : [];
  return Object.fromEntries(
    list.map((r) => [r.id, mapGitProvider(r.provider)] as const)
  );
}
