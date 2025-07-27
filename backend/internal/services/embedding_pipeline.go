// ABOUTME: Background embedding pipeline for processing repositories asynchronously
// ABOUTME: Handles job queue management, batch processing, and progress tracking
package services

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github-analyzer/internal/config"
	"github-analyzer/internal/database"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// EmbeddingJob represents a background embedding job
type EmbeddingJob struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	RepositoryID primitive.ObjectID `bson:"repositoryId" json:"repositoryId"`
	Status       string             `bson:"status" json:"status"` // pending, processing, completed, failed
	Priority     int                `bson:"priority" json:"priority"` // 1=high, 2=normal, 3=low
	Attempts     int                `bson:"attempts" json:"attempts"`
	MaxAttempts  int                `bson:"maxAttempts" json:"maxAttempts"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	StartedAt    *time.Time         `bson:"startedAt,omitempty" json:"startedAt,omitempty"`
	CompletedAt  *time.Time         `bson:"completedAt,omitempty" json:"completedAt,omitempty"`
	ErrorMessage string             `bson:"errorMessage,omitempty" json:"errorMessage,omitempty"`
}

// EmbeddingPipeline manages background embedding processing
type EmbeddingPipeline struct {
	embeddingService *EmbeddingService
	mongoDB          *database.MongoDB
	config           *config.Config
	
	// Job processing
	jobQueue     chan EmbeddingJob
	workers      int
	workerPool   sync.WaitGroup
	shutdownChan chan struct{}
	running      bool
	mu           sync.RWMutex
}

// NewEmbeddingPipeline creates a new background embedding pipeline
func NewEmbeddingPipeline(embeddingService *EmbeddingService, mongoDB *database.MongoDB, config *config.Config) *EmbeddingPipeline {
	return &EmbeddingPipeline{
		embeddingService: embeddingService,
		mongoDB:          mongoDB,
		config:           config,
		jobQueue:         make(chan EmbeddingJob, 100), // Buffer for 100 jobs
		workers:          config.CodeProcessing.EmbeddingWorkersNum,
		shutdownChan:     make(chan struct{}),
		running:          false,
	}
}

// Start begins the background processing pipeline
func (ep *EmbeddingPipeline) Start(ctx context.Context) error {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	
	if ep.running {
		return fmt.Errorf("pipeline is already running")
	}
	
	ep.running = true
	log.Printf("Starting embedding pipeline with %d workers", ep.workers)
	
	// Start worker goroutines
	for i := 0; i < ep.workers; i++ {
		ep.workerPool.Add(1)
		go ep.worker(ctx, i)
	}
	
	// Start job scheduler
	go ep.jobScheduler(ctx)
	
	return nil
}

// Stop gracefully shuts down the pipeline
func (ep *EmbeddingPipeline) Stop() error {
	ep.mu.Lock()
	defer ep.mu.Unlock()
	
	if !ep.running {
		return nil
	}
	
	log.Println("Stopping embedding pipeline...")
	close(ep.shutdownChan)
	
	// Wait for all workers to finish
	ep.workerPool.Wait()
	ep.running = false
	
	log.Println("Embedding pipeline stopped")
	return nil
}

// QueueRepository adds a repository to the embedding queue
func (ep *EmbeddingPipeline) QueueRepository(ctx context.Context, repositoryID primitive.ObjectID, priority int) error {
	// Check if job already exists and is pending/processing
	collection := ep.mongoDB.Database().Collection("embedding_jobs")
	
	existingJob := collection.FindOne(ctx, bson.M{
		"repositoryId": repositoryID,
		"status": bson.M{"$in": []string{"pending", "processing"}},
	})
	
	if existingJob.Err() == nil {
		return fmt.Errorf("repository %s already has a pending embedding job", repositoryID.Hex())
	}
	
	// Create new job
	job := EmbeddingJob{
		RepositoryID: repositoryID,
		Status:       "pending",
		Priority:     priority,
		Attempts:     0,
		MaxAttempts:  3,
		CreatedAt:    time.Now(),
	}
	
	result, err := collection.InsertOne(ctx, job)
	if err != nil {
		return fmt.Errorf("failed to queue embedding job: %w", err)
	}
	
	job.ID = result.InsertedID.(primitive.ObjectID)
	log.Printf("Queued embedding job for repository %s with priority %d", repositoryID.Hex(), priority)
	
	return nil
}

// QueueAllRepositories queues all repositories that need embedding
func (ep *EmbeddingPipeline) QueueAllRepositories(ctx context.Context) error {
	// Find repositories that need embedding (no embedding status or failed status)
	repoCollection := ep.mongoDB.Database().Collection("repositories")
	
	filter := bson.M{
		"$or": []bson.M{
			{"embeddingStatus": bson.M{"$exists": false}},
			{"embeddingStatus": ""},
			{"embeddingStatus": "failed"},
			{"embeddingStatus": "pending"},
		},
	}
	
	cursor, err := repoCollection.Find(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to find repositories needing embedding: %w", err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Printf("Failed to close cursor: %v", err)
		}
	}()
	
	count := 0
	for cursor.Next(ctx) {
		var repo struct {
			ID primitive.ObjectID `bson:"_id"`
		}
		if err := cursor.Decode(&repo); err != nil {
			log.Printf("Failed to decode repository: %v", err)
			continue
		}
		
		if err := ep.QueueRepository(ctx, repo.ID, 2); err != nil {
			log.Printf("Failed to queue repository %s: %v", repo.ID.Hex(), err)
			continue
		}
		count++
	}
	
	log.Printf("Queued %d repositories for embedding", count)
	return nil
}

// GetJobStatus returns the status of embedding jobs
func (ep *EmbeddingPipeline) GetJobStatus(ctx context.Context, repositoryID *primitive.ObjectID) ([]EmbeddingJob, error) {
	collection := ep.mongoDB.Database().Collection("embedding_jobs")
	
	filter := bson.M{}
	if repositoryID != nil {
		filter["repositoryId"] = *repositoryID
	}
	
	cursor, err := collection.Find(ctx, filter, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get job status: %w", err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Printf("Failed to close cursor: %v", err)
		}
	}()
	
	var jobs []EmbeddingJob
	if err := cursor.All(ctx, &jobs); err != nil {
		return nil, fmt.Errorf("failed to decode jobs: %w", err)
	}
	
	return jobs, nil
}

// jobScheduler continuously schedules pending jobs to workers
func (ep *EmbeddingPipeline) jobScheduler(ctx context.Context) {
	ticker := time.NewTicker(3 * time.Second) // Check for new jobs every 3 seconds for faster responsiveness
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ep.shutdownChan:
			return
		case <-ticker.C:
			ep.schedulePendingJobs(ctx)
		}
	}
}

// schedulePendingJobs finds pending jobs and sends them to workers
func (ep *EmbeddingPipeline) schedulePendingJobs(ctx context.Context) {
	collection := ep.mongoDB.Database().Collection("embedding_jobs")
	
	// Find pending jobs ordered by priority (1=high, 2=normal, 3=low) then by creation time
	cursor, err := collection.Find(ctx, bson.M{"status": "pending"}, nil)
	if err != nil {
		log.Printf("Failed to find pending jobs: %v", err)
		return
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Printf("Failed to close cursor: %v", err)
		}
	}()
	
	for cursor.Next(ctx) {
		var job EmbeddingJob
		if err := cursor.Decode(&job); err != nil {
			log.Printf("Failed to decode job: %v", err)
			continue
		}
		
		// Try to send job to worker (non-blocking)
		select {
		case ep.jobQueue <- job:
			// Mark job as processing
			if err := ep.updateJobStatus(ctx, job.ID, "processing", ""); err != nil {
				log.Printf("Failed to update job status: %v", err)
			}
		default:
			// Queue is full, will try again on next iteration
			continue
		}
	}
}

// worker processes embedding jobs from the queue
func (ep *EmbeddingPipeline) worker(ctx context.Context, workerID int) {
	defer ep.workerPool.Done()
	
	log.Printf("Worker %d started", workerID)
	
	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %d stopped due to context cancellation", workerID)
			return
		case <-ep.shutdownChan:
			log.Printf("Worker %d stopped due to shutdown", workerID)
			return
		case job := <-ep.jobQueue:
			ep.processJob(ctx, job, workerID)
		}
	}
}

// processJob processes a single embedding job
func (ep *EmbeddingPipeline) processJob(ctx context.Context, job EmbeddingJob, workerID int) {
	log.Printf("Worker %d processing job %s for repository %s", workerID, job.ID.Hex(), job.RepositoryID.Hex())
	
	startTime := time.Now()
	if err := ep.updateJobStartTime(ctx, job.ID, startTime); err != nil {
		log.Printf("Failed to update job start time: %v", err)
	}
	
	// Process the repository
	err := ep.embeddingService.ProcessRepository(ctx, job.RepositoryID)
	
	if err != nil {
		log.Printf("Worker %d failed to process repository %s: %v", workerID, job.RepositoryID.Hex(), err)
		
		// Increment attempts and check if we should retry
		newAttempts := job.Attempts + 1
		if newAttempts >= job.MaxAttempts {
			// Max attempts reached, mark as failed
			if updateErr := ep.updateJobStatus(ctx, job.ID, "failed", err.Error()); updateErr != nil {
				log.Printf("Failed to update job status to failed: %v", updateErr)
			}
		} else {
			// Retry later
			if updateErr := ep.retryJob(ctx, job.ID, newAttempts); updateErr != nil {
				log.Printf("Failed to retry job: %v", updateErr)
			}
		}
		return
	}
	
	// Success
	completedAt := time.Now()
	if err := ep.updateJobCompleted(ctx, job.ID, completedAt); err != nil {
		log.Printf("Failed to update job completion: %v", err)
	}
	
	duration := completedAt.Sub(startTime)
	log.Printf("Worker %d completed job %s for repository %s in %v", 
		workerID, job.ID.Hex(), job.RepositoryID.Hex(), duration)
}

// updateJobStatus updates the status of a job
func (ep *EmbeddingPipeline) updateJobStatus(ctx context.Context, jobID primitive.ObjectID, status, errorMessage string) error {
	collection := ep.mongoDB.Database().Collection("embedding_jobs")
	
	update := bson.M{
		"$set": bson.M{
			"status": status,
		},
	}
	
	if errorMessage != "" {
		update["$set"].(bson.M)["errorMessage"] = errorMessage
	}
	
	_, err := collection.UpdateOne(ctx, bson.M{"_id": jobID}, update)
	return err
}

// updateJobStartTime updates the start time of a job
func (ep *EmbeddingPipeline) updateJobStartTime(ctx context.Context, jobID primitive.ObjectID, startTime time.Time) error {
	collection := ep.mongoDB.Database().Collection("embedding_jobs")
	
	update := bson.M{
		"$set": bson.M{
			"startedAt": startTime,
		},
	}
	
	_, err := collection.UpdateOne(ctx, bson.M{"_id": jobID}, update)
	return err
}

// updateJobCompleted marks a job as completed
func (ep *EmbeddingPipeline) updateJobCompleted(ctx context.Context, jobID primitive.ObjectID, completedAt time.Time) error {
	collection := ep.mongoDB.Database().Collection("embedding_jobs")
	
	update := bson.M{
		"$set": bson.M{
			"status":      "completed",
			"completedAt": completedAt,
		},
	}
	
	_, err := collection.UpdateOne(ctx, bson.M{"_id": jobID}, update)
	return err
}

// retryJob schedules a job for retry
func (ep *EmbeddingPipeline) retryJob(ctx context.Context, jobID primitive.ObjectID, newAttempts int) error {
	collection := ep.mongoDB.Database().Collection("embedding_jobs")
	
	update := bson.M{
		"$set": bson.M{
			"status":   "pending",
			"attempts": newAttempts,
		},
		"$unset": bson.M{
			"startedAt":    "",
			"errorMessage": "",
		},
	}
	
	_, err := collection.UpdateOne(ctx, bson.M{"_id": jobID}, update)
	return err
}

// GetPipelineStats returns statistics about the embedding pipeline
func (ep *EmbeddingPipeline) GetPipelineStats(ctx context.Context) (map[string]interface{}, error) {
	collection := ep.mongoDB.Database().Collection("embedding_jobs")
	
	// Aggregate job status counts
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$status"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}
	
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get pipeline stats: %w", err)
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Printf("Failed to close cursor: %v", err)
		}
	}()
	
	stats := map[string]interface{}{
		"pending":    0,
		"processing": 0,
		"completed":  0,
		"failed":     0,
		"running":    ep.IsRunning(),
		"workers":    ep.workers,
	}
	
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err := cursor.Decode(&result); err != nil {
			continue
		}
		stats[result.ID] = result.Count
	}
	
	return stats, nil
}

// IsRunning returns whether the pipeline is currently running
func (ep *EmbeddingPipeline) IsRunning() bool {
	ep.mu.RLock()
	defer ep.mu.RUnlock()
	return ep.running
}

// DeleteRepositoryVectors deletes all vector embeddings for a repository from Qdrant
func (ep *EmbeddingPipeline) DeleteRepositoryVectors(ctx context.Context, repositoryID primitive.ObjectID) (int, error) {
	// Get all code chunks for this repository to collect vector IDs
	codeChunksCollection := ep.mongoDB.Database().Collection("codechunks")
	
	filter := bson.M{"repositoryId": repositoryID, "vectorId": bson.M{"$ne": ""}}
	cursor, err := codeChunksCollection.Find(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to find code chunks: %w", err)
	}
	defer func() {
		if closeErr := cursor.Close(ctx); closeErr != nil {
			log.Printf("Failed to close cursor: %v", closeErr)
		}
	}()
	
	var vectorIDs []string
	for cursor.Next(ctx) {
		var chunk struct {
			VectorID string `bson:"vectorId"`
		}
		if err := cursor.Decode(&chunk); err != nil {
			log.Printf("Failed to decode chunk: %v", err)
			continue
		}
		if chunk.VectorID != "" {
			vectorIDs = append(vectorIDs, chunk.VectorID)
		}
	}
	
	if err := cursor.Err(); err != nil {
		return 0, fmt.Errorf("cursor error: %w", err)
	}
	
	if len(vectorIDs) == 0 {
		log.Printf("No vectors found for repository %s", repositoryID.Hex())
		return 0, nil
	}
	
	// Delete vectors from Qdrant
	collectionName := ep.config.Database.QdrantCollectionName
	if err := ep.embeddingService.qdrantClient.DeletePoints(ctx, collectionName, vectorIDs); err != nil {
		return 0, fmt.Errorf("failed to delete vectors from Qdrant: %w", err)
	}
	
	log.Printf("Deleted %d vectors from Qdrant for repository %s", len(vectorIDs), repositoryID.Hex())
	return len(vectorIDs), nil
}