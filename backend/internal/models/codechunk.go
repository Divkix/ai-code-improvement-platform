// ABOUTME: CodeChunk model for storing processed code segments with metadata
// ABOUTME: Includes MongoDB tags, validation, and helper methods for search operations

package models

import (
    "crypto/sha256"
    "fmt"
    "time"

    "go.mongodb.org/mongo-driver/bson/primitive"
)

// CodeChunk represents a processed code segment stored in MongoDB
type CodeChunk struct {
    ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    RepositoryID primitive.ObjectID `bson:"repositoryId" json:"repositoryId"`
    FilePath     string             `bson:"filePath" json:"filePath"`
    FileName     string             `bson:"fileName" json:"fileName"`
    Language     string             `bson:"language" json:"language"`
    StartLine    int                `bson:"startLine" json:"startLine"`
    EndLine      int                `bson:"endLine" json:"endLine"`
    Content      string             `bson:"content" json:"content"`
    ContentHash  string             `bson:"contentHash" json:"contentHash"`
    Imports      []string           `bson:"imports,omitempty" json:"imports,omitempty"`
    Metadata     ChunkMetadata      `bson:"metadata" json:"metadata"`
    VectorID     string             `bson:"vectorId,omitempty" json:"vectorId,omitempty"`
    CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
    UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// ChunkMetadata contains extracted metadata from code chunk
type ChunkMetadata struct {
    Functions  []string `bson:"functions,omitempty" json:"functions,omitempty"`
    Classes    []string `bson:"classes,omitempty" json:"classes,omitempty"`
    Variables  []string `bson:"variables,omitempty" json:"variables,omitempty"`
    Types      []string `bson:"types,omitempty" json:"types,omitempty"`
    Complexity int      `bson:"complexity,omitempty" json:"complexity,omitempty"`
}

// SearchResult represents a search result with relevance score
type SearchResult struct {
    CodeChunk
    Score     float64 `json:"score"`
    Highlight string  `json:"highlight,omitempty"`
}

// SearchRequest represents search request parameters
type SearchRequest struct {
    Query        string `json:"query" binding:"required"`
    RepositoryID string `json:"repositoryId,omitempty"`
    Language     string `json:"language,omitempty"`
    FileType     string `json:"fileType,omitempty"`
    Limit        int    `json:"limit,omitempty"`
    Offset       int    `json:"offset,omitempty"`
}

// SearchResponse represents search response with pagination
type SearchResponse struct {
    Results []SearchResult `json:"results"`
    Total   int64          `json:"total"`
    HasMore bool           `json:"hasMore"`
    Query   string         `json:"query"`
}

// VectorSearchRequest represents vector search request parameters
type VectorSearchRequest struct {
    Query        string `json:"query" binding:"required"`
    RepositoryID string `json:"repositoryId,omitempty"`
    Language     string `json:"language,omitempty"`
    FileType     string `json:"fileType,omitempty"`
    Limit        int    `json:"limit,omitempty"`
    Offset       int    `json:"offset,omitempty"`
}

// HybridSearchRequest represents hybrid (text + vector) search request parameters
type HybridSearchRequest struct {
    Query        string  `json:"query" binding:"required"`
    RepositoryID string  `json:"repositoryId,omitempty"`
    Language     string  `json:"language,omitempty"`
    FileType     string  `json:"fileType,omitempty"`
    VectorWeight float64 `json:"vectorWeight,omitempty"` // 0.0 to 1.0, defaults to 0.7
    TextWeight   float64 `json:"textWeight,omitempty"`   // 0.0 to 1.0, defaults to 0.3
    Limit        int     `json:"limit,omitempty"`
    Offset       int     `json:"offset,omitempty"`
}

// SimilarityResult represents a vector similarity search result with distance scores
type SimilarityResult struct {
    CodeChunk
    Score      float32 `json:"score"`      // Cosine similarity score (0.0 to 1.0)
    Distance   float32 `json:"distance"`   // Cosine distance (1.0 - score)
    Relevance  string  `json:"relevance"`  // "high", "medium", "low" based on score
    Highlight  string  `json:"highlight,omitempty"`
}

// EmbeddingStatusResponse represents the status of embedding processing for a repository
type EmbeddingStatusResponse struct {
    RepositoryID         string        `json:"repositoryId"`
    Status               string        `json:"status"`               // pending, processing, completed, failed
    Progress             int           `json:"progress"`             // 0-100
    TotalChunks          *int          `json:"totalChunks,omitempty"`
    ProcessedChunks      *int          `json:"processedChunks,omitempty"`
    FailedChunks         *int          `json:"failedChunks,omitempty"`
    StartedAt            *time.Time    `json:"startedAt,omitempty"`
    CompletedAt          *time.Time    `json:"completedAt,omitempty"`
    EstimatedTimeRemaining *time.Duration `json:"estimatedTimeRemaining,omitempty"`
}

// CreateCodeChunkRequest represents the request payload for creating a code chunk
type CreateCodeChunkRequest struct {
    RepositoryID string        `json:"repositoryId" binding:"required"`
    FilePath     string        `json:"filePath" binding:"required"`
    Language     string        `json:"language" binding:"required"`
    StartLine    int           `json:"startLine" binding:"required"`
    EndLine      int           `json:"endLine" binding:"required"`
    Content      string        `json:"content" binding:"required"`
    Imports      []string      `json:"imports,omitempty"`
    Metadata     ChunkMetadata `json:"metadata"`
}

// NewCodeChunk creates a new CodeChunk with default values
func NewCodeChunk(repositoryID primitive.ObjectID, req CreateCodeChunkRequest) *CodeChunk {
    now := time.Now()
    
    // Extract file name from path
    fileName := extractFileName(req.FilePath)
    
    // Generate content hash
    contentHash := generateContentHash(req.Content)
    
    return &CodeChunk{
        RepositoryID: repositoryID,
        FilePath:     req.FilePath,
        FileName:     fileName,
        Language:     req.Language,
        StartLine:    req.StartLine,
        EndLine:      req.EndLine,
        Content:      req.Content,
        ContentHash:  contentHash,
        Imports:      req.Imports,
        Metadata:     req.Metadata,
        CreatedAt:    now,
        UpdatedAt:    now,
    }
}

// Update applies updates to the code chunk
func (c *CodeChunk) Update(content string, metadata ChunkMetadata, imports []string) {
    c.Content = content
    c.ContentHash = generateContentHash(content)
    c.Metadata = metadata
    c.Imports = imports
    c.UpdatedAt = time.Now()
}

// SetVectorID sets the vector database ID for this chunk
func (c *CodeChunk) SetVectorID(vectorID string) {
    c.VectorID = vectorID
    c.UpdatedAt = time.Now()
}

// IsIndexed returns true if the chunk has been indexed in the vector database
func (c *CodeChunk) IsIndexed() bool {
    return c.VectorID != ""
}

// GetLineCount returns the number of lines in this chunk
func (c *CodeChunk) GetLineCount() int {
    return c.EndLine - c.StartLine + 1
}

// GetFileExtension returns the file extension
func (c *CodeChunk) GetFileExtension() string {
    if len(c.FileName) == 0 {
        return ""
    }
    
    for i := len(c.FileName) - 1; i >= 0; i-- {
        if c.FileName[i] == '.' {
            return c.FileName[i+1:]
        }
    }
    return ""
}

// HasFunction checks if the chunk contains a specific function
func (c *CodeChunk) HasFunction(functionName string) bool {
    for _, fn := range c.Metadata.Functions {
        if fn == functionName {
            return true
        }
    }
    return false
}

// HasClass checks if the chunk contains a specific class
func (c *CodeChunk) HasClass(className string) bool {
    for _, cls := range c.Metadata.Classes {
        if cls == className {
            return true
        }
    }
    return false
}

// extractFileName extracts the filename from a file path
func extractFileName(filePath string) string {
    if len(filePath) == 0 {
        return ""
    }
    
    // Find the last occurrence of '/' or '\' 
    lastSlash := -1
    for i := len(filePath) - 1; i >= 0; i-- {
        if filePath[i] == '/' || filePath[i] == '\\' {
            lastSlash = i
            break
        }
    }
    
    if lastSlash == -1 {
        return filePath // No path separator found, return the whole string
    }
    
    return filePath[lastSlash+1:]
}

// generateContentHash generates a SHA256 hash of the content
func generateContentHash(content string) string {
    hash := sha256.Sum256([]byte(content))
    return fmt.Sprintf("%x", hash)
}

// ValidateSearchRequest validates a search request
func ValidateSearchRequest(req SearchRequest) error {
    if req.Query == "" {
        return fmt.Errorf("query is required")
    }
    
    if len(req.Query) > 500 {
        return fmt.Errorf("query too long (max 500 characters)")
    }
    
    if req.Limit < 0 {
        return fmt.Errorf("limit must be non-negative")
    }
    
    if req.Limit > 100 {
        return fmt.Errorf("limit too large (max 100)")
    }
    
    if req.Offset < 0 {
        return fmt.Errorf("offset must be non-negative")
    }
    
    return nil
}

// NormalizeLanguage normalizes language names to standard forms
func NormalizeLanguage(language string) string {
    switch language {
    case "js", "javascript":
        return "javascript"
    case "ts", "typescript":
        return "typescript"
    case "py", "python":
        return "python"
    case "go", "golang":
        return "go"
    case "java":
        return "java"
    case "cpp", "c++", "cxx":
        return "cpp"
    case "c":
        return "c"
    case "cs", "csharp", "c#":
        return "csharp"
    case "php":
        return "php"
    case "rb", "ruby":
        return "ruby"
    case "rs", "rust":
        return "rust"
    case "sh", "bash", "shell":
        return "shell"
    case "sql":
        return "sql"
    case "html":
        return "html"
    case "css":
        return "css"
    case "json":
        return "json"
    case "yaml", "yml":
        return "yaml"
    case "xml":
        return "xml"
    case "md", "markdown":
        return "markdown"
    default:
        return language
    }
}

// Validate validates a vector search request
func (req *VectorSearchRequest) Validate() error {
    return ValidateVectorSearchRequest(*req)
}

// Validate validates a hybrid search request
func (req *HybridSearchRequest) Validate() error {
    return ValidateHybridSearchRequest(*req)
}

// ValidateVectorSearchRequest validates a vector search request
func ValidateVectorSearchRequest(req VectorSearchRequest) error {
    if req.Query == "" {
        return fmt.Errorf("query is required")
    }
    
    if len(req.Query) > 500 {
        return fmt.Errorf("query too long (max 500 characters)")
    }
    
    if req.Limit < 0 {
        return fmt.Errorf("limit must be non-negative")
    }
    
    if req.Limit > 100 {
        return fmt.Errorf("limit too large (max 100)")
    }
    
    if req.Offset < 0 {
        return fmt.Errorf("offset must be non-negative")
    }
    
    return nil
}

// ValidateHybridSearchRequest validates a hybrid search request
func ValidateHybridSearchRequest(req HybridSearchRequest) error {
    if req.Query == "" {
        return fmt.Errorf("query is required")
    }
    
    if len(req.Query) > 500 {
        return fmt.Errorf("query too long (max 500 characters)")
    }
    
    if req.Limit < 0 {
        return fmt.Errorf("limit must be non-negative")
    }
    
    if req.Limit > 100 {
        return fmt.Errorf("limit too large (max 100)")
    }
    
    if req.Offset < 0 {
        return fmt.Errorf("offset must be non-negative")
    }
    
    // Validate weights
    if req.VectorWeight < 0 || req.VectorWeight > 1 {
        return fmt.Errorf("vectorWeight must be between 0 and 1")
    }
    
    if req.TextWeight < 0 || req.TextWeight > 1 {
        return fmt.Errorf("textWeight must be between 0 and 1")
    }
    
    // Weights should sum to approximately 1
    if req.VectorWeight > 0 && req.TextWeight > 0 {
        sum := req.VectorWeight + req.TextWeight
        if sum < 0.95 || sum > 1.05 {
            return fmt.Errorf("vectorWeight and textWeight should sum to 1.0 (got %.2f)", sum)
        }
    }
    
    return nil
}

// SetDefaultWeights sets default weights for hybrid search if not provided
func (req *HybridSearchRequest) SetDefaultWeights() {
    if req.VectorWeight == 0 && req.TextWeight == 0 {
        req.VectorWeight = 0.7
        req.TextWeight = 0.3
    }
}

// SetDefaultLimits sets default limits for search requests if not provided
func (req *VectorSearchRequest) SetDefaultLimits() {
    if req.Limit == 0 {
        req.Limit = 20
    }
}

func (req *HybridSearchRequest) SetDefaultLimits() {
    if req.Limit == 0 {
        req.Limit = 20
    }
}

// CalculateRelevance determines relevance level based on similarity score
func (sr *SimilarityResult) CalculateRelevance() {
    sr.Distance = 1.0 - sr.Score
    
    if sr.Score >= 0.85 {
        sr.Relevance = "high"
    } else if sr.Score >= 0.6 {
        sr.Relevance = "medium"
    } else {
        sr.Relevance = "low"
    }
}

// NewSimilarityResult creates a new SimilarityResult with calculated relevance
func NewSimilarityResult(chunk CodeChunk, score float32) SimilarityResult {
    result := SimilarityResult{
        CodeChunk: chunk,
        Score:     score,
    }
    result.CalculateRelevance()
    return result
}