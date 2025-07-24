// ABOUTME: Chat handlers for AI-powered conversations about code repositories
// ABOUTME: Implements CRUD operations and streaming chat with RAG pipeline

package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github-analyzer/internal/generated"
	"github-analyzer/internal/models"
	"github-analyzer/internal/services"
)

// ChatHandler handles chat-related HTTP requests
type ChatHandler struct {
	db          *mongo.Database
	chatService *services.ChatRAGService
}

// NewChatHandler creates a new chat handler
func NewChatHandler(db *mongo.Database, chatService *services.ChatRAGService) *ChatHandler {
	return &ChatHandler{
		db:          db,
		chatService: chatService,
	}
}

// CreateChatSession handles POST /api/chat/sessions
func (h *ChatHandler) CreateChatSession(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "User not authenticated"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_user_id", "message": "Invalid user ID format"})
		return
	}

	var req models.CreateChatSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Allow empty request body
		req = models.CreateChatSessionRequest{}
	}

	session := models.NewChatSession(userObjID, req)

	collection := h.db.Collection(models.ChatSessionCollection)
	result, err := collection.InsertOne(c.Request.Context(), session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database_error", "message": "Failed to create chat session"})
		return
	}

	session.ID = result.InsertedID.(primitive.ObjectID)
	c.JSON(http.StatusCreated, session)
}

// ListChatSessions handles GET /api/chat/sessions
func (h *ChatHandler) ListChatSessions(c *gin.Context, params generated.ListChatSessionsParams) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "User not authenticated"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_user_id", "message": "Invalid user ID format"})
		return
	}

	// Parse query parameters from struct
	limit := 20
	if params.Limit != nil && *params.Limit > 0 && *params.Limit <= 100 {
		limit = *params.Limit
	}

	offset := 0
	if params.Offset != nil && *params.Offset >= 0 {
		offset = *params.Offset
	}

	collection := h.db.Collection(models.ChatSessionCollection)
	filter := bson.M{"userId": userObjID}

	// Get total count
	total, err := collection.CountDocuments(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database_error", "message": "Failed to count chat sessions"})
		return
	}

	// Get sessions with pagination, sorted by updatedAt desc
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "updatedAt", Value: -1}})

	cursor, err := collection.Find(c.Request.Context(), filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database_error", "message": "Failed to fetch chat sessions"})
		return
	}
	defer func() {
		if err := cursor.Close(c.Request.Context()); err != nil {
			log.Printf("[ERROR] Failed to close cursor: %v", err)
		}
	}()

	var sessions []models.ChatSession
	if err := cursor.All(c.Request.Context(), &sessions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database_error", "message": "Failed to decode chat sessions"})
		return
	}

	if sessions == nil {
		sessions = []models.ChatSession{}
	}

	response := models.ChatSessionListResponse{
		Sessions: sessions,
		Total:    total,
	}

	c.JSON(http.StatusOK, response)
}

// GetChatSession handles GET /api/chat/sessions/{id}
func (h *ChatHandler) GetChatSession(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "User not authenticated"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_user_id", "message": "Invalid user ID format"})
		return
	}

	sessionID := c.Param("id")
	sessionObjID, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_session_id", "message": "Invalid session ID format"})
		return
	}

	collection := h.db.Collection(models.ChatSessionCollection)
	filter := bson.M{
		"_id":    sessionObjID,
		"userId": userObjID, // Ensure user owns the session
	}

	var session models.ChatSession
	err = collection.FindOne(c.Request.Context(), filter).Decode(&session)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			c.JSON(http.StatusNotFound, gin.H{"error": "session_not_found", "message": "Chat session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database_error", "message": "Failed to fetch chat session"})
		return
	}

	c.JSON(http.StatusOK, session)
}

