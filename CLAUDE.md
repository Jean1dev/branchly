# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

Branchly is a full-stack job automation platform integrated with GitHub. It consists of three services:
- **branchly-api** — Go backend (Gin + MongoDB + GitHub OAuth2)
- **branchly-runner** — Go job execution service (stub, port 8081)
- **branchly-web** — Next.js 16 frontend (React 19, TypeScript, Tailwind CSS)

## Development Commands

### Full Stack (Docker)
```bash
docker-compose up          # Start all services (API, runner, MongoDB)
docker-compose up --build  # Rebuild and start
```

### Frontend (branchly-web)
```bash
cd branchly-web
pnpm install
pnpm dev      # Dev server at http://localhost:3000
pnpm build
pnpm lint
```

Node >= 20.9.0 and pnpm 10.18.2 are required.

### Backend (branchly-api / branchly-runner)
```bash
cd branchly-api
go run ./cmd/api          # Run API locally
go build ./cmd/api        # Build binary
go test ./...             # Run all tests
go test ./internal/...    # Run tests in a specific package
```

### Environment Setup
Copy `branchly-api/.env.example` to `branchly-api/.env` and fill in:
- `GITHUB_CLIENT_ID` / `GITHUB_CLIENT_SECRET` — GitHub OAuth App credentials
- `JWT_SECRET` — random secret for JWT signing
- `ENCRYPTION_KEY` — 32-byte key (hex or raw) for AES-256-GCM token encryption
- `INTERNAL_API_SECRET` — shared secret for runner → API callbacks

## Architecture

### branchly-api (Clean Architecture)
```
cmd/api/          Entry point
internal/
  config/         Env-based config loading and validation
  domain/         Models: User, Job, Repository + repository interfaces
  handler/        HTTP handlers (thin, delegates to services)
  service/        Business logic
  repository/     MongoDB implementations of domain interfaces
  middleware/      JWT auth, CORS
  infra/          MongoDB client, AES-GCM crypto, HTTP clients
  respond/        Shared response helpers
```

Dependency flow: `handler → service → repository → MongoDB`

### API Routes
- `GET /health`
- `GET /auth/github`, `GET /auth/github/callback`, `GET /auth/me`, `POST /auth/logout`
- `GET|POST /repositories`, `DELETE /repositories/:id`, `GET /repositories/github`
- `GET|POST /jobs`, `GET /jobs/:id`, `GET /jobs/:id/logs` (SSE)
- `POST /internal/jobs/:id/status`, `POST /internal/jobs/:id/logs` (runner callbacks, protected by `INTERNAL_API_SECRET`)

### branchly-runner
Stub HTTP server on port 8081. Accepts `POST /jobs`, returns 202. Intended for GitHub Actions / CI integration later.

### branchly-web (Next.js App Router)
```
app/
  (app)/          Protected route group — dashboard, jobs, repositories, settings
  (marketing)/    Public route group — home, login
components/
  ui/             Reusable UI primitives
  features/       Feature-specific components
  layout/         Layout components
  providers/      React context providers
  skeletons/      Loading skeleton components
lib/
  mock-auth.tsx   Auth mock utilities (development)
  mock-data.ts    Fake data for development/UI testing
  utils.ts        Utility helpers
types/index.ts    Shared TypeScript types
```

### Data Flow
1. **Auth:** Frontend → `/auth/github` → GitHub OAuth → JWT cookie
2. **Jobs:** Frontend creates job → API saves to MongoDB → dispatches to Runner → Runner calls back via internal API → Frontend streams logs via SSE (`/jobs/:id/logs`)
3. **Repositories:** API fetches from GitHub using encrypted stored token

### MongoDB Collections
- `users` — unique on `(provider, provider_id)`, GitHub tokens encrypted with AES-256-GCM
- `repositories` — unique on `(user_id, github_repo_id)`
- `jobs` — indexed on `user_id`, `repository_id`, `status`; embedded `logs` array with timestamp + severity
