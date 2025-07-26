// ABOUTME: Comprehensive unit tests for CodeChunk model and related validation functions
// ABOUTME: Tests all helper methods, validation logic, and business rules

package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestNewCodeChunk(t *testing.T) {
	repositoryID := primitive.NewObjectID()
	req := CreateCodeChunkRequest{
		RepositoryID: repositoryID.Hex(),
		FilePath:     "/path/to/file.go",
		Language:     "go",
		StartLine:    1,
		EndLine:      10,
		Content:      "package main\n\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}",
		Imports:      []string{"fmt"},
		Metadata: ChunkMetadata{
			Functions: []string{"main"},
		},
	}

	chunk := NewCodeChunk(repositoryID, req)

	assert.Equal(t, repositoryID, chunk.RepositoryID)
	assert.Equal(t, "/path/to/file.go", chunk.FilePath)
	assert.Equal(t, "file.go", chunk.FileName)
	assert.Equal(t, "go", chunk.Language)
	assert.Equal(t, 1, chunk.StartLine)
	assert.Equal(t, 10, chunk.EndLine)
	assert.Equal(t, req.Content, chunk.Content)
	assert.NotEmpty(t, chunk.ContentHash)
	assert.Equal(t, req.Imports, chunk.Imports)
	assert.Equal(t, req.Metadata, chunk.Metadata)
	assert.NotZero(t, chunk.CreatedAt)
	assert.NotZero(t, chunk.UpdatedAt)
}

func TestCodeChunk_Update(t *testing.T) {
	chunk := &CodeChunk{
		Content:     "old content",
		ContentHash: "old_hash",
		UpdatedAt:   time.Now().Add(-time.Hour),
	}

	newContent := "new content"
	newMetadata := ChunkMetadata{
		Functions: []string{"newFunction"},
	}
	newImports := []string{"newImport"}

	oldUpdatedAt := chunk.UpdatedAt
	chunk.Update(newContent, newMetadata, newImports)

	assert.Equal(t, newContent, chunk.Content)
	assert.NotEqual(t, "old_hash", chunk.ContentHash)
	assert.Equal(t, newMetadata, chunk.Metadata)
	assert.Equal(t, newImports, chunk.Imports)
	assert.True(t, chunk.UpdatedAt.After(oldUpdatedAt))
}

func TestCodeChunk_SetVectorID(t *testing.T) {
	chunk := &CodeChunk{
		VectorID:  "",
		UpdatedAt: time.Now().Add(-time.Hour),
	}

	oldUpdatedAt := chunk.UpdatedAt
	vectorID := "vector_12345"
	chunk.SetVectorID(vectorID)

	assert.Equal(t, vectorID, chunk.VectorID)
	assert.True(t, chunk.UpdatedAt.After(oldUpdatedAt))
}

func TestCodeChunk_IsIndexed(t *testing.T) {
	tests := []struct {
		name     string
		vectorID string
		expected bool
	}{
		{
			name:     "indexed chunk",
			vectorID: "vector_12345",
			expected: true,
		},
		{
			name:     "not indexed chunk",
			vectorID: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := &CodeChunk{VectorID: tt.vectorID}
			assert.Equal(t, tt.expected, chunk.IsIndexed())
		})
	}
}

func TestCodeChunk_GetLineCount(t *testing.T) {
	tests := []struct {
		name      string
		startLine int
		endLine   int
		expected  int
	}{
		{
			name:      "single line",
			startLine: 1,
			endLine:   1,
			expected:  1,
		},
		{
			name:      "multiple lines",
			startLine: 1,
			endLine:   10,
			expected:  10,
		},
		{
			name:      "range with gap",
			startLine: 5,
			endLine:   15,
			expected:  11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := &CodeChunk{
				StartLine: tt.startLine,
				EndLine:   tt.endLine,
			}
			assert.Equal(t, tt.expected, chunk.GetLineCount())
		})
	}
}

func TestCodeChunk_GetFileExtension(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		expected string
	}{
		{
			name:     "go file",
			fileName: "main.go",
			expected: "go",
		},
		{
			name:     "javascript file",
			fileName: "app.js",
			expected: "js",
		},
		{
			name:     "typescript file",
			fileName: "component.tsx",
			expected: "tsx",
		},
		{
			name:     "no extension",
			fileName: "Makefile",
			expected: "",
		},
		{
			name:     "empty filename",
			fileName: "",
			expected: "",
		},
		{
			name:     "multiple dots",
			fileName: "config.test.js",
			expected: "js",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chunk := &CodeChunk{FileName: tt.fileName}
			assert.Equal(t, tt.expected, chunk.GetFileExtension())
		})
	}
}

