// ABOUTME: Unit tests for chat prompt template functionality
// ABOUTME: Tests template rendering, code snippet formatting, and truncation logic

package prompts

import (
	"strings"
	"testing"
)

func TestRenderChatPrompt(t *testing.T) {
	t.Run("renders prompt with single snippet", func(t *testing.T) {
		snippets := []CodeSnippet{
			{
				FilePath:  "src/main.go",
				Content:   "func main() {\n    fmt.Println(\"Hello, world!\")\n}",
				StartLine: 10,
				EndLine:   12,
				Language:  "go",
			},
		}
		
		data := BuildPromptData(snippets, "What does this function do?")
		
		result, err := RenderChatPrompt(data)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		
		// Check that the prompt contains expected elements
		if !strings.Contains(result, "src/main.go") {
			t.Error("Expected prompt to contain file path")
		}
		
		if !strings.Contains(result, "(10-12)") {
			t.Error("Expected prompt to contain line range")
		}
		
		if !strings.Contains(result, "What does this function do?") {
			t.Error("Expected prompt to contain user question")
		}
		
		if !strings.Contains(result, "func main()") {
			t.Error("Expected prompt to contain code content")
		}
	})
	
	t.Run("renders prompt with multiple snippets", func(t *testing.T) {
		snippets := []CodeSnippet{
			{
				FilePath:  "src/main.go",
				Content:   "func main() { ... }",
				StartLine: 10,
				EndLine:   15,
			},
			{
				FilePath:  "src/utils.go",
				Content:   "func helper() { ... }",
				StartLine: 5,
				EndLine:   8,
			},
		}
		
		data := BuildPromptData(snippets, "How do these functions work together?")
		
		result, err := RenderChatPrompt(data)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		
		// Check both files are included
		if !strings.Contains(result, "src/main.go") {
			t.Error("Expected prompt to contain first file path")
		}
		
		if !strings.Contains(result, "src/utils.go") {
			t.Error("Expected prompt to contain second file path")
		}
		
		// Check both code snippets are included
		if !strings.Contains(result, "func main()") {
			t.Error("Expected prompt to contain first function")
		}
		
		if !strings.Contains(result, "func helper()") {
			t.Error("Expected prompt to contain second function")
		}
	})
	
	t.Run("handles empty snippets", func(t *testing.T) {
		data := BuildPromptData([]CodeSnippet{}, "What is this about?")
		
		result, err := RenderChatPrompt(data)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		
		if !strings.Contains(result, "What is this about?") {
			t.Error("Expected prompt to contain user question")
		}
		
		// Should not contain any file references
		if strings.Contains(result, "File:") {
			t.Error("Expected no file references for empty snippets")
		}
	})
}

func TestBuildPromptData(t *testing.T) {
	snippets := []CodeSnippet{
		{FilePath: "test.go", Content: "test content"},
	}
	question := "  What is this?  "
	
	data := BuildPromptData(snippets, question)
	
	if len(data.Snippets) != 1 {
		t.Errorf("Expected 1 snippet, got %v", len(data.Snippets))
	}
	
	if data.Question != "What is this?" {
		t.Errorf("Expected trimmed question 'What is this?', got %v", data.Question)
	}
}

func TestFormatCodeSnippet(t *testing.T) {
	t.Run("formats code snippet correctly", func(t *testing.T) {
		content := "func test() {\n    return true\n}"
		snippet := FormatCodeSnippet("test.go", content, 5, 7, "go")
		
		if snippet.FilePath != "test.go" {
			t.Errorf("Expected FilePath 'test.go', got %v", snippet.FilePath)
		}
		
		if snippet.StartLine != 5 {
			t.Errorf("Expected StartLine 5, got %v", snippet.StartLine)
		}
		
		if snippet.EndLine != 7 {
			t.Errorf("Expected EndLine 7, got %v", snippet.EndLine)
		}
		
		if snippet.Language != "go" {
			t.Errorf("Expected Language 'go', got %v", snippet.Language)
		}
		
		if snippet.Content != content {
			t.Errorf("Expected content to be preserved")
		}
	})
	
	t.Run("trims trailing whitespace", func(t *testing.T) {
		content := "func test() {   \n    return true   \n}  "
		snippet := FormatCodeSnippet("test.go", content, 1, 3, "go")
		
		expected := "func test() {\n    return true\n}"
		if snippet.Content != expected {
			t.Errorf("Expected trimmed content:\n%q\nGot:\n%q", expected, snippet.Content)
		}
	})
}

func TestTruncateIfTooLong(t *testing.T) {
	t.Run("does not truncate short content", func(t *testing.T) {
		content := "This is a short prompt"
		result := TruncateIfTooLong(content, 100)
		
		if result != content {
			t.Errorf("Expected content to remain unchanged, got %v", result)
		}
	})
	
	t.Run("truncates long content", func(t *testing.T) {
		content := strings.Repeat("This is a very long prompt that should be truncated. ", 100)
		result := TruncateIfTooLong(content, 200)
		
		if len(result) >= len(content) {
			t.Error("Expected content to be truncated")
		}
		
		if !strings.Contains(result, "truncated") {
			t.Error("Expected truncation indicator")
		}
	})
	
	t.Run("truncates at code block boundary", func(t *testing.T) {
		content := `Some content
--- File: test.go
Code here
--- File: test2.go
More code
` + strings.Repeat("Additional content ", 100)
		
		result := TruncateIfTooLong(content, 200)
		
		if !strings.Contains(result, "Content truncated due to length") {
			t.Error("Expected code block truncation message")
		}
	})
}

func TestEstimateTokenCount(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		expected int
	}{
		{
			name:     "empty string",
			content:  "",
			expected: 0,
		},
		{
			name:     "short content",
			content:  "test",
			expected: 1,
		},
		{
			name:     "longer content",
			content:  "This is a test message with more content",
			expected: 10, // 40 chars / 4 = 10
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := EstimateTokenCount(tc.content)
			if result != tc.expected {
				t.Errorf("Expected %v tokens, got %v", tc.expected, result)
			}
		})
	}
}

func TestRenderSystemPrompt(t *testing.T) {
	result := RenderSystemPrompt()
	
	if result == "" {
		t.Error("Expected non-empty system prompt")
	}
	
	if !strings.Contains(result, "AI assistant") {
		t.Error("Expected system prompt to mention AI assistant")
	}
	
	if !strings.Contains(result, "code analysis") {
		t.Error("Expected system prompt to mention code analysis")
	}
}

func TestMaxRecommendedPromptLength(t *testing.T) {
	// Just a sanity check that the constant is reasonable
	if MaxRecommendedPromptLength <= 0 {
		t.Error("Expected MaxRecommendedPromptLength to be positive")
	}
	
	if MaxRecommendedPromptLength < 1000 {
		t.Error("Expected MaxRecommendedPromptLength to be at least 1000 characters")
	}
}