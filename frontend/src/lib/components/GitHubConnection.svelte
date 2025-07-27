<!-- ABOUTME: GitHub OAuth connection management with secure authentication flow -->
<!-- ABOUTME: Handles GitHub account linking, disconnection, and OAuth callbacks -->

<script lang="ts">
	import { onMount } from 'svelte';
	import { replaceState } from '$app/navigation';
	import {
		handleGitHubCallback,
		disconnectGitHub as disconnectGitHubAPI,
		getGitHubLoginUrl
	} from '$lib/api/hooks';
	import { type User } from '$lib/api';
	import { authStore } from '$lib/stores/auth';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import * as Alert from '$lib/components/ui/alert';
	import { Github, Unlink, Loader2, Check, AlertCircle } from '@lucide/svelte';
	import * as AlertDialog from '$lib/components/ui/alert-dialog';
	import { toast } from 'svelte-sonner';

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
				toast.success('GitHub account connected successfully!');

				// Clean up the URL
				replaceState(window.location.pathname, {});
			} catch (err) {
				error = err instanceof Error ? err.message : 'Failed to connect GitHub account';
				toast.error(error);
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
			toast.error(error);
			connecting = false;
		}
	}

	let showDisconnectDialog = $state(false);

	async function disconnectGitHub() {
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
			toast.success('GitHub account disconnected successfully.');
			showDisconnectDialog = false;
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to disconnect GitHub account';
			toast.error(error);
		} finally {
			connecting = false;
		}
	}

	function clearMessages() {
		error = '';
		success = '';
	}
</script>

<Card.Root>
	<Card.Header>
		<Card.Title class="flex items-center gap-2">
			<Github class="h-5 w-5" />
			GitHub Connection
		</Card.Title>
		<Card.Description>
			Connect your GitHub account to import and analyze your repositories
		</Card.Description>
	</Card.Header>
	<Card.Content class="space-y-4">
		<div class="flex items-center justify-between">
			{#if user.githubConnected}
				<div class="flex items-center gap-2">
					<div class="h-2 w-2 rounded-full bg-green-500"></div>
					<span class="text-sm">Connected as <strong>{user.githubUsername}</strong></span>
				</div>
				<AlertDialog.Root bind:open={showDisconnectDialog}>
					<AlertDialog.Trigger>
						<Button
							variant="outline"
							size="sm"
							disabled={connecting}
							class="text-destructive hover:text-destructive"
						>
							{#if connecting}
								<Loader2 class="mr-2 h-4 w-4 animate-spin" />
								Disconnecting...
							{:else}
								<Unlink class="mr-2 h-4 w-4" />
								Disconnect
							{/if}
						</Button>
					</AlertDialog.Trigger>
					<AlertDialog.Content>
						<AlertDialog.Header>
							<AlertDialog.Title>Disconnect GitHub Account</AlertDialog.Title>
							<AlertDialog.Description>
								Are you sure you want to disconnect your GitHub account? You will lose access to
								your GitHub repositories and any imported data.
							</AlertDialog.Description>
						</AlertDialog.Header>
						<AlertDialog.Footer>
							<AlertDialog.Cancel>Cancel</AlertDialog.Cancel>
							<AlertDialog.Action
								onclick={disconnectGitHub}
								class="text-destructive-foreground bg-destructive hover:bg-destructive/90"
							>
								Disconnect
							</AlertDialog.Action>
						</AlertDialog.Footer>
					</AlertDialog.Content>
				</AlertDialog.Root>
			{:else}
				<div class="flex items-center gap-2">
					<div class="h-2 w-2 rounded-full bg-muted-foreground"></div>
					<span class="text-sm text-muted-foreground">Not connected</span>
				</div>
				<Button onclick={connectGitHub} disabled={connecting}>
					{#if connecting}
						<Loader2 class="mr-2 h-4 w-4 animate-spin" />
						Connecting...
					{:else}
						<Github class="mr-2 h-4 w-4" />
						Connect GitHub
					{/if}
				</Button>
			{/if}
		</div>

		{#if error}
			<Alert.Root variant="destructive">
				<AlertCircle class="h-4 w-4" />
				<Alert.Title>Connection Error</Alert.Title>
				<Alert.Description>
					{error}
					<Button
						variant="ghost"
						size="sm"
						onclick={clearMessages}
						class="mt-2 h-auto p-0 text-destructive underline hover:text-destructive"
					>
						Dismiss
					</Button>
				</Alert.Description>
			</Alert.Root>
		{/if}

		{#if success}
			<Alert.Root>
				<Check class="h-4 w-4" />
				<Alert.Title>Success</Alert.Title>
				<Alert.Description>
					{success}
					<Button
						variant="ghost"
						size="sm"
						onclick={clearMessages}
						class="mt-2 h-auto p-0 text-muted-foreground underline hover:text-foreground"
					>
						Dismiss
					</Button>
				</Alert.Description>
			</Alert.Root>
		{/if}

		{#if user.githubConnected}
			<div class="rounded-md bg-muted/50 p-3 text-sm text-muted-foreground">
				<p>Your GitHub account is connected and ready to import repositories.</p>
			</div>
		{:else}
			<div class="space-y-2 text-sm text-muted-foreground">
				<p>Connect your GitHub account to:</p>
				<ul class="ml-4 space-y-1">
					<li class="flex items-start gap-2">
						<div class="mt-2 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-muted-foreground"></div>
						Import repositories directly from GitHub
					</li>
					<li class="flex items-start gap-2">
						<div class="mt-2 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-muted-foreground"></div>
						Access private repositories you own
					</li>
					<li class="flex items-start gap-2">
						<div class="mt-2 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-muted-foreground"></div>
						Get real-time repository statistics
					</li>
					<li class="flex items-start gap-2">
						<div class="mt-2 h-1.5 w-1.5 flex-shrink-0 rounded-full bg-muted-foreground"></div>
						Automatically sync repository changes
					</li>
				</ul>
			</div>
		{/if}
	</Card.Content>
</Card.Root>
