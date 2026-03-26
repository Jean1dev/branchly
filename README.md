# Branchly

A full-stack job automation platform integrated with GitHub. Create and monitor automated jobs tied to your repositories, with real-time log streaming.

## Services

| Service | Stack | Port |
|---|---|---|
| **branchly-api** | Go, Gin, MongoDB | 8080 |
| **branchly-runner** | Go (stub) | 8081 |
| **branchly-web** | Next.js 16, React 19, TypeScript | 3000 |

## Prerequisites

- Docker & Docker Compose
- Node.js >= 20.9.0 + pnpm 10.18.2 (frontend only)
- Go 1.22+ (backend only)
- A GitHub OAuth App

## Quick Start (Docker)

**1. Configure the API**

```bash
cp branchly-api/.env.example branchly-api/.env
```

Edit `branchly-api/.env`:

```env
JWT_SECRET=<random-secret>
ENCRYPTION_KEY=<64-char-hex-string>   # 32-byte AES-256-GCM key
INTERNAL_API_SECRET=<shared-secret>   # must match branchly-web
```

**2. Configure the frontend**

```bash
cp branchly-web/.env.example branchly-web/.env
```

Set your GitHub OAuth credentials and the `INTERNAL_API_SECRET` (same value as the API).

**3. Start all services**

```bash
docker-compose up --build
```

The app will be available at `http://localhost:3000`.

## Development

### Backend (branchly-api)

```bash
cd branchly-api
go run ./cmd/api        # Start dev server at :8080
go build ./cmd/api      # Build binary
go test ./...           # Run all tests
```

### Frontend (branchly-web)

```bash
cd branchly-web
pnpm install
pnpm dev                # Dev server at http://localhost:3000
pnpm build
pnpm lint
```

## Architecture

### Auth Flow

1. User signs in via GitHub through **NextAuth** (branchly-web)
2. NextAuth calls `POST /internal/auth/upsert` on the API with the GitHub token
3. API stores the user (token encrypted with AES-256-GCM) and returns a JWT
4. JWT is stored in the NextAuth session; browser communicates only with Next.js `/api/*` proxy

### API Routes

**Public**
- `GET /health`

**Authenticated** (JWT Bearer)
- `GET /auth/me` — current user
- `POST /auth/logout`
- `GET /repositories` — connected repositories
- `POST /repositories` — connect a repository
- `DELETE /repositories/:id`
- `GET /repositories/github` — list GitHub repos available to connect
- `GET /jobs` — list jobs
- `POST /jobs` — create a job
- `GET /jobs/:id`
- `GET /jobs/:id/logs` — SSE stream of job logs

**Internal** (`X-Internal-Secret` header)
- `POST /internal/auth/upsert` — called by NextAuth on sign-in
- `POST /internal/jobs/:id/status` — runner status callback
- `POST /internal/jobs/:id/logs` — runner log callback

### Backend Structure (Clean Architecture)

```
branchly-api/
  cmd/api/          Entry point
  internal/
    config/         Env-based config
    domain/         Models + repository interfaces
    handler/        HTTP handlers
    service/        Business logic
    repository/     MongoDB implementations
    middleware/      JWT auth, CORS, internal secret
    infra/          MongoDB client, AES-GCM crypto, runner HTTP client
    respond/        Shared response helpers
```

### Frontend Structure

```
branchly-web/
  app/
    (app)/          Protected routes — dashboard, jobs, repositories, settings
    (marketing)/    Public routes — home, login
    api/            Next.js API routes (proxy to branchly-api)
  components/
    ui/             Reusable UI primitives
    features/       Feature-specific components
    layout/         Layout components
    providers/      React context providers
  lib/
    auth.ts         NextAuth configuration
    api-client.ts   Typed API client
  types/index.ts    Shared TypeScript types
```

### MongoDB Collections

| Collection | Notes |
|---|---|
| `users` | Unique on `(provider, provider_id)`; GitHub token encrypted |
| `repositories` | Unique on `(user_id, github_repo_id)` |
| `jobs` | Indexed on `user_id`, `repository_id`, `status`; embedded `logs` array |

## Environment Variables

### branchly-api

| Variable | Description |
|---|---|
| `PORT` | HTTP port (default `8080`) |
| `MONGODB_URI` | MongoDB connection string |
| `MONGODB_DATABASE` | Database name (default `branchly`) |
| `JWT_SECRET` | Secret for JWT signing |
| `JWT_TTL_DAYS` | JWT expiry in days (default `7`) |
| `ENCRYPTION_KEY` | 64-char hex (32 bytes) for AES-256-GCM. **Must match the runner.** Generate: `openssl rand -hex 32` |
| `MAX_ACTIVE_JOBS_PER_USER` | Max concurrent jobs per user (default `3`) |
| `RUNNER_URL` | Runner service URL |
| `FRONTEND_URL` | Frontend origin (for CORS) |
| `ALLOWED_ORIGINS` | Comma-separated allowed CORS origins |
| `INTERNAL_API_SECRET` | Shared secret for internal endpoints |

> GitHub OAuth credentials (`GITHUB_CLIENT_ID` / `GITHUB_CLIENT_SECRET`) belong in **branchly-web** (NextAuth), not the API.

---

## Security

### 1 — GitHub token encryption (AES-256-GCM)

GitHub OAuth tokens are **never stored in plaintext**. At sign-in the API encrypts each token with AES-256-GCM before writing it to MongoDB. The runner decrypts the token in memory, uses it for git operations and the GitHub API call, then explicitly zeros the variable — the plaintext never touches disk or logs.

**Generate a valid key:**

```bash
openssl rand -hex 32
```

Set the output as `ENCRYPTION_KEY` in **both** `branchly-api/.env` and `branchly-runner/.env`. The two services must share the same key — the API encrypts, the runner decrypts.

> In `docker-compose.yml` both services load their key from their respective `.env` files. Ensure both files contain the **same** `ENCRYPTION_KEY` value.

### 2 — Per-user rate limiting

A user can have at most **3 active jobs** (`pending` + `running`) simultaneously. Attempting to create a 4th returns **HTTP 429**:

```json
{
  "error": {
    "code": "RATE_LIMIT_EXCEEDED",
    "message": "you have reached the maximum of 3 active jobs — wait for one to complete"
  }
}
```

Configurable via `MAX_ACTIVE_JOBS_PER_USER` (default: `3`).

### 3 — Repository ownership validation (defence in depth)

| Layer | Where | What is checked |
|---|---|---|
| **API** | `job_service.Create` | The `repository_id` belongs to the authenticated user. Returns HTTP **404** for both "not found" and "wrong owner" — never 403 — to avoid leaking repository existence. |
| **Runner** | `executor.Run` (before any I/O) | The repository document in MongoDB has `user_id == job.user_id` **and** `full_name == repository_name` from the payload. A tampered payload that swaps the clone URL while keeping a valid ID is rejected before any filesystem or network operation. |

Security-relevant errors are logged server-side with full context (`job_id`, `user_id`, `repository_id`) but return generic messages to clients.
