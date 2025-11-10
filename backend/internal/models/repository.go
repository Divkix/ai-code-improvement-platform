// ABOUTME: Repository model for MongoDB storage with all repository metadata
// ABOUTME: Includes validation, indexing, and JSON marshalling for API responses

package models

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Repository represents a GitHub repository in our database
type Repository struct {
	ID                  primitive.ObjectID   `bson:"_id,omitempty" json:"id"`
	UserID              primitive.ObjectID   `bson:"userId" json:"userId"`
	GitHubRepoID        *int64               `bson:"githubRepoId,omitempty" json:"githubRepoId,omitempty"`
	Owner               string               `bson:"owner" json:"owner"`
	Name                string               `bson:"name" json:"name"`
	FullName            string               `bson:"fullName" json:"fullName"`
	Description         *string              `bson:"description,omitempty" json:"description,omitempty"`
	PrimaryLanguage     *string              `bson:"primaryLanguage,omitempty" json:"primaryLanguage,omitempty"`
	IsPrivate           bool                 `bson:"isPrivate" json:"isPrivate"`
	IndexedAt           *time.Time           `bson:"indexedAt,omitempty" json:"indexedAt,omitempty"`
	LastSyncedAt        *time.Time           `bson:"lastSyncedAt,omitempty" json:"lastSyncedAt,omitempty"`
	Status              string               `bson:"status" json:"status"`
	ImportProgress      int                  `bson:"importProgress" json:"importProgress"`
	EmbeddingStatus     string               `bson:"embeddingStatus,omitempty" json:"embeddingStatus,omitempty"`
	EmbeddingProgress   int                  `bson:"embeddingProgress,omitempty" json:"embeddingProgress,omitempty"`
	EmbeddedChunksCount int                  `bson:"embeddedChunksCount,omitempty" json:"embeddedChunksCount,omitempty"`
	LastEmbeddedAt      *time.Time           `bson:"lastEmbeddedAt,omitempty" json:"lastEmbeddedAt,omitempty"`
	Stats               *RepositoryStats     `bson:"stats,omitempty" json:"stats,omitempty"`
	ImportFailures      *RepositoryFailures  `bson:"importFailures,omitempty" json:"importFailures,omitempty"`
	CreatedAt           time.Time            `bson:"createdAt" json:"createdAt"`
	UpdatedAt           time.Time            `bson:"updatedAt" json:"updatedAt"`
}

// RepositoryStats contains detailed statistics about a repository
type RepositoryStats struct {
	TotalFiles     int            `bson:"totalFiles" json:"totalFiles"`
	TotalLines     int            `bson:"totalLines" json:"totalLines"`
	Languages      map[string]int `bson:"languages,omitempty" json:"languages,omitempty"`
	LastCommitDate *time.Time     `bson:"lastCommitDate,omitempty" json:"lastCommitDate,omitempty"`
}

// RepositoryFailures tracks import/processing failures for repositories
type RepositoryFailures struct {
	TotalChunks   int      `bson:"totalChunks" json:"totalChunks"`
	FailedChunks  int      `bson:"failedChunks" json:"failedChunks"`
	SuccessRate   float64  `bson:"successRate" json:"successRate"`
	FailedBatches []string `bson:"failedBatches,omitempty" json:"failedBatches,omitempty"`
	Message       string   `bson:"message,omitempty" json:"message,omitempty"`
	Timestamp     time.Time `bson:"timestamp" json:"timestamp"`
}

// CreateRepositoryRequest represents the request payload for creating a repository
type CreateRepositoryRequest struct {
	Name            string  `json:"name" binding:"required"`
	Owner           string  `json:"owner" binding:"required"`
	FullName        string  `json:"fullName" binding:"required"`
	Description     *string `json:"description,omitempty"`
	GitHubRepoID    *int64  `json:"githubRepoId,omitempty"`
	PrimaryLanguage *string `json:"primaryLanguage,omitempty"`
	IsPrivate       *bool   `json:"isPrivate,omitempty"`
}

// UpdateRepositoryRequest represents the request payload for updating a repository
type UpdateRepositoryRequest struct {
	Name            *string `json:"name,omitempty"`
	Description     *string `json:"description,omitempty"`
	PrimaryLanguage *string `json:"primaryLanguage,omitempty"`
}

