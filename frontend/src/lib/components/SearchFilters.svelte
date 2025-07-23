<!-- ABOUTME: Search filters component for language, file type, and repository selection -->
<!-- ABOUTME: Provides dropdown filters with clear options and responsive design -->

<script lang="ts">
	import { createEventDispatcher } from 'svelte';

	export let selectedLanguage: string = '';
	export let selectedFileType: string = '';
	export let selectedRepository: string = '';
	export let languages: string[] = [];
	export let repositories: Array<{ id: string; name: string; fullName: string }> = [];
	export let disabled: boolean = false;

	const dispatch = createEventDispatcher<{
		languageChange: string;
		fileTypeChange: string;
		repositoryChange: string;
		clearFilters: void;
	}>();

	// Common file types for quick selection
	const commonFileTypes = [
		{ value: 'js', label: 'JavaScript (.js)' },
		{ value: 'ts', label: 'TypeScript (.ts)' },
		{ value: 'py', label: 'Python (.py)' },
		{ value: 'go', label: 'Go (.go)' },
		{ value: 'java', label: 'Java (.java)' },
		{ value: 'cpp', label: 'C++ (.cpp)' },
		{ value: 'c', label: 'C (.c)' },
		{ value: 'php', label: 'PHP (.php)' },
		{ value: 'rb', label: 'Ruby (.rb)' },
		{ value: 'rs', label: 'Rust (.rs)' },
		{ value: 'html', label: 'HTML (.html)' },
		{ value: 'css', label: 'CSS (.css)' },
		{ value: 'json', label: 'JSON (.json)' },
		{ value: 'yaml', label: 'YAML (.yaml)' },
		{ value: 'yml', label: 'YAML (.yml)' },
		{ value: 'md', label: 'Markdown (.md)' },
		{ value: 'sh', label: 'Shell (.sh)' },
		{ value: 'sql', label: 'SQL (.sql)' }
	];

	function handleLanguageChange(event: Event) {
		const target = event.target as HTMLSelectElement;
		selectedLanguage = target.value;
		dispatch('languageChange', selectedLanguage);
	}

	function handleFileTypeChange(event: Event) {
		const target = event.target as HTMLSelectElement;
		selectedFileType = target.value;
		dispatch('fileTypeChange', selectedFileType);
	}

	function handleRepositoryChange(event: Event) {
		const target = event.target as HTMLSelectElement;
		selectedRepository = target.value;
		dispatch('repositoryChange', selectedRepository);
	}

	function clearAllFilters() {
		selectedLanguage = '';
		selectedFileType = '';
		selectedRepository = '';
		dispatch('clearFilters');
	}

	$: hasActiveFilters = selectedLanguage || selectedFileType || selectedRepository;
</script>

