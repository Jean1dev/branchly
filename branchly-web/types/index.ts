export type JobStatus = "pending" | "running" | "completed" | "failed";

export type JobLogLevel = "info" | "success" | "warning" | "error";

export interface User {
  id: string;
  name: string;
  email: string;
  avatar: string;
  githubUsername: string;
}

export interface Repository {
  id: string;
  fullName: string;
  defaultBranch: string;
  language: string;
  lastJobAt: string;
  jobsCount: number;
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
  prompt: string;
  status: JobStatus;
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
