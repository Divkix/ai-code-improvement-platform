import { expect, test } from '@playwright/test';

test('dashboard page loads correctly', async ({ page }) => {
	await page.goto('/');

	// Check that we're redirected to auth if not logged in, or dashboard loads
	await page.waitForLoadState('networkidle');

	// Check if we're on login page (redirect) or dashboard is loaded
	const currentUrl = page.url();
	if (currentUrl.includes('/auth/login')) {
		await expect(page.locator('form')).toBeVisible();
		console.log('Not authenticated - redirected to login');
	} else {
		// Should be on dashboard, check for dashboard elements
		await expect(page.locator('text=Dashboard')).toBeVisible();
	}
});
