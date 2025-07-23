// ABOUTME: Search API client for handling search requests and responses
// ABOUTME: Provides centralized search functionality with error handling and type safety

import type { 
    SearchRequest, 
    SearchResponse, 
    QuickSearchResponse, 
    SearchSuggestionsResponse, 
    LanguagesResponse,
    RecentChunksResponse
} from './search-types';

class SearchClient {
    private baseUrl: string;

    constructor(baseUrl = '/api') {
        this.baseUrl = baseUrl;
    }

    /**
     * Perform global search across all repositories
     */
    async search(request: SearchRequest): Promise<SearchResponse> {
        const response = await fetch(`${this.baseUrl}/search`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(request),
        });
        
        if (!response.ok) {
            throw new Error(`Search failed: ${response.status} ${response.statusText}`);
        }
        
        return response.json();
    }

    /**
     * Search within a specific repository
     */
    async searchRepository(repositoryId: string, request: SearchRequest): Promise<SearchResponse> {
        const response = await fetch(`${this.baseUrl}/repositories/${repositoryId}/search`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(request),
        });
        
        if (!response.ok) {
            throw new Error(`Repository search failed: ${response.status} ${response.statusText}`);
        }
        
        return response.json();
    }

    /**
     * Perform quick search (lightweight results)
     */
    async quickSearch(query: string, limit = 5): Promise<QuickSearchResponse> {
        const params = new URLSearchParams({
            query,
            limit: limit.toString()
        });
        
        const response = await fetch(`${this.baseUrl}/search/quick?${params}`);
        
        if (!response.ok) {
            throw new Error(`Quick search failed: ${response.status} ${response.statusText}`);
        }
        
        return response.json();
    }

    /**
     * Get search suggestions based on partial query
     */
    async getSuggestions(query: string, limit = 10): Promise<SearchSuggestionsResponse> {
        const params = new URLSearchParams({
            query,
            limit: limit.toString()
        });
        
        const response = await fetch(`${this.baseUrl}/search/suggestions?${params}`);
        
        if (!response.ok) {
            throw new Error(`Get suggestions failed: ${response.status} ${response.statusText}`);
        }
        
        return response.json();
    }

    /**
     * Get available programming languages
     */
    async getLanguages(): Promise<LanguagesResponse> {
        const response = await fetch(`${this.baseUrl}/search/languages`);
        
        if (!response.ok) {
            throw new Error(`Get languages failed: ${response.status} ${response.statusText}`);
        }
        
        return response.json();
    }

    /**
     * Get available languages for a specific repository
     */
    async getRepositoryLanguages(repositoryId: string): Promise<LanguagesResponse> {
        const response = await fetch(`${this.baseUrl}/repositories/${repositoryId}/search/languages`);
        
        if (!response.ok) {
            throw new Error(`Get repository languages failed: ${response.status} ${response.statusText}`);
        }
        
        return response.json();
    }

    /**
     * Get recent code chunks
     */
    async getRecentChunks(limit = 10): Promise<RecentChunksResponse> {
        const params = new URLSearchParams({
            limit: limit.toString()
        });
        
        const response = await fetch(`${this.baseUrl}/search/recent?${params}`);
        
        if (!response.ok) {
            throw new Error(`Get recent chunks failed: ${response.status} ${response.statusText}`);
        }
        
        return response.json();
    }

    /**
     * Build URL search parameters from SearchRequest
     */
    private buildSearchParams(request: SearchRequest): URLSearchParams {
        const params = new URLSearchParams();
        
        params.set('query', request.query);
        
        if (request.repositoryId) {
            params.set('repositoryId', request.repositoryId);
        }
        
        if (request.language) {
            params.set('language', request.language);
        }
        
        if (request.fileType) {
            params.set('fileType', request.fileType);
        }
        
        if (request.limit !== undefined) {
            params.set('limit', request.limit.toString());
        }
        
        if (request.offset !== undefined) {
            params.set('offset', request.offset.toString());
        }
        
        return params;
    }
}

// Export a singleton instance
export const searchClient = new SearchClient();

// Also export the class for custom instances
export { SearchClient };