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
	{:else if results && results.results.length > 0}
		<div class="results-header">
			<h3>Search Results</h3>
			<p class="results-meta">
				{results.total} result{results.total === 1 ? '' : 's'} for "{query}"
			</p>
		</div>

		<div class="results-list">
			{#each results.results as result (result.id)}
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
							content={result.highlight || result.content}
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
								{#each result.metadata.functions.slice(0, 3) as func}
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
								{#each result.metadata.classes.slice(0, 3) as cls}
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

	.loading-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 64px 20px;
		text-align: center;
	}

	.loading-state .spinner {
		width: 32px;
		height: 32px;
		border: 3px solid #e5e7eb;
		border-top: 3px solid #3b82f6;
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin-bottom: 16px;
	}

	.loading-state p {
		color: #6b7280;
		font-size: 16px;
	}

	.error-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 64px 20px;
		text-align: center;
		color: #dc2626;
	}

	.error-state svg {
		margin-bottom: 16px;
	}

	.error-state h3 {
		margin: 0 0 8px 0;
		font-size: 20px;
		font-weight: 600;
	}

	.error-message {
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

	.results-header {
		margin-bottom: 24px;
	}

	.results-header h3 {
		font-size: 24px;
		font-weight: 600;
		margin: 0 0 8px 0;
		color: #1f2937;
	}

	.results-meta {
		color: #6b7280;
		font-size: 14px;
		margin: 0;
	}

	.results-list {
		display: flex;
		flex-direction: column;
		gap: 16px;
	}

	.result-item {
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 8px;
		padding: 20px;
		cursor: pointer;
		transition: all 0.2s;
		position: relative;
	}

	.result-item:hover {
		border-color: #3b82f6;
		box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
	}

	.result-item:focus {
		outline: none;
		border-color: #3b82f6;
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
	}

	.result-header {
		margin-bottom: 16px;
	}

	.file-info {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}

	.file-path-container {
		display: flex;
		align-items: center;
		gap: 8px;
	}

	.file-path {
		font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
		font-size: 14px;
		font-weight: 500;
		color: #1f2937;
		flex: 1;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.open-external {
		background: none;
		border: none;
		color: #6b7280;
		cursor: pointer;
		padding: 2px;
		border-radius: 3px;
		transition: color 0.2s;
		flex-shrink: 0;
	}

	.open-external:hover {
		color: #3b82f6;
	}

	.file-meta {
		display: flex;
		flex-wrap: wrap;
		gap: 12px;
		align-items: center;
	}

	.language-badge {
		display: inline-block;
		padding: 2px 8px;
		border-radius: 12px;
		font-size: 12px;
		font-weight: 500;
		color: white;
		text-transform: capitalize;
	}

	.line-range {
		font-size: 12px;
		color: #6b7280;
		font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
	}

	.relevance-score {
		font-size: 12px;
		color: #6b7280;
		padding: 2px 6px;
		background-color: #f3f4f6;
		border-radius: 4px;
		display: inline-flex;
		align-items: center;
		gap: 4px;
		font-weight: 500;
	}

	.relevance-score.semantic {
		background-color: rgba(59, 130, 246, 0.1);
		border: 1px solid rgba(59, 130, 246, 0.2);
	}

	.relevance-score svg {
		opacity: 0.8;
	}

	.code-preview {
		margin-bottom: 16px;
	}

	.metadata {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 12px;
		padding-top: 12px;
		border-top: 1px solid #f3f4f6;
		margin-top: 12px;
	}

	.metadata:first-of-type {
		border-top: none;
		padding-top: 0;
		margin-top: 0;
	}

	.metadata-label {
		color: #6b7280;
		font-weight: 500;
		flex-shrink: 0;
	}

	.metadata-items {
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
		align-items: center;
	}

	.metadata-item {
		background-color: #eff6ff;
		color: #1d4ed8;
		padding: 2px 6px;
		border-radius: 4px;
		font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
	}

	.metadata-more {
		color: #6b7280;
		font-style: italic;
	}

	.load-more {
		text-align: center;
		margin-top: 32px;
	}

	.load-more-button {
		background-color: #3b82f6;
		color: white;
		border: none;
		padding: 12px 24px;
		border-radius: 6px;
		font-weight: 500;
		cursor: pointer;
		transition: background-color 0.2s;
	}

	.load-more-button:hover:not(:disabled) {
		background-color: #2563eb;
	}

	.load-more-button:disabled {
		background-color: #9ca3af;
		cursor: not-allowed;
	}

	.no-results {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 64px 20px;
		text-align: center;
		color: #6b7280;
	}

	.no-results svg {
		margin-bottom: 24px;
		color: #d1d5db;
	}

	.no-results h3 {
		color: #1f2937;
		margin: 0 0 8px 0;
		font-size: 20px;
		font-weight: 600;
	}

	.no-results > p {
		margin-bottom: 24px;
		font-size: 16px;
	}

	.suggestions {
		text-align: left;
		max-width: 300px;
	}

	.suggestions h4 {
		color: #1f2937;
		margin: 0 0 12px 0;
		font-size: 16px;
		font-weight: 500;
	}

	.suggestions ul {
		list-style-type: disc;
		padding-left: 20px;
		margin: 0;
	}

	.suggestions li {
		margin-bottom: 4px;
		font-size: 14px;
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
		.result-item {
			padding: 16px;
		}

		.file-meta {
			flex-direction: column;
			align-items: flex-start;
			gap: 8px;
		}

		.results-header h3 {
			font-size: 20px;
		}

		.file-path-container {
			flex-direction: column;
			align-items: flex-start;
			gap: 4px;
		}

		.metadata {
			flex-direction: column;
			align-items: flex-start;
			gap: 4px;
		}
	}

	/* High contrast mode support */
	@media (prefers-contrast: high) {
		.result-item {
			border-color: #000;
		}

		.result-item:hover,
		.result-item:focus {
			border-color: #0066cc;
		}

		.metadata {
			border-top-color: #000;
		}
	}

	/* Reduced motion support */
	@media (prefers-reduced-motion: reduce) {
		.result-item,
		.load-more-button,
		.retry-button,
		.open-external,
		.loading-state .spinner {
			transition: none;
			animation: none;
		}
	}
</style>
