// ABOUTME: Unit tests for chat session model functionality
// ABOUTME: Tests chat session creation, message handling, and helper methods

package models

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestNewChatSession(t *testing.T) {
	userID := primitive.NewObjectID()
	
	t.Run("creates session with default title", func(t *testing.T) {
		req := CreateChatSessionRequest{}
		session := NewChatSession(userID, req)
		
		if session.UserID != userID {
			t.Errorf("Expected UserID %v, got %v", userID, session.UserID)
		}
		
		if session.Title != "New Chat" {
			t.Errorf("Expected title 'New Chat', got %v", session.Title)
		}
		
		if session.RepositoryID != nil {
			t.Errorf("Expected RepositoryID to be nil, got %v", session.RepositoryID)
		}
		
		if len(session.Messages) != 0 {
			t.Errorf("Expected empty messages array, got %v", len(session.Messages))
		}
	})
	
	t.Run("creates session with custom title", func(t *testing.T) {
		title := "My Custom Chat"
		req := CreateChatSessionRequest{Title: &title}
		session := NewChatSession(userID, req)
		
		if session.Title != title {
			t.Errorf("Expected title %v, got %v", title, session.Title)
		}
	})
	
	t.Run("creates session with repository ID", func(t *testing.T) {
		repoID := primitive.NewObjectID()
		repoIDStr := repoID.Hex()
		req := CreateChatSessionRequest{RepositoryID: &repoIDStr}
		session := NewChatSession(userID, req)
		
		if session.RepositoryID == nil {
			t.Error("Expected RepositoryID to be set")
		} else if *session.RepositoryID != repoID {
			t.Errorf("Expected RepositoryID %v, got %v", repoID, *session.RepositoryID)
		}
	})
	
	t.Run("handles invalid repository ID", func(t *testing.T) {
		invalidID := "invalid-id"
		req := CreateChatSessionRequest{RepositoryID: &invalidID}
		session := NewChatSession(userID, req)
		
		if session.RepositoryID != nil {
			t.Error("Expected RepositoryID to be nil for invalid ID")
		}
	})
}

func TestChatSession_AppendMessage(t *testing.T) {
	userID := primitive.NewObjectID()
	session := NewChatSession(userID, CreateChatSessionRequest{})
	
	t.Run("appends user message", func(t *testing.T) {
		content := "Hello, how are you?"
		session.AppendMessage(RoleUser, content, nil, nil)
		
		if len(session.Messages) != 1 {
			t.Errorf("Expected 1 message, got %v", len(session.Messages))
		}
		
		message := session.Messages[0]
		if message.Role != RoleUser {
			t.Errorf("Expected role %v, got %v", RoleUser, message.Role)
		}
		
		if message.Content != content {
			t.Errorf("Expected content %v, got %v", content, message.Content)
		}
		
		if message.TokensUsed != nil {
			t.Errorf("Expected nil TokensUsed, got %v", *message.TokensUsed)
		}
	})
	
	t.Run("appends assistant message with retrieved chunks", func(t *testing.T) {
		content := "Here's what I found in your code:"
		chunks := []RetrievedChunk{
			{
				ChunkID:   primitive.NewObjectID(),
				FilePath:  "src/main.go",
				StartLine: 10,
				EndLine:   20,
				Content:   "func main() { ... }",
			},
		}
		tokensUsed := 150
		
		session.AppendMessage(RoleAssistant, content, chunks, &tokensUsed)
		
		if len(session.Messages) != 2 {
			t.Errorf("Expected 2 messages, got %v", len(session.Messages))
		}
		
		message := session.Messages[1]
		if message.Role != RoleAssistant {
			t.Errorf("Expected role %v, got %v", RoleAssistant, message.Role)
		}
		
		if len(message.RetrievedChunks) != 1 {
			t.Errorf("Expected 1 retrieved chunk, got %v", len(message.RetrievedChunks))
		}
		
		if message.TokensUsed == nil || *message.TokensUsed != tokensUsed {
			t.Errorf("Expected tokensUsed %v, got %v", tokensUsed, message.TokensUsed)
		}
		
		chunk := message.RetrievedChunks[0]
		if chunk.FilePath != "src/main.go" {
			t.Errorf("Expected FilePath 'src/main.go', got %v", chunk.FilePath)
		}
	})
}

