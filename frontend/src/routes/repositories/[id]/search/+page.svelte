<!-- ABOUTME: Repository-specific search page for searching within a single repository -->
<!-- ABOUTME: Provides focused search interface with repository context and breadcrumbs -->

<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import SearchBox from '$lib/components/SearchBox.svelte';
	import SearchResults from '$lib/components/SearchResults.svelte';
	import SearchFilters from '$lib/components/SearchFilters.svelte';
	import { searchClient } from '$lib/api/search-client';
	import type { SearchResponse, SearchRequest } from '$lib/api/search-types';
	import type { components } from '$lib/api/types';

	type Repository = components['schemas']['Repository'];

	// Repository context
	let repository: Repository | null = null;
	let repositoryLoading = true;
	let repositoryError: string | null = null;

	// Search state
	let searchQuery = '';
	let searchResults: SearchResponse | null = null;
	let loading = false;
	let error: string | null = null;

	// Filter state (no repository filter needed since we're in a specific repo)
	let selectedLanguage = '';
	let selectedFileType = '';
	let availableLanguages: string[] = [];

	// Pagination
	let currentOffset = 0;
	const limit = 10;

	$: repositoryId = $page.params.id;

	onMount(async () => {
		await loadRepositoryData();
	});

	async function loadRepositoryData() {
		if (!repositoryId) return;

		repositoryLoading = true;
		repositoryError = null;

		try {
			// Load repository information
			const repoResponse = await fetch(`/api/repositories/${repositoryId}`);
			if (!repoResponse.ok) {
				throw new Error(
					`Failed to load repository: ${repoResponse.status} ${repoResponse.statusText}`
				);
			}
			repository = await repoResponse.json();

			// Load available languages for this repository
			const languagesData = await searchClient.getRepositoryLanguages(repositoryId);
			availableLanguages = languagesData.languages || [];
		} catch (err) {
			repositoryError = err instanceof Error ? err.message : 'Failed to load repository data';
			console.error('Repository loading error:', err);
		} finally {
			repositoryLoading = false;
		}
	}

	async function performSearch(query: string, offset = 0, append = false) {
		if (!query.trim() || !repositoryId) {
			searchResults = null;
			return;
		}

		loading = true;
		error = null;

		try {
			const searchRequest: SearchRequest = {
				query,
				limit,
				offset,
				language: selectedLanguage || undefined,
				fileType: selectedFileType || undefined
			};

			const data = await searchClient.searchRepository(repositoryId, searchRequest);

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
			loading = false;
		}
	}

	async function handleSearch(
		event: CustomEvent<{ query: string; mode: 'text' | 'vector' | 'hybrid' }>
	) {
		searchQuery = event.detail.query;
		currentOffset = 0;
		await performSearch(searchQuery);
	}

	async function handleLoadMore() {
		const newOffset = currentOffset + limit;
		await performSearch(searchQuery, newOffset, true);
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

	function handleClearFilters() {
		selectedLanguage = '';
		selectedFileType = '';
		handleFilterChange();
	}

	function handleResultSelect(event: CustomEvent) {
		const result = event.detail;
		// Navigate to the specific file/line (implementation depends on your routing setup)
		console.log('Selected result in repository:', repository?.name, result);
		// Example: navigate to `/repositories/${repositoryId}/files?path=${result.filePath}&line=${result.startLine}`
	}

	function handleRetry() {
		if (searchQuery) {
			performSearch(searchQuery);
		}
	}
</script>

<svelte:head>
	<title
		>{repository ? `Search in ${repository.name}` : 'Repository Search'} - GitHub Analyzer</title
	>
	<meta
		name="description"
		content={repository ? `Search code in ${repository.fullName}` : 'Search repository code'}
	/>
</svelte:head>

<div class="repository-search-page">
	{#if repositoryLoading}
		<div class="repository-loading">
			<div class="spinner"></div>
			<p>Loading repository...</p>
		</div>
	{:else if repositoryError}
		<div class="repository-error">
			<h2>Error Loading Repository</h2>
			<p>{repositoryError}</p>
			<button class="retry-button" on:click={loadRepositoryData}> Try Again </button>
		</div>
	{:else if repository}
		<!-- Repository Header -->
		<div class="repository-header">
			<nav class="breadcrumb">
				<a href="/repositories" class="breadcrumb-link">Repositories</a>
				<span class="breadcrumb-separator">/</span>
				<a href="/repositories/{repository.id}" class="breadcrumb-link">{repository.name}</a>
				<span class="breadcrumb-separator">/</span>
				<span class="breadcrumb-current">Search</span>
			</nav>

			<div class="header-content">
				<h1>Search in {repository.name}</h1>
				<p class="repository-description">
					{#if repository.description}
						{repository.description}
					{:else}
						Search through code in this repository to find functions, classes, and implementations.
					{/if}
				</p>

				<div class="repository-meta">
					<span class="repo-fullname">{repository.fullName}</span>
					{#if repository.primaryLanguage}
						<span class="repo-language">{repository.primaryLanguage}</span>
					{/if}
					{#if repository.stats?.totalFiles}
						<span class="repo-stats">{repository.stats.totalFiles} files</span>
					{/if}
				</div>
			</div>
		</div>

		<div class="search-container">
			<!-- Search Input -->
			<SearchBox
				placeholder="Search code in {repository.name}..."
				{loading}
				on:search={handleSearch}
				on:clear={() => {
					searchQuery = '';
					searchResults = null;
				}}
			/>

			<!-- Search Filters (no repository filter since we're in a specific repo) -->
			<SearchFilters
				bind:selectedLanguage
				bind:selectedFileType
				languages={availableLanguages}
				repositories={[]}
				disabled={loading}
				on:languageChange={handleLanguageChange}
				on:fileTypeChange={handleFileTypeChange}
				on:clearFilters={handleClearFilters}
			/>

			<!-- Search Results -->
			<SearchResults
				results={searchResults}
				{loading}
				{error}
				query={searchQuery}
				on:loadMore={handleLoadMore}
				on:selectResult={handleResultSelect}
				on:retry={handleRetry}
			/>
		</div>

		<!-- Repository-Specific Search Tips -->
		{#if !searchQuery && !loading}
			<div class="search-tips">
				<h3>Search Tips for {repository.name}</h3>
				<div class="tips-grid">
					<div class="tip-card">
						<h4>🎯 Focused Search</h4>
						<p>Search is limited to this repository, giving you more relevant results</p>
					</div>
					<div class="tip-card">
						<h4>📁 File Structure</h4>
						<p>Results will show exact file paths within {repository.name}</p>
					</div>
					{#if repository.primaryLanguage}
						<div class="tip-card">
							<h4>💻 {repository.primaryLanguage}</h4>
							<p>This repository primarily uses {repository.primaryLanguage}</p>
						</div>
					{/if}
					<div class="tip-card">
						<h4>🔍 Quick Access</h4>
						<p>Click on search results to navigate directly to the code</p>
					</div>
				</div>

				{#if repository.stats?.languages && Object.keys(repository.stats.languages).length > 1}
					<div class="language-breakdown">
						<h4>Languages in this repository:</h4>
						<div class="language-tags">
							{#each Object.entries(repository.stats.languages) as [language, count]}
								<span class="language-tag">
									{language}
									<span class="language-count">({count})</span>
								</span>
							{/each}
						</div>
					</div>
				{/if}
			</div>
		{/if}
	{/if}
</div>

<style>
	.repository-search-page {
		max-width: 1200px;
		margin: 0 auto;
		padding: 24px;
		min-height: calc(100vh - 80px);
	}

	.repository-loading {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 64px 20px;
		text-align: center;
	}

	.repository-loading .spinner {
		width: 32px;
		height: 32px;
		border: 3px solid #e5e7eb;
		border-top: 3px solid #3b82f6;
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin-bottom: 16px;
	}

	.repository-loading p {
		color: #6b7280;
		font-size: 16px;
	}

	.repository-error {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 64px 20px;
		text-align: center;
		color: #dc2626;
	}

	.repository-error h2 {
		margin: 0 0 8px 0;
		font-size: 24px;
		font-weight: 600;
	}

	.repository-error p {
		color: #6b7280;
		margin-bottom: 16px;
	}

	.retry-button {
		background-color: #3b82f6;
		color: white;
		border: none;
		padding: 8px 16px;
		border-radius: 6px;
		font-weight: 500;
		cursor: pointer;
		transition: background-color 0.2s;
	}

	.retry-button:hover {
		background-color: #2563eb;
	}

	.repository-header {
		margin-bottom: 48px;
	}

	.breadcrumb {
		margin-bottom: 24px;
		font-size: 14px;
	}

	.breadcrumb-link {
		color: #3b82f6;
		text-decoration: none;
		transition: color 0.2s;
	}

	.breadcrumb-link:hover {
		color: #1d4ed8;
	}

	.breadcrumb-separator {
		margin: 0 8px;
		color: #6b7280;
	}

	.breadcrumb-current {
		color: #1f2937;
		font-weight: 500;
	}

	.header-content h1 {
		font-size: 36px;
		font-weight: 700;
		color: #1f2937;
		margin: 0 0 12px 0;
	}

	.repository-description {
		font-size: 18px;
		color: #6b7280;
		margin: 0 0 16px 0;
		line-height: 1.5;
	}

	.repository-meta {
		display: flex;
		flex-wrap: wrap;
		gap: 16px;
		align-items: center;
		font-size: 14px;
	}

	.repo-fullname {
		font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
		background-color: #f3f4f6;
		color: #1f2937;
		padding: 4px 8px;
		border-radius: 4px;
		font-weight: 500;
	}

	.repo-language {
		background-color: #eff6ff;
		color: #1d4ed8;
		padding: 4px 8px;
		border-radius: 4px;
		font-weight: 500;
	}

	.repo-stats {
		color: #6b7280;
		font-weight: 500;
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
		grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
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

	.language-breakdown {
		margin-top: 32px;
		text-align: left;
		max-width: 600px;
		margin-left: auto;
		margin-right: auto;
	}

	.language-breakdown h4 {
		font-size: 16px;
		font-weight: 600;
		color: #1f2937;
		margin: 0 0 12px 0;
		text-align: center;
	}

	.language-tags {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		justify-content: center;
	}

	.language-tag {
		background-color: #f3f4f6;
		color: #1f2937;
		padding: 4px 8px;
		border-radius: 6px;
		font-size: 12px;
		font-weight: 500;
	}

	.language-count {
		color: #6b7280;
		font-weight: 400;
	}

	@keyframes spin {
		0% {
			transform: rotate(0deg);
		}
		100% {
			transform: rotate(360deg);
		}
	}

	/* Mobile responsiveness */
	@media (max-width: 768px) {
		.repository-search-page {
			padding: 16px;
		}

		.header-content h1 {
			font-size: 28px;
		}

		.repository-description {
			font-size: 16px;
		}

		.repository-meta {
			flex-direction: column;
			align-items: flex-start;
			gap: 8px;
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

		.breadcrumb {
			font-size: 13px;
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

		.repo-fullname {
			background-color: #e0e0e0;
			color: #000;
		}

		.repo-language {
			background-color: #e0f2fe;
			color: #000;
		}

		.language-tag {
			background-color: #e0e0e0;
			color: #000;
		}
	}

	/* Reduced motion support */
	@media (prefers-reduced-motion: reduce) {
		.tip-card,
		.retry-button,
		.breadcrumb-link,
		.repository-loading .spinner {
			transition: none;
			animation: none;
		}
	}
</style>
