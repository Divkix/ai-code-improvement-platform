# Backend – AI Code Improvement Platform

A production-ready Go (1.24+) service that powers the AI Code Improvement Platform.  
It exposes a **type-safe REST API** (OpenAPI 3.1) for semantic code search, chat-based RAG, GitHub repository ingestion and dashboard analytics.  The backend is stateless and can run anywhere Docker is supported.

---

## ✨ Features

- **Gin-powered REST API** – minimal, ultra-fast router with middleware support.
- **OpenAPI-first** – single source of truth (`api/openapi.yaml`) with generated server stubs and client SDKs.
- **Authentication & Authorization** – JWT sessions, GitHub OAuth, route-based RBAC.
- **Vector Search** – Qdrant integration for code embeddings & hybrid BM25 queries.
- **MongoDB Storage** – metadata & user/session persistence with index-aware queries.
- **Background Workers** – repository embedding pipeline and scheduled maintenance jobs.
- **Modular Layers** – `handlers`, `services`, `models`, `middleware`, `database` for clean separation of concerns.
- **Container-first** – multi-stage `Dockerfile` + `docker-compose` for local and prod.

---

## 📂 Project Structure

```
backend/
├── api/                 # OpenAPI spec + generation helpers
├── cmd/                 # CLI entrypoints (server, utilities)
│   └── server/
│       └── main.go
├── internal/
│   ├── auth/            # JWT utilities, OAuth helpers
│   ├── config/          # env var parsing & validation
│   ├── database/        # Mongo & Qdrant clients
│   ├── handlers/        # HTTP route handlers
│   ├── middleware/      # Gin middleware (auth, logging)
│   ├── models/          # Domain entities & persistence helpers
│   ├── services/        # Business logic (chat, embeddings, GH sync)
│   └── server/          # HTTP server wiring
├── scripts/             # Helper bash scripts
├── Dockerfile           # Production image
└── Dockerfile.dev       # Hot-reload dev image
```

---

## 🛠️ Local Development

### Prerequisites

- Go 1.24+
- Docker / Docker Compose (for Mongo + Qdrant)
- Make (optional but recommended)

### Getting Started

```bash
cd backend

# Spin up Mongo & Qdrant plus a hot-reload server
go run ./cmd/server        # or: make backend-dev

# or via Docker Compose (recommended)
make up env=dev            # mounts source & reflex-rebuild
```

The dev compose file (`../docker-compose.dev.yml`) exposes the API on <http://localhost:8080>.

---

## 🔧 Configuration

All configuration is supplied via **environment variables** and parsed in `internal/config/config.go`. Below is a non-exhaustive table:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | HTTP listen port |
| `MONGODB_URI` | mongodb://mongo:27017/github-analyzer | Mongo connection string |
| `QDRANT_URL` | http://qdrant:6334 | Qdrant API base |
| `JWT_SECRET` | – | HMAC secret for JWT tokens |
| `GITHUB_CLIENT_ID` / `GITHUB_CLIENT_SECRET` | – | GitHub OAuth credentials |
| `LLM_API_KEY` | – | API key for OpenAI-compatible LLM |
| `EMBEDDING_API_KEY` | – | API key for embedding model |

### Loading `.env` Locally

```bash
cp .env.example .env
# edit values as needed
source .env
```

---

## 🧪 Testing

```bash
# Run unit tests with race detection & coverage
make test            # equivalent to: go test ./... -race -cover

# Static analysis & linting
make validate        # go vet + golangci-lint + go test
```

Test coverage reports are generated in `coverage.out`. Combine with tools like `go tool cover` or `gocov` for HTML visualization.

---

## 📝 API Documentation

The server hosts Swagger-UI at `/docs` when `ENVIRONMENT != prod`.

Regenerate server stubs & client SDK after modifying the spec:

```bash
make generate        # runs go generate + openapi-generator
```

---

## 📦 Docker

```bash
# Build production image
docker build -t ai-code-platform-backend .

# Run
docker run -p 8080:8080 --env-file .env ai-code-platform-backend
```

Multi-stage build produces a slim image (~25 MB) with **distroless** base.

---

## 🚀 Deployment

The backend is stateless; scale horizontally behind a load-balancer.  
For Kubernetes, see the sample manifest in `deploy/k8s` (if/when added).

---

## 🤝 Contributing Guidelines

1. Fork / branch from `main`.
2. Follow Conventional Commits for commit messages.
3. Ensure `make validate` passes before opening a PR.
4. Write tests for new functionality.

---

## 📄 License

MIT © 2024 – Contributors to the AI Code Improvement Platform 