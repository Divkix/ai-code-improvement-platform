// ABOUTME: Code processing service for chunking files and extracting metadata
// ABOUTME: Implements 150-line chunks with overlap and language-aware processing

package services

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github-analyzer/internal/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// CodeProcessor handles code processing and chunking operations
type CodeProcessor struct {
	chunkSize    int // Lines per chunk
	overlapSize  int // Lines to overlap between chunks
}

// NewCodeProcessor creates a new code processor with default settings
func NewCodeProcessor() *CodeProcessor {
	return &CodeProcessor{
		chunkSize:   150, // 150 lines per chunk
		overlapSize: 50,  // 50 lines overlap
	}
}

// ProcessAndChunkFiles processes repository files and creates code chunks
func (cp *CodeProcessor) ProcessAndChunkFiles(files []*models.RepositoryFile, repositoryID primitive.ObjectID) ([]*models.CodeChunk, error) {
	var allChunks []*models.CodeChunk
	
	log.Printf("Processing %d files for chunking", len(files))
	
	for _, file := range files {
		if !file.IsValidForProcessing() {
			continue
		}
		
		chunks, err := cp.ChunkFile(file, repositoryID)
		if err != nil {
			log.Printf("Failed to chunk file %s: %v", file.Path, err)
			continue
		}
		
		allChunks = append(allChunks, chunks...)
	}
	
	log.Printf("Created %d chunks from %d files", len(allChunks), len(files))
	return allChunks, nil
}

// ChunkFile splits a file into overlapping chunks
func (cp *CodeProcessor) ChunkFile(file *models.RepositoryFile, repositoryID primitive.ObjectID) ([]*models.CodeChunk, error) {
	if file.Content == "" {
		return nil, fmt.Errorf("file content is empty")
	}
	
	lines := strings.Split(file.Content, "\n")
	if len(lines) < 5 {
		return nil, fmt.Errorf("file too short for chunking")
	}
	
	var chunks []*models.CodeChunk
	chunkIndex := 0
	
	// Create overlapping chunks
	for start := 0; start < len(lines); start += (cp.chunkSize - cp.overlapSize) {
		end := start + cp.chunkSize
		if end > len(lines) {
			end = len(lines)
		}
		
		// Skip if chunk is too small (less than 10 lines)
		if end-start < 10 {
			break
		}
		
		chunkLines := lines[start:end]
		chunkContent := strings.Join(chunkLines, "\n")
		
		// Extract metadata from chunk
		metadata := cp.ExtractMetadata(chunkContent, file.Language)
		
		// Create chunk request
		chunkReq := models.CreateCodeChunkRequest{
			RepositoryID: repositoryID.Hex(),
			FilePath:     file.Path,
			Language:     file.Language,
			StartLine:    start + 1, // 1-indexed
			EndLine:      end,
			Content:      chunkContent,
			Imports:      cp.ExtractImports(chunkContent, file.Language),
			Metadata:     metadata,
		}
		
		// Create code chunk
		chunk := models.NewCodeChunk(repositoryID, chunkReq)
		chunks = append(chunks, chunk)
		chunkIndex++
	}
	
	log.Printf("Created %d chunks for file %s (%d lines)", len(chunks), file.Path, len(lines))
	return chunks, nil
}

// ExtractMetadata extracts metadata from code content based on language
func (cp *CodeProcessor) ExtractMetadata(content, language string) models.ChunkMetadata {
	metadata := models.ChunkMetadata{
		Functions:  cp.extractFunctions(content, language),
		Classes:    cp.extractClasses(content, language),
		Variables:  cp.extractVariables(content, language),
		Types:      cp.extractTypes(content, language),
		Complexity: cp.calculateComplexity(content),
	}
	
	return metadata
}

