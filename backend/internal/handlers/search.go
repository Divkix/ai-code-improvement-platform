// ABOUTME: HTTP handlers for search endpoints with validation and error handling
// ABOUTME: Provides REST API for code search functionality with proper response formatting

package handlers

import (
    "net/http"
    "strconv"
    "strings"
    
    "github.com/gin-gonic/gin"
    
    "github-analyzer/internal/models"
    "github-analyzer/internal/services"
)

// SearchHandler handles search-related HTTP requests
type SearchHandler struct {
    searchService *services.SearchService
}

// NewSearchHandler creates a new search handler
func NewSearchHandler(searchService *services.SearchService) *SearchHandler {
    return &SearchHandler{
        searchService: searchService,
    }
}

// GlobalSearch handles POST /api/search - search across all accessible repositories
func (h *SearchHandler) GlobalSearch(c *gin.Context) {
    var req models.SearchRequest
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "invalid_request",
            "message": "Invalid search request format: " + err.Error(),
        })
        return
    }
    
    // Parse query parameters (they can override JSON body)
    h.parseQueryParameters(c, &req)
    
    // Perform search
    response, err := h.searchService.SearchCodeChunks(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "search_failed",
            "message": "Search operation failed",
        })
        return
    }
    
    c.JSON(http.StatusOK, response)
}

// RepositorySearch handles POST /api/repositories/{id}/search - search within specific repository
func (h *SearchHandler) RepositorySearch(c *gin.Context) {
    repositoryID := c.Param("id")
    if repositoryID == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "missing_repository_id",
            "message": "Repository ID is required",
        })
        return
    }
    
    var req models.SearchRequest
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "invalid_request", 
            "message": "Invalid search request format: " + err.Error(),
        })
        return
    }
    
    // Set repository ID from URL parameter
    req.RepositoryID = repositoryID
    
    // Parse query parameters (they can override JSON body)
    h.parseQueryParameters(c, &req)
    
    // Perform repository-specific search
    response, err := h.searchService.SearchByRepository(c.Request.Context(), repositoryID, req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "search_failed",
            "message": "Repository search operation failed",
        })
        return
    }
    
    c.JSON(http.StatusOK, response)
}

// GetSearchSuggestions handles GET /api/search/suggestions - get search suggestions
func (h *SearchHandler) GetSearchSuggestions(c *gin.Context) {
    query := c.Query("q")
    repositoryID := c.Query("repositoryId")
    
    if query == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "missing_query",
            "message": "Query parameter 'q' is required",
        })
        return
    }
    
    if len(query) < 2 {
        c.JSON(http.StatusOK, gin.H{
            "suggestions": []string{},
        })
        return
    }
    
    // For now, return simple suggestions based on common patterns
    // This could be enhanced with actual database queries for function/class names
    suggestions := h.generateSuggestions(query, repositoryID)
    
    c.JSON(http.StatusOK, gin.H{
        "suggestions": suggestions,
        "query":       query,
    })
}

// GetLanguages handles GET /api/search/languages - get available programming languages
func (h *SearchHandler) GetLanguages(c *gin.Context) {
    repositoryID := c.Query("repositoryId")
    
    languages, err := h.searchService.GetLanguages(c.Request.Context(), repositoryID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "failed_to_get_languages",
            "message": "Failed to retrieve available languages",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "languages": languages,
    })
}

// GetSearchStats handles GET /api/search/stats - get search statistics
func (h *SearchHandler) GetSearchStats(c *gin.Context) {
    repositoryID := c.Query("repositoryId")
    
    stats, err := h.searchService.GetStats(c.Request.Context(), repositoryID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "failed_to_get_stats",
            "message": "Failed to retrieve search statistics",
        })
        return
    }
    
    c.JSON(http.StatusOK, stats)
}

// GetRecentChunks handles GET /api/search/recent - get recently added code chunks
func (h *SearchHandler) GetRecentChunks(c *gin.Context) {
    repositoryID := c.Query("repositoryId")
    
    // Parse limit
    limitStr := c.DefaultQuery("limit", "20")
    limit, err := strconv.Atoi(limitStr)
    if err != nil || limit <= 0 || limit > 100 {
        limit = 20
    }
    
    chunks, err := h.searchService.GetRecentChunks(c.Request.Context(), repositoryID, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "failed_to_get_recent_chunks",
            "message": "Failed to retrieve recent code chunks",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "chunks": chunks,
        "total":  len(chunks),
    })
}

