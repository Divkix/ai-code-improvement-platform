# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an AI-powered code analysis platform that helps development teams onboard faster, maintain code quality, and modernize legacy codebases through intelligent conversation with their repositories.

**Technology Stack:**
- Frontend: SvelteKit (Bun runtime)
- Backend: Go 1.24+ with Gin framework + oapi-codegen
- API Documentation: OpenAPI 3.0 specification (auto-generated)
- Database: MongoDB 8.0
- Vector Database: Qdrant 1.7+
- Embedding Model: Voyage AI (voyage-code-3)
- LLM: OpenAI-compatible (GPT-4o-mini default) or Claude 4 sonnet
- Containerization: Docker Compose

## Development Commands

**Root Level (Make commands):**
```bash
# Start all services in production mode
make up

# Start all services in development mode (with hot reload)
make up env=dev

# Stop all services and clean up
make down
make clean  # Also removes volumes

# View logs (all services or specific: make logs service=backend)
make logs

# Get shell access to containers
make sh service=backend

# Generate Go code from OpenAPI spec
make generate

# Build backend binary
make build

# Run backend in development mode
make backend-dev

# Validate OpenAPI spec and run linting
make validate

# Run all tests with coverage
make test

# Run tests for a specific service
cd backend && go test -v ./internal/services/...
```

**Backend (Go) - Direct Commands:**
```bash
# From backend/ directory:
cd backend

# Generate API code from OpenAPI spec
go generate ./internal/generated/...

# Run backend server
go run cmd/server/main.go

# Build backend
go build -o bin/server cmd/server/main.go

# Test backend with coverage
go test -v -race -coverprofile=coverage.out ./...

# Lint backend
golangci-lint run

# View test coverage
go tool cover -html=coverage.out -o coverage.html
```

**Frontend (SvelteKit):**
```bash
# From frontend/ directory:
cd frontend

# Install dependencies
bun install

# Run development server
bun run dev

# Build frontend
bun run build

# Preview production build
bun run preview

# Type checking
bun run check

# Linting and formatting
bun run lint
bun run format

# Generate TypeScript types from OpenAPI
bun run generate-api

# Run tests
bun run test
bun run test:unit
bun run test:e2e
```

**Docker Environment:**
```bash
# Production mode
docker-compose up --build -d
docker-compose down

# Development mode (with hot reload)
docker-compose -f docker-compose.dev.yml up --build -d
docker-compose -f docker-compose.dev.yml down
```

## Architecture Overview

The system follows a microservices architecture with the following core components:

### Backend Architecture
- **API Layer**: Gin framework with oapi-codegen for type-safe OpenAPI implementation
- **Authentication**: JWT-based authentication with bcrypt password hashing
- **Database Layer**: MongoDB for document storage, Qdrant for vector embeddings
- **External APIs**: GitHub API for repository access, Voyage AI for embeddings, OpenAI-compatible LLM or Claude for chat responses
- **Processing Pipeline**: Async repository import with code chunking and embedding generation

### Key Data Models
- **Users**: Authentication and GitHub token storage (encrypted)
- **Repositories**: GitHub repository metadata and import status tracking
- **CodeChunks**: Processed code segments with metadata and vector references
- **ChatSessions**: Conversation history with retrieved context chunks
- **AnalyticsEvents**: Usage tracking and optimization metrics

### RAG Pipeline
The core intelligence comes from a Retrieval-Augmented Generation (RAG) pipeline:

1. **Code Chunking**: Files are split into 150-line overlapping chunks
2. **Embedding Generation**: Chunks are converted to vectors using Voyage AI's voyage-code-3 model
3. **Vector Storage**: Embeddings stored in Qdrant with metadata
4. **Query Processing**: User questions are embedded and matched against code chunks
5. **Context Construction**: Relevant chunks are assembled into prompts for the LLM
6. **Response Generation**: LLM provides context-aware answers referencing specific code

### Search Capabilities
The platform provides sophisticated multi-modal search functionality:
- **Text Search**: MongoDB full-text search with relevance scoring
- **Vector Search**: Semantic similarity using Qdrant embeddings
- **Hybrid Search**: Configurable text + vector weight combination
- **Similar Code**: Code recommendation based on semantic similarity
- **Advanced Filtering**: Language, file type, and repository scoping

### Frontend Architecture
- **SvelteKit**: TypeScript-based reactive framework
- **State Management**: Svelte stores for authentication and chat state
- **API Client**: Type-safe client generated from OpenAPI specification
- **UI Components**: Responsive design with TailwindCSS
- **Chat Interface**: Real-time conversation with syntax highlighting

## Implementation Strategy

The project uses vertical slicing - each slice delivers a complete, testable feature:

1. **Foundation**: Docker environment, basic authentication, health checks
2. **Dashboard**: Visual metrics showing code analysis value and cost savings
3. **Repository Management**: CRUD operations for repository tracking
4. **GitHub Integration**: OAuth authentication and repository import with progress tracking
5. **Code Processing**: File fetching, chunking, and metadata extraction
6. **Search Foundation**: Basic text search through code chunks
7. **Vector RAG**: Semantic search with embeddings and vector storage
8. **AI Chat**: Complete conversational interface with LLM integration
9. **Polish**: Error handling, performance optimization, demo preparation

## Key Implementation Notes

