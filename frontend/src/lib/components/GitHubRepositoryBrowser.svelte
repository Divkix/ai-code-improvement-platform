<script lang="ts">
	import { onMount } from 'svelte';
	import { getGitHubRepositories } from '$lib/api/hooks';
	import { type GitHubRepository, type User } from '$lib/api';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Alert from '$lib/components/ui/alert/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import {
		Github,
		Lock,
		Unlock,
		Star,
		GitFork,
		Loader2,
		AlertCircle,
		Package,
		Calendar
	} from '@lucide/svelte';

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

			const response = await getGitHubRepositories(page);

			if (page === 1) {
				repositories = response.repositories.map((repo) => ({
					...repo,
					stargazersCount: repo.stargazersCount ?? 0,
					forksCount: repo.forksCount ?? 0,
					size: repo.size ?? 0
				}));
			} else {
				repositories = [
					...repositories,
					...response.repositories.map((repo) => ({
						...repo,
						stargazersCount: repo.stargazersCount ?? 0,
						forksCount: repo.forksCount ?? 0,
						size: repo.size ?? 0
					}))
				];
			}

			currentPage = page;
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
	<Card.Root>
		<Card.Content class="p-6 text-center">
			<Github class="mx-auto h-12 w-12 text-muted-foreground" />
			<Card.Title class="mt-2 text-sm">GitHub Not Connected</Card.Title>
			<Card.Description class="mt-1">
				Connect your GitHub account to browse and import repositories.
			</Card.Description>
		</Card.Content>
	</Card.Root>
{:else if loading && repositories.length === 0}
	<Card.Root>
		<Card.Content class="py-12 text-center">
			<Loader2 class="mx-auto h-8 w-8 animate-spin" />
			<p class="mt-2 text-muted-foreground">Loading your GitHub repositories...</p>
		</Card.Content>
	</Card.Root>
{:else if error && repositories.length === 0}
	<Alert.Root variant="destructive">
		<AlertCircle class="h-4 w-4" />
		<Alert.Title>Error</Alert.Title>
		<Alert.Description>
			<p>{error}</p>
			<Button variant="outline" size="sm" onclick={() => loadRepositories(1)} class="mt-2">
				Try Again
			</Button>
		</Alert.Description>
	</Alert.Root>
{:else if repositories.length === 0}
	<Card.Root>
		<Card.Content class="py-12 text-center">
			<Package class="mx-auto h-12 w-12 text-muted-foreground" />
			<Card.Title class="mt-2 text-sm">No repositories found</Card.Title>
			<Card.Description class="mt-1">
				You don't have any repositories in your GitHub account yet.
			</Card.Description>
		</Card.Content>
	</Card.Root>
{:else}
	<div class="space-y-4">
		{#if error}
			<Alert.Root variant="destructive">
				<AlertCircle class="h-4 w-4" />
				<Alert.Title>Error</Alert.Title>
				<Alert.Description>
					{error}
				</Alert.Description>
			</Alert.Root>
		{/if}

		<div class="grid grid-cols-1 gap-4 md:grid-cols-2">
			{#each repositories as repo (repo.id)}
				<Card.Root class="transition-shadow hover:shadow-md">
					<Card.Header class="pb-3">
						<div class="flex items-start justify-between">
							<div class="min-w-0 flex-1">
								<div class="flex items-center space-x-2">
									{#if repo.private}
										<Lock class="h-4 w-4 flex-shrink-0 text-amber-500" />
									{:else}
										<Unlock class="h-4 w-4 flex-shrink-0 text-green-500" />
									{/if}
									<Card.Title class="text-sm leading-tight font-semibold break-words"
										>{repo.name}</Card.Title
									>
								</div>
								<Card.Description class="text-xs break-words text-muted-foreground">
									{repo.owner}/{repo.name}
								</Card.Description>
							</div>
						</div>
					</Card.Header>

					<Card.Content class="space-y-3">
						{#if repo.description}
							<p class="line-clamp-2 text-sm text-muted-foreground">{repo.description}</p>
						{/if}

						<div class="flex items-center justify-between text-xs text-muted-foreground">
							<div class="flex items-center space-x-4">
								{#if repo.language}
									<Badge variant="secondary" class="text-xs">
										{repo.language}
									</Badge>
								{/if}
								<div class="flex items-center space-x-1">
									<Star class="h-3 w-3" />
									<span>{repo.stargazersCount}</span>
								</div>
								<div class="flex items-center space-x-1">
									<GitFork class="h-3 w-3" />
									<span>{repo.forksCount}</span>
								</div>
								<span>{repo.size ? formatSize(repo.size) : 'N/A'}</span>
							</div>
						</div>

						<div class="flex items-center space-x-1 text-xs text-muted-foreground">
							<Calendar class="h-3 w-3" />
							<span>Updated {repo.updatedAt ? formatDate(repo.updatedAt) : 'N/A'}</span>
						</div>

						<Button
							onclick={() => importRepository(repo)}
							disabled={importing.has(repo.id)}
							class="w-full"
						>
							{#if importing.has(repo.id)}
								<Loader2 class="mr-2 h-4 w-4 animate-spin" />
								Importing...
							{:else}
								Import Repository
							{/if}
						</Button>
					</Card.Content>
				</Card.Root>
			{/each}
		</div>

		{#if hasMore}
			<div class="mt-6 text-center">
				<Button variant="outline" onclick={loadMore} disabled={loading}>
					{#if loading}
						<Loader2 class="mr-2 h-4 w-4 animate-spin" />
						Loading more...
					{:else}
						Load More Repositories
					{/if}
				</Button>
			</div>
		{/if}
	</div>
{/if}
