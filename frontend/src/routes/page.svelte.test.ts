import { page } from '@vitest/browser/context';
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { render, waitFor } from 'vitest-browser-svelte';
import Page from './+page.svelte';
import { mockDashboardStats, mockActivityItems, mockTrendDataPoints } from '$lib/test-utils';

// Mock Chart.js
vi.mock('chart.js', () => ({
	Chart: vi.fn().mockImplementation(() => ({
		destroy: vi.fn(),
		update: vi.fn(),
		render: vi.fn()
	})),
	registerables: []
}));

// Mock API hooks with successful responses by default
vi.mock('$lib/api/hooks', () => ({
	getDashboardStats: vi.fn().mockResolvedValue(mockDashboardStats),
	getDashboardActivity: vi.fn().mockResolvedValue(mockActivityItems),
	getDashboardTrends: vi.fn().mockResolvedValue(mockTrendDataPoints)
}));

describe('/+page.svelte', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	it('should render dashboard stats after loading', async () => {
		render(Page);

		// Wait for loading to complete and content to appear
		await waitFor(() => {
			const statsSection = page.getByTestId('dashboard-stats');
			expect.element(statsSection).toBeInTheDocument();
		});

		// Check that dashboard stats are displayed
		const statsSection = page.getByTestId('dashboard-stats');
		await expect.element(statsSection).toBeInTheDocument();

		// Verify repository count is displayed
		const repoCount = page.getByText('5');
		await expect.element(repoCount).toBeInTheDocument();
	});

	it('should show loading state initially', async () => {
		render(Page);

		// Should show skeleton loading initially - check for any skeleton element
		const loadingIndicator = page.getByTestId('loading-skeleton');
		await expect.element(loadingIndicator).toBeInTheDocument();
	});

	it('should display formatted metrics correctly', async () => {
		render(Page);

		// Wait for content to load
		await waitFor(() => {
			const statsSection = page.getByTestId('dashboard-stats');
			expect.element(statsSection).toBeInTheDocument();
		});

		// Check for formatted code chunks count (12.5K)
		const codeChunks = page.getByText('12.5K');
		await expect.element(codeChunks).toBeInTheDocument();

		// Check for cost savings
		const costSavings = page.getByText('15K');
		await expect.element(costSavings).toBeInTheDocument();
	});
});
