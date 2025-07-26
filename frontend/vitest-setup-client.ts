/// <reference types="@vitest/browser/matchers" />
/// <reference types="@vitest/browser/providers/playwright" />

import { beforeEach, afterEach, vi } from 'vitest';
import { setupGlobalMocks } from '$lib/test-utils';

// Setup global mocks
setupGlobalMocks();

// Setup basic mocks before each test
beforeEach(() => {
	// Mock Chart.js globally
	vi.stubGlobal(
		'Chart',
		vi.fn().mockImplementation(() => ({
			destroy: vi.fn(),
			update: vi.fn(),
			render: vi.fn()
		}))
	);
});

// Clean up after each test
afterEach(() => {
	vi.clearAllMocks();
	vi.restoreAllMocks();
});
