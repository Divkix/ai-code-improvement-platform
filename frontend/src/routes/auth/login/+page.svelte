<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/stores/auth';
	let email = $state('');
	let password = $state('');
	let error = $state('');

	function useDemoCredentials() {
		email = 'demo@github-analyzer.com';
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
	<title>Login - GitHub Analyzer</title>
</svelte:head>

<div class="mx-auto mt-16 max-w-md">
	<div class="rounded-lg bg-white px-6 py-8 shadow">
		<div class="text-center">
			<h2 class="mb-6 text-2xl font-bold text-gray-900">Sign in to your account</h2>
		</div>

		<div class="mb-6 rounded-lg border border-blue-200 bg-blue-50 p-4">
			<div class="flex items-start">
				<div class="flex-shrink-0">
					<svg class="h-5 w-5 text-blue-400" fill="currentColor" viewBox="0 0 20 20">
						<path
							fill-rule="evenodd"
							d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z"
							clip-rule="evenodd"
						></path>
					</svg>
				</div>
				<div class="ml-3 flex-1">
					<h3 class="text-sm font-medium text-blue-800">Demo Access Available</h3>
					<div class="mt-2 text-sm text-blue-700">
						<p class="mb-2">Use these credentials to explore the AI-powered code analyzer:</p>
						<div class="rounded border border-blue-300 bg-white p-3 font-mono text-xs">
							<div><strong>Email:</strong> demo@github-analyzer.com</div>
							<div><strong>Password:</strong> demo123456</div>
						</div>
					</div>
					<div class="mt-3">
						<button
							type="button"
							onclick={useDemoCredentials}
							class="text-sm font-medium text-blue-600 hover:text-blue-500"
						>
							Fill demo credentials →
						</button>
					</div>
				</div>
			</div>
		</div>

		<form onsubmit={handleLogin} class="space-y-6">
			{#if error}
				<div class="rounded border border-red-200 bg-red-50 px-4 py-3 text-red-700">
					{error}
				</div>
			{/if}

			<div>
				<label for="email" class="mb-2 block text-sm font-medium text-gray-700">
					Email address
				</label>
				<input
					id="email"
					type="email"
					bind:value={email}
					required
					class="w-full rounded-md border border-gray-300 px-3 py-2 shadow-sm focus:border-blue-500 focus:ring-blue-500 focus:outline-none"
					placeholder="Enter your email"
				/>
			</div>

			<div>
				<label for="password" class="mb-2 block text-sm font-medium text-gray-700">
					Password
				</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					required
					class="w-full rounded-md border border-gray-300 px-3 py-2 shadow-sm focus:border-blue-500 focus:ring-blue-500 focus:outline-none"
					placeholder="Enter your password"
				/>
			</div>

			<div class="flex items-center">
				<input
					id="remember-me"
					type="checkbox"
					class="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
				/>
				<label for="remember-me" class="ml-2 block text-sm text-gray-900"> Remember me </label>
			</div>

			<button
				type="submit"
				disabled={$authStore.isLoading}
				class="flex w-full justify-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none disabled:cursor-not-allowed disabled:opacity-50"
			>
				{$authStore.isLoading ? 'Signing in...' : 'Sign in'}
			</button>

			<div class="text-center">
				<p class="text-sm text-gray-500">
					Experience the power of AI-powered code analysis with the demo account above.
				</p>
			</div>
		</form>
	</div>
</div>
