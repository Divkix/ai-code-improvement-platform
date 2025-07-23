<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { authStore } from '$lib/stores/auth';
	import '../app.css';
	let { children } = $props();

	let mobileMenuOpen = $state(false);

	// Initialize the store as soon as the component is created on the client
	if (browser) {
		authStore.init();
	}

	// This effect is the single source of truth for auth-based navigation.
	$effect(() => {
		if (!browser || $authStore.isLoading) return; // Wait for the auth check to complete

		const { isAuthenticated } = $authStore;
		const path = $page.url.pathname;
		const isAuthRoute = path.startsWith('/auth');

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
							<a
								href="/search"
								class="px-3 py-2 text-sm font-medium text-gray-500 hover:text-gray-700">Search</a
							>
							<a
								href="/chat"
								class="px-3 py-2 text-sm font-medium text-gray-500 hover:text-gray-700">Chat</a
							>
						</div>
					{/if}
				</div>
				<div class="flex items-center space-x-4">
					{#if $authStore.isAuthenticated}
						<!-- Mobile menu button -->
						<button
							onclick={() => (mobileMenuOpen = !mobileMenuOpen)}
							class="inline-flex items-center justify-center rounded-md p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-500 focus:ring-2 focus:ring-blue-500 focus:outline-none md:hidden"
							aria-expanded={mobileMenuOpen}
							aria-controls="mobile-menu"
						>
							<span class="sr-only">Open main menu</span>
							<!-- Hamburger icon -->
							<svg
								class="h-6 w-6"
								fill="none"
								viewBox="0 0 24 24"
								stroke-width="1.5"
								stroke="currentColor"
							>
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5"
								/>
							</svg>
						</button>

						<span class="hidden text-sm text-gray-700 sm:block"
							>Welcome, {$authStore.user?.name}</span
						>
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

		<!-- Mobile menu -->
		{#if $authStore.isAuthenticated && mobileMenuOpen}
			<div class="md:hidden" id="mobile-menu">
				<div class="space-y-1 border-t border-gray-200 bg-white px-2 pt-2 pb-3 sm:px-3">
					<a
						href="/"
						class="block px-3 py-2 text-base font-medium text-gray-500 hover:bg-gray-50 hover:text-gray-700"
						onclick={() => (mobileMenuOpen = false)}
					>
						Dashboard
					</a>
					<a
						href="/repositories"
						class="block px-3 py-2 text-base font-medium text-gray-500 hover:bg-gray-50 hover:text-gray-700"
						onclick={() => (mobileMenuOpen = false)}
					>
						Repositories
					</a>
					<a
						href="/search"
						class="block px-3 py-2 text-base font-medium text-gray-500 hover:bg-gray-50 hover:text-gray-700"
						onclick={() => (mobileMenuOpen = false)}
					>
						Search
					</a>
					<a
						href="/chat"
						class="block px-3 py-2 text-base font-medium text-gray-500 hover:bg-gray-50 hover:text-gray-700"
						onclick={() => (mobileMenuOpen = false)}
					>
						Chat
					</a>
				</div>
			</div>
		{/if}
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
		{:else}
			{@render children()}
		{/if}
	</main>
</div>
