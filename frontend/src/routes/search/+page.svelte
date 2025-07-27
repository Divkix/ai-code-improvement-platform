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
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Alert from '$lib/components/ui/alert/index.js';

	// Generic response shape used for API client calls
	type ApiResponse = { error?: { message?: string }; data?: unknown };

	// Pipeline stats polling
	let pipelineStats: {
		pending: number;
		processing: number;
		completed: number;
		failed: number;
	} | null = $state(null);
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
	let searchQuery = $state('');
	let searchMode: 'text' | 'vector' | 'hybrid' = $state('text');
	let searchResults: SearchResponse | null = $state(null);
	let loading = $state(false);
	let error: string | null = $state(null);

	// Filter state
	let selectedLanguage = $state('');
	let selectedFileType = $state('');
	let selectedRepository = $state('');
	let availableLanguages: string[] = $state([]);
	let availableRepositories: Repository[] = $state([]);

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
	<title>Search Code - ACIP</title>
	<meta
		name="description"
		content="Search through your code repositories to find functions, classes, and implementations quickly."
	/>
</svelte:head>

<div class="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
	<!-- Search Header -->
	<div class="mb-12 text-center">
		<h1 class="text-4xl font-bold tracking-tight text-foreground sm:text-5xl">Search Code</h1>
		<p class="mx-auto mt-3 max-w-2xl text-lg text-muted-foreground sm:mt-4">
			Find functions, classes, variables, and any code patterns across all your repositories.
		</p>
	</div>

	<div class="mx-auto max-w-4xl">
		{#if pipelineStats && pipelineStats.pending + pipelineStats.processing > 0}
			<Alert.Root class="mb-6">
				<Alert.Description>
					⚡ Vector embeddings are currently being generated for your repositories. Results may be
					incomplete until processing finishes.
				</Alert.Description>
			</Alert.Root>
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
		<div class="mx-auto mt-12 max-w-4xl">
			<h3 class="mb-8 text-center text-2xl font-semibold">Search Tips</h3>
			<div class="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
				<Card.Root class="transition-shadow hover:shadow-md">
					<Card.Header>
						<Card.Title class="text-base">🔍 Function Names</Card.Title>
					</Card.Header>
					<Card.Content>
						<p class="text-sm text-muted-foreground">
							Search for specific functions like <code class="rounded bg-muted px-1 py-0.5 text-xs"
								>getUserData</code
							>
							or
							<code class="rounded bg-muted px-1 py-0.5 text-xs">calculateTotal</code>
						</p>
					</Card.Content>
				</Card.Root>

				<Card.Root class="transition-shadow hover:shadow-md">
					<Card.Header>
						<Card.Title class="text-base">🏗️ Class Names</Card.Title>
					</Card.Header>
					<Card.Content>
						<p class="text-sm text-muted-foreground">
							Find classes with names like <code class="rounded bg-muted px-1 py-0.5 text-xs"
								>UserService</code
							>
							or <code class="rounded bg-muted px-1 py-0.5 text-xs">DatabaseConnection</code>
						</p>
					</Card.Content>
				</Card.Root>

				<Card.Root class="transition-shadow hover:shadow-md">
					<Card.Header>
						<Card.Title class="text-base">📝 Code Patterns</Card.Title>
					</Card.Header>
					<Card.Content>
						<p class="text-sm text-muted-foreground">
							Search for patterns like <code class="rounded bg-muted px-1 py-0.5 text-xs"
								>async function</code
							>
							or <code class="rounded bg-muted px-1 py-0.5 text-xs">try catch</code>
						</p>
					</Card.Content>
				</Card.Root>

				<Card.Root class="transition-shadow hover:shadow-md">
					<Card.Header>
						<Card.Title class="text-base">🔤 Variable Names</Card.Title>
					</Card.Header>
					<Card.Content>
						<p class="text-sm text-muted-foreground">
							Find variables like <code class="rounded bg-muted px-1 py-0.5 text-xs">apiKey</code>
							or <code class="rounded bg-muted px-1 py-0.5 text-xs">databaseUrl</code>
						</p>
					</Card.Content>
				</Card.Root>

				<Card.Root class="transition-shadow hover:shadow-md">
					<Card.Header>
						<Card.Title class="text-base">🎯 Specific Terms</Card.Title>
					</Card.Header>
					<Card.Content>
						<p class="text-sm text-muted-foreground">
							Search for keywords like <code class="rounded bg-muted px-1 py-0.5 text-xs"
								>authentication</code
							>
							or <code class="rounded bg-muted px-1 py-0.5 text-xs">validation</code>
						</p>
					</Card.Content>
				</Card.Root>

				<Card.Root class="transition-shadow hover:shadow-md">
					<Card.Header>
						<Card.Title class="text-base">⚙️ Use Filters</Card.Title>
					</Card.Header>
					<Card.Content>
						<p class="text-sm text-muted-foreground">
							Narrow results by programming language, file type, or repository
						</p>
					</Card.Content>
				</Card.Root>
			</div>
		</div>
	{/if}
</div>
