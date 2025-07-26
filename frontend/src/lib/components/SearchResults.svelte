<!-- ABOUTME: Search results display component with pagination and result highlighting -->
<!-- ABOUTME: Handles loading states, error states, and empty results with helpful suggestions -->

<script lang="ts">
	import type { SearchResponse, SearchResult } from '../api/search-types';
	import CodeSnippet from './CodeSnippet.svelte';
	import { createEventDispatcher } from 'svelte';

	export let results: SearchResponse | null = null;
	export let loading = false;
	export let error: string | null = null;
	export let query = '';
	export let searchMode: 'text' | 'vector' | 'hybrid' = 'text';

	const dispatch = createEventDispatcher<{
		loadMore: void;
		selectResult: SearchResult;
		retry: void;
	}>();

	function handleLoadMore() {
		dispatch('loadMore');
	}

	function handleResultClick(result: SearchResult) {
		dispatch('selectResult', result);
	}

	function handleRetry() {
		dispatch('retry');
	}

	function getLanguageColor(language: string): string {
		const colors: Record<string, string> = {
			javascript: '#f7df1e',
			typescript: '#3178c6',
			python: '#3776ab',
			go: '#00add8',
			java: '#ed8b00',
			php: '#777bb4',
			cpp: '#00599c',
			csharp: '#239120',
			rust: '#000000',
			ruby: '#cc342d',
			html: '#e34f26',
			css: '#1572b6',
			json: '#000000',
			yaml: '#cb171e',
			shell: '#4eaa25',
			sql: '#336791'
		};
		return colors[language.toLowerCase()] || '#6b7280';
	}

	function formatScore(score: number, mode: 'text' | 'vector' | 'hybrid'): string {
		if (mode === 'text') {
			return score.toFixed(2);
		}
		// Vector and hybrid scores are percentages
		return (score * 100).toFixed(1) + '%';
	}

	function formatLineRange(startLine: number, endLine: number): string {
		if (startLine === endLine) {
			return `Line ${startLine}`;
		}
		return `Lines ${startLine}-${endLine}`;
	}

	function getScoreLabel(searchMode: 'text' | 'vector' | 'hybrid'): string {
		switch (searchMode) {
			case 'vector':
				return 'Similarity';
			case 'hybrid':
				return 'Relevance';
			default:
				return 'Score';
		}
	}

	function getRelevanceLevel(score: number, mode: 'text' | 'vector' | 'hybrid'): string {
		if (mode === 'vector' || mode === 'hybrid') {
			if (score >= 0.85) return 'high';
			if (score >= 0.6) return 'medium';
			return 'low';
		}
		// Text search score thresholds
		if (score >= 5) return 'high';
		if (score >= 2) return 'medium';
		return 'low';
	}

	function getRelevanceColor(level: string): string {
		switch (level) {
			case 'high':
				return '#10b981'; // green
			case 'medium':
				return '#f59e0b'; // amber
			case 'low':
				return '#6b7280'; // gray
			default:
				return '#6b7280';
		}
	}
</script>

