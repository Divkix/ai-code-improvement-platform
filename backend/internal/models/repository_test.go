// ABOUTME: Unit tests for Repository model methods and validation logic
// ABOUTME: Tests creation, updates, status management, and progress tracking

package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestValidStatus(t *testing.T) {
	t.Parallel()
	
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{
			name:     "valid status - pending",
			status:   StatusPending,
			expected: true,
		},
		{
			name:     "valid status - importing",
			status:   StatusImporting,
			expected: true,
		},
		{
			name:     "valid status - ready",
			status:   StatusReady,
			expected: true,
		},
		{
			name:     "valid status - error",
			status:   StatusError,
			expected: true,
		},
		{
			name:     "invalid status - empty",
			status:   "",
			expected: false,
		},
		{
			name:     "invalid status - random",
			status:   "invalid-status",
			expected: false,
		},
		{
			name:     "invalid status - case sensitive",
			status:   "PENDING",
			expected: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := ValidStatus(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewRepository(t *testing.T) {
	t.Parallel()
	
	userID := primitive.NewObjectID()
	githubRepoID := int64(12345)
	description := "Test repository"
	primaryLanguage := "Go"
	isPrivate := true
	
	tests := []struct {
		name string
		req  CreateRepositoryRequest
	}{
		{
			name: "complete repository request",
			req: CreateRepositoryRequest{
				Name:            "test-repo",
				Owner:           "test-owner",
				FullName:        "test-owner/test-repo",
				Description:     &description,
				GitHubRepoID:    &githubRepoID,
				PrimaryLanguage: &primaryLanguage,
				IsPrivate:       &isPrivate,
			},
		},
		{
			name: "minimal repository request",
			req: CreateRepositoryRequest{
				Name:     "minimal-repo",
				Owner:    "minimal-owner",
				FullName: "minimal-owner/minimal-repo",
			},
		},
		{
			name: "repository with nil IsPrivate",
			req: CreateRepositoryRequest{
				Name:      "nil-private-repo",
				Owner:     "nil-owner",
				FullName:  "nil-owner/nil-private-repo",
				IsPrivate: nil,
			},
		},
		{
			name: "repository with false IsPrivate",
			req: CreateRepositoryRequest{
				Name:      "false-private-repo",
				Owner:     "false-owner",
				FullName:  "false-owner/false-private-repo",
				IsPrivate: func() *bool { b := false; return &b }(),
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			startTime := time.Now()
			repo := NewRepository(userID, tt.req)
			endTime := time.Now()
			
			// Basic fields
			assert.Equal(t, userID, repo.UserID)
			assert.Equal(t, tt.req.Name, repo.Name)
			assert.Equal(t, tt.req.Owner, repo.Owner)
			assert.Equal(t, tt.req.FullName, repo.FullName)
			assert.Equal(t, tt.req.Description, repo.Description)
			assert.Equal(t, tt.req.GitHubRepoID, repo.GitHubRepoID)
			assert.Equal(t, tt.req.PrimaryLanguage, repo.PrimaryLanguage)
			
			// Default values
			assert.Equal(t, StatusPending, repo.Status)
			assert.Equal(t, 0, repo.ImportProgress)
			assert.Nil(t, repo.IndexedAt)
			assert.Nil(t, repo.LastSyncedAt)
			assert.Nil(t, repo.Stats)
			
			// IsPrivate handling
			expectedPrivate := tt.req.IsPrivate != nil && *tt.req.IsPrivate
			assert.Equal(t, expectedPrivate, repo.IsPrivate)
			
			// Timestamps
			assert.True(t, repo.CreatedAt.After(startTime) || repo.CreatedAt.Equal(startTime))
			assert.True(t, repo.CreatedAt.Before(endTime) || repo.CreatedAt.Equal(endTime))
			assert.Equal(t, repo.CreatedAt, repo.UpdatedAt)
		})
	}
}

func TestRepository_Update(t *testing.T) {
	t.Parallel()
	
	userID := primitive.NewObjectID()
	originalRepo := NewRepository(userID, CreateRepositoryRequest{
		Name:     "original-repo",
		Owner:    "original-owner",
		FullName: "original-owner/original-repo",
	})
	
	// Wait a bit to ensure UpdatedAt changes
	time.Sleep(time.Millisecond)
	
	tests := []struct {
		name string
		req  UpdateRepositoryRequest
	}{
		{
			name: "update all fields",
			req: UpdateRepositoryRequest{
				Name:            func() *string { s := "updated-name"; return &s }(),
				Description:     func() *string { s := "updated description"; return &s }(),
				PrimaryLanguage: func() *string { s := "JavaScript"; return &s }(),
			},
		},
		{
			name: "update only name",
			req: UpdateRepositoryRequest{
				Name: func() *string { s := "only-name-updated"; return &s }(),
			},
		},
		{
			name: "update only description",
			req: UpdateRepositoryRequest{
				Description: func() *string { s := "only description updated"; return &s }(),
			},
		},
		{
			name: "update only primary language",
			req: UpdateRepositoryRequest{
				PrimaryLanguage: func() *string { s := "Python"; return &s }(),
			},
		},
		{
			name: "empty update request",
			req:  UpdateRepositoryRequest{},
		},
		{
			name: "update with nil description",
			req: UpdateRepositoryRequest{
				Description: nil,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			// Create a copy of the original repo
			repo := *originalRepo
			originalUpdatedAt := repo.UpdatedAt
			
			startTime := time.Now()
			repo.Update(tt.req)
			endTime := time.Now()
			
			// Check that UpdatedAt was changed
			assert.True(t, repo.UpdatedAt.After(originalUpdatedAt))
			assert.True(t, repo.UpdatedAt.After(startTime) || repo.UpdatedAt.Equal(startTime))
			assert.True(t, repo.UpdatedAt.Before(endTime) || repo.UpdatedAt.Equal(endTime))
			
			// Check field updates
			if tt.req.Name != nil {
				assert.Equal(t, *tt.req.Name, repo.Name)
			} else {
				assert.Equal(t, originalRepo.Name, repo.Name)
			}
			
			if tt.req.Description != nil {
				assert.Equal(t, tt.req.Description, repo.Description)
			} else {
				assert.Equal(t, originalRepo.Description, repo.Description)
			}
			
			if tt.req.PrimaryLanguage != nil {
				assert.Equal(t, tt.req.PrimaryLanguage, repo.PrimaryLanguage)
			} else {
				assert.Equal(t, originalRepo.PrimaryLanguage, repo.PrimaryLanguage)
			}
			
			// Check that other fields remain unchanged
			assert.Equal(t, originalRepo.UserID, repo.UserID)
			assert.Equal(t, originalRepo.Owner, repo.Owner)
			assert.Equal(t, originalRepo.FullName, repo.FullName)
			assert.Equal(t, originalRepo.Status, repo.Status)
			assert.Equal(t, originalRepo.ImportProgress, repo.ImportProgress)
		})
	}
}

func TestRepository_SetStatus(t *testing.T) {
	t.Parallel()
	
	userID := primitive.NewObjectID()
	repo := NewRepository(userID, CreateRepositoryRequest{
		Name:     "test-repo",
		Owner:    "test-owner",
		FullName: "test-owner/test-repo",
	})
	
	originalUpdatedAt := repo.UpdatedAt
	time.Sleep(time.Millisecond)
	
	startTime := time.Now()
	repo.SetStatus(StatusImporting)
	endTime := time.Now()
	
	assert.Equal(t, StatusImporting, repo.Status)
	assert.True(t, repo.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, repo.UpdatedAt.After(startTime) || repo.UpdatedAt.Equal(startTime))
	assert.True(t, repo.UpdatedAt.Before(endTime) || repo.UpdatedAt.Equal(endTime))
}

func TestRepository_SetImportProgress(t *testing.T) {
	t.Parallel()
	
	userID := primitive.NewObjectID()
	
	tests := []struct {
		name           string
		progress       int
		expectedProgress int
		expectedStatus string
	}{
		{
			name:           "progress 0",
			progress:       0,
			expectedProgress: 0,
			expectedStatus: StatusPending,
		},
		{
			name:           "progress 50",
			progress:       50,
			expectedProgress: 50,
			expectedStatus: StatusImporting,
		},
		{
			name:           "progress 100",
			progress:       100,
			expectedProgress: 100,
			expectedStatus: StatusReady,
		},
		{
			name:           "progress negative",
			progress:       -10,
			expectedProgress: 0,
			expectedStatus: StatusPending,
		},
		{
			name:           "progress over 100",
			progress:       150,
			expectedProgress: 100,
			expectedStatus: StatusReady,
		},
		{
			name:           "progress 1",
			progress:       1,
			expectedProgress: 1,
			expectedStatus: StatusImporting,
		},
		{
			name:           "progress 99",
			progress:       99,
			expectedProgress: 99,
			expectedStatus: StatusImporting,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			
			repo := NewRepository(userID, CreateRepositoryRequest{
				Name:     "test-repo",
				Owner:    "test-owner",
				FullName: "test-owner/test-repo",
			})
			
			originalUpdatedAt := repo.UpdatedAt
			time.Sleep(time.Millisecond)
			
			startTime := time.Now()
			repo.SetImportProgress(tt.progress)
			endTime := time.Now()
			
			assert.Equal(t, tt.expectedProgress, repo.ImportProgress)
			assert.Equal(t, tt.expectedStatus, repo.Status)
			assert.True(t, repo.UpdatedAt.After(originalUpdatedAt))
			assert.True(t, repo.UpdatedAt.After(startTime) || repo.UpdatedAt.Equal(startTime))
			assert.True(t, repo.UpdatedAt.Before(endTime) || repo.UpdatedAt.Equal(endTime))
		})
	}
}

func TestRepository_SetStats(t *testing.T) {
	t.Parallel()
	
	userID := primitive.NewObjectID()
	repo := NewRepository(userID, CreateRepositoryRequest{
		Name:     "test-repo",
		Owner:    "test-owner",
		FullName: "test-owner/test-repo",
	})
	
	originalUpdatedAt := repo.UpdatedAt
	time.Sleep(time.Millisecond)
	
	lastCommitDate := time.Now().Add(-24 * time.Hour)
	stats := &RepositoryStats{
		TotalFiles: 42,
		TotalLines: 1337,
		Languages: map[string]int{
			"Go":         800,
			"JavaScript": 400,
			"CSS":        137,
		},
		LastCommitDate: &lastCommitDate,
	}
	
	startTime := time.Now()
	repo.SetStats(stats)
	endTime := time.Now()
	
	require.NotNil(t, repo.Stats)
	assert.Equal(t, stats.TotalFiles, repo.Stats.TotalFiles)
	assert.Equal(t, stats.TotalLines, repo.Stats.TotalLines)
	assert.Equal(t, stats.Languages, repo.Stats.Languages)
	assert.Equal(t, stats.LastCommitDate, repo.Stats.LastCommitDate)
	assert.True(t, repo.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, repo.UpdatedAt.After(startTime) || repo.UpdatedAt.Equal(startTime))
	assert.True(t, repo.UpdatedAt.Before(endTime) || repo.UpdatedAt.Equal(endTime))
	
	// Test with nil stats
	repo.SetStats(nil)
	assert.Nil(t, repo.Stats)
}

func TestRepository_MarkIndexed(t *testing.T) {
	t.Parallel()
	
	userID := primitive.NewObjectID()
	repo := NewRepository(userID, CreateRepositoryRequest{
		Name:     "test-repo",
		Owner:    "test-owner",
		FullName: "test-owner/test-repo",
	})
	
	originalUpdatedAt := repo.UpdatedAt
	time.Sleep(time.Millisecond)
	
	startTime := time.Now()
	repo.MarkIndexed()
	endTime := time.Now()
	
	require.NotNil(t, repo.IndexedAt)
	require.NotNil(t, repo.LastSyncedAt)
	
	assert.True(t, repo.IndexedAt.After(startTime) || repo.IndexedAt.Equal(startTime))
	assert.True(t, repo.IndexedAt.Before(endTime) || repo.IndexedAt.Equal(endTime))
	
	assert.True(t, repo.LastSyncedAt.After(startTime) || repo.LastSyncedAt.Equal(startTime))
	assert.True(t, repo.LastSyncedAt.Before(endTime) || repo.LastSyncedAt.Equal(endTime))
	
	assert.True(t, repo.UpdatedAt.After(originalUpdatedAt))
	assert.True(t, repo.UpdatedAt.After(startTime) || repo.UpdatedAt.Equal(startTime))
	assert.True(t, repo.UpdatedAt.Before(endTime) || repo.UpdatedAt.Equal(endTime))
	
	// Check that all three timestamps are approximately the same
	assert.True(t, repo.IndexedAt.Equal(*repo.LastSyncedAt))
	timeDiff := repo.UpdatedAt.Sub(*repo.IndexedAt)
	assert.True(t, timeDiff >= 0 && timeDiff < time.Second)
}

func TestRepositoryStats(t *testing.T) {
	t.Parallel()
	
	lastCommitDate := time.Now().Add(-48 * time.Hour)
	
	stats := &RepositoryStats{
		TotalFiles: 100,
		TotalLines: 5000,
		Languages: map[string]int{
			"Go":         3000,
			"JavaScript": 1500,
			"HTML":       400,
			"CSS":        100,
		},
		LastCommitDate: &lastCommitDate,
	}
	
	assert.Equal(t, 100, stats.TotalFiles)
	assert.Equal(t, 5000, stats.TotalLines)
	assert.Equal(t, 4, len(stats.Languages))
	assert.Equal(t, 3000, stats.Languages["Go"])
	assert.Equal(t, 1500, stats.Languages["JavaScript"])
	assert.Equal(t, 400, stats.Languages["HTML"])
	assert.Equal(t, 100, stats.Languages["CSS"])
	assert.Equal(t, &lastCommitDate, stats.LastCommitDate)
}

func TestRepositoryConstants(t *testing.T) {
	t.Parallel()
	
	assert.Equal(t, "pending", StatusPending)
	assert.Equal(t, "importing", StatusImporting)
	assert.Equal(t, "ready", StatusReady)
	assert.Equal(t, "error", StatusError)
}