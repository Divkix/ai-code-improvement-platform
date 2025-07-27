// ABOUTME: API hooks and utilities for common operations using generated client
// ABOUTME: Provides easy-to-use functions for dashboard, repositories, and GitHub operations
import { apiClient } from './client';
import type {
	DashboardStats,
	ActivityItem,
	TrendDataPoint,
	Repository,
	RepositoryListResponse,
	GitHubRepositoriesResponse,
	GitHubRepositorySearchResponse,
	CreateRepositoryRequest,
	UpdateRepositoryRequest,
	GitHubOAuthRequest,
	HealthCheck
} from './index';

// Health endpoints
export async function getHealth(): Promise<HealthCheck> {
	const { data, error } = await apiClient.GET('/health');
	if (error) throw new Error(error.message || 'Health check failed');
	return data;
}

export async function getApiHealth(): Promise<HealthCheck> {
	const { data, error } = await apiClient.GET('/api/health');
	if (error) throw new Error(error.message || 'API health check failed');
	return data;
}

// Dashboard operations
export async function getDashboardStats(): Promise<DashboardStats> {
	const { data, error } = await apiClient.GET('/api/dashboard/stats');
	if (error) throw new Error(error.message || 'Failed to fetch dashboard stats');
	return data;
}

export async function getDashboardActivity(limit?: number): Promise<ActivityItem[]> {
	const { data, error } = await apiClient.GET('/api/dashboard/activity', {
		params: {
			query: limit ? { limit } : undefined
		}
	});
	if (error) throw new Error(error.message || 'Failed to fetch dashboard activity');
	return data;
}

export async function getDashboardTrends(days?: number): Promise<TrendDataPoint[]> {
	const { data, error } = await apiClient.GET('/api/dashboard/trends', {
		params: {
			query: days ? { days } : undefined
		}
	});
	if (error) throw new Error(error.message || 'Failed to fetch dashboard trends');
	return data;
}

// Repository operations
export async function getRepositories(params?: {
	limit?: number;
	offset?: number;
	status?: 'pending' | 'importing' | 'ready' | 'error';
}): Promise<RepositoryListResponse> {
	const { data, error } = await apiClient.GET('/api/repositories', {
		params: {
			query: params
		}
	});
	if (error) throw new Error(error.message || 'Failed to fetch repositories');
	return data;
}

export async function createRepository(repository: CreateRepositoryRequest): Promise<Repository> {
	const { data, error } = await apiClient.POST('/api/repositories', {
		body: repository
	});
	if (error) throw new Error(error.message || 'Failed to create repository');
	return data;
}

export async function getRepository(id: string): Promise<Repository> {
	const { data, error } = await apiClient.GET('/api/repositories/{id}', {
		params: {
			path: { id }
		}
	});
	if (error) throw new Error(error.message || 'Failed to fetch repository');
	return data;
}

export async function updateRepository(
	id: string,
	updates: UpdateRepositoryRequest
): Promise<Repository> {
	const { data, error } = await apiClient.PUT('/api/repositories/{id}', {
		params: {
			path: { id }
		},
		body: updates
	});
	if (error) throw new Error(error.message || 'Failed to update repository');
	return data;
}

export async function deleteRepository(id: string): Promise<void> {
	const { error } = await apiClient.DELETE('/api/repositories/{id}', {
		params: {
			path: { id }
		}
	});
	if (error) throw new Error(error.message || 'Failed to delete repository');
}

// GitHub operations
export async function getGitHubRepositories(page?: number): Promise<GitHubRepositoriesResponse> {
	const { data, error } = await apiClient.GET('/api/github/repositories', {
		params: {
			query: page ? { page } : undefined
		}
	});
	if (error) throw new Error(error.message || 'Failed to fetch GitHub repositories');
	return data;
}

export async function validateGitHubRepository(owner: string, repo: string) {
	const { data, error } = await apiClient.GET('/api/github/repositories/{owner}/{repo}/validate', {
		params: {
			path: { owner, repo }
		}
	});
	if (error) throw new Error(error.message || 'Failed to validate GitHub repository');
	return data;
}

export async function searchGitHubRepositories(
	query: string,
	limit?: number
): Promise<GitHubRepositorySearchResponse> {
	const { data, error } = await apiClient.GET('/api/github/repositories/search', {
		params: {
			query: { q: query, limit }
		}
	});
	if (error) throw new Error(error.message || 'Failed to search GitHub repositories');
	return data;
}

export async function getRecentGitHubRepositories(
	limit?: number
): Promise<GitHubRepositorySearchResponse> {
	const { data, error } = await apiClient.GET('/api/github/repositories/recent', {
		params: {
			query: limit ? { limit } : undefined
		}
	});
	if (error) throw new Error(error.message || 'Failed to fetch recent GitHub repositories');
	return data;
}

// GitHub OAuth
export async function getGitHubLoginUrl(redirectUri?: string) {
	const { data, error } = await apiClient.GET('/api/auth/github/login', {
		params: {
			query: redirectUri ? { redirect_uri: redirectUri } : undefined
		}
	});
	if (error) throw new Error(error.message || 'Failed to get GitHub login URL');
	return data;
}

export async function handleGitHubCallback(request: GitHubOAuthRequest) {
	const { data, error } = await apiClient.POST('/api/auth/github/callback', {
		body: request
	});
	if (error) throw new Error(error.message || 'GitHub OAuth callback failed');
	return data;
}

export async function disconnectGitHub() {
	const { data, error } = await apiClient.POST('/api/auth/github/disconnect');
	if (error) throw new Error(error.message || 'Failed to disconnect GitHub');
	return data;
}
