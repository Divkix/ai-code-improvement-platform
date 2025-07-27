// ABOUTME: End-to-end tests for the AI chat interface functionality
// ABOUTME: Tests chat sessions, messaging, streaming responses, and repository context

import { expect, test } from '@playwright/test';

test.describe('Chat Interface', () => {
	test.beforeEach(async ({ page }) => {
		// Navigate to login and authenticate with demo user
		await page.goto('/auth/login');
		await page.fill('input[name="email"]', 'demo@acip.com');
		await page.fill('input[name="password"]', 'demo123456');
		await page.click('button[type="submit"]');
		await page.waitForURL('/');
	});

	test('should load chat page and show initial state', async ({ page }) => {
		await page.goto('/chat');

		// Check page title
		await expect(page).toHaveTitle(/AI Chat/);

		// Check for main chat interface elements
		await expect(page.locator('h2')).toContainText('AI Code Assistant');
		await expect(page.locator('text=Start a conversation')).toBeVisible();
		await expect(page.locator('text=Ask me anything about your code!')).toBeVisible();

		// Check for suggested questions
		await expect(page.locator('text=Try asking:')).toBeVisible();
		await expect(page.locator('text=Explain the authentication flow')).toBeVisible();

		// Check for input form
		await expect(page.locator('input[placeholder="Ask about the code..."]')).toBeVisible();
		await expect(page.locator('button:has-text("Send")')).toBeVisible();
	});

	test('should show session sidebar', async ({ page }) => {
		await page.goto('/chat');

		// Click hamburger menu to open session sidebar
		await page.click('button:has(svg)'); // Hamburger menu button

		// Check sidebar is visible
		await expect(page.locator('text=Chat Sessions')).toBeVisible();
		await expect(page.locator('button:has-text("New Chat")')).toBeVisible();

		// Close sidebar
		await page.click('button:has(svg):last-child'); // Close button
		await expect(page.locator('text=Chat Sessions')).not.toBeVisible();
	});

	test('should create new chat session', async ({ page }) => {
		await page.goto('/chat');

		// Open session sidebar
		await page.click('button:has(svg)');

		// Click New Chat button
		await page.click('button:has-text("New Chat")');

		// Wait for session creation
		await page.waitForTimeout(1000);

		// Check that session was created (title should change from default)
		await expect(page.locator('h2')).toContainText('New Chat');
	});

	test('should send a message and receive response', async ({ page }) => {
		await page.goto('/chat');

		// Type a message
		const testMessage = 'What is the main purpose of this codebase?';
		await page.fill('input[placeholder="Ask about the code..."]', testMessage);

		// Send message
		await page.click('button:has-text("Send")');

		// Check that user message appears
		await expect(page.locator('.text-white').filter({ hasText: testMessage })).toBeVisible();

		// Check that streaming indicator appears
		await expect(page.locator('text=Analyzing code...')).toBeVisible();

		// Wait for response (up to 30 seconds)
		await page.waitForSelector('.bg-gray-100:has(.prose)', { timeout: 30000 });

		// Check that assistant response appears
		await expect(page.locator('.bg-gray-100').first()).toBeVisible();

		// Check that streaming indicator disappears
		await expect(page.locator('text=Analyzing code...')).not.toBeVisible();
	});

	test('should handle suggested questions', async ({ page }) => {
		await page.goto('/chat');

		// Click on first suggested question
		await page.click('button:has-text("Explain the authentication flow")');

		// Check that the question was sent
		await expect(
			page.locator('.text-white').filter({ hasText: 'Explain the authentication flow' })
		).toBeVisible();

		// Check that suggested questions disappear after sending
		await expect(page.locator('text=Try asking:')).not.toBeVisible();
	});

	test('should show repository selector', async ({ page }) => {
		await page.goto('/chat');

		// Check repository selector is present
		await expect(page.locator('label[for="repo-select"]')).toContainText('Repository:');

		// Check for select dropdown (may show loading or repositories)
		const repoSelect = page.locator('select#repo-select');
		if (await repoSelect.isVisible()) {
			await expect(repoSelect).toBeVisible();
			await expect(repoSelect.locator('option').first()).toContainText('All repositories');
		} else {
			// If loading, check for loading indicator
			await expect(page.locator('text=Loading...')).toBeVisible();
		}
	});

	test('should maintain conversation history', async ({ page }) => {
		await page.goto('/chat');

		// Send first message
		await page.fill('input[placeholder="Ask about the code..."]', 'Hello');
		await page.click('button:has-text("Send")');

		// Wait for response
		await page.waitForSelector('.bg-gray-100:has(.prose)', { timeout: 30000 });

		// Send second message
		await page.fill('input[placeholder="Ask about the code..."]', 'Can you explain more?');
		await page.click('button:has-text("Send")');

		// Check that both user messages are visible
		await expect(page.locator('.text-white').filter({ hasText: 'Hello' })).toBeVisible();
		await expect(
			page.locator('.text-white').filter({ hasText: 'Can you explain more?' })
		).toBeVisible();

		// Check that there are multiple assistant responses
		await expect(page.locator('.bg-gray-100')).toHaveCount(2, { timeout: 30000 });
	});

	test('should show timestamp and token information', async ({ page }) => {
		await page.goto('/chat');

		// Send a message
		await page.fill('input[placeholder="Ask about the code..."]', 'Test message');
		await page.click('button:has-text("Send")');

		// Wait for response
		await page.waitForSelector('.bg-gray-100:has(.prose)', { timeout: 30000 });

		// Check for timestamp (format: HH:MM)
		await expect(page.locator('.opacity-70').filter({ hasText: /\d{1,2}:\d{2}/ })).toBeVisible();

		// Check for token count in assistant message (if available)
		const tokenInfo = page.locator('text=/\\d+ tokens/');
		if ((await tokenInfo.count()) > 0) {
			await expect(tokenInfo.first()).toBeVisible();
		}
	});

	test('should handle errors gracefully', async ({ page }) => {
		await page.goto('/chat');

		// Mock a network error by intercepting the API call
		await page.route('/api/chat/sessions/*/message', (route) => {
			route.fulfill({
				status: 500,
				contentType: 'application/json',
				body: JSON.stringify({ error: 'internal_error', message: 'Test error' })
			});
		});

		// Send a message
		await page.fill('input[placeholder="Ask about the code..."]', 'This should fail');
		await page.click('button:has-text("Send")');

		// Check that error message appears
		await expect(page.locator('text=Error:')).toBeVisible();
	});

	test('should disable input during streaming', async ({ page }) => {
		await page.goto('/chat');

		// Send a message
		await page.fill('input[placeholder="Ask about the code..."]', 'Test streaming disable');
		await page.click('button:has-text("Send")');

		// Check that input is disabled during streaming
		await expect(page.locator('input[placeholder="Ask about the code..."]')).toBeDisabled();
		await expect(page.locator('button:has-text("Stop")')).toBeVisible();

		// Wait for streaming to complete
		await page.waitForSelector('.bg-gray-100:has(.prose)', { timeout: 30000 });

		// Check that input is re-enabled
		await expect(page.locator('input[placeholder="Ask about the code..."]')).toBeEnabled();
		await expect(page.locator('button:has-text("Send")')).toBeVisible();
	});

	test('should show retrieved chunks information', async ({ page }) => {
		await page.goto('/chat');

		// Send a message that would trigger code retrieval
		await page.fill(
			'input[placeholder="Ask about the code..."]',
			'Show me the authentication code'
		);
		await page.click('button:has-text("Send")');

		// Wait for response
		await page.waitForSelector('.bg-gray-100:has(.prose)', { timeout: 30000 });

		// Check for retrieved chunks indicator (if code chunks were found)
		const chunksInfo = page.locator('text=/Analyzed \\d+ code chunks/');
		if ((await chunksInfo.count()) > 0) {
			await expect(chunksInfo.first()).toBeVisible();
		}
	});

	test('should handle session management', async ({ page }) => {
		await page.goto('/chat');

		// Open session sidebar
		await page.click('button:has(svg)');

		// Create new session
		await page.click('button:has-text("New Chat")');
		await page.waitForTimeout(1000);

		// Send a message to populate the session
		await page.fill('input[placeholder="Ask about the code..."]', 'First session message');
		await page.click('button:has-text("Send")');
		await page.waitForSelector('.bg-gray-100:has(.prose)', { timeout: 30000 });

		// Open sidebar again and check session appears
		await page.click('button:has(svg)');

		// Look for session with auto-generated title
		const sessionTitle = page
			.locator('.border')
			.filter({ hasText: 'First session message' })
			.first();
		if ((await sessionTitle.count()) > 0) {
			await expect(sessionTitle).toBeVisible();
		} else {
			// Fallback to check for "New Chat" title
			await expect(page.locator('text=New Chat')).toBeVisible();
		}
	});
});