// RepositoryListResponse represents the response for listing repositories
type RepositoryListResponse struct {
	Repositories []Repository `json:"repositories"`
	Total        int64        `json:"total"`
}

// Repository status constants
const (
	StatusPending         = "pending"
	StatusImporting       = "importing"
	StatusPartial         = "partial"
	StatusQueuedEmbedding = "queued-embedding"
	StatusEmbedding       = "embedding"
	StatusReady           = "ready"
	StatusError           = "error"
)

// ValidStatus checks if the provided status is valid
func ValidStatus(status string) bool {
	switch status {
	case StatusPending, StatusImporting, StatusPartial, StatusQueuedEmbedding, StatusEmbedding, StatusReady, StatusError:
		return true
	default:
		return false
	}
}

// NewRepository creates a new repository with default values
func NewRepository(userID primitive.ObjectID, req CreateRepositoryRequest) *Repository {
	now := time.Now()
	return &Repository{
		UserID:          userID,
		GitHubRepoID:    req.GitHubRepoID,
		Owner:           req.Owner,
		Name:            req.Name,
		FullName:        req.FullName,
		Description:     req.Description,
		PrimaryLanguage: req.PrimaryLanguage,
		IsPrivate:       req.IsPrivate != nil && *req.IsPrivate,
		Status:          StatusPending,
		ImportProgress:  0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// Update applies updates to the repository
func (r *Repository) Update(req UpdateRepositoryRequest) {
	now := time.Now()
	if req.Name != nil {
		r.Name = *req.Name
	}
	if req.Description != nil {
		r.Description = req.Description
	}
	if req.PrimaryLanguage != nil {
		r.PrimaryLanguage = req.PrimaryLanguage
	}
	r.UpdatedAt = now
}

// SetStatus updates the repository status and timestamp
func (r *Repository) SetStatus(status string) {
	r.Status = status
	r.UpdatedAt = time.Now()
}

// SetImportProgress updates the import progress and timestamp
func (r *Repository) SetImportProgress(progress int) {
	if progress < 0 {
		progress = 0
	} else if progress > 100 {
		progress = 100
	}
	r.ImportProgress = progress
	r.UpdatedAt = time.Now()

	// Automatically update status based on progress
	switch progress {
	case 0:
		r.Status = StatusPending
	case 100:
		r.Status = StatusReady
	default:
		r.Status = StatusImporting
	}
}

// SetStats updates repository statistics
func (r *Repository) SetStats(stats *RepositoryStats) {
	r.Stats = stats
	r.UpdatedAt = time.Now()
}

// SetImportFailures updates repository import failure information
func (r *Repository) SetImportFailures(failures *RepositoryFailures) {
	r.ImportFailures = failures
	r.UpdatedAt = time.Now()
}

// MarkIndexed marks the repository as indexed with current timestamp
func (r *Repository) MarkIndexed() {
	now := time.Now()
	r.IndexedAt = &now
	r.LastSyncedAt = &now
	r.UpdatedAt = now
}

// GetCompositeStatus returns the composite status for display based on both import and embedding status
func (r *Repository) GetCompositeStatus() string {
	// Handle error states first
	if r.Status == StatusError {
		return StatusError
	}
	if r.EmbeddingStatus == "failed" {
		return StatusError
	}

	// Handle partial import state
	if r.Status == StatusPartial {
		return StatusPartial
	}

	// Handle pending/importing states
	if r.Status == StatusPending {
		return StatusPending
	}
	if r.Status == StatusImporting {
		return StatusImporting
	}

	// Handle post-import states based on embedding status
	if r.Status == StatusReady || r.Status == StatusQueuedEmbedding {
		switch r.EmbeddingStatus {
		case "", "pending":
			return StatusQueuedEmbedding
		case "processing":
			return StatusEmbedding
		case "completed":
			return StatusReady
		default:
			// Unknown embedding status, assume queued for embedding
			return StatusQueuedEmbedding
		}
	}

	// Fallback to original status
	return r.Status
}

// MarshalJSON provides custom JSON marshalling that uses composite status for API responses
func (r *Repository) MarshalJSON() ([]byte, error) {
	// Create an anonymous struct that mirrors Repository but with composite status
	type RepositoryAlias Repository
	return json.Marshal(&struct {
		*RepositoryAlias
		Status string `json:"status"`
	}{
		RepositoryAlias: (*RepositoryAlias)(r),
		Status:          r.GetCompositeStatus(),
	})
}
