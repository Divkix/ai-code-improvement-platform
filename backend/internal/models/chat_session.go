// ABOUTME: Chat session model for MongoDB storage with conversation history
// ABOUTME: Includes message handling, retrieved chunks tracking, and token usage

package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatSession represents a conversation session between user and AI
type ChatSession struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"userId" json:"userId"`
	RepositoryID *primitive.ObjectID `bson:"repositoryId,omitempty" json:"repositoryId,omitempty"`
	Title        string             `bson:"title" json:"title"`
	Messages     []ChatMessage      `bson:"messages" json:"messages"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// ChatMessage represents a single message in a chat session
type ChatMessage struct {
	ID              primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	Role            string             `bson:"role" json:"role"` // "user" or "assistant"
	Content         string             `bson:"content" json:"content"`
	RetrievedChunks []RetrievedChunk   `bson:"retrievedChunks,omitempty" json:"retrievedChunks,omitempty"`
	TokensUsed     *int               `bson:"tokensUsed,omitempty" json:"tokensUsed,omitempty"`
	Timestamp      time.Time          `bson:"timestamp" json:"timestamp"`
}

// RetrievedChunk represents a code chunk that was retrieved for context
type RetrievedChunk struct {
	ChunkID     primitive.ObjectID `bson:"chunkId" json:"chunkId"`
	FilePath    string             `bson:"filePath" json:"filePath"`
	StartLine   int                `bson:"startLine" json:"startLine"`
	EndLine     int                `bson:"endLine" json:"endLine"`
	Content     string             `bson:"content" json:"content"`
	Similarity  *float64           `bson:"similarity,omitempty" json:"similarity,omitempty"`
	Language    *string            `bson:"language,omitempty" json:"language,omitempty"`
}

// CreateChatSessionRequest represents the request payload for creating a chat session
type CreateChatSessionRequest struct {
	RepositoryID *string `json:"repositoryId,omitempty"`
	Title        *string `json:"title,omitempty"`
}

// SendMessageRequest represents the request payload for sending a message
type SendMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

// ChatSessionListResponse represents the response for listing chat sessions
type ChatSessionListResponse struct {
	Sessions []ChatSession `json:"sessions"`
	Total    int64         `json:"total"`
}

// Message role constants
const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
)

// ValidRole checks if the provided role is valid
func ValidRole(role string) bool {
	switch role {
	case RoleUser, RoleAssistant, RoleSystem:
		return true
	default:
		return false
	}
}

// NewChatSession creates a new chat session with default values
func NewChatSession(userID primitive.ObjectID, req CreateChatSessionRequest) *ChatSession {
	now := time.Now()
	
	var repositoryID *primitive.ObjectID
	if req.RepositoryID != nil && *req.RepositoryID != "" {
		if objID, err := primitive.ObjectIDFromHex(*req.RepositoryID); err == nil {
			repositoryID = &objID
		}
	}
	
	title := "New Chat"
	if req.Title != nil && *req.Title != "" {
		title = *req.Title
	}
	
	return &ChatSession{
		UserID:       userID,
		RepositoryID: repositoryID,
		Title:        title,
		Messages:     []ChatMessage{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// AppendMessage adds a new message to the chat session
func (cs *ChatSession) AppendMessage(role, content string, retrievedChunks []RetrievedChunk, tokensUsed *int) {
	now := time.Now()
	
	message := ChatMessage{
		ID:              primitive.NewObjectID(),
		Role:            role,
		Content:         content,
		RetrievedChunks: retrievedChunks,
		TokensUsed:     tokensUsed,
		Timestamp:      now,
	}
	
	cs.Messages = append(cs.Messages, message)
	cs.UpdatedAt = now
}

// GetLastMessage returns the last message in the session, or nil if empty
func (cs *ChatSession) GetLastMessage() *ChatMessage {
	if len(cs.Messages) == 0 {
		return nil
	}
	return &cs.Messages[len(cs.Messages)-1]
}

// GetMessageCount returns the total number of messages in the session
func (cs *ChatSession) GetMessageCount() int {
	return len(cs.Messages)
}

// GetTotalTokensUsed calculates the total tokens used across all messages
func (cs *ChatSession) GetTotalTokensUsed() int {
	total := 0
	for _, message := range cs.Messages {
		if message.TokensUsed != nil {
			total += *message.TokensUsed
		}
	}
	return total
}

// UpdateTitle updates the chat session title
func (cs *ChatSession) UpdateTitle(title string) {
	cs.Title = title
	cs.UpdatedAt = time.Now()
}

// GenerateTitle creates a title from the first user message (up to 50 chars)
func (cs *ChatSession) GenerateTitle() string {
	for _, message := range cs.Messages {
		if message.Role == RoleUser && len(message.Content) > 0 {
			if len(message.Content) <= 50 {
				return message.Content
			}
			return message.Content[:47] + "..."
		}
	}
	return "New Chat"
}

// HasRepository checks if the session is associated with a repository
func (cs *ChatSession) HasRepository() bool {
	return cs.RepositoryID != nil
}

// Collection name for MongoDB
const ChatSessionCollection = "chat_sessions"