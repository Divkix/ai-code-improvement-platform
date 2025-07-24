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

	let user = $state<User | null>(null);
	let repositories = $state<Repository[]>([]);
	let loading = $state(true);
	let error = $state('');
	let showAddModal = $state(false);
	let showGitHubBrowser = $state(false);
	let githubUrl = $state('');
	let importMethod = $state<'url' | 'github'>('github');
	let pollingInterval: ReturnType<typeof setInterval> | null = null;

	// Subscribe to auth store to get current user
	authStore.subscribe((auth) => {
		if (auth.user && (!user || user.id !== auth.user.id)) {
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
		} finally {
			loading = false;
		}
	}

	function manageProgressPolling() {
		const hasImportingRepos = repositories.some(
			(repo) => repo.status === 'importing' || repo.status === 'pending'
		);

		if (hasImportingRepos && !pollingInterval) {
			// Start polling every 3 seconds
			pollingInterval = setInterval(updateRepositoryProgress, 3000);
		} else if (!hasImportingRepos && pollingInterval) {
			// Stop polling when no repositories are importing
			clearInterval(pollingInterval);
			pollingInterval = null;
		}
	}

	async function updateRepositoryProgress() {
		try {
			const importingRepos = repositories.filter(
				(repo) => repo.status === 'importing' || repo.status === 'pending'
			);

			if (importingRepos.length === 0) {
				// No repos to poll, stop polling
				if (pollingInterval) {
					clearInterval(pollingInterval);
					pollingInterval = null;
				}
				return;
			}

			// Poll each importing repository for updates
			const updatePromises = importingRepos.map(async (repo) => {
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

	async function handleAddRepository(event: Event) {
		event.preventDefault();
		if (!githubUrl.trim()) {
			error = 'Please enter a GitHub repository URL';
			return;
		}

		try {
			// Parse GitHub URL to extract owner and repo name
			const parsed = parseGitHubUrl(githubUrl);
			if (!parsed) {
				error = 'Invalid GitHub repository URL. Please use format: https://github.com/owner/repo';
				return;
			}

			const newRepo = await createRepository({
				name: parsed.name,
				owner: parsed.owner,
				fullName: parsed.fullName,
				isPrivate: false // We'll assume public for now since we can't detect this from URL
			});
			repositories = [{ ...newRepo, isPrivate: newRepo.isPrivate ?? false }, ...repositories];
			showAddModal = false;
			githubUrl = '';
			error = '';

			// Start polling if the new repository is importing
			manageProgressPolling();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create repository';
		}
	}

	function parseGitHubUrl(url: string): { owner: string; name: string; fullName: string } | null {
		try {
			// Handle different GitHub URL formats
			const cleanUrl = url.trim().replace(/\.git$/, '');
			let match;

			// Match https://github.com/owner/repo
			match = cleanUrl.match(/https?:\/\/github\.com\/([^/]+)\/([^/]+)/);
			if (match) {
				const owner = match[1];
				const name = match[2];
				return { owner, name, fullName: `${owner}/${name}` };
			}

			// Match owner/repo format
			match = cleanUrl.match(/^([^/]+)\/([^/]+)$/);
			if (match) {
				const owner = match[1];
				const name = match[2];
				return { owner, name, fullName: `${owner}/${name}` };
			}

			return null;
		} catch {
			return null;
		}
	}

	async function handleDeleteRepository(repo: Repository) {
		if (!confirm(`Are you sure you want to delete "${repo.name}"?`)) {
			return;
		}

		try {
			await deleteRepository(repo.id);
			repositories = repositories.filter((r) => r.id !== repo.id);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to delete repository';
		}
	}

	function getStatusColor(status: string) {
		switch (status) {
			case 'ready':
				return 'bg-green-100 text-green-800';
			case 'importing':
				return 'bg-blue-100 text-blue-800';
			case 'pending':
				return 'bg-yellow-100 text-yellow-800';
			case 'error':
				return 'bg-red-100 text-red-800';
			default:
				return 'bg-gray-100 text-gray-800';
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

	function openAddModal() {
		showAddModal = true;
		// Default to GitHub import if user is connected, otherwise URL
		importMethod = user?.githubConnected ? 'github' : 'url';
	}

	function closeAddModal() {
		showAddModal = false;
		showGitHubBrowser = false;
		githubUrl = '';
		error = '';
		importMethod = 'github';
	}

	function openGitHubBrowser() {
		showGitHubBrowser = true;
	}

	function closeGitHubBrowser() {
		showGitHubBrowser = false;
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
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to import repository';
		}
	}
</script>

<svelte:head>
	<title>Repositories - GitHub Analyzer</title>
</svelte:head>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Repositories</h1>
			<p class="text-gray-600">Manage and analyze your GitHub repositories</p>
		</div>
		<div class="flex space-x-3">
			{#if user?.githubConnected}
				<button
					onclick={openGitHubBrowser}
					class="inline-flex items-center rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
				>
					<svg class="mr-2 h-4 w-4" fill="currentColor" viewBox="0 0 20 20">
						<path
							fill-rule="evenodd"
							d="M10 0C4.477 0 0 4.484 0 10.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0110 4.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.203 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.942.359.31.678.921.678 1.856 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0020 10.017C20 4.484 15.522 0 10 0z"
							clip-rule="evenodd"
						/>
					</svg>
					Browse GitHub
				</button>
			{/if}
			<button
				onclick={openAddModal}
				class="inline-flex items-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
			>
				Add Repository
			</button>
		</div>
	</div>

	<!-- GitHub Connection Status -->
	{#if user}
		<GitHubConnection {user} />
	{/if}

	<!-- Error Display -->
	{#if error}
		<div class="rounded-md bg-red-50 p-4">
			<div class="flex">
				<div class="ml-3">
					<h3 class="text-sm font-medium text-red-800">Error</h3>
					<p class="mt-2 text-sm text-red-700">{error}</p>
					<div class="mt-4">
						<button
							onclick={() => {
								error = '';
								loadRepositories();
							}}
							class="rounded-md bg-red-100 px-2 py-1 text-sm font-medium text-red-800 hover:bg-red-200"
						>
							Dismiss
						</button>
					</div>
				</div>
			</div>
		</div>
	{/if}

	{#if loading}
		<div class="py-12 text-center">
			<div
				class="mx-auto h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent"
			></div>
			<p class="mt-2 text-gray-600">Loading repositories...</p>
		</div>
	{:else if !repositories || repositories.length === 0}
		<div class="py-12 text-center">
			<div class="mx-auto h-12 w-12 text-gray-400">
				<svg fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						stroke-width="2"
						d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2H5a2 2 0 00-2-2z"
					/>
				</svg>
			</div>
			<h3 class="mt-2 text-sm font-medium text-gray-900">No repositories</h3>
			<p class="mt-1 text-sm text-gray-500">Get started by importing your first repository.</p>
			<div class="mt-6">
				<button
					onclick={openAddModal}
					class="inline-flex items-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
				>
					Add Repository
				</button>
			</div>
		</div>
	{:else}
		<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
			{#each repositories as repo (repo.id)}
				<div class="overflow-hidden rounded-lg bg-white shadow">
					<div class="p-6">
						<div class="mb-4 flex items-center justify-between">
							<div class="flex items-center">
								<div class="flex-shrink-0">
									<div class="flex h-8 w-8 items-center justify-center rounded-full bg-gray-200">
										<svg class="h-4 w-4 text-gray-600" fill="currentColor" viewBox="0 0 20 20">
											<path
												fill-rule="evenodd"
												d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z"
												clip-rule="evenodd"
											/>
										</svg>
									</div>
								</div>
								<div class="ml-3">
									<h3 class="text-lg font-medium text-gray-900">{repo.name}</h3>
									<p class="text-sm text-gray-500">{repo.fullName}</p>
								</div>
							</div>
							<span
								class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {getStatusColor(
									repo.status
								)}"
							>
								{repo.status}
							</span>
						</div>

						<p class="mb-4 text-sm text-gray-700">{repo.description || 'No description'}</p>

						<!-- Import Progress Bar for importing repositories -->
						{#if repo.status === 'importing' || repo.status === 'pending'}
							<div class="mb-4">
								<div class="mb-2 flex items-center justify-between">
									<span class="text-xs font-medium text-gray-700">
										{repo.status === 'pending' ? 'Preparing import...' : 'Importing repository...'}
									</span>
									<span class="text-xs text-gray-500">{repo.importProgress}%</span>
								</div>
								<div class="h-2 w-full rounded-full bg-gray-200">
									<div
										class="h-2 rounded-full bg-blue-600 transition-all duration-300 ease-out"
										style="width: {repo.importProgress}%"
									></div>
								</div>
								{#if repo.status === 'importing'}
									<div class="mt-2 flex items-center text-xs text-blue-600">
										<svg class="-ml-1 mr-2 h-3 w-3 animate-spin" fill="none" viewBox="0 0 24 24">
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
										Processing repository files...
									</div>
								{/if}
							</div>
						{/if}

						<div class="mb-4 grid grid-cols-2 gap-4">
							<div>
								<dt class="text-xs font-medium uppercase tracking-wide text-gray-500">Language</dt>
								<dd class="text-sm text-gray-900">{repo.primaryLanguage || 'Unknown'}</dd>
							</div>
							<div>
								<dt class="text-xs font-medium uppercase tracking-wide text-gray-500">
									Lines of Code
								</dt>
								<dd class="text-sm text-gray-900">{getLinesOfCode(repo).toLocaleString()}</dd>
							</div>
							<div>
								<dt class="text-xs font-medium uppercase tracking-wide text-gray-500">Progress</dt>
								<dd class="text-sm text-gray-900">{repo.importProgress}%</dd>
							</div>
							<div>
								<dt class="text-xs font-medium uppercase tracking-wide text-gray-500">
									Last Updated
								</dt>
								<dd class="text-sm text-gray-900">{getLastAnalyzed(repo)}</dd>
							</div>
						</div>

						<div class="flex space-x-2">
							<a
								href="/chat?repo={repo.id}"
								class="flex-1 rounded-md bg-blue-600 px-3 py-2 text-center text-sm font-medium text-white hover:bg-blue-700"
							>
								Analyze Code
							</a>
							<button
								type="button"
								onclick={() => handleDeleteRepository(repo)}
								class="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
							>
								Delete
							</button>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

{#if showAddModal}
	<div class="fixed inset-0 z-50 overflow-y-auto">
		<div class="flex min-h-screen items-center justify-center p-4">
			<div
				class="fixed inset-0 bg-black bg-opacity-50"
				onclick={closeAddModal}
				onkeydown={closeAddModal}
				role="button"
				tabindex="0"
			></div>
			<div class="relative w-full max-w-3xl rounded-lg bg-white p-6 shadow-lg">
				<h3 class="mb-4 text-lg font-medium text-gray-900">Add Repository</h3>

				<!-- Import Method Selection -->
				{#if user?.githubConnected}
					<div class="mb-6">
						<fieldset>
							<legend class="text-base font-medium text-gray-900">Import Method</legend>
							<div class="mt-2 space-y-2">
								<label class="flex items-center">
									<input
										type="radio"
										bind:group={importMethod}
										value="github"
										class="h-4 w-4 border-gray-300 text-blue-600 focus:ring-blue-500"
									/>
									<span class="ml-3 block text-sm font-medium text-gray-700">
										Browse your GitHub repositories
									</span>
								</label>
								<label class="flex items-center">
									<input
										type="radio"
										bind:group={importMethod}
										value="url"
										class="h-4 w-4 border-gray-300 text-blue-600 focus:ring-blue-500"
									/>
									<span class="ml-3 block text-sm font-medium text-gray-700">
										Enter repository URL manually
									</span>
								</label>
							</div>
						</fieldset>
					</div>
				{/if}

				{#if importMethod === 'github' && user?.githubConnected}
					<!-- GitHub Repository Browser -->
					<div class="mb-4">
						<GitHubRepositoryBrowser {user} onRepositoryImport={handleGitHubRepositoryImport} />
					</div>

					<div class="flex justify-end">
						<button
							type="button"
							onclick={closeAddModal}
							class="rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
						>
							Cancel
						</button>
					</div>
				{:else}
					<!-- Manual URL Entry Form -->
					<form onsubmit={handleAddRepository}>
						<div class="mb-4">
							<label for="githubUrl" class="block text-sm font-medium text-gray-700"
								>GitHub Repository URL</label
							>
							<input
								type="url"
								id="githubUrl"
								bind:value={githubUrl}
								class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-blue-500"
								placeholder="https://github.com/owner/repository or owner/repository"
								required
							/>
							<p class="mt-1 text-xs text-gray-500">
								Enter a GitHub repository URL or owner/repository format
							</p>
						</div>

						<div class="flex space-x-3">
							<button
								type="button"
								onclick={closeAddModal}
								class="flex-1 rounded-md border border-gray-300 bg-white px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
							>
								Cancel
							</button>
							<button
								type="submit"
								class="flex-1 rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
							>
								Add Repository
							</button>
						</div>
					</form>
				{/if}
			</div>
		</div>
	</div>
{/if}

<!-- GitHub Repository Browser Modal -->
{#if showGitHubBrowser && user?.githubConnected}
	<div class="fixed inset-0 z-50 overflow-y-auto">
		<div class="flex min-h-screen items-center justify-center p-4">
			<div
				class="fixed inset-0 bg-black bg-opacity-50"
				onclick={closeGitHubBrowser}
				onkeydown={closeGitHubBrowser}
				role="button"
				tabindex="0"
			></div>
			<div class="relative w-full max-w-4xl rounded-lg bg-white p-6 shadow-lg">
				<div class="mb-4 flex items-center justify-between">
					<h3 class="text-lg font-medium text-gray-900">Browse GitHub Repositories</h3>
					<button
						onclick={closeGitHubBrowser}
						class="rounded-md p-1 text-gray-400 hover:text-gray-500"
						aria-label="Close GitHub repository browser"
					>
						<svg class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M6 18L18 6M6 6l12 12"
							/>
						</svg>
					</button>
				</div>

				<GitHubRepositoryBrowser {user} onRepositoryImport={handleGitHubRepositoryImport} />
			</div>
		</div>
	</div>
{/if}
