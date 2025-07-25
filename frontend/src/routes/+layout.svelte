<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { authStore } from '$lib/stores/auth';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Sheet from '$lib/components/ui/sheet/index.js';
	import { Menu, Loader2 } from '@lucide/svelte';
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
						<div class="hidden md:ml-10 md:flex md:space-x-2">
							<Button variant="ghost" size="sm" href="/">Dashboard</Button>
							<Button variant="ghost" size="sm" href="/repositories">Repositories</Button>
							<Button variant="ghost" size="sm" href="/search">Search</Button>
							<Button variant="ghost" size="sm" href="/chat">Chat</Button>
						</div>
					{/if}
				</div>
				<div class="flex items-center space-x-4">
					{#if $authStore.isAuthenticated}
						<Sheet.Root bind:open={mobileMenuOpen}>
							<Sheet.Trigger
								class="inline-flex items-center justify-center rounded-md p-2 text-gray-400 hover:bg-gray-100 hover:text-gray-500 focus:ring-2 focus:ring-blue-500 focus:outline-none md:hidden"
							>
								<Menu class="h-6 w-6" />
								<span class="sr-only">Open main menu</span>
							</Sheet.Trigger>
							<Sheet.Content side="left">
								<Sheet.Header>
									<Sheet.Title>Navigation</Sheet.Title>
								</Sheet.Header>
								<div class="mt-4 flex flex-col space-y-2">
									<Button variant="ghost" href="/" onclick={() => (mobileMenuOpen = false)}>
										Dashboard
									</Button>
									<Button
										variant="ghost"
										href="/repositories"
										onclick={() => (mobileMenuOpen = false)}
									>
										Repositories
									</Button>
									<Button variant="ghost" href="/search" onclick={() => (mobileMenuOpen = false)}>
										Search
									</Button>
									<Button variant="ghost" href="/chat" onclick={() => (mobileMenuOpen = false)}>
										Chat
									</Button>
								</div>
							</Sheet.Content>
						</Sheet.Root>

						<span class="hidden text-sm text-gray-700 sm:block"
							>Welcome, {$authStore.user?.name}</span
						>
						<Button
							variant="ghost"
							size="sm"
							onclick={() => {
								authStore.logout();
							}}
						>
							Logout
						</Button>
					{:else if !$authStore.isLoading && $page.url.pathname !== '/auth/login'}
						<Button href="/auth/login">Login</Button>
					{/if}
				</div>
			</div>
		</div>
	</nav>

	<main class="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
		{#if $authStore.isLoading}
			<div class="flex h-96 items-center justify-center">
				<div class="flex items-center justify-center">
					<Loader2 class="h-8 w-8 animate-spin" />
					<span class="ml-2">Authenticating...</span>
				</div>
			</div>
		{:else}
			{@render children()}
		{/if}
	</main>
</div>
