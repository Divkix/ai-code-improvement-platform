// ABOUTME: Authentication utilities for route protection and auth validation
// ABOUTME: Provides helper functions for checking auth status and redirecting users

import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import { authStore } from './stores/auth';

/**
 * Validates JWT token format and expiration
 * @param token - JWT token to validate
 * @returns true if token is valid and not expired, false otherwise
 */
function isValidJWT(token: string): boolean {
	try {
		// Check JWT format: should have 3 parts separated by dots
		const parts = token.split('.');
		if (parts.length !== 3) {
			return false;
		}

		// Decode the payload (second part)
		const payload = JSON.parse(atob(parts[1]));

		// Check if token has expiration claim
		if (!payload.exp) {
			return false;
		}

		// Check if token is expired (exp is in seconds, Date.now() is in milliseconds)
		const currentTime = Math.floor(Date.now() / 1000);
		if (payload.exp < currentTime) {
			return false;
		}

		return true;
	} catch {
		return false;
	}
}

/**
 * Check authentication status by validating token expiration
 * Does not make API calls, only validates token locally
 * @returns true if user has valid token, false otherwise
 */
export async function checkAuthStatus(): Promise<boolean> {
	if (!browser) return false;

	try {
		const token = localStorage.getItem('auth_token');

		if (!token) {
			return false;
		}

		// Validate token format and expiration
		if (!isValidJWT(token)) {
			console.warn('Auth token is invalid or expired, logging out');
			// Remove invalid token and user data
			localStorage.removeItem('auth_token');
			localStorage.removeItem('auth_user');
			authStore.logout();
			return false;
		}

		return true;
	} catch (error) {
		console.error('Error checking auth status:', error);
		return false;
	}
}

// Check if user is authenticated
export function isAuthenticated(): boolean {
	return authStore.current.isAuthenticated;
}

/**
 * Redirect to login if not authenticated
 * Makes this async to properly validate token before allowing access
 * Throws error if not authenticated to stop execution flow
 */
export async function requireAuth(): Promise<void> {
	if (!browser) return;

	try {
		// Check if token exists and is valid
		const isValid = await checkAuthStatus();

		if (!isValid) {
			// Token is invalid or expired, redirect to login
			console.warn('Authentication required, redirecting to login');
			await goto('/auth/login');
			throw new Error('Authentication required');
		}

		// Also check store state (for consistency)
		if (!authStore.current.isAuthenticated && !authStore.current.isLoading) {
			console.warn('Auth store shows not authenticated, redirecting to login');
			await goto('/auth/login');
			throw new Error('Authentication required');
		}

		// TODO: Implement token refresh mechanism when token is close to expiration
		// TODO: Use BroadcastChannel to sync auth state across tabs
	} catch (error) {
		if (error instanceof Error && error.message === 'Authentication required') {
			throw error;
		}
		console.error('Error in requireAuth:', error);
		await goto('/auth/login');
		throw new Error('Authentication required');
	}
}

// Redirect to dashboard if already authenticated
export function requireGuest(): void {
	if (!browser) return;

	if (authStore.current.isAuthenticated) {
		goto('/');
	}
}

// Get current user
export function getCurrentUser() {
	return authStore.current.user;
}

// Get auth token
export function getAuthToken(): string | null {
	if (!browser) return null;
	return localStorage.getItem('auth_token');
}
