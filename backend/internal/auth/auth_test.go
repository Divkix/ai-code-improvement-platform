// ABOUTME: Unit tests for JWT authentication service
// ABOUTME: Tests password hashing, JWT generation, and token validation

package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAuthService(t *testing.T) {
	t.Parallel()
	
	secret := "test-secret-key"
	service := NewAuthService(secret)
	
	assert.NotNil(t, service)
	assert.Equal(t, []byte(secret), service.jwtSecret)
}

func TestHashPassword(t *testing.T) {
	t.Parallel()
	
	service := NewAuthService("test-secret")
	
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false,
		},
		{
			name:     "long password",
			password: "very-long-password-with-special-characters-!@#$%^&*()",
			wantErr:  false,
		},
		{
			name:     "unicode password",
			password: "пароль123",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			hashedPassword, err := service.HashPassword(tt.password)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, hashedPassword)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hashedPassword)
				assert.NotEqual(t, tt.password, hashedPassword)
				// Ensure hash starts with bcrypt prefix
				assert.Contains(t, hashedPassword, "$2a$")
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	t.Parallel()
	
	service := NewAuthService("test-secret")
	
	tests := []struct {
		name           string
		password       string
		checkPassword  string
		expectMatch    bool
	}{
		{
			name:          "correct password",
			password:      "password123",
			checkPassword: "password123",
			expectMatch:   true,
		},
		{
			name:          "incorrect password",
			password:      "password123",
			checkPassword: "wrongpassword",
			expectMatch:   false,
		},
		{
			name:          "empty passwords",
			password:      "",
			checkPassword: "",
			expectMatch:   true,
		},
		{
			name:          "case sensitive",
			password:      "Password123",
			checkPassword: "password123",
			expectMatch:   false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			// First hash the password
			hashedPassword, err := service.HashPassword(tt.password)
			require.NoError(t, err)
			
			// Then check if the check password matches
			err = service.CheckPassword(hashedPassword, tt.checkPassword)
			
			if tt.expectMatch {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	t.Parallel()
	
	service := NewAuthService("test-secret-key")
	
	tests := []struct {
		name    string
		userID  string
		email   string
		wantErr bool
	}{
		{
			name:    "valid user data",
			userID:  "user123",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "empty user ID",
			userID:  "",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "empty email",
			userID:  "user123",
			email:   "",
			wantErr: false,
		},
		{
			name:    "both empty",
			userID:  "",
			email:   "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			token, err := service.GenerateToken(tt.userID, tt.email)
			
			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
				
				// Verify token structure (should have 3 parts separated by dots)
				assert.Len(t, splitToken(token), 3)
			}
		})
	}
}

func TestValidateToken(t *testing.T) {
	t.Parallel()
	
	service := NewAuthService("test-secret-key")
	userID := "user123"
	email := "user@example.com"
	
	t.Run("valid token", func(t *testing.T) {
		t.Parallel()
		
		token, err := service.GenerateToken(userID, email)
		require.NoError(t, err)
		
		claims, err := service.ValidateToken(token)
		assert.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
		assert.Equal(t, userID, claims.Subject)
		assert.True(t, claims.ExpiresAt.After(time.Now()))
	})
	
	t.Run("invalid token format", func(t *testing.T) {
		t.Parallel()
		
		claims, err := service.ValidateToken("invalid-token")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
	
	t.Run("empty token", func(t *testing.T) {
		t.Parallel()
		
		claims, err := service.ValidateToken("")
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
	
	t.Run("token with wrong secret", func(t *testing.T) {
		t.Parallel()
		
		wrongService := NewAuthService("wrong-secret")
		token, err := wrongService.GenerateToken(userID, email)
		require.NoError(t, err)
		
		claims, err := service.ValidateToken(token)
		assert.Error(t, err)
		assert.Nil(t, claims)
	})
	
	t.Run("expired token", func(t *testing.T) {
		t.Parallel()
		
		// Create a token that expires immediately
		claims := Claims{
			UserID: userID,
			Email:  email,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)), // Already expired
				IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
				NotBefore: jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
				Subject:   userID,
			},
		}
		
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		expiredToken, err := token.SignedString(service.jwtSecret)
		require.NoError(t, err)
		
		validatedClaims, err := service.ValidateToken(expiredToken)
		assert.Error(t, err)
		assert.Nil(t, validatedClaims)
	})
	
	t.Run("token with wrong signing method", func(t *testing.T) {
		t.Parallel()
		
		// Create a malformed token with wrong signing method (RS256 instead of HS256)
		malformedToken := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWV9.malformed"
		
		validatedClaims, err := service.ValidateToken(malformedToken)
		assert.Error(t, err)
		assert.Nil(t, validatedClaims)
	})
}

func TestTokenRoundTrip(t *testing.T) {
	t.Parallel()
	
	service := NewAuthService("test-secret-key")
	userID := "user123"
	email := "user@example.com"
	
	// Generate token
	token, err := service.GenerateToken(userID, email)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	
	// Validate token
	claims, err := service.ValidateToken(token)
	require.NoError(t, err)
	require.NotNil(t, claims)
	
	// Verify claims
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, userID, claims.Subject)
	assert.True(t, claims.ExpiresAt.After(time.Now()))
	assert.True(t, claims.IssuedAt.Before(time.Now().Add(time.Second)))
	assert.True(t, claims.NotBefore.Before(time.Now().Add(time.Second)))
}

// Helper function to split JWT token
func splitToken(token string) []string {
	result := []string{}
	current := ""
	
	for _, char := range token {
		if char == '.' {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	
	if current != "" {
		result = append(result, current)
	}
	
	return result
}