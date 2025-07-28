# Backend – AI Code Fixing Platform

A production-ready Go (1.24+) service that powers the AI Code Fixing Platform.  
It exposes a **type-safe REST API** (OpenAPI 3.1) for automated code fixing, AST-based analysis, knowledge graph queries, and repository-wide understanding. Transforms from "smart text search" into an automated code fixing engine that generates complete, validated solutions for technical debt and code issues.

---

## ✨ Features

- **Automated Code Fixing** – AST-based problem detection with generated solutions and validation
- **Knowledge Graph Analysis** – Neo4j integration for code relationships and dependency traversal
- **Multi-Modal Understanding** – combines code, comments, tests, and documentation for context
- **Hierarchical Code Summarization** – semantic clustering from functions to system architecture
- **Program Dependence Graphs** – control and data flow analysis for change impact prediction
- **Incremental Analysis** – real-time updates with smart caching and change propagation
- **Fix Validation System** – syntax, compilation, behavioral, and security validation
- **Repository-Level Reasoning** – CodePlan-inspired planning for complex architectural queries
- **OpenAPI-first Architecture** – single source of truth (`api/openapi.yaml`) with generated types
- **Gin-powered REST API** – minimal, ultra-fast router with comprehensive middleware support
- **Authentication & Authorization** – JWT sessions, GitHub OAuth, route-based RBAC
- **Vector & Graph Search** – Qdrant + Neo4j for semantic similarity and relationship queries
- **MongoDB Storage** – metadata persistence with advanced indexing for complex queries
- **Background Processing** – embedding pipeline, AST analysis, and fix generation workers
- **Modular Architecture** – clean separation with `handlers`, `services`, `models`, `middleware`
- **Container-first Deployment** – multi-stage `Dockerfile` + `docker-compose` for all environments

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
│   ├── database/        # Mongo, Qdrant & Neo4j clients
│   ├── handlers/        # HTTP route handlers (fix generation, analysis)
│   ├── middleware/      # Gin middleware (auth, logging, validation)
│   ├── models/          # Domain entities & persistence (fixes, analysis)
│   ├── services/        # Business logic (AST analysis, fix generation, validation)
│   └── server/          # HTTP server wiring
├── scripts/             # Helper bash scripts
├── Dockerfile           # Production image
└── Dockerfile.dev       # Hot-reload dev image
```

---

## 🛠️ Local Development

### Prerequisites

- Go 1.24+
- Docker / Docker Compose (for Mongo + Qdrant + Neo4j)
- Make (optional but recommended)
- Tree-sitter parsers (for AST analysis)

### Getting Started

```bash
cd backend

# Spin up Mongo, Qdrant, Neo4j plus a hot-reload server
go run ./cmd/server        # or: make backend-dev

# or via Docker Compose (recommended)
make up env=dev            # mounts source & reflex-rebuild with all databases
```

The dev compose file (`../docker-compose.dev.yml`) exposes the API on <http://localhost:8080> with full infrastructure stack.

---

## 🔧 Configuration

All configuration is supplied via **environment variables** and parsed in `internal/config/config.go`. Below is a non-exhaustive table:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 8080 | HTTP listen port |
| `MONGODB_URI` | mongodb://mongo:27017/acip | Mongo connection string |
| `QDRANT_URL` | http://qdrant:6334 | Qdrant vector database API |
| `NEO4J_URI` | bolt://neo4j:7687 | Neo4j knowledge graph database |
| `NEO4J_USER` / `NEO4J_PASSWORD` | neo4j/password | Neo4j authentication |
| `JWT_SECRET` | – | HMAC secret for JWT tokens |
| `GITHUB_CLIENT_ID` / `GITHUB_CLIENT_SECRET` | – | GitHub OAuth credentials |
| `LLM_API_KEY` | – | API key for OpenAI-compatible LLM |
| `EMBEDDING_API_KEY` | – | API key for embedding model |
| `ENABLE_AST_ANALYSIS` | true | Enable AST-based code analysis |
| `ENABLE_KNOWLEDGE_GRAPH` | true | Enable knowledge graph features |
| `ANALYSIS_DEPTH` | semantic | Analysis depth: basic, ast, semantic, full |
| `TREE_SITTER_PATH` | /usr/local/lib/tree-sitter | Path to tree-sitter parsers |

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