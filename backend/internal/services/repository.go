// ABOUTME: Repository service for MongoDB operations including CRUD and repository management
// ABOUTME: Handles repository creation, updates, deletion and statistics with user ownership validation

package services

import (
	"context"
	"errors"
	"time"

	"github-analyzer/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrRepositoryNotFound    = errors.New("repository not found")
	ErrRepositoryExists      = errors.New("repository already exists")
	ErrUnauthorized         = errors.New("unauthorized access to repository")
)

const RepositoryCollection = "repositories"

// RepositoryService provides repository-related operations
type RepositoryService struct {
	collection *mongo.Collection
}

// NewRepositoryService creates a new repository service
func NewRepositoryService(db *mongo.Database) *RepositoryService {
	return &RepositoryService{
		collection: db.Collection(RepositoryCollection),
	}
}

// CreateRepository creates a new repository
func (s *RepositoryService) CreateRepository(ctx context.Context, userID primitive.ObjectID, req models.CreateRepositoryRequest) (*models.Repository, error) {
	// Check if repository already exists for this user
	existing, err := s.GetRepositoryByFullName(ctx, userID, req.FullName)
	if err != nil && err != ErrRepositoryNotFound {
		return nil, err
	}
	if existing != nil {
		return nil, ErrRepositoryExists
	}

	// Create new repository
	repo := models.NewRepository(userID, req)
	result, err := s.collection.InsertOne(ctx, repo)
	if err != nil {
		return nil, err
	}

	repo.ID = result.InsertedID.(primitive.ObjectID)
	return repo, nil
}

// GetRepositories retrieves repositories for a user with pagination and filtering
func (s *RepositoryService) GetRepositories(ctx context.Context, userID primitive.ObjectID, limit, offset int, statusFilter string) (*models.RepositoryListResponse, error) {
	filter := bson.M{"userId": userID}
	if statusFilter != "" && models.ValidStatus(statusFilter) {
		filter["status"] = statusFilter
	}

	// Count total repositories
	total, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Find repositories with pagination
	opts := options.Find().
		SetSkip(int64(offset)).
		SetLimit(int64(limit)).
		SetSort(bson.D{{"updatedAt", -1}}) // Sort by most recently updated

	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var repositories []models.Repository
	if err := cursor.All(ctx, &repositories); err != nil {
		return nil, err
	}

	return &models.RepositoryListResponse{
		Repositories: repositories,
		Total:        total,
	}, nil
}

// GetRepository retrieves a repository by ID with user ownership check
func (s *RepositoryService) GetRepository(ctx context.Context, userID primitive.ObjectID, repoID string) (*models.Repository, error) {
	objectID, err := primitive.ObjectIDFromHex(repoID)
	if err != nil {
		return nil, ErrRepositoryNotFound
	}

	var repo models.Repository
	filter := bson.M{"_id": objectID, "userId": userID}
	err = s.collection.FindOne(ctx, filter).Decode(&repo)
	if err == mongo.ErrNoDocuments {
		return nil, ErrRepositoryNotFound
	}
	if err != nil {
		return nil, err
	}

	return &repo, nil
}

// GetRepositoryByFullName retrieves a repository by full name for a user
func (s *RepositoryService) GetRepositoryByFullName(ctx context.Context, userID primitive.ObjectID, fullName string) (*models.Repository, error) {
	var repo models.Repository
	filter := bson.M{"userId": userID, "fullName": fullName}
	err := s.collection.FindOne(ctx, filter).Decode(&repo)
	if err == mongo.ErrNoDocuments {
		return nil, ErrRepositoryNotFound
	}
	if err != nil {
		return nil, err
	}

	return &repo, nil
}

