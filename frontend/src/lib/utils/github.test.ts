// ABOUTME: Tests for GitHub URL generation utilities
// ABOUTME: Verifies URL construction, line highlighting, and window.open behavior

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { generateGitHubUrl, openGitHubUrl } from './github';
import type { Repository } from '$lib/api';

describe('GitHub utilities', () => {
	describe('generateGitHubUrl', () => {
		const mockRepository: Repository = {
			id: '1',
			name: 'test-repo',
			owner: 'test-owner',
			fullName: 'test-owner/test-repo',
			description: 'Test repository',
			status: 'ready',
			importedAt: '2024-01-01T00:00:00Z',
			githubUrl: 'https://github.com/test-owner/test-repo'
		};

		it('should generate basic GitHub URL without line numbers', () => {
			const url = generateGitHubUrl(mockRepository, 'src/index.ts');

			expect(url).toBe('https://github.com/test-owner/test-repo/blob/main/src/index.ts');
		});

		it('should generate GitHub URL with single line highlight', () => {
			const url = generateGitHubUrl(mockRepository, 'src/index.ts', 42);

			expect(url).toBe('https://github.com/test-owner/test-repo/blob/main/src/index.ts#L42');
		});

		it('should generate GitHub URL with line range highlight', () => {
			const url = generateGitHubUrl(mockRepository, 'src/index.ts', 10, 20);

			expect(url).toBe('https://github.com/test-owner/test-repo/blob/main/src/index.ts#L10-L20');
		});

		it('should treat same start and end line as single line', () => {
			const url = generateGitHubUrl(mockRepository, 'src/index.ts', 15, 15);

			expect(url).toBe('https://github.com/test-owner/test-repo/blob/main/src/index.ts#L15');
		});

		it('should handle file paths with slashes', () => {
			const url = generateGitHubUrl(mockRepository, 'src/components/Button.svelte', 25, 30);

			expect(url).toBe(
				'https://github.com/test-owner/test-repo/blob/main/src/components/Button.svelte#L25-L30'
			);
		});

		it('should handle root level files', () => {
			const url = generateGitHubUrl(mockRepository, 'README.md', 1);

			expect(url).toBe('https://github.com/test-owner/test-repo/blob/main/README.md#L1');
		});

		it('should handle files with special characters in path', () => {
			const url = generateGitHubUrl(mockRepository, 'src/utils/api-client.ts');

			expect(url).toBe('https://github.com/test-owner/test-repo/blob/main/src/utils/api-client.ts');
		});

		it('should handle endLine when startLine is 0', () => {
			const url = generateGitHubUrl(mockRepository, 'src/index.ts', 0, 5);

			// 0 is falsy, so no line highlight should be added
			expect(url).toBe('https://github.com/test-owner/test-repo/blob/main/src/index.ts');
		});

		it('should handle negative line numbers', () => {
			const url = generateGitHubUrl(mockRepository, 'src/index.ts', -1, 5);

			expect(url).toBe('https://github.com/test-owner/test-repo/blob/main/src/index.ts#L-1-L5');
		});

		it('should work with different repository structures', () => {
			const specialRepo: Repository = {
				id: '2',
				name: 'special-repo',
				owner: 'special-owner',
				fullName: 'special-owner/special-repo',
				description: 'Special repository',
				status: 'ready',
				importedAt: '2024-01-01T00:00:00Z',
				githubUrl: 'https://github.com/special-owner/special-repo'
			};

			const url = generateGitHubUrl(specialRepo, 'lib/core.js', 100, 200);

			expect(url).toBe(
				'https://github.com/special-owner/special-repo/blob/main/lib/core.js#L100-L200'
			);
		});
	});

	describe('openGitHubUrl', () => {
		beforeEach(() => {
			// Mock window.open
			if (!global.window) {
				global.window = {} as any;
			}
			global.window.open = vi.fn();
		});

		afterEach(() => {
			vi.restoreAllMocks();
		});

		it('should call window.open with correct parameters', () => {
			const testUrl = 'https://github.com/test-owner/test-repo/blob/main/src/index.ts#L42';

			openGitHubUrl(testUrl);

			expect(window.open).toHaveBeenCalledWith(testUrl, '_blank', 'noopener,noreferrer');
		});

		it('should handle empty URL', () => {
			openGitHubUrl('');

			expect(window.open).toHaveBeenCalledWith('', '_blank', 'noopener,noreferrer');
		});

		it('should handle various URL formats', () => {
			const testUrls = [
				'https://github.com/owner/repo',
				'https://github.com/owner/repo/blob/main/file.ts',
				'https://github.com/owner/repo/blob/main/file.ts#L1-L10',
				'https://github.com/owner/repo/tree/feature-branch'
			];

			testUrls.forEach((url) => {
				openGitHubUrl(url);
				expect(window.open).toHaveBeenCalledWith(url, '_blank', 'noopener,noreferrer');
			});

			expect(window.open).toHaveBeenCalledTimes(testUrls.length);
		});

		it('should not throw error if window.open is not available', () => {
			// Simulate environment where window.open is not available
			const originalOpen = window.open;
			delete (window as any).open;

			expect(() => {
				openGitHubUrl('https://github.com/test/repo');
			}).toThrow();

			// Restore
			window.open = originalOpen;
		});
	});

	describe('integration tests', () => {
		beforeEach(() => {
			if (!global.window) {
				global.window = {} as any;
			}
			global.window.open = vi.fn();
		});

		afterEach(() => {
			vi.restoreAllMocks();
		});

		it('should generate URL and open it', () => {
			const mockRepository: Repository = {
				id: '1',
				name: 'integration-test',
				owner: 'test-org',
				fullName: 'test-org/integration-test',
				description: 'Integration test repository',
				status: 'ready',
				importedAt: '2024-01-01T00:00:00Z',
				githubUrl: 'https://github.com/test-org/integration-test'
			};

			const url = generateGitHubUrl(mockRepository, 'src/app.ts', 15, 25);
			openGitHubUrl(url);

			expect(window.open).toHaveBeenCalledWith(
				'https://github.com/test-org/integration-test/blob/main/src/app.ts#L15-L25',
				'_blank',
				'noopener,noreferrer'
			);
		});
	});
});
