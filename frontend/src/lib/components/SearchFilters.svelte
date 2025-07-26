<!-- ABOUTME: Search filters component for language, file type, and repository selection -->
<!-- ABOUTME: Provides dropdown filters with clear options and responsive design -->

<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import * as Select from '$lib/components/ui/select/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { X } from '@lucide/svelte';

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

	function handleLanguageChange(value?: string) {
		selectedLanguage = value || '';
		dispatch('languageChange', selectedLanguage);
	}

	function handleFileTypeChange(value?: string) {
		selectedFileType = value || '';
		dispatch('fileTypeChange', selectedFileType);
	}

	function handleRepositoryChange(value?: string) {
		selectedRepository = value || '';
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

<div class="mb-6 rounded-lg border bg-card p-6">
	<div class="mb-4 flex items-center justify-between">
		<h3 class="text-base font-semibold">Filters</h3>
		{#if hasActiveFilters}
			<Button variant="outline" size="sm" onclick={clearAllFilters} {disabled} class="text-xs">
				Clear All
			</Button>
		{/if}
	</div>

	<div class="mb-4 grid grid-cols-1 gap-4 md:grid-cols-3">
		<!-- Language Filter -->
		<div class="space-y-2">
			<label class="text-sm font-medium text-foreground"> Programming Language </label>
			<Select.Root bind:value={selectedLanguage} onValueChange={handleLanguageChange} {disabled}>
				<Select.Trigger class="w-full">
					{selectedLanguage
						? selectedLanguage.charAt(0).toUpperCase() + selectedLanguage.slice(1)
						: 'All Languages'}
				</Select.Trigger>
				<Select.Content>
					<Select.Item value="">All Languages</Select.Item>
					{#each languages as language (language)}
						<Select.Item value={language}>
							{language.charAt(0).toUpperCase() + language.slice(1)}
						</Select.Item>
					{/each}
				</Select.Content>
			</Select.Root>
		</div>

		<!-- File Type Filter -->
		<div class="space-y-2">
			<label class="text-sm font-medium text-foreground"> File Type </label>
			<Select.Root bind:value={selectedFileType} onValueChange={handleFileTypeChange} {disabled}>
				<Select.Trigger class="w-full">
					{selectedFileType
						? commonFileTypes.find((ft) => ft.value === selectedFileType)?.label || selectedFileType
						: 'All File Types'}
				</Select.Trigger>
				<Select.Content>
					<Select.Item value="">All File Types</Select.Item>
					{#each commonFileTypes as fileType (fileType.value)}
						<Select.Item value={fileType.value}>
							{fileType.label}
						</Select.Item>
					{/each}
				</Select.Content>
			</Select.Root>
		</div>

		<!-- Repository Filter -->
		{#if repositories.length > 0}
			<div class="space-y-2">
				<label class="text-sm font-medium text-foreground"> Repository </label>
				<Select.Root
					bind:value={selectedRepository}
					onValueChange={handleRepositoryChange}
					{disabled}
				>
					<Select.Trigger class="w-full">
						{selectedRepository
							? repositories.find((r) => r.id === selectedRepository)?.fullName ||
								selectedRepository
							: 'All Repositories'}
					</Select.Trigger>
					<Select.Content>
						<Select.Item value="">All Repositories</Select.Item>
						{#each repositories as repo (repo.id)}
							<Select.Item value={repo.id}>
								{repo.fullName}
							</Select.Item>
						{/each}
					</Select.Content>
				</Select.Root>
			</div>
		{/if}
	</div>

	<!-- Active Filters Display -->
	{#if hasActiveFilters}
		<div class="border-t pt-4">
			<span class="mb-2 block text-sm font-medium text-muted-foreground">Active filters:</span>
			<div class="flex flex-wrap gap-2">
				{#if selectedLanguage}
					<Badge variant="secondary" class="gap-1">
						<span class="text-xs">Language: {selectedLanguage}</span>
						<Button
							variant="ghost"
							size="sm"
							class="hover:text-destructive-foreground h-4 w-4 p-0 hover:bg-destructive"
							onclick={() => {
								selectedLanguage = '';
								dispatch('languageChange', '');
							}}
							aria-label="Remove language filter"
						>
							<X class="h-3 w-3" />
						</Button>
					</Badge>
				{/if}

				{#if selectedFileType}
					<Badge variant="secondary" class="gap-1">
						<span class="text-xs">Type: .{selectedFileType}</span>
						<Button
							variant="ghost"
							size="sm"
							class="hover:text-destructive-foreground h-4 w-4 p-0 hover:bg-destructive"
							onclick={() => {
								selectedFileType = '';
								dispatch('fileTypeChange', '');
							}}
							aria-label="Remove file type filter"
						>
							<X class="h-3 w-3" />
						</Button>
					</Badge>
				{/if}

				{#if selectedRepository}
					{@const repo = repositories.find((r) => r.id === selectedRepository)}
					<Badge variant="secondary" class="gap-1">
						<span class="text-xs">Repo: {repo?.name || selectedRepository}</span>
						<Button
							variant="ghost"
							size="sm"
							class="hover:text-destructive-foreground h-4 w-4 p-0 hover:bg-destructive"
							onclick={() => {
								selectedRepository = '';
								dispatch('repositoryChange', '');
							}}
							aria-label="Remove repository filter"
						>
							<X class="h-3 w-3" />
						</Button>
					</Badge>
				{/if}
			</div>
		</div>
	{/if}
</div>
