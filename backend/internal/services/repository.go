// ABOUTME: Repository service for MongoDB operations including CRUD and repository management
// ABOUTME: Handles repository creation, updates, deletion and statistics with user ownership validation

package services

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"github-analyzer/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrRepositoryNotFound = errors.New("repository not found")
	ErrRepositoryExists   = errors.New("repository already exists")
	ErrUnauthorized       = errors.New("unauthorized access to repository")
)

const RepositoryCollection = "repositories"

// RepositoryService provides repository-related operations
type RepositoryService struct {
	collection        *mongo.Collection
	githubService     *GitHubService
	userService       *UserService
	embeddingPipeline *EmbeddingPipeline
}

// NewRepositoryService creates a new repository service
func NewRepositoryService(db *mongo.Database, githubService *GitHubService, userService *UserService, embeddingPipeline *EmbeddingPipeline) *RepositoryService {
	return &RepositoryService{
		collection:        db.Collection(RepositoryCollection),
		githubService:     githubService,
		userService:       userService,
		embeddingPipeline: embeddingPipeline,
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
		SetSort(bson.D{{Key: "updatedAt", Value: -1}}) // Sort by most recently updated

	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := cursor.Close(ctx); closeErr != nil {
			// Log the error but don't override the main error
			// In production, you'd want proper logging here
			_ = closeErr // Explicitly ignore for now
		}
	}()

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
	switch progress {
	case 0:
		status = models.StatusPending
	case 100:
		status = models.StatusReady
	}

	filter := bson.M{"_id": objectID, "userId": userID}
	update := bson.M{"$set": bson.M{
		"importProgress": progress,
		"status":         status,
		"updatedAt":      time.Now(),
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
		"repositoryId":  repo.ID.Hex(),
		"totalFiles":    0,
		"totalLines":    0,
		"languages":     make(map[string]int),
		"codeChunks":    0,
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

// ImportRepositoryFromGitHub creates a repository by importing from GitHub
func (s *RepositoryService) ImportRepositoryFromGitHub(ctx context.Context, userID primitive.ObjectID, owner, repoName string) (*models.Repository, error) {
	// Get user and check GitHub connection
	user, err := s.userService.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user.GitHubToken == "" {
		return nil, errors.New("github account is not connected")
	}

	// Decrypt GitHub token
	accessToken, err := s.githubService.DecryptToken(user.GitHubToken)
	if err != nil {
		return nil, errors.New("failed to decrypt GitHub token")
	}

	// Validate repository exists and user has access
	githubRepo, err := s.githubService.ValidateRepository(ctx, accessToken, owner, repoName)
	if err != nil {
		return nil, err
	}

	fullName := owner + "/" + repoName

	// Check if repository already exists for this user
	existing, err := s.GetRepositoryByFullName(ctx, userID, fullName)
	if err != nil && err != ErrRepositoryNotFound {
		return nil, err
	}
	if existing != nil {
		return nil, ErrRepositoryExists
	}

	// Create repository request from GitHub data
	req := models.CreateRepositoryRequest{
		Name:            githubRepo.Name,
		Owner:           githubRepo.Owner,
		FullName:        githubRepo.FullName,
		Description:     githubRepo.Description,
		GitHubRepoID:    &githubRepo.ID,
		PrimaryLanguage: githubRepo.Language,
		IsPrivate:       githubRepo.Private,
	}

	// Create repository with pending status
	repo := models.NewRepository(userID, req)
	repo.Status = models.StatusPending
	repo.ImportProgress = 0

	// Insert into database
	result, err := s.collection.InsertOne(ctx, repo)
	if err != nil {
		return nil, err
	}

	repo.ID = result.InsertedID.(primitive.ObjectID)

	// Start async import process
	go s.processRepositoryImport(context.Background(), repo.ID, userID, accessToken, githubRepo)

	return repo, nil
}

// processRepositoryImport handles the async import process
func (s *RepositoryService) processRepositoryImport(ctx context.Context, repoID primitive.ObjectID, userID primitive.ObjectID, accessToken string, githubRepo *GitHubRepository) {
	repoIDStr := repoID.Hex()

	// Update status to importing
	if err := s.UpdateRepositoryStatus(ctx, userID, repoIDStr, models.StatusImporting); err != nil {
		return
	}
	if err := s.UpdateRepositoryProgress(ctx, userID, repoIDStr, 10); err != nil {
		return
	}

	// Fetch repository statistics from GitHub
	stats, err := s.githubService.GetRepositoryStatistics(ctx, accessToken, githubRepo.Owner, githubRepo.Name)
	if err != nil {
		// Mark as error but don't fail completely
		if updateErr := s.UpdateRepositoryStatus(ctx, userID, repoIDStr, models.StatusError); updateErr != nil {
			// Log error but continue
			_ = updateErr
		}
		return
	}

	if err := s.UpdateRepositoryProgress(ctx, userID, repoIDStr, 50); err != nil {
		return
	}

	// Convert GitHub stats to our repository stats format
	repoStats := &models.RepositoryStats{
		TotalFiles:     stats["total_files"].(int),
		TotalLines:     stats["total_lines"].(int),
		Languages:      stats["languages"].(map[string]int),
		LastCommitDate: stats["last_commit_date"].(*time.Time),
	}

	// Update repository with statistics
	err = s.UpdateRepositoryStats(ctx, userID, repoIDStr, repoStats)
	if err != nil {
		if updateErr := s.UpdateRepositoryStatus(ctx, userID, repoIDStr, models.StatusError); updateErr != nil {
			// Log error but continue
			_ = updateErr
		}
		return
	}

	if err := s.UpdateRepositoryProgress(ctx, userID, repoIDStr, 80); err != nil {
		return
	}

	// Mark repository as indexed and ready
	if err := s.MarkRepositoryIndexed(ctx, userID, repoIDStr); err != nil {
		return
	}
	if err := s.UpdateRepositoryProgress(ctx, userID, repoIDStr, 100); err != nil {
		// Final progress update failed, but repository is indexed
		_ = err
	}
	
	// Queue repository for embedding processing if pipeline is available
	if s.embeddingPipeline != nil {
		if err := s.embeddingPipeline.QueueRepository(ctx, repoID, 2); err != nil {
			// Log error but don't fail the import
			// The embedding can be triggered manually later
			log.Printf("Failed to queue repository %s for embedding processing: %v", repoID.Hex(), err)
		}
	}
}

// CreateRepositoryFromGitHub creates a repository with GitHub integration
func (s *RepositoryService) CreateRepositoryFromGitHub(ctx context.Context, userID primitive.ObjectID, githubURL string) (*models.Repository, error) {
	// Parse GitHub URL to extract owner and repo name
	owner, repoName, err := s.parseGitHubURL(githubURL)
	if err != nil {
		return nil, errors.New("invalid GitHub repository URL")
	}

	return s.ImportRepositoryFromGitHub(ctx, userID, owner, repoName)
}

// parseGitHubURL parses a GitHub URL and returns owner and repository name
func (s *RepositoryService) parseGitHubURL(url string) (string, string, error) {
	// Handle different GitHub URL formats
	// https://github.com/owner/repo
	// owner/repo

	// Remove .git suffix if present
	url = strings.TrimSuffix(url, ".git")

	// Handle https://github.com/owner/repo format
	if strings.Contains(url, "github.com/") {
		parts := strings.Split(url, "github.com/")
		if len(parts) != 2 {
			return "", "", errors.New("invalid GitHub URL format")
		}
		url = parts[1]
	}

	// Now we should have owner/repo format
	parts := strings.Split(url, "/")
	if len(parts) != 2 {
		return "", "", errors.New("invalid GitHub URL format")
	}

	return parts[0], parts[1], nil
}
