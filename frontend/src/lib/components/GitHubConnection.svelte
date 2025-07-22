<script lang="ts">
	import { onMount } from 'svelte';
	import { handleGitHubCallback, disconnectGitHub as disconnectGitHubAPI, getGitHubLoginUrl } from '$lib/api/hooks';
	import { type User } from '$lib/api';
	import { authStore } from '$lib/stores/auth';

	let { user }: { user: User } = $props();

	let connecting = $state(false);
	let error = $state('');
	let success = $state('');

	onMount(() => {
		// Handle GitHub OAuth callback
		handleOAuthCallback();
	});

	async function handleOAuthCallback() {
		const urlParams = new URLSearchParams(window.location.search);
		const code = urlParams.get('code');
		const state = urlParams.get('state');

		if (code && state && !user.githubConnected) {
			try {
				connecting = true;
				error = '';

				// Handle the OAuth callback
				const updatedUser = await handleGitHubCallback({ code, state });

				// Update the auth store with the new user data
				authStore.setUser(updatedUser);
				user = {
					...updatedUser,
					githubConnected: updatedUser.githubConnected ?? false,
					githubUsername: updatedUser.githubUsername ?? undefined
				};

				success = 'GitHub account connected successfully!';

				// Clean up the URL
				window.history.replaceState({}, document.title, window.location.pathname);
			} catch (err) {
				error = err instanceof Error ? err.message : 'Failed to connect GitHub account';
			} finally {
				connecting = false;
			}
		}
	}

	async function connectGitHub() {
		try {
			connecting = true;
			error = '';
			success = '';

			// Get the GitHub OAuth URL from the backend API
			const redirectUri = `${window.location.origin}/repositories`;
			const { auth_url } = await getGitHubLoginUrl(redirectUri);

			// Redirect to the GitHub OAuth URL
			window.location.href = auth_url;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to initiate GitHub connection';
			connecting = false;
		}
	}

	async function disconnectGitHub() {
		if (
			!confirm(
				'Are you sure you want to disconnect your GitHub account? You will lose access to your GitHub repositories.'
			)
		) {
			return;
		}

		try {
			connecting = true;
			error = '';
			success = '';

			// Disconnect GitHub account
			const updatedUser = await disconnectGitHubAPI();

			// Update the auth store
			authStore.setUser(updatedUser);
			user = {
				...updatedUser,
				githubConnected: updatedUser.githubConnected ?? false,
				githubUsername: updatedUser.githubUsername ?? undefined
			};

			success = 'GitHub account disconnected successfully.';
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to disconnect GitHub account';
		} finally {
			connecting = false;
		}
	}

	function clearMessages() {
		error = '';
		success = '';
	}
</script>

<div class="rounded-lg border border-gray-200 bg-white p-6">
	<div class="flex items-center justify-between">
		<div>
			<h3 class="text-lg font-medium text-gray-900">GitHub Connection</h3>
			<p class="mt-1 text-sm text-gray-600">
				Connect your GitHub account to import and analyze your repositories
			</p>
		</div>

		<div class="flex items-center space-x-3">
			{#if user.githubConnected}
				<div class="flex items-center text-sm text-green-600">
					<svg class="mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M5 13l4 4L19 7"
						/>
					</svg>
					Connected as {user.githubUsername}
				</div>
				<button
					onclick={disconnectGitHub}
					disabled={connecting}
					class="rounded-md border border-red-300 bg-white px-4 py-2 text-sm font-medium text-red-700 hover:bg-red-50 disabled:cursor-not-allowed disabled:opacity-50"
				>
					{connecting ? 'Disconnecting...' : 'Disconnect'}
				</button>
			{:else}
				<div class="flex items-center text-sm text-gray-500">
					<svg class="mr-2 h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M6 18L18 6M6 6l12 12"
						/>
					</svg>
					Not connected
				</div>
				<button
					onclick={connectGitHub}
					disabled={connecting}
					class="inline-flex items-center rounded-md bg-gray-900 px-4 py-2 text-sm font-medium text-white hover:bg-gray-800 disabled:cursor-not-allowed disabled:opacity-50"
				>
					{#if connecting}
						<svg class="mr-2 h-4 w-4 animate-spin" fill="none" viewBox="0 0 24 24">
							<circle
								class="opacity-25"
								cx="12"
								cy="12"
								r="10"
								stroke="currentColor"
								stroke-width="4"
							></circle>
							<path
								class="opacity-75"
								fill="currentColor"
								d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
							></path>
						</svg>
						Connecting...
					{:else}
						<svg class="mr-2 h-4 w-4" fill="currentColor" viewBox="0 0 20 20">
							<path
								fill-rule="evenodd"
								d="M10 0C4.477 0 0 4.484 0 10.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0110 4.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.203 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.942.359.31.678.921.678 1.856 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0020 10.017C20 4.484 15.522 0 10 0z"
								clip-rule="evenodd"
							/>
						</svg>
						Connect GitHub
					{/if}
				</button>
			{/if}
		</div>
	</div>

	{#if error}
		<div class="mt-4 rounded-md bg-red-50 p-4">
			<div class="flex">
				<div class="flex-shrink-0">
					<svg class="h-5 w-5 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
						/>
					</svg>
				</div>
				<div class="ml-3">
					<h3 class="text-sm font-medium text-red-800">Connection Error</h3>
					<p class="mt-2 text-sm text-red-700">{error}</p>
					<div class="mt-4">
						<button
							onclick={clearMessages}
							class="rounded-md bg-red-100 px-2 py-1 text-sm font-medium text-red-800 hover:bg-red-200"
						>
							Dismiss
						</button>
					</div>
				</div>
			</div>
		</div>
	{/if}

	{#if success}
		<div class="mt-4 rounded-md bg-green-50 p-4">
			<div class="flex">
				<div class="flex-shrink-0">
					<svg class="h-5 w-5 text-green-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M5 13l4 4L19 7"
						/>
					</svg>
				</div>
				<div class="ml-3">
					<h3 class="text-sm font-medium text-green-800">Success</h3>
					<p class="mt-2 text-sm text-green-700">{success}</p>
					<div class="mt-4">
						<button
							onclick={clearMessages}
							class="rounded-md bg-green-100 px-2 py-1 text-sm font-medium text-green-800 hover:bg-green-200"
						>
							Dismiss
						</button>
					</div>
				</div>
			</div>
		</div>
	{/if}

	{#if user.githubConnected}
		<div class="mt-4 text-sm text-gray-600">
			<p>Your GitHub account is connected and ready to import repositories.</p>
		</div>
	{:else}
		<div class="mt-4 text-sm text-gray-600">
			<p>Connect your GitHub account to:</p>
			<ul class="mt-2 list-inside list-disc space-y-1">
				<li>Import repositories directly from GitHub</li>
				<li>Access private repositories you own</li>
				<li>Get real-time repository statistics</li>
				<li>Automatically sync repository changes</li>
			</ul>
		</div>
	{/if}
</div>
