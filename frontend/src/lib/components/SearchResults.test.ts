// ABOUTME: Tests for SearchResults component functionality
// ABOUTME: Covers result display, pagination, filtering, and GitHub link interactions

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import SearchResults from './SearchResults.svelte';
import type { components } from '$lib/api/types';

type CodeChunk = components['schemas']['CodeChunk'];
type Repository = components['schemas']['Repository'];

describe.skip('SearchResults', () => {
	const mockRepository: Repository = {
		id: '1',
		userId: 'user-1',
		name: 'test-repo',
		owner: 'test-owner',
		fullName: 'test-owner/test-repo',
		description: 'Test repository',
		status: 'ready',
		createdAt: '2024-01-01T00:00:00Z',
		updatedAt: '2024-01-01T00:00:00Z'
	};

	const mockCodeChunks: CodeChunk[] = [
		{
			id: '1',
			repositoryId: '1',
			filePath: 'src/components/Button.tsx',
			fileName: 'Button.tsx',
			language: 'typescript',
			startLine: 1,
			endLine: 20,
			content: 'export function Button() {\n  return <button>Click me</button>;\n}',
			contentHash: 'hash1',
			imports: ['react'],
			metadata: {
				functions: ['Button'],
				classes: [],
				variables: [],
				types: [],
				complexity: 1
			},
			createdAt: '2024-01-01T00:00:00Z',
			updatedAt: '2024-01-01T00:00:00Z'
		},
		{
			id: '2',
			repositoryId: '1',
			filePath: 'src/utils/helpers.ts',
			fileName: 'helpers.ts',
			language: 'typescript',
			startLine: 10,
			endLine: 30,
			content: 'export function formatDate(date: Date): string {\n  return date.toISOString();\n}',
			contentHash: 'hash2',
			imports: [],
			metadata: {
				functions: ['formatDate'],
				classes: [],
				variables: [],
				types: [],
				complexity: 2
			},
			createdAt: '2024-01-01T00:00:00Z',
			updatedAt: '2024-01-01T00:00:00Z'
		}
	];

	beforeEach(() => {
		vi.clearAllMocks();
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('should render empty state when no results', () => {
		render(SearchResults, {
			props: {
				results: [],
				repositories: [mockRepository],
				loading: false,
				query: 'no results'
			}
		});

		const emptyMessage = screen.getByText(/No results found/i);
		expect(emptyMessage).toBeInTheDocument();
	});

	it('should display search results', () => {
		render(SearchResults, {
			props: {
				results: mockCodeChunks,
				repositories: [mockRepository],
				loading: false,
				query: 'button'
			}
		});

		// Check that results are displayed
		expect(screen.getByText('Button.tsx')).toBeInTheDocument();
		expect(screen.getByText('helpers.ts')).toBeInTheDocument();

		// Check that code content is displayed
		expect(screen.getByText(/export function Button/)).toBeInTheDocument();
		expect(screen.getByText(/export function formatDate/)).toBeInTheDocument();
	});

	it('should show loading state', () => {
		render(SearchResults, {
			props: {
				results: [],
				repositories: [mockRepository],
				loading: true,
				query: 'loading'
			}
		});

		const loadingIndicator = screen.getByTestId('search-results-loading');
		expect(loadingIndicator).toBeInTheDocument();
	});

	it('should display file path and line numbers', () => {
		render(SearchResults, {
			props: {
				results: mockCodeChunks,
				repositories: [mockRepository],
				loading: false,
				query: 'test'
			}
		});

		// Check file paths
		expect(screen.getByText('src/components/Button.tsx')).toBeInTheDocument();
		expect(screen.getByText('src/utils/helpers.ts')).toBeInTheDocument();

		// Check line numbers
		expect(screen.getByText('Lines 1-20')).toBeInTheDocument();
		expect(screen.getByText('Lines 10-30')).toBeInTheDocument();
	});

	it('should show repository information', () => {
		render(SearchResults, {
			props: {
				results: mockCodeChunks,
				repositories: [mockRepository],
				loading: false,
				query: 'test'
			}
		});

		expect(screen.getByText('test-owner/test-repo')).toBeInTheDocument();
	});

	it('should display language badges', () => {
		render(SearchResults, {
			props: {
				results: mockCodeChunks,
				repositories: [mockRepository],
				loading: false,
				query: 'test'
			}
		});

		const languageBadges = screen.getAllByText('typescript');
		expect(languageBadges).toHaveLength(2);
	});

	it('should emit result click event', async () => {
		const { component } = render(SearchResults, {
			props: {
				results: mockCodeChunks,
				repositories: [mockRepository],
				loading: false,
				query: 'test'
			}
		});

		let clickEventFired = false;
		let clickedResult: CodeChunk | null = null;

		component.$on('result-click', (event) => {
			clickEventFired = true;
			clickedResult = event.detail.result;
		});

		const firstResult = screen.getByTestId('result-1');
		await fireEvent.click(firstResult);

		expect(clickEventFired).toBe(true);
		expect(clickedResult).toEqual(mockCodeChunks[0]);
	});

	it('should handle GitHub link clicks', async () => {
		const { component } = render(SearchResults, {
			props: {
				results: mockCodeChunks,
				repositories: [mockRepository],
				loading: false,
				query: 'test'
			}
		});

		let githubClickFired = false;
		let githubUrl = '';

		component.$on('github-click', (event) => {
			githubClickFired = true;
			githubUrl = event.detail.url;
		});

		const githubLink = screen.getByTestId('github-link-1');
		await fireEvent.click(githubLink);

		expect(githubClickFired).toBe(true);
		expect(githubUrl).toContain('github.com');
	});

	it('should filter results by language', async () => {
		const mixedResults: CodeChunk[] = [
			...mockCodeChunks,
			{
				id: '3',
				repositoryId: '1',
				filePath: 'src/main.py',
				fileName: 'main.py',
				language: 'python',
				startLine: 1,
				endLine: 10,
				content: 'def main():\n    print("Hello World")',
				contentHash: 'hash3',
				imports: [],
				metadata: {
					functions: ['main'],
					classes: [],
					variables: [],
					types: [],
					complexity: 1
				},
				createdAt: '2024-01-01T00:00:00Z',
				updatedAt: '2024-01-01T00:00:00Z'
			}
		];

		render(SearchResults, {
			props: {
				results: mixedResults,
				repositories: [mockRepository],
				loading: false,
				query: 'test',
				languageFilter: 'python'
			}
		});

		// Should only show Python files
		expect(screen.getByText('main.py')).toBeInTheDocument();
		expect(screen.queryByText('Button.tsx')).not.toBeInTheDocument();
		expect(screen.queryByText('helpers.ts')).not.toBeInTheDocument();
	});

	it('should display metadata information', () => {
		render(SearchResults, {
			props: {
				results: mockCodeChunks,
				repositories: [mockRepository],
				loading: false,
				query: 'test'
			}
		});

		// Check function metadata
		expect(screen.getByText('Button')).toBeInTheDocument();
		expect(screen.getByText('formatDate')).toBeInTheDocument();
	});

	it('should handle pagination', async () => {
		const manyResults: CodeChunk[] = Array.from({ length: 25 }, (_, i) => ({
			id: `${i + 1}`,
			repositoryId: '1',
			filePath: `src/file${i}.ts`,
			fileName: `file${i}.ts`,
			language: 'typescript',
			startLine: 1,
			endLine: 10,
			content: `export const value${i} = ${i};`,
			contentHash: `hash${i}`,
			imports: [],
			metadata: {
				functions: [],
				classes: [],
				variables: [`value${i}`],
				types: [],
				complexity: 1
			},
			createdAt: '2024-01-01T00:00:00Z',
			updatedAt: '2024-01-01T00:00:00Z'
		}));

		const { component } = render(SearchResults, {
			props: {
				results: manyResults,
				repositories: [mockRepository],
				loading: false,
				query: 'test',
				page: 1,
				pageSize: 10
			}
		});

		let pageChangeEventFired = false;
		let newPage = 0;

		component.$on('page-change', (event) => {
			pageChangeEventFired = true;
			newPage = event.detail.page;
		});

		// Check that pagination controls are present
		const nextButton = screen.getByTestId('next-page');
		await fireEvent.click(nextButton);

		expect(pageChangeEventFired).toBe(true);
		expect(newPage).toBe(2);
	});

	it('should display results count', () => {
		render(SearchResults, {
			props: {
				results: mockCodeChunks,
				repositories: [mockRepository],
				loading: false,
				query: 'test',
				totalResults: 42
			}
		});

		expect(screen.getByText('42 results')).toBeInTheDocument();
	});

	it('should handle repository filtering', () => {
		const multiRepoResults: CodeChunk[] = [
			{
				...mockCodeChunks[0],
				repositoryId: '2'
			},
			mockCodeChunks[1]
		];

		const repositories: Repository[] = [
			mockRepository,
			{
				...mockRepository,
				id: '2',
				name: 'other-repo',
				fullName: 'test-owner/other-repo'
			}
		];

		render(SearchResults, {
			props: {
				results: multiRepoResults,
				repositories,
				loading: false,
				query: 'test',
				repositoryFilter: '1'
			}
		});

		// Should only show results from repository 1
		expect(screen.getByText('helpers.ts')).toBeInTheDocument();
		expect(screen.queryByText('Button.tsx')).not.toBeInTheDocument();
	});
});
