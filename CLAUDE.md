# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an AI-powered automated code fixing platform that transforms from "smart text search" into an automated code fixing engine. It generates complete, validated solutions for technical debt and code issues through AST-based analysis, knowledge graphs, and multi-modal understanding.

**Technology Stack:**
- Frontend: SvelteKit (Bun runtime) with fix validation UI
- Backend: Go 1.24+ with Gin framework + oapi-codegen
- API Documentation: OpenAPI 3.0 specification (auto-generated)
- Database: MongoDB 8.0 (metadata, fixes, analysis results)
- Vector Database: Qdrant 1.7+ (semantic embeddings)
- Knowledge Graph: Neo4j (code relationships, dependencies)
- AST Analysis: Tree-sitter parsers (40+ languages)
- Embedding Model: Voyage AI (voyage-code-3)
- LLM: OpenAI-compatible (GPT-4o-mini default) or Claude 4 sonnet
- Containerization: Docker Compose with full infrastructure stack

## Important Code Conventions

- All code files MUST start with 2-line ABOUTME comments explaining the file's purpose
- Use 4 spaces for indentation (not tabs)
- NEVER use `--no-verify` when committing code
- Always run `golangci-lint run` for Go linting after making changes
- Match existing code style and formatting within each file

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
bun run generate  # Alias for generate-api

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
- **Database Layer**: MongoDB for document storage, Qdrant for vector embeddings, Neo4j for knowledge graphs
- **AST Analysis Engine**: Tree-sitter parsers for structural code understanding across 40+ languages
- **Fix Generation Engine**: Problem detection, solution planning, code generation, and validation
- **Knowledge Graph Service**: Code relationships, dependencies, and architectural pattern analysis
- **Multi-Modal Analysis**: Combines code, comments, tests, documentation, and commit history
- **External APIs**: GitHub API for repository access, Voyage AI for embeddings, OpenAI-compatible LLM
- **Processing Pipeline**: AST analysis, knowledge graph population, incremental change tracking

### Key Data Models
- **Users**: Authentication and GitHub token storage (encrypted)
- **Repositories**: GitHub repository metadata, analysis status, and knowledge graph references
- **CodeChunks**: AST-processed code segments with structural metadata and vector references
- **GeneratedFixes**: Complete fix solutions with validation results and confidence scores
- **KnowledgeGraphNodes**: AST nodes, relationships, and dependency information
- **ProgramDependenceGraphs**: Control and data flow analysis for change impact prediction
- **ChatSessions**: Conversation history with enhanced context retrieval
- **AnalyticsEvents**: Usage tracking, fix success rates, and cost savings metrics

### Enhanced Analysis Pipeline
The core intelligence combines multiple analysis approaches for automated code fixing:

1. **AST-Based Code Analysis**: Tree-sitter parsers extract structural information (functions, classes, dependencies)
2. **Knowledge Graph Population**: Code relationships stored in Neo4j for traversal and impact analysis
3. **Program Dependence Analysis**: Control and data flow graphs for change impact prediction
4. **Multi-Modal Context Extraction**: Combines code, comments, tests, docs, and commit history
5. **Hierarchical Summarization**: Semantic clustering from functions to system architecture
6. **Fix Generation Pipeline**: Problem detection → solution planning → code generation → validation
7. **Incremental Updates**: Real-time change tracking with smart caching and propagation
8. **Repository-Level Reasoning**: CodePlan-inspired planning for complex architectural queries

### Analysis and Fix Capabilities
The platform provides comprehensive code analysis and automated fixing:
- **Structural Analysis**: AST-based understanding of code organization and relationships
- **Knowledge Graph Queries**: Traverse code dependencies and architectural patterns
- **Semantic Search**: Vector similarity using Qdrant embeddings with context enrichment
- **Multi-Modal Context**: Combine code structure, comments, tests, and documentation
- **Fix Generation**: Automated problem detection with validated solution generation
- **Impact Analysis**: Predict change effects through program dependence graphs
- **Incremental Processing**: Smart caching with real-time updates and change propagation

