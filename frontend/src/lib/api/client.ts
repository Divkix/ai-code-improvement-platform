// ABOUTME: Generated API client using openapi-fetch and generated types
// ABOUTME: Provides type-safe API calls with automatic request/response validation
import createClient from 'openapi-fetch';
import type { paths } from './types';

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

// Export the client as default
export default apiClient;
