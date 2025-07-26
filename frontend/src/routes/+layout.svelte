<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { browser } from '$app/environment';
	import { authStore, type AuthState } from '$lib/stores/auth';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Sheet from '$lib/components/ui/sheet/index.js';
	import * as NavigationMenu from '$lib/components/ui/navigation-menu/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import { Toaster } from '$lib/components/ui/sonner/index.js';
	import { Menu } from '@lucide/svelte';
	import '../app.css';
	let { children } = $props();

	let mobileMenuOpen = $state(false);

	// Initialize the store as soon as the component is created on the client
	if (browser) {
		authStore.init();
	}

	// Subscribe to auth store reactively
	let authState = $state<AuthState>({
		user: null,
		token: null,
		isAuthenticated: false,
		isLoading: true
	});
	$effect(() => {
		const unsubscribe = authStore.subscribe((state) => {
			authState = state;
		});
		return unsubscribe;
	});

	// This effect is the single source of truth for auth-based navigation.
	$effect(() => {
		if (!browser || !authState || authState.isLoading) return; // Wait for the auth check to complete

		const { isAuthenticated } = authState;
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

<div class="min-h-screen bg-background">
	<header
		class="sticky top-0 z-50 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60"
	>
		<div class="container mx-auto px-4 sm:px-6 lg:px-8">
			<div class="flex h-16 items-center justify-between">
				<div class="flex items-center">
					<a href="/" class="mr-6 flex items-center space-x-2">
						<span class="text-xl font-bold">GitHub Analyzer</span>
					</a>
					{#if authState?.isAuthenticated}
						<NavigationMenu.Root class="hidden md:flex">
							<NavigationMenu.List>
								<NavigationMenu.Item>
									<NavigationMenu.Link
										href="/"
										class="group inline-flex h-10 w-max items-center justify-center rounded-md bg-background px-4 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground focus:outline-none disabled:pointer-events-none disabled:opacity-50 data-[active]:bg-accent/50 data-[state=open]:bg-accent/50"
									>
										Dashboard
									</NavigationMenu.Link>
								</NavigationMenu.Item>
								<NavigationMenu.Item>
									<NavigationMenu.Link
										href="/repositories"
										class="group inline-flex h-10 w-max items-center justify-center rounded-md bg-background px-4 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground focus:outline-none disabled:pointer-events-none disabled:opacity-50 data-[active]:bg-accent/50 data-[state=open]:bg-accent/50"
									>
										Repositories
									</NavigationMenu.Link>
								</NavigationMenu.Item>
								<NavigationMenu.Item>
									<NavigationMenu.Link
										href="/search"
										class="group inline-flex h-10 w-max items-center justify-center rounded-md bg-background px-4 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground focus:outline-none disabled:pointer-events-none disabled:opacity-50 data-[active]:bg-accent/50 data-[state=open]:bg-accent/50"
									>
										Search
									</NavigationMenu.Link>
								</NavigationMenu.Item>
								<NavigationMenu.Item>
									<NavigationMenu.Link
										href="/chat"
										class="group inline-flex h-10 w-max items-center justify-center rounded-md bg-background px-4 py-2 text-sm font-medium transition-colors hover:bg-accent hover:text-accent-foreground focus:bg-accent focus:text-accent-foreground focus:outline-none disabled:pointer-events-none disabled:opacity-50 data-[active]:bg-accent/50 data-[state=open]:bg-accent/50"
									>
										Chat
									</NavigationMenu.Link>
								</NavigationMenu.Item>
							</NavigationMenu.List>
						</NavigationMenu.Root>
					{/if}
				</div>
				<div class="flex items-center space-x-4">
					{#if authState?.isAuthenticated}
						<Sheet.Root bind:open={mobileMenuOpen}>
							<Sheet.Trigger>
								<Button variant="ghost" size="icon" class="md:hidden">
									<Menu class="h-5 w-5" />
									<span class="sr-only">Toggle menu</span>
								</Button>
							</Sheet.Trigger>
							<Sheet.Content side="left">
								<Sheet.Header>
									<Sheet.Title>Navigation</Sheet.Title>
								</Sheet.Header>
								<div class="mt-6 flex flex-col space-y-3">
									<Button
										variant="ghost"
										href="/"
										onclick={() => (mobileMenuOpen = false)}
										class="justify-start"
									>
										Dashboard
									</Button>
									<Button
										variant="ghost"
										href="/repositories"
										onclick={() => (mobileMenuOpen = false)}
										class="justify-start"
									>
										Repositories
									</Button>
									<Button
										variant="ghost"
										href="/search"
										onclick={() => (mobileMenuOpen = false)}
										class="justify-start"
									>
										Search
									</Button>
									<Button
										variant="ghost"
										href="/chat"
										onclick={() => (mobileMenuOpen = false)}
										class="justify-start"
									>
										Chat
									</Button>
								</div>
							</Sheet.Content>
						</Sheet.Root>

						<span class="hidden text-sm text-muted-foreground sm:block">
							Welcome, {authState?.user?.name}
						</span>
						<Button
							variant="outline"
							size="sm"
							onclick={() => {
								authStore.logout();
							}}
						>
							Logout
						</Button>
					{:else if !authState?.isLoading && $page.url.pathname !== '/auth/login'}
						<Button href="/auth/login">Login</Button>
					{/if}
				</div>
			</div>
		</div>
	</header>

	<main class="container mx-auto px-4 py-8 sm:px-6 lg:px-8">
		{#if authState?.isLoading}
			<div class="mx-auto max-w-4xl space-y-4">
				<Skeleton class="h-12 w-full" />
				<div class="grid grid-cols-1 gap-6 md:grid-cols-3">
					<Skeleton class="h-32 w-full" />
					<Skeleton class="h-32 w-full" />
					<Skeleton class="h-32 w-full" />
				</div>
				<Skeleton class="h-64 w-full" />
			</div>
		{:else}
			{@render children()}
		{/if}
	</main>
</div>

<Toaster />
