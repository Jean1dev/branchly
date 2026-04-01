export type JobStatus = "pending" | "running" | "completed" | "failed";

export type AgentType = "claude-code" | "gemini";

export type GitProvider = "github" | "gitlab" | "azure-devops";

export const AGENTS = [
  {
    id: "claude-code" as AgentType,
    name: "Claude Code",
    provider: "Anthropic",
    description: "Best quality. Uses your Anthropic API key.",
    badge: null,
  },
  {
    id: "gemini" as AgentType,
    name: "Gemini CLI",
    provider: "Google",
    description: "Free tier available. Great for testing.",
    badge: "Free tier",
  },
] as const;

export type JobLogLevel = "info" | "success" | "warning" | "error";

export interface User {
  id: string;
  name: string;
  email: string;
  avatar: string;
  githubUsername: string;
}

export interface GitIntegration {
  id: string;
  provider: GitProvider;
  tokenType: "oauth" | "pat";
  connectedAt: string;
}

export interface Repository {
  id: string;
  integrationId: string;
  provider: GitProvider;
  externalId: string;
  fullName: string;
  cloneUrl: string;
  defaultBranch: string;
  language: string;
  lastJobAt: string;
  jobsCount: number;
}

export interface ProviderRepo {
  externalId: string;
  fullName: string;
  cloneUrl: string;
  defaultBranch: string;
  language: string;
  provider: GitProvider;
}

export interface JobCost {
  inputTokens: number;
  outputTokens: number;
  totalTokens: number;
  estimatedUSD: number;
  modelUsed: string;
  durationSecs: number;
  isEstimate: boolean;
}

export interface Job {
  id: string;
  repositoryId: string;
  repositoryName: string;
  repositoryProvider: GitProvider;
  prompt: string;
  status: JobStatus;
  agentType: AgentType;
  branchName: string;
  prUrl: string | null;
  createdAt: string;
  completedAt: string | null;
  cost: JobCost | null;
}

export interface JobLog {
  timestamp: string;
  level: JobLogLevel;
  message: string;
}
