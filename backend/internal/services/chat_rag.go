// ABOUTME: Chat RAG service orchestrating retrieval-augmented generation for code conversations
// ABOUTME: Combines vector search, prompt building, and LLM streaming for context-aware responses

package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github-analyzer/internal/models"
	"github-analyzer/internal/prompts"
)

// ChatRAGService handles RAG pipeline for chat conversations
type ChatRAGService struct {
	db            *mongo.Database
	searchService *SearchService
	llmService    *LLMService
}

// ChatStreamChunk represents a streaming chat response chunk
type ChatStreamChunk struct {
	Type    string `json:"type"`    // "content", "done", "error"
	Content string `json:"content"` // Text content or error message
	Delta   string `json:"delta"`   // Incremental content for streaming
}

// ErrSessionNotFound is returned when a chat session is not found
var ErrSessionNotFound = errors.New("chat session not found")

// NewChatRAGService creates a new chat RAG service
func NewChatRAGService(db *mongo.Database, searchService *SearchService, llmService *LLMService) *ChatRAGService {
	return &ChatRAGService{
		db:            db,
		searchService: searchService,
		llmService:    llmService,
	}
}

// ProcessMessage processes a user message and returns updated session (non-streaming)
func (s *ChatRAGService) ProcessMessage(ctx context.Context, userID, sessionID primitive.ObjectID, message string) (*models.ChatSession, error) {
	startTime := time.Now()
	
	log.Printf("[CHAT] Processing message for user=%s session=%s message_length=%d",
		userID.Hex(), sessionID.Hex(), len(message))
	
	// Retrieve the session
	session, err := s.getSession(ctx, userID, sessionID)
	if err != nil {
		log.Printf("[CHAT] Failed to retrieve session user=%s session=%s error=%v",
			userID.Hex(), sessionID.Hex(), err)
		return nil, err
	}

	// Add user message to session
	session.AppendMessage(models.RoleUser, message, nil, nil)

	// Get repository context if session is associated with a repository
	var retrievedChunks []models.RetrievedChunk
	var retrievalTime time.Duration
	if session.HasRepository() {
		retrievalStart := time.Now()
		chunks, err := s.retrieveContext(ctx, *session.RepositoryID, message)
		retrievalTime = time.Since(retrievalStart)
		
		if err != nil {
			log.Printf("[CHAT] Failed to retrieve context user=%s session=%s repo=%s error=%v retrieval_time=%v",
				userID.Hex(), sessionID.Hex(), session.RepositoryID.Hex(), err, retrievalTime)
		} else {
			retrievedChunks = chunks
			log.Printf("[CHAT] Retrieved context user=%s session=%s repo=%s chunks_count=%d retrieval_time=%v",
				userID.Hex(), sessionID.Hex(), session.RepositoryID.Hex(), len(chunks), retrievalTime)
		}
	}

	// Build prompt
	promptData := s.buildPromptData(retrievedChunks, message)
	systemPrompt := prompts.RenderSystemPrompt()
	fullPrompt, err := prompts.RenderChatPrompt(promptData)
	if err != nil {
		return nil, fmt.Errorf("failed to render prompt: %w", err)
	}

	// Generate response using LLM
	llmStart := time.Now()
	messages := s.llmService.BuildMessages(systemPrompt, fullPrompt)
	response, err := s.llmService.ChatCompletion(ctx, messages, DefaultChatOptions)
	llmDuration := time.Since(llmStart)
	
	if err != nil {
		log.Printf("[CHAT] LLM generation failed user=%s session=%s llm_time=%v error=%v",
			userID.Hex(), sessionID.Hex(), llmDuration, err)
		return nil, fmt.Errorf("failed to generate response: %w", err)
	}

	if len(response.Choices) == 0 {
		log.Printf("[CHAT] No response choices user=%s session=%s llm_time=%v",
			userID.Hex(), sessionID.Hex(), llmDuration)
		return nil, fmt.Errorf("no response generated")
	}

	// Extract response content and token usage
	assistantContent := response.Choices[0].Message.Content
	tokensUsed := response.Usage.TotalTokens

	log.Printf("[CHAT] LLM generation successful user=%s session=%s llm_time=%v tokens_used=%d response_length=%d",
		userID.Hex(), sessionID.Hex(), llmDuration, tokensUsed, len(assistantContent))

	// Add assistant message to session
	session.AppendMessage(models.RoleAssistant, assistantContent, retrievedChunks, &tokensUsed)

	// Auto-generate title if this is the first exchange
	if session.GetMessageCount() == 2 && session.Title == "New Chat" {
		session.UpdateTitle(session.GenerateTitle())
	}

	// Save updated session
	if err := s.saveSession(ctx, session); err != nil {
		log.Printf("[CHAT] Failed to save session user=%s session=%s error=%v",
			userID.Hex(), sessionID.Hex(), err)
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	totalDuration := time.Since(startTime)
	log.Printf("[CHAT] Message processing completed user=%s session=%s total_time=%v retrieval_time=%v llm_time=%v tokens_used=%d",
		userID.Hex(), sessionID.Hex(), totalDuration, retrievalTime, llmDuration, tokensUsed)

	return session, nil
}

// ProcessMessageStreaming processes a user message and returns streaming response
func (s *ChatRAGService) ProcessMessageStreaming(ctx context.Context, userID, sessionID primitive.ObjectID, message string) (<-chan ChatStreamChunk, error) {
	startTime := time.Now()
	
	log.Printf("[CHAT] Processing streaming message for user=%s session=%s message_length=%d",
		userID.Hex(), sessionID.Hex(), len(message))
	
	// Create response channel
	responseChan := make(chan ChatStreamChunk, 10)

	go func() {
		defer close(responseChan)

		// Retrieve the session
		session, err := s.getSession(ctx, userID, sessionID)
		if err != nil {
			log.Printf("[CHAT] Failed to retrieve session for streaming user=%s session=%s error=%v",
				userID.Hex(), sessionID.Hex(), err)
			responseChan <- ChatStreamChunk{Type: "error", Content: err.Error()}
			return
		}

		// Add user message to session
		session.AppendMessage(models.RoleUser, message, nil, nil)

		// Get repository context if session is associated with a repository
		var retrievedChunks []models.RetrievedChunk
		var retrievalTime time.Duration
		if session.HasRepository() {
			retrievalStart := time.Now()
			chunks, err := s.retrieveContext(ctx, *session.RepositoryID, message)
			retrievalTime = time.Since(retrievalStart)
			
			if err != nil {
				log.Printf("[CHAT] Failed to retrieve context for streaming user=%s session=%s repo=%s error=%v retrieval_time=%v",
					userID.Hex(), sessionID.Hex(), session.RepositoryID.Hex(), err, retrievalTime)
			} else {
				retrievedChunks = chunks
				log.Printf("[CHAT] Retrieved context for streaming user=%s session=%s repo=%s chunks_count=%d retrieval_time=%v",
					userID.Hex(), sessionID.Hex(), session.RepositoryID.Hex(), len(chunks), retrievalTime)
			}
		}

		// Build prompt
		promptData := s.buildPromptData(retrievedChunks, message)
		systemPrompt := prompts.RenderSystemPrompt()
		fullPrompt, err := prompts.RenderChatPrompt(promptData)
		if err != nil {
			responseChan <- ChatStreamChunk{Type: "error", Content: fmt.Sprintf("Failed to render prompt: %v", err)}
			return
		}

		// Generate streaming response using LLM
		llmStart := time.Now()
		messages := s.llmService.BuildMessages(systemPrompt, fullPrompt)
		stream, err := s.llmService.ChatStream(ctx, messages, DefaultChatOptions)
		
		if err != nil {
			llmDuration := time.Since(llmStart)
			log.Printf("[CHAT] Streaming LLM generation failed user=%s session=%s llm_time=%v error=%v",
				userID.Hex(), sessionID.Hex(), llmDuration, err)
			responseChan <- ChatStreamChunk{Type: "error", Content: fmt.Sprintf("Failed to generate response: %v", err)}
			return
		}

		// Collect response content and stream to client
		var fullContent strings.Builder
		var totalTokens int

		for chunk := range stream {
			select {
			case <-ctx.Done():
				return
			default:
				if len(chunk.Choices) > 0 {
					delta := chunk.Choices[0].Delta.Content
					if delta != "" {
						fullContent.WriteString(delta)
						responseChan <- ChatStreamChunk{
							Type:    "content",
							Content: fullContent.String(),
							Delta:   delta,
						}
					}
				}

				// Update token count if available
				if chunk.Usage != nil {
					totalTokens = chunk.Usage.TotalTokens
				}
			}
		}

		// Add assistant message to session
		assistantContent := fullContent.String()
		llmDuration := time.Since(llmStart)
		
		if assistantContent != "" {
			var tokensPtr *int
			if totalTokens > 0 {
				tokensPtr = &totalTokens
			}
			session.AppendMessage(models.RoleAssistant, assistantContent, retrievedChunks, tokensPtr)

			log.Printf("[CHAT] Streaming LLM generation successful user=%s session=%s llm_time=%v tokens_used=%d response_length=%d",
				userID.Hex(), sessionID.Hex(), llmDuration, totalTokens, len(assistantContent))

			// Auto-generate title if this is the first exchange
			if session.GetMessageCount() == 2 && session.Title == "New Chat" {
				session.UpdateTitle(session.GenerateTitle())
			}

			// Save updated session
			if err := s.saveSession(ctx, session); err != nil {
				log.Printf("[CHAT] Failed to save streaming session user=%s session=%s error=%v",
					userID.Hex(), sessionID.Hex(), err)
				responseChan <- ChatStreamChunk{Type: "error", Content: fmt.Sprintf("Failed to save session: %v", err)}
				return
			}
		}

		totalDuration := time.Since(startTime)
		log.Printf("[CHAT] Streaming message processing completed user=%s session=%s total_time=%v retrieval_time=%v llm_time=%v tokens_used=%d",
			userID.Hex(), sessionID.Hex(), totalDuration, retrievalTime, llmDuration, totalTokens)

		// Send completion signal
		responseChan <- ChatStreamChunk{Type: "done", Content: ""}
	}()

	return responseChan, nil
}

// retrieveContext retrieves relevant code chunks for the given query
func (s *ChatRAGService) retrieveContext(ctx context.Context, repositoryID primitive.ObjectID, query string) ([]models.RetrievedChunk, error) {
	// Use hybrid search to get the most relevant chunks (8 chunks as specified in slice9.md)
	const contextChunks = 8
	const vectorWeight = 0.7 // Favor vector search over text search

	results, err := s.searchService.HybridSearch(ctx, repositoryID, query, contextChunks, vectorWeight)
	if err != nil {
		return nil, fmt.Errorf("failed to perform hybrid search: %w", err)
	}

	// Convert search results to retrieved chunks
	var chunks []models.RetrievedChunk
	for _, result := range results {
		similarity := float64(result.Score) // Convert float32 to float64
		chunk := models.RetrievedChunk{
			ChunkID:    result.ID,
			FilePath:   result.FilePath,
			StartLine:  result.StartLine,
			EndLine:    result.EndLine,
			Content:    result.Content,
			Similarity: &similarity,
			Language:   &result.Language,
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// buildPromptData creates prompt data from retrieved chunks and user query
func (s *ChatRAGService) buildPromptData(chunks []models.RetrievedChunk, query string) prompts.ChatPromptData {
	var snippets []prompts.CodeSnippet
	
	for _, chunk := range chunks {
		language := ""
		if chunk.Language != nil {
			language = *chunk.Language
		}
		
		snippet := prompts.FormatCodeSnippet(
			chunk.FilePath,
			chunk.Content,
			chunk.StartLine,
			chunk.EndLine,
			language,
		)
		snippets = append(snippets, snippet)
	}

	return prompts.BuildPromptData(snippets, query)
}

// getSession retrieves a chat session by ID and user ID
func (s *ChatRAGService) getSession(ctx context.Context, userID, sessionID primitive.ObjectID) (*models.ChatSession, error) {
	collection := s.db.Collection(models.ChatSessionCollection)
	
	filter := bson.M{
		"_id":    sessionID,
		"userId": userID,
	}

	var session models.ChatSession
	err := collection.FindOne(ctx, filter).Decode(&session)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to retrieve session: %w", err)
	}

	return &session, nil
}

// saveSession saves a chat session to the database
func (s *ChatRAGService) saveSession(ctx context.Context, session *models.ChatSession) error {
	collection := s.db.Collection(models.ChatSessionCollection)
	
	filter := bson.M{"_id": session.ID}
	update := bson.M{"$set": session}
	
	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}