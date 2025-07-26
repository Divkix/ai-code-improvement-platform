<!-- ABOUTME: Search results display component with pagination and result highlighting -->
<!-- ABOUTME: Handles loading states, error states, and empty results with helpful suggestions -->

<script lang="ts">
	import type { SearchResponse, SearchResult } from '../api/search-types';
	import CodeSnippet from './CodeSnippet.svelte';
	import { createEventDispatcher } from 'svelte';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { ScrollArea } from '$lib/components/ui/scroll-area/index.js';
	import * as Alert from '$lib/components/ui/alert/index.js';
	import {
		FileText,
		ExternalLink,
		Search,
		AlertCircle,
		Loader2,
		Sparkles,
		ArrowUpDown
	} from '@lucide/svelte';

	interface Props {
		results?: SearchResponse | null;
		loading?: boolean;
		error?: string | null;
		query?: string;
		searchMode?: 'text' | 'vector' | 'hybrid';
	}

	let {
		results = null,
		loading = false,
		error = null,
		query = '',
		searchMode = 'text'
	}: Props = $props();

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

<div class="mx-auto w-full max-w-4xl">
	{#if loading}
		<Card.Root>
			<Card.Content class="flex items-center justify-center py-8">
				<Loader2 class="mr-3 h-8 w-8 animate-spin" />
				<p class="text-muted-foreground">Searching code...</p>
			</Card.Content>
		</Card.Root>
	{:else if error}
		<Alert.Root variant="destructive">
			<AlertCircle class="h-4 w-4" />
			<Alert.Title>Search Error</Alert.Title>
			<Alert.Description class="mt-2">
				{error}
				<Button variant="outline" size="sm" class="mt-3" onclick={handleRetry}>Try Again</Button>
			</Alert.Description>
		</Alert.Root>
	{:else if results && Array.isArray(results.results) && results.results.length > 0}
		<div class="mb-6">
			<h3 class="mb-2 text-lg font-semibold text-foreground">Search Results</h3>
			<p class="text-sm text-muted-foreground">
				{results.total} result{results.total === 1 ? '' : 's'} for "{query}"
			</p>
		</div>

		<ScrollArea class="h-[600px] w-full">
			<div class="space-y-4 pr-4">
				{#each results.results as result, i (`${result.id}-${result.startLine}-${i}`)}
					<Card.Root
						class="cursor-pointer transition-colors hover:bg-muted/50"
						onclick={() => handleResultClick(result)}
						onkeydown={(e) => e.key === 'Enter' && handleResultClick(result)}
						role="button"
						tabindex={0}
						aria-label="View code chunk in {result.fileName}"
					>
						<Card.Header class="pb-3">
							<div class="flex items-center justify-between">
								<div class="flex min-w-0 flex-1 items-center space-x-2">
									<FileText class="h-4 w-4 flex-shrink-0 text-muted-foreground" />
									<span class="truncate text-sm font-medium" title={result.filePath}>
										{result.filePath}
									</span>
								</div>
								<Button
									variant="ghost"
									size="sm"
									class="h-8 w-8 flex-shrink-0 p-0"
									title="Open in new tab"
									aria-label="Open code chunk in new tab"
									onclick={(e) => {
										e.stopPropagation();
										handleResultClick(result);
									}}
								>
									<ExternalLink class="h-4 w-4" />
								</Button>
							</div>
							<div class="mt-2 flex items-center gap-2">
								<Badge
									variant="outline"
									class="text-xs"
									style="background-color: {getLanguageColor(result.language)}; color: white;"
								>
									{result.language}
								</Badge>
								<Badge variant="secondary" class="text-xs">
									{formatLineRange(result.startLine, result.endLine)}
								</Badge>
								<Badge
									variant="outline"
									class="gap-1 text-xs"
									style="color: {getRelevanceColor(getRelevanceLevel(result.score, searchMode))}"
									title="{getScoreLabel(searchMode)}: {formatScore(result.score, searchMode)}"
								>
									{#if searchMode === 'vector'}
										<Sparkles class="h-3 w-3" />
									{:else if searchMode === 'hybrid'}
										<ArrowUpDown class="h-3 w-3" />
									{/if}
									{getScoreLabel(searchMode)}: {formatScore(result.score, searchMode)}
								</Badge>
							</div>
						</Card.Header>

						<Card.Content class="pt-0">
							<div class="mb-4">
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
								<div class="mb-2">
									<span class="mr-2 text-xs font-medium text-muted-foreground">Functions:</span>
									<div class="inline-flex flex-wrap gap-1">
										{#each result.metadata.functions.slice(0, 3) as func (func)}
											<Badge variant="secondary" class="text-xs">
												{func}
											</Badge>
										{/each}
										{#if result.metadata.functions.length > 3}
											<Badge variant="outline" class="text-xs">
												+{result.metadata.functions.length - 3} more
											</Badge>
										{/if}
									</div>
								</div>
							{/if}

							{#if result.metadata?.classes && result.metadata.classes.length > 0}
								<div class="mb-2">
									<span class="mr-2 text-xs font-medium text-muted-foreground">Classes:</span>
									<div class="inline-flex flex-wrap gap-1">
										{#each result.metadata.classes.slice(0, 3) as cls (cls)}
											<Badge variant="secondary" class="text-xs">
												{cls}
											</Badge>
										{/each}
										{#if result.metadata.classes.length > 3}
											<Badge variant="outline" class="text-xs">
												+{result.metadata.classes.length - 3} more
											</Badge>
										{/if}
									</div>
								</div>
							{/if}
						</Card.Content>
					</Card.Root>
				{/each}
			</div>
		</ScrollArea>

		{#if results.hasMore}
			<div class="mt-6 text-center">
				<Button variant="outline" onclick={handleLoadMore} disabled={loading}>
					{#if loading}
						<Loader2 class="mr-2 h-4 w-4 animate-spin" />
						Loading...
					{:else}
						Load More Results
					{/if}
				</Button>
			</div>
		{/if}
	{:else if query}
		<Card.Root>
			<Card.Content class="py-12 text-center">
				<Search class="mx-auto mb-4 h-16 w-16 text-muted-foreground" />
				<h3 class="mb-2 text-lg font-semibold text-foreground">No Results Found</h3>
				<p class="mb-6 text-muted-foreground">No code chunks found for "{query}"</p>
				<div class="mx-auto max-w-md text-left">
					<h4 class="mb-3 text-sm font-medium text-foreground">Try:</h4>
					<ul class="list-inside list-disc space-y-1 text-sm text-muted-foreground">
						<li>Different keywords or phrases</li>
						<li>Broader search terms</li>
						<li>Checking your spelling</li>
						<li>Searching for function or class names</li>
						<li>Using specific programming constructs</li>
					</ul>
				</div>
			</Card.Content>
		</Card.Root>
	{/if}
</div>
