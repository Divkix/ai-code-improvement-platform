// ABOUTME: MSW request handlers for mocking API endpoints in tests
// ABOUTME: Provides consistent mock responses for dashboard, auth, and repository endpoints

import { http, HttpResponse } from 'msw';
import { mockDashboardStats, mockActivityItems, mockTrendDataPoints } from '$lib/test-utils';

const API_BASE = 'http://localhost:8080/api';

export const handlers = [
	// Dashboard endpoints
	http.get(`${API_BASE}/dashboard/stats`, () => {
		return HttpResponse.json(mockDashboardStats);
	}),

	http.get(`${API_BASE}/dashboard/activity`, ({ request }) => {
		const url = new URL(request.url);
		const limit = url.searchParams.get('limit');
		const limitNum = limit ? parseInt(limit, 10) : undefined;

		const items = limitNum ? mockActivityItems.slice(0, limitNum) : mockActivityItems;
		return HttpResponse.json(items);
	}),

	http.get(`${API_BASE}/dashboard/trends`, ({ request }) => {
		const url = new URL(request.url);
		const days = url.searchParams.get('days');
		const daysNum = days ? parseInt(days, 10) : undefined;

		const trends = daysNum ? mockTrendDataPoints.slice(-daysNum) : mockTrendDataPoints;
		return HttpResponse.json(trends);
	}),

	// Health endpoints
	http.get('/health', () => {
		return HttpResponse.json({ status: 'ok', timestamp: new Date().toISOString() });
	}),

	http.get(`${API_BASE}/health`, () => {
		return HttpResponse.json({ status: 'ok', timestamp: new Date().toISOString() });
	}),

	// Repository endpoints
	http.get(`${API_BASE}/repositories`, () => {
		return HttpResponse.json({
			repositories: [
				{
					id: '1',
					name: 'frontend',
					fullName: 'test-org/frontend',
					description: 'Frontend application',
					status: 'ready',
					importedAt: new Date().toISOString(),
					githubUrl: 'https://github.com/test-org/frontend'
				},
				{
					id: '2',
					name: 'backend',
					fullName: 'test-org/backend',
					description: 'Backend API',
					status: 'ready',
					importedAt: new Date().toISOString(),
					githubUrl: 'https://github.com/test-org/backend'
				}
			],
			total: 2,
			offset: 0,
			limit: 20
		});
	}),

	// GitHub endpoints
	http.get(`${API_BASE}/github/repositories`, () => {
		return HttpResponse.json({
			repositories: [
				{
					id: 123,
					name: 'test-repo',
					fullName: 'test-org/test-repo',
					description: 'A test repository',
					private: false,
					htmlUrl: 'https://github.com/test-org/test-repo'
				}
			],
			totalCount: 1,
			hasNextPage: false
		});
	}),

	// Auth endpoints
	http.post(`${API_BASE}/auth/login`, async ({ request }) => {
		const body = (await request.json()) as { email: string; password: string };

		if (body.email === 'demo@github-analyzer.com' && body.password === 'demo123456') {
			return HttpResponse.json({
				token: 'mock-jwt-token',
				user: {
					id: '1',
					email: 'demo@github-analyzer.com',
					name: 'Demo User'
				}
			});
		}

		return HttpResponse.json({ message: 'Invalid credentials' }, { status: 401 });
	}),

	// Error simulation for testing error states
	http.get(`${API_BASE}/dashboard/stats-error`, () => {
		return HttpResponse.json({ message: 'Internal server error' }, { status: 500 });
	})
];
