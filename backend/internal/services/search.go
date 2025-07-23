// ABOUTME: Search service implementing MongoDB text search for code chunks
// ABOUTME: Provides relevance scoring, filtering, and pagination capabilities

package services

import (
    "context"
    "fmt"
    "strings"
    
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    
    "github-analyzer/internal/models"
)

// SearchService handles code search operations
type SearchService struct {
    db           *mongo.Database
    codeChunks   *mongo.Collection
}

// NewSearchService creates a new search service
func NewSearchService(db *mongo.Database) *SearchService {
    return &SearchService{
        db:         db,
        codeChunks: db.Collection("code_chunks"),
    }
}

// CreateTextIndex creates the compound text search index
func (s *SearchService) CreateTextIndex(ctx context.Context) error {
    indexModel := mongo.IndexModel{
        Keys: bson.D{
            {Key: "content", Value: "text"},
            {Key: "metadata.functions", Value: "text"},
            {Key: "metadata.classes", Value: "text"},
            {Key: "filePath", Value: "text"},
            {Key: "fileName", Value: "text"},
            {Key: "imports", Value: "text"},
        },
        Options: options.Index().
            SetWeights(bson.D{
                {Key: "content", Value: 10},
                {Key: "metadata.functions", Value: 8},
                {Key: "metadata.classes", Value: 8},
                {Key: "filePath", Value: 5},
                {Key: "fileName", Value: 5},
                {Key: "imports", Value: 2},
            }).
            SetName("CodeSearchIndex"),
    }
    
    _, err := s.codeChunks.Indexes().CreateOne(ctx, indexModel)
    return err
}

// EnsureIndexes creates all required indexes for the search service
func (s *SearchService) EnsureIndexes(ctx context.Context) error {
    // Create text search index
    if err := s.CreateTextIndex(ctx); err != nil {
        return fmt.Errorf("failed to create text search index: %w", err)
    }
    
    // Create additional indexes for performance
    indexes := []mongo.IndexModel{
        {
            Keys: bson.D{{Key: "repositoryId", Value: 1}},
            Options: options.Index().SetName("repositoryId_1"),
        },
        {
            Keys: bson.D{{Key: "language", Value: 1}},
            Options: options.Index().SetName("language_1"),
        },
        {
            Keys: bson.D{{Key: "fileName", Value: 1}},
            Options: options.Index().SetName("fileName_1"),
        },
        {
            Keys: bson.D{{Key: "contentHash", Value: 1}},
            Options: options.Index().SetName("contentHash_1").SetUnique(true),
        },
        {
            Keys: bson.D{{Key: "createdAt", Value: 1}},
            Options: options.Index().SetName("createdAt_1"),
        },
    }
    
    _, err := s.codeChunks.Indexes().CreateMany(ctx, indexes)
    if err != nil {
        return fmt.Errorf("failed to create additional indexes: %w", err)
    }
    
    return nil
}

