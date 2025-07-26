<!-- ABOUTME: Global search page for searching across all repositories and code chunks -->
<!-- ABOUTME: Provides comprehensive search interface with filters and pagination -->

<script lang="ts">
	import { onMount } from 'svelte';
	import SearchBox from '$lib/components/SearchBox.svelte';
	import SearchResults from '$lib/components/SearchResults.svelte';
	import SearchFilters from '$lib/components/SearchFilters.svelte';
	import { apiClient, vectorSearchAPI } from '$lib/api/client';
	import type { SearchResponse } from '$lib/api/search-types';
	import type { components } from '$lib/api/types';
	import { onDestroy } from 'svelte';
	import { generateGitHubUrl, openGitHubUrl } from '$lib/utils/github';

	// Generic response shape used for API client calls
	type ApiResponse = { error?: { message?: string }; data?: unknown };

	// Pipeline stats polling
	let pipelineStats: {
		pending: number;
		processing: number;
		completed: number;
		failed: number;
	} | null = null;
	let statsInterval: NodeJS.Timeout;

	async function fetchPipelineStats() {
		try {
			pipelineStats = await vectorSearchAPI.getPipelineStats();
		} catch {
			// ignore
		}
	}

	onMount(() => {
		fetchPipelineStats();
		statsInterval = setInterval(fetchPipelineStats, 5000);
	});

	onDestroy(() => {
		if (statsInterval) clearInterval(statsInterval);
	});

	type Repository = components['schemas']['Repository'];

	// Search state
	let searchQuery = '';
	let searchMode: 'text' | 'vector' | 'hybrid' = 'text';
	let searchResults: SearchResponse | null = null;
	let loading = false;
	let error: string | null = null;

	// Filter state
	let selectedLanguage = '';
	let selectedFileType = '';
	let selectedRepository = '';
	let availableLanguages: string[] = [];
	let availableRepositories: Repository[] = [];

	// Pagination
	let currentOffset = 0;
	const limit = 10;

	// Track latest search to prevent race conditions where slower previous
	// requests overwrite the most recent results. A simple incrementing id
	// is sufficient because performSearch is always called synchronously
	// from the event handlers in this component.
	let latestSearchId = 0;

	onMount(async () => {
		await loadInitialData();
	});

	async function loadInitialData() {
		try {
			// Load available languages and repositories for filters
			const [languagesResponse, repositoriesResponse] = await Promise.all([
				apiClient.GET('/api/search/languages').catch(() => ({ data: null, error: null })),
				apiClient.GET('/api/repositories').catch(() => ({ data: null, error: null }))
			]);

			if (languagesResponse.data && languagesResponse.data.languages) {
				availableLanguages = languagesResponse.data.languages;
			} else {
				availableLanguages = [];
			}

			if (repositoriesResponse.data && repositoriesResponse.data.repositories) {
				availableRepositories = repositoriesResponse.data.repositories;
			} else {
				availableRepositories = [];
			}
		} catch (err) {
			console.warn('Failed to load initial data:', err);
			// Ensure arrays are never null
			availableLanguages = [];
			availableRepositories = [];
		}
	}

	async function performSearch(
		query: string,
		mode: 'text' | 'vector' | 'hybrid' = searchMode,
		offset = 0,
		append = false
	) {
		// Increment the global search id and capture a local copy for this call.
		const searchId = ++latestSearchId;

		if (!query.trim()) {
			searchResults = null;
			return;
		}

		loading = true;
		error = null;

		try {
			let response: unknown;

			if (mode === 'vector') {
				// Vector search using apiClient
				response = await apiClient.POST('/api/search/vector', {
					body: {
						query,
						repositoryId: selectedRepository || undefined,
						language: selectedLanguage || undefined,
						fileType: selectedFileType || undefined,
						limit,
						offset
					}
				});
			} else if (mode === 'hybrid') {
				// Hybrid search using apiClient
				response = await apiClient.POST('/api/search/hybrid', {
					body: {
						query,
						repositoryId: selectedRepository || undefined,
						language: selectedLanguage || undefined,
						fileType: selectedFileType || undefined,
						vectorWeight: 0.7,
						textWeight: 0.3,
						limit,
						offset
					}
				});
			} else {
				// Traditional text search using apiClient
				response = await apiClient.POST('/api/search', {
					body: {
						query,
						limit,
						offset,
						language: selectedLanguage || undefined,
						fileType: selectedFileType || undefined,
						repositoryId: selectedRepository || undefined
					}
				});
			}

			const apiRes = response as ApiResponse;
			if (apiRes.error) {
				throw new Error(apiRes.error.message || 'Search failed');
			}

			const data = apiRes.data as SearchResponse;

			// If a newer search has been initiated while this one was awaiting,
			// discard this response to avoid overwriting fresher results.
			if (searchId !== latestSearchId) {
				return;
			}

			if (append && searchResults) {
				// Append results for pagination
				searchResults = {
					...data,
					results: [...searchResults.results, ...data.results]
				};
			} else {
				searchResults = data;
			}

			currentOffset = offset;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Search failed';
			console.error('Search error:', err);
		} finally {
			// Only clear the loading state if this search is still the latest one.
			if (searchId === latestSearchId) {
				loading = false;
			}
		}
	}

	async function handleSearch(
		event: CustomEvent<{ query: string; mode: 'text' | 'vector' | 'hybrid' }>
	) {
		searchQuery = event.detail.query;
		searchMode = event.detail.mode;
		currentOffset = 0;
		await performSearch(searchQuery, searchMode);
	}

	function handleModeChange(event: CustomEvent<'text' | 'vector' | 'hybrid'>) {
		searchMode = event.detail;
		if (searchQuery) {
			performSearch(searchQuery, searchMode);
		}
	}

	async function handleLoadMore() {
		const newOffset = currentOffset + limit;
		await performSearch(searchQuery, searchMode, newOffset, true);
	}

	async function handleFilterChange() {
		// Reset pagination and search with new filters
		currentOffset = 0;
		await performSearch(searchQuery);
	}

	function handleLanguageChange(event: CustomEvent<string>) {
		selectedLanguage = event.detail;
		handleFilterChange();
	}

	function handleFileTypeChange(event: CustomEvent<string>) {
		selectedFileType = event.detail;
		handleFilterChange();
	}

	function handleRepositoryChange(event: CustomEvent<string>) {
		selectedRepository = event.detail;
		handleFilterChange();
	}

	function handleClearFilters() {
		selectedLanguage = '';
		selectedFileType = '';
		selectedRepository = '';
		handleFilterChange();
	}

	function handleResultSelect(event: CustomEvent) {
		const result = event.detail;
		// Find the repository information from availableRepositories
		const repository = availableRepositories.find((repo) => repo.id === result.repositoryId);

		if (repository) {
			// Generate GitHub URL with line highlighting
			const githubUrl = generateGitHubUrl(
				repository,
				result.filePath,
				result.startLine,
				result.endLine
			);
			openGitHubUrl(githubUrl);
		} else {
			console.error('Repository not found for ID:', result.repositoryId);
		}
	}

	function handleRetry() {
		if (searchQuery) {
			performSearch(searchQuery);
		}
	}
