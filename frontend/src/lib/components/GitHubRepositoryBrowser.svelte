<script lang="ts">
	import { onMount } from 'svelte';
	import apiClient, { type GitHubRepository, type User } from '$lib/api';

	let {
		user,
		onRepositoryImport = () => {}
	}: {
		user: User;
		onRepositoryImport?: (repo: GitHubRepository) => void;
	} = $props();

	let repositories = $state<GitHubRepository[]>([]);
	let loading = $state(false);
	let error = $state('');
	let currentPage = $state(1);
	let hasMore = $state(false);
	let importing = $state<Set<number>>(new Set());

	onMount(() => {
		if (user.githubConnected) {
			loadRepositories();
		}
	});

	async function loadRepositories(page: number = 1) {
		if (!user.githubConnected) return;

		try {
			loading = true;
			error = '';

			const response = await apiClient.getGitHubRepositories(page);

			if (page === 1) {
				repositories = response.repositories;
			} else {
				repositories = [...repositories, ...response.repositories];
			}

			currentPage = response.currentPage;
			hasMore = response.hasMore;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load GitHub repositories';
			console.error('Error loading GitHub repositories:', err);
		} finally {
			loading = false;
		}
	}

	async function loadMore() {
		if (!hasMore || loading) return;
		await loadRepositories(currentPage + 1);
	}

	async function importRepository(repo: GitHubRepository) {
		try {
			importing = new Set([...importing, repo.id]);
			error = '';

			// Call the parent's import handler
			onRepositoryImport(repo);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to import repository';
		} finally {
			importing = new Set([...importing].filter((id) => id !== repo.id));
		}
	}

	function formatDate(dateString: string) {
		const date = new Date(dateString);
		const now = new Date();
		const diffInHours = (now.getTime() - date.getTime()) / (1000 * 60 * 60);

		if (diffInHours < 1) return 'Just now';
		if (diffInHours < 24) return `${Math.floor(diffInHours)} hours ago`;

		const diffInDays = Math.floor(diffInHours / 24);
		if (diffInDays < 30) return `${diffInDays} day${diffInDays > 1 ? 's' : ''} ago`;

		const diffInMonths = Math.floor(diffInDays / 30);
		if (diffInMonths < 12) return `${diffInMonths} month${diffInMonths > 1 ? 's' : ''} ago`;

		const diffInYears = Math.floor(diffInMonths / 12);
		return `${diffInYears} year${diffInYears > 1 ? 's' : ''} ago`;
	}

	function formatSize(sizeInKB: number) {
		if (sizeInKB < 1024) return `${sizeInKB} KB`;
		const sizeInMB = Math.round(sizeInKB / 1024);
		if (sizeInMB < 1024) return `${sizeInMB} MB`;
		const sizeInGB = Math.round(sizeInMB / 1024);
		return `${sizeInGB} GB`;
	}
</script>