<div class="search-results">
	{#if loading}
		<div class="loading-state">
			<div class="spinner"></div>
			<p>Searching code...</p>
		</div>
	{:else if error}
		<div class="error-state">
			<svg
				width="48"
				height="48"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
			>
				<circle cx="12" cy="12" r="10" />
				<line x1="15" y1="9" x2="9" y2="15" />
				<line x1="9" y1="9" x2="15" y2="15" />
			</svg>
			<h3>Search Error</h3>
			<p class="error-message">{error}</p>
			<button class="retry-button" on:click={handleRetry}> Try Again </button>
		</div>
	{:else if results && Array.isArray(results.results) && results.results.length > 0}
		<div class="results-header">
			<h3>Search Results</h3>
			<p class="results-meta">
				{results.total} result{results.total === 1 ? '' : 's'} for "{query}"
			</p>
		</div>

		<div class="results-list">
			{#each results.results as result, i (`${result.id}-${result.startLine}-${i}`)}
				<div
					class="result-item"
					on:click={() => handleResultClick(result)}
					on:keydown={(e) => e.key === 'Enter' && handleResultClick(result)}
					role="button"
					tabindex="0"
					aria-label="View code chunk in {result.fileName}"
				>
					<div class="result-header">
						<div class="file-info">
							<div class="file-path-container">
								<span class="file-path" title={result.filePath}>
									{result.filePath}
								</span>
								<button
									class="open-external"
									title="Open in new tab"
									aria-label="Open code chunk in new tab"
									on:click|stopPropagation={() => handleResultClick(result)}
								>
									<svg
										width="14"
										height="14"
										viewBox="0 0 24 24"
										fill="none"
										stroke="currentColor"
										stroke-width="2"
									>
										<path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
										<polyline points="15,3 21,3 21,9" />
										<line x1="10" y1="14" x2="21" y2="3" />
									</svg>
								</button>
							</div>
							<div class="file-meta">
								<span
									class="language-badge"
									style="background-color: {getLanguageColor(result.language)}"
								>
									{result.language}
								</span>
								<span class="line-range">
									{formatLineRange(result.startLine, result.endLine)}
								</span>
								<span
									class="relevance-score"
									class:semantic={searchMode === 'vector' || searchMode === 'hybrid'}
									style="color: {getRelevanceColor(getRelevanceLevel(result.score, searchMode))}"
									title="{getScoreLabel(searchMode)}: {formatScore(result.score, searchMode)}"
								>
									{#if searchMode === 'vector'}
										<svg
											width="12"
											height="12"
											viewBox="0 0 24 24"
											fill="none"
											stroke="currentColor"
											stroke-width="2"
										>
											<path d="M9 19c-5 0-8-2.5-8-5s3-5 8-5 8 2.5 8 5-3 5-8 5Z" />
											<path d="m8 19 8-14" />
											<path d="m1 14 8-14" />
											<path d="m15 5 4 14" />
										</svg>
									{:else if searchMode === 'hybrid'}
										<svg
											width="12"
											height="12"
											viewBox="0 0 24 24"
											fill="none"
											stroke="currentColor"
											stroke-width="2"
										>
											<path d="M12 3v18m0-18 4 4m-4-4-4 4" />
											<path d="m8 17 4 4 4-4" />
										</svg>
									{/if}
									{getScoreLabel(searchMode)}: {formatScore(result.score, searchMode)}
								</span>
							</div>
						</div>
					</div>

					<div class="code-preview">
						<CodeSnippet
							content={result.highlight && result.highlight.trim().length > 0
								? result.highlight
								: result.content}
							language={result.language}
							searchTerm={query}
							maxLines={8}
							showLineNumbers={true}
							startLine={result.startLine}
						/>
					</div>

					{#if result.metadata?.functions && result.metadata.functions.length > 0}
						<div class="metadata">
							<span class="metadata-label">Functions:</span>
							<div class="metadata-items">
								{#each result.metadata.functions.slice(0, 3) as func (func)}
									<span class="metadata-item">{func}</span>
								{/each}
								{#if result.metadata.functions.length > 3}
									<span class="metadata-more">
										+{result.metadata.functions.length - 3} more
									</span>
								{/if}
							</div>
						</div>
					{/if}

					{#if result.metadata?.classes && result.metadata.classes.length > 0}
						<div class="metadata">
							<span class="metadata-label">Classes:</span>
							<div class="metadata-items">
								{#each result.metadata.classes.slice(0, 3) as cls (cls)}
									<span class="metadata-item">{cls}</span>
								{/each}
								{#if result.metadata.classes.length > 3}
									<span class="metadata-more">
										+{result.metadata.classes.length - 3} more
									</span>
								{/if}
							</div>
						</div>
					{/if}
				</div>
			{/each}
		</div>

		{#if results.hasMore}
			<div class="load-more">
				<button class="load-more-button" on:click={handleLoadMore} disabled={loading}>
					{loading ? 'Loading...' : 'Load More Results'}
				</button>
			</div>
		{/if}
	{:else if query}
		<div class="no-results">
			<svg
				width="64"
				height="64"
				viewBox="0 0 24 24"
				fill="none"
				stroke="currentColor"
				stroke-width="1.5"
			>
				<circle cx="11" cy="11" r="8" />
				<path d="M21 21l-4.35-4.35" />
			</svg>
			<h3>No Results Found</h3>
			<p>No code chunks found for "{query}"</p>
			<div class="suggestions">
				<h4>Try:</h4>
				<ul>
					<li>Different keywords or phrases</li>
					<li>Broader search terms</li>
					<li>Checking your spelling</li>
					<li>Searching for function or class names</li>
					<li>Using specific programming constructs</li>
				</ul>
			</div>
		</div>
	{/if}
</div>

<style>
	.search-results {
		width: 100%;
		max-width: 900px;
		margin: 0 auto;
	}
</style>