func TestChatSession_GetLastMessage(t *testing.T) {
	userID := primitive.NewObjectID()
	session := NewChatSession(userID, CreateChatSessionRequest{})
	
	t.Run("returns nil for empty session", func(t *testing.T) {
		lastMessage := session.GetLastMessage()
		if lastMessage != nil {
			t.Error("Expected nil for empty session")
		}
	})
	
	t.Run("returns last message", func(t *testing.T) {
		session.AppendMessage(RoleUser, "First message", nil, nil)
		session.AppendMessage(RoleAssistant, "Second message", nil, nil)
		
		lastMessage := session.GetLastMessage()
		if lastMessage == nil {
			t.Error("Expected last message, got nil")
		} else if lastMessage.Content != "Second message" {
			t.Errorf("Expected 'Second message', got %v", lastMessage.Content)
		}
	})
}

func TestChatSession_GetTotalTokensUsed(t *testing.T) {
	userID := primitive.NewObjectID()
	session := NewChatSession(userID, CreateChatSessionRequest{})
	
	t.Run("returns 0 for empty session", func(t *testing.T) {
		total := session.GetTotalTokensUsed()
		if total != 0 {
			t.Errorf("Expected 0 tokens, got %v", total)
		}
	})
	
	t.Run("calculates total tokens correctly", func(t *testing.T) {
		tokens1 := 100
		tokens2 := 150
		
		session.AppendMessage(RoleUser, "Message 1", nil, &tokens1)
		session.AppendMessage(RoleAssistant, "Message 2", nil, &tokens2)
		session.AppendMessage(RoleUser, "Message 3", nil, nil) // No tokens
		
		total := session.GetTotalTokensUsed()
		expected := tokens1 + tokens2
		if total != expected {
			t.Errorf("Expected %v tokens, got %v", expected, total)
		}
	})
}

func TestChatSession_GenerateTitle(t *testing.T) {
	userID := primitive.NewObjectID()
	session := NewChatSession(userID, CreateChatSessionRequest{})
	
	t.Run("returns default title for empty session", func(t *testing.T) {
		title := session.GenerateTitle()
		if title != "New Chat" {
			t.Errorf("Expected 'New Chat', got %v", title)
		}
	})
	
	t.Run("generates title from first user message", func(t *testing.T) {
		session.AppendMessage(RoleUser, "How do I implement authentication?", nil, nil)
		session.AppendMessage(RoleAssistant, "Here's how...", nil, nil)
		
		title := session.GenerateTitle()
		if title != "How do I implement authentication?" {
			t.Errorf("Expected 'How do I implement authentication?', got %v", title)
		}
	})
	
	t.Run("truncates long titles", func(t *testing.T) {
		longMessage := "This is a very long message that should be truncated because it exceeds the maximum length allowed for titles"
		session2 := NewChatSession(userID, CreateChatSessionRequest{})
		session2.AppendMessage(RoleUser, longMessage, nil, nil)
		
		title := session2.GenerateTitle()
		if len(title) > 50 {
			t.Errorf("Expected title length <= 50, got %v", len(title))
		}
		
		if title[len(title)-3:] != "..." {
			t.Errorf("Expected truncated title to end with '...', got %v", title)
		}
	})
}

func TestChatSession_HasRepository(t *testing.T) {
	userID := primitive.NewObjectID()
	
	t.Run("returns false when no repository", func(t *testing.T) {
		session := NewChatSession(userID, CreateChatSessionRequest{})
		if session.HasRepository() {
			t.Error("Expected false for session without repository")
		}
	})
	
	t.Run("returns true when has repository", func(t *testing.T) {
		repoID := primitive.NewObjectID()
		repoIDStr := repoID.Hex()
		req := CreateChatSessionRequest{RepositoryID: &repoIDStr}
		session := NewChatSession(userID, req)
		
		if !session.HasRepository() {
			t.Error("Expected true for session with repository")
		}
	})
}

func TestValidRole(t *testing.T) {
	validRoles := []string{RoleUser, RoleAssistant, RoleSystem}
	invalidRoles := []string{"invalid", "admin", "moderator", ""}
	
	for _, role := range validRoles {
		if !ValidRole(role) {
			t.Errorf("Expected role %v to be valid", role)
		}
	}
	
	for _, role := range invalidRoles {
		if ValidRole(role) {
			t.Errorf("Expected role %v to be invalid", role)
		}
	}
}

func TestChatSession_UpdateTitle(t *testing.T) {
	userID := primitive.NewObjectID()
	session := NewChatSession(userID, CreateChatSessionRequest{})
	originalUpdatedAt := session.UpdatedAt
	
	// Wait a bit to ensure timestamp changes
	time.Sleep(time.Millisecond)
	
	newTitle := "Updated Title"
	session.UpdateTitle(newTitle)
	
	if session.Title != newTitle {
		t.Errorf("Expected title %v, got %v", newTitle, session.Title)
	}
	
	if !session.UpdatedAt.After(originalUpdatedAt) {
		t.Error("Expected UpdatedAt to be updated")
	}
}