// ABOUTME: Unit tests for JWT authentication middleware
// ABOUTME: Tests authorization header validation, token processing, and context handling

package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"acip.divkix.me/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

func TestAuthMiddleware_Success(t *testing.T) {
	t.Parallel()
	
	// Gin test mode set globally in TestMain
	
	// Create auth service and generate a valid token
	authService := auth.NewAuthService("test-secret")
	userID := "user123"
	email := "user@example.com"
	
	token, err := authService.GenerateToken(userID, email)
	require.NoError(t, err)
	
	// Create test handler
	testHandler := func(c *gin.Context) {
		// Verify user info is set in context
		contextUserID, exists := GetUserIDFromContext(c)
		assert.True(t, exists)
		assert.Equal(t, userID, contextUserID)
		
		contextEmail, exists := GetUserEmailFromContext(c)
		assert.True(t, exists)
		assert.Equal(t, email, contextEmail)
		
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	}
	
	// Setup router with middleware
	router := gin.New()
	router.Use(AuthMiddleware(authService))
	router.GET("/protected", testHandler)
	
	// Create request with valid token
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Verify response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

func TestAuthMiddleware_MissingAuthHeader(t *testing.T) {
	t.Parallel()
	
	// Gin test mode set globally in TestMain
	
	authService := auth.NewAuthService("test-secret")
	
	// Create test handler (should not be reached)
	testHandler := func(c *gin.Context) {
		t.Error("Handler should not be called when auth fails")
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	}
	
	// Setup router with middleware
	router := gin.New()
	router.Use(AuthMiddleware(authService))
	router.GET("/protected", testHandler)
	
	// Create request without auth header
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Verify response
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Missing authorization header")
	assert.Contains(t, w.Body.String(), "unauthorized")
}

func TestAuthMiddleware_InvalidAuthHeaderFormat(t *testing.T) {
	t.Parallel()
	
	// Gin test mode set globally in TestMain
	
	authService := auth.NewAuthService("test-secret")
	
	tests := []struct {
		name      string
		authHeader string
	}{
		{
			name:      "no bearer prefix",
			authHeader: "invalid-token",
		},
		{
			name:      "wrong prefix",
			authHeader: "Basic invalid-token",
		},
		{
			name:      "missing token",
			authHeader: "Bearer",
		},
		{
			name:      "case sensitive bearer",
			authHeader: "bearer token123",
		},
	}
	
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			// Create test handler (should not be reached)
			testHandler := func(c *gin.Context) {
				t.Error("Handler should not be called when auth fails")
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			}
			
			// Setup router with middleware
			router := gin.New()
			router.Use(AuthMiddleware(authService))
			router.GET("/protected", testHandler)
			
			// Create request with invalid auth header
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", tt.authHeader)
			
			// Record response
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			// Verify response
			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.Contains(t, w.Body.String(), "Invalid authorization header format")
			assert.Contains(t, w.Body.String(), "unauthorized")
		})
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	t.Parallel()
	
	// Gin test mode set globally in TestMain
	
	authService := auth.NewAuthService("test-secret")
	
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "malformed token",
			token: "invalid.token.format",
		},
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "whitespace only token",
			token: "   ",
		},
		{
			name:  "token with leading space",
			token: " token123",
		},
		{
			name:  "random string",
			token: "random-string-not-jwt",
		},
		{
			name:  "wrong signing method",
			token: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.invalid",
		},
	}
	
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			// Create test handler (should not be reached)
			testHandler := func(c *gin.Context) {
				t.Error("Handler should not be called when auth fails")
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			}
			
			// Setup router with middleware
			router := gin.New()
			router.Use(AuthMiddleware(authService))
			router.GET("/protected", testHandler)
			
			// Create request with invalid token
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", "Bearer "+tt.token)
			
			// Record response
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			
			// Verify response
			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.Contains(t, w.Body.String(), "Invalid or expired token")
			assert.Contains(t, w.Body.String(), "unauthorized")
		})
	}
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	t.Parallel()
	
	// Gin test mode set globally in TestMain
	
	// Create token with one secret
	tokenService := auth.NewAuthService("correct-secret")
	userID := "user123"
	email := "user@example.com"
	
	token, err := tokenService.GenerateToken(userID, email)
	require.NoError(t, err)
	
	// Create middleware with different secret
	authService := auth.NewAuthService("wrong-secret")
	
	// Create test handler (should not be reached)
	testHandler := func(c *gin.Context) {
		t.Error("Handler should not be called when auth fails")
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	}
	
	// Setup router with middleware
	router := gin.New()
	router.Use(AuthMiddleware(authService))
	router.GET("/protected", testHandler)
	
	// Create request with token signed by different secret
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	
	// Record response
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Verify response
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired token")
	assert.Contains(t, w.Body.String(), "unauthorized")
}