// extractFunctions extracts function names from code
func (cp *CodeProcessor) extractFunctions(content, language string) []string {
	var functions []string
	var patterns []*regexp.Regexp
	
	switch language {
	case "javascript", "typescript":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`function\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\(`),
			regexp.MustCompile(`const\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=\s*\(`),
			regexp.MustCompile(`([a-zA-Z_$][a-zA-Z0-9_$]*)\s*:\s*\([^)]*\)\s*=>`),
			regexp.MustCompile(`([a-zA-Z_$][a-zA-Z0-9_$]*)\s*\([^)]*\)\s*{\s*$`),
		}
	case "python":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`def\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`),
			regexp.MustCompile(`async\s+def\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`),
		}
	case "go":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`func\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`),
			regexp.MustCompile(`func\s+\([^)]*\)\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`),
		}
	case "java", "csharp":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`(?:public|private|protected|static|\s)+\s*\w+\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`),
		}
	case "cpp", "c":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`(?:static|extern|inline|\s)*\s*\w+\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\([^)]*\)\s*{`),
		}
	case "rust":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`fn\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`),
		}
	case "php":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`function\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`),
		}
	case "ruby":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`def\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
		}
	default:
		// Generic pattern for unknown languages
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`function\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`),
			regexp.MustCompile(`def\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`),
		}
	}
	
	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 && match[1] != "" {
				functions = append(functions, match[1])
			}
		}
	}
	
	return cp.deduplicateStrings(functions)
}

// extractClasses extracts class names from code
func (cp *CodeProcessor) extractClasses(content, language string) []string {
	var classes []string
	var patterns []*regexp.Regexp
	
	switch language {
	case "javascript", "typescript":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`class\s+([a-zA-Z_$][a-zA-Z0-9_$]*)`),
			regexp.MustCompile(`interface\s+([a-zA-Z_$][a-zA-Z0-9_$]*)`),
		}
	case "python":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`class\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
		}
	case "java", "csharp":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`(?:public|private|protected|\s)*\s*class\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
			regexp.MustCompile(`(?:public|private|protected|\s)*\s*interface\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
		}
	case "cpp":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`class\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
			regexp.MustCompile(`struct\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
		}
	case "go":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`type\s+([a-zA-Z_][a-zA-Z0-9_]*)\s+struct`),
			regexp.MustCompile(`type\s+([a-zA-Z_][a-zA-Z0-9_]*)\s+interface`),
		}
	case "rust":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`struct\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
			regexp.MustCompile(`trait\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
			regexp.MustCompile(`enum\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
		}
	}
	
	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 && match[1] != "" {
				classes = append(classes, match[1])
			}
		}
	}
	
	return cp.deduplicateStrings(classes)
}

// extractVariables extracts variable declarations
func (cp *CodeProcessor) extractVariables(content, language string) []string {
	var variables []string
	var patterns []*regexp.Regexp
	
	switch language {
	case "javascript", "typescript":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`(?:const|let|var)\s+([a-zA-Z_$][a-zA-Z0-9_$]*)`),
		}
	case "python":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*)\s*=`),
		}
	case "go":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`var\s+([a-zA-Z_][a-zA-Z0-9_]*)`),
			regexp.MustCompile(`([a-zA-Z_][a-zA-Z0-9_]*)\s*:=`),
		}
	}
	
	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 && match[1] != "" {
				variables = append(variables, match[1])
			}
		}
	}
	
	// Limit to first 20 variables to avoid noise
	if len(variables) > 20 {
		variables = variables[:20]
	}
	
	return cp.deduplicateStrings(variables)
}

// extractTypes extracts type definitions
func (cp *CodeProcessor) extractTypes(content, language string) []string {
	var types []string
	var patterns []*regexp.Regexp
	
	switch language {
	case "typescript":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`type\s+([a-zA-Z_$][a-zA-Z0-9_$]*)\s*=`),
			regexp.MustCompile(`interface\s+([a-zA-Z_$][a-zA-Z0-9_$]*)`),
		}
	case "go":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`type\s+([a-zA-Z_][a-zA-Z0-9_]*)\s+`),
		}
	case "rust":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`type\s+([a-zA-Z_][a-zA-Z0-9_]*)\s*=`),
		}
	}
	
	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 && match[1] != "" {
				types = append(types, match[1])
			}
		}
	}
	
	return cp.deduplicateStrings(types)
}

// ExtractImports extracts import statements from code
func (cp *CodeProcessor) ExtractImports(content, language string) []string {
	var imports []string
	var patterns []*regexp.Regexp
	
	switch language {
	case "javascript", "typescript":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`import\s+.*\s+from\s+['"]([^'"]+)['"]`),
			regexp.MustCompile(`import\s+['"]([^'"]+)['"]`),
			regexp.MustCompile(`require\(['"]([^'"]+)['"]\)`),
		}
	case "python":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`import\s+([a-zA-Z_][a-zA-Z0-9_.]*)`),
			regexp.MustCompile(`from\s+([a-zA-Z_][a-zA-Z0-9_.]*)\s+import`),
		}
	case "go":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`import\s+"([^"]+)"`),
			regexp.MustCompile(`import\s+\(\s*"([^"]+)"`),
		}
	case "java":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`import\s+([a-zA-Z_][a-zA-Z0-9_.]*)`),
		}
	case "csharp":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`using\s+([a-zA-Z_][a-zA-Z0-9_.]*)`),
		}
	case "rust":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`use\s+([a-zA-Z_][a-zA-Z0-9_:]*)`),
		}
	}
	
	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) > 1 && match[1] != "" {
				imports = append(imports, match[1])
			}
		}
	}
	
	return cp.deduplicateStrings(imports)
}

// calculateComplexity estimates cyclomatic complexity
func (cp *CodeProcessor) calculateComplexity(content string) int {
	complexity := 1 // Base complexity
	
	// Count decision points
	complexityPatterns := []string{
		`if\s*\(`, `else\s+if`, `else`,
		`for\s*\(`, `while\s*\(`, `do\s+{`,
		`switch\s*\(`, `case\s+`, `default:`,
		`catch\s*\(`, `except:`, `finally:`,
		`\?\s*:`, `&&`, `\|\|`,
	}
	
	for _, pattern := range complexityPatterns {
		regex := regexp.MustCompile(pattern)
		matches := regex.FindAllString(content, -1)
		complexity += len(matches)
	}
	
	return complexity
}

// deduplicateStrings removes duplicate strings from slice
func (cp *CodeProcessor) deduplicateStrings(items []string) []string {
	seen := make(map[string]bool)
	var result []string
	
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	
	return result
}