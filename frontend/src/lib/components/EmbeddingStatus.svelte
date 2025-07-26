<!-- ABOUTME: Repository embedding status display with progress tracking and controls -->
<!-- ABOUTME: Shows embedding progress, statistics, and provides re-embedding triggers -->

<script lang="ts">
	import { createEventDispatcher, onMount } from 'svelte';
	import { vectorSearchAPI } from '../api/client';
	import * as Card from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Progress } from '$lib/components/ui/progress';
	import { Separator } from '$lib/components/ui/separator';
	import { Skeleton } from '$lib/components/ui/skeleton';
	import { RefreshCw, Play, Loader2, Check, X, AlertCircle, Clock, Zap } from '@lucide/svelte';

	export let repositoryId: string;
	export let autoRefresh = true;
	export let refreshInterval = 5000; // 5 seconds

	interface EmbeddingStatus {
		repositoryId: string;
		status: 'pending' | 'processing' | 'completed' | 'failed';
		progress: number;
		totalChunks?: number;
		processedChunks?: number;
		failedChunks?: number;
		startedAt?: string;
		completedAt?: string;
		estimatedTimeRemaining?: number;
	}

	let status: EmbeddingStatus | null = null;
	let loading = false;
	let error: string | null = null;
	let refreshTimeout: NodeJS.Timeout;

	const dispatch = createEventDispatcher<{
		statusChange: EmbeddingStatus;
		triggerEmbedding: void;
		error: string;
	}>();

	onMount(() => {
		fetchStatus();
		if (autoRefresh) {
			startAutoRefresh();
		}
		return () => {
			if (refreshTimeout) {
				clearTimeout(refreshTimeout);
			}
		};
	});

	function startAutoRefresh() {
		if (refreshTimeout) {
			clearTimeout(refreshTimeout);
		}
		refreshTimeout = setTimeout(() => {
			fetchStatus();
			if (autoRefresh && status?.status === 'processing') {
				startAutoRefresh();
			}
		}, refreshInterval);
	}

	async function fetchStatus() {
		if (!repositoryId) return;

		try {
			loading = true;
			error = null;
			const response = await vectorSearchAPI.getRepositoryEmbeddingStatus(repositoryId);
			status = response as EmbeddingStatus;
			dispatch('statusChange', status);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to fetch embedding status';
			dispatch('error', error);
		} finally {
			loading = false;
		}
	}

	async function triggerEmbedding() {
		if (!repositoryId) return;

		try {
			loading = true;
			error = null;
			await vectorSearchAPI.triggerRepositoryEmbedding(repositoryId);
			dispatch('triggerEmbedding');
			// Fetch status after triggering
			setTimeout(fetchStatus, 1000);
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to trigger embedding';
			dispatch('error', error);
		} finally {
			loading = false;
		}
	}

	function getStatusIcon(status: string) {
		switch (status) {
			case 'completed':
				return Check;
			case 'processing':
				return Zap;
			case 'failed':
				return X;
			case 'pending':
			default:
				return Clock;
		}
	}

	function getStatusVariant(status: string): 'default' | 'secondary' | 'destructive' | 'outline' {
		switch (status) {
			case 'completed':
				return 'default';
			case 'processing':
				return 'secondary';
			case 'failed':
				return 'destructive';
			case 'pending':
			default:
				return 'outline';
		}
	}

	function formatDuration(seconds: number): string {
		if (seconds < 60) {
			return `${Math.round(seconds)}s`;
		}
		const minutes = Math.floor(seconds / 60);
		const remainingSeconds = seconds % 60;
		if (minutes < 60) {
			return `${minutes}m ${Math.round(remainingSeconds)}s`;
		}
		const hours = Math.floor(minutes / 60);
		const remainingMinutes = minutes % 60;
		return `${hours}h ${remainingMinutes}m`;
	}

	function formatTimestamp(timestamp: string): string {
		return new Date(timestamp).toLocaleString();
	}
</script>