func TestGetUserIDFromContext_Success(t *testing.T) {
	t.Parallel()
	
	// Gin test mode set globally in TestMain
	
	expectedUserID := "user123"
	
	// Create gin context and set user ID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", expectedUserID)
	
	// Test GetUserIDFromContext
	userID, exists := GetUserIDFromContext(c)
	assert.True(t, exists)
	assert.Equal(t, expectedUserID, userID)
}

func TestGetUserIDFromContext_NotExists(t *testing.T) {
	t.Parallel()
	
	// Gin test mode set globally in TestMain
	
	// Create gin context without setting user ID
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	// Test GetUserIDFromContext
	userID, exists := GetUserIDFromContext(c)
	assert.False(t, exists)
	assert.Empty(t, userID)
}

func TestGetUserEmailFromContext_Success(t *testing.T) {
	t.Parallel()
	
	// Gin test mode set globally in TestMain
	
	expectedEmail := "user@example.com"
	
	// Create gin context and set user email
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_email", expectedEmail)
	
	// Test GetUserEmailFromContext
	email, exists := GetUserEmailFromContext(c)
	assert.True(t, exists)
	assert.Equal(t, expectedEmail, email)
}

func TestGetUserEmailFromContext_NotExists(t *testing.T) {
	t.Parallel()
	
	// Gin test mode set globally in TestMain
	
	// Create gin context without setting user email
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	
	// Test GetUserEmailFromContext
	email, exists := GetUserEmailFromContext(c)
	assert.False(t, exists)
	assert.Empty(t, email)
}

func TestAuthMiddleware_MultipleCalls(t *testing.T) {
	t.Parallel()
	
	// Gin test mode set globally in TestMain
	
	authService := auth.NewAuthService("test-secret")
	userID := "user123"
	email := "user@example.com"
	
	token, err := authService.GenerateToken(userID, email)
	require.NoError(t, err)
	
	callCount := 0
	
	// Create test handler that increments counter
	testHandler := func(c *gin.Context) {
		callCount++
		
		// Verify user info is consistently available
		contextUserID, exists := GetUserIDFromContext(c)
		assert.True(t, exists)
		assert.Equal(t, userID, contextUserID)
		
		contextEmail, exists := GetUserEmailFromContext(c)
		assert.True(t, exists)
		assert.Equal(t, email, contextEmail)
		
		c.JSON(http.StatusOK, gin.H{"message": "success", "call": callCount})
	}
	
	// Setup router with middleware
	router := gin.New()
	router.Use(AuthMiddleware(authService))
	router.GET("/protected", testHandler)
	
	// Make multiple requests
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "success")
	}
	
	assert.Equal(t, 3, callCount)
}

func TestAuthMiddleware_ContextIsolation(t *testing.T) {
	t.Parallel()
	
	// Gin test mode set globally in TestMain
	
	authService := auth.NewAuthService("test-secret")
	
	// Create two different users
	user1ID := "user1"
	user1Email := "user1@example.com"
	token1, err := authService.GenerateToken(user1ID, user1Email)
	require.NoError(t, err)
	
	user2ID := "user2"
	user2Email := "user2@example.com"
	token2, err := authService.GenerateToken(user2ID, user2Email)
	require.NoError(t, err)
	
	// Create test handler that stores user info
	var capturedUsers []map[string]string
	testHandler := func(c *gin.Context) {
		userID, exists := GetUserIDFromContext(c)
		require.True(t, exists)
		
		email, exists := GetUserEmailFromContext(c)
		require.True(t, exists)
		
		capturedUsers = append(capturedUsers, map[string]string{
			"id":    userID,
			"email": email,
		})
		
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	}
	
	// Setup router with middleware
	router := gin.New()
	router.Use(AuthMiddleware(authService))
	router.GET("/protected", testHandler)
	
	// Make request with user1 token
	req1 := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req1.Header.Set("Authorization", "Bearer "+token1)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	
	// Make request with user2 token
	req2 := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req2.Header.Set("Authorization", "Bearer "+token2)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	
	// Verify both requests succeeded
	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Equal(t, http.StatusOK, w2.Code)
	
	// Verify context isolation - each request should have correct user info
	require.Len(t, capturedUsers, 2)
	
	assert.Equal(t, user1ID, capturedUsers[0]["id"])
	assert.Equal(t, user1Email, capturedUsers[0]["email"])
	
	assert.Equal(t, user2ID, capturedUsers[1]["id"])
	assert.Equal(t, user2Email, capturedUsers[1]["email"])
}