import type { Job, JobLog, Repository, User } from "@/types";

export const mockUser: User = {
  id: "user_01",
  name: "Lucas Mendes",
  email: "lucas@example.com",
  avatar: "https://api.dicebear.com/7.x/initials/svg?seed=LM",
  githubUsername: "lucasmendes",
};

export const mockRepositories: Repository[] = [
  {
    id: "repo_01",
    fullName: "lucasmendes/ecommerce-api",
    defaultBranch: "main",
    language: "Go",
    lastJobAt: "2024-01-15T10:30:00Z",
    jobsCount: 12,
  },
  {
    id: "repo_02",
    fullName: "lucasmendes/frontend-dashboard",
    defaultBranch: "main",
    language: "TypeScript",
    lastJobAt: "2024-01-14T16:45:00Z",
    jobsCount: 7,
  },
  {
    id: "repo_03",
    fullName: "lucasmendes/auth-service",
    defaultBranch: "develop",
    language: "Go",
    lastJobAt: "2024-01-13T09:00:00Z",
    jobsCount: 3,
  },
];

export const mockJobs: Job[] = [
  {
    id: "job_01",
    repositoryId: "repo_01",
    repositoryName: "lucasmendes/ecommerce-api",
    prompt:
      "Add pagination to the /products endpoint using cursor-based pagination",
    status: "completed",
    branchName: "branchly/add-pagination-products",
    prUrl: "https://github.com/lucasmendes/ecommerce-api/pull/47",
    createdAt: "2024-01-15T10:30:00Z",
    completedAt: "2024-01-15T10:34:22Z",
  },
  {
    id: "job_02",
    repositoryId: "repo_02",
    repositoryName: "lucasmendes/frontend-dashboard",
    prompt: "Create a reusable DataTable component with sorting and filtering",
    status: "running",
    branchName: "branchly/create-datatable-component",
    prUrl: null,
    createdAt: "2024-01-15T11:00:00Z",
    completedAt: null,
  },
  {
    id: "job_03",
    repositoryId: "repo_01",
    repositoryName: "lucasmendes/ecommerce-api",
    prompt: "Write unit tests for the order service",
    status: "failed",
    branchName: "branchly/unit-tests-order-service",
    prUrl: null,
    createdAt: "2024-01-14T16:45:00Z",
    completedAt: "2024-01-14T16:51:00Z",
  },
];

export const mockJobLogs: JobLog[] = [
  { timestamp: "10:34:01", level: "info", message: "Job started" },
  {
    timestamp: "10:34:02",
    level: "info",
    message: "Cloning repository lucasmendes/ecommerce-api...",
  },
  {
    timestamp: "10:34:05",
    level: "info",
    message: "Repository cloned successfully",
  },
  {
    timestamp: "10:34:06",
    level: "info",
    message: "Indexing codebase (47 files)...",
  },
  {
    timestamp: "10:34:08",
    level: "info",
    message: "Codebase indexed. Sending context to agent...",
  },
  {
    timestamp: "10:34:10",
    level: "info",
    message: "Agent analyzing task: 'Add pagination...'",
  },
  {
    timestamp: "10:34:15",
    level: "info",
    message: "Editing file: handlers/products.go",
  },
  {
    timestamp: "10:34:18",
    level: "info",
    message: "Editing file: models/pagination.go (new file)",
  },
  {
    timestamp: "10:34:22",
    level: "success",
    message: "Changes committed to branch branchly/add-pagination-products",
  },
  {
    timestamp: "10:34:23",
    level: "success",
    message: "Pull Request opened: https://github.com/...",
  },
];

export function getRepositoryById(id: string): Repository | undefined {
  return mockRepositories.find((r) => r.id === id);
}

export function getJobById(id: string): Job | undefined {
  return mockJobs.find((j) => j.id === id);
}

export function getJobsByRepositoryId(repositoryId: string): Job[] {
  return mockJobs.filter((j) => j.repositoryId === repositoryId);
}
