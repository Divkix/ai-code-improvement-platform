// ABOUTME: Tests for API client functionality and authorization handling
// ABOUTME: Covers API client setup, auth token management, and vector search methods

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { apiClient, setAuthToken, vectorSearchAPI } from './client';
import { mockLocalStorage, createMockResponse, createMockErrorResponse } from '$lib/test-utils';

// Mock openapi-fetch
vi.mock('openapi-fetch', () => {
	const mockPost = vi.fn();
	const mockGet = vi.fn();
	const mockUse = vi.fn();
	const mockEject = vi.fn();

	return {
		default: vi.fn(() => ({
			POST: mockPost,
			GET: mockGet,
			use: mockUse,
			eject: mockEject
		}))
	};
});

describe('API Client', () => {
	let mockStorage: ReturnType<typeof mockLocalStorage>;
	let mockPost: any;
	let mockGet: any;
	let mockUse: any;
	let mockEject: any;

	beforeEach(async () => {
		// Setup localStorage mock
		mockStorage = mockLocalStorage();
		Object.defineProperty(global, 'localStorage', {
			value: mockStorage,
			writable: true
		});

		// Mock fetch for getPipelineStats
		global.fetch = vi.fn();

		// Get mock functions from the mocked module
		const { apiClient } = await import('./client');
		mockPost = (apiClient as any).POST;
		mockGet = (apiClient as any).GET;
		mockUse = (apiClient as any).use;
		mockEject = (apiClient as any).eject;

		// Clear all mocks
		vi.clearAllMocks();
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	describe('setAuthToken', () => {
		it('should set authorization header when token provided', () => {
			const token = 'test-token';

			setAuthToken(token);

			expect(mockEject).toHaveBeenCalled();
			expect(mockUse).toHaveBeenCalledWith({
				onRequest: expect.any(Function)
			});
		});

		it('should clear middleware when token is null', () => {
			setAuthToken(null);

			expect(mockEject).toHaveBeenCalled();
			expect(mockUse).not.toHaveBeenCalled();
		});

		it('should properly set Authorization header in request', () => {
			const token = 'test-token';
			setAuthToken(token);

			// Get the middleware function
			const middlewareCall = mockUse.mock.calls[0][0];
			const mockRequest = {
				headers: {
					set: vi.fn()
				}
			};

			middlewareCall.onRequest({ request: mockRequest });

			expect(mockRequest.headers.set).toHaveBeenCalledWith('Authorization', `Bearer ${token}`);
		});
	});

	describe('vectorSearchAPI', () => {
		describe('vectorSearch', () => {
			it('should perform vector search successfully', async () => {
				const mockSearchResults = {
					chunks: [
						{
							id: 'chunk-1',
							content: 'test content',
							score: 0.95,
							metadata: {}
						}
					],
					total: 1
				};

				mockPost.mockResolvedValueOnce({
					data: mockSearchResults,
					error: null
				});

				const result = await vectorSearchAPI.vectorSearch('test query', 'repo-1', 10, 0);

				expect(mockPost).toHaveBeenCalledWith('/api/search/vector', {
					body: {
						query: 'test query',
						repositoryId: 'repo-1',
						limit: 10,
						offset: 0
					}
				});

				expect(result).toEqual(mockSearchResults);
			});

			it('should use default values for optional parameters', async () => {
				const mockSearchResults = { chunks: [], total: 0 };

				mockPost.mockResolvedValueOnce({
					data: mockSearchResults,
					error: null
				});

				await vectorSearchAPI.vectorSearch('test query');

				expect(mockPost).toHaveBeenCalledWith('/api/search/vector', {
					body: {
						query: 'test query',
						repositoryId: undefined,
						limit: 20,
						offset: 0
					}
				});
			});

			it('should throw error when API returns error', async () => {
				mockPost.mockResolvedValueOnce({
					data: null,
					error: { error: 'Search failed' }
				});

				await expect(vectorSearchAPI.vectorSearch('test query')).rejects.toThrow('Search failed');
			});

			it('should throw generic error when no error message provided', async () => {
				mockPost.mockResolvedValueOnce({
					data: null,
					error: {}
				});

				await expect(vectorSearchAPI.vectorSearch('test query')).rejects.toThrow(
					'Vector search failed'
				);
			});
		});

		describe('hybridSearch', () => {
			it('should perform hybrid search successfully', async () => {
				const mockSearchResults = {
					chunks: [
						{
							id: 'chunk-1',
							content: 'test content',
							score: 0.95,
							metadata: {}
						}
					],
					total: 1
				};

				mockPost.mockResolvedValueOnce({
					data: mockSearchResults,
					error: null
				});

				const result = await vectorSearchAPI.hybridSearch('test query', 'repo-1', 0.8, 10, 0);

				expect(mockPost).toHaveBeenCalledWith('/api/search/hybrid', {
					body: {
						query: 'test query',
						repositoryId: 'repo-1',
						vectorWeight: 0.8,
						textWeight: expect.closeTo(0.2, 5), // Use closeTo for floating point comparison
						limit: 10,
						offset: 0
					}
				});

				expect(result).toEqual(mockSearchResults);
			});

			it('should use default vector weight of 0.7', async () => {
				const mockSearchResults = { chunks: [], total: 0 };

				mockPost.mockResolvedValueOnce({
					data: mockSearchResults,
					error: null
				});

				await vectorSearchAPI.hybridSearch('test query');

				expect(mockPost).toHaveBeenCalledWith('/api/search/hybrid', {
					body: {
						query: 'test query',
						repositoryId: undefined,
						vectorWeight: 0.7,
						textWeight: expect.closeTo(0.3, 5), // Use closeTo for floating point comparison
						limit: 20,
						offset: 0
					}
				});
			});

			it('should throw error when API returns error', async () => {
				mockPost.mockResolvedValueOnce({
					data: null,
					error: { error: 'Hybrid search failed' }
				});

				await expect(vectorSearchAPI.hybridSearch('test query')).rejects.toThrow(
					'Hybrid search failed'
				);
			});
		});

		describe('getSimilarChunks', () => {
			it('should get similar chunks successfully', async () => {
				const mockSimilarChunks = {
					chunks: [
						{
							id: 'similar-chunk-1',
							content: 'similar content',
							score: 0.85,
							metadata: {}
						}
					],
					total: 1
				};

				mockGet.mockResolvedValueOnce({
					data: mockSimilarChunks,
					error: null
				});

				const result = await vectorSearchAPI.getSimilarChunks('chunk-1', 5);

				expect(mockGet).toHaveBeenCalledWith('/api/search/similar/{chunkId}', {
					params: {
						path: { chunkId: 'chunk-1' },
						query: { limit: 5 }
					}
				});

				expect(result).toEqual(mockSimilarChunks);
			});

			it('should throw error when API returns error', async () => {
				mockGet.mockResolvedValueOnce({
					data: null,
					error: { error: 'Similar chunks search failed' }
				});

				await expect(vectorSearchAPI.getSimilarChunks('chunk-1')).rejects.toThrow(
					'Similar chunks search failed'
				);
			});
		});

		describe('triggerRepositoryEmbedding', () => {
			it('should trigger repository embedding successfully', async () => {
				const mockResponse = {
					message: 'Embedding started',
					jobId: 'job-123'
				};

				mockPost.mockResolvedValueOnce({
					data: mockResponse,
					error: null
				});

				const result = await vectorSearchAPI.triggerRepositoryEmbedding('repo-1');

				expect(mockPost).toHaveBeenCalledWith('/api/repositories/{id}/embed', {
					params: {
						path: { id: 'repo-1' }
					}
				});

				expect(result).toEqual(mockResponse);
			});

			it('should throw error when triggering fails', async () => {
				mockPost.mockResolvedValueOnce({
					data: null,
					error: { error: 'Failed to trigger embedding' }
				});

				await expect(vectorSearchAPI.triggerRepositoryEmbedding('repo-1')).rejects.toThrow(
					'Failed to trigger embedding'
				);
			});
		});

		describe('getRepositoryEmbeddingStatus', () => {
			it('should get embedding status successfully', async () => {
				const mockStatus = {
					status: 'processing',
					progress: 0.5,
					totalChunks: 100,
					processedChunks: 50
				};

				mockGet.mockResolvedValueOnce({
					data: mockStatus,
					error: null
				});

				const result = await vectorSearchAPI.getRepositoryEmbeddingStatus('repo-1');

				expect(mockGet).toHaveBeenCalledWith('/api/repositories/{id}/embedding-status', {
					params: {
						path: { id: 'repo-1' }
					}
				});

				expect(result).toEqual(mockStatus);
			});

			it('should throw error when getting status fails', async () => {
				mockGet.mockResolvedValueOnce({
					data: null,
					error: { error: 'Failed to get embedding status' }
				});

				await expect(vectorSearchAPI.getRepositoryEmbeddingStatus('repo-1')).rejects.toThrow(
					'Failed to get embedding status'
				);
			});
		});

		describe('getPipelineStats', () => {
			it('should get pipeline stats successfully', async () => {
				const mockStats = {
					pending: 10,
					processing: 5,
					completed: 100,
					failed: 2
				};

				(global.fetch as any).mockResolvedValueOnce({
					ok: true,
					json: () => Promise.resolve(mockStats)
				});

				const result = await vectorSearchAPI.getPipelineStats();

				expect(global.fetch).toHaveBeenCalledWith(
					'http://localhost:8080/api/embedding/pipeline-stats',
					{
						headers: {
							'Content-Type': 'application/json'
						}
					}
				);

				expect(result).toEqual(mockStats);
			});

			it('should include auth token when available', async () => {
				mockStorage.setItem('auth_token', 'test-token');

				const mockStats = {
					pending: 10,
					processing: 5,
					completed: 100,
					failed: 2
				};

				(global.fetch as any).mockResolvedValueOnce({
					ok: true,
					json: () => Promise.resolve(mockStats)
				});

				await vectorSearchAPI.getPipelineStats();

				expect(global.fetch).toHaveBeenCalledWith(
					'http://localhost:8080/api/embedding/pipeline-stats',
					{
						headers: {
							'Content-Type': 'application/json',
							Authorization: 'Bearer test-token'
						}
					}
				);
			});

			it('should throw error when fetch fails', async () => {
				(global.fetch as any).mockResolvedValueOnce({
					ok: false,
					status: 500
				});

				await expect(vectorSearchAPI.getPipelineStats()).rejects.toThrow(
					'Failed to fetch pipeline stats: 500'
				);
			});
		});

		describe('importRepository', () => {
			it('should trigger repository import successfully', async () => {
				const mockResponse = {
					message: 'Repository import started',
					jobId: 'import-job-123'
				};

				mockPost.mockResolvedValueOnce({
					data: mockResponse,
					error: null
				});

				const result = await vectorSearchAPI.importRepository('repo-1');

				expect(mockPost).toHaveBeenCalledWith('/api/repositories/{id}/import', {
					params: {
						path: { id: 'repo-1' }
					}
				});

				expect(result).toEqual(mockResponse);
			});

			it('should throw error when import fails', async () => {
				mockPost.mockResolvedValueOnce({
					data: null,
					error: { error: 'Failed to trigger repository import' }
				});

				await expect(vectorSearchAPI.importRepository('repo-1')).rejects.toThrow(
					'Failed to trigger repository import'
				);
			});
		});
	});
});
