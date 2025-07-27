# ACIP - AI Code Improvement Platform

<!-- Uncomment once CI is live -->
<!--[![CI](https://github.com/your-org/ai-code-improvement-platform/actions/workflows/ci.yml/badge.svg)](https://github.com/your-org/ai-code-improvement-platform/actions/workflows/ci.yml)-->

A fully-containerised end-to-end system that helps developers explore, understand and improve codebases using modern AI techniques.

## Table of Contents

- [ACIP - AI Code Improvement Platform](#acip---ai-code-improvement-platform)
  - [Table of Contents](#table-of-contents)
  - [✨ Key Features](#-key-features)
  - [🏗️ Tech Stack](#️-tech-stack)
  - [📂 Repository Layout (top-level)](#-repository-layout-top-level)
  - [🚀 Quick Start (Production-like)](#-quick-start-production-like)
    - [Local Development Hot-Reload](#local-development-hot-reload)
  - [🛠️ Backend Development](#️-backend-development)
    - [Environment Variables (partial list)](#environment-variables-partial-list)
    - [API Documentation](#api-documentation)
  - [🌐 Frontend Development](#-frontend-development)
    - [Regenerate Typed API Client](#regenerate-typed-api-client)
  - [🧪 Testing All Services](#-testing-all-services)
  - [🖥️ Makefile Cheat-Sheet](#️-makefile-cheat-sheet)
  - [🤝 Contributing](#-contributing)
  - [📄 License](#-license)

---

## ✨ Key Features

- **Semantic Code Search** – vector & hybrid search across all indexed repositories.
- **AI Chat (RAG)** – ask questions about your code in natural language; answers are grounded in retrieved code snippets.
- **Repository Dashboard** – language statistics, recent activity and trend insights.
- **GitHub OAuth** – securely connect your GitHub account for repository import.
- **Automated Embedding Pipeline** – background worker keeps vectors in-sync as code changes.
- **Type-Safe API** – single OpenAPI spec shared by Go backend & SvelteKit frontend.

---

## 🏗️ Tech Stack

| Layer       | Technology |
|-------------|------------|
| **Frontend**| SvelteKit + TypeScript, TailwindCSS, openapi-fetch |
| **Backend** | Go 1.22 (Gin), MongoDB, Qdrant, OpenAI-compatible LLM/Embedding APIs |
| **Auth**    | JWT for session, GitHub OAuth for repository access |
| **Infra**   | Docker Compose, Makefile utilities |

---

## 📂 Repository Layout (top-level)

```
backend/    # Go services, API, generation & Dockerfiles
frontend/   # SvelteKit application
mongodb-init/  # Mongo seed user scripts
Makefile    # helper commands (docker compose, build, test)
docker-compose[.dev].yml  # multi-service orchestration
```

---

## 🚀 Quick Start (Production-like)

The simplest way to spin everything up is Docker Compose:

```bash
# Build & start in the background
make up               # defaults to env=prod → docker-compose.yml

# Tail logs
make logs

# Shutdown
make down
```

The **prod** stack uses the regular `Dockerfile` images for both frontend and backend.

### Local Development Hot-Reload

```bash
# Start dev versions (hot-reload, vite, go run, etc.)
make up env=dev       # uses docker-compose.dev.yml
```

Behind the scenes the dev compose file mounts your source directories and uses **`Dockerfile.dev`** images that have nodemon / reflex / vite dev servers.

---

## 🛠️ Backend Development

```bash
cd backend

# Run unit tests with race detector & coverage
make test

# Lint & vet
make validate

# Hot-reload standalone (without docker)
make backend-dev
```

### Environment Variables (partial list)

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | HTTP listen port |
| `MONGODB_URI` | mongodb://localhost:27017/github-analyzer | Mongo connection string |
| `QDRANT_URL` | http://localhost:6334 | Qdrant gRPC/HTTP base |
| `JWT_SECRET` | *required* | HMAC secret for JWT tokens |
| `GITHUB_CLIENT_ID` / `GITHUB_CLIENT_SECRET` | – | GitHub OAuth credentials |
| `LLM_API_KEY` | *required* | API key for your OpenAI-compatible LLM |
| `EMBEDDING_API_KEY` | – | API key for embedding provider (may reuse LLM key) |

See `backend/internal/config/config.go` for the full configuration matrix.

### API Documentation

Swagger-UI is automatically served at:

```
http://localhost:8080/docs/
```

The spec lives in `backend/api/openapi.yaml`. Convert to JSON via:

```bash
make -C backend openapi-json
```

---

## 🌐 Frontend Development

```bash
cd frontend
bun install            # or npm / pnpm / yarn
cp .env.example .env.local
bun run dev            # localhost:3000
```

`VITE_API_URL` must point to the backend URL (defaults to `http://localhost:8080`).

### Regenerate Typed API Client

```bash
bun run generate-api   # parses backend OpenAPI spec → src/lib/api/types.ts
```

Unit tests: `bun run test:unit`

E2E tests (Playwright): `bun run test:e2e`

---

## 🧪 Testing All Services

1. Ensure **MongoDB** & **Qdrant** services are running (`make up env=dev`).
2. Run backend tests: `make test`.
3. Run frontend unit & e2e tests: `cd frontend && bun run test`.

---

## 🖥️ Makefile Cheat-Sheet

```
make up [env=dev|prod]       # build & start stack
make down [env=..]           # stop stack
make clean                   # down + remove volumes
make logs service=<name>     # follow logs
make generate                # go generate API stubs
make backend-dev             # run backend with hot reload
```

---

## 🤝 Contributing

1. Fork & clone the repo.
2. Create a feature branch: `git checkout -b feat/my-feature`.
3. Run `make validate` to ensure code passes lints/tests.
4. Submit a pull request – please describe the motivation and include screenshots / logs where relevant.

---

## 📄 License

This project is licensed under the MIT License. See the `LICENSE` file for details.
