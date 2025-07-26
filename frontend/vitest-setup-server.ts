// ABOUTME: Vitest setup for server-side tests (Node.js environment)
// ABOUTME: Configures test environment for testing stores, utilities, and server-side logic

import { beforeEach, afterEach, vi } from 'vitest';
import { setupGlobalMocks } from '$lib/test-utils';

// Setup global mocks for server environment
setupGlobalMocks();

// Setup environment for server tests
beforeEach(() => {
	// Additional server-specific setup if needed
});

// Clean up after each test
afterEach(() => {
	vi.clearAllMocks();
	vi.restoreAllMocks();
});
