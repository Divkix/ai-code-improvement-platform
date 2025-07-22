// ABOUTME: Authentication store using Svelte stores for user state management
// ABOUTME: Handles token storage, user data, and authentication state throughout the app

import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import apiClient from '../api';

export interface User {
	id: string;
	email: string;
	name: string;
	githubConnected: boolean;
	githubUsername?: string;
	createdAt: string;
}

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

// Create the auth store
function createAuthStore() {
	const { subscribe, set, update } = writable<AuthState>(initialState);

	return {
		subscribe,

		// Initialize auth state from localStorage
		init: () => {
			if (!browser) {
				set(loggedOutState);
				return;
			}

			const token = localStorage.getItem('auth_token');
			const userStr = localStorage.getItem('auth_user');

			if (token && userStr) {
				try {
					const user = JSON.parse(userStr);
					update((state) => ({
						...state,
						user,
						token,
						isAuthenticated: true
					}));

					// Validate the token against the backend. This will set isLoading to false.
					authStore.validateToken();
				} catch (error) {
					console.error('Failed to parse stored user data:', error);
					authStore.logout();
				}
			} else {
				// No token found, so the user is not authenticated.
				set(loggedOutState);
			}
		},

		// Login user
		login: async (email: string, password: string) => {
			update((state) => ({ ...state, isLoading: true }));

			try {
				const response = await apiClient.login(email, password);

				// Store token and user data
				if (browser) {
					localStorage.setItem('auth_token', response.token);
					localStorage.setItem('auth_user', JSON.stringify(response.user));
				}

				update((state) => ({
					...state,
					user: response.user,
					token: response.token,
					isAuthenticated: true,
					isLoading: false
				}));

				return response;
			} catch (error) {
				update((state) => ({ ...state, isLoading: false }));
				throw error;
			}
		},

		// Logout user
		logout: () => {
			if (browser) {
				localStorage.removeItem('auth_token');
				localStorage.removeItem('auth_user');
			}
			set(loggedOutState);
		},

		// Validate current token
		validateToken: async () => {
			try {
				const user = await apiClient.getCurrentUser();
				// Token is valid, update user info and set loading to false.
				update((state) => ({ ...state, user, isLoading: false }));
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

			update((state) => ({ ...state, user }));
		},

		// Set user information (used by GitHub OAuth callbacks)
		setUser: (user: User) => {
			if (browser) {
				localStorage.setItem('auth_user', JSON.stringify(user));
			}

			update((state) => ({ ...state, user }));
		}
	};
}

export const authStore = createAuthStore();
