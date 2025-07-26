// ABOUTME: Authentication utilities for route protection and auth validation
// ABOUTME: Provides helper functions for checking auth status and redirecting users

import { browser } from '$app/environment';
import { goto } from '$app/navigation';
import { authStore } from './stores/auth';

// Check if user is authenticated
export function isAuthenticated(): boolean {
	return authStore.current.isAuthenticated;
}

// Redirect to login if not authenticated
export function requireAuth(): void {
	if (!browser) return;

	if (!authStore.current.isAuthenticated && !authStore.current.isLoading) {
		goto('/auth/login');
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