// SearchCodeChunks performs text search on code chunks
func (s *SearchService) SearchCodeChunks(ctx context.Context, req models.SearchRequest) (*models.SearchResponse, error) {
    // Validate request
    if err := models.ValidateSearchRequest(req); err != nil {
        return nil, fmt.Errorf("invalid search request: %w", err)
    }
    
    // Set defaults
    if req.Limit <= 0 || req.Limit > 100 {
        req.Limit = 20
    }
    if req.Offset < 0 {
        req.Offset = 0
    }
    
    // Build match stage
    matchStage := bson.D{
        {Key: "$text", Value: bson.D{{Key: "$search", Value: req.Query}}},
    }
    
    // Add repository filter if specified
    if req.RepositoryID != "" {
        repoID, err := primitive.ObjectIDFromHex(req.RepositoryID)
        if err != nil {
            return nil, fmt.Errorf("invalid repository ID: %w", err)
        }
        matchStage = append(matchStage, bson.E{Key: "repositoryId", Value: repoID})
    }
    
    // Add language filter if specified
    if req.Language != "" {
        normalizedLang := models.NormalizeLanguage(req.Language)
        matchStage = append(matchStage, bson.E{Key: "language", Value: normalizedLang})
    }
    
    // Add file type filter if specified
    if req.FileType != "" {
        matchStage = append(matchStage, bson.E{Key: "fileName", Value: primitive.Regex{
            Pattern: fmt.Sprintf("\\.%s$", req.FileType),
            Options: "i",
        }})
    }
    
    // Build aggregation pipeline
    pipeline := mongo.Pipeline{
        bson.D{{Key: "$match", Value: matchStage}},
        bson.D{{Key: "$project", Value: bson.D{
            {Key: "repositoryId", Value: 1},
            {Key: "filePath", Value: 1},
            {Key: "fileName", Value: 1},
            {Key: "language", Value: 1},
            {Key: "startLine", Value: 1},
            {Key: "endLine", Value: 1},
            {Key: "content", Value: 1},
            {Key: "contentHash", Value: 1},
            {Key: "imports", Value: 1},
            {Key: "metadata", Value: 1},
            {Key: "vectorId", Value: 1},
            {Key: "createdAt", Value: 1},
            {Key: "updatedAt", Value: 1},
            {Key: "score", Value: bson.D{{Key: "$meta", Value: "textScore"}}},
        }}},
        bson.D{{Key: "$sort", Value: bson.D{{Key: "score", Value: bson.D{{Key: "$meta", Value: "textScore"}}}}}},
        bson.D{{Key: "$skip", Value: req.Offset}},
        bson.D{{Key: "$limit", Value: req.Limit + 1}}, // +1 to check if there are more results
    }
    
    // Execute search
    cursor, err := s.codeChunks.Aggregate(ctx, pipeline)
    if err != nil {
        return nil, fmt.Errorf("search query failed: %w", err)
    }
    defer func() { _ = cursor.Close(ctx) }()
    
    // Parse results
    var searchResults []models.SearchResult
    for cursor.Next(ctx) {
        var result models.SearchResult
        if err := cursor.Decode(&result); err != nil {
            continue // Skip malformed results
        }
        
        // Generate highlight snippet
        result.Highlight = s.generateHighlight(result.Content, req.Query)
        searchResults = append(searchResults, result)
    }
    
    // Check for pagination
    hasMore := len(searchResults) > req.Limit
    if hasMore {
        searchResults = searchResults[:req.Limit] // Remove the extra result
    }
    
    // Get total count for the query (without limit/offset)
    total, err := s.getSearchCount(ctx, matchStage)
    if err != nil {
        // Log error but don't fail the request
        total = int64(len(searchResults))
    }
    
    return &models.SearchResponse{
        Results: searchResults,
        Total:   total,
        HasMore: hasMore,
        Query:   req.Query,
    }, nil
}

// getSearchCount gets the total count of results for a search query
func (s *SearchService) getSearchCount(ctx context.Context, matchStage bson.D) (int64, error) {
    countPipeline := mongo.Pipeline{
        bson.D{{Key: "$match", Value: matchStage}},
        bson.D{{Key: "$count", Value: "total"}},
    }
    
    cursor, err := s.codeChunks.Aggregate(ctx, countPipeline)
    if err != nil {
        return 0, fmt.Errorf("count query failed: %w", err)
    }
    defer func() { _ = cursor.Close(ctx) }()
    
    if cursor.Next(ctx) {
        var countResult struct {
            Total int64 `bson:"total"`
        }
        if err := cursor.Decode(&countResult); err == nil {
            return countResult.Total, nil
        }
    }
    
    return 0, nil
}

// generateHighlight creates a highlighted snippet of the content
func (s *SearchService) generateHighlight(content, query string) string {
    words := strings.Fields(strings.ToLower(query))
    if len(words) == 0 {
        return s.truncateContent(content, 200)
    }
    
    lowerContent := strings.ToLower(content)
    
    // Find the best matching position
    bestPos := -1
    maxMatches := 0
    
    for i := 0; i < len(content)-100; i += 50 {
        matches := 0
        section := lowerContent[i:min(i+200, len(lowerContent))]
        
        for _, word := range words {
            if strings.Contains(section, word) {
                matches++
            }
        }
        
        if matches > maxMatches {
            maxMatches = matches
            bestPos = i
        }
    }
    
    if bestPos == -1 {
        return s.truncateContent(content, 200)
    }
    
    start := max(0, bestPos)
    end := min(len(content), start+200)
    
    result := content[start:end]
    if end < len(content) {
        result += "..."
    }
    if start > 0 {
        result = "..." + result
    }
    
    return result
}

// truncateContent truncates content to specified length
func (s *SearchService) truncateContent(content string, maxLen int) string {
    if len(content) <= maxLen {
        return content
    }
    return content[:maxLen] + "..."
}

// SearchByRepository searches code chunks within a specific repository
func (s *SearchService) SearchByRepository(ctx context.Context, repositoryID string, req models.SearchRequest) (*models.SearchResponse, error) {
    req.RepositoryID = repositoryID
    return s.SearchCodeChunks(ctx, req)
}

