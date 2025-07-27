// ABOUTME: Chat prompt templates for RAG-based code assistance conversations
// ABOUTME: Provides structured prompts with code snippets and formatting guidelines

package prompts

import (
	"bytes"
	"strings"
	"text/template"
)

// CodeSnippet represents a code snippet to be included in the prompt
type CodeSnippet struct {
	FilePath  string
	Content   string
	StartLine int
	EndLine   int
	Language  string
}

// ChatPromptData contains the data used to render chat prompts
type ChatPromptData struct {
	Snippets []CodeSnippet
	Question string
}

// ChatPromptTemplate is the main template for chat interactions
const ChatPromptTemplate = `You are an AI assistant helping developers understand code. Use ONLY the provided code snippets to answer questions.

{{range .Snippets}}
--- File: {{.FilePath}} ({{.StartLine}}-{{.EndLine}})
{{.Content}}

{{end}}
User question: {{.Question}}

Guidelines:
- Reference file paths & line numbers when discussing code
- Be concise and focused in your responses
- If you're not certain about something, say "I'm not certain"
- Only use information from the provided code snippets
- Format code references as: filename:line_number`

// SystemPromptTemplate provides the system-level instructions
const SystemPromptTemplate = `You are a helpful AI assistant specialized in code analysis and software development. You help developers understand and improve their codebase by analyzing provided code snippets and answering questions about them.

Key principles:
- Always base your answers on the provided code snippets
- Reference specific files and line numbers when relevant
- Be precise and accurate in your explanations
- If information is not available in the provided snippets, clearly state this
- Format code blocks with proper syntax highlighting when possible`

// RenderChatPrompt renders the chat prompt template with the provided data
func RenderChatPrompt(data ChatPromptData) (string, error) {
	tmpl, err := template.New("chatPrompt").Parse(ChatPromptTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// RenderSystemPrompt returns the system prompt (currently static)
func RenderSystemPrompt() string {
	return SystemPromptTemplate
}

// BuildPromptData creates ChatPromptData from code snippets and user question
func BuildPromptData(snippets []CodeSnippet, question string) ChatPromptData {
	return ChatPromptData{
		Snippets: snippets,
		Question: strings.TrimSpace(question),
	}
}

// FormatCodeSnippet creates a formatted code snippet with proper indentation
func FormatCodeSnippet(filePath, content string, startLine, endLine int, language string) CodeSnippet {
	// Clean up the content - remove excessive whitespace but preserve structure
	lines := strings.Split(content, "\n")
	var cleanLines []string

	for _, line := range lines {
		// Preserve empty lines but trim trailing whitespace
		cleanLines = append(cleanLines, strings.TrimRight(line, " \t"))
	}

	cleanContent := strings.Join(cleanLines, "\n")

	return CodeSnippet{
		FilePath:  filePath,
		Content:   cleanContent,
		StartLine: startLine,
		EndLine:   endLine,
		Language:  language,
	}
}

// TruncateIfTooLong truncates the prompt if it exceeds the maximum length
func TruncateIfTooLong(prompt string, maxLength int) string {
	if len(prompt) <= maxLength {
		return prompt
	}

	// Try to truncate at a reasonable point (end of a code block or sentence)
	truncated := prompt[:maxLength-100] // Leave some buffer

	// Find the last complete code block or sentence
	if lastBlock := strings.LastIndex(truncated, "\n---"); lastBlock > 0 {
		return truncated[:lastBlock] + "\n\n[Content truncated due to length...]"
	}

	// Fallback: truncate at word boundary
	if lastSpace := strings.LastIndex(truncated, " "); lastSpace > maxLength/2 {
		return truncated[:lastSpace] + "... [truncated]"
	}

	return truncated + "... [truncated]"
}

// EstimateTokenCount provides a rough estimate of token count for the prompt
func EstimateTokenCount(prompt string) int {
	// Rough approximation: ~4 characters per token on average
	return len(prompt) / 4
}

// GetMaxRecommendedPromptLength returns the maximum prompt length from config
func GetMaxRecommendedPromptLength(maxLength int) int {
	return maxLength
}

