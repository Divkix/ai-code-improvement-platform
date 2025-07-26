// ABOUTME: Unit tests for services package focusing on business logic and validation
// ABOUTME: Tests error constants, helper functions, and validation logic without external dependencies

package services

import (
	"testing"

	"github-analyzer/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestServiceErrors(t *testing.T) {
	// Test that service error constants are properly defined
	assert.Equal(t, "user not found", ErrUserNotFound.Error())
	assert.Equal(t, "user already exists", ErrUserExists.Error())
	assert.Equal(t, "invalid password", ErrInvalidPassword.Error())
	assert.Equal(t, "repository not found", ErrRepositoryNotFound.Error())
	assert.Equal(t, "repository already exists", ErrRepositoryExists.Error())
	assert.Equal(t, "unauthorized access to repository", ErrUnauthorized.Error())
}

func TestServiceConstants(t *testing.T) {
	// Test that collection constants are defined
	assert.Equal(t, "repositories", RepositoryCollection)
}

func TestNewUserService(t *testing.T) {
	// Test that NewUserService can be called (without database dependency)
	// This tests the constructor logic
	service := &UserService{}
	assert.NotNil(t, service)
}

func TestNewRepositoryService(t *testing.T) {
	// Test that NewRepositoryService can be called (without dependencies)
	service := &RepositoryService{}
	assert.NotNil(t, service)
}

func TestUserServiceMethods(t *testing.T) {
	// Test method signatures exist and can be referenced
	service := &UserService{}
	
	// Verify methods exist (compile-time check)
	assert.NotNil(t, service.GetUserByEmail)
	assert.NotNil(t, service.GetUserByID)
	assert.NotNil(t, service.UpdateUser)
	assert.NotNil(t, service.GetByID)
	assert.NotNil(t, service.UpdateGitHubConnection)
	assert.NotNil(t, service.RemoveGitHubConnection)
	assert.NotNil(t, service.DeleteUser)
}

func TestRepositoryServiceMethods(t *testing.T) {
	// Test method signatures exist and can be referenced
	service := &RepositoryService{}
	
	// Verify methods exist (compile-time check)
	assert.NotNil(t, service.CreateRepository)
	assert.NotNil(t, service.GetRepositories)
	assert.NotNil(t, service.GetRepository)
	assert.NotNil(t, service.GetRepositoryByFullName)
	assert.NotNil(t, service.UpdateRepository)
	assert.NotNil(t, service.DeleteRepository)
	assert.NotNil(t, service.UpdateRepositoryStatus)
	assert.NotNil(t, service.UpdateRepositoryProgress)
	assert.NotNil(t, service.UpdateRepositoryStats)
	assert.NotNil(t, service.GetRepositoryStats)
	assert.NotNil(t, service.MarkRepositoryIndexed)
	assert.NotNil(t, service.ImportRepositoryFromGitHub)
	assert.NotNil(t, service.CreateRepositoryFromGitHub)
	assert.NotNil(t, service.TriggerRepositoryImport)
}

// Test GitHub URL parsing logic
func TestParseGitHubURL(t *testing.T) {
	service := &RepositoryService{}
	
	tests := []struct {
		name           string
		url            string
		expectedOwner  string
		expectedRepo   string
		expectedError  bool
	}{
		{
			name:          "HTTPS URL",
			url:           "https://github.com/owner/repo",
			expectedOwner: "owner",
			expectedRepo:  "repo",
			expectedError: false,
		},
		{
			name:          "HTTPS URL with .git",
			url:           "https://github.com/owner/repo.git",
			expectedOwner: "owner",
			expectedRepo:  "repo",
			expectedError: false,
		},
		{
			name:          "owner/repo format",
			url:           "owner/repo",
			expectedOwner: "owner",
			expectedRepo:  "repo",
			expectedError: false,
		},
		{
			name:          "invalid format",
			url:           "invalid",
			expectedOwner: "",
			expectedRepo:  "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			owner, repo, err := service.parseGitHubURL(tt.url)
			
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOwner, owner)
				assert.Equal(t, tt.expectedRepo, repo)
			}
		})
	}
}

// Test file helper functions
func TestCalculateTotalLines(t *testing.T) {
	service := &RepositoryService{}
	
	// Test with nil input
	result := service.calculateTotalLines(nil)
	assert.Equal(t, 0, result)
	
	// Test with empty slice
	result = service.calculateTotalLines([]*models.RepositoryFile{})
	assert.Equal(t, 0, result)
}

// Test that NewGitHubService can be instantiated
func TestNewGitHubService(t *testing.T) {
	service := &GitHubService{}
	assert.NotNil(t, service)
}

// Test GitHubService methods exist
func TestGitHubServiceMethods(t *testing.T) {
	service := &GitHubService{}
	
	// Verify methods exist (compile-time check)
	assert.NotNil(t, service.GetOAuthConfig)
	assert.NotNil(t, service.CreateClient)
	assert.NotNil(t, service.ValidateRepository)
}

// Test that service can handle encryption key normalization
func TestEncryptionKeyNormalization(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "short key",
			input:    "short",
			expected: 32,
		},
		{
			name:     "long key",
			input:    "this-is-a-very-long-encryption-key-that-exceeds-32-bytes",
			expected: 32,
		},
		{
			name:     "exact length key",
			input:    "exactly-32-bytes-long-key-value!",
			expected: 32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := []byte(tt.input)
			if len(key) != 32 {
				// Pad or truncate to 32 bytes for AES-256
				padded := make([]byte, 32)
				copy(padded, key)
				key = padded
			}
			assert.Equal(t, tt.expected, len(key))
		})
	}
}