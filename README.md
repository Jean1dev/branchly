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
| `ENCRYPTION_KEY` | 32-byte hex key for AES-256-GCM token encryption |
| `RUNNER_URL` | Runner service URL |
| `FRONTEND_URL` | Frontend origin (for CORS) |
| `ALLOWED_ORIGINS` | Comma-separated allowed CORS origins |
| `INTERNAL_API_SECRET` | Shared secret for internal endpoints |

> GitHub OAuth credentials (`GITHUB_CLIENT_ID` / `GITHUB_CLIENT_SECRET`) belong in **branchly-web** (NextAuth), not the API.