func TestCodeChunk_HasFunction(t *testing.T) {
	chunk := &CodeChunk{
		Metadata: ChunkMetadata{
			Functions: []string{"main", "helper", "process"},
		},
	}

	tests := []struct {
		name         string
		functionName string
		expected     bool
	}{
		{
			name:         "existing function",
			functionName: "main",
			expected:     true,
		},
		{
			name:         "another existing function",
			functionName: "helper",
			expected:     true,
		},
		{
			name:         "non-existing function",
			functionName: "nonexistent",
			expected:     false,
		},
		{
			name:         "empty function name",
			functionName: "",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, chunk.HasFunction(tt.functionName))
		})
	}
}

func TestCodeChunk_HasClass(t *testing.T) {
	chunk := &CodeChunk{
		Metadata: ChunkMetadata{
			Classes: []string{"User", "Repository", "CodeProcessor"},
		},
	}

	tests := []struct {
		name      string
		className string
		expected  bool
	}{
		{
			name:      "existing class",
			className: "User",
			expected:  true,
		},
		{
			name:      "another existing class",
			className: "Repository",
			expected:  true,
		},
		{
			name:      "non-existing class",
			className: "NonExistent",
			expected:  false,
		},
		{
			name:      "empty class name",
			className: "",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, chunk.HasClass(tt.className))
		})
	}
}

func TestExtractFileName(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{
			name:     "unix path",
			filePath: "/path/to/file.go",
			expected: "file.go",
		},
		{
			name:     "windows path",
			filePath: "C:\\path\\to\\file.js",
			expected: "file.js",
		},
		{
			name:     "no path separator",
			filePath: "filename.txt",
			expected: "filename.txt",
		},
		{
			name:     "empty path",
			filePath: "",
			expected: "",
		},
		{
			name:     "root file",
			filePath: "/file.go",
			expected: "file.go",
		},
		{
			name:     "mixed separators",
			filePath: "/path\\to/file.py",
			expected: "file.py",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFileName(tt.filePath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateContentHash(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "simple content",
			content: "hello world",
		},
		{
			name:    "code content",
			content: "package main\n\nfunc main() {\n    fmt.Println(\"Hello, World!\")\n}",
		},
		{
			name:    "empty content",
			content: "",
		},
		{
			name:    "special characters",
			content: "!@#$%^&*()_+-={}[]|\\:;\"'<>?,./",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := generateContentHash(tt.content)
			hash2 := generateContentHash(tt.content)
			
			// Same content should produce same hash
			assert.Equal(t, hash1, hash2)
			
			// Hash should not be empty
			assert.NotEmpty(t, hash1)
			
			// Hash should be 64 characters (SHA256 hex)
			assert.Len(t, hash1, 64)
		})
	}
}

