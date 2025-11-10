// ABOUTME: Generated API client using openapi-fetch and generated types
// ABOUTME: Provides type-safe API calls with automatic request/response validation
import createClient from 'openapi-fetch';
import type { paths, operations } from './types';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

// Create the API client with the generated types
export const apiClient = createClient<paths>({
	baseUrl: API_BASE_URL
});

/**
 * Validates JWT token format and expiration
 * @param token - JWT token to validate
 * @returns true if token is valid, false otherwise
 */
function isValidJWT(token: string): boolean {
	try {
		// Check JWT format: should have 3 parts separated by dots
		const parts = token.split('.');
		if (parts.length !== 3) {
			console.warn('Invalid JWT format: token does not have 3 parts');
			return false;
		}

		// Decode the payload (second part)
		const payload = JSON.parse(atob(parts[1]));

		// Check if token has expiration claim
		if (!payload.exp) {
			console.warn('JWT missing expiration claim');
			return false;
		}

		// Check if token is expired (exp is in seconds, Date.now() is in milliseconds)
		const currentTime = Math.floor(Date.now() / 1000);
		if (payload.exp < currentTime) {
			console.warn('JWT token has expired');
			return false;
		}

		return true;
	} catch (error) {
		console.error('JWT validation error:', error);
		return false;
	}
}

// Set up auth token interceptor
export function setAuthToken(token: string | null) {
	if (token) {
		// Validate token before using it
		if (!isValidJWT(token)) {
			console.warn('Attempting to set invalid JWT token, removing from localStorage');
			// Remove invalid token from localStorage
			if (typeof localStorage !== 'undefined') {
				localStorage.removeItem('auth_token');
			}
			// Clear all middleware when token is invalid
			apiClient.eject();
			return;
		}

		// Clear existing middleware and add new one with token
		apiClient.eject();
		apiClient.use({
			onRequest({ request }) {
				request.headers.set('Authorization', `Bearer ${token}`);
			}
		});
	} else {
		// Clear all middleware when no token
		apiClient.eject();
	}
}

// Initialize auth token from localStorage if available
if (typeof localStorage !== 'undefined') {
	const storedToken = localStorage.getItem('auth_token');
	if (storedToken) {
		// Validate token before using it
		if (isValidJWT(storedToken)) {
			setAuthToken(storedToken);
		} else {
			// Remove invalid token from localStorage
			console.warn('Stored JWT token is invalid or expired, removing');
			localStorage.removeItem('auth_token');
			localStorage.removeItem('auth_user');
		}
	}
}

// Vector Search API methods
export const vectorSearchAPI: {
	vectorSearch(
		query: string,
		repositoryId?: string,
		limit?: number,
		offset?: number
	): Promise<operations['vectorSearch']['responses']['200']['content']['application/json']>;
	hybridSearch(
		query: string,
		repositoryId?: string,
		vectorWeight?: number,
		limit?: number,
		offset?: number
	): Promise<operations['hybridSearch']['responses']['200']['content']['application/json']>;
	getSimilarChunks(
		chunkId: string,
		limit?: number
	): Promise<operations['findSimilarChunks']['responses']['200']['content']['application/json']>;
	triggerRepositoryEmbedding(
		repositoryId: string
	): Promise<
		operations['triggerRepositoryEmbedding']['responses']['202']['content']['application/json']
	>;
	getRepositoryEmbeddingStatus(
		repositoryId: string
	): Promise<
		operations['getRepositoryEmbeddingStatus']['responses']['200']['content']['application/json']
	>;

	// Embedding pipeline stats (custom endpoint)
	getPipelineStats(): Promise<{
		pending: number;
		processing: number;
		completed: number;
		failed: number;
	}>;
	importRepository(
		repositoryId: string
	): Promise<
		operations['triggerRepositoryImport']['responses']['202']['content']['application/json']
	>;
} = {
	// Perform semantic vector search
	async vectorSearch(query: string, repositoryId?: string, limit?: number, offset?: number) {
		const { data, error } = await apiClient.POST('/api/search/vector', {
			body: {
				query,
				repositoryId,
				limit: limit || 20,
				offset: offset || 0
			}
		});

		if (error) {
			throw new Error(error.error || 'Vector search failed');
		}

		return data;
	},

	// Perform hybrid search (text + vector)
	async hybridSearch(
		query: string,
		repositoryId?: string,
		vectorWeight?: number,
		limit?: number,
		offset?: number
	) {
		const { data, error } = await apiClient.POST('/api/search/hybrid', {
			body: {
				query,
				repositoryId,
				vectorWeight: vectorWeight || 0.7,
				textWeight: 1 - (vectorWeight || 0.7),
				limit: limit || 20,
				offset: offset || 0
			}
		});

		if (error) {
			throw new Error(error.error || 'Hybrid search failed');
		}

		return data;
	},

	// Find similar code chunks to a specific chunk
	async getSimilarChunks(chunkId: string, limit?: number) {
		const { data, error } = await apiClient.GET('/api/search/similar/{chunkId}', {
			params: {
				path: { chunkId },
				query: { limit: limit }
			}
		});

		if (error) {
			throw new Error(error.error || 'Similar chunks search failed');
		}

		return data;
	},

	// Trigger embedding processing for a repository
	async triggerRepositoryEmbedding(repositoryId: string) {
		const { data, error } = await apiClient.POST('/api/repositories/{id}/embed', {
			params: {
				path: { id: repositoryId }
			}
		});

		if (error) {
			throw new Error(error.error || 'Failed to trigger embedding');
		}

		return data;
	},

	// Get embedding status for a repository
	async getRepositoryEmbeddingStatus(repositoryId: string) {
		const { data, error } = await apiClient.GET('/api/repositories/{id}/embedding-status', {
			params: {
				path: { id: repositoryId }
			}
		});

		if (error) {
			throw new Error(error.error || 'Failed to get embedding status');
		}

		return data;
	},

	// Get embedding pipeline stats. This endpoint is protected, so we attach
	// the JWT from localStorage if it exists. We use fetch() here because the
	// generated openapi-fetch types currently mis-handle this specific path.
	async getPipelineStats() {
		const headers: Record<string, string> = {
			'Content-Type': 'application/json'
		};

		if (typeof localStorage !== 'undefined') {
			const token = localStorage.getItem('auth_token');
			if (token && isValidJWT(token)) {
				headers['Authorization'] = `Bearer ${token}`;
			} else if (token) {
				// Token exists but is invalid, remove it
				console.warn('Stored JWT token is invalid, removing');
				localStorage.removeItem('auth_token');
				localStorage.removeItem('auth_user');
			}
		}

		const res = await fetch(`${API_BASE_URL}/api/embedding/pipeline-stats`, {
			headers
		});

		if (!res.ok) {
			throw new Error(`Failed to fetch pipeline stats: ${res.status}`);
		}

		return (await res.json()) as {
			pending: number;
			processing: number;
			completed: number;
			failed: number;
		};
	},

	// Manually trigger repository import for stuck/pending repositories
	async importRepository(repositoryId: string) {
		const { data, error } = await apiClient.POST('/api/repositories/{id}/import', {
			params: {
				path: { id: repositoryId }
			}
		});

		if (error) {
			throw new Error(error.error || 'Failed to trigger repository import');
		}

		return data;
	}
};

// Export the client as default
export default apiClient;
