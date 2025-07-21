<script lang="ts">
	import { onMount } from 'svelte';
	import { authStore } from '$lib/stores/auth';
	import '../app.css';

	let { children } = $props();

	// Initialize auth store on app start
	onMount(() => {
		authStore.init();
	});
</script>

<div class="min-h-screen bg-gray-50">
	<!-- Navigation -->
	<nav class="border-b border-gray-200 bg-white shadow-sm">
		<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
			<div class="flex h-16 justify-between">
				<div class="flex items-center">
					<div class="flex-shrink-0">
						<h1 class="text-xl font-semibold text-gray-900">GitHub Analyzer</h1>
					</div>
					<div class="hidden md:ml-10 md:flex md:space-x-8">
						<a href="/" class="px-3 py-2 text-sm font-medium text-gray-500 hover:text-gray-700"
							>Dashboard</a
						>
						<a
							href="/repositories"
							class="px-3 py-2 text-sm font-medium text-gray-500 hover:text-gray-700"
							>Repositories</a
						>
						<a href="/chat" class="px-3 py-2 text-sm font-medium text-gray-500 hover:text-gray-700"
							>Chat</a
						>
					</div>
				</div>
				<div class="flex items-center space-x-4">
					{#if $authStore.isAuthenticated}
						<!-- Authenticated user navigation -->
						<span class="text-sm text-gray-700">Welcome, {$authStore.user?.name}</span>
						<button
							onclick={() => authStore.logout()}
							class="px-3 py-2 text-sm font-medium text-gray-500 hover:text-gray-700"
						>
							Logout
						</button>
					{:else}
						<!-- Guest navigation -->
						<a
							href="/auth/login"
							class="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700">Login</a
						>
					{/if}
				</div>
			</div>
		</div>
	</nav>

	<!-- Main content -->
	<main class="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
		{@render children()}
	</main>
</div>
