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
	"log"
	"time"

	"github-analyzer/internal/models"
	"github-analyzer/internal/utils"

	"github.com/google/go-github/v73/github"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/oauth2"
	oauth2_github "golang.org/x/oauth2/github"
)

// GitHubService provides GitHub API operations
type GitHubService struct {
	db            *mongo.Database
	clientID      string
	clientSecret  string
	encryptionKey []byte
	oauthConfig   *oauth2.Config
	batchSize     int
	maxFileSize   int
}

// GitHubRepository represents repository data from GitHub API
type GitHubRepository struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	FullName        string     `json:"fullName"`
	Owner           string     `json:"owner"`
	Description     *string    `json:"description"`
	Private         bool       `json:"private"`
	Language        *string    `json:"language"`
	StargazersCount int        `json:"stargazersCount"`
	ForksCount      int        `json:"forksCount"`
	Size            int        `json:"size"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	PushedAt        *time.Time `json:"pushedAt,omitempty"`
}

// GitHubImportProgress represents the progress of importing a repository
type GitHubImportProgress struct {
	Stage        string `json:"stage"`
	Progress     int    `json:"progress"`
	Message      string `json:"message"`
	Error        string `json:"error,omitempty"`
	FilesFetched int    `json:"files_fetched"`
	TotalFiles   int    `json:"total_files"`
}

// NewGitHubService creates a new GitHub service
func NewGitHubService(db *mongo.Database, clientID, clientSecret, encryptionKey string, batchSize, maxFileSize int) *GitHubService {
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
		batchSize:     batchSize,
		maxFileSize:   maxFileSize,
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

// SearchUserRepositories searches repositories for the authenticated user including organization repositories
func (s *GitHubService) SearchUserRepositories(ctx context.Context, accessToken, query string, limit int) ([]*GitHubRepository, error) {
	client := s.CreateClient(accessToken)

	// Get all repositories accessible to the user (including organization repos)
	// First get user's own repositories with the search term
	userRepos, err := s.searchUserOwnedRepositories(ctx, client, query, limit)
	if err != nil {
		return nil, err
	}

	// Then get organization repositories with the search term
	orgRepos, err := s.searchOrganizationRepositories(ctx, client, query, limit-len(userRepos))
	if err != nil {
		// Don't fail completely if org search fails, just log and continue with user repos
		log.Printf("Warning: Failed to search organization repositories: %v", err)
	}

	// Combine and deduplicate results
	repoMap := make(map[int64]*GitHubRepository)
	
	// Add user repositories
	for _, repo := range userRepos {
		repoMap[repo.ID] = repo
	}
	
	// Add organization repositories (avoiding duplicates)
	for _, repo := range orgRepos {
		if _, exists := repoMap[repo.ID]; !exists {
			repoMap[repo.ID] = repo
		}
	}

	// Convert map to slice and sort by updated time
	var allRepos []*GitHubRepository
	for _, repo := range repoMap {
		allRepos = append(allRepos, repo)
	}

	// Sort by updated time (most recent first)
	for i := 0; i < len(allRepos)-1; i++ {
		for j := i + 1; j < len(allRepos); j++ {
			if allRepos[i].UpdatedAt.Before(allRepos[j].UpdatedAt) {
				allRepos[i], allRepos[j] = allRepos[j], allRepos[i]
			}
		}
	}

	// Limit results
	if len(allRepos) > limit {
		allRepos = allRepos[:limit]
	}

	return allRepos, nil
}

// searchUserOwnedRepositories searches repositories owned by the authenticated user
func (s *GitHubService) searchUserOwnedRepositories(ctx context.Context, client *github.Client, query string, limit int) ([]*GitHubRepository, error) {
	// Get the authenticated user
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get authenticated user: %w", err)
	}

	// Build search query for user's own repositories
	searchQuery := fmt.Sprintf("%s user:%s", query, user.GetLogin())

	// Search repositories
	opt := &github.SearchOptions{
		Sort:  "updated",
		Order: "desc",
		ListOptions: github.ListOptions{
			PerPage: limit,
		},
	}

	result, _, err := client.Search.Repositories(ctx, searchQuery, opt)
	if err != nil {
		if s.IsRateLimited(err) {
			return nil, fmt.Errorf("github rate limit exceeded: %w", err)
		}
		return nil, fmt.Errorf("failed to search user repositories: %w", err)
	}

	var ghRepos []*GitHubRepository
	for _, repo := range result.Repositories {
		ghRepos = append(ghRepos, s.convertToGitHubRepository(repo))
	}

	return ghRepos, nil
}

// searchOrganizationRepositories searches repositories from organizations the user has access to
func (s *GitHubService) searchOrganizationRepositories(ctx context.Context, client *github.Client, query string, limit int) ([]*GitHubRepository, error) {
	if limit <= 0 {
		return []*GitHubRepository{}, nil
	}

	// Get user's organizations
	orgs, _, err := client.Organizations.List(ctx, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user organizations: %w", err)
	}

	var allOrgRepos []*GitHubRepository
	
	// Search each organization's repositories
	for _, org := range orgs {
		if len(allOrgRepos) >= limit {
			break
		}

		orgLogin := org.GetLogin()
		searchQuery := fmt.Sprintf("%s org:%s", query, orgLogin)

		opt := &github.SearchOptions{
			Sort:  "updated",
			Order: "desc",
			ListOptions: github.ListOptions{
				PerPage: limit - len(allOrgRepos),
			},
		}

		result, _, err := client.Search.Repositories(ctx, searchQuery, opt)
		if err != nil {
			// Log error but continue with other orgs
			log.Printf("Warning: Failed to search repositories for organization %s: %v", orgLogin, err)
			continue
		}

		for _, repo := range result.Repositories {
			if len(allOrgRepos) >= limit {
				break
			}
			allOrgRepos = append(allOrgRepos, s.convertToGitHubRepository(repo))
		}
	}

	return allOrgRepos, nil
}

// GetRecentUserRepositories fetches the most recently updated repositories for the authenticated user
func (s *GitHubService) GetRecentUserRepositories(ctx context.Context, accessToken string, limit int) ([]*GitHubRepository, error) {
	client := s.CreateClient(accessToken)

	opt := &github.RepositoryListByAuthenticatedUserOptions{
		Visibility: "all",
		Sort:       "updated",
		ListOptions: github.ListOptions{
			PerPage: limit,
		},
	}

	repos, _, err := client.Repositories.ListByAuthenticatedUser(ctx, opt)
	if err != nil {
		if s.IsRateLimited(err) {
			return nil, fmt.Errorf("github rate limit exceeded: %w", err)
		}
		return nil, fmt.Errorf("failed to fetch recent repositories: %w", err)
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

	// Safely convert GitHub Timestamp values to *time.Time for JSON/BSON compatibility
	var lastCommitPtr *time.Time
	if repository.PushedAt != nil {
		t := repository.GetPushedAt().Time
		lastCommitPtr = &t
	}

	createdAt := repository.GetCreatedAt().Time
	updatedAt := repository.GetUpdatedAt().Time

	stats := map[string]interface{}{
		"total_files":      repository.GetSize() / 50, // Rough approximation
		"total_lines":      totalLines,
		"languages":        languages,
		"stargazers_count": repository.GetStargazersCount(),
		"forks_count":      repository.GetForksCount(),
		"size":             repository.GetSize(),
		"last_commit_date": lastCommitPtr,
		"created_at":       createdAt,
		"updated_at":       updatedAt,
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

// FetchRepositoryFiles fetches all files from a GitHub repository
func (s *GitHubService) FetchRepositoryFiles(ctx context.Context, accessToken, owner, repo string) ([]*models.RepositoryFile, error) {
	client := s.CreateClient(accessToken)

	log.Printf("Fetching files for repository: %s/%s", owner, repo)

	// Get the repository tree (recursive)
	tree, _, err := client.Git.GetTree(ctx, owner, repo, "HEAD", true)
	if err != nil {
		if s.IsRateLimited(err) {
			return nil, fmt.Errorf("github rate limit exceeded while fetching tree: %w", err)
		}
		return nil, fmt.Errorf("failed to fetch repository tree: %w", err)
	}

	var files []*models.RepositoryFile
	var totalFiles int
	var processedFiles int

	// Count total files to process
	for _, entry := range tree.Entries {
		if entry.GetType() == "blob" {
			totalFiles++
		}
	}

	log.Printf("Found %d files to process in %s/%s", totalFiles, owner, repo)

	// Process files in batches to avoid rate limits (configurable via GITHUB_BATCH_SIZE)
	batchSize := s.batchSize

	for i := 0; i < len(tree.Entries); i += batchSize {
		end := i + batchSize
		if end > len(tree.Entries) {
			end = len(tree.Entries)
		}

		batch := tree.Entries[i:end]
		batchFiles, err := s.processBatch(ctx, client, owner, repo, batch)
		if err != nil {
			log.Printf("Failed to process batch %d-%d: %v", i, end-1, err)
			continue
		}

		files = append(files, batchFiles...)
		processedFiles += len(batchFiles)

		log.Printf("Processed %d/%d files", processedFiles, totalFiles)
	}

	log.Printf("Successfully fetched %d files from %s/%s", len(files), owner, repo)
	return files, nil
}

// processBatch processes a batch of repository files with enhanced error handling
func (s *GitHubService) processBatch(ctx context.Context, client *github.Client, owner, repo string, entries []*github.TreeEntry) ([]*models.RepositoryFile, error) {
	var files []*models.RepositoryFile
	var failedFiles []string
	var skippedFiles int

	log.Printf("Processing batch of %d entries for %s/%s", len(entries), owner, repo)

	for _, entry := range entries {
		// Only process files (blobs), not directories (trees)
		if entry.GetType() != "blob" {
			continue
		}

		// Create basic file info first
		path := entry.GetPath()
		sha := entry.GetSHA()
		size := entry.GetSize()

		// Skip files that are too large or clearly not code files (size limit configurable via GITHUB_MAX_FILE_SIZE)
		maxFileSize := s.maxFileSize

		if size > maxFileSize { // Skip files above the configurable limit
			log.Printf("Skipping large file: %s (%d bytes) - exceeds 1MB limit", path, size)
			skippedFiles++
			continue
		}

		// Create repository file with empty content first
		repoFile := models.NewRepositoryFile(path, "", sha, size)

		// Skip non-code files early
		if !repoFile.IsCodeFile() {
			skippedFiles++
			continue
		}

		// Fetch file content with retry logic
		content, err := s.getFileContentWithRetry(ctx, client, owner, repo, path, 3)
		if err != nil {
			log.Printf("Failed to fetch content for %s after retries: %v", path, err)
			failedFiles = append(failedFiles, path)
			continue
		}

		// Validate and clean content before storing
		if valid, reason := utils.ValidateContentForStorage(content); !valid {
			log.Printf("Skipping file %s due to encoding issue: %s", path, reason)
			skippedFiles++
			continue
		}
		
		// Clean content to ensure safe storage
		cleanedContent := utils.CleanContent(content)
		
		// Update file with cleaned content
		repoFile.Content = cleanedContent
		repoFile.Size = len(cleanedContent)

		// Validate file is suitable for processing
		if !repoFile.IsValidForProcessing() {
			log.Printf("Skipping file %s - not valid for processing (size: %d, lines: %d)",
				path, repoFile.Size, repoFile.GetLineCount())
			skippedFiles++
			continue
		}

		files = append(files, repoFile)
	}

	// Log batch processing summary
	log.Printf("Batch processing complete for %s/%s: %d files processed, %d skipped, %d failed",
		owner, repo, len(files), skippedFiles, len(failedFiles))

	if len(failedFiles) > 0 {
		log.Printf("Failed files: %v", failedFiles)
	}

	return files, nil
}

// getFileContentWithRetry fetches file content with retry logic for rate limiting
func (s *GitHubService) getFileContentWithRetry(ctx context.Context, client *github.Client, owner, repo, path string, maxRetries int) (string, error) {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		content, err := s.GetFileContent(ctx, client, owner, repo, path)
		if err == nil {
			return content, nil
		}

		lastErr = err

		// If it's a rate limit error, wait and retry
		if s.IsRateLimited(err) {
			if resetTime := s.GetRateLimitResetTime(err); resetTime != nil {
				waitDuration := time.Until(*resetTime)
				if waitDuration > 0 && waitDuration < 5*time.Minute {
					log.Printf("Rate limited, waiting %v before retry %d/%d for %s", waitDuration, attempt, maxRetries, path)
					time.Sleep(waitDuration)
					continue
				}
			}
			// Exponential backoff for rate limits without reset time
			backoff := time.Duration(attempt*attempt) * time.Second
			log.Printf("Rate limited, backing off %v before retry %d/%d for %s", backoff, attempt, maxRetries, path)
			time.Sleep(backoff)
			continue
		}

		// For other errors, don't retry immediately
		if attempt < maxRetries {
			backoff := time.Duration(attempt) * time.Second
			log.Printf("Error fetching %s, retrying in %v (attempt %d/%d): %v", path, backoff, attempt, maxRetries, err)
			time.Sleep(backoff)
		}
	}

	return "", fmt.Errorf("failed to fetch %s after %d attempts: %w", path, maxRetries, lastErr)
}

// GetFileContent fetches the content of a specific file from GitHub
func (s *GitHubService) GetFileContent(ctx context.Context, client *github.Client, owner, repo, path string) (string, error) {
	fileContent, _, _, err := client.Repositories.GetContents(ctx, owner, repo, path, nil)
	if err != nil {
		if s.IsRateLimited(err) {
			return "", fmt.Errorf("github rate limit exceeded while fetching %s: %w", path, err)
		}
		return "", fmt.Errorf("failed to fetch file content for %s: %w", path, err)
	}

	if fileContent == nil {
		return "", fmt.Errorf("file content is nil for %s", path)
	}

	content, err := fileContent.GetContent()
	if err != nil {
		return "", fmt.Errorf("failed to decode file content for %s: %w", path, err)
	}

	return content, nil
}

// GetRepositoryTree fetches the repository tree structure
func (s *GitHubService) GetRepositoryTree(ctx context.Context, accessToken, owner, repo string) (*models.RepositoryTree, error) {
	client := s.CreateClient(accessToken)

	tree, _, err := client.Git.GetTree(ctx, owner, repo, "HEAD", true)
	if err != nil {
		if s.IsRateLimited(err) {
			return nil, fmt.Errorf("github rate limit exceeded while fetching tree: %w", err)
		}
		return nil, fmt.Errorf("failed to fetch repository tree: %w", err)
	}

	repoTree := &models.RepositoryTree{
		SHA: tree.GetSHA(),
		URL: "", // GitHub Tree object doesn't have URL field
	}

	for _, entry := range tree.Entries {
		item := &models.RepositoryTreeItem{
			Path: entry.GetPath(),
			Mode: entry.GetMode(),
			Type: entry.GetType(),
			SHA:  entry.GetSHA(),
			URL:  entry.GetURL(),
		}

		if entry.Size != nil {
			size := *entry.Size
			item.Size = &size
		}

		repoTree.Tree = append(repoTree.Tree, item)
	}

	return repoTree, nil
}

// GetRepositoryLanguages fetches the languages used in a repository with detailed stats
func (s *GitHubService) GetRepositoryLanguages(ctx context.Context, accessToken, owner, repo string) (map[string]int, error) {
	client := s.CreateClient(accessToken)

	languages, _, err := client.Repositories.ListLanguages(ctx, owner, repo)
	if err != nil {
		if s.IsRateLimited(err) {
			return nil, fmt.Errorf("github rate limit exceeded while fetching languages: %w", err)
		}
		return nil, fmt.Errorf("failed to fetch repository languages: %w", err)
	}

	return languages, nil
}