// parseQueryParameters extracts and sets search parameters from URL query params
func (h *SearchHandler) parseQueryParameters(c *gin.Context, req *models.SearchRequest) {
    // Override with query parameters if provided
    if limit := c.Query("limit"); limit != "" {
        if l, err := strconv.Atoi(limit); err == nil && l > 0 {
            req.Limit = l
        }
    }
    
    if offset := c.Query("offset"); offset != "" {
        if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
            req.Offset = o
        }
    }
    
    if req.Language == "" {
        req.Language = c.Query("language")
    }
    
    if req.FileType == "" {
        req.FileType = c.Query("fileType")
    }
    
    if req.RepositoryID == "" {
        req.RepositoryID = c.Query("repositoryId")
    }
}

// generateSuggestions creates search suggestions based on query
func (h *SearchHandler) generateSuggestions(query, repositoryID string) []string {
    suggestions := []string{}
    
    // Common programming patterns and keywords
    commonPatterns := []string{
        "function", "class", "method", "variable", "const", "let", "var",
        "interface", "type", "struct", "enum", "import", "export",
        "async", "await", "promise", "error", "exception", "handler",
        "service", "controller", "model", "view", "component",
        "test", "spec", "mock", "fixture", "helper", "util",
        "config", "setting", "parameter", "option", "flag",
        "create", "update", "delete", "get", "set", "find", "search",
        "validate", "parse", "format", "convert", "transform",
        "connect", "disconnect", "login", "logout", "auth", "user",
        "database", "query", "schema", "migration", "index",
        "api", "endpoint", "route", "middleware", "request", "response",
    }
    
    // Filter patterns that start with or contain the query
    for _, pattern := range commonPatterns {
        if len(suggestions) >= 10 {
            break
        }
        
        if containsIgnoreCase(pattern, query) {
            suggestions = append(suggestions, pattern)
        }
    }
    
    // Programming language specific suggestions
    if repositoryID != "" {
        // Add language-specific suggestions based on the repository
        // This could be enhanced with actual database queries
        langSuggestions := []string{
            query + " function",
            query + " class",
            query + " method",
            query + " variable",
            query + " type",
        }
        
        for _, suggestion := range langSuggestions {
            if len(suggestions) >= 10 {
                break
            }
            suggestions = append(suggestions, suggestion)
        }
    }
    
    return suggestions
}

// containsIgnoreCase checks if s contains substr (case insensitive)
func containsIgnoreCase(s, substr string) bool {
    s = strings.ToLower(s)
    substr = strings.ToLower(substr)
    return strings.Contains(s, substr)
}

// QuickSearch handles GET /api/search/quick - lightweight search for autocomplete
func (h *SearchHandler) QuickSearch(c *gin.Context) {
    query := c.Query("q")
    if query == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "error":   "missing_query",
            "message": "Query parameter 'q' is required",
        })
        return
    }
    
    if len(query) < 2 {
        c.JSON(http.StatusOK, gin.H{
            "results": []models.SearchResult{},
            "total":   0,
        })
        return
    }
    
    // Quick search with limited results
    req := models.SearchRequest{
        Query:        query,
        Limit:        5, // Limit to 5 results for quick search
        Offset:       0,
        RepositoryID: c.Query("repositoryId"),
        Language:     c.Query("language"),
        FileType:     c.Query("fileType"),
    }
    
    response, err := h.searchService.SearchCodeChunks(c.Request.Context(), req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error":   "search_failed",
            "message": "Quick search operation failed",
        })
        return
    }
    
    // Return simplified results for quick search
    quickResults := make([]gin.H, len(response.Results))
    for i, result := range response.Results {
        quickResults[i] = gin.H{
            "id":        result.ID,
            "filePath":  result.FilePath,
            "fileName":  result.FileName,
            "language":  result.Language,
            "highlight": result.Highlight,
            "score":     result.Score,
        }
    }
    
    c.JSON(http.StatusOK, gin.H{
        "results": quickResults,
        "total":   response.Total,
        "query":   query,
    })
}