{#if !user.githubConnected}
	<div class="rounded-lg border border-gray-200 bg-gray-50 p-6 text-center">
		<svg
			class="mx-auto h-12 w-12 text-gray-400"
			fill="none"
			viewBox="0 0 24 24"
			stroke="currentColor"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"
			/>
		</svg>
		<h3 class="mt-2 text-sm font-medium text-gray-900">GitHub Not Connected</h3>
		<p class="mt-1 text-sm text-gray-500">
			Connect your GitHub account to browse and import repositories.
		</p>
	</div>
{:else if loading && repositories.length === 0}
	<div class="py-12 text-center">
		<div
			class="mx-auto h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent"
		></div>
		<p class="mt-2 text-gray-600">Loading your GitHub repositories...</p>
	</div>
{:else if error && repositories.length === 0}
	<div class="rounded-md bg-red-50 p-4">
		<div class="flex">
			<div class="ml-3">
				<h3 class="text-sm font-medium text-red-800">Error</h3>
				<p class="mt-2 text-sm text-red-700">{error}</p>
				<div class="mt-4">
					<button
						onclick={() => loadRepositories(1)}
						class="rounded-md bg-red-100 px-2 py-1 text-sm font-medium text-red-800 hover:bg-red-200"
					>
						Try Again
					</button>
				</div>
			</div>
		</div>
	</div>
{:else if repositories.length === 0}
	<div class="py-12 text-center">
		<svg
			class="mx-auto h-12 w-12 text-gray-400"
			fill="none"
			viewBox="0 0 24 24"
			stroke="currentColor"
		>
			<path
				stroke-linecap="round"
				stroke-linejoin="round"
				stroke-width="2"
				d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"
			/>
		</svg>
		<h3 class="mt-2 text-sm font-medium text-gray-900">No repositories found</h3>
		<p class="mt-1 text-sm text-gray-500">
			You don't have any repositories in your GitHub account yet.
		</p>
	</div>
{:else}
	<div class="space-y-4">
		{#if error}
			<div class="rounded-md bg-red-50 p-4">
				<div class="flex">
					<div class="ml-3">
						<h3 class="text-sm font-medium text-red-800">Error</h3>
						<p class="mt-2 text-sm text-red-700">{error}</p>
					</div>
				</div>
			</div>
		{/if}

		<div class="grid gap-4 sm:grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
			{#each repositories as repo (repo.id)}
				<div
					class="overflow-hidden rounded-lg border border-gray-200 bg-white transition-shadow hover:shadow-md"
				>
					<div class="p-4">
						<div class="flex items-start justify-between">
							<div class="min-w-0 flex-1">
								<div class="flex items-center space-x-2">
									{#if repo.private}
										<svg class="h-4 w-4 text-yellow-500" fill="currentColor" viewBox="0 0 20 20">
											<path
												fill-rule="evenodd"
												d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z"
												clip-rule="evenodd"
											/>
										</svg>
									{:else}
										<svg class="h-4 w-4 text-green-500" fill="currentColor" viewBox="0 0 20 20">
											<path
												fill-rule="evenodd"
												d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z"
												clip-rule="evenodd"
											/>
										</svg>
									{/if}
									<h3 class="truncate text-sm font-medium text-gray-900">{repo.name}</h3>
								</div>
								<p class="mt-1 text-xs text-gray-500">
									{repo.owner}/{repo.name}
								</p>
							</div>
						</div>

						{#if repo.description}
							<p class="mt-2 line-clamp-2 text-sm text-gray-600">{repo.description}</p>
						{/if}

						<div class="mt-3 flex items-center justify-between text-xs text-gray-500">
							<div class="flex items-center space-x-4">
								{#if repo.language}
									<div class="flex items-center">
										<div class="mr-1 h-2 w-2 rounded-full bg-blue-400"></div>
										{repo.language}
									</div>
								{/if}
								<div class="flex items-center">
									<svg class="mr-1 h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M11.049 2.927c.3-.921 1.603-.921 1.902 0l1.519 4.674a1 1 0 00.95.69h4.915c.969 0 1.371 1.24.588 1.81l-3.976 2.888a1 1 0 00-.363 1.118l1.518 4.674c.3.922-.755 1.688-1.538 1.118l-3.976-2.888a1 1 0 00-1.176 0l-3.976 2.888c-.783.57-1.838-.197-1.538-1.118l1.518-4.674a1 1 0 00-.363-1.118l-3.976-2.888c-.784-.57-.38-1.81.588-1.81h4.914a1 1 0 00.951-.69l1.519-4.674z"
										/>
									</svg>
									{repo.stargazersCount}
								</div>
								<div class="flex items-center">
									<svg class="mr-1 h-3 w-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.367 2.684 3 3 0 00-5.367-2.684z"
										/>
									</svg>
									{repo.forksCount}
								</div>
								<div>{formatSize(repo.size)}</div>
							</div>
						</div>

						<div class="mt-3 text-xs text-gray-500">
							Updated {formatDate(repo.updatedAt)}
						</div>

						<div class="mt-4">
							<button
								onclick={() => importRepository(repo)}
								disabled={importing.has(repo.id)}
								class="w-full rounded-md bg-blue-600 px-3 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
							>
								{#if importing.has(repo.id)}
									<div class="flex items-center justify-center">
										<svg class="mr-2 h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
											<circle
												class="opacity-25"
												cx="12"
												cy="12"
												r="10"
												stroke="currentColor"
												stroke-width="4"
											></circle>
											<path
												class="opacity-75"
												fill="currentColor"
												d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
											></path>
										</svg>
										Importing...
									</div>
								{:else}
									Import Repository
								{/if}
							</button>
						</div>
					</div>
				</div>
			{/each}
		</div>

		{#if hasMore}
			<div class="mt-6 text-center">
				<button
					onclick={loadMore}
					disabled={loading}
					class="inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50 disabled:cursor-not-allowed disabled:opacity-50"
				>
					{#if loading}
						<svg class="mr-2 h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
							<circle
								class="opacity-25"
								cx="12"
								cy="12"
								r="10"
								stroke="currentColor"
								stroke-width="4"
							></circle>
							<path
								class="opacity-75"
								fill="currentColor"
								d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
							></path>
						</svg>
						Loading more...
					{:else}
						Load More Repositories
					{/if}
				</button>
			</div>
		{/if}
	</div>
{/if}

<style>
	.line-clamp-2 {
		display: -webkit-box;
		-webkit-line-clamp: 2;
		line-clamp: 2;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}
</style>
