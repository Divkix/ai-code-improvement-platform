// ABOUTME: Embedding processing service for generating and storing vector embeddings
// ABOUTME: Handles batch processing, progress tracking, and error recovery for code chunks
package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github-analyzer/internal/config"
	"github-analyzer/internal/database"
	"github-analyzer/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type EmbeddingService struct {
	provider EmbeddingProvider
	qdrantClient  *database.Qdrant
	mongoDB       *database.MongoDB
	config        *config.Config
}

type EmbeddingStatus string

const (
	EmbeddingStatusPending    EmbeddingStatus = "pending"
	EmbeddingStatusProcessing EmbeddingStatus = "processing"
	EmbeddingStatusCompleted  EmbeddingStatus = "completed"
	EmbeddingStatusFailed     EmbeddingStatus = "failed"
)

type ProcessingStats struct {
	TotalChunks      int
	ProcessedChunks  int
	FailedChunks     int
	SkippedChunks    int
	StartedAt        time.Time
	CompletedAt      *time.Time
	EstimatedTimeRemaining *time.Duration
}

func NewEmbeddingService(provider EmbeddingProvider, qdrant *database.Qdrant, mongoDB *database.MongoDB, config *config.Config) *EmbeddingService {
	return &EmbeddingService{
		provider: provider,
		qdrantClient:  qdrant,
		mongoDB:       mongoDB,
		config:        config,
	}
}

func (es *EmbeddingService) InitializeCollection(ctx context.Context) error {
	collectionName := es.config.Database.QdrantCollectionName
	vectorDim := es.config.Database.VectorDimension

	// Check if collection already exists
	exists, err := es.qdrantClient.CollectionExists(ctx, collectionName)
	if err != nil {
		return fmt.Errorf("failed to check collection existence: %w", err)
	}

	if !exists {
		log.Printf("Creating Qdrant collection: %s with dimension: %d", collectionName, vectorDim)
		if err := es.qdrantClient.CreateCollection(ctx, collectionName, vectorDim); err != nil {
			return fmt.Errorf("failed to create collection: %w", err)
		}
	}

	return nil
}

func (es *EmbeddingService) ProcessRepository(ctx context.Context, repositoryID primitive.ObjectID) error {
	log.Printf("Starting embedding processing for repository: %s", repositoryID.Hex())

	// Ensure collection exists
	if err := es.InitializeCollection(ctx); err != nil {
		return fmt.Errorf("failed to initialize collection: %w", err)
	}

	// Update repository status to processing
	if err := es.updateRepositoryEmbeddingStatus(ctx, repositoryID, EmbeddingStatusProcessing); err != nil {
		return fmt.Errorf("failed to update repository status: %w", err)
	}

	// Get unprocessed chunks for the repository
	chunks, err := es.getUnprocessedChunks(ctx, repositoryID)
	if err != nil {
		if updateErr := es.updateRepositoryEmbeddingStatus(ctx, repositoryID, EmbeddingStatusFailed); updateErr != nil {
			log.Printf("Failed to update repository status to failed: %v", updateErr)
		}
		return fmt.Errorf("failed to get unprocessed chunks: %w", err)
	}

	if len(chunks) == 0 {
		log.Printf("No unprocessed chunks found for repository: %s", repositoryID.Hex())
		if err := es.updateRepositoryEmbeddingStatus(ctx, repositoryID, EmbeddingStatusCompleted); err != nil {
			log.Printf("Failed to update repository status to completed: %v", err)
		}
		return nil
	}

	stats := ProcessingStats{
		TotalChunks:     len(chunks),
		ProcessedChunks: 0,
		FailedChunks:    0,
		StartedAt:       time.Now(),
	}

	const batchSize = 50 // Process 50 chunks at a time
	for i := 0; i < len(chunks); i += batchSize {
		end := i + batchSize
		if end > len(chunks) {
			end = len(chunks)
		}

		batch := chunks[i:end]
		if err := es.processBatch(ctx, batch, &stats); err != nil {
			log.Printf("Failed to process batch %d-%d for repository %s: %v", i, end-1, repositoryID.Hex(), err)
			// Continue with next batch instead of failing entirely
		}

		// Update progress
		progress := (stats.ProcessedChunks + stats.FailedChunks) * 100 / stats.TotalChunks
		if err := es.updateRepositoryProgress(ctx, repositoryID, progress); err != nil {
			log.Printf("Failed to update repository progress: %v", err)
		}

		log.Printf("Processed %d/%d chunks for repository %s", stats.ProcessedChunks+stats.FailedChunks, stats.TotalChunks, repositoryID.Hex())
	}

	completedAt := time.Now()
	stats.CompletedAt = &completedAt

	// Determine final status based on results
	finalStatus := EmbeddingStatusCompleted
	if stats.FailedChunks > 0 && stats.ProcessedChunks == 0 {
		finalStatus = EmbeddingStatusFailed
	}

	if err := es.updateRepositoryEmbeddingStatus(ctx, repositoryID, finalStatus); err != nil {
		log.Printf("Failed to update final repository status: %v", err)
	}

	log.Printf("Completed embedding processing for repository %s. Processed: %d, Failed: %d, Total time: %v", 
		repositoryID.Hex(), stats.ProcessedChunks, stats.FailedChunks, completedAt.Sub(stats.StartedAt))

	return nil
}

