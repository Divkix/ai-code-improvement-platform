// ABOUTME: GitHub OAuth handlers for authentication and repository operations
// ABOUTME: Implements GitHub OAuth flow, token management, and repository validation

package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strconv"
	"sync"
	"time"

	"acip.divkix.me/internal/middleware"
	"acip.divkix.me/internal/services"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/oauth2"
)

// OAuthState represents a stored OAuth state with expiration
type OAuthState struct {
	UserID    string
	CreatedAt time.Time
}

// GitHubHandler handles GitHub OAuth and repository operations
type GitHubHandler struct {
	githubService *services.GitHubService
	userService   *services.UserService
	stateStore    map[string]OAuthState
	stateMutex    sync.RWMutex
}

// NewGitHubHandler creates a new GitHub handler
func NewGitHubHandler(githubService *services.GitHubService, userService *services.UserService) *GitHubHandler {
	return &GitHubHandler{
		githubService: githubService,
		userService:   userService,
		stateStore:    make(map[string]OAuthState),
	}
}

// GitHubLoginRequest represents the GitHub OAuth login request
type GitHubLoginRequest struct {
	RedirectURI string `json:"redirect_uri"`
}

// GitHubCallbackRequest represents the GitHub OAuth callback request
type GitHubCallbackRequest struct {
	Code  string `json:"code" binding:"required"`
	State string `json:"state" binding:"required"`
}

// GitHubLogin handles GitHub OAuth login initiation
func (h *GitHubHandler) GitHubLogin(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not found in context",
		})
		return
	}

	// Generate state for CSRF protection
	state := generateRandomState()

	// Store state with user association and expiration
	h.storeOAuthState(state, userID)
	
	redirectURI := c.DefaultQuery("redirect_uri", "http://localhost:3000/auth/github/callback")
	
	oauthConfig := h.githubService.GetOAuthConfig()
	oauthConfig.RedirectURL = redirectURI

	authURL := oauthConfig.AuthCodeURL(state,
		oauth2.SetAuthURLParam("scope", "repo user:email"),
		oauth2.AccessTypeOffline,
	)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

// GitHubCallback handles GitHub OAuth callback
func (h *GitHubHandler) GitHubCallback(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not found in context",
		})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Invalid user ID",
		})
		return
	}

	var req GitHubCallbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	// Validate state parameter against stored state for CSRF protection
	if !h.validateOAuthState(req.State, userID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_state",
			"message": "Invalid or expired OAuth state parameter",
		})
		return
	}

	// Exchange code for token
	oauthConfig := h.githubService.GetOAuthConfig()
	token, err := oauthConfig.Exchange(c.Request.Context(), req.Code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "oauth_error",
			"message": "Failed to exchange code for token: " + err.Error(),
		})
		return
	}

	// Get GitHub user information
	client := h.githubService.CreateClient(token.AccessToken)
	githubUser, _, err := client.Users.Get(c.Request.Context(), "")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "github_error",
			"message": "Failed to get GitHub user information: " + err.Error(),
		})
		return
	}

	// Encrypt and store the GitHub token
	encryptedToken, err := h.githubService.EncryptToken(token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to encrypt GitHub token",
		})
		return
	}

	// Update user with GitHub information
	user, err := h.userService.UpdateGitHubConnection(c.Request.Context(), objectID, encryptedToken, githubUser.GetLogin())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to update user with GitHub information: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GitHubDisconnect handles disconnecting GitHub account
func (h *GitHubHandler) GitHubDisconnect(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not found in context",
		})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Invalid user ID",
		})
		return
	}

	// Remove GitHub connection from user
	user, err := h.userService.RemoveGitHubConnection(c.Request.Context(), objectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to disconnect GitHub account: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetGitHubRepositories handles fetching user's GitHub repositories
func (h *GitHubHandler) GetGitHubRepositories(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not found in context",
		})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Invalid user ID",
		})
		return
	}

	// Get user and check GitHub connection
	user, err := h.userService.GetByID(c.Request.Context(), objectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to get user information",
		})
		return
	}

	if user.GitHubToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "github_not_connected",
			"message": "GitHub account is not connected",
		})
		return
	}

	// Decrypt GitHub token
	accessToken, err := h.githubService.DecryptToken(user.GitHubToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to decrypt GitHub token",
		})
		return
	}

	// Parse page parameter
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	// Get repositories from GitHub
	repositories, err := h.githubService.GetUserRepositories(c.Request.Context(), accessToken, page)
	if err != nil {
		if h.githubService.IsRateLimited(err) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limited",
				"message": "GitHub rate limit exceeded",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "github_error",
			"message": "Failed to fetch repositories: " + err.Error(),
		})
		return
	}

	// Check if there are more repositories (simplified check)
	hasMore := len(repositories) == 20

	c.JSON(http.StatusOK, gin.H{
		"repositories": repositories,
		"has_more":     hasMore,
		"current_page": page,
	})
}

