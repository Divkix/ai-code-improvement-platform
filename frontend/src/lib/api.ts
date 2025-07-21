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

	// Generic ping
	async ping(): Promise<{ message: string; service: string }> {
		return this.request<{ message: string; service: string }>('/api/ping');
	}
}

export const apiClient = new ApiClient();
export default apiClient;
