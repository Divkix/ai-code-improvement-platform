<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/stores/auth';
	import { requireGuest } from '$lib/auth';

	let email = '';
	let password = '';
	let error = '';

	// Redirect if already logged in
	onMount(() => {
		requireGuest();
	});

	async function handleLogin() {
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

		<form on:submit|preventDefault={handleLogin} class="space-y-6">
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

			<div class="flex items-center justify-between">
				<div class="flex items-center">
					<input
						id="remember-me"
						type="checkbox"
						class="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
					/>
					<label for="remember-me" class="ml-2 block text-sm text-gray-900"> Remember me </label>
				</div>

				<div class="text-sm">
					<a href="/auth/forgot-password" class="font-medium text-blue-600 hover:text-blue-500">
						Forgot your password?
					</a>
				</div>
			</div>

			<button
				type="submit"
				disabled={$authStore.isLoading}
				class="flex w-full justify-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700 focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 focus:outline-none disabled:cursor-not-allowed disabled:opacity-50"
			>
				{$authStore.isLoading ? 'Signing in...' : 'Sign in'}
			</button>

			<div class="text-center">
				<span class="text-sm text-gray-500">
					For demo access, contact your administrator
				</span>
			</div>
		</form>
	</div>
</div>
