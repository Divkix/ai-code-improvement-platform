// ABOUTME: Modern API client using generated OpenAPI types and openapi-fetch
// ABOUTME: Replaces manual API client with type-safe generated version
export { default as apiClient, setAuthToken } from './client';
import type { paths, components } from './types';
export type { paths, components };

// Re-export specific types for convenience
export type User = components['schemas']['User'];
export type AuthResponse = components['schemas']['AuthResponse'];
export type HealthCheck = components['schemas']['HealthCheck'];
export type DashboardStats = components['schemas']['DashboardStats'];
export type ActivityItem = components['schemas']['ActivityItem'];
export type TrendDataPoint = components['schemas']['TrendDataPoint'];
export type Repository = components['schemas']['Repository'];
export type CreateRepositoryRequest = components['schemas']['CreateRepositoryRequest'];
export type UpdateRepositoryRequest = components['schemas']['UpdateRepositoryRequest'];
export type GitHubRepository = components['schemas']['GitHubRepository'];
export type GitHubRepositoriesResponse = components['schemas']['GitHubRepositoriesResponse'];
export type GitHubRepositorySearchResponse =
	components['schemas']['GitHubRepositorySearchResponse'];
export type RepositoryListResponse = components['schemas']['RepositoryListResponse'];
export type ApiError = components['schemas']['Error'];
export type GitHubOAuthRequest = components['schemas']['GitHubOAuthRequest'];
export type LoginRequest = components['schemas']['LoginRequest'];

// Export some commonly used request/response types
export type LoginUserJSONRequestBody =
	paths['/api/auth/login']['post']['requestBody']['content']['application/json'];
export type GetRepositoriesParams = paths['/api/repositories']['get']['parameters']['query'];
export type GetDashboardActivityParams =
	paths['/api/dashboard/activity']['get']['parameters']['query'];
export type GetDashboardTrendsParams = paths['/api/dashboard/trends']['get']['parameters']['query'];
