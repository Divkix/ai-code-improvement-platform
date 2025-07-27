<script lang="ts">
	import { onMount } from 'svelte';
	import {
		getRepositories,
		createRepository,
		getRepository,
		deleteRepository
	} from '$lib/api/hooks';
	import { type Repository, type GitHubRepository, type User } from '$lib/api';
	import { authStore } from '$lib/stores/auth';
	import GitHubConnection from '$lib/components/GitHubConnection.svelte';
	import GitHubRepositoryBrowser from '$lib/components/GitHubRepositoryBrowser.svelte';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as Alert from '$lib/components/ui/alert/index.js';
	import { Progress } from '$lib/components/ui/progress/index.js';
	import { Loader2, Github, FolderGit2 } from '@lucide/svelte';
	import { toast } from 'svelte-sonner';

	let user = $state<User | null>(null);
	let repositories = $state<Repository[]>([]);
	let loading = $state(true);
	let error = $state('');
	let showGitHubBrowser = $state(false);
	let pollingInterval: ReturnType<typeof setInterval> | null = null;

	// Watch for auth store changes to get current user
	$effect(() => {
		const auth = authStore.current;
		if (
			auth.user &&
			(!user || user.id !== auth.user.id || user.githubConnected !== auth.user.githubConnected)
		) {
			user = auth.user;
		}
	});

	onMount(() => {
		loadRepositories();

		// Cleanup polling on component destroy
		return () => {
			if (pollingInterval) {
				clearInterval(pollingInterval);
				pollingInterval = null;
			}
		};
	});
	async function loadRepositories() {
		try {
			loading = true;
			error = '';
			const response = await getRepositories();
			repositories = (response.repositories || []).map((repo) => ({
				...repo,
				isPrivate: repo.isPrivate ?? false
			}));

			// Start or stop progress polling based on repository statuses
			manageProgressPolling();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load repositories';
			console.error('Error loading repositories:', err);
			toast.error('Failed to load repositories', {
				description: err instanceof Error ? err.message : 'An unexpected error occurred'
			});
		} finally {
			loading = false;
		}
	}

	function manageProgressPolling() {
		const hasProcessingRepos = repositories.some(
			(repo) =>
				repo.status === 'importing' ||
				repo.status === 'pending' ||
				repo.status === 'queued-embedding' ||
				repo.status === 'embedding'
		);

		if (hasProcessingRepos && !pollingInterval) {
			// Start polling every 3 seconds
			pollingInterval = setInterval(updateRepositoryProgress, 3000);
		} else if (!hasProcessingRepos && pollingInterval) {
			// Stop polling when no repositories are processing
			clearInterval(pollingInterval);
			pollingInterval = null;
		}
	}

	async function updateRepositoryProgress() {
		try {
			const processingRepos = repositories.filter(
				(repo) =>
					repo.status === 'importing' ||
					repo.status === 'pending' ||
					repo.status === 'queued-embedding' ||
					repo.status === 'embedding'
			);

			if (processingRepos.length === 0) {
				// No repos to poll, stop polling
				if (pollingInterval) {
					clearInterval(pollingInterval);
					pollingInterval = null;
				}
				return;
			}

			// Poll each processing repository for updates
			const updatePromises = processingRepos.map(async (repo) => {
				try {
					const updatedRepo = await getRepository(repo.id);
					// Update the repository in our local state
					repositories = repositories.map((r) =>
						r.id === repo.id ? { ...updatedRepo, isPrivate: updatedRepo.isPrivate ?? false } : r
					);
					return updatedRepo;
				} catch (err) {
					console.error(`Failed to update repository ${repo.id}:`, err);
					return null;
				}
			});

			await Promise.all(updatePromises);

			// Check if we should continue polling
			manageProgressPolling();
		} catch (err) {
			console.error('Error updating repository progress:', err);
		}
	}

	async function handleDeleteRepository(repo: Repository) {
		if (!confirm(`Are you sure you want to delete "${repo.name}"?`)) {
			return;
		}

		try {
			await deleteRepository(repo.id);
			repositories = repositories.filter((r) => r.id !== repo.id);
			toast.success('Repository deleted', {
				description: `${repo.name} has been removed from your repositories.`
			});
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete repository';
			toast.error('Failed to delete repository', {
				description: err instanceof Error ? err.message : 'An unexpected error occurred'
			});
		}
	}

	function getStatusText(status: string) {
		switch (status) {
			case 'pending':
				return 'pending import';
			case 'queued-embedding':
				return 'embedding pending';
			case 'importing':
				return 'importing';
			case 'embedding':
				return 'embedding';
			case 'ready':
				return 'ready';
			case 'error':
				return 'error';
			default:
				return status;
		}
	}

	function formatDate(dateString: string) {
		const date = new Date(dateString);
		const now = new Date();
		const diffInHours = (now.getTime() - date.getTime()) / (1000 * 60 * 60);
		if (diffInHours < 1) {
			return 'Just now';
		} else if (diffInHours < 24) {
			return `${Math.floor(diffInHours)} hours ago`;
		} else {
			const diffInDays = Math.floor(diffInHours / 24);
			return `${diffInDays} day${diffInDays > 1 ? 's' : ''} ago`;
		}
	}

	function getLinesOfCode(repo: Repository) {
		return repo.stats?.totalLines || 0;
	}

	function getLastAnalyzed(repo: Repository) {
		if (repo.status === 'importing') {
			return `Importing... ${repo.importProgress}%`;
		}
		if (repo.lastSyncedAt) {
			return formatDate(repo.lastSyncedAt);
		}
		return 'Never';
	}

	function openGitHubBrowser() {
		showGitHubBrowser = true;
	}

	function closeGitHubBrowser() {
		showGitHubBrowser = false;
		error = '';
	}

	async function handleGitHubRepositoryImport(githubRepo: GitHubRepository) {
		try {
			// Use the new GitHub import API endpoint
			const newRepo = await createRepository({
				name: githubRepo.name,
				owner: githubRepo.owner,
				fullName: githubRepo.fullName,
				description: githubRepo.description,
				githubRepoId: githubRepo.id,
				primaryLanguage: githubRepo.language,
				isPrivate: githubRepo.private
			});

			repositories = [{ ...newRepo, isPrivate: newRepo.isPrivate ?? false }, ...repositories];
			closeGitHubBrowser();
			error = '';

			// Start polling if the new repository is importing
			manageProgressPolling();

			toast.success('Repository imported successfully', {
				description: `${githubRepo.fullName} has been imported and is being processed.`
			});
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to import repository';
			toast.error('Failed to import repository', {
				description: err instanceof Error ? err.message : 'An unexpected error occurred'
			});
		}
	}