</script>

<svelte:head>
	<title>Search Code - GitHub Analyzer</title>
	<meta
		name="description"
		content="Search through your code repositories to find functions, classes, and implementations quickly."
	/>
</svelte:head>

<div class="search-page">
	<div class="search-header">
		<div class="header-content">
			<h1>Search Code</h1>
			<p class="search-description">
				Find functions, classes, variables, and any code patterns across all your repositories.
			</p>
		</div>
	</div>

	<div class="search-container">
		{#if pipelineStats && pipelineStats.pending + pipelineStats.processing > 0}
			<div class="embedding-banner" role="status" aria-live="polite">
				⚡ Vector embeddings are currently being generated for your repositories. Results may be
				incomplete until processing finishes.
			</div>
		{/if}

		<!-- Search Input -->
		<SearchBox
			placeholder="Search for functions, classes, variables, or any code pattern..."
			{loading}
			bind:searchMode
			showModeSelector={true}
			on:search={handleSearch}
			on:modeChange={handleModeChange}
			on:clear={() => {
				searchQuery = '';
				searchResults = null;
			}}
		/>

		<!-- Search Filters -->
		<SearchFilters
			bind:selectedLanguage
			bind:selectedFileType
			bind:selectedRepository
			languages={availableLanguages}
			repositories={availableRepositories}
			disabled={loading}
			on:languageChange={handleLanguageChange}
			on:fileTypeChange={handleFileTypeChange}
			on:repositoryChange={handleRepositoryChange}
			on:clearFilters={handleClearFilters}
		/>

		<!-- Search Results -->
		<SearchResults
			results={searchResults}
			{loading}
			{error}
			query={searchQuery}
			{searchMode}
			on:loadMore={handleLoadMore}
			on:selectResult={handleResultSelect}
			on:retry={handleRetry}
		/>
	</div>

	<!-- Search Tips -->
	{#if !searchQuery && !loading}
		<div class="search-tips">
			<h3>Search Tips</h3>
			<div class="tips-grid">
				<div class="tip-card">
					<h4>🔍 Function Names</h4>
					<p>
						Search for specific functions like <code>getUserData</code> or
						<code>calculateTotal</code>
					</p>
				</div>
				<div class="tip-card">
					<h4>🏗️ Class Names</h4>
					<p>
						Find classes with names like <code>UserService</code> or <code>DatabaseConnection</code>
					</p>
				</div>
				<div class="tip-card">
					<h4>📝 Code Patterns</h4>
					<p>Search for patterns like <code>async function</code> or <code>try catch</code></p>
				</div>
				<div class="tip-card">
					<h4>🔤 Variable Names</h4>
					<p>Find variables like <code>apiKey</code> or <code>databaseUrl</code></p>
				</div>
				<div class="tip-card">
					<h4>🎯 Specific Terms</h4>
					<p>Search for keywords like <code>authentication</code> or <code>validation</code></p>
				</div>
				<div class="tip-card">
					<h4>⚙️ Use Filters</h4>
					<p>Narrow results by programming language, file type, or repository</p>
				</div>
			</div>
		</div>
	{/if}
</div>

<style>
	.search-page {
		max-width: 1200px;
		margin: 0 auto;
		padding: 24px;
		min-height: calc(100vh - 80px);
	}

	.search-header {
		text-align: center;
		margin-bottom: 48px;
	}

	.header-content h1 {
		font-size: 36px;
		font-weight: 700;
		color: #1f2937;
		margin: 0 0 12px 0;
	}

	.search-description {
		font-size: 18px;
		color: #6b7280;
		margin: 0;
		max-width: 600px;
		margin-left: auto;
		margin-right: auto;
	}

	.search-container {
		max-width: 900px;
		margin: 0 auto;
	}

	.search-tips {
		max-width: 900px;
		margin: 48px auto 0;
		text-align: center;
	}

	.search-tips h3 {
		font-size: 24px;
		font-weight: 600;
		color: #1f2937;
		margin: 0 0 24px 0;
	}

	.tips-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
		gap: 20px;
		margin-top: 24px;
	}

	.tip-card {
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 12px;
		padding: 20px;
		text-align: left;
		transition: all 0.2s;
	}

	.tip-card:hover {
		border-color: #3b82f6;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}

	.tip-card h4 {
		font-size: 16px;
		font-weight: 600;
		color: #1f2937;
		margin: 0 0 8px 0;
	}

	.tip-card p {
		font-size: 14px;
		color: #6b7280;
		line-height: 1.5;
		margin: 0;
	}

	.tip-card code {
		background-color: #f3f4f6;
		color: #1f2937;
		padding: 2px 4px;
		border-radius: 3px;
		font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
		font-size: 13px;
	}

	/* Mobile responsiveness */
	@media (max-width: 768px) {
		.search-page {
			padding: 16px;
		}

		.header-content h1 {
			font-size: 28px;
		}

		.search-description {
			font-size: 16px;
		}

		.tips-grid {
			grid-template-columns: 1fr;
			gap: 16px;
		}

		.search-tips {
			margin-top: 32px;
		}

		.search-tips h3 {
			font-size: 20px;
		}
	}

	/* High contrast mode support */
	@media (prefers-contrast: high) {
		.tip-card {
			border-color: #000;
		}

		.tip-card:hover {
			border-color: #0066cc;
		}

		.tip-card code {
			background-color: #e0e0e0;
			color: #000;
		}
	}

	/* Reduced motion support */
	@media (prefers-reduced-motion: reduce) {
		.tip-card {
			transition: none;
		}
	}
</style>