// GetLanguages returns all unique languages in the code chunks
func (s *SearchService) GetLanguages(ctx context.Context, repositoryID string) ([]string, error) {
    matchStage := bson.D{}
    if repositoryID != "" {
        repoID, err := primitive.ObjectIDFromHex(repositoryID)
        if err != nil {
            return nil, fmt.Errorf("invalid repository ID: %w", err)
        }
        matchStage = bson.D{{Key: "repositoryId", Value: repoID}}
    }
    
    pipeline := mongo.Pipeline{
        bson.D{{Key: "$match", Value: matchStage}},
        bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$language"}}}},
        bson.D{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
    }
    
    cursor, err := s.codeChunks.Aggregate(ctx, pipeline)
    if err != nil {
        return nil, fmt.Errorf("failed to get languages: %w", err)
    }
    defer func() { _ = cursor.Close(ctx) }()
    
    var languages []string
    for cursor.Next(ctx) {
        var result struct {
            ID string `bson:"_id"`
        }
        if err := cursor.Decode(&result); err != nil {
            continue
        }
        if result.ID != "" {
            languages = append(languages, result.ID)
        }
    }
    
    return languages, nil
}

// GetRecentChunks returns the most recently created code chunks
func (s *SearchService) GetRecentChunks(ctx context.Context, repositoryID string, limit int) ([]models.CodeChunk, error) {
    if limit <= 0 || limit > 100 {
        limit = 20
    }
    
    filter := bson.D{}
    if repositoryID != "" {
        repoID, err := primitive.ObjectIDFromHex(repositoryID)
        if err != nil {
            return nil, fmt.Errorf("invalid repository ID: %w", err)
        }
        filter = bson.D{{Key: "repositoryId", Value: repoID}}
    }
    
    opts := options.Find().
        SetSort(bson.D{{Key: "createdAt", Value: -1}}).
        SetLimit(int64(limit))
    
    cursor, err := s.codeChunks.Find(ctx, filter, opts)
    if err != nil {
        return nil, fmt.Errorf("failed to get recent chunks: %w", err)
    }
    defer func() { _ = cursor.Close(ctx) }()
    
    var chunks []models.CodeChunk
    if err := cursor.All(ctx, &chunks); err != nil {
        return nil, fmt.Errorf("failed to decode chunks: %w", err)
    }
    
    return chunks, nil
}

// GetStats returns search statistics
func (s *SearchService) GetStats(ctx context.Context, repositoryID string) (*SearchStats, error) {
    matchStage := bson.D{}
    if repositoryID != "" {
        repoID, err := primitive.ObjectIDFromHex(repositoryID)
        if err != nil {
            return nil, fmt.Errorf("invalid repository ID: %w", err)
        }
        matchStage = bson.D{{Key: "repositoryId", Value: repoID}}
    }
    
    pipeline := mongo.Pipeline{
        bson.D{{Key: "$match", Value: matchStage}},
        bson.D{{Key: "$group", Value: bson.D{
            {Key: "_id", Value: nil},
            {Key: "totalChunks", Value: bson.D{{Key: "$sum", Value: 1}}},
            {Key: "totalLines", Value: bson.D{{Key: "$sum", Value: bson.D{{Key: "$subtract", Value: bson.A{"$endLine", "$startLine"}}}}}},
            {Key: "avgComplexity", Value: bson.D{{Key: "$avg", Value: "$metadata.complexity"}}},
            {Key: "languages", Value: bson.D{{Key: "$addToSet", Value: "$language"}}},
        }}},
    }
    
    cursor, err := s.codeChunks.Aggregate(ctx, pipeline)
    if err != nil {
        return nil, fmt.Errorf("failed to get stats: %w", err)
    }
    defer func() { _ = cursor.Close(ctx) }()
    
    if cursor.Next(ctx) {
        var result struct {
            TotalChunks   int      `bson:"totalChunks"`
            TotalLines    int      `bson:"totalLines"`
            AvgComplexity float64  `bson:"avgComplexity"`
            Languages     []string `bson:"languages"`
        }
        
        if err := cursor.Decode(&result); err != nil {
            return nil, fmt.Errorf("failed to decode stats: %w", err)
        }
        
        return &SearchStats{
            TotalChunks:   result.TotalChunks,
            TotalLines:    result.TotalLines,
            AvgComplexity: result.AvgComplexity,
            Languages:     result.Languages,
        }, nil
    }
    
    return &SearchStats{}, nil
}

// SearchStats represents search statistics
type SearchStats struct {
    TotalChunks   int      `json:"totalChunks"`
    TotalLines    int      `json:"totalLines"`
    AvgComplexity float64  `json:"avgComplexity"`
    Languages     []string `json:"languages"`
}

// Helper functions
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}