<div class="search-filters">
	<div class="filters-header">
		<h3>Filters</h3>
		{#if hasActiveFilters}
			<button
				class="clear-filters-button"
				on:click={clearAllFilters}
				{disabled}
				title="Clear all filters"
			>
				Clear All
			</button>
		{/if}
	</div>

	<div class="filters-grid">
		<!-- Language Filter -->
		<div class="filter-group">
			<label for="language-filter" class="filter-label"> Programming Language </label>
			<div class="select-wrapper">
				<select
					id="language-filter"
					bind:value={selectedLanguage}
					on:change={handleLanguageChange}
					{disabled}
					class="filter-select"
				>
					<option value="">All Languages</option>
					{#each languages as language}
						<option value={language}>
							{language.charAt(0).toUpperCase() + language.slice(1)}
						</option>
					{/each}
				</select>
				<svg
					class="select-arrow"
					width="12"
					height="12"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
				>
					<path d="M6 9l6 6 6-6" />
				</svg>
			</div>
		</div>

		<!-- File Type Filter -->
		<div class="filter-group">
			<label for="filetype-filter" class="filter-label"> File Type </label>
			<div class="select-wrapper">
				<select
					id="filetype-filter"
					bind:value={selectedFileType}
					on:change={handleFileTypeChange}
					{disabled}
					class="filter-select"
				>
					<option value="">All File Types</option>
					{#each commonFileTypes as fileType}
						<option value={fileType.value}>
							{fileType.label}
						</option>
					{/each}
				</select>
				<svg
					class="select-arrow"
					width="12"
					height="12"
					viewBox="0 0 24 24"
					fill="none"
					stroke="currentColor"
					stroke-width="2"
				>
					<path d="M6 9l6 6 6-6" />
				</svg>
			</div>
		</div>

		<!-- Repository Filter -->
		{#if repositories.length > 0}
			<div class="filter-group">
				<label for="repository-filter" class="filter-label"> Repository </label>
				<div class="select-wrapper">
					<select
						id="repository-filter"
						bind:value={selectedRepository}
						on:change={handleRepositoryChange}
						{disabled}
						class="filter-select"
					>
						<option value="">All Repositories</option>
						{#each repositories as repo}
							<option value={repo.id}>
								{repo.fullName}
							</option>
						{/each}
					</select>
					<svg
						class="select-arrow"
						width="12"
						height="12"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
					>
						<path d="M6 9l6 6 6-6" />
					</svg>
				</div>
			</div>
		{/if}
	</div>

	<!-- Active Filters Display -->
	{#if hasActiveFilters}
		<div class="active-filters">
			<span class="active-filters-label">Active filters:</span>
			<div class="active-filter-tags">
				{#if selectedLanguage}
					<span class="filter-tag">
						<span class="filter-tag-label">Language:</span>
						<span class="filter-tag-value">{selectedLanguage}</span>
						<button
							class="remove-filter"
							on:click={() => {
								selectedLanguage = '';
								dispatch('languageChange', '');
							}}
							aria-label="Remove language filter"
						>
							<svg
								width="12"
								height="12"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
							>
								<line x1="18" y1="6" x2="6" y2="18" />
								<line x1="6" y1="6" x2="18" y2="18" />
							</svg>
						</button>
					</span>
				{/if}

				{#if selectedFileType}
					<span class="filter-tag">
						<span class="filter-tag-label">Type:</span>
						<span class="filter-tag-value">.{selectedFileType}</span>
						<button
							class="remove-filter"
							on:click={() => {
								selectedFileType = '';
								dispatch('fileTypeChange', '');
							}}
							aria-label="Remove file type filter"
						>
							<svg
								width="12"
								height="12"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
							>
								<line x1="18" y1="6" x2="6" y2="18" />
								<line x1="6" y1="6" x2="18" y2="18" />
							</svg>
						</button>
					</span>
				{/if}

				{#if selectedRepository}
					{@const repo = repositories.find((r) => r.id === selectedRepository)}
					<span class="filter-tag">
						<span class="filter-tag-label">Repo:</span>
						<span class="filter-tag-value">{repo?.name || selectedRepository}</span>
						<button
							class="remove-filter"
							on:click={() => {
								selectedRepository = '';
								dispatch('repositoryChange', '');
							}}
							aria-label="Remove repository filter"
						>
							<svg
								width="12"
								height="12"
								viewBox="0 0 24 24"
								fill="none"
								stroke="currentColor"
								stroke-width="2"
							>
								<line x1="18" y1="6" x2="6" y2="18" />
								<line x1="6" y1="6" x2="18" y2="18" />
							</svg>
						</button>
					</span>
				{/if}
			</div>
		</div>
	{/if}
</div>

<style>
	.search-filters {
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 8px;
		padding: 20px;
		margin-bottom: 24px;
	}

	.filters-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 16px;
	}

	.filters-header h3 {
		margin: 0;
		font-size: 16px;
		font-weight: 600;
		color: #1f2937;
	}

	.clear-filters-button {
		background: none;
		border: 1px solid #d1d5db;
		color: #6b7280;
		padding: 4px 8px;
		border-radius: 4px;
		font-size: 12px;
		cursor: pointer;
		transition: all 0.2s;
	}

	.clear-filters-button:hover:not(:disabled) {
		border-color: #9ca3af;
		color: #374151;
	}

	.clear-filters-button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.filters-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 16px;
		margin-bottom: 16px;
	}

	.filter-group {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.filter-label {
		font-size: 13px;
		font-weight: 500;
		color: #374151;
		margin-bottom: 4px;
	}

	.select-wrapper {
		position: relative;
		display: flex;
		align-items: center;
	}

	.filter-select {
		width: 100%;
		padding: 8px 32px 8px 12px;
		border: 1px solid #d1d5db;
		border-radius: 6px;
		background: white;
		font-size: 14px;
		color: #1f2937;
		cursor: pointer;
		transition:
			border-color 0.2s,
			box-shadow 0.2s;
		appearance: none;
		-webkit-appearance: none;
		-moz-appearance: none;
	}

	.filter-select:focus {
		outline: none;
		border-color: #3b82f6;
		box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
	}

	.filter-select:disabled {
		background-color: #f9fafb;
		color: #6b7280;
		cursor: not-allowed;
	}

	.select-arrow {
		position: absolute;
		right: 10px;
		top: 50%;
		transform: translateY(-50%);
		color: #6b7280;
		pointer-events: none;
	}

	.active-filters {
		border-top: 1px solid #e5e7eb;
		padding-top: 16px;
		margin-top: 16px;
	}

	.active-filters-label {
		font-size: 13px;
		font-weight: 500;
		color: #6b7280;
		margin-bottom: 8px;
		display: block;
	}

	.active-filter-tags {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}

	.filter-tag {
		display: inline-flex;
		align-items: center;
		gap: 4px;
		background-color: #eff6ff;
		border: 1px solid #bfdbfe;
		border-radius: 16px;
		padding: 4px 8px;
		font-size: 12px;
		max-width: 200px;
	}

	.filter-tag-label {
		color: #1e40af;
		font-weight: 500;
	}

	.filter-tag-value {
		color: #1d4ed8;
		font-weight: 600;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}

	.remove-filter {
		background: none;
		border: none;
		color: #6b7280;
		cursor: pointer;
		padding: 2px;
		border-radius: 2px;
		transition:
			color 0.2s,
			background-color 0.2s;
		display: flex;
		align-items: center;
		justify-content: center;
		flex-shrink: 0;
	}

	.remove-filter:hover {
		color: #dc2626;
		background-color: rgba(220, 38, 38, 0.1);
	}

	.remove-filter:focus {
		outline: 2px solid #3b82f6;
		outline-offset: 1px;
	}

	/* Mobile responsiveness */
	@media (max-width: 640px) {
		.search-filters {
			padding: 16px;
		}

		.filters-grid {
			grid-template-columns: 1fr;
			gap: 12px;
		}

		.filters-header {
			flex-direction: column;
			align-items: flex-start;
			gap: 8px;
		}

		.active-filter-tags {
			flex-direction: column;
			align-items: flex-start;
		}

		.filter-tag {
			max-width: 100%;
		}
	}

	/* High contrast mode support */
	@media (prefers-contrast: high) {
		.search-filters {
			border-color: #000;
		}

		.filter-select {
			border-color: #000;
		}

		.filter-select:focus {
			border-color: #0066cc;
			box-shadow: 0 0 0 3px rgba(0, 102, 204, 0.3);
		}

		.filter-tag {
			background-color: #e0f2fe;
			border-color: #000;
		}

		.active-filters {
			border-top-color: #000;
		}
	}

	/* Reduced motion support */
	@media (prefers-reduced-motion: reduce) {
		.filter-select,
		.clear-filters-button,
		.remove-filter {
			transition: none;
		}
	}
</style>