// DeleteChatSession handles DELETE /api/chat/sessions/{id}
func (h *ChatHandler) DeleteChatSession(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "User not authenticated"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_user_id", "message": "Invalid user ID format"})
		return
	}

	sessionID := c.Param("id")
	sessionObjID, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_session_id", "message": "Invalid session ID format"})
		return
	}

	collection := h.db.Collection(models.ChatSessionCollection)
	filter := bson.M{
		"_id":    sessionObjID,
		"userId": userObjID, // Ensure user owns the session
	}

	result, err := collection.DeleteOne(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database_error", "message": "Failed to delete chat session"})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "session_not_found", "message": "Chat session not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

// SendChatMessage handles POST /api/chat/sessions/{id}/message
func (h *ChatHandler) SendChatMessage(c *gin.Context) {
	startTime := time.Now()
	
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "User not authenticated"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_user_id", "message": "Invalid user ID format"})
		return
	}

	sessionID := c.Param("id")
	sessionObjID, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_session_id", "message": "Invalid session ID format"})
		return
	}

	log.Printf("[HTTP] Chat message request user=%s session=%s client_ip=%s",
		userObjID.Hex(), sessionObjID.Hex(), c.ClientIP())

	var req models.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "message": "Invalid request body"})
		return
	}

	// Check if client wants streaming response
	acceptHeader := c.GetHeader("Accept")
	isStreaming := acceptHeader == "text/event-stream"

	log.Printf("[HTTP] Chat message processing user=%s session=%s streaming=%v message_length=%d",
		userObjID.Hex(), sessionObjID.Hex(), isStreaming, len(req.Content))

	if isStreaming {
		h.handleStreamingMessage(c, userObjID, sessionObjID, req.Content, startTime)
	} else {
		h.handleNonStreamingMessage(c, userObjID, sessionObjID, req.Content, startTime)
	}
}

// handleStreamingMessage handles streaming chat responses
func (h *ChatHandler) handleStreamingMessage(c *gin.Context, userObjID, sessionObjID primitive.ObjectID, message string, startTime time.Time) {
	// Set headers for Server-Sent Events
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	// Use the RAG service to process the message with streaming
	responseStream, err := h.chatService.ProcessMessageStreaming(ctx, userObjID, sessionObjID, message)
	if err != nil {
		// Send error as SSE event
		if _, writeErr := fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", err.Error()); writeErr != nil {
			log.Printf("[ERROR] Failed to write error to stream: %v", writeErr)
		}
		c.Writer.Flush()
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "streaming_not_supported", "message": "Streaming not supported"})
		return
	}

	// Stream the response
	for chunk := range responseStream {
		select {
		case <-ctx.Done():
			return
		default:
			// Send chunk as SSE event
			data, _ := json.Marshal(chunk)
			if _, writeErr := fmt.Fprintf(c.Writer, "data: %s\n\n", data); writeErr != nil {
				log.Printf("[ERROR] Failed to write chunk to stream: %v", writeErr)
			}
			flusher.Flush()
		}
	}

	// Send final event to indicate completion
	if _, writeErr := fmt.Fprintf(c.Writer, "event: done\ndata: {}\n\n"); writeErr != nil {
		log.Printf("[ERROR] Failed to write completion event to stream: %v", writeErr)
	}
	flusher.Flush()
	
	totalDuration := time.Since(startTime)
	log.Printf("[HTTP] Streaming chat completed user=%s session=%s total_time=%v",
		userObjID.Hex(), sessionObjID.Hex(), totalDuration)
}

// handleNonStreamingMessage handles regular JSON chat responses
func (h *ChatHandler) handleNonStreamingMessage(c *gin.Context, userObjID, sessionObjID primitive.ObjectID, message string, startTime time.Time) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer cancel()

	// Use the RAG service to process the message
	updatedSession, err := h.chatService.ProcessMessage(ctx, userObjID, sessionObjID, message)
	if err != nil {
		if errors.Is(err, services.ErrSessionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "session_not_found", "message": "Chat session not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "processing_error", "message": err.Error()})
		return
	}

	totalDuration := time.Since(startTime)
	log.Printf("[HTTP] Non-streaming chat completed user=%s session=%s total_time=%v",
		userObjID.Hex(), sessionObjID.Hex(), totalDuration)
	
	c.JSON(http.StatusOK, updatedSession)
}

// The methods above already implement the generated ServerInterface correctly
// No additional wrappers needed since the signatures match