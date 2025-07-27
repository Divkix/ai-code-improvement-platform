// ABOUTME: User model for MongoDB with authentication fields and validation
// ABOUTME: Implements bcrypt password hashing and JWT token generation

package models

import (
	"time"

	"acip.divkix.me/internal/generated"
	"github.com/oapi-codegen/runtime/types"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system
type User struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email           string             `bson:"email" json:"email"`
	Password        string             `bson:"password" json:"-"` // Never include password in JSON
	Name            string             `bson:"name" json:"name"`
	GitHubToken     string             `bson:"githubToken,omitempty" json:"-"` // Encrypted GitHub token, never in JSON
	GitHubUsername  string             `bson:"githubUsername,omitempty" json:"githubUsername,omitempty"`
	GitHubConnected bool               `bson:"githubConnected" json:"githubConnected"`
	CreatedAt       time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt       time.Time          `bson:"updatedAt" json:"updatedAt"`
}


// LoginUserRequest represents the request to login a user
type LoginUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserResponse represents the user data returned in API responses
type UserResponse struct {
	ID              string    `json:"id"`
	Email           string    `json:"email"`
	Name            string    `json:"name"`
	GitHubConnected bool      `json:"githubConnected"`
	GitHubUsername  string    `json:"githubUsername,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
}

// AuthResponse represents the response after successful authentication
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// ToResponse converts a User model to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:              u.ID.Hex(),
		Email:           u.Email,
		Name:            u.Name,
		GitHubConnected: u.GitHubConnected,
		GitHubUsername:  u.GitHubUsername,
		CreatedAt:       u.CreatedAt,
	}
}

// ToGeneratedUser converts a User model to generated.User
func (u *User) ToGeneratedUser() generated.User {
	email := types.Email(u.Email)
	githubConnected := u.GitHubConnected
	var githubUsername *string
	if u.GitHubUsername != "" {
		githubUsername = &u.GitHubUsername
	}

	return generated.User{
		Id:              u.ID.Hex(),
		Email:           email,
		Name:            u.Name,
		GithubConnected: &githubConnected,
		GithubUsername:  githubUsername,
		CreatedAt:       u.CreatedAt,
	}
}

// Collection name for MongoDB
const UserCollection = "users"