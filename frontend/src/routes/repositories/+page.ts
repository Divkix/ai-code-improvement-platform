// ABOUTME: Page load function for repositories page with authentication protection
// ABOUTME: Redirects to login if user is not authenticated

import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url, fetch }) => {
	// Check if user is authenticated by checking for auth token
	const authToken = typeof window !== 'undefined' ? localStorage.getItem('auth_token') : null;
	
	if (!authToken) {
		throw redirect(302, '/auth/login');
	}
	
	// If authenticated, return empty object (we can load data here later)
	return {};
};