// ABOUTME: Unified server implementation that implements generated.ServerInterface
// ABOUTME: Delegates all methods to individual handlers following OpenAPI specification
package server

import (
	"github-analyzer/internal/generated"
	"github-analyzer/internal/handlers"
	"github-analyzer/internal/services"

	"github.com/gin-gonic/gin"
)

// Server implements generated.ServerInterface by delegating to individual handlers
type Server struct {
	health       *handlers.HealthHandler
	auth         *handlers.AuthHandler
	dashboard    *handlers.DashboardHandler
	repository   *handlers.RepositoryHandler
	github       *handlers.GitHubHandler
	search       *handlers.SearchHandler
	vectorSearch *handlers.VectorSearchHandler
	chat         *handlers.ChatHandler
	pipeline     *services.EmbeddingPipeline
}

// Ensure Server implements generated.ServerInterface
var _ generated.ServerInterface = (*Server)(nil)

// NewServer creates a new unified server with all handlers
func NewServer(
	health *handlers.HealthHandler,
	auth *handlers.AuthHandler,
	dashboard *handlers.DashboardHandler,
	repository *handlers.RepositoryHandler,
	github *handlers.GitHubHandler,
	search *handlers.SearchHandler,
	vectorSearch *handlers.VectorSearchHandler,
	chat *handlers.ChatHandler,
	pipeline *services.EmbeddingPipeline,
) *Server {
	return &Server{
		health:       health,
		auth:         auth,
		dashboard:    dashboard,
		repository:   repository,
		github:       github,
		search:       search,
		vectorSearch: vectorSearch,
		chat:         chat,
		pipeline:     pipeline,
	}
}

// Health endpoints
func (s *Server) GetHealth(c *gin.Context) {
	s.health.GetHealth(c)
}

func (s *Server) GetApiHealth(c *gin.Context) {
	s.health.GetApiHealth(c)
}

// Auth endpoints
func (s *Server) LoginUser(c *gin.Context) {
	s.auth.LoginUser(c)
}

func (s *Server) GetCurrentUser(c *gin.Context) {
	s.auth.GetCurrentUser(c)
}

// GitHub Auth endpoints
func (s *Server) GithubLogin(c *gin.Context, params generated.GithubLoginParams) {
	s.github.GitHubLogin(c)
}

func (s *Server) GithubCallback(c *gin.Context) {
	s.github.GitHubCallback(c)
}

func (s *Server) GithubDisconnect(c *gin.Context) {
	s.github.GitHubDisconnect(c)
}

// Dashboard endpoints
func (s *Server) GetDashboardStats(c *gin.Context) {
	s.dashboard.GetDashboardStats(c)
}

func (s *Server) GetDashboardActivity(c *gin.Context, params generated.GetDashboardActivityParams) {
	s.dashboard.GetDashboardActivity(c, params)
}

func (s *Server) GetDashboardTrends(c *gin.Context, params generated.GetDashboardTrendsParams) {
	s.dashboard.GetDashboardTrends(c, params)
}

// GitHub API endpoints
func (s *Server) GetGitHubRepositories(c *gin.Context, params generated.GetGitHubRepositoriesParams) {
	s.github.GetGitHubRepositories(c)
}

func (s *Server) ValidateGitHubRepository(c *gin.Context, owner string, repo string) {
	s.github.ValidateGitHubRepository(c)
}

// Repository endpoints
func (s *Server) GetRepositories(c *gin.Context, params generated.GetRepositoriesParams) {
	s.repository.GetRepositories(c)
}

func (s *Server) CreateRepository(c *gin.Context) {
	s.repository.CreateRepository(c)
}

func (s *Server) GetRepository(c *gin.Context, id string) {
	s.repository.GetRepository(c)
}

func (s *Server) UpdateRepository(c *gin.Context, id string) {
	s.repository.UpdateRepository(c)
}

func (s *Server) DeleteRepository(c *gin.Context, id string) {
	s.repository.DeleteRepository(c)
}

func (s *Server) GetRepositoryStats(c *gin.Context, id string) {
	s.repository.GetRepositoryStats(c)
}

// TriggerRepositoryImport triggers manual repository import for stuck repositories
func (s *Server) TriggerRepositoryImport(c *gin.Context, id string) {
	s.repository.TriggerRepositoryImport(c)
}

// Search endpoints
func (s *Server) GlobalSearch(c *gin.Context, params generated.GlobalSearchParams) {
	s.search.GlobalSearch(c)
}

func (s *Server) RepositorySearch(c *gin.Context, id string, params generated.RepositorySearchParams) {
	s.search.RepositorySearch(c)
}

func (s *Server) GetSearchSuggestions(c *gin.Context, params generated.GetSearchSuggestionsParams) {
	s.search.GetSearchSuggestions(c)
}

func (s *Server) QuickSearch(c *gin.Context, params generated.QuickSearchParams) {
	s.search.QuickSearch(c)
}

func (s *Server) GetLanguages(c *gin.Context, params generated.GetLanguagesParams) {
	s.search.GetLanguages(c)
}

func (s *Server) GetRecentChunks(c *gin.Context, params generated.GetRecentChunksParams) {
	s.search.GetRecentChunks(c)
}

func (s *Server) GetSearchStats(c *gin.Context, params generated.GetSearchStatsParams) {
	s.search.GetSearchStats(c)
}

// Vector Search endpoints
func (s *Server) VectorSearch(c *gin.Context) {
	s.vectorSearch.VectorSearch(c)
}

func (s *Server) HybridSearch(c *gin.Context) {
	s.vectorSearch.HybridSearch(c)
}

func (s *Server) FindSimilarChunks(c *gin.Context, chunkId string, params generated.FindSimilarChunksParams) {
	s.vectorSearch.FindSimilar(c)
}

// Repository embedding endpoints
func (s *Server) TriggerRepositoryEmbedding(c *gin.Context, id string) {
	s.vectorSearch.EmbedRepository(c)
}

func (s *Server) GetRepositoryEmbeddingStatus(c *gin.Context, id string) {
	s.vectorSearch.GetEmbeddingStatus(c)
}

// Embedding pipeline stats endpoint
func (s *Server) GetEmbeddingPipelineStats(c *gin.Context) {
	if s.pipeline == nil {
		c.JSON(500, gin.H{"error": "pipeline not configured"})
		return
	}
	stats, err := s.pipeline.GetPipelineStats(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, stats)
}

// Chat endpoints
func (s *Server) CreateChatSession(c *gin.Context) {
	s.chat.CreateChatSession(c)
}

func (s *Server) ListChatSessions(c *gin.Context, params generated.ListChatSessionsParams) {
	s.chat.ListChatSessions(c, params)
}

func (s *Server) GetChatSession(c *gin.Context, id string) {
	c.Params = append(c.Params, gin.Param{Key: "id", Value: id})
	s.chat.GetChatSession(c)
}

func (s *Server) DeleteChatSession(c *gin.Context, id string) {
	c.Params = append(c.Params, gin.Param{Key: "id", Value: id})
	s.chat.DeleteChatSession(c)
}

func (s *Server) UpdateChatSession(c *gin.Context, id string) {
	s.chat.UpdateChatSession(c, id)
}

func (s *Server) SendChatMessage(c *gin.Context, id string) {
	c.Params = append(c.Params, gin.Param{Key: "id", Value: id})
	s.chat.SendChatMessage(c)
}