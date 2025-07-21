// ABOUTME: Dashboard service for generating statistics, activity, and trend data
// ABOUTME: Provides compelling demo metrics with realistic calculations for cost savings

package services

import (
	"context"
	"math/rand"
	"strconv"
	"time"

	"github-analyzer/internal/models"
)

// DashboardService handles dashboard-related operations
type DashboardService struct {
	// userService *UserService // Will be used when we have real data
}

// NewDashboardService creates a new dashboard service
func NewDashboardService() *DashboardService {
	return &DashboardService{}
}

// GetStats returns dashboard statistics with compelling demo data
func (s *DashboardService) GetStats(ctx context.Context, userID string) (*models.DashboardStats, error) {
	// For demo purposes, generate impressive but realistic metrics
	// In production, these would come from actual database queries
	
	baseRepos := 3 + rand.Intn(5) // 3-7 repositories
	chunksPerRepo := 1500 + rand.Intn(2000) // 1500-3500 chunks per repo
	totalChunks := baseRepos * chunksPerRepo
	
	// Calculate cost savings based on developer time saved
	// Assumptions: 
	// - Without AI: 2 hours to understand unfamiliar code
	// - With AI: 15 minutes to get same understanding
	// - Developer hourly rate: $120/hour
	// - Average queries per day: 12
	hoursPerQuery := 1.75 // 2 hours - 15 minutes = 1.75 hours saved
	queriesPerMonth := 12 * 30 // 12 queries per day * 30 days
	hourlyRate := 120.0
	monthlySavings := float64(queriesPerMonth) * hoursPerQuery * hourlyRate
	
	stats := &models.DashboardStats{
		TotalRepositories:       baseRepos,
		CodeChunksProcessed:     totalChunks,
		AvgResponseTime:         1.2 + rand.Float64()*0.8, // 1.2-2.0 seconds
		CostSavingsMonthly:      monthlySavings,
		IssuesPreventedMonthly:  15 + rand.Intn(20), // 15-35 issues
		DeveloperHoursReclaimed: float64(queriesPerMonth) * hoursPerQuery,
	}
	
	return stats, nil
}

// GetActivity returns recent activity items with demo data
func (s *DashboardService) GetActivity(ctx context.Context, userID string, limit int) ([]*models.ActivityItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	
	activities := make([]*models.ActivityItem, 0, limit)
	
	// Generate realistic activity items
	activityTemplates := []struct {
		actType    models.ActivityType
		message    string
		severity   models.ActivitySeverity
		repoName   string
	}{
		{models.ActivityRepositoryImported, "Successfully imported React codebase with 847 files", models.SeveritySuccess, "frontend-app"},
		{models.ActivityAnalysisCompleted, "Code quality analysis completed - 94% coverage achieved", models.SeveritySuccess, "api-service"},
		{models.ActivityIssueDetected, "Potential memory leak detected in user authentication module", models.SeverityWarning, "auth-service"},
		{models.ActivityOptimizationFound, "Database query optimization opportunity found - 67% faster", models.SeverityInfo, "data-processor"},
		{models.ActivityRepositoryImported, "Node.js microservice imported successfully", models.SeveritySuccess, "payment-gateway"},
		{models.ActivityIssueDetected, "Deprecated API usage found in 3 components", models.SeverityWarning, "frontend-app"},
		{models.ActivityAnalysisCompleted, "Security scan completed - no vulnerabilities found", models.SeveritySuccess, "api-service"},
		{models.ActivityOptimizationFound, "Code duplication reduced by 23% through refactoring suggestions", models.SeverityInfo, "shared-utils"},
	}
	
	for i := 0; i < limit && i < len(activityTemplates); i++ {
		template := activityTemplates[i%len(activityTemplates)]
		
		activity := &models.ActivityItem{
			ID:             strconv.Itoa(i + 1),
			Type:           template.actType,
			Message:        template.message,
			Timestamp:      time.Now().Add(-time.Duration(i*2+rand.Intn(60)) * time.Minute),
			Severity:       template.severity,
			RepositoryName: template.repoName,
		}
		
		activities = append(activities, activity)
	}
	
	return activities, nil
}

// GetTrends returns trend data for charts with demo data
func (s *DashboardService) GetTrends(ctx context.Context, userID string, days int) ([]*models.TrendDataPoint, error) {
	if days < 7 || days > 90 {
		days = 30
	}
	
	trends := make([]*models.TrendDataPoint, 0, days)
	
	// Generate realistic trending data showing improvement over time
	baseQuality := 65.0
	basePerformance := 70.0
	baseIssues := 2
	
	for i := 0; i < days; i++ {
		date := time.Now().AddDate(0, 0, -days+i+1)
		
		// Simulate gradual improvement with some variance
		qualityTrend := float64(i) * 0.5 // Gradual improvement
		performanceTrend := float64(i) * 0.4
		issuesTrend := i / 10 // Fewer issues over time
		
		// Add some realistic variance
		qualityVariance := (rand.Float64() - 0.5) * 4
		performanceVariance := (rand.Float64() - 0.5) * 5
		
		trend := &models.TrendDataPoint{
			Date:             date.Format("2006-01-02"),
			CodeQuality:      baseQuality + qualityTrend + qualityVariance,
			IssuesResolved:   baseIssues + issuesTrend + rand.Intn(3),
			PerformanceScore: basePerformance + performanceTrend + performanceVariance,
		}
		
		// Ensure values stay within realistic bounds
		if trend.CodeQuality > 98 {
			trend.CodeQuality = 98
		}
		if trend.CodeQuality < 60 {
			trend.CodeQuality = 60
		}
		if trend.PerformanceScore > 97 {
			trend.PerformanceScore = 97
		}
		if trend.PerformanceScore < 65 {
			trend.PerformanceScore = 65
		}
		
		trends = append(trends, trend)
	}
	
	return trends, nil
}