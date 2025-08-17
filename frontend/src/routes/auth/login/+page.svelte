<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/stores/auth';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import { Checkbox } from '$lib/components/ui/checkbox/index.js';
	import * as Alert from '$lib/components/ui/alert/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Info } from '@lucide/svelte';
	let email = $state('');
	let password = $state('');
	let error = $state('');

	function useDemoCredentials() {
		email = 'demo@acip.com';
		password = 'demo123456';
	}

	// Redirect if already logged in
	onMount(() => {});

	async function handleLogin(event: Event) {
		event.preventDefault();
		if (!email || !password) {
			error = 'Please fill in all fields';
			return;
		}

		error = '';

		try {
			await authStore.login(email, password);
			// Redirect to dashboard on success
			goto('/');
		} catch (e) {
			error = e instanceof Error ? e.message : 'Login failed. Please try again.';
		}
	}
</script>

<svelte:head>
	<title>Login - ACIP</title>
</svelte:head>

<div class="mx-auto mt-16 max-w-md">
	<Card.Root>
		<Card.Header class="text-center">
			<Card.Title class="text-2xl">Sign in to your account</Card.Title>
		</Card.Header>

		<Card.Content class="space-y-6">
			<Alert.Root class="mb-6">
				<Info class="h-4 w-4" />
				<Alert.Title>Demo Access Available</Alert.Title>
				<Alert.Description>
					<p class="mb-2">Use these credentials to explore the AI-powered code analyzer:</p>
					<div class="mt-2 rounded border bg-background p-3 font-mono text-xs">
						<div><strong>Email:</strong> demo@acip.com</div>
						<div><strong>Password:</strong> demo123456</div>
					</div>
					<div class="mt-3">
						<Button variant="link" size="sm" onclick={useDemoCredentials} class="h-auto p-0">
							Fill demo credentials →
						</Button>
					</div>
				</Alert.Description>
			</Alert.Root>

			<form onsubmit={handleLogin} class="space-y-6">
				{#if error}
					<Alert.Root variant="destructive">
						<Alert.Description>{error}</Alert.Description>
					</Alert.Root>
				{/if}

				<div class="space-y-2">
					<Label for="email">Email address</Label>
					<Input
						id="email"
						type="email"
						bind:value={email}
						required
						placeholder="Enter your email"
					/>
				</div>

				<div class="space-y-2">
					<Label for="password">Password</Label>
					<Input
						id="password"
						type="password"
						bind:value={password}
						required
						placeholder="Enter your password"
					/>
				</div>

				<div class="flex items-center space-x-2">
					<Checkbox id="remember-me" />
					<Label for="remember-me" class="text-sm font-normal">Remember me</Label>
				</div>

				<Button type="submit" disabled={authStore.current.isLoading} class="w-full">
					{authStore.current.isLoading ? 'Signing in...' : 'Sign in'}
				</Button>

				<div class="text-center">
					<p class="text-sm text-muted-foreground">
						Experience the power of AI-powered code analysis with the demo account above.
					</p>
				</div>
			</form>
		</Card.Content>
	</Card.Root>
</div>
