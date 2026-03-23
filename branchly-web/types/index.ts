export type JobStatus = "completed" | "running" | "failed";

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
}

export interface JobLog {
  timestamp: string;
  level: JobLogLevel;
  message: string;
}
