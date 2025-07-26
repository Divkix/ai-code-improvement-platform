// ABOUTME: Test utilities and setup helpers for Vitest and testing-library
// ABOUTME: Provides common mocks, utilities, and setup for consistent testing across components

import { vi } from 'vitest';
import type { DashboardStats, ActivityItem, TrendDataPoint } from '$lib/api';

// Mock data for tests
export const mockDashboardStats: DashboardStats = {
	totalRepositories: 5,
	codeChunksProcessed: 12500,
	costSavingsMonthly: 15000,
	developerHoursReclaimed: 120,
	issuesPreventedMonthly: 25,
	knowledgeRadius: 8500
};

export const mockActivityItems: ActivityItem[] = [
	{
		id: '1',
		type: 'repository_imported',
		message: 'Repository "frontend" has been imported successfully',
		severity: 'success',
		timestamp: new Date(Date.now() - 1000 * 60 * 30).toISOString(), // 30 minutes ago
		repositoryName: 'frontend'
	},
	{
		id: '2',
		type: 'analysis_completed',
		message: 'Code analysis completed for backend',
		severity: 'success',
		timestamp: new Date(Date.now() - 1000 * 60 * 60).toISOString(), // 1 hour ago
		repositoryName: 'backend'
	},
	{
		id: '3',
		type: 'issue_detected',
		message: 'Potential performance issue detected in auth service',
		severity: 'warning',
		timestamp: new Date(Date.now() - 1000 * 60 * 60 * 2).toISOString(), // 2 hours ago
		repositoryName: 'auth-service'
	}
];

export const mockTrendDataPoints: TrendDataPoint[] = [
	{ date: '2024-01-01', codeQuality: 85, performanceScore: 92, issuesResolved: 4 },
	{ date: '2024-01-02', codeQuality: 87, performanceScore: 90, issuesResolved: 5 },
	{ date: '2024-01-03', codeQuality: 89, performanceScore: 94, issuesResolved: 6 },
	{ date: '2024-01-04', codeQuality: 88, performanceScore: 93, issuesResolved: 5 },
	{ date: '2024-01-05', codeQuality: 91, performanceScore: 95, issuesResolved: 7 }
];

// Mock localStorage for tests
export function mockLocalStorage() {
	const store: Record<string, string> = {};

	return {
		getItem: vi.fn((key: string) => store[key] || null),
		setItem: vi.fn((key: string, value: string) => {
			store[key] = value;
		}),
		removeItem: vi.fn((key: string) => {
			delete store[key];
		}),
		clear: vi.fn(() => {
			Object.keys(store).forEach((key) => delete store[key]);
		}),
		key: vi.fn((index: number) => Object.keys(store)[index] || null),
		get length() {
			return Object.keys(store).length;
		}
	};
}

// Mock Chart.js
export function mockChartJs() {
	return {
		Chart: vi.fn().mockImplementation(() => ({
			destroy: vi.fn(),
			update: vi.fn(),
			render: vi.fn()
		})),
		registerables: []
	};
}

// Mock canvas context
export function mockCanvasContext() {
	return {
		getContext: vi.fn(() => ({
			clearRect: vi.fn(),
			fillRect: vi.fn(),
			strokeRect: vi.fn(),
			beginPath: vi.fn(),
			closePath: vi.fn(),
			stroke: vi.fn(),
			fill: vi.fn(),
			moveTo: vi.fn(),
			lineTo: vi.fn(),
			arc: vi.fn(),
			canvas: {
				width: 400,
				height: 300
			}
		}))
	};
}

// Setup global mocks for tests
export function setupGlobalMocks() {
	// Mock localStorage (both browser and Node.js)
	if (typeof global !== 'undefined') {
		Object.defineProperty(global, 'localStorage', {
			value: mockLocalStorage(),
			writable: true
		});
	}

	// Mock Canvas API (only in browser environment or when HTMLCanvasElement exists)
	if (typeof HTMLCanvasElement !== 'undefined') {
		Object.defineProperty(HTMLCanvasElement.prototype, 'getContext', {
			value: vi.fn(() => mockCanvasContext().getContext()),
			writable: true
		});
	} else if (typeof global !== 'undefined') {
		// Create HTMLCanvasElement mock for Node.js environment

		// eslint-disable-next-line @typescript-eslint/ban-ts-comment
		// @ts-ignore - defining HTMLCanvasElement for test environment
		// eslint-disable-next-line @typescript-eslint/no-explicit-any
		(global as any).HTMLCanvasElement = class HTMLCanvasElement {
			getContext() {
				return mockCanvasContext().getContext();
			}
		} as unknown as { new (): HTMLCanvasElement; prototype: HTMLCanvasElement };
	}

	// Mock fetch
	if (typeof global !== 'undefined') {
		global.fetch = vi.fn();
	}

	// Mock window.location (only in environments that have window)
	if (typeof window !== 'undefined') {
		Object.defineProperty(window, 'location', {
			value: {
				href: 'http://localhost:3000',
				origin: 'http://localhost:3000',
				pathname: '/',
				search: '',
				hash: ''
			},
			writable: true
		});
	}
}

// Helper to create mock fetch responses
export function createMockResponse<T>(data: T, status = 200) {
	return Promise.resolve({
		ok: status >= 200 && status < 300,
		status,
		json: () => Promise.resolve(data),
		text: () => Promise.resolve(JSON.stringify(data))
	} as Response);
}

// Error response helper
export function createMockErrorResponse(message: string, status = 500) {
	return Promise.resolve({
		ok: false,
		status,
		json: () => Promise.resolve({ message, error: message }),
		text: () => Promise.resolve(JSON.stringify({ message, error: message }))
	} as Response);
}
