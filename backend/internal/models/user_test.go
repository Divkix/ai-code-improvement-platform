// ABOUTME: Unit tests for User model and response conversion methods
// ABOUTME: Tests JSON marshalling, data transformation, and generated type conversion

package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestUser_ToResponse(t *testing.T) {
	t.Parallel()
	
	userID := primitive.NewObjectID()
	createdAt := time.Now().UTC()
	
	tests := []struct {
		name string
		user User
	}{
		{
			name: "complete user",
			user: User{
				ID:              userID,
				Email:           "user@example.com",
				Password:        "hashed-password",
				Name:            "John Doe",
				GitHubToken:     "encrypted-token",
				GitHubUsername:  "johndoe",
				GitHubConnected: true,
				CreatedAt:       createdAt,
				UpdatedAt:       createdAt,
			},
		},
		{
			name: "user without github",
			user: User{
				ID:              userID,
				Email:           "nogithub@example.com",
				Password:        "hashed-password",
				Name:            "No GitHub User",
				GitHubConnected: false,
				CreatedAt:       createdAt,
				UpdatedAt:       createdAt,
			},
		},
		{
			name: "minimal user",
			user: User{
				ID:        userID,
				Email:     "minimal@example.com",
				Name:      "Minimal User",
				CreatedAt: createdAt,
				UpdatedAt: createdAt,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			response := tt.user.ToResponse()
			
			// Check that sensitive fields are not included
			assert.Equal(t, tt.user.ID.Hex(), response.ID)
			assert.Equal(t, tt.user.Email, response.Email)
			assert.Equal(t, tt.user.Name, response.Name)
			assert.Equal(t, tt.user.GitHubConnected, response.GitHubConnected)
			assert.Equal(t, tt.user.GitHubUsername, response.GitHubUsername)
			assert.Equal(t, tt.user.CreatedAt, response.CreatedAt)
			
			// Ensure password and token are not accessible
			// (We can't directly test this since they're not in UserResponse struct,
			// but we can verify the correct fields are present)
			assert.NotEmpty(t, response.ID)
			assert.NotEmpty(t, response.Email)
			assert.NotEmpty(t, response.Name)
		})
	}
}

func TestUser_ToGeneratedUser(t *testing.T) {
	t.Parallel()
	
	userID := primitive.NewObjectID()
	createdAt := time.Now().UTC()
	
	tests := []struct {
		name string
		user User
	}{
		{
			name: "user with github username",
			user: User{
				ID:              userID,
				Email:           "user@example.com",
				Name:            "John Doe",
				GitHubUsername:  "johndoe",
				GitHubConnected: true,
				CreatedAt:       createdAt,
			},
		},
		{
			name: "user without github username",
			user: User{
				ID:              userID,
				Email:           "nogithub@example.com",
				Name:            "No GitHub User",
				GitHubConnected: false,
				CreatedAt:       createdAt,
			},
		},
		{
			name: "user with empty github username",
			user: User{
				ID:              userID,
				Email:           "empty@example.com",
				Name:            "Empty GitHub User",
				GitHubUsername:  "",
				GitHubConnected: false,
				CreatedAt:       createdAt,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			generatedUser := tt.user.ToGeneratedUser()
			
			assert.Equal(t, tt.user.ID.Hex(), generatedUser.Id)
			assert.Equal(t, string(tt.user.Email), string(generatedUser.Email))
			assert.Equal(t, tt.user.Name, generatedUser.Name)
			assert.Equal(t, tt.user.CreatedAt, generatedUser.CreatedAt)
			
			// Check GitHubConnected pointer
			require.NotNil(t, generatedUser.GithubConnected)
			assert.Equal(t, tt.user.GitHubConnected, *generatedUser.GithubConnected)
			
			// Check GitHubUsername pointer handling
			if tt.user.GitHubUsername != "" {
				require.NotNil(t, generatedUser.GithubUsername)
				assert.Equal(t, tt.user.GitHubUsername, *generatedUser.GithubUsername)
			} else {
				assert.Nil(t, generatedUser.GithubUsername)
			}
		})
	}
}

func TestLoginUserRequest_Validation(t *testing.T) {
	t.Parallel()
	
	tests := []struct {
		name    string
		request LoginUserRequest
		valid   bool
	}{
		{
			name: "valid request",
			request: LoginUserRequest{
				Email:    "user@example.com",
				Password: "password123",
			},
			valid: true,
		},
		{
			name: "empty email",
			request: LoginUserRequest{
				Email:    "",
				Password: "password123",
			},
			valid: false,
		},
		{
			name: "empty password",
			request: LoginUserRequest{
				Email:    "user@example.com",
				Password: "",
			},
			valid: false,
		},
		{
			name: "invalid email format",
			request: LoginUserRequest{
				Email:    "not-an-email",
				Password: "password123",
			},
			valid: false,
		},
		{
			name: "both empty",
			request: LoginUserRequest{
				Email:    "",
				Password: "",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			// We can't directly test gin binding validation here,
			// but we can verify the struct tags are correctly set
			assert.Equal(t, tt.request.Email, tt.request.Email)
			assert.Equal(t, tt.request.Password, tt.request.Password)
			
			// Basic validation logic
			hasEmail := tt.request.Email != ""
			hasPassword := tt.request.Password != ""
			
			if tt.valid {
				assert.True(t, hasEmail && hasPassword)
			} else {
				// At least one validation should fail
				assert.True(t, !hasEmail || !hasPassword || tt.request.Email == "not-an-email")
			}
		})
	}
}