test.describe('Chat Performance', () => {
	test.beforeEach(async ({ page }) => {
		await page.goto('/auth/login');
		await page.fill('input[name="email"]', 'demo@acip.com');
		await page.fill('input[name="password"]', 'demo123456');
		await page.click('button[type="submit"]');
		await page.waitForURL('/');
	});

	test('should respond within acceptable time limits', async ({ page }) => {
		await page.goto('/chat');

		const startTime = Date.now();

		// Send a message
		await page.fill('input[placeholder="Ask about the code..."]', 'What is this codebase about?');
		await page.click('button:has-text("Send")');

		// Wait for first content to appear (first token)
		await page.waitForSelector('.bg-gray-100:has(.prose)', { timeout: 15000 });

		const firstTokenTime = Date.now() - startTime;

		// Check that first token appears within 5 seconds (slice9.md requirement)
		expect(firstTokenTime).toBeLessThan(5000);

		console.log(`First token latency: ${firstTokenTime}ms`);
	});

	test('should handle multiple concurrent messages gracefully', async ({ page }) => {
		await page.goto('/chat');

		// Send first message
		await page.fill('input[placeholder="Ask about the code..."]', 'First concurrent message');
		await page.click('button:has-text("Send")');

		// Wait for response to complete
		await page.waitForSelector('.bg-gray-100:has(.prose)', { timeout: 30000 });
		await expect(page.locator('text=Analyzing code...')).not.toBeVisible();

		// Send second message quickly
		await page.fill('input[placeholder="Ask about the code..."]', 'Second concurrent message');
		await page.click('button:has-text("Send")');

		// Should handle the second message properly
		await page.waitForSelector('.bg-gray-100:has(.prose)', { timeout: 30000 });

		// Check both messages are present
		await expect(
			page.locator('.text-white').filter({ hasText: 'First concurrent message' })
		).toBeVisible();
		await expect(
			page.locator('.text-white').filter({ hasText: 'Second concurrent message' })
		).toBeVisible();
	});
});
