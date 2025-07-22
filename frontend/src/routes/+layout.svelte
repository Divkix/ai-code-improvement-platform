<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { authStore } from '$lib/stores/auth';
	import '../app.css';
	let { children } = $props();

	// Initialize the store as soon as the component is created on the client
	if (browser) {
		authStore.init();
	}

	// This effect is the single source of truth for auth-based navigation.
	$effect(() => {
		if (!browser) return;

		const { isLoading, isAuthenticated } = $authStore;
		const path = $page.url.pathname;
		const isAuthRoute = path.startsWith('/auth');

		// Don't do anything while the store is initializing
		if (isLoading) return;

		// If not authenticated and not on an auth page, redirect to login
		if (!isAuthenticated && !isAuthRoute) {
			goto('/auth/login');
		}

		// If authenticated and on an auth page, redirect to the dashboard
		if (isAuthenticated && isAuthRoute) {
			goto('/');
		}
	});
</script>

<div class="min-h-screen bg-gray-50">
	<nav class="border-b border-gray-200 bg-white shadow-sm">
		<div class="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
			<div class="flex h-16 justify-between">
				<div class="flex items-center">
					<div class="flex-shrink-0">
						<a href="/" class="text-xl font-semibold text-gray-900">GitHub Analyzer</a>
					</div>
					{#if $authStore.isAuthenticated}
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
					{/if}
				</div>
				<div class="flex items-center space-x-4">
					{#if $authStore.isAuthenticated}
						<span class="text-sm text-gray-700">Welcome, {$authStore.user?.name}</span>
						<button
							onclick={() => {
								authStore.logout();
							}}
							class="px-3 py-2 text-sm font-medium text-gray-500 hover:text-gray-700"
						>
							Logout
						</button>
					{:else if !$authStore.isLoading && $page.url.pathname !== '/auth/login'}
						<a
							href="/auth/login"
							class="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
							>Login</a
						>
					{/if}
				</div>
			</div>
		</div>
	</nav>

	<main class="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
		{#if $authStore.isLoading}
			<div class="flex h-96 items-center justify-center">
				<div class="text-center">
					<div
						class="inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-blue-600 border-r-transparent"
					></div>
					<p class="mt-4 text-gray-600">Authenticating...</p>
				</div>
			</div>
		{:else if $authStore.isAuthenticated || $page.url.pathname.startsWith('/auth')}
			{@render children()}
		{:else}
			<div class="flex h-96 items-center justify-center">
				<p class="mt-4 text-gray-600">Redirecting...</p>
			</div>
		{/if}
	</main>
</div>