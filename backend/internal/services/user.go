// ABOUTME: User service for MongoDB operations including CRUD and authentication
// ABOUTME: Handles user registration, login, and profile management

package services

import (
	"context"
	"errors"
	"time"

	"acip.divkix.me/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrUserExists      = errors.New("user already exists")
	ErrInvalidPassword = errors.New("invalid password")
)

// UserService provides user-related operations
type UserService struct {
	collection *mongo.Collection
}

// NewUserService creates a new user service
func NewUserService(db *mongo.Database) *UserService {
	return &UserService{
		collection: db.Collection(models.UserCollection),
	}
}


// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := s.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	var user models.User
	err = s.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(ctx context.Context, userID string, updates bson.M) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return ErrUserNotFound
	}

	updates["updatedAt"] = time.Now()
	_, err = s.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": updates})
	return err
}

// GetByID retrieves a user by ObjectID
func (s *UserService) GetByID(ctx context.Context, userID primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := s.collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateGitHubConnection updates user with GitHub token and username
func (s *UserService) UpdateGitHubConnection(ctx context.Context, userID primitive.ObjectID, encryptedToken, githubUsername string) (*models.User, error) {
	updates := bson.M{
		"githubToken":     encryptedToken,
		"githubUsername":  githubUsername,
		"githubConnected": true,
		"updatedAt":       time.Now(),
	}

	_, err := s.collection.UpdateOne(ctx, bson.M{"_id": userID}, bson.M{"$set": updates})
	if err != nil {
		return nil, err
	}

	// Return updated user
	return s.GetByID(ctx, userID)
}

// RemoveGitHubConnection removes GitHub connection from user
func (s *UserService) RemoveGitHubConnection(ctx context.Context, userID primitive.ObjectID) (*models.User, error) {
	updates := bson.M{
		"githubConnected": false,
		"updatedAt":       time.Now(),
	}

	unset := bson.M{
		"githubToken":    "",
		"githubUsername": "",
	}

	_, err := s.collection.UpdateOne(ctx, 
		bson.M{"_id": userID}, 
		bson.M{
			"$set":   updates,
			"$unset": unset,
		})
	if err != nil {
		return nil, err
	}

	// Return updated user
	return s.GetByID(ctx, userID)
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return ErrUserNotFound
	}

	_, err = s.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	return err
}