### Frontend Architecture
- **SvelteKit**: TypeScript-based reactive framework with fix validation UI
- **State Management**: Svelte stores for authentication, fix tracking, and analysis state
- **API Client**: Type-safe client generated from OpenAPI specification
- **UI Components**: Responsive design with TailwindCSS and fix visualization
- **Fix Generation Interface**: Real-time progress tracking and validation result display
- **Analysis Dashboard**: Code structure visualization and fix success metrics

## Implementation Strategy

The project implements a 15-month roadmap transforming from text-based analysis to automated code fixing:

**Phase 1 (Months 1-3): Foundation Enhancement**
- AST-based code analysis engine with Tree-sitter parsers
- Knowledge graph infrastructure using Neo4j
- Enhanced metadata extraction beyond simple text patterns

**Phase 2 (Months 4-6): Semantic Understanding**
- Program dependence graph implementation for control/data flow
- Hierarchical code summarization with semantic clustering
- Multi-modal context integration (code + comments + tests + docs)

**Phase 3 (Months 7-9): Repository-Level Reasoning**
- CodePlan-inspired planning system for complex queries
- Advanced context window management with compression
- Repository-wide architectural understanding

**Phase 4 (Months 10-12): Real-Time Updates**
- Incremental analysis engine with smart caching
- Change propagation system for impact tracking
- 90% reduction in processing time for updates

**Phase 5 (Months 13-15): Automated Fix Generation**
- Complete fix generation engine with validation
- Multi-level validation (syntax, compilation, behavior, security)
- 90%+ automation of common code fixes and technical debt elimination

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
- Configurable chunk sizes via environment variables (CHUNK_SIZE, CHUNK_OVERLAP_SIZE)
- Default: 30 lines per chunk with 10-line overlap (optimized for local embedding models)
- For Voyage AI: Recommend 150 lines per chunk with 50-line overlap
- Language-aware chunking that respects function/class boundaries
- Metadata extraction (functions, classes, imports, file paths)
- Content deduplication via SHA256 hashing to avoid redundant embeddings
- Stored in MongoDB with vector references in Qdrant

### Configuration Management
Environment-based configuration in `internal/config/config.go`:
- All configuration uses environment variables loaded from `.env` file
- **Database configs**: `MONGODB_URI`, `QDRANT_URL`, `NEO4J_URI`, `NEO4J_USER`, `NEO4J_PASSWORD`
- **AI service configs**: `EMBEDDING_BASE_URL`, `LLM_BASE_URL`, `EMBEDDING_API_KEY`, `LLM_API_KEY`
- **Analysis configs**: `ENABLE_AST_ANALYSIS`, `ENABLE_KNOWLEDGE_GRAPH`, `ANALYSIS_DEPTH`, `TREE_SITTER_PATH`
- **Performance configs**: `MAX_CONCURRENT_WORKERS`, `ANALYSIS_CACHE_SIZE`, `MAX_CONTEXT_TOKENS`
- **Chunking strategy**: `CHUNK_SIZE` and `CHUNK_OVERLAP_SIZE` (default: 30/10 lines for local models, 150/50 for Voyage AI)
- Vector dimensions configurable via `VECTOR_DIMENSION` (256, 512, 768, 1024, 2048)
- See `.env.example` for complete configuration reference

## Environment Setup

Copy `.env.example` to `.env` and configure the required variables:

