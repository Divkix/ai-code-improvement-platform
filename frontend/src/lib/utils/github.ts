// ABOUTME: GitHub URL generation utilities for linking to repository files with line highlights
// ABOUTME: Provides functions to construct GitHub URLs with line number anchors

import type { Repository } from '$lib/api';

/**
 * Generates a GitHub URL for a specific file with line highlighting
 * @param repository - Repository object containing owner and name
 * @param filePath - Path to the file within the repository
 * @param startLine - Starting line number (optional)
 * @param endLine - Ending line number (optional)
 * @returns GitHub URL string
 */
export function generateGitHubUrl(
	repository: Repository,
	filePath: string,
	startLine?: number,
	endLine?: number
): string {
	const baseUrl = `https://github.com/${repository.owner}/${repository.name}/blob/main/${filePath}`;
	
	if (!startLine) {
		return baseUrl;
	}
	
	// GitHub uses #L format for line numbers
	if (endLine && endLine !== startLine) {
		return `${baseUrl}#L${startLine}-L${endLine}`;
	} else {
		return `${baseUrl}#L${startLine}`;
	}
}

/**
 * Opens a GitHub URL in a new tab
 * @param url - GitHub URL to open
 */
export function openGitHubUrl(url: string): void {
	window.open(url, '_blank', 'noopener,noreferrer');
}