# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an AI-powered code analysis platform that helps development teams onboard faster, maintain code quality, and modernize legacy codebases through intelligent conversation with their repositories.

**Technology Stack:**
- Frontend: SvelteKit (Bun runtime)
- Backend: Go 1.21+ with Gin framework + oapi-codegen
- API Documentation: OpenAPI 3.0 specification (auto-generated)
- Database: MongoDB 8.0
- Vector Database: Qdrant 1.7+
- Embedding Model: Voyage AI (voyage-code-3)
- LLM: Claude 4 sonnet
- Containerization: Docker Compose

## Development Commands

Since this is a new project with detailed specifications but no implementation yet, the following commands are planned:

**Backend (Go):**
```bash
# Run backend server
go run cmd/server/main.go

# Build backend
go build -o bin/server cmd/server/main.go

# Test backend
go test ./...

# Lint backend
golangci-lint run
```

**Frontend (SvelteKit):**
```bash
# Install dependencies
bun install

# Run development server
bun run dev

# Build frontend
bun run build

# Preview production build
bun run preview
```

**Docker Environment:**
```bash
# Start all services
docker-compose up

# Start in background
docker-compose up -d

# Stop all services
docker-compose down

# Rebuild and start
docker-compose up --build
```

## Architecture Overview

The system follows a microservices architecture with the following core components:

### Backend Architecture
- **API Layer**: Gin framework with oapi-codegen for type-safe OpenAPI implementation
- **Authentication**: JWT-based authentication with bcrypt password hashing
- **Database Layer**: MongoDB for document storage, Qdrant for vector embeddings
- **External APIs**: GitHub API for repository access, Voyage AI for embeddings, Claude for chat responses
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
5. **Context Construction**: Relevant chunks are assembled into prompts for Claude
6. **Response Generation**: Claude provides context-aware answers referencing specific code

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
8. **AI Chat**: Complete conversational interface with Claude integration
9. **Polish**: Error handling, performance optimization, demo preparation

## Key Implementation Notes

### OpenAPI-First Development
This project uses oapi-codegen for type-safe API development:
- Define endpoints in `api/openapi.yaml`
- Generate server stubs and client types
- Automatic request/response validation
- Built-in Swagger UI documentation

### Embedding Strategy
Code chunks are optimized for retrieval:
- 150 lines per chunk with 50-line overlap
- Language-aware chunking respecting function boundaries
- Metadata extraction (functions, classes, imports)
- Content deduplication via SHA256 hashing

### Performance Considerations
- Batch processing for repository imports
- Concurrent goroutines with semaphore limiting
- Caching for unchanged file embeddings
- Progress tracking for long-running operations

### Demo-Ready Features
The application is designed for compelling demonstrations:
- Strategic dashboard with cost savings calculations
- Real-time import progress indicators
- Conversational AI that references specific code
- Professional UI with smooth animations

## Environment Setup

Required environment variables:
```bash
JWT_SECRET=your-secret-key
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret
VOYAGE_API_KEY=your-voyage-api-key
ANTHROPIC_API_KEY=your-anthropic-api-key
MONGODB_URI=mongodb://localhost:27017/github-analyzer
QDRANT_URL=http://localhost:6333
```

## Demo User Access

A demo user is automatically created on application startup:

**Login Credentials:**
- **Email:** demo@github-analyzer.com
- **Password:** demo123456

No user registration is available - access is managed by administrators.

## Project Status

This project is currently in the planning phase with comprehensive specifications completed. The next steps involve implementing the foundation slice with Docker environment setup, basic Go server, and SvelteKit frontend initialization.

The implementation follows a 21-day timeline with each slice building incrementally toward a complete MVP demonstrating AI-powered code analysis capabilities.
