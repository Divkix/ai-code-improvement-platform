// ABOUTME: MongoDB database connection and client management
// ABOUTME: Provides connection setup, health checks, and client access
package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	client *mongo.Client
	dbName string
}

func NewMongoDB(uri, dbName string) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	// Test the connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &MongoDB{
		client: client,
		dbName: dbName,
	}, nil
}

func (m *MongoDB) Client() *mongo.Client {
	return m.client
}

func (m *MongoDB) Database() *mongo.Database {
	return m.client.Database(m.dbName)
}

func (m *MongoDB) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return m.client.Disconnect(ctx)
}

func (m *MongoDB) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.client.Ping(ctx, nil)
}

// EnsureIndexes creates all required indexes for the application
func (m *MongoDB) EnsureIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db := m.Database()

	// Code chunks text search index
	codeChunksCollection := db.Collection("codechunks")
	textIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "content", Value: "text"},
			{Key: "metadata.functions", Value: "text"},
			{Key: "metadata.classes", Value: "text"},
			{Key: "filePath", Value: "text"},
			{Key: "fileName", Value: "text"},
			{Key: "imports", Value: "text"},
		},
		Options: options.Index().
			SetWeights(bson.D{
				{Key: "content", Value: 10},
				{Key: "metadata.functions", Value: 8},
				{Key: "metadata.classes", Value: 8},
				{Key: "filePath", Value: 5},
				{Key: "fileName", Value: 5},
				{Key: "imports", Value: 2},
			}).
			SetDefaultLanguage("none").
			SetLanguageOverride("language_override").
			SetName("CodeSearchIndex"),
	}

	// Additional indexes for code_chunks
	codeChunkIndexes := []mongo.IndexModel{
		textIndexModel,
		{
			Keys:    bson.D{{Key: "repositoryId", Value: 1}},
			Options: options.Index().SetName("repositoryId_1"),
		},
		{
			Keys:    bson.D{{Key: "language", Value: 1}},
			Options: options.Index().SetName("language_1"),
		},
		{
			Keys:    bson.D{{Key: "fileName", Value: 1}},
			Options: options.Index().SetName("fileName_1"),
		},
		{
			Keys:    bson.D{{Key: "contentHash", Value: 1}},
			Options: options.Index().SetName("contentHash_1"),
		},
		{
			Keys:    bson.D{{Key: "createdAt", Value: 1}},
			Options: options.Index().SetName("createdAt_1"),
		},
	}

	// Create code_chunks indexes
	_, err := codeChunksCollection.Indexes().CreateMany(ctx, codeChunkIndexes)
	if err != nil {
		log.Printf("Warning: Failed to create code_chunks indexes: %v", err)
	} else {
		log.Println("Successfully created code_chunks indexes")
	}

	// Users collection indexes
	usersCollection := db.Collection("users")
	userIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetName("email_1").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "githubId", Value: 1}},
			Options: options.Index().SetName("githubId_1").SetSparse(true),
		},
	}

	_, err = usersCollection.Indexes().CreateMany(ctx, userIndexes)
	if err != nil {
		log.Printf("Warning: Failed to create users indexes: %v", err)
	} else {
		log.Println("Successfully created users indexes")
	}

	// Repositories collection indexes
	repositoriesCollection := db.Collection("repositories")
	repoIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "userId", Value: 1}},
			Options: options.Index().SetName("userId_1"),
		},
		{
			Keys:    bson.D{{Key: "fullName", Value: 1}},
			Options: options.Index().SetName("fullName_1"),
		},
		{
			Keys:    bson.D{{Key: "status", Value: 1}},
			Options: options.Index().SetName("status_1"),
		},
		{
			Keys:    bson.D{{Key: "githubRepoId", Value: 1}},
			Options: options.Index().SetName("githubRepoId_1").SetSparse(true),
		},
	}

	_, err = repositoriesCollection.Indexes().CreateMany(ctx, repoIndexes)
	if err != nil {
		log.Printf("Warning: Failed to create repositories indexes: %v", err)
	} else {
		log.Println("Successfully created repositories indexes")
	}

	return nil
}

// InitializeCollections ensures all required collections exist
func (m *MongoDB) InitializeCollections() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := m.Database()

	// List of required collections
	collections := []string{"users", "repositories", "code_chunks"}

	// Get existing collections
	existingCollections, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	existingSet := make(map[string]bool)
	for _, name := range existingCollections {
		existingSet[name] = true
	}

	// Create missing collections
	for _, collectionName := range collections {
		if !existingSet[collectionName] {
			err := db.CreateCollection(ctx, collectionName)
			if err != nil {
				log.Printf("Warning: Failed to create collection %s: %v", collectionName, err)
			} else {
				log.Printf("Created collection: %s", collectionName)
			}
		}
	}

	return nil
}