</script>

<svelte:head>
	<title>Repositories - ACIP</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Repositories</h1>
			<p class="text-gray-600">Manage and analyze your GitHub repositories</p>
		</div>
		{#if user?.githubConnected}
			<Button onclick={openGitHubBrowser}>
				<Github class="mr-2 h-4 w-4" />
				Browse GitHub
			</Button>
		{/if}
	</div>

	<!-- GitHub Connection Status -->
	{#if user}
		<GitHubConnection {user} />
	{/if}

	<!-- Error Display -->
	{#if error}
		<Alert.Root variant="destructive">
			<Alert.Title>Error</Alert.Title>
			<Alert.Description>
				<p>{error}</p>
				<div class="mt-4">
					<Button
						variant="outline"
						size="sm"
						onclick={() => {
							error = '';
							loadRepositories();
						}}
					>
						Dismiss
					</Button>
				</div>
			</Alert.Description>
		</Alert.Root>
	{/if}

	{#if loading}
		<div class="py-12 text-center">
			<Loader2 class="mx-auto h-8 w-8 animate-spin" />
			<p class="mt-2 text-muted-foreground">Loading repositories...</p>
		</div>
	{:else if !repositories || repositories.length === 0}
		<Card.Root class="py-12">
			<Card.Content class="text-center">
				<div class="mx-auto mb-4 h-12 w-12 text-muted-foreground">
					<FolderGit2 class="h-12 w-12" />
				</div>
				<Card.Title class="text-sm">No repositories</Card.Title>
				<Card.Description class="mt-1">
					Get started by connecting your GitHub account and importing your first repository.
				</Card.Description>
				{#if user?.githubConnected}
					<div class="mt-6">
						<Button onclick={openGitHubBrowser}>
							<Github class="mr-2 h-4 w-4" />
							Browse GitHub
						</Button>
					</div>
				{/if}
			</Card.Content>
		</Card.Root>
	{:else}
		<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
			{#each repositories as repo (repo.id)}
				<Card.Root>
					<Card.Header>
						<div class="flex items-center justify-between">
							<div class="flex items-center">
								<div class="flex-shrink-0">
									<div class="flex h-8 w-8 items-center justify-center rounded-full bg-muted">
										<FolderGit2 class="h-4 w-4 text-muted-foreground" />
									</div>
								</div>
								<div class="ml-3">
									<Card.Title class="text-lg">
										<a
											href="https://github.com/{repo.fullName}"
											target="_blank"
											rel="noopener noreferrer"
											class="transition-colors hover:text-primary"
										>
											{repo.name}
										</a>
									</Card.Title>
									<Card.Description>{repo.fullName}</Card.Description>
								</div>
							</div>
							<Badge variant={repo.status === 'ready' ? 'default' : 'secondary'}>
								{getStatusText(repo.status)}
							</Badge>
						</div>
					</Card.Header>

					<Card.Content>
						<p class="mb-4 text-sm text-muted-foreground">{repo.description || 'No description'}</p>

						<!-- Import Progress Bar for importing repositories -->
						{#if repo.status === 'importing' || repo.status === 'pending'}
							<div class="mb-4">
								<div class="mb-2 flex items-center justify-between">
									<span class="text-xs font-medium">
										{repo.status === 'pending' ? 'Preparing import...' : 'Importing repository...'}
									</span>
									<span class="text-xs text-muted-foreground">{repo.importProgress}%</span>
								</div>
								<Progress value={repo.importProgress} class="w-full" />
								{#if repo.status === 'importing'}
									<div class="mt-2 flex items-center text-xs text-primary">
										<Loader2 class="mr-2 h-3 w-3 animate-spin" />
										Processing repository files...
									</div>
								{/if}
							</div>
						{/if}

						<div class="mb-4 grid grid-cols-2 gap-4">
							<div>
								<dt class="text-xs font-medium tracking-wide text-muted-foreground uppercase">
									Language
								</dt>
								<dd class="text-sm">{repo.primaryLanguage || 'Unknown'}</dd>
							</div>
							<div>
								<dt class="text-xs font-medium tracking-wide text-muted-foreground uppercase">
									Lines of Code
								</dt>
								<dd class="text-sm">{getLinesOfCode(repo).toLocaleString()}</dd>
							</div>
							<div>
								<dt class="text-xs font-medium tracking-wide text-muted-foreground uppercase">
									Progress
								</dt>
								<dd class="text-sm">{repo.importProgress}%</dd>
							</div>
							<div>
								<dt class="text-xs font-medium tracking-wide text-muted-foreground uppercase">
									Last Updated
								</dt>
								<dd class="text-sm">{getLastAnalyzed(repo)}</dd>
							</div>
						</div>
					</Card.Content>

					<Card.Footer class="flex space-x-2">
						<Button href="/chat?repo={repo.id}" class="flex-1">Analyze Code</Button>
						<Button variant="outline" onclick={() => handleDeleteRepository(repo)}>Delete</Button>
					</Card.Footer>
				</Card.Root>
			{/each}
		</div>
	{/if}
</div>

<!-- GitHub Repository Browser Modal -->
<Dialog.Root bind:open={showGitHubBrowser}>
	<Dialog.Content class="max-h-[90vh] max-w-[calc(100%-2rem)] overflow-y-auto sm:max-w-4xl">
		<Dialog.Header>
			<Dialog.Title>Browse GitHub Repositories</Dialog.Title>
		</Dialog.Header>
		{#if user}
			<GitHubRepositoryBrowser {user} onRepositoryImport={handleGitHubRepositoryImport} />
		{:else}
			<div class="flex items-center justify-center p-8">
				<p class="text-muted-foreground">Loading user information...</p>
			</div>
		{/if}
		<Dialog.Footer>
			<Button variant="outline" onclick={closeGitHubBrowser}>Close</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
