<script lang="ts">
	import { onMount } from 'svelte';
	import { requireAuth } from '$lib/auth';
	
	// Ensure user is authenticated
	onMount(() => {
		requireAuth();
	});

	// Mock data for demo
	const repositories = [
		{
			id: 1,
			name: 'backend-api',
			fullName: 'company/backend-api',
			description: 'Main backend API service for the application',
			language: 'Go',
			status: 'ready',
			lastAnalyzed: '2 hours ago',
			linesOfCode: 45000,
			issues: 3
		},
		{
			id: 2,
			name: 'frontend-web',
			fullName: 'company/frontend-web',
			description: 'React frontend application',
			language: 'TypeScript',
			status: 'importing',
			lastAnalyzed: 'Importing...',
			linesOfCode: 25000,
			issues: 0
		},
		{
			id: 3,
			name: 'mobile-app',
			fullName: 'company/mobile-app',
			description: 'React Native mobile application',
			language: 'JavaScript',
			status: 'ready',
			lastAnalyzed: '1 day ago',
			linesOfCode: 30000,
			issues: 7
		}
	];

	function getStatusColor(status: string) {
		switch (status) {
			case 'ready':
				return 'bg-green-100 text-green-800';
			case 'importing':
				return 'bg-blue-100 text-blue-800';
			case 'error':
				return 'bg-red-100 text-red-800';
			default:
				return 'bg-gray-100 text-gray-800';
		}
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
		<a
			href="/repositories/import"
			class="inline-flex items-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
		>
			Import Repository
		</a>
	</div>

	<!-- Repository List -->
	{#if repositories.length === 0}
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
				<a
					href="/repositories/import"
					class="inline-flex items-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
				>
					Import Repository
				</a>
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

						<p class="mb-4 text-sm text-gray-700">{repo.description}</p>

						<div class="mb-4 grid grid-cols-2 gap-4">
							<div>
								<dt class="text-xs font-medium tracking-wide text-gray-500 uppercase">Language</dt>
								<dd class="text-sm text-gray-900">{repo.language}</dd>
							</div>
							<div>
								<dt class="text-xs font-medium tracking-wide text-gray-500 uppercase">
									Lines of Code
								</dt>
								<dd class="text-sm text-gray-900">{repo.linesOfCode.toLocaleString()}</dd>
							</div>
							<div>
								<dt class="text-xs font-medium tracking-wide text-gray-500 uppercase">
									Issues Found
								</dt>
								<dd class="text-sm text-gray-900">{repo.issues}</dd>
							</div>
							<div>
								<dt class="text-xs font-medium tracking-wide text-gray-500 uppercase">
									Last Analyzed
								</dt>
								<dd class="text-sm text-gray-900">{repo.lastAnalyzed}</dd>
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
								class="rounded-md border border-gray-300 bg-white px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
							>
								Settings
							</button>
						</div>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
