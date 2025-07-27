// ABOUTME: Repository service for MongoDB operations including CRUD and repository management
// ABOUTME: Handles repository creation, updates, deletion and statistics with user ownership validation

package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"acip.divkix.me/internal/config"
	"acip.divkix.me/internal/models"

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
	db                *mongo.Database
	githubService     *GitHubService
	userService       *UserService
	embeddingPipeline *EmbeddingPipeline
	config            *config.Config
}

// NewRepositoryService creates a new repository service
func NewRepositoryService(db *mongo.Database, githubService *GitHubService, userService *UserService, embeddingPipeline *EmbeddingPipeline, config *config.Config) *RepositoryService {
	return &RepositoryService{
		collection:        db.Collection(RepositoryCollection),
		db:                db,
		githubService:     githubService,
		userService:       userService,
		embeddingPipeline: embeddingPipeline,
		config:            config,
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
	
	// Auto-trigger import for GitHub repositories
	if repo.GitHubRepoID != nil && repo.Owner != "" {
		log.Printf("🚀 Auto-triggering import for GitHub repository: %s", repo.FullName)
		go s.autoTriggerGitHubImport(context.Background(), repo.ID, userID, repo.Owner, repo.Name)
	}
	
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

// DeleteRepository deletes a repository and all associated data (code chunks, vectors, chat sessions)
func (s *RepositoryService) DeleteRepository(ctx context.Context, userID primitive.ObjectID, repoID string) error {
	objectID, err := primitive.ObjectIDFromHex(repoID)
	if err != nil {
		return ErrRepositoryNotFound
	}

	// First verify the repository exists and belongs to the user
	filter := bson.M{"_id": objectID, "userId": userID}
	var repository models.Repository
	err = s.collection.FindOne(ctx, filter).Decode(&repository)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return ErrRepositoryNotFound
		}
		return fmt.Errorf("failed to find repository: %w", err)
	}

	log.Printf("Starting cascade deletion for repository %s (%s)", repository.Name, repository.ID.Hex())

	// Use MongoDB session for transaction support
	session, err := s.db.Client().StartSession()
	if err != nil {
		return fmt.Errorf("failed to start session: %w", err)
	}
	defer session.EndSession(ctx)

	// Perform deletion operations in transaction
	_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		var deletionCounts struct {
			vectors      int
			codeChunks   int64
			chatSessions int64
		}

		// Step 1: Delete vectors from Qdrant (before deleting code chunks)
		log.Printf("Deleting vectors from Qdrant for repository %s", objectID.Hex())
		vectorCount, err := s.embeddingPipeline.DeleteRepositoryVectors(sc, objectID)
		if err != nil {
			log.Printf("Warning: Failed to delete vectors from Qdrant: %v", err)
			// Continue with MongoDB cleanup even if Qdrant deletion fails
		}
		deletionCounts.vectors = vectorCount

		// Step 2: Delete code chunks from MongoDB
		log.Printf("Deleting code chunks from MongoDB for repository %s", objectID.Hex())
		codeChunksCollection := s.db.Collection("codechunks")
		chunkFilter := bson.M{"repositoryId": objectID}
		chunkResult, err := codeChunksCollection.DeleteMany(sc, chunkFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to delete code chunks: %w", err)
		}
		deletionCounts.codeChunks = chunkResult.DeletedCount

		// Step 3: Delete chat sessions from MongoDB
		log.Printf("Deleting chat sessions from MongoDB for repository %s", objectID.Hex())
		chatSessionsCollection := s.db.Collection("chat_sessions")
		sessionFilter := bson.M{"repositoryId": objectID}
		sessionResult, err := chatSessionsCollection.DeleteMany(sc, sessionFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to delete chat sessions: %w", err)
		}
		deletionCounts.chatSessions = sessionResult.DeletedCount

		// Step 4: Delete embedding jobs from MongoDB
		log.Printf("Deleting embedding jobs from MongoDB for repository %s", objectID.Hex())
		embeddingJobsCollection := s.db.Collection("embedding_jobs")
		jobFilter := bson.M{"repositoryId": objectID}
		jobResult, err := embeddingJobsCollection.DeleteMany(sc, jobFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to delete embedding jobs: %w", err)
		}

		// Step 5: Finally delete the repository itself
		log.Printf("Deleting repository record from MongoDB for repository %s", objectID.Hex())
		result, err := s.collection.DeleteOne(sc, filter)
		if err != nil {
			return nil, fmt.Errorf("failed to delete repository: %w", err)
		}

		if result.DeletedCount == 0 {
			return nil, ErrRepositoryNotFound
		}

		log.Printf("Successfully deleted repository %s: %d vectors, %d code chunks, %d chat sessions, %d embedding jobs", 
			objectID.Hex(), deletionCounts.vectors, deletionCounts.codeChunks, 
			deletionCounts.chatSessions, jobResult.DeletedCount)

		return nil, nil
	})

	if err != nil {
		return fmt.Errorf("transaction failed during repository deletion: %w", err)
	}

	log.Printf("Cascade deletion completed for repository %s", objectID.Hex())
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
		status = models.StatusQueuedEmbedding
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
		log.Printf("❌ Failed to get user %s: %v", userID.Hex(), err)
		return nil, err
	}

	log.Printf("🔍 User found: %s, GitHubToken length: %d, GitHubUsername: %s", 
		user.Email, len(user.GitHubToken), user.GitHubUsername)

	if user.GitHubToken == "" {
		log.Printf("❌ GitHub account is not connected for user %s", user.Email)
		return nil, errors.New("github account is not connected - please connect your GitHub account first")
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
		IsPrivate:       &githubRepo.Private,
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
	
	log.Printf("🔄 Starting processRepositoryImport for repository %s (%s/%s)", repoIDStr, githubRepo.Owner, githubRepo.Name)
	
	// Update status to importing
	log.Printf("📝 Updating repository status to importing...")
	if err := s.UpdateRepositoryStatus(ctx, userID, repoIDStr, models.StatusImporting); err != nil {
		log.Printf("❌ CRITICAL: Failed to update repository status to importing: %v", err)
		log.Printf("   Repository ID: %s, User ID: %s", repoIDStr, userID.Hex())
		return
	}
	log.Printf("✅ Repository status updated to importing")
	
	log.Printf("Starting repository import for %s/%s", githubRepo.Owner, githubRepo.Name)
	
	// Step 1: Fetch repository statistics (10% progress)
	if err := s.UpdateRepositoryProgress(ctx, userID, repoIDStr, 10); err != nil {
		log.Printf("Failed to update progress: %v", err)
	}
	
	stats, err := s.githubService.GetRepositoryStatistics(ctx, accessToken, githubRepo.Owner, githubRepo.Name)
	if err != nil {
		log.Printf("Failed to fetch repository statistics: %v", err)
		if updateErr := s.UpdateRepositoryStatus(ctx, userID, repoIDStr, models.StatusError); updateErr != nil {
			log.Printf("Failed to update status to error: %v", updateErr)
		}
		return
	}
	
	// Step 2: Fetch repository files (30% progress)
	if err := s.UpdateRepositoryProgress(ctx, userID, repoIDStr, 30); err != nil {
		log.Printf("Failed to update progress: %v", err)
	}
	
	log.Printf("Fetching files from %s/%s", githubRepo.Owner, githubRepo.Name)
	files, err := s.githubService.FetchRepositoryFiles(ctx, accessToken, githubRepo.Owner, githubRepo.Name)
	if err != nil {
		log.Printf("Failed to fetch repository files: %v", err)
		if updateErr := s.UpdateRepositoryStatus(ctx, userID, repoIDStr, models.StatusError); updateErr != nil {
			log.Printf("Failed to update status to error: %v", updateErr)
		}
		return
	}
	
	log.Printf("Fetched %d files from %s/%s", len(files), githubRepo.Owner, githubRepo.Name)
	
	// Step 3: Process and chunk files (50% progress) 
	if err := s.UpdateRepositoryProgress(ctx, userID, repoIDStr, 50); err != nil {
		log.Printf("Failed to update progress: %v", err)
	}
	
	processor := NewCodeProcessor(s.config)
	chunks, err := processor.ProcessAndChunkFiles(files, repoID)
	if err != nil {
		log.Printf("Failed to process and chunk files: %v", err)
		if updateErr := s.UpdateRepositoryStatus(ctx, userID, repoIDStr, models.StatusError); updateErr != nil {
			log.Printf("Failed to update status to error: %v", updateErr)
		}
		return
	}
	
	log.Printf("Created %d code chunks from %d files", len(chunks), len(files))
	
	// Step 4: Store code chunks in MongoDB (70% progress)
	if err := s.UpdateRepositoryProgress(ctx, userID, repoIDStr, 70); err != nil {
		log.Printf("Failed to update progress: %v", err)
	}
	
	if err := s.storeCodeChunks(ctx, chunks); err != nil {
		log.Printf("Failed to store code chunks: %v", err)
		if updateErr := s.UpdateRepositoryStatus(ctx, userID, repoIDStr, models.StatusError); updateErr != nil {
			log.Printf("Failed to update status to error: %v", updateErr)
		}
		return
	}
	
	log.Printf("Stored %d code chunks in MongoDB", len(chunks))
	
	// Step 5: Update repository statistics (85% progress)
	if err := s.UpdateRepositoryProgress(ctx, userID, repoIDStr, 85); err != nil {
		log.Printf("Failed to update progress: %v", err)
	}
	
	// Convert GitHub stats to our repository stats format with actual file data
	repoStats := &models.RepositoryStats{
		TotalFiles:     len(files),
		TotalLines:     s.calculateTotalLines(files),
		Languages:      stats["languages"].(map[string]int),
		LastCommitDate: stats["last_commit_date"].(*time.Time),
	}
	
	// Update repository with statistics
	err = s.UpdateRepositoryStats(ctx, userID, repoIDStr, repoStats)
	if err != nil {
		log.Printf("Failed to update repository stats: %v", err)
		if updateErr := s.UpdateRepositoryStatus(ctx, userID, repoIDStr, models.StatusError); updateErr != nil {
			log.Printf("Failed to update status to error: %v", updateErr)
		}
		return
	}
	
	// Step 6: Complete import and queue for embedding (100% progress)
	if err := s.UpdateRepositoryProgress(ctx, userID, repoIDStr, 95); err != nil {
		log.Printf("Failed to update progress: %v", err)
	}
	
	// Complete import - this will set status to queued-embedding
	if err := s.UpdateRepositoryProgress(ctx, userID, repoIDStr, 100); err != nil {
		log.Printf("Failed to update final progress: %v", err)
	}
	
	log.Printf("Successfully completed import for repository %s/%s", githubRepo.Owner, githubRepo.Name)
	
	// Queue repository for embedding processing if pipeline is available
	if s.embeddingPipeline != nil {
		if err := s.embeddingPipeline.QueueRepository(ctx, repoID, 2); err != nil {
			log.Printf("Failed to queue repository %s for embedding processing: %v", repoID.Hex(), err)
		} else {
			log.Printf("Queued repository %s for embedding processing", repoID.Hex())
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

// storeCodeChunks stores code chunks in MongoDB with enhanced batch processing and error handling
func (s *RepositoryService) storeCodeChunks(ctx context.Context, chunks []*models.CodeChunk) error {
	if len(chunks) == 0 {
		log.Printf("No chunks to store")
		return nil
	}
	
	collection := s.collection.Database().Collection("codechunks")
	
	// Process chunks in batches of 100 for better performance
	const batchSize = 100
	var totalInserted int
	var failedBatches []string
	
	log.Printf("Storing %d code chunks in %d batches", len(chunks), (len(chunks)+batchSize-1)/batchSize)
	
	for i := 0; i < len(chunks); i += batchSize {
		end := i + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}
		
		batch := chunks[i:end]
		batchNum := (i / batchSize) + 1
		
		// Convert to []interface{} for bulk insert
		documents := make([]interface{}, len(batch))
		for j, chunk := range batch {
			documents[j] = chunk
		}
		
		// Insert batch with enhanced error handling and UTF-8 validation
		var insertErr error
		var successfulInserts int
		
		for attempt := 1; attempt <= 3; attempt++ {
			_, err := collection.InsertMany(ctx, documents, options.InsertMany().SetOrdered(false))
			if err == nil {
				totalInserted += len(documents)
				log.Printf("Successfully inserted batch %d: %d chunks", batchNum, len(documents))
				break
			}

			// Handle bulk write exceptions with detailed error analysis
			if we, ok := err.(mongo.BulkWriteException); ok {
				successfulInserts = len(documents) - len(we.WriteErrors)
				totalInserted += successfulInserts
				
				// Analyze error types
				duplicateErrors := 0
				utf8Errors := 0
				otherErrors := 0
				
				for _, writeErr := range we.WriteErrors {
					switch writeErr.Code {
					case 11000: // duplicate key
						duplicateErrors++
					default:
						if strings.Contains(writeErr.Message, "invalid UTF-8") || 
						   strings.Contains(writeErr.Message, "text contains invalid UTF-8") {
							utf8Errors++
						} else {
							otherErrors++
						}
					}
				}
				
				if duplicateErrors > 0 {
					log.Printf("Batch %d: %d duplicates skipped", batchNum, duplicateErrors)
				}
				if utf8Errors > 0 {
					log.Printf("Batch %d: %d UTF-8 encoding errors", batchNum, utf8Errors)
				}
				if otherErrors > 0 {
					log.Printf("Batch %d: %d other errors", batchNum, otherErrors)
				}
				
				log.Printf("Batch %d partial success: %d/%d chunks inserted", 
					batchNum, successfulInserts, len(documents))
				
				// If we have some successful inserts, consider this batch partially successful
				if successfulInserts > 0 {
					break
				}
			}
			
			insertErr = err
			if attempt < 3 {
				backoff := time.Duration(attempt) * time.Second
				log.Printf("Failed to insert batch %d (attempt %d/3), retrying in %v: %v", 
					batchNum, attempt, backoff, err)
				time.Sleep(backoff)
			} else {
				log.Printf("Failed to insert batch %d after 3 attempts: %v", batchNum, err)
				
				// Only mark as completely failed if no chunks were inserted
				if successfulInserts == 0 {
					failedBatches = append(failedBatches, fmt.Sprintf("batch-%d", batchNum))
				}
			}
		}
		
		// If we still have an error after retries, continue with next batch but log it
		if insertErr != nil {
			log.Printf("Skipping failed batch %d, continuing with remaining batches", batchNum)
		}
	}
	
	log.Printf("Chunk storage complete: %d/%d chunks inserted successfully", totalInserted, len(chunks))
	
	// Calculate success rate
	successRate := float64(totalInserted) / float64(len(chunks))
	
	if len(failedBatches) > 0 {
		log.Printf("Warning: %d batches had failures, but %d/%d chunks were successfully inserted (%.1f%% success rate)", 
			len(failedBatches), totalInserted, len(chunks), successRate*100)
		
		// Only return error if success rate is below 50%
		if successRate < 0.5 {
			return fmt.Errorf("chunk storage failed: only %d/%d chunks inserted (%.1f%% success rate)", 
				totalInserted, len(chunks), successRate*100)
		}
		
		// Log warning but continue processing
		log.Printf("Proceeding with import despite %d failed batches due to acceptable success rate (%.1f%%)", 
			len(failedBatches), successRate*100)
	}
	
	return nil
}

// calculateTotalLines calculates total lines from repository files
func (s *RepositoryService) calculateTotalLines(files []*models.RepositoryFile) int {
	totalLines := 0
	for _, file := range files {
		if file.IsValidForProcessing() {
			totalLines += file.GetLineCount()
		}
	}
	return totalLines
}

// autoTriggerGitHubImport automatically triggers import for newly created GitHub repositories
func (s *RepositoryService) autoTriggerGitHubImport(ctx context.Context, repoID primitive.ObjectID, userID primitive.ObjectID, owner, repoName string) {
	log.Printf("🔄 Starting auto-import for repository %s (%s/%s)", repoID.Hex(), owner, repoName)
	
	// Get user and check GitHub connection
	user, err := s.userService.GetByID(ctx, userID)
	if err != nil {
		log.Printf("❌ Failed to get user %s for auto-import: %v", userID.Hex(), err)
		if errStatus := s.UpdateRepositoryStatus(ctx, userID, repoID.Hex(), models.StatusError); errStatus != nil {
			log.Printf("Failed to set repository status to error: %v", errStatus)
		}
		return
	}

	log.Printf("🔍 Auto-import user found: %s, GitHubToken length: %d, GitHubUsername: %s", 
		user.Email, len(user.GitHubToken), user.GitHubUsername)

	if user.GitHubToken == "" {
		log.Printf("❌ GitHub account is not connected for user %s during auto-import", user.Email)
		if errStatus := s.UpdateRepositoryStatus(ctx, userID, repoID.Hex(), models.StatusError); errStatus != nil {
			log.Printf("Failed to set repository status to error: %v", errStatus)
		}
		return
	}

	// Decrypt GitHub token
	accessToken, err := s.githubService.DecryptToken(user.GitHubToken)
	if err != nil {
		log.Printf("❌ Failed to decrypt GitHub token for auto-import: %v", err)
		if errStatus := s.UpdateRepositoryStatus(ctx, userID, repoID.Hex(), models.StatusError); errStatus != nil {
			log.Printf("Failed to set repository status to error: %v", errStatus)
		}
		return
	}

	// Validate repository still exists and is accessible
	githubRepo, err := s.githubService.ValidateRepository(ctx, accessToken, owner, repoName)
	if err != nil {
		log.Printf("❌ Repository validation failed during auto-import: %v", err)
		if errStatus := s.UpdateRepositoryStatus(ctx, userID, repoID.Hex(), models.StatusError); errStatus != nil {
			log.Printf("Failed to set repository status to error: %v", errStatus)
		}
		return
	}

	log.Printf("✅ GitHub repository validated for auto-import: %s", githubRepo.FullName)

	// Start the import process
	s.processRepositoryImport(ctx, repoID, userID, accessToken, githubRepo)
}

// TriggerRepositoryImport manually triggers repository import for repositories stuck in pending/error status
func (s *RepositoryService) TriggerRepositoryImport(ctx context.Context, userID primitive.ObjectID, repoID string) error {
	// Get repository details
	repo, err := s.GetRepository(ctx, userID, repoID)
	if err != nil {
		return err
	}

	// Get user and check GitHub connection
	user, err := s.userService.GetByID(ctx, userID)
	if err != nil {
		log.Printf("❌ Failed to get user %s: %v", userID.Hex(), err)
		return err
	}

	log.Printf("🔍 User found: %s, GitHubToken length: %d, GitHubUsername: %s", 
		user.Email, len(user.GitHubToken), user.GitHubUsername)

	if user.GitHubToken == "" {
		log.Printf("❌ GitHub account is not connected for user %s", user.Email)
		return errors.New("github account is not connected - please connect your GitHub account first")
	}

	// Decrypt GitHub token
	accessToken, err := s.githubService.DecryptToken(user.GitHubToken)
	if err != nil {
		return errors.New("failed to decrypt GitHub token")
	}

	// Parse owner and repo from full name
	parts := strings.Split(repo.FullName, "/")
	if len(parts) != 2 {
		return errors.New("invalid repository full name format")
	}
	owner, repoName := parts[0], parts[1]

	// Validate repository still exists and is accessible
	githubRepo, err := s.githubService.ValidateRepository(ctx, accessToken, owner, repoName)
	if err != nil {
		return fmt.Errorf("repository validation failed: %w", err)
	}

	log.Printf("🚀 Manually triggering import for repository %s (%s)", repo.ID.Hex(), repo.FullName)

	// Reset repository status and progress
	if err := s.UpdateRepositoryStatus(ctx, userID, repoID, models.StatusImporting); err != nil {
		return fmt.Errorf("failed to update repository status: %w", err)
	}

	if err := s.UpdateRepositoryProgress(ctx, userID, repoID, 0); err != nil {
		return fmt.Errorf("failed to reset repository progress: %w", err)
	}

	// Start async import process
	go s.processRepositoryImport(context.Background(), repo.ID, userID, accessToken, githubRepo)

	return nil
}

