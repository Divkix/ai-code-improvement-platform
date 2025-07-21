// ABOUTME: Dashboard models for statistics, activity, and trend data
// ABOUTME: Implements compelling demo metrics and cost savings calculations

package models

import (
	"time"
)

// DashboardStats represents key metrics for the dashboard
type DashboardStats struct {
	TotalRepositories        int     `json:"totalRepositories"`
	CodeChunksProcessed      int     `json:"codeChunksProcessed"`
	AvgResponseTime          float64 `json:"avgResponseTime"`
	CostSavingsMonthly       float64 `json:"costSavingsMonthly"`
	IssuesPreventedMonthly   int     `json:"issuesPreventedMonthly"`
	DeveloperHoursReclaimed  float64 `json:"developerHoursReclaimed"`
}

// ActivityType represents different types of activity
type ActivityType string

const (
	ActivityRepositoryImported ActivityType = "repository_imported"
	ActivityAnalysisCompleted  ActivityType = "analysis_completed"
	ActivityIssueDetected      ActivityType = "issue_detected"
	ActivityOptimizationFound  ActivityType = "optimization_found"
)

// ActivitySeverity represents the severity level of an activity
type ActivitySeverity string

const (
	SeverityInfo    ActivitySeverity = "info"
	SeverityWarning ActivitySeverity = "warning"
	SeverityError   ActivitySeverity = "error"
	SeveritySuccess ActivitySeverity = "success"
)

// ActivityItem represents an activity item in the feed
type ActivityItem struct {
	ID             string           `json:"id"`
	Type           ActivityType     `json:"type"`
	Message        string           `json:"message"`
	Timestamp      time.Time        `json:"timestamp"`
	Severity       ActivitySeverity `json:"severity"`
	RepositoryName string           `json:"repositoryName,omitempty"`
}

// TrendDataPoint represents a data point for trend charts
type TrendDataPoint struct {
	Date             string  `json:"date"`
	CodeQuality      float64 `json:"codeQuality"`
	IssuesResolved   int     `json:"issuesResolved"`
	PerformanceScore float64 `json:"performanceScore"`
}