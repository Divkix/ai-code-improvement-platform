// ABOUTME: GitHub API service for repository operations and OAuth authentication
// ABOUTME: Provides methods for GitHub integration, rate limiting, and error handling

package services

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/google/go-github/v73/github"
	"golang.org/x/oauth2"
	oauth2_github "golang.org/x/oauth2/github"
	"go.mongodb.org/mongo-driver/mongo"
)

// GitHubService provides GitHub API operations
type GitHubService struct {
	db              *mongo.Database
	clientID        string
	clientSecret    string
	encryptionKey   []byte
	oauthConfig     *oauth2.Config
}

// GitHubRepository represents repository data from GitHub API
type GitHubRepository struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	FullName        string    `json:"fullName"`
	Owner           string    `json:"owner"`
	Description     *string   `json:"description"`
	Private         bool      `json:"private"`
	Language        *string   `json:"language"`
	StargazersCount int       `json:"stargazersCount"`
	ForksCount      int       `json:"forksCount"`
	Size            int       `json:"size"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
	PushedAt        *time.Time `json:"pushedAt,omitempty"`
}

// GitHubImportProgress represents the progress of importing a repository
type GitHubImportProgress struct {
	Stage       string `json:"stage"`
	Progress    int    `json:"progress"`
	Message     string `json:"message"`
	Error       string `json:"error,omitempty"`
	FilesFetched int   `json:"files_fetched"`
	TotalFiles  int    `json:"total_files"`
}

// NewGitHubService creates a new GitHub service
func NewGitHubService(db *mongo.Database, clientID, clientSecret, encryptionKey string) *GitHubService {
	key := []byte(encryptionKey)
	if len(key) != 32 {
		// Pad or truncate to 32 bytes for AES-256
		padded := make([]byte, 32)
		copy(padded, key)
		key = padded
	}

	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"repo", "user:email"},
		Endpoint:     oauth2_github.Endpoint,
	}

	return &GitHubService{
		db:            db,
		clientID:      clientID,
		clientSecret:  clientSecret,
		encryptionKey: key,
		oauthConfig:   oauthConfig,
	}
}

// GetOAuthConfig returns the OAuth configuration for GitHub
func (s *GitHubService) GetOAuthConfig() *oauth2.Config {
	return s.oauthConfig
}

// CreateClient creates a GitHub client with the provided access token
func (s *GitHubService) CreateClient(accessToken string) *github.Client {
	return github.NewClient(nil).WithAuthToken(accessToken)
}

// ValidateRepository validates that a repository exists and is accessible
func (s *GitHubService) ValidateRepository(ctx context.Context, accessToken, owner, repo string) (*GitHubRepository, error) {
	client := s.CreateClient(accessToken)

	repository, _, err := client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		if s.IsRateLimited(err) {
			return nil, fmt.Errorf("github rate limit exceeded: %w", err)
		}
		return nil, fmt.Errorf("failed to fetch repository: %w", err)
	}

	return s.convertToGitHubRepository(repository), nil
}

// GetUserRepositories fetches repositories for the authenticated user
func (s *GitHubService) GetUserRepositories(ctx context.Context, accessToken string, page int) ([]*GitHubRepository, error) {
	client := s.CreateClient(accessToken)

	opt := &github.RepositoryListByAuthenticatedUserOptions{
		Visibility: "all",
		Sort:       "updated",
		ListOptions: github.ListOptions{
			Page:    page,
			PerPage: 20,
		},
	}

	repos, _, err := client.Repositories.ListByAuthenticatedUser(ctx, opt)
	if err != nil {
		if s.IsRateLimited(err) {
			return nil, fmt.Errorf("github rate limit exceeded: %w", err)
		}
		return nil, fmt.Errorf("failed to fetch repositories: %w", err)
	}

	var ghRepos []*GitHubRepository
	for _, repo := range repos {
		ghRepos = append(ghRepos, s.convertToGitHubRepository(repo))
	}

	return ghRepos, nil
}

// GetRepositoryStatistics fetches detailed statistics for a repository
func (s *GitHubService) GetRepositoryStatistics(ctx context.Context, accessToken, owner, repo string) (map[string]interface{}, error) {
	client := s.CreateClient(accessToken)

	// Get repository details
	repository, _, err := client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository: %w", err)
	}

	// Get languages
	languages, _, err := client.Repositories.ListLanguages(ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch languages: %w", err)
	}

	// Calculate total lines (approximation based on repository size)
	totalLines := repository.GetSize() * 10 // Rough approximation

	stats := map[string]interface{}{
		"total_files":      repository.GetSize() / 50, // Rough approximation
		"total_lines":      totalLines,
		"languages":        languages,
		"stargazers_count": repository.GetStargazersCount(),
		"forks_count":      repository.GetForksCount(),
		"size":            repository.GetSize(),
		"last_commit_date": repository.GetPushedAt(),
		"created_at":      repository.GetCreatedAt(),
		"updated_at":      repository.GetUpdatedAt(),
	}

	return stats, nil
}

// CheckRateLimit returns the current rate limit status
func (s *GitHubService) CheckRateLimit(ctx context.Context, accessToken string) (*github.RateLimits, error) {
	client := s.CreateClient(accessToken)
	
	rateLimits, _, err := client.RateLimit.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check rate limit: %w", err)
	}

	return rateLimits, nil
}

// EncryptToken encrypts a GitHub access token for storage
func (s *GitHubService) EncryptToken(token string) (string, error) {
	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(token), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptToken decrypts a GitHub access token from storage
func (s *GitHubService) DecryptToken(encryptedToken string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedToken)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.encryptionKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// convertToGitHubRepository converts a GitHub API repository to our internal format
func (s *GitHubService) convertToGitHubRepository(repo *github.Repository) *GitHubRepository {
	var owner string
	if repo.Owner != nil {
		owner = repo.Owner.GetLogin()
	}

	var pushedAt *time.Time
	if repo.PushedAt != nil {
		pushedTime := repo.GetPushedAt().Time
		pushedAt = &pushedTime
	}

	return &GitHubRepository{
		ID:              repo.GetID(),
		Name:            repo.GetName(),
		FullName:        repo.GetFullName(),
		Owner:           owner,
		Description:     repo.Description,
		Private:         repo.GetPrivate(),
		Language:        repo.Language,
		StargazersCount: repo.GetStargazersCount(),
		ForksCount:      repo.GetForksCount(),
		Size:            repo.GetSize(),
		CreatedAt:       repo.GetCreatedAt().Time,
		UpdatedAt:       repo.GetUpdatedAt().Time,
		PushedAt:        pushedAt,
	}
}

// IsRateLimited checks if an error is due to rate limiting
func (s *GitHubService) IsRateLimited(err error) bool {
	if err == nil {
		return false
	}
	
	_, isRateLimit := err.(*github.RateLimitError)
	_, isAbuse := err.(*github.AbuseRateLimitError)
	
	return isRateLimit || isAbuse
}

// GetRateLimitResetTime extracts the rate limit reset time from an error
func (s *GitHubService) GetRateLimitResetTime(err error) *time.Time {
	if rateLimitErr, ok := err.(*github.RateLimitError); ok {
		return &rateLimitErr.Rate.Reset.Time
	}
	if abuseErr, ok := err.(*github.AbuseRateLimitError); ok {
		if abuseErr.RetryAfter != nil {
			retryTime := time.Now().Add(*abuseErr.RetryAfter)
			return &retryTime
		}
	}
	return nil
}