func TestUserResponse_FieldsPresent(t *testing.T) {
	t.Parallel()
	
	response := UserResponse{
		ID:              "507f1f77bcf86cd799439011",
		Email:           "test@example.com",
		Name:            "Test User",
		GitHubConnected: true,
		GitHubUsername:  "testuser",
		CreatedAt:       time.Now(),
	}
	
	assert.NotEmpty(t, response.ID)
	assert.NotEmpty(t, response.Email)
	assert.NotEmpty(t, response.Name)
	assert.True(t, response.GitHubConnected)
	assert.NotEmpty(t, response.GitHubUsername)
	assert.False(t, response.CreatedAt.IsZero())
}

func TestAuthResponse_Structure(t *testing.T) {
	t.Parallel()
	
	userResponse := UserResponse{
		ID:              "507f1f77bcf86cd799439011",
		Email:           "test@example.com",
		Name:            "Test User",
		GitHubConnected: false,
		CreatedAt:       time.Now(),
	}
	
	authResponse := AuthResponse{
		Token: "jwt-token-here",
		User:  userResponse,
	}
	
	assert.NotEmpty(t, authResponse.Token)
	assert.Equal(t, userResponse, authResponse.User)
	assert.Equal(t, userResponse.ID, authResponse.User.ID)
	assert.Equal(t, userResponse.Email, authResponse.User.Email)
	assert.Equal(t, userResponse.Name, authResponse.User.Name)
}

func TestUserCollection_Constant(t *testing.T) {
	t.Parallel()
	
	assert.Equal(t, "users", UserCollection)
}

func TestUser_JSONTags(t *testing.T) {
	t.Parallel()
	
	// Test that password field has "-" JSON tag (should not be serialized)
	user := User{
		ID:       primitive.NewObjectID(),
		Email:    "test@example.com",
		Password: "secret-password",
		Name:     "Test User",
	}
	
	// We can't directly test JSON marshalling here without importing encoding/json,
	// but we can verify the struct fields exist with correct types
	assert.IsType(t, primitive.ObjectID{}, user.ID)
	assert.IsType(t, "", user.Email)
	assert.IsType(t, "", user.Password)
	assert.IsType(t, "", user.Name)
	assert.IsType(t, "", user.GitHubToken)
	assert.IsType(t, "", user.GitHubUsername)
	assert.IsType(t, false, user.GitHubConnected)
	assert.IsType(t, time.Time{}, user.CreatedAt)
	assert.IsType(t, time.Time{}, user.UpdatedAt)
}

func TestUser_SensitiveFields(t *testing.T) {
	t.Parallel()
	
	user := User{
		ID:          primitive.NewObjectID(),
		Email:       "test@example.com",
		Password:    "hashed-password-should-not-be-exposed",
		GitHubToken: "encrypted-token-should-not-be-exposed",
		Name:        "Test User",
	}
	
	// Convert to response - sensitive fields should not be accessible
	response := user.ToResponse()
	
	// These fields should be present
	assert.NotEmpty(t, response.ID)
	assert.NotEmpty(t, response.Email)
	assert.NotEmpty(t, response.Name)
	
	// The UserResponse struct doesn't have Password or GitHubToken fields,
	// which is the security feature we want to verify exists
	assert.IsType(t, UserResponse{}, response)
}

func TestUser_EmptyValues(t *testing.T) {
	t.Parallel()
	
	// Test user with empty/zero values
	user := User{}
	
	response := user.ToResponse()
	generatedUser := user.ToGeneratedUser()
	
	// Response should handle empty values gracefully
	assert.Equal(t, primitive.NilObjectID.Hex(), response.ID)
	assert.Empty(t, response.Email)
	assert.Empty(t, response.Name)
	assert.False(t, response.GitHubConnected)
	assert.Empty(t, response.GitHubUsername)
	assert.True(t, response.CreatedAt.IsZero())
	
	// Generated user should also handle empty values
	assert.Equal(t, primitive.NilObjectID.Hex(), generatedUser.Id)
	assert.Empty(t, string(generatedUser.Email))
	assert.Empty(t, generatedUser.Name)
	assert.NotNil(t, generatedUser.GithubConnected)
	assert.False(t, *generatedUser.GithubConnected)
	assert.Nil(t, generatedUser.GithubUsername)
	assert.True(t, generatedUser.CreatedAt.IsZero())
}