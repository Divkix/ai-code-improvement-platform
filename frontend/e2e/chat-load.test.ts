// ABOUTME: Load testing for chat interface to verify concurrent user handling
// ABOUTME: Tests system performance under multiple simultaneous chat sessions

import { expect, test } from '@playwright/test';

test.describe('Chat Load Testing', () => {
	const CONCURRENT_USERS = 5;
	const MESSAGES_PER_USER = 3;

	test('should handle multiple users chatting simultaneously', async ({ browser }) => {
		const contexts = [];
		const pages = [];

		// Create multiple browser contexts (simulate different users)
		for (let i = 0; i < CONCURRENT_USERS; i++) {
			const context = await browser.newContext();
			const page = await context.newPage();
			contexts.push(context);
			pages.push(page);
		}

		try {
			// Login all users
			await Promise.all(
				pages.map(async (page, index) => {
					await page.goto('/auth/login');
					await page.fill('input[name="email"]', 'demo@github-analyzer.com');
					await page.fill('input[name="password"]', 'demo123456');
					await page.click('button[type="submit"]');
					await page.waitForURL('/');
					await page.goto('/chat');
				})
			);

			// Have all users send messages concurrently
			const messagePromises = [];

			for (let userIndex = 0; userIndex < CONCURRENT_USERS; userIndex++) {
				const page = pages[userIndex];

				for (let msgIndex = 0; msgIndex < MESSAGES_PER_USER; msgIndex++) {
					messagePromises.push(
						(async () => {
							const message = `User ${userIndex + 1} Message ${msgIndex + 1}: What can you tell me about this codebase?`;

							await page.fill('input[placeholder="Ask about the code..."]', message);
							await page.click('button:has-text("Send")');

							// Wait for user message to appear
							await expect(page.locator('.text-white').filter({ hasText: message })).toBeVisible();

							// Wait for assistant response with timeout
							await page.waitForSelector('.bg-gray-100:has(.prose)', { timeout: 60000 });

							return { user: userIndex + 1, message: msgIndex + 1, success: true };
						})()
					);

					// Stagger messages slightly to simulate realistic usage
					await new Promise((resolve) => setTimeout(resolve, 500));
				}
			}

			// Wait for all messages to complete
			const startTime = Date.now();
			const results = await Promise.allSettled(messagePromises);
			const endTime = Date.now();

			console.log(`Load test completed in ${endTime - startTime}ms`);
			console.log(
				`Successful requests: ${results.filter((r) => r.status === 'fulfilled').length}/${results.length}`
			);

			// Check that most requests succeeded (allow for some failures under load)
			const successCount = results.filter((r) => r.status === 'fulfilled').length;
			const successRate = successCount / results.length;

			expect(successRate).toBeGreaterThan(0.8); // 80% success rate minimum

			// Verify that each user's page still shows their messages
			await Promise.all(
				pages.map(async (page, userIndex) => {
					const userMessages = page
						.locator('.text-white')
						.filter({ hasText: `User ${userIndex + 1}` });
					const messageCount = await userMessages.count();
					expect(messageCount).toBeGreaterThan(0);
				})
			);
		} finally {
			// Clean up contexts
			await Promise.all(contexts.map((context) => context.close()));
		}
	});

	test('should maintain performance under sustained load', async ({ browser }) => {
		const context = await browser.newContext();
		const page = await context.newPage();

		try {
			// Login
			await page.goto('/auth/login');
			await page.fill('input[name="email"]', 'demo@github-analyzer.com');
			await page.fill('input[name="password"]', 'demo123456');
			await page.click('button[type="submit"]');
			await page.waitForURL('/');
			await page.goto('/chat');

			const responseTimes = [];
			const MESSAGE_COUNT = 10;

			// Send multiple messages in sequence and measure response times
			for (let i = 0; i < MESSAGE_COUNT; i++) {
				const startTime = Date.now();

				const message = `Performance test message ${i + 1}: Explain the codebase structure`;
				await page.fill('input[placeholder="Ask about the code..."]', message);
				await page.click('button:has-text("Send")');

				// Wait for first token (streaming start)
				await page.waitForSelector('.bg-gray-100:has(.prose)', { timeout: 30000 });

				const responseTime = Date.now() - startTime;
				responseTimes.push(responseTime);

				console.log(`Message ${i + 1} response time: ${responseTime}ms`);

				// Wait for streaming to complete before next message
				await expect(page.locator('text=Analyzing code...')).not.toBeVisible();
			}

			// Calculate performance metrics
			const avgResponseTime = responseTimes.reduce((a, b) => a + b) / responseTimes.length;
			const maxResponseTime = Math.max(...responseTimes);
			const minResponseTime = Math.min(...responseTimes);

			console.log(`Average response time: ${avgResponseTime}ms`);
			console.log(`Max response time: ${maxResponseTime}ms`);
			console.log(`Min response time: ${minResponseTime}ms`);

			// Performance assertions
			expect(avgResponseTime).toBeLessThan(10000); // Average under 10 seconds
			expect(maxResponseTime).toBeLessThan(30000); // Max under 30 seconds

			// Check that performance doesn't degrade significantly over time
			const firstHalf = responseTimes.slice(0, Math.floor(MESSAGE_COUNT / 2));
			const secondHalf = responseTimes.slice(Math.floor(MESSAGE_COUNT / 2));

			const firstHalfAvg = firstHalf.reduce((a, b) => a + b) / firstHalf.length;
			const secondHalfAvg = secondHalf.reduce((a, b) => a + b) / secondHalf.length;

			// Second half shouldn't be more than 50% slower than first half
			expect(secondHalfAvg).toBeLessThan(firstHalfAvg * 1.5);
		} finally {
			await context.close();
		}
	});

	test('should handle rapid message bursts', async ({ browser }) => {
		const context = await browser.newContext();
		const page = await context.newPage();

		try {
			// Login
			await page.goto('/auth/login');
			await page.fill('input[name="email"]', 'demo@github-analyzer.com');
			await page.fill('input[name="password"]', 'demo123456');
			await page.click('button[type="submit"]');
			await page.waitForURL('/');
			await page.goto('/chat');

			// Send messages in rapid succession
			const messages = [
				'What is this project about?',
				'Show me the main components',
				'Explain the database schema',
				'How does authentication work?',
				'What are the API endpoints?'
			];

			// Send all messages quickly
			for (const message of messages) {
				await page.fill('input[placeholder="Ask about the code..."]', message);
				await page.click('button:has-text("Send")');

				// Brief pause to allow UI to update
				await page.waitForTimeout(100);
			}

			// Wait for all user messages to appear
			for (const message of messages) {
				await expect(page.locator('.text-white').filter({ hasText: message })).toBeVisible();
			}

			// Wait for at least one assistant response
			await page.waitForSelector('.bg-gray-100:has(.prose)', { timeout: 60000 });

			// The system should handle this gracefully without crashes
			// Check that the page is still responsive
			await expect(page.locator('input[placeholder="Ask about the code..."]')).toBeVisible();
			await expect(page.locator('button:has-text("Send")')).toBeVisible();
		} finally {
			await context.close();
		}
	});
});