### OpenAPI-First Development
This project uses oapi-codegen for type-safe API development:
- Define endpoints in `backend/api/openapi.yaml`
- Run `make generate` or `go generate ./internal/generated/...` to generate server stubs and client types
- Server interface is implemented in `internal/server/server.go` which delegates to individual handlers
- Frontend API types are generated with `bun run generate-api`
- Built-in Swagger UI available at `/docs/`
- **CRITICAL**: Always regenerate types after modifying the OpenAPI spec before running tests

### Development Workflow
- All code files should start with 2-line ABOUTME comments explaining the file's purpose
- Use `golangci-lint run` for Go linting (never skip this step)
- Services are stateless and follow dependency injection patterns
- Database operations use MongoDB transactions where data consistency is critical

### Code Architecture Patterns
- **Handler Pattern**: Each API domain (auth, github, search, etc.) has dedicated handlers in `internal/handlers/`
- **Service Layer**: Business logic is encapsulated in services (`internal/services/`)
- **Repository Pattern**: Database operations are abstracted through repository interfaces
- **Dependency Injection**: All components are wired together in `cmd/server/main.go`
- **Generated Types**: API types and interfaces are generated from OpenAPI spec in `internal/generated/`

### Embedding Pipeline Architecture
The system uses a sophisticated background processing pipeline:
- **EmbeddingPipeline** (`internal/services/embedding_pipeline.go`): Manages background job queue
- **EmbeddingService** (`internal/services/embedding.go`): Handles vector generation and storage
- **Multiple Providers**: Supports both Voyage AI and local embedding models
- **Batch Processing**: Concurrent processing with configurable worker pools
- **Progress Tracking**: Real-time status updates stored in MongoDB
- **Status Management**: Uses EmbeddingStatus enum (pending, processing, completed, failed)
- **Error Recovery**: Failed chunks can be reprocessed without affecting completed ones

### Code Chunking Strategy
Files are processed into optimal chunks for retrieval:
- 150 lines per chunk with 50-line overlap to preserve context
- Language-aware chunking that respects function/class boundaries
- Metadata extraction (functions, classes, imports, file paths)
- Content deduplication via SHA256 hashing to avoid redundant embeddings
- Stored in MongoDB with vector references in Qdrant

### Configuration Management
Environment-based configuration in `internal/config/config.go`:
- Supports both Voyage AI (`EMBEDDING_PROVIDER=voyage`) and local models (`EMBEDDING_PROVIDER=local`)
- Vector dimensions configurable via `VECTOR_DIMENSION` (256, 512, 1024, 2048 for Voyage)
- Database connections, API keys, and server settings all configurable

## Environment Setup

Required environment variables:
```bash
# Core Configuration
JWT_SECRET=your-secret-key
MONGODB_URI=mongodb://localhost:27017/github-analyzer
QDRANT_URL=http://localhost:6334

# GitHub OAuth
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret
GITHUB_ENCRYPTION_KEY=your-16-24-32-byte-aes-key

# AI Services - LLM (OpenAI-compatible)
LLM_BASE_URL=https://api.openai.com/v1
LLM_MODEL=gpt-4o-mini
LLM_API_KEY=your-llm-api-key
LLM_REQUEST_TIMEOUT=30s

# Legacy Anthropic support (optional)
LLM_API_KEY=your-llm-api-key

# Embedding Provider (voyage or local)
EMBEDDING_PROVIDER=voyage
VOYAGE_API_KEY=your-voyage-api-key  # Required if EMBEDDING_PROVIDER=voyage
LOCAL_EMBEDDING_URL=http://localhost:1234  # Required if EMBEDDING_PROVIDER=local
LOCAL_EMBEDDING_MODEL=text-embedding-nomic-embed-text-v1.5  # For local provider

# Vector Configuration
VECTOR_DIMENSION=1024  # Must be 256, 512, 1024, or 2048 for Voyage
QDRANT_COLLECTION_NAME=codechunks

# Server Configuration (Optional)
PORT=8080
HOST=0.0.0.0
GIN_MODE=debug

# Frontend Configuration (.env in frontend/)
VITE_API_URL=http://localhost:8080
```

## Demo User Access

A demo user is automatically created on application startup:

**Login Credentials:**
- **Email:** demo@github-analyzer.com
- **Password:** demo123456

No user registration is available - access is managed by administrators.

## Testing and Quality

**Backend Testing:**
```bash
# Run all tests with coverage from backend/ directory
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Run specific test package
go test -v ./internal/services/...

# Run tests with verbose output
go test -v -run TestEmbeddingService ./internal/services/
```

**Frontend Testing:**
```bash
# From frontend/ directory
bun run test        # All tests (unit + e2e) - uses npm run internally
bun run test:unit   # Vitest unit tests
bun run test:e2e    # Playwright e2e tests
```

**Linting:**
```bash
# Backend linting
cd backend && golangci-lint run

# Frontend linting  
cd frontend && bun run lint
```

## Project Status

This is a fully implemented AI-powered code analysis platform with the following completed features:

**✅ Completed:**
- Docker containerization with development and production modes
- Go backend with OpenAPI-first development using oapi-codegen
- SvelteKit frontend with TypeScript and TailwindCSS
- MongoDB for document storage, Qdrant for vector embeddings
- GitHub OAuth integration for repository access
- Background embedding pipeline with job queue management
- Support for both Voyage AI and local embedding models
- Vector-based semantic search and hybrid search capabilities
- Real-time repository import with progress tracking
- Dashboard with analytics and cost savings calculations
- Full test coverage for critical components

The platform demonstrates enterprise-ready AI-powered code analysis with sophisticated RAG (Retrieval-Augmented Generation) capabilities for conversational code exploration.
