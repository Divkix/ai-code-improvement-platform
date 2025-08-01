<!-- ABOUTME: Search input component with debounced input handling and clear functionality -->
<!-- ABOUTME: Provides keyboard shortcuts and loading states for enhanced user experience -->

<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { debounce } from '../utils/debounce';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Search, X, BookOpen, Sparkles, ArrowUpDown, Loader2 } from '@lucide/svelte';

	interface Props {
		value?: string;
		placeholder?: string;
		disabled?: boolean;
		loading?: boolean;
		searchMode?: 'text' | 'vector' | 'hybrid';
		showModeSelector?: boolean;
	}

	let {
		value = $bindable(''),
		placeholder = 'Search code...',
		disabled = false,
		loading = false,
		searchMode = $bindable('text'),
		showModeSelector = true
	}: Props = $props();
	// Remove autofocus prop to improve accessibility

	const dispatch = createEventDispatcher<{
		search: { query: string; mode: 'text' | 'vector' | 'hybrid' };
		clear: void;
		focus: void;
		blur: void;
		modeChange: 'text' | 'vector' | 'hybrid';
	}>();

	// Debounced search function. We trigger a leading-edge call for instant
	// feedback and rely on the debounce to coalesce rapid subsequent input.
	const debouncedSearch = debounce((query: string) => {
		// Trailing edge execution – ensures we also run when typing stops.
		if (query.trim()) {
			dispatch('search', { query: query.trim(), mode: searchMode });
		} else {
			dispatch('clear');
		}
	}, 250);

	function handleInput(event: Event) {
		const target = event.target as HTMLInputElement;
		value = target.value;

		// Immediately dispatch for a snappy UI, then debounce to limit requests
		if (value.trim()) {
			dispatch('search', { query: value.trim(), mode: searchMode });
		} else {
			dispatch('clear');
		}

		// Schedule trailing-edge search update to capture final state after pause
		debouncedSearch(value);
	}

	function handleKeydown(event: KeyboardEvent) {
		if (event.key === 'Enter' && value.trim()) {
			// Cancel debounce and search immediately on Enter
			event.preventDefault();
			dispatch('search', { query: value.trim(), mode: searchMode });
		}

		if (event.key === 'Escape') {
			value = '';
			dispatch('clear');
			(event.target as HTMLInputElement).blur();
		}
	}

	function handleClear() {
		value = '';
		dispatch('clear');
	}

	function handleFocus() {
		dispatch('focus');
	}

	function handleBlur() {
		dispatch('blur');
	}

	function handleModeChange(mode: 'text' | 'vector' | 'hybrid') {
		searchMode = mode;
		dispatch('modeChange', mode);

		// Re-trigger search if there's a query
		if (value.trim()) {
			dispatch('search', { query: value.trim(), mode: searchMode });
		}
	}

	// Get mode description for accessibility
	function getModeDescription(mode: 'text' | 'vector' | 'hybrid') {
		switch (mode) {
			case 'text':
				return 'Text-based keyword search';
			case 'vector':
				return 'Semantic AI-powered search';
			case 'hybrid':
				return 'Combined text and semantic search';
			default:
				return '';
		}
	}
</script>

<div class="w-full max-w-2xl">
	<!-- Search Mode Selector -->
	{#if showModeSelector}
		<div class="mb-3 flex gap-1 rounded-lg bg-muted p-1" role="radiogroup" aria-label="Search mode">
			<Button
				variant={searchMode === 'text' ? 'default' : 'ghost'}
				size="sm"
				onclick={() => handleModeChange('text')}
				aria-pressed={searchMode === 'text'}
				aria-label={getModeDescription('text')}
				{disabled}
				class="flex-1"
			>
				<BookOpen class="mr-2 h-4 w-4" />
				Text
			</Button>
			<Button
				variant={searchMode === 'vector' ? 'default' : 'ghost'}
				size="sm"
				onclick={() => handleModeChange('vector')}
				aria-pressed={searchMode === 'vector'}
				aria-label={getModeDescription('vector')}
				{disabled}
				class="flex-1"
			>
				<Sparkles class="mr-2 h-4 w-4" />
				Semantic
			</Button>
			<Button
				variant={searchMode === 'hybrid' ? 'default' : 'ghost'}
				size="sm"
				onclick={() => handleModeChange('hybrid')}
				aria-pressed={searchMode === 'hybrid'}
				aria-label={getModeDescription('hybrid')}
				{disabled}
				class="flex-1"
			>
				<ArrowUpDown class="mr-2 h-4 w-4" />
				Hybrid
			</Button>
		</div>
	{/if}

	<div class="relative">
		<Input
			type="text"
			bind:value
			oninput={handleInput}
			onkeydown={handleKeydown}
			onfocus={handleFocus}
			onblur={handleBlur}
			{placeholder}
			{disabled}
			autocomplete="off"
			spellcheck="false"
			class="pr-10 pl-10"
		/>

		<!-- Search icon or loading spinner -->
		<div class="absolute top-1/2 left-3 -translate-y-1/2 transform text-muted-foreground">
			{#if loading}
				<Loader2 class="h-4 w-4 animate-spin" aria-label="Searching..." />
			{:else}
				<Search class="h-4 w-4" />
			{/if}
		</div>

		<!-- Clear button -->
		{#if value && !disabled}
			<Button
				variant="ghost"
				size="sm"
				class="absolute top-1/2 right-1 h-8 w-8 -translate-y-1/2 transform p-0"
				onclick={handleClear}
				aria-label="Clear search"
				tabindex={-1}
			>
				<X class="h-4 w-4" />
			</Button>
		{/if}
	</div>
</div>