<Card.Root>
	<Card.Header>
		<Card.Title class="flex items-center justify-between">
			Vector Embeddings
			{#if !loading}
				<Button
					variant="outline"
					size="sm"
					onclick={fetchStatus}
					title="Refresh status"
					aria-label="Refresh embedding status"
				>
					<RefreshCw class="h-4 w-4" />
				</Button>
			{/if}
		</Card.Title>
	</Card.Header>
	<Card.Content class="space-y-4">
		{#if loading && !status}
			<div class="space-y-4">
				<div class="flex items-center justify-between">
					<Skeleton class="h-6 w-20" />
					<Skeleton class="h-4 w-16" />
				</div>
				<Skeleton class="h-2 w-full" />
				<div class="grid grid-cols-3 gap-4">
					<div class="space-y-2">
						<Skeleton class="h-3 w-16" />
						<Skeleton class="h-6 w-12" />
					</div>
					<div class="space-y-2">
						<Skeleton class="h-3 w-16" />
						<Skeleton class="h-6 w-12" />
					</div>
					<div class="space-y-2">
						<Skeleton class="h-3 w-16" />
						<Skeleton class="h-6 w-12" />
					</div>
				</div>
			</div>
		{:else if error}
			<div class="flex flex-col items-center space-y-4 py-6 text-center">
				<AlertCircle class="h-8 w-8 text-destructive" />
				<p class="text-destructive">{error}</p>
				<Button variant="outline" onclick={fetchStatus}>
					<RefreshCw class="mr-2 h-4 w-4" />
					Try Again
				</Button>
			</div>
		{:else if status}
			<!-- Status Overview -->
			<div class="flex items-center justify-between">
				<Badge variant={getStatusVariant(status.status)} class="gap-1">
					{@const StatusIcon = getStatusIcon(status.status)}
					<StatusIcon class="h-3 w-3" />
					{status.status.charAt(0).toUpperCase() + status.status.slice(1)}
				</Badge>

				{#if status.status === 'processing'}
					<div class="flex flex-col items-end space-y-1 text-sm text-muted-foreground">
						<span>{status.progress}% complete</span>
						{#if status.estimatedTimeRemaining}
							<span>ETA: {formatDuration(status.estimatedTimeRemaining)}</span>
						{/if}
					</div>
				{/if}
			</div>

			<!-- Progress Bar -->
			{#if status.status === 'processing' || status.status === 'completed'}
				<div class="space-y-2">
					<div class="flex justify-between text-sm">
						<span>Progress</span>
						<span>{status.processedChunks || 0} / {status.totalChunks || 0}</span>
					</div>
					<Progress value={status.progress} class="h-2" />
				</div>
			{/if}

			<!-- Statistics -->
			{#if status.totalChunks || status.processedChunks}
				<Separator />
				<div class="grid grid-cols-3 gap-4">
					{#if status.totalChunks}
						<div class="space-y-1">
							<p class="text-xs font-medium text-muted-foreground">Total Chunks</p>
							<p class="text-lg font-semibold">{status.totalChunks.toLocaleString()}</p>
						</div>
					{/if}
					{#if status.processedChunks}
						<div class="space-y-1">
							<p class="text-xs font-medium text-muted-foreground">Processed</p>
							<p class="text-lg font-semibold">{status.processedChunks.toLocaleString()}</p>
						</div>
					{/if}
					{#if status.failedChunks && status.failedChunks > 0}
						<div class="space-y-1">
							<p class="text-xs font-medium text-muted-foreground">Failed</p>
							<p class="text-lg font-semibold text-destructive">
								{status.failedChunks.toLocaleString()}
							</p>
						</div>
					{/if}
				</div>
			{/if}

			<!-- Timestamps -->
			{#if status.startedAt || status.completedAt}
				<Separator />
				<div class="space-y-2 rounded-md bg-muted/50 p-3 text-xs">
					{#if status.startedAt}
						<div class="flex justify-between">
							<span class="font-medium text-muted-foreground">Started:</span>
							<span>{formatTimestamp(status.startedAt)}</span>
						</div>
					{/if}
					{#if status.completedAt}
						<div class="flex justify-between">
							<span class="font-medium text-muted-foreground">Completed:</span>
							<span>{formatTimestamp(status.completedAt)}</span>
						</div>
					{/if}
				</div>
			{/if}

			<!-- Action Buttons -->
			<div class="flex justify-end">
				{#if status.status === 'pending' || status.status === 'failed'}
					<Button onclick={triggerEmbedding} disabled={loading}>
						{#if loading}
							<Loader2 class="mr-2 h-4 w-4 animate-spin" />
						{:else}
							<Play class="mr-2 h-4 w-4" />
						{/if}
						{status.status === 'failed' ? 'Retry Embedding' : 'Start Embedding'}
					</Button>
				{:else if status.status === 'completed'}
					<Button variant="outline" onclick={triggerEmbedding} disabled={loading}>
						{#if loading}
							<Loader2 class="mr-2 h-4 w-4 animate-spin" />
						{:else}
							<RefreshCw class="mr-2 h-4 w-4" />
						{/if}
						Re-embed Repository
					</Button>
				{/if}
			</div>
		{:else}
			<div class="flex flex-col items-center space-y-4 py-8 text-center text-muted-foreground">
				<Clock class="h-8 w-8" />
				<p>No embedding status available</p>
				<Button onclick={triggerEmbedding} disabled={loading}>
					<Play class="mr-2 h-4 w-4" />
					Start Embedding
				</Button>
			</div>
		{/if}
	</Card.Content>
</Card.Root>