// SearchGitHubRepositories handles searching user's GitHub repositories
func (h *GitHubHandler) SearchGitHubRepositories(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not found in context",
		})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Invalid user ID",
		})
		return
	}

	// Get query parameter
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Query parameter 'q' is required",
		})
		return
	}

	// Parse limit parameter
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "6"))
	if limit < 1 || limit > 20 {
		limit = 6
	}

	// Get user and check GitHub connection
	user, err := h.userService.GetByID(c.Request.Context(), objectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to get user information",
		})
		return
	}

	if user.GitHubToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "github_not_connected",
			"message": "GitHub account is not connected",
		})
		return
	}

	// Decrypt GitHub token
	accessToken, err := h.githubService.DecryptToken(user.GitHubToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to decrypt GitHub token",
		})
		return
	}

	// Search repositories
	repositories, err := h.githubService.SearchUserRepositories(c.Request.Context(), accessToken, query, limit)
	if err != nil {
		if h.githubService.IsRateLimited(err) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limited",
				"message": "GitHub rate limit exceeded",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "github_error",
			"message": "Failed to search repositories: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"repositories": repositories,
		"query":        query,
		"total":        len(repositories),
	})
}

// GetRecentGitHubRepositories handles fetching recent user's GitHub repositories
func (h *GitHubHandler) GetRecentGitHubRepositories(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not found in context",
		})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Invalid user ID",
		})
		return
	}

	// Parse limit parameter
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "6"))
	if limit < 1 || limit > 20 {
		limit = 6
	}

	// Get user and check GitHub connection
	user, err := h.userService.GetByID(c.Request.Context(), objectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to get user information",
		})
		return
	}

	if user.GitHubToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "github_not_connected",
			"message": "GitHub account is not connected",
		})
		return
	}

	// Decrypt GitHub token
	accessToken, err := h.githubService.DecryptToken(user.GitHubToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to decrypt GitHub token",
		})
		return
	}

	// Get recent repositories
	repositories, err := h.githubService.GetRecentUserRepositories(c.Request.Context(), accessToken, limit)
	if err != nil {
		if h.githubService.IsRateLimited(err) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limited",
				"message": "GitHub rate limit exceeded",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "github_error",
			"message": "Failed to fetch recent repositories: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"repositories": repositories,
		"total":        len(repositories),
	})
}

// ValidateGitHubRepository handles validating a specific GitHub repository
func (h *GitHubHandler) ValidateGitHubRepository(c *gin.Context) {
	userID, exists := middleware.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "User not found in context",
		})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "Invalid user ID",
		})
		return
	}

	owner := c.Param("owner")
	repo := c.Param("repo")

	if owner == "" || repo == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "Owner and repository name are required",
		})
		return
	}

	// Get user and check GitHub connection
	user, err := h.userService.GetByID(c.Request.Context(), objectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to get user information",
		})
		return
	}

	if user.GitHubToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "github_not_connected",
			"message": "GitHub account is not connected",
		})
		return
	}

	// Decrypt GitHub token
	accessToken, err := h.githubService.DecryptToken(user.GitHubToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to decrypt GitHub token",
		})
		return
	}

	// Validate repository
	repository, err := h.githubService.ValidateRepository(c.Request.Context(), accessToken, owner, repo)
	if err != nil {
		if h.githubService.IsRateLimited(err) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limited",
				"message": "GitHub rate limit exceeded",
			})
			return
		}

		c.JSON(http.StatusNotFound, gin.H{
			"error":   "repository_not_found",
			"message": "Repository not found or not accessible: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, repository)
}

// generateRandomState generates a random state string for OAuth CSRF protection
func generateRandomState() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

// storeOAuthState stores an OAuth state with user association and expiration
func (h *GitHubHandler) storeOAuthState(state, userID string) {
	h.stateMutex.Lock()
	defer h.stateMutex.Unlock()
	
	h.stateStore[state] = OAuthState{
		UserID:    userID,
		CreatedAt: time.Now(),
	}
	
	// Clean up expired states (older than 10 minutes)
	h.cleanupExpiredStates()
}

// validateOAuthState validates an OAuth state parameter
func (h *GitHubHandler) validateOAuthState(state, userID string) bool {
	h.stateMutex.RLock()
	defer h.stateMutex.RUnlock()
	
	storedState, exists := h.stateStore[state]
	if !exists {
		return false
	}
	
	// Check if state has expired (10 minutes)
	if time.Since(storedState.CreatedAt) > 10*time.Minute {
		// Clean up expired state
		delete(h.stateStore, state)
		return false
	}
	
	// Verify the state belongs to the same user
	if storedState.UserID != userID {
		return false
	}
	
	// Remove state after successful validation (one-time use)
	delete(h.stateStore, state)
	return true
}

// cleanupExpiredStates removes expired OAuth states (called with mutex locked)
func (h *GitHubHandler) cleanupExpiredStates() {
	now := time.Now()
	for state, oauthState := range h.stateStore {
		if now.Sub(oauthState.CreatedAt) > 10*time.Minute {
			delete(h.stateStore, state)
		}
	}
}