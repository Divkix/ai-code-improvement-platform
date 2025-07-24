// ABOUTME: Generated API client using openapi-fetch and generated types
// ABOUTME: Provides type-safe API calls with automatic request/response validation
import createClient from 'openapi-fetch';
import type { paths, operations } from './types';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

// Create the API client with the generated types
export const apiClient = createClient<paths>({
	baseUrl: API_BASE_URL
});

// Set up auth token interceptor
export function setAuthToken(token: string | null) {
	if (token) {
		apiClient.use({
			onRequest({ request }) {
				request.headers.set('Authorization', `Bearer ${token}`);
			}
		});
	}
}

// Initialize auth token from localStorage if available
if (typeof localStorage !== 'undefined') {
	const storedToken = localStorage.getItem('auth_token');
	if (storedToken) {
		setAuthToken(storedToken);
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

	// Get embedding pipeline stats
	async getPipelineStats() {
		const res = await fetch(`${API_BASE_URL}/api/embedding/pipeline-stats`, {
			headers: {
				'Content-Type': 'application/json'
			}
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
