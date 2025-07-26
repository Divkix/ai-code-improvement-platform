// ABOUTME: Authentication store using Svelte stores for SSR compatibility with user state management
// ABOUTME: Handles token storage, user data, and authentication state throughout the app

import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import { apiClient, setAuthToken, type User } from '../api';

export interface AuthState {
	user: User | null;
	token: string | null;
	isAuthenticated: boolean;
	isLoading: boolean;
}

// Set isLoading to true initially to prevent premature redirects on page load.
const initialState: AuthState = {
	user: null,
	token: null,
	isAuthenticated: false,
	isLoading: true
};

// Define a clear state for when the user is logged out and loading is complete.
const loggedOutState: AuthState = {
	user: null,
	token: null,
	isAuthenticated: false,
	isLoading: false
};

// Create the auth store - using Svelte stores for SSR compatibility
function createAuthStore() {
	const { subscribe, set } = writable<AuthState>(initialState);

	// Internal state tracking (SSR-safe)
	let _currentState: AuthState = initialState;

	// Subscribe to keep internal state in sync
	subscribe((value) => {
		_currentState = value;
	});

	// Helper to update state
	const updateState = (newState: AuthState) => {
		_currentState = newState;
		set(newState);
	};

	return {
		subscribe,

		// Direct state access (Svelte 5 style)
		get current() {
			return _currentState;
		},

		// Initialize auth state from localStorage
		init: () => {
			if (!browser) {
				updateState(loggedOutState);
				return;
			}

			const token = localStorage.getItem('auth_token');
			const userStr = localStorage.getItem('auth_user');

			if (token && userStr) {
				try {
					const user = JSON.parse(userStr);
					setAuthToken(token);
					updateState({
						..._currentState,
						user,
						token,
						isAuthenticated: true
					});

					// Validate the token against the backend. This will set isLoading to false.
					authStore.validateToken();
				} catch (error) {
					console.error('Failed to parse stored user data:', error);
					authStore.logout();
				}
			} else {
				// No token found, so the user is not authenticated.
				updateState(loggedOutState);
			}
		},

		// Login user
		login: async (email: string, password: string) => {
			updateState({ ..._currentState, isLoading: true });

			try {
				const { data, error } = await apiClient.POST('/api/auth/login', {
					body: { email, password }
				});

				if (error) {
					throw new Error(error.message || 'Login failed');
				}

				// Store token and user data
				if (browser) {
					localStorage.setItem('auth_token', data.token);
					localStorage.setItem('auth_user', JSON.stringify(data.user));
					setAuthToken(data.token);
				}

				updateState({
					..._currentState,
					user: data.user,
					token: data.token,
					isAuthenticated: true,
					isLoading: false
				});

				return data;
			} catch (error) {
				updateState({ ..._currentState, isLoading: false });
				throw error;
			}
		},

		// Logout user
		logout: () => {
			if (browser) {
				localStorage.removeItem('auth_token');
				localStorage.removeItem('auth_user');
				setAuthToken(null);
			}
			updateState(loggedOutState);
		},

		// Validate current token
		validateToken: async () => {
			try {
				const { data, error } = await apiClient.GET('/api/auth/me');

				if (error) {
					throw new Error(error.message || 'Token validation failed');
				}

				// Token is valid, update user info and set loading to false.
				updateState({ ..._currentState, user: data, isLoading: false });
			} catch (error) {
				console.error('Token validation failed:', error);
				// Token is invalid, log the user out.
				authStore.logout();
			}
		},

		// Update user information
		updateUser: (user: User) => {
			if (browser) {
				localStorage.setItem('auth_user', JSON.stringify(user));
			}

			updateState({ ..._currentState, user });
		},

		// Set user information (used by GitHub OAuth callbacks)
		setUser: (user: User) => {
			if (browser) {
				localStorage.setItem('auth_user', JSON.stringify(user));
			}

			updateState({ ..._currentState, user });
		}
	};
}

export const authStore = createAuthStore();
