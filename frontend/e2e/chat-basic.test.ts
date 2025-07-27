// ABOUTME: Basic E2E tests for chat functionality with proper authentication handling
// ABOUTME: Tests fundamental chat interface features without dependencies on external services

import { expect, test } from '@playwright/test';

test.describe('Chat Interface - Basic', () => {
	test('should handle authentication flow', async ({ page }) => {
		// Try to access chat page
		await page.goto('/chat');

		// Should redirect to login if not authenticated
		await page.waitForLoadState('networkidle');

		const currentUrl = page.url();
		console.log('Current URL after navigation:', currentUrl);

		if (currentUrl.includes('/auth/login')) {
			// Login with demo credentials
			await page.fill('input[name="email"]', 'demo@acip.com');
			await page.fill('input[name="password"]', 'demo123456');
			await page.click('button[type="submit"]');

			// Wait for redirect after login
			await page.waitForURL(/(?:\/|\/chat)/, { timeout: 10000 });

			// Navigate to chat if we're on dashboard
			if (!page.url().includes('/chat')) {
				await page.goto('/chat');
			}
		}

		// Now we should be on the chat page
		await expect(page.locator('text=AI Code Assistant')).toBeVisible({ timeout: 10000 });
	});

	test('should show chat interface elements', async ({ context }) => {
		// Create a new page with context
		const page = await context.newPage();

		// Login first
		await page.goto('/auth/login');
		await page.fill('input[name="email"]', 'demo@acip.com');
		await page.fill('input[name="password"]', 'demo123456');
		await page.click('button[type="submit"]');
		await page.waitForURL(/(?:\/|\/dashboard)/, { timeout: 10000 });

		// Navigate to chat
		await page.goto('/chat');

		// Check for main interface elements
		await expect(page.locator('text=AI Code Assistant')).toBeVisible();
		await expect(page.locator('input[placeholder="Ask about the code..."]')).toBeVisible();
		await expect(page.locator('button:has-text("Send")')).toBeVisible();

		// Check for suggested questions when no messages
		await expect(page.locator('text=Try asking:')).toBeVisible();
		await expect(page.locator('text=Start a conversation')).toBeVisible();
	});

	test('should allow typing in input field', async ({ context }) => {
		const page = await context.newPage();

		// Login
		await page.goto('/auth/login');
		await page.fill('input[name="email"]', 'demo@acip.com');
		await page.fill('input[name="password"]', 'demo123456');
		await page.click('button[type="submit"]');
		await page.waitForURL(/(?:\/|\/dashboard)/, { timeout: 10000 });

		// Navigate to chat
		await page.goto('/chat');

		// Type in the input field
		const testMessage = 'This is a test message';
		await page.fill('input[placeholder="Ask about the code..."]', testMessage);

		// Verify the text was entered
		const inputValue = await page.inputValue('input[placeholder="Ask about the code..."]');
		expect(inputValue).toBe(testMessage);
	});

	test('should show repository selector', async ({ context }) => {
		const page = await context.newPage();

		// Login
		await page.goto('/auth/login');
		await page.fill('input[name="email"]', 'demo@acip.com');
		await page.fill('input[name="password"]', 'demo123456');
		await page.click('button[type="submit"]');
		await page.waitForURL(/(?:\/|\/dashboard)/, { timeout: 10000 });

		// Navigate to chat
		await page.goto('/chat');

		// Check for repository selector
		await expect(page.locator('text=Repository:')).toBeVisible();

		// Should show either loading state or actual selector
		const hasSelect = (await page.locator('select#repo-select').count()) > 0;
		const hasLoading = (await page.locator('text=Loading...').count()) > 0;
		const hasNoRepos = (await page.locator('text=No repositories found').count()) > 0;

		expect(hasSelect || hasLoading || hasNoRepos).toBe(true);
	});

	test('should handle session sidebar', async ({ context }) => {
		const page = await context.newPage();

		// Login
		await page.goto('/auth/login');
		await page.fill('input[name="email"]', 'demo@acip.com');
		await page.fill('input[name="password"]', 'demo123456');
		await page.click('button[type="submit"]');
		await page.waitForURL(/(?:\/|\/dashboard)/, { timeout: 10000 });

		// Navigate to chat
		await page.goto('/chat');

		// Find and click the hamburger menu button
		const menuButton = page
			.locator('button')
			.filter({ has: page.locator('svg') })
			.first();
		await menuButton.click();

		// Check if sidebar appears
		await expect(page.locator('text=Chat Sessions')).toBeVisible();
		await expect(page.locator('button:has-text("New Chat")')).toBeVisible();

		// Close sidebar by clicking close button or clicking outside
		const closeButton = page
			.locator('button')
			.filter({ has: page.locator('svg') })
			.nth(1);
		if ((await closeButton.count()) > 0) {
			await closeButton.click();
		}
	});

	test('should handle suggested questions click', async ({ context }) => {
		const page = await context.newPage();

		// Login
		await page.goto('/auth/login');
		await page.fill('input[name="email"]', 'demo@acip.com');
		await page.fill('input[name="password"]', 'demo123456');
		await page.click('button[type="submit"]');
		await page.waitForURL(/(?:\/|\/dashboard)/, { timeout: 10000 });

		// Navigate to chat
		await page.goto('/chat');

		// Wait for suggested questions to appear
		await expect(page.locator('text=Try asking:')).toBeVisible();

		// Click on first suggested question
		const firstQuestion = page
			.locator('button')
			.filter({ hasText: 'Explain the authentication flow' });
		await firstQuestion.click();

		// Check that the question was filled into the input
		const inputValue = await page.inputValue('input[placeholder="Ask about the code..."]');
		expect(inputValue).toBe('Explain the authentication flow');
	});

	test('should show different states correctly', async ({ context }) => {
		const page = await context.newPage();

		// Login
		await page.goto('/auth/login');
		await page.fill('input[name="email"]', 'demo@acip.com');
		await page.fill('input[name="password"]', 'demo123456');
		await page.click('button[type="submit"]');
		await page.waitForURL(/(?:\/|\/dashboard)/, { timeout: 10000 });

		// Navigate to chat
		await page.goto('/chat');

		// Initial state - should show green dot (not streaming)
		await expect(page.locator('.bg-green-500')).toBeVisible();

		// Should show empty conversation state
		await expect(page.locator('text=Start a conversation')).toBeVisible();
		await expect(page.locator('text=Ask me anything about your code!')).toBeVisible();
	});
});
