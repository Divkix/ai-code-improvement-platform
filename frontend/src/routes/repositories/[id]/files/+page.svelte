<script lang="ts">
	import { page } from '$app/stores';
	import { onMount } from 'svelte';
	import CodeSnippet from '$lib/components/CodeSnippet.svelte';
	import { apiClient } from '$lib/api/client';
	import type { components } from '$lib/api/types';
	import { get } from 'svelte/store';

	let loading = true;
	let error: string | null = null;
	let content = '';
	let language = '';

	const currentPage = get(page);
	const repoId = currentPage.params.id;
	const filePath = currentPage.url.searchParams.get('path') ?? '';
	const startLine = Number(currentPage.url.searchParams.get('line') ?? '1');
	const endLine = Number(currentPage.url.searchParams.get('endLine') ?? startLine);
	const searchTerm = currentPage.url.searchParams.get('q') ?? '';

	onMount(async () => {
		try {
			// fetch the first chunk of this file via search
			const { data, error: apiErr } = await apiClient.POST('/api/search', {
				body: {
					query: filePath,
					repositoryId: repoId,
					limit: 1,
					offset: 0
				}
			});

			if (apiErr) {
				throw new Error(apiErr.message || 'Failed to fetch file');
			}

			if (data && data.results.length > 0) {
				const chunk = data.results[0];
				content = chunk.content;
				language = chunk.language;
			} else {
				error = 'File not found in repository';
			}
		} catch (err) {
			error = err instanceof Error ? err.message : 'Unknown error';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>{filePath}</title>
</svelte:head>

{#if loading}
	<p style="padding:1rem">Loading file...</p>
{:else if error}
	<p style="padding:1rem;color:#dc2626">{error}</p>
{:else}
	<div class="viewer">
		<h2 class="file-path">{filePath}</h2>
		<CodeSnippet {content} {language} {searchTerm} showLineNumbers={true} {startLine} />
	</div>
{/if}

<style>
	.viewer {
		max-width: 900px;
		margin: 32px auto;
		padding: 0 16px;
	}

	.file-path {
		font-size: 18px;
		font-weight: 600;
		color: #1f2937;
		margin-bottom: 16px;
		word-break: break-all;
	}
</style>