func TestValidateSearchRequest(t *testing.T) {
	tests := []struct {
		name        string
		request     SearchRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid request",
			request: SearchRequest{
				Query: "test query",
				Limit: 10,
			},
			expectError: false,
		},
		{
			name: "empty query",
			request: SearchRequest{
				Query: "",
			},
			expectError: true,
			errorMsg:    "query is required",
		},
		{
			name: "query too long",
			request: SearchRequest{
				Query: string(make([]byte, 501)), // 501 characters
			},
			expectError: true,
			errorMsg:    "query too long",
		},
		{
			name: "negative limit",
			request: SearchRequest{
				Query: "test",
				Limit: -1,
			},
			expectError: true,
			errorMsg:    "limit must be non-negative",
		},
		{
			name: "limit too large",
			request: SearchRequest{
				Query: "test",
				Limit: 101,
			},
			expectError: true,
			errorMsg:    "limit too large",
		},
		{
			name: "negative offset",
			request: SearchRequest{
				Query:  "test",
				Offset: -1,
			},
			expectError: true,
			errorMsg:    "offset must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSearchRequest(tt.request)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNormalizeLanguage(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"js", "javascript"},
		{"javascript", "javascript"},
		{"ts", "typescript"},
		{"typescript", "typescript"},
		{"py", "python"},
		{"python", "python"},
		{"go", "go"},
		{"golang", "go"},
		{"cpp", "cpp"},
		{"c++", "cpp"},
		{"cxx", "cpp"},
		{"c", "c"},
		{"cs", "csharp"},
		{"csharp", "csharp"},
		{"c#", "csharp"},
		{"rb", "ruby"},
		{"ruby", "ruby"},
		{"rs", "rust"},
		{"rust", "rust"},
		{"sh", "shell"},
		{"bash", "shell"},
		{"shell", "shell"},
		{"yaml", "yaml"},
		{"yml", "yaml"},
		{"md", "markdown"},
		{"markdown", "markdown"},
		{"unknown", "unknown"}, // Should return as-is
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeLanguage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHybridSearchRequest_SetDefaultWeights(t *testing.T) {
	tests := []struct {
		name           string
		request        HybridSearchRequest
		expectedVector float64
		expectedText   float64
	}{
		{
			name:           "no weights set",
			request:        HybridSearchRequest{},
			expectedVector: 0.7,
			expectedText:   0.3,
		},
		{
			name: "weights already set",
			request: HybridSearchRequest{
				VectorWeight: 0.6,
				TextWeight:   0.4,
			},
			expectedVector: 0.6,
			expectedText:   0.4,
		},
		{
			name: "only vector weight set",
			request: HybridSearchRequest{
				VectorWeight: 0.8,
			},
			expectedVector: 0.8,
			expectedText:   0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.request
			req.SetDefaultWeights()
			
			assert.Equal(t, tt.expectedVector, req.VectorWeight)
			assert.Equal(t, tt.expectedText, req.TextWeight)
		})
	}
}

func TestSimilarityResult_CalculateRelevance(t *testing.T) {
	tests := []struct {
		name             string
		score            float32
		expectedRelevance string
		expectedDistance float32
	}{
		{
			name:             "high relevance",
			score:            0.9,
			expectedRelevance: "high",
			expectedDistance: 0.1,
		},
		{
			name:             "medium relevance",
			score:            0.7,
			expectedRelevance: "medium",
			expectedDistance: 0.3,
		},
		{
			name:             "low relevance",
			score:            0.4,
			expectedRelevance: "low",
			expectedDistance: 0.6,
		},
		{
			name:             "edge case high",
			score:            0.85,
			expectedRelevance: "high",
			expectedDistance: 0.15,
		},
		{
			name:             "edge case medium",
			score:            0.6,
			expectedRelevance: "medium",
			expectedDistance: 0.4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &SimilarityResult{Score: tt.score}
			result.CalculateRelevance()
			
			assert.Equal(t, tt.expectedRelevance, result.Relevance)
			assert.InDelta(t, tt.expectedDistance, result.Distance, 0.001)
		})
	}
}

func TestNewSimilarityResult(t *testing.T) {
	chunk := CodeChunk{
		ID:       primitive.NewObjectID(),
		FilePath: "/test/file.go",
		Content:  "test content",
	}
	score := float32(0.8)

	result := NewSimilarityResult(chunk, score)

	assert.Equal(t, chunk, result.CodeChunk)
	assert.Equal(t, score, result.Score)
	assert.Equal(t, "medium", result.Relevance)
	assert.InDelta(t, 0.2, result.Distance, 0.001)
}

func TestValidateHybridSearchRequest(t *testing.T) {
	tests := []struct {
		name        string
		request     HybridSearchRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid request",
			request: HybridSearchRequest{
				Query:        "test",
				VectorWeight: 0.7,
				TextWeight:   0.3,
			},
			expectError: false,
		},
		{
			name: "weights don't sum to 1",
			request: HybridSearchRequest{
				Query:        "test",
				VectorWeight: 0.3,
				TextWeight:   0.3,
			},
			expectError: true,
			errorMsg:    "should sum to 1.0",
		},
		{
			name: "vector weight out of range",
			request: HybridSearchRequest{
				Query:        "test",
				VectorWeight: 1.5,
				TextWeight:   0.3,
			},
			expectError: true,
			errorMsg:    "vectorWeight must be between 0 and 1",
		},
		{
			name: "text weight out of range",
			request: HybridSearchRequest{
				Query:      "test",
				TextWeight: -0.1,
			},
			expectError: true,
			errorMsg:    "textWeight must be between 0 and 1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHybridSearchRequest(tt.request)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}