// UpdateRepository updates repository information
func (s *RepositoryService) UpdateRepository(ctx context.Context, userID primitive.ObjectID, repoID string, req models.UpdateRepositoryRequest) (*models.Repository, error) {
	// First check if repository exists and user owns it
	repo, err := s.GetRepository(ctx, userID, repoID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	repo.Update(req)

	// Update in database
	objectID, _ := primitive.ObjectIDFromHex(repoID)
	filter := bson.M{"_id": objectID, "userId": userID}
	update := bson.M{"$set": bson.M{
		"name":            repo.Name,
		"description":     repo.Description,
		"primaryLanguage": repo.PrimaryLanguage,
		"updatedAt":       repo.UpdatedAt,
	}}

	_, err = s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

// DeleteRepository deletes a repository with user ownership check
func (s *RepositoryService) DeleteRepository(ctx context.Context, userID primitive.ObjectID, repoID string) error {
	objectID, err := primitive.ObjectIDFromHex(repoID)
	if err != nil {
		return ErrRepositoryNotFound
	}

	filter := bson.M{"_id": objectID, "userId": userID}
	result, err := s.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrRepositoryNotFound
	}

	return nil
}

// UpdateRepositoryStatus updates the status of a repository
func (s *RepositoryService) UpdateRepositoryStatus(ctx context.Context, userID primitive.ObjectID, repoID string, status string) error {
	if !models.ValidStatus(status) {
		return errors.New("invalid repository status")
	}

	objectID, err := primitive.ObjectIDFromHex(repoID)
	if err != nil {
		return ErrRepositoryNotFound
	}

	filter := bson.M{"_id": objectID, "userId": userID}
	update := bson.M{"$set": bson.M{
		"status":    status,
		"updatedAt": time.Now(),
	}}

	result, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrRepositoryNotFound
	}

	return nil
}

// UpdateRepositoryProgress updates the import progress of a repository
func (s *RepositoryService) UpdateRepositoryProgress(ctx context.Context, userID primitive.ObjectID, repoID string, progress int) error {
	objectID, err := primitive.ObjectIDFromHex(repoID)
	if err != nil {
		return ErrRepositoryNotFound
	}

	// Validate progress
	if progress < 0 {
		progress = 0
	} else if progress > 100 {
		progress = 100
	}

	// Determine status based on progress
	status := models.StatusImporting
	if progress == 0 {
		status = models.StatusPending
	} else if progress == 100 {
		status = models.StatusReady
	}

	filter := bson.M{"_id": objectID, "userId": userID}
	update := bson.M{"$set": bson.M{
		"importProgress": progress,
		"status":        status,
		"updatedAt":     time.Now(),
	}}

	result, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrRepositoryNotFound
	}

	return nil
}

// UpdateRepositoryStats updates repository statistics
func (s *RepositoryService) UpdateRepositoryStats(ctx context.Context, userID primitive.ObjectID, repoID string, stats *models.RepositoryStats) error {
	objectID, err := primitive.ObjectIDFromHex(repoID)
	if err != nil {
		return ErrRepositoryNotFound
	}

	filter := bson.M{"_id": objectID, "userId": userID}
	update := bson.M{"$set": bson.M{
		"stats":     stats,
		"updatedAt": time.Now(),
	}}

	result, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrRepositoryNotFound
	}

	return nil
}

// GetRepositoryStats retrieves detailed statistics for a repository
func (s *RepositoryService) GetRepositoryStats(ctx context.Context, userID primitive.ObjectID, repoID string) (map[string]interface{}, error) {
	repo, err := s.GetRepository(ctx, userID, repoID)
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"repositoryId": repo.ID.Hex(),
		"totalFiles":   0,
		"totalLines":   0,
		"languages":    make(map[string]int),
		"codeChunks":   0,
		"avgComplexity": 0.0,
	}

	if repo.Stats != nil {
		stats["totalFiles"] = repo.Stats.TotalFiles
		stats["totalLines"] = repo.Stats.TotalLines
		if repo.Stats.Languages != nil {
			stats["languages"] = repo.Stats.Languages
		}
		if repo.Stats.LastCommitDate != nil {
			stats["lastCommitDate"] = repo.Stats.LastCommitDate
		}
	}

	return stats, nil
}

// MarkRepositoryIndexed marks a repository as indexed
func (s *RepositoryService) MarkRepositoryIndexed(ctx context.Context, userID primitive.ObjectID, repoID string) error {
	objectID, err := primitive.ObjectIDFromHex(repoID)
	if err != nil {
		return ErrRepositoryNotFound
	}

	now := time.Now()
	filter := bson.M{"_id": objectID, "userId": userID}
	update := bson.M{"$set": bson.M{
		"indexedAt":    now,
		"lastSyncedAt": now,
		"updatedAt":    now,
	}}

	result, err := s.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrRepositoryNotFound
	}

	return nil
}