# Slice 7: Vector RAG Implementation Plan

## Overview
Transform the current text search into a powerful Vector RAG system with semantic code understanding using Voyage AI embeddings and Qdrant vector database.

## Phase 1: Dependencies & Infrastructure

### 1.1 Update Go Dependencies
```bash
cd backend
go get github.com/qdrant/go-client@latest
go mod tidy
```

### 1.2 Environment Variables  
Add to `.env`:
```bash
VOYAGE_API_KEY=your-voyage-api-key
QDRANT_COLLECTION_NAME=code_chunks
VECTOR_DIMENSION=1536
```

### 1.3 Configuration Updates
Update `backend/internal/config/config.go`:
- Add VoyageAPIKey field
- Add QdrantCollectionName field  
- Add VectorDimension field

## Phase 2: Core Services Implementation

### 2.1 Create Voyage AI Service
**File: `backend/internal/services/voyage.go`**
- HTTP client for Voyage AI API
- `GenerateEmbeddings(texts []string) ([][]float32, error)` method
- voyage-code-3 model integration
- Batch processing (max 128 texts per request)
- Rate limiting compliance (60 RPM)
- Error handling and retries

### 2.2 Enhanced Qdrant Service  
**File: `backend/internal/database/qdrant.go`** (enhance existing)
- Full Qdrant client with vector operations
- `CreateCollection()` method with 1536-dim vectors, Cosine distance
- `UpsertPoints()` method for batch vector storage
- `SearchSimilar()` method for vector similarity search  
- `DeletePoints()` method for cleanup
- Health checks and connection management

### 2.3 Embedding Processing Service
**File: `backend/internal/services/embedding.go`**
- `ProcessRepository(repoID string)` - batch embed all chunks
- `ProcessCodeChunk(chunk *models.CodeChunk)` - embed single chunk
- Progress tracking with status updates
- Error recovery and retry logic
- MongoDB vector ID storage integration

## Phase 3: API Layer Updates

### 3.1 OpenAPI Specification
**File: `backend/api/openapi.yaml`** - Add endpoints:
- `POST /api/search/vector` - pure vector search
- `POST /api/search/hybrid` - combined text + vector  
- `GET /api/search/similar/{chunkId}` - find similar chunks
- `POST /api/repositories/{id}/embed` - trigger embedding processing
- `GET /api/repositories/{id}/embedding-status` - check progress

### 3.2 New Handlers
**File: `backend/internal/handlers/vector_search.go`**
- VectorSearchHandler for semantic search
- HybridSearchHandler combining text + vector scores
- SimilarChunksHandler for finding related code
- EmbeddingStatusHandler for progress tracking

### 3.3 Enhanced Search Service
**File: `backend/internal/services/search.go`** (enhance existing)
- `VectorSearch()` method using Qdrant similarity search
- `HybridSearch()` method combining text + vector scores (0.7 vector + 0.3 text)
- Result deduplication and ranking
- Performance optimizations with caching

## Phase 4: Data Models Updates

### 4.1 Search Models
**File: `backend/internal/models/codechunk.go`** (enhance existing)
- Add VectorSearchRequest struct
- Add HybridSearchRequest struct  
- Add SimilarityResult struct with distance scores
- Add EmbeddingStatus enum (pending, processing, completed, failed)

### 4.2 Repository Models
**File: `backend/internal/models/repository.go`** (enhance existing)
- Add EmbeddingStatus field to Repository struct
- Add EmbeddedChunksCount field
- Add LastEmbeddedAt timestamp field

## Phase 5: Background Processing

### 5.1 Embedding Pipeline
**File: `backend/internal/services/embedding_pipeline.go`**
- Background job processing existing repositories
- Batch processing with configurable size (50 chunks/batch)
- Progress tracking in MongoDB
- Error handling and partial failure recovery
- Admin endpoint to trigger re-indexing

### 5.2 Real-time Processing
Enhance `backend/internal/services/repository.go`:
- Trigger embedding for new code chunks
- Async processing to avoid blocking imports
- Queue-based processing with retry logic

## Phase 6: Frontend Integration

### 6.1 API Client Updates
**File: `frontend/src/lib/api/client.ts`** (enhance existing)
- Add vectorSearch() method
- Add hybridSearch() method  
- Add getSimilarChunks() method
- Add repository embedding status methods

### 6.2 Search Components Enhancement
**File: `frontend/src/lib/components/SearchBox.svelte`** (enhance existing)
- Add search mode toggle (text/vector/hybrid)
- Add "semantic search" option checkbox
- Enhanced search filters for vector search

**File: `frontend/src/lib/components/SearchResults.svelte`** (enhance existing)  
- Display vector similarity scores
- Show "semantic similarity" indicators
- Enhanced result ranking visualization
- Similar chunks suggestions

### 6.3 New Components
**File: `frontend/src/lib/components/EmbeddingStatus.svelte`**
- Show repository embedding progress
- Display embedding statistics
- Trigger re-embedding controls

## Phase 7: Testing Implementation

### 7.1 Backend Tests
**File: `backend/internal/services/voyage_test.go`**
- Test embedding generation
- Test batch processing
- Test error handling

**File: `backend/internal/database/qdrant_test.go`**  
- Test vector operations
- Test collection management
- Test search accuracy

**File: `backend/internal/services/embedding_test.go`**
- Test end-to-end pipeline
- Test progress tracking
- Test error recovery

### 7.2 Integration Tests
**File: `backend/test/vector_search_test.go`**
- Test hybrid search accuracy
- Test performance benchmarks
- Test vector-text score combination

## Phase 8: Database Migrations & Setup

### 8.1 MongoDB Indexes
Add to `backend/internal/services/search.go` EnsureIndexes():
- Index on `vectorId` field
- Index on repository embedding status
- Compound indexes for hybrid search

### 8.2 Qdrant Setup
**File: `backend/internal/database/qdrant_setup.go`**
- Initialize code_chunks collection
- Configure vector parameters (1536-dim, Cosine)
- Setup payload indexing for metadata filtering

## Phase 9: Performance & Monitoring

### 9.1 Caching Layer
- Cache frequent vector searches
- Cache embedding results for unchanged chunks
- Implement TTL for vector cache

### 9.2 Metrics Collection
- Track embedding generation time
- Monitor vector search performance  
- Log hybrid search accuracy metrics

## Implementation Order

1. **Dependencies & Config** (0.5 day)
2. **VoyageService** (1 day)  
3. **Enhanced QdrantService** (1.5 days)
4. **EmbeddingService** (1.5 days)
5. **API Endpoints** (1 day)
6. **Enhanced SearchService** (1.5 days)
7. **Frontend Updates** (2 days)
8. **Background Pipeline** (1.5 days) 
9. **Testing & Integration** (2 days)
10. **Performance Optimization** (1.5 days)

**Total Estimated Time: 12 days**

## Testing Verification

After implementation, verify:
- [ ] Vector embeddings generated for all code chunks
- [ ] Hybrid search returns relevant semantic results
- [ ] Performance meets requirements (sub-500ms search)
- [ ] Background processing handles large repositories
- [ ] Frontend displays vector similarity scores
- [ ] Error handling works for API failures
- [ ] Repository embedding status tracking works

This comprehensive plan transforms the current text search into a powerful Vector RAG system that understands code semantically, providing developers with intelligent code discovery capabilities.