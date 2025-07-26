// ABOUTME: Tests for authentication store functionality
// ABOUTME: Covers login, logout, token validation, and localStorage persistence

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { get } from 'svelte/store';
import { authStore } from './auth';
import { mockLocalStorage, createMockResponse, createMockErrorResponse } from '$lib/test-utils';

// Mock the browser environment
vi.mock('$app/environment', () => ({
	browser: true
}));

// Mock the API client
vi.mock('../api', () => ({
	apiClient: {
		POST: vi.fn(),
		GET: vi.fn()
	},
	setAuthToken: vi.fn()
}));

describe('authStore', () => {
	let mockStorage: ReturnType<typeof mockLocalStorage>;
	let mockApiClient: any;

	beforeEach(async () => {
		// Setup localStorage mock
		mockStorage = mockLocalStorage();
		Object.defineProperty(global, 'localStorage', {
			value: mockStorage,
			writable: true
		});

		// Get the mocked API client
		const { apiClient } = await import('../api');
		mockApiClient = apiClient;

		// Clear all mocks
		vi.clearAllMocks();

		// Reset auth store to initial state
		authStore.logout();
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	describe('initial state', () => {
		it('should have correct initial state', () => {
			const state = get(authStore);
			expect(state.user).toBeNull();
			expect(state.token).toBeNull();
			expect(state.isAuthenticated).toBe(false);
			expect(state.isLoading).toBe(false); // After logout
		});
	});

	describe('init', () => {
		it('should initialize from localStorage when valid token exists', () => {
			const mockUser = { id: '1', email: 'test@example.com', name: 'Test User' };
			const mockToken = 'valid-token';

			mockStorage.setItem('auth_token', mockToken);
			mockStorage.setItem('auth_user', JSON.stringify(mockUser));

			// Mock successful token validation
			mockApiClient.GET.mockResolvedValueOnce({
				data: mockUser,
				error: null
			});

			authStore.init();

			const state = get(authStore);
			expect(state.user).toEqual(mockUser);
			expect(state.token).toBe(mockToken);
			expect(state.isAuthenticated).toBe(true);
		});

		it('should clear invalid token from localStorage', () => {
			mockStorage.setItem('auth_token', 'invalid-token');
			mockStorage.setItem('auth_user', 'invalid-json');

			authStore.init();

			expect(mockStorage.removeItem).toHaveBeenCalledWith('auth_token');
			expect(mockStorage.removeItem).toHaveBeenCalledWith('auth_user');

			const state = get(authStore);
			expect(state.isAuthenticated).toBe(false);
		});

		it('should handle missing localStorage data', () => {
			authStore.init();

			const state = get(authStore);
			expect(state.user).toBeNull();
			expect(state.token).toBeNull();
			expect(state.isAuthenticated).toBe(false);
			expect(state.isLoading).toBe(false);
		});
	});

	describe('login', () => {
		it('should login successfully with valid credentials', async () => {
			const mockUser = { id: '1', email: 'test@example.com', name: 'Test User' };
			const mockToken = 'auth-token';

			mockApiClient.POST.mockResolvedValueOnce({
				data: { user: mockUser, token: mockToken },
				error: null
			});

			const result = await authStore.login('test@example.com', 'password');

			expect(mockApiClient.POST).toHaveBeenCalledWith('/api/auth/login', {
				body: { email: 'test@example.com', password: 'password' }
			});

			expect(mockStorage.setItem).toHaveBeenCalledWith('auth_token', mockToken);
			expect(mockStorage.setItem).toHaveBeenCalledWith('auth_user', JSON.stringify(mockUser));

			const state = get(authStore);
			expect(state.user).toEqual(mockUser);
			expect(state.token).toBe(mockToken);
			expect(state.isAuthenticated).toBe(true);
			expect(state.isLoading).toBe(false);

			expect(result).toEqual({ user: mockUser, token: mockToken });
		});

		it('should handle login failure', async () => {
			mockApiClient.POST.mockResolvedValueOnce({
				data: null,
				error: { message: 'Invalid credentials' }
			});

			await expect(authStore.login('test@example.com', 'wrongpassword')).rejects.toThrow(
				'Invalid credentials'
			);

			const state = get(authStore);
			expect(state.isAuthenticated).toBe(false);
			expect(state.isLoading).toBe(false);
		});

		it('should handle network errors during login', async () => {
			mockApiClient.POST.mockRejectedValueOnce(new Error('Network error'));

			await expect(authStore.login('test@example.com', 'password')).rejects.toThrow(
				'Network error'
			);

			const state = get(authStore);
			expect(state.isAuthenticated).toBe(false);
			expect(state.isLoading).toBe(false);
		});
	});

	describe('logout', () => {
		it('should clear user data and localStorage', () => {
			// Set up authenticated state
			mockStorage.setItem('auth_token', 'token');
			mockStorage.setItem('auth_user', '{"id":"1"}');

			authStore.logout();

			expect(mockStorage.removeItem).toHaveBeenCalledWith('auth_token');
			expect(mockStorage.removeItem).toHaveBeenCalledWith('auth_user');

			const state = get(authStore);
			expect(state.user).toBeNull();
			expect(state.token).toBeNull();
			expect(state.isAuthenticated).toBe(false);
			expect(state.isLoading).toBe(false);
		});
	});

	describe('validateToken', () => {
		it('should validate token successfully', async () => {
			const mockUser = { id: '1', email: 'test@example.com', name: 'Test User' };

			mockApiClient.GET.mockResolvedValueOnce({
				data: mockUser,
				error: null
			});

			await authStore.validateToken();

			expect(mockApiClient.GET).toHaveBeenCalledWith('/api/auth/me');

			const state = get(authStore);
			expect(state.user).toEqual(mockUser);
			expect(state.isLoading).toBe(false);
		});

		it('should logout on token validation failure', async () => {
			mockApiClient.GET.mockResolvedValueOnce({
				data: null,
				error: { message: 'Token invalid' }
			});

			// Spy on console.error to avoid noise in test output
			const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

			await authStore.validateToken();

			const state = get(authStore);
			expect(state.isAuthenticated).toBe(false);
			expect(state.user).toBeNull();
			expect(state.token).toBeNull();

			consoleSpy.mockRestore();
		});
	});

	describe('updateUser', () => {
		it('should update user information', () => {
			const updatedUser = { id: '1', email: 'updated@example.com', name: 'Updated User' };

			authStore.updateUser(updatedUser);

			expect(mockStorage.setItem).toHaveBeenCalledWith('auth_user', JSON.stringify(updatedUser));

			const state = get(authStore);
			expect(state.user).toEqual(updatedUser);
		});
	});

	describe('setUser', () => {
		it('should set user information', () => {
			const newUser = { id: '2', email: 'new@example.com', name: 'New User' };

			authStore.setUser(newUser);

			expect(mockStorage.setItem).toHaveBeenCalledWith('auth_user', JSON.stringify(newUser));

			const state = get(authStore);
			expect(state.user).toEqual(newUser);
		});
	});

	describe('current getter', () => {
		it('should provide direct access to current state', () => {
			const currentState = authStore.current;
			expect(currentState).toBeDefined();
			expect(typeof currentState.isAuthenticated).toBe('boolean');
			expect(typeof currentState.isLoading).toBe('boolean');
		});
	});
});