**Essential Variables (Must be set):**
```bash
# Authentication
JWT_SECRET=your-secret-key

# Database connections
MONGODB_URI=mongodb://mongodb:27017/acip  # Use 'mongodb' for Docker
QDRANT_URL=http://localhost:6334
NEO4J_URI=bolt://neo4j:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=password

# AI Services
EMBEDDING_BASE_URL=https://api.openai.com/v1
EMBEDDING_MODEL=voyage-code-3
EMBEDDING_API_KEY=your-api-key

LLM_BASE_URL=https://api.openai.com/v1
LLM_MODEL=gpt-4o-mini
LLM_API_KEY=your-api-key

# GitHub OAuth (for repository import)
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret
GITHUB_ENCRYPTION_KEY=your-16-24-32-byte-aes-key

# Frontend
VITE_API_URL=http://localhost:8080

# Analysis Configuration
ENABLE_AST_ANALYSIS=true
ENABLE_KNOWLEDGE_GRAPH=true
ANALYSIS_DEPTH=semantic
TREE_SITTER_PATH=/usr/local/lib/tree-sitter
```

**Important Notes:**
- For Docker: Use service names (`mongodb`, `neo4j`) as hostnames, `localhost` for direct runs
- Local embedding models: Set `EMBEDDING_BASE_URL=http://host.docker.internal:1234/v1`
- AST analysis requires Tree-sitter parsers for supported languages
- See `.env.example` for complete configuration with defaults and comments

## Demo User Access

A demo user is automatically created on application startup:

**Login Credentials:**
- **Email:** demo@acip.com
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

# Note: Package management uses Bun, but some scripts delegate to npm for compatibility
```

**Linting:**
```bash
# Backend linting
cd backend && golangci-lint run

# Frontend linting  
cd frontend && bun run lint
```

## Project Status

This is an AI-powered automated code fixing platform implementing a 15-month roadmap to transform from text analysis to automated code fixing:

**✅ Current Foundation (Phase 0):**
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

**🚧 In Development (Phases 1-5):**
- AST-based code analysis engine with Tree-sitter parsers (Phase 1)
- Neo4j knowledge graph infrastructure (Phase 1)  
- Program dependence graphs for control/data flow analysis (Phase 2)
- Hierarchical code summarization with semantic clustering (Phase 2)
- Multi-modal context integration (Phase 3)
- Incremental analysis with smart caching (Phase 4)
- Complete automated fix generation engine (Phase 5)

The platform is evolving from conversational code exploration to automated code fixing with validated solution generation.

## Key File Locations

**Backend Architecture:**
- `cmd/server/main.go` - Main application entry point and dependency injection
- `internal/server/server.go` - HTTP server implementation with route handlers
- `internal/handlers/` - API endpoint handlers (auth, github, search, chat, etc.)
- `internal/services/` - Business logic layer (embedding, repository, search, etc.)
- `internal/models/` - Data models and MongoDB schemas
- `internal/generated/` - Auto-generated API types from OpenAPI spec
- `api/openapi.yaml` - OpenAPI specification (source of truth for API)

**Frontend Architecture:**
- `src/routes/` - SvelteKit page routes and API endpoints
- `src/lib/components/` - Reusable Svelte components
- `src/lib/api/` - API client and type definitions
- `src/lib/stores/` - Svelte stores for state management

**Configuration:**
- `.env.example` - Complete environment variable reference
- `docker-compose.yml` - Production container orchestration
- `docker-compose.dev.yml` - Development with hot-reload
- `Makefile` - Development commands and shortcuts

## Troubleshooting

**Common Issues:**
1. **Tests failing after OpenAPI changes**: Run `make generate` to regenerate types
2. **Frontend API errors**: Regenerate types with `bun run generate-api`
3. **Docker networking**: Use service names (`mongodb`, `neo4j`) in Docker, `localhost` for direct runs
4. **Embedding pipeline stuck**: Check Qdrant connection and vector dimensions match
5. **Neo4j connection errors**: Verify `NEO4J_URI`, `NEO4J_USER`, and `NEO4J_PASSWORD` are set correctly
6. **AST analysis failures**: Ensure Tree-sitter parsers are installed at `TREE_SITTER_PATH`
7. **Build failures**: Ensure `golangci-lint` is installed and run `make validate`
8. **Knowledge graph queries failing**: Check Neo4j service is running and accessible
