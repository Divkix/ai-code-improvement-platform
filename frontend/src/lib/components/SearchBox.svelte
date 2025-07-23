<!-- ABOUTME: Search input component with debounced input handling and clear functionality -->
<!-- ABOUTME: Provides keyboard shortcuts and loading states for enhanced user experience -->

<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { debounce } from '../utils/debounce';

	export let value = '';
	export let placeholder = 'Search code...';
	export let disabled = false;
	export let loading = false;
	export let searchMode: 'text' | 'vector' | 'hybrid' = 'text';
	export let showModeSelector = true;
	// Remove autofocus prop to improve accessibility

	const dispatch = createEventDispatcher<{
		search: { query: string; mode: 'text' | 'vector' | 'hybrid' };
		clear: void;
		focus: void;
		blur: void;
		modeChange: 'text' | 'vector' | 'hybrid';
	}>();

	// Debounced search function to avoid too many API calls
	const debouncedSearch = debounce((query: string) => {
		if (query.trim()) {
			dispatch('search', { query: query.trim(), mode: searchMode });
		} else {
			dispatch('clear');
		}
	}, 300);

	function handleInput(event: Event) {
		const target = event.target as HTMLInputElement;
		value = target.value;

		// Always trigger the debounced search
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

<div class="search-box">
	<!-- Search Mode Selector -->
	{#if showModeSelector}
		<div class="mode-selector" role="radiogroup" aria-label="Search mode">
			<button
				type="button"
				class="mode-button"
				class:active={searchMode === 'text'}
				aria-pressed={searchMode === 'text'}
				aria-label={getModeDescription('text')}
				on:click={() => handleModeChange('text')}
				{disabled}
			>
				<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="m19 21-7-4-7 4V5a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2v16z"/>
				</svg>
				Text
			</button>
			<button
				type="button"
				class="mode-button"
				class:active={searchMode === 'vector'}
				aria-pressed={searchMode === 'vector'}
				aria-label={getModeDescription('vector')}
				on:click={() => handleModeChange('vector')}
				{disabled}
			>
				<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M9 19c-5 0-8-2.5-8-5s3-5 8-5 8 2.5 8 5-3 5-8 5Z"/>
					<path d="m8 19 8-14"/>
					<path d="m1 14 8-14"/>
					<path d="m15 5 4 14"/>
				</svg>
				Semantic
			</button>
			<button
				type="button"
				class="mode-button"
				class:active={searchMode === 'hybrid'}
				aria-pressed={searchMode === 'hybrid'}
				aria-label={getModeDescription('hybrid')}
				on:click={() => handleModeChange('hybrid')}
				{disabled}
			>
				<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<path d="M12 3v18m0-18 4 4m-4-4-4 4"/>
					<path d="m8 17 4 4 4-4"/>
				</svg>
				Hybrid
			</button>
		</div>
	{/if}

	<div class="search-container">
		<input
			type="text"
			bind:value
			on:input={handleInput}
			on:keydown={handleKeydown}
			on:focus={handleFocus}
			on:blur={handleBlur}
			{placeholder}
			{disabled}
			class="search-input"
			class:loading
			autocomplete="off"
			spellcheck="false"
		/>

		<!-- Search icon or loading spinner -->
		<div class="search-icon">
			{#if loading}
				<div class="spinner" aria-label="Searching..."></div>
			{:else}
				<svg
					width="20"
					height="20"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
				>
					<circle cx="11" cy="11" r="8" />
					<path d="M21 21l-4.35-4.35" />
				</svg>
			{/if}
		</div>

		<!-- Clear button -->
		{#if value && !disabled}
			<button
				type="button"
				class="clear-button"
				on:click={handleClear}
				aria-label="Clear search"
				tabindex="-1"
			>
				<svg
					width="16"
					height="16"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
				>
					<line x1="18" y1="6" x2="6" y2="18" />
					<line x1="6" y1="6" x2="18" y2="18" />
				</svg>
			</button>
		{/if}
	</div>
</div>

<style>
	.search-box {
		width: 100%;
		max-width: 600px;
	}

	.mode-selector {
		display: flex;
		gap: 2px;
		background: #f3f4f6;
		border-radius: 8px;
		padding: 4px;
		margin-bottom: 12px;
		border: 1px solid #e5e7eb;
	}

	.mode-button {
		flex: 1;
		display: flex;
		align-items: center;
		justify-content: center;
		gap: 6px;
		padding: 8px 12px;
		border: none;
		background: transparent;
		border-radius: 6px;
		font-size: 14px;
		font-weight: 500;
		color: #6b7280;
		cursor: pointer;
		transition: all 0.2s;
	}

	.mode-button:hover {
		background: rgba(255, 255, 255, 0.8);
		color: #374151;
	}

	.mode-button.active {
		background: white;
		color: #1f2937;
		box-shadow: 0 1px 2px rgba(0, 0, 0, 0.05);
	}

	.mode-button:disabled {
		cursor: not-allowed;
		opacity: 0.5;
	}

	.mode-button svg {
		width: 16px;
		height: 16px;
	}

	.search-container {
		position: relative;
		display: flex;
		align-items: center;
	}

	.search-input {
		width: 100%;
		padding: 12px 48px 12px 48px;
		border: 2px solid #e5e7eb;
		border-radius: 8px;
		font-size: 16px;
		font-family: inherit;
		background: white;
		transition:
			border-color 0.2s,
			box-shadow 0.2s;
		outline: none;
	}

	.search-input:focus {
		border-color: #3b82f6;
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
	}

	.search-input:disabled {
		background-color: #f9fafb;
		color: #6b7280;
		cursor: not-allowed;
	}

	.search-input.loading {
		padding-right: 56px;
	}

	.search-icon {
		position: absolute;
		left: 14px;
		top: 50%;
		transform: translateY(-50%);
		color: #6b7280;
		pointer-events: none;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.clear-button {
		position: absolute;
		right: 12px;
		top: 50%;
		transform: translateY(-50%);
		background: none;
		border: none;
		color: #6b7280;
		cursor: pointer;
		padding: 6px;
		border-radius: 4px;
		transition:
			color 0.2s,
			background-color 0.2s;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.clear-button:hover {
		color: #374151;
		background-color: #f3f4f6;
	}

	.clear-button:focus {
		outline: 2px solid #3b82f6;
		outline-offset: 2px;
	}

	.spinner {
		width: 20px;
		height: 20px;
		border: 2px solid #e5e7eb;
		border-top: 2px solid #3b82f6;
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	@keyframes spin {
		0% {
			transform: rotate(0deg);
		}
		100% {
			transform: rotate(360deg);
		}
	}

	/* Responsive adjustments */
	@media (max-width: 640px) {
		.search-input {
			font-size: 16px; /* Prevent zoom on iOS */
			padding: 10px 40px 10px 40px;
		}

		.search-input.loading {
			padding-right: 48px;
		}

		.search-icon {
			left: 12px;
		}

		.clear-button {
			right: 10px;
		}
	}

	/* High contrast mode support */
	@media (prefers-contrast: high) {
		.search-input {
			border-color: #000;
		}

		.search-input:focus {
			border-color: #0066cc;
			box-shadow: 0 0 0 3px rgba(0, 102, 204, 0.3);
		}
	}

	/* Reduced motion support */
	@media (prefers-reduced-motion: reduce) {
		.search-input,
		.clear-button,
		.spinner {
			transition: none;
			animation: none;
		}
	}
</style>
