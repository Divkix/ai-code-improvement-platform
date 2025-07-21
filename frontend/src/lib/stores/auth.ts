// ABOUTME: Authentication store using Svelte stores for user state management
// ABOUTME: Handles token storage, user data, and authentication state throughout the app

import { writable } from 'svelte/store';
import { browser } from '$app/environment';
import apiClient from '../api';

export interface User {
	id: string;
	email: string;
	name: string;
	createdAt: string;
}

export interface AuthState {
	user: User | null;
	token: string | null;
	isAuthenticated: boolean;
	isLoading: boolean;
}

const initialState: AuthState = {
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
			if (!browser) return;
			
			update(state => ({ ...state, isLoading: true }));
			
			const token = localStorage.getItem('auth_token');
			const userStr = localStorage.getItem('auth_user');
			
			if (token && userStr) {
				try {
					const user = JSON.parse(userStr);
					update(state => ({
						...state,
						user,
						token,
						isAuthenticated: true,
						isLoading: false
					}));
					
					// Validate token by fetching current user
					authStore.validateToken();
				} catch (error) {
					console.error('Failed to parse stored user data:', error);
					authStore.logout();
				}
			} else {
				update(state => ({ ...state, isLoading: false }));
			}
		},
		
		// Login user
		login: async (email: string, password: string) => {
			update(state => ({ ...state, isLoading: true }));
			
			try {
				const response = await apiClient.login(email, password);
				
				// Store token and user data
				if (browser) {
					localStorage.setItem('auth_token', response.token);
					localStorage.setItem('auth_user', JSON.stringify(response.user));
				}
				
				update(state => ({
					...state,
					user: response.user,
					token: response.token,
					isAuthenticated: true,
					isLoading: false
				}));
				
				return response;
			} catch (error) {
				update(state => ({ ...state, isLoading: false }));
				throw error;
			}
		},
		
		
		// Logout user
		logout: () => {
			if (browser) {
				localStorage.removeItem('auth_token');
				localStorage.removeItem('auth_user');
			}
			
			set(initialState);
		},
		
		// Validate current token
		validateToken: async () => {
			try {
				const user = await apiClient.getCurrentUser();
				update(state => ({ ...state, user, isLoading: false }));
			} catch (error) {
				console.error('Token validation failed:', error);
				authStore.logout();
			}
		},
		
		// Update user information
		updateUser: (user: User) => {
			if (browser) {
				localStorage.setItem('auth_user', JSON.stringify(user));
			}
			
			update(state => ({ ...state, user }));
		}
	};
}

export const authStore = createAuthStore();