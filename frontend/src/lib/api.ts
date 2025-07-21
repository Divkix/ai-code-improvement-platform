// ABOUTME: API client utilities for communicating with the backend
// ABOUTME: Provides typed interfaces and error handling for all API endpoints

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export interface ApiError {
	error: string;
	message: string;
}

export interface HealthCheck {
	status: 'healthy' | 'degraded' | 'unhealthy';
	services: {
		mongodb: 'connected' | 'disconnected';
		qdrant: 'connected' | 'disconnected';
	};
	timestamp: string;
}

export interface User {
	id: string;
	email: string;
	name: string;
	createdAt: string;
}

export interface LoginRequest {
	email: string;
	password: string;
}

export interface AuthResponse {
	token: string;
	user: User;
}

export interface DashboardStats {
	totalRepositories: number;
	codeChunksProcessed: number;
	avgResponseTime: number;
	costSavingsMonthly: number;
	issuesPreventedMonthly: number;
	developerHoursReclaimed: number;
}

export interface ActivityItem {
	id: string;
	type: 'repository_imported' | 'analysis_completed' | 'issue_detected' | 'optimization_found';
	message: string;
	timestamp: string;
	severity: 'info' | 'warning' | 'error' | 'success';
	repositoryName?: string;
}

export interface TrendDataPoint {
	date: string;
	codeQuality: number;
	issuesResolved: number;
	performanceScore: number;
}

export interface Repository {
	id: string;
	userId: string;
	githubRepoId?: number;
	owner: string;
	name: string;
	fullName: string;
	description?: string;
	primaryLanguage?: string;
	isPrivate: boolean;
	indexedAt?: string;
	lastSyncedAt?: string;
	status: 'pending' | 'importing' | 'ready' | 'error';
	importProgress: number;
	stats?: {
		totalFiles: number;
		totalLines: number;
		languages?: Record<string, number>;
		lastCommitDate?: string;
	};
	createdAt: string;
	updatedAt: string;
}

export interface CreateRepositoryRequest {
	name: string;
	owner: string;
	fullName: string;
	description?: string;
	githubRepoId?: number;
	primaryLanguage?: string;
	isPrivate: boolean;
}

export interface CreateRepositoryFromUrlRequest {
	githubUrl: string;
}

export interface UpdateRepositoryRequest {
	name?: string;
	description?: string;
	primaryLanguage?: string;
}

export interface RepositoryListResponse {
	repositories: Repository[];
	total: number;
}

export interface RepositoryStats {
	repositoryId: string;
	totalFiles: number;
	totalLines: number;
	languages: Record<string, number>;
	lastCommitDate?: string;
	codeChunks: number;
	avgComplexity: number;
}

class ApiClient {
	private baseUrl: string;

	constructor(baseUrl: string = API_BASE_URL) {
		this.baseUrl = baseUrl;
	}

	private async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
		const url = `${this.baseUrl}${endpoint}`;

		const config: RequestInit = {
			headers: {
				'Content-Type': 'application/json',
				...options.headers
			},
			...options
		};

		// Add auth token if available
		const token = localStorage.getItem('auth_token');
		if (token) {
			config.headers = {
				...config.headers,
				Authorization: `Bearer ${token}`
			};
		}

		try {
			const response = await fetch(url, config);

			if (!response.ok) {
				const error: ApiError = await response.json();
				throw new Error(error.message || `HTTP ${response.status}`);
			}

			// Handle 204 No Content responses (like DELETE operations)
			if (response.status === 204) {
				return null as T;
			}

			return await response.json();
		} catch (error) {
			if (error instanceof Error) {
				throw error;
			}
			throw new Error('Network error');
		}
	}

	// Health checks
	async getHealth(): Promise<HealthCheck> {
		return this.request<HealthCheck>('/health');
	}

	async getApiHealth(): Promise<HealthCheck> {
		return this.request<HealthCheck>('/api/health');
	}

	// Authentication
	async login(email: string, password: string): Promise<AuthResponse> {
		const requestBody: LoginRequest = { email, password };
		return this.request<AuthResponse>('/api/auth/login', {
			method: 'POST',
			body: JSON.stringify(requestBody)
		});
	}

	async getCurrentUser(): Promise<User> {
		return this.request<User>('/api/auth/me');
	}

	// Dashboard
	async getDashboardStats(): Promise<DashboardStats> {
		return this.request<DashboardStats>('/api/dashboard/stats');
	}

	async getDashboardActivity(limit?: number): Promise<ActivityItem[]> {
		const query = limit ? `?limit=${limit}` : '';
		return this.request<ActivityItem[]>(`/api/dashboard/activity${query}`);
	}

	async getDashboardTrends(days?: number): Promise<TrendDataPoint[]> {
		const query = days ? `?days=${days}` : '';
		return this.request<TrendDataPoint[]>(`/api/dashboard/trends${query}`);
	}

	// Repositories
	async getRepositories(limit?: number, offset?: number, status?: string): Promise<RepositoryListResponse> {
		const params = new URLSearchParams();
		if (limit) params.append('limit', limit.toString());
		if (offset) params.append('offset', offset.toString());
		if (status) params.append('status', status);
		
		const query = params.toString() ? `?${params.toString()}` : '';
		return this.request<RepositoryListResponse>(`/api/repositories${query}`);
	}

	async createRepository(data: CreateRepositoryRequest): Promise<Repository> {
		return this.request<Repository>('/api/repositories', {
			method: 'POST',
			body: JSON.stringify(data)
		});
	}

	async getRepository(id: string): Promise<Repository> {
		return this.request<Repository>(`/api/repositories/${id}`);
	}

	async updateRepository(id: string, data: UpdateRepositoryRequest): Promise<Repository> {
		return this.request<Repository>(`/api/repositories/${id}`, {
			method: 'PUT',
			body: JSON.stringify(data)
		});
	}

	async deleteRepository(id: string): Promise<void> {
		await this.request<void>(`/api/repositories/${id}`, {
			method: 'DELETE'
		});
	}

	async getRepositoryStats(id: string): Promise<RepositoryStats> {
		return this.request<RepositoryStats>(`/api/repositories/${id}/stats`);
	}

	// Generic ping
	async ping(): Promise<{ message: string; service: string }> {
		return this.request<{ message: string; service: string }>('/api/ping');
	}
}

export const apiClient = new ApiClient();
export default apiClient;