func (es *EmbeddingService) ProcessCodeChunk(ctx context.Context, chunk *models.CodeChunk) error {
	if chunk.IsIndexed() {
		return nil // Already processed
	}

	// Generate embedding for the chunk content
	embeddings, err := es.provider.GenerateEmbeddings(ctx, []string{chunk.Content})
	if err != nil {
		return fmt.Errorf("failed to generate embedding: %w", err)
	}

	if len(embeddings) == 0 {
		return fmt.Errorf("no embeddings returned")
	}

	// Create vector point for Qdrant
	vectorPoint := database.VectorPoint{
		ID:     chunk.ID.Hex(),
		Vector: embeddings[0],
		Payload: map[string]any{
			"repositoryId": chunk.RepositoryID.Hex(),
			"filePath":     chunk.FilePath,
			"fileName":     chunk.FileName,
			"language":     chunk.Language,
			"startLine":    chunk.StartLine,
			"endLine":      chunk.EndLine,
			"functions":    chunk.Metadata.Functions,
			"classes":      chunk.Metadata.Classes,
			"contentHash":  chunk.ContentHash,
		},
	}

	// Store in Qdrant
	if err := es.qdrantClient.UpsertPoints(ctx, es.config.Database.QdrantCollectionName, []database.VectorPoint{vectorPoint}); err != nil {
		return fmt.Errorf("failed to store vector in Qdrant: %w", err)
	}

	// Update chunk with vector ID
	chunk.SetVectorID(chunk.ID.Hex())

	// Update chunk in MongoDB
	if err := es.updateChunkVectorID(ctx, chunk); err != nil {
		return fmt.Errorf("failed to update chunk vector ID: %w", err)
	}

	return nil
}

func (es *EmbeddingService) processBatch(ctx context.Context, chunks []*models.CodeChunk, stats *ProcessingStats) error {
	if len(chunks) == 0 {
		return nil
	}

	// Extract content from chunks
	contents := make([]string, len(chunks))
	for i, chunk := range chunks {
		contents[i] = chunk.Content
	}

	// Generate embeddings for the batch
	embeddings, err := es.provider.GenerateEmbeddings(ctx, contents)
	if err != nil {
		stats.FailedChunks += len(chunks)
		return fmt.Errorf("failed to generate embeddings for batch: %w", err)
	}

	if len(embeddings) != len(chunks) {
		stats.FailedChunks += len(chunks)
		return fmt.Errorf("embedding count mismatch: expected %d, got %d", len(chunks), len(embeddings))
	}

	// Create vector points
	vectorPoints := make([]database.VectorPoint, len(chunks))
	for i, chunk := range chunks {
		vectorPoints[i] = database.VectorPoint{
			ID:     chunk.ID.Hex(),
			Vector: embeddings[i],
			Payload: map[string]any{
				"repositoryId": chunk.RepositoryID.Hex(),
				"filePath":     chunk.FilePath,
				"fileName":     chunk.FileName,
				"language":     chunk.Language,
				"startLine":    chunk.StartLine,
				"endLine":      chunk.EndLine,
				"functions":    chunk.Metadata.Functions,
				"classes":      chunk.Metadata.Classes,
				"contentHash":  chunk.ContentHash,
			},
		}
	}

	// Store vectors in Qdrant
	if err := es.qdrantClient.UpsertPoints(ctx, es.config.Database.QdrantCollectionName, vectorPoints); err != nil {
		stats.FailedChunks += len(chunks)
		return fmt.Errorf("failed to store vectors in Qdrant: %w", err)
	}

	// Update chunks with vector IDs in MongoDB
	for _, chunk := range chunks {
		chunk.SetVectorID(chunk.ID.Hex())
		if err := es.updateChunkVectorID(ctx, chunk); err != nil {
			log.Printf("Failed to update chunk vector ID for %s: %v", chunk.ID.Hex(), err)
			stats.FailedChunks++
		} else {
			stats.ProcessedChunks++
		}
	}

	return nil
}

