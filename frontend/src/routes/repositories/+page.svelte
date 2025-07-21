<script lang="ts">
	import { onMount } from 'svelte';
	import { requireAuth } from '$lib/auth';
	import apiClient, { type Repository } from '$lib/api';

	// Ensure user is authenticated
	onMount(() => {
		requireAuth();
		loadRepositories();
	});

	let repositories: Repository[] = [];
	let loading = true;
	let error = '';
	let showAddModal = false;
	let addForm = {
		name: '',
		owner: '',
		fullName: '',
		description: '',
		primaryLanguage: '',
		isPrivate: false
	};

	async function loadRepositories() {
		try {
			loading = true;
			error = '';
			const response = await apiClient.getRepositories();
			repositories = response.repositories;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load repositories';
			console.error('Error loading repositories:', err);
		} finally {
			loading = false;
		}
	}

	async function handleAddRepository() {
		if (!addForm.name || !addForm.owner || !addForm.fullName) {
			return;
		}

		try {
			const newRepo = await apiClient.createRepository({
				name: addForm.name,
				owner: addForm.owner,
				fullName: addForm.fullName,
				description: addForm.description || undefined,
				primaryLanguage: addForm.primaryLanguage || undefined,
				isPrivate: addForm.isPrivate
			});

			repositories = [newRepo, ...repositories];
			showAddModal = false;
			
			// Reset form
			addForm = {
				name: '',
				owner: '',
				fullName: '',
				description: '',
				primaryLanguage: '',
				isPrivate: false
			};
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create repository';
		}
	}

	async function handleDeleteRepository(repo: Repository) {
		if (!confirm(`Are you sure you want to delete "${repo.name}"?`)) {
			return;
		}

		try {
			await apiClient.deleteRepository(repo.id);
			repositories = repositories.filter(r => r.id !== repo.id);
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
	}

	function closeAddModal() {
		showAddModal = false;
		addForm = {
			name: '',
			owner: '',
			fullName: '',
			description: '',
			primaryLanguage: '',
			isPrivate: false
		};
	}

	// Auto-fill fullName when owner and name are entered
	$: if (addForm.owner && addForm.name) {
		addForm.fullName = `${addForm.owner}/${addForm.name}`;
	}
</script>

<svelte:head>
	<title>Repositories - GitHub Analyzer</title>
</svelte:head>

<div class="space-y-6">
	<!-- Header -->
	<div class="flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold text-gray-900">Repositories</h1>
			<p class="text-gray-600">Manage and analyze your GitHub repositories</p>
		</div>
		<button
			on:click={openAddModal}
			class="inline-flex items-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
		>
			Add Repository
		</button>
	</div>

	<!-- Error Message -->
	{#if error}
		<div class="rounded-md bg-red-50 p-4">
			<div class="flex">
				<div class="ml-3">
					<h3 class="text-sm font-medium text-red-800">Error</h3>
					<p class="mt-2 text-sm text-red-700">{error}</p>
					<div class="mt-4">
						<button
							on:click={loadRepositories}
							class="rounded-md bg-red-100 px-2 py-1 text-sm font-medium text-red-800 hover:bg-red-200"
						>
							Try Again
						</button>
					</div>
				</div>
			</div>
		</div>
	{/if}

	<!-- Loading State -->
	{#if loading}
		<div class="py-12 text-center">
			<div class="mx-auto h-8 w-8 animate-spin rounded-full border-4 border-blue-600 border-t-transparent"></div>
			<p class="mt-2 text-gray-600">Loading repositories...</p>
		</div>
	{:else if repositories.length === 0}
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
					on:click={openAddModal}
					class="inline-flex items-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
				>
					Add Repository
				</button>
			</div>
		</div>
	{:else}
		<div class="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
			{#each repositories as repo}
				<div class="overflow-hidden rounded-lg bg-white shadow">
					<div class="p-6">
						<div class="mb-4 flex items-center justify-between">
							<div class="flex items-center">
								<div class="flex-shrink-0">
									<div class="flex h-8 w-8 items-center justify-center rounded-full bg-gray-200">
										<svg class="h-4 w-4 text-gray-600" fill="currentColor" viewBox="0 0 20 20">
											<path
												fillRule="evenodd"
												d="M3 4a1 1 0 011-1h12a1 1 0 011 1v2a1 1 0 01-1 1H4a1 1 0 01-1-1V4zM3 10a1 1 0 011-1h6a1 1 0 011 1v6a1 1 0 01-1 1H4a1 1 0 01-1-1v-6zM14 9a1 1 0 00-1 1v6a1 1 0 001 1h2a1 1 0 001-1v-6a1 1 0 00-1-1h-2z"
												clipRule="evenodd"
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

						<div class="mb-4 grid grid-cols-2 gap-4">
							<div>
								<dt class="text-xs font-medium tracking-wide text-gray-500 uppercase">Language</dt>
								<dd class="text-sm text-gray-900">{repo.primaryLanguage || 'Unknown'}</dd>
							</div>
							<div>
								<dt class="text-xs font-medium tracking-wide text-gray-500 uppercase">
									Lines of Code
								</dt>
								<dd class="text-sm text-gray-900">{getLinesOfCode(repo).toLocaleString()}</dd>
							</div>
							<div>
								<dt class="text-xs font-medium tracking-wide text-gray-500 uppercase">
									Progress
								</dt>
								<dd class="text-sm text-gray-900">{repo.importProgress}%</dd>
							</div>
							<div>
								<dt class="text-xs font-medium tracking-wide text-gray-500 uppercase">
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
								on:click={() => handleDeleteRepository(repo)}
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

<!-- Add Repository Modal -->
{#if showAddModal}
	<div class="fixed inset-0 z-50 overflow-y-auto">
		<div class="flex min-h-screen items-center justify-center p-4">
			<div class="fixed inset-0 bg-black bg-opacity-50" on:click={closeAddModal} on:keydown={closeAddModal}></div>
			<div class="relative w-full max-w-md rounded-lg bg-white p-6 shadow-lg">
				<h3 class="mb-4 text-lg font-medium text-gray-900">Add Repository</h3>
				
				<form on:submit|preventDefault={handleAddRepository}>
					<div class="mb-4">
						<label for="owner" class="block text-sm font-medium text-gray-700">Owner</label>
						<input
							type="text"
							id="owner"
							bind:value={addForm.owner}
							class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-blue-500"
							placeholder="e.g., microsoft"
							required
						/>
					</div>

					<div class="mb-4">
						<label for="name" class="block text-sm font-medium text-gray-700">Repository Name</label>
						<input
							type="text"
							id="name"
							bind:value={addForm.name}
							class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-blue-500"
							placeholder="e.g., vscode"
							required
						/>
					</div>

					<div class="mb-4">
						<label for="fullName" class="block text-sm font-medium text-gray-700">Full Name</label>
						<input
							type="text"
							id="fullName"
							bind:value={addForm.fullName}
							class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm bg-gray-50"
							placeholder="owner/repository"
							readonly
						/>
					</div>

					<div class="mb-4">
						<label for="description" class="block text-sm font-medium text-gray-700">Description (Optional)</label>
						<textarea
							id="description"
							bind:value={addForm.description}
							rows="2"
							class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-blue-500"
							placeholder="Brief description of the repository"
						></textarea>
					</div>

					<div class="mb-4">
						<label for="primaryLanguage" class="block text-sm font-medium text-gray-700">Primary Language (Optional)</label>
						<input
							type="text"
							id="primaryLanguage"
							bind:value={addForm.primaryLanguage}
							class="mt-1 block w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-blue-500"
							placeholder="e.g., TypeScript"
						/>
					</div>

					<div class="mb-6">
						<label class="flex items-center">
							<input
								type="checkbox"
								bind:checked={addForm.isPrivate}
								class="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
							/>
							<span class="ml-2 text-sm text-gray-700">Private repository</span>
						</label>
					</div>

					<div class="flex space-x-3">
						<button
							type="button"
							on:click={closeAddModal}
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
			</div>
		</div>
	</div>
{/if}