func (es *EmbeddingService) getUnprocessedChunks(ctx context.Context, repositoryID primitive.ObjectID) ([]*models.CodeChunk, error) {
	collection := es.mongoDB.Database().Collection("codechunks")

	// Find chunks without vector IDs
	filter := bson.M{
		"repositoryId": repositoryID,
		"$or": []bson.M{
			{"vectorId": bson.M{"$exists": false}},
			{"vectorId": ""},
		},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Printf("Failed to close cursor: %v", err)
		}
	}()

	var chunks []*models.CodeChunk
	if err := cursor.All(ctx, &chunks); err != nil {
		return nil, err
	}

	return chunks, nil
}

func (es *EmbeddingService) updateChunkVectorID(ctx context.Context, chunk *models.CodeChunk) error {
	collection := es.mongoDB.Database().Collection("codechunks")

	filter := bson.M{"_id": chunk.ID}
	update := bson.M{
		"$set": bson.M{
			"vectorId":  chunk.VectorID,
			"updatedAt": chunk.UpdatedAt,
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

func (es *EmbeddingService) updateRepositoryEmbeddingStatus(ctx context.Context, repositoryID primitive.ObjectID, status EmbeddingStatus) error {
	collection := es.mongoDB.Database().Collection("repositories")

	filter := bson.M{"_id": repositoryID}
	update := bson.M{
		"$set": bson.M{
			"embeddingStatus": string(status),
			"updatedAt":       time.Now(),
		},
	}

	if status == EmbeddingStatusCompleted {
		update["$set"].(bson.M)["lastEmbeddedAt"] = time.Now()
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

func (es *EmbeddingService) updateRepositoryProgress(ctx context.Context, repositoryID primitive.ObjectID, progress int) error {
	collection := es.mongoDB.Database().Collection("repositories")

	filter := bson.M{"_id": repositoryID}
	update := bson.M{
		"$set": bson.M{
			"embeddingProgress": progress,
			"updatedAt":         time.Now(),
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

func (es *EmbeddingService) GetEmbeddingStatus(ctx context.Context, repositoryID primitive.ObjectID) (EmbeddingStatus, int, error) {
	collection := es.mongoDB.Database().Collection("repositories")

	var result struct {
		EmbeddingStatus   string `bson:"embeddingStatus"`
		EmbeddingProgress int    `bson:"embeddingProgress"`
	}

	err := collection.FindOne(ctx, bson.M{"_id": repositoryID}).Decode(&result)
	if err == mongo.ErrNoDocuments {
		return EmbeddingStatusPending, 0, nil
	}
	if err != nil {
		return EmbeddingStatusPending, 0, err
	}

	status := EmbeddingStatus(result.EmbeddingStatus)
	if status == "" {
		status = EmbeddingStatusPending
	}

	return status, result.EmbeddingProgress, nil
}

func (es *EmbeddingService) GetProcessingStats(ctx context.Context, repositoryID primitive.ObjectID) (*ProcessingStats, error) {
	// Get total chunks count
	chunkCollection := es.mongoDB.Database().Collection("codechunks")
	totalChunks, err := chunkCollection.CountDocuments(ctx, bson.M{"repositoryId": repositoryID})
	if err != nil {
		return nil, err
	}

	// Get processed chunks count
	processedChunks, err := chunkCollection.CountDocuments(ctx, bson.M{
		"repositoryId": repositoryID,
		"vectorId":     bson.M{"$ne": ""},
	})
	if err != nil {
		return nil, err
	}

	// Get repository info for timestamps
	repoCollection := es.mongoDB.Database().Collection("repositories")
	var repo struct {
		EmbeddingStatus string     `bson:"embeddingStatus"`
		CreatedAt       time.Time  `bson:"createdAt"`
		UpdatedAt       time.Time  `bson:"updatedAt"`
		LastEmbeddedAt  *time.Time `bson:"lastEmbeddedAt"`
	}

	err = repoCollection.FindOne(ctx, bson.M{"_id": repositoryID}).Decode(&repo)
	if err != nil {
		return nil, err
	}

	stats := &ProcessingStats{
		TotalChunks:     int(totalChunks),
		ProcessedChunks: int(processedChunks),
		StartedAt:       repo.CreatedAt,
	}

	if repo.LastEmbeddedAt != nil {
		stats.CompletedAt = repo.LastEmbeddedAt
	}

	// Calculate estimated time remaining if processing
	if repo.EmbeddingStatus == string(EmbeddingStatusProcessing) && processedChunks > 0 {
		elapsed := time.Since(stats.StartedAt)
		avgTimePerChunk := elapsed / time.Duration(processedChunks)
		remaining := avgTimePerChunk * time.Duration(totalChunks-processedChunks)
		stats.EstimatedTimeRemaining = &remaining
	}

	return stats, nil
}