<!-- ABOUTME: Repository embedding status display with progress tracking and controls -->
<!-- ABOUTME: Shows embedding progress, statistics, and provides re-embedding triggers -->

<script lang="ts">
	import { createEventDispatcher, onMount } from 'svelte';
	import { vectorSearchAPI } from '../api/client';

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
				return '✅';
			case 'processing':
				return '⚡';
			case 'failed':
				return '❌';
			case 'pending':
			default:
				return '⏳';
		}
	}

	function getStatusColor(status: string) {
		switch (status) {
			case 'completed':
				return '#10b981'; // green
			case 'processing':
				return '#3b82f6'; // blue
			case 'failed':
				return '#ef4444'; // red
			case 'pending':
			default:
				return '#f59e0b'; // amber
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

<div class="embedding-status">
	<div class="status-header">
		<h3>Vector Embeddings</h3>
		{#if !loading}
			<button
				class="refresh-button"
				on:click={fetchStatus}
				title="Refresh status"
				aria-label="Refresh embedding status"
			>
				<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
					<polyline points="23,4 23,10 17,10"/>
					<polyline points="1,20 1,14 7,14"/>
					<path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10"/>
					<path d="M3.51 15a9 9 0 0 0 14.85 3.36L23 14"/>
				</svg>
			</button>
		{/if}
	</div>

	{#if loading && !status}
		<div class="loading-state">
			<div class="spinner"></div>
			<p>Loading embedding status...</p>
		</div>
	{:else if error}
		<div class="error-state">
			<p class="error-message">{error}</p>
			<button class="retry-button" on:click={fetchStatus}>
				Try Again
			</button>
		</div>
	{:else if status}
		<div class="status-content">
			<!-- Status Overview -->
			<div class="status-overview">
				<div class="status-badge" style="background-color: {getStatusColor(status.status)}">
					<span class="status-icon">{getStatusIcon(status.status)}</span>
					<span class="status-text">{status.status.charAt(0).toUpperCase() + status.status.slice(1)}</span>
				</div>
				
				{#if status.status === 'processing'}
					<div class="progress-info">
						<span class="progress-text">{status.progress}% complete</span>
						{#if status.estimatedTimeRemaining}
							<span class="time-remaining">
								ETA: {formatDuration(status.estimatedTimeRemaining)}
							</span>
						{/if}
					</div>
				{/if}
			</div>

			<!-- Progress Bar -->
			{#if status.status === 'processing' || status.status === 'completed'}
				<div class="progress-bar">
					<div 
						class="progress-fill" 
						class:completed={status.status === 'completed'}
						style="width: {status.progress}%"
					></div>
				</div>
			{/if}

			<!-- Statistics -->
			{#if status.totalChunks || status.processedChunks}
				<div class="statistics">
					<div class="stat-grid">
						{#if status.totalChunks}
							<div class="stat-item">
								<span class="stat-label">Total Chunks</span>
								<span class="stat-value">{status.totalChunks.toLocaleString()}</span>
							</div>
						{/if}
						{#if status.processedChunks}
							<div class="stat-item">
								<span class="stat-label">Processed</span>
								<span class="stat-value">{status.processedChunks.toLocaleString()}</span>
							</div>
						{/if}
						{#if status.failedChunks && status.failedChunks > 0}
							<div class="stat-item">
								<span class="stat-label">Failed</span>
								<span class="stat-value error">{status.failedChunks.toLocaleString()}</span>
							</div>
						{/if}
					</div>
				</div>
			{/if}

			<!-- Timestamps -->
			{#if status.startedAt || status.completedAt}
				<div class="timestamps">
					{#if status.startedAt}
						<div class="timestamp">
							<span class="timestamp-label">Started:</span>
							<span class="timestamp-value">{formatTimestamp(status.startedAt)}</span>
						</div>
					{/if}
					{#if status.completedAt}
						<div class="timestamp">
							<span class="timestamp-label">Completed:</span>
							<span class="timestamp-value">{formatTimestamp(status.completedAt)}</span>
						</div>
					{/if}
				</div>
			{/if}

			<!-- Action Buttons -->
			<div class="actions">
				{#if status.status === 'pending' || status.status === 'failed'}
					<button 
						class="action-button primary"
						on:click={triggerEmbedding}
						disabled={loading}
					>
						{#if loading}
							<div class="button-spinner"></div>
						{:else}
							<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<polygon points="5,3 19,12 5,21 5,3"/>
							</svg>
						{/if}
						{status.status === 'failed' ? 'Retry Embedding' : 'Start Embedding'}
					</button>
				{:else if status.status === 'completed'}
					<button 
						class="action-button secondary"
						on:click={triggerEmbedding}
						disabled={loading}
					>
						{#if loading}
							<div class="button-spinner"></div>
						{:else}
							<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
								<polyline points="23,4 23,10 17,10"/>
								<polyline points="1,20 1,14 7,14"/>
								<path d="M20.49 9A9 9 0 0 0 5.64 5.64L1 10"/>
								<path d="M3.51 15a9 9 0 0 0 14.85 3.36L23 14"/>
							</svg>
						{/if}
						Re-embed Repository
					</button>
				{/if}
			</div>
		</div>
	{:else}
		<div class="empty-state">
			<p>No embedding status available</p>
			<button class="action-button primary" on:click={triggerEmbedding} disabled={loading}>
				Start Embedding
			</button>
		</div>
	{/if}
</div>

<style>
	.embedding-status {
		background: white;
		border: 1px solid #e5e7eb;
		border-radius: 8px;
		padding: 20px;
		font-family: system-ui, -apple-system, sans-serif;
	}

	.status-header {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 16px;
	}

	.status-header h3 {
		margin: 0;
		font-size: 18px;
		font-weight: 600;
		color: #1f2937;
	}

	.refresh-button {
		background: none;
		border: 1px solid #d1d5db;
		border-radius: 6px;
		padding: 6px;
		cursor: pointer;
		color: #6b7280;
		transition: all 0.2s;
	}

	.refresh-button:hover {
		background: #f9fafb;
		border-color: #9ca3af;
		color: #374151;
	}

	.loading-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 32px;
		gap: 12px;
		color: #6b7280;
	}

	.error-state {
		text-align: center;
		padding: 24px;
	}

	.error-message {
		color: #ef4444;
		margin: 0 0 12px 0;
	}

	.status-overview {
		display: flex;
		justify-content: space-between;
		align-items: center;
		margin-bottom: 16px;
	}

	.status-badge {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		padding: 6px 12px;
		border-radius: 6px;
		color: white;
		font-weight: 500;
		font-size: 14px;
	}

	.progress-info {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 4px;
		font-size: 12px;
		color: #6b7280;
	}

	.progress-bar {
		background: #f3f4f6;
		border-radius: 4px;
		height: 8px;
		margin-bottom: 16px;
		overflow: hidden;
	}

	.progress-fill {
		height: 100%;
		background: #3b82f6;
		transition: width 0.3s ease;
		border-radius: 4px;
	}

	.progress-fill.completed {
		background: #10b981;
	}

	.statistics {
		margin-bottom: 16px;
	}

	.stat-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
		gap: 16px;
	}

	.stat-item {
		display: flex;
		flex-direction: column;
		gap: 4px;
	}

	.stat-label {
		font-size: 12px;
		color: #6b7280;
		font-weight: 500;
	}

	.stat-value {
		font-size: 18px;
		font-weight: 600;
		color: #1f2937;
	}

	.stat-value.error {
		color: #ef4444;
	}

	.timestamps {
		margin-bottom: 16px;
		padding: 12px;
		background: #f9fafb;
		border-radius: 6px;
		font-size: 12px;
	}

	.timestamp {
		display: flex;
		justify-content: space-between;
		margin-bottom: 4px;
	}

	.timestamp:last-child {
		margin-bottom: 0;
	}

	.timestamp-label {
		color: #6b7280;
		font-weight: 500;
	}

	.timestamp-value {
		color: #374151;
	}

	.actions {
		display: flex;
		justify-content: center;
	}

	.action-button {
		display: inline-flex;
		align-items: center;
		gap: 8px;
		padding: 10px 16px;
		border-radius: 6px;
		font-weight: 500;
		cursor: pointer;
		transition: all 0.2s;
		border: 1px solid transparent;
	}

	.action-button.primary {
		background: #3b82f6;
		color: white;
	}

	.action-button.primary:hover {
		background: #2563eb;
	}

	.action-button.secondary {
		background: white;
		color: #374151;
		border-color: #d1d5db;
	}

	.action-button.secondary:hover {
		background: #f9fafb;
		border-color: #9ca3af;
	}

	.action-button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
	}

	.retry-button {
		background: #3b82f6;
		color: white;
		border: none;
		padding: 8px 16px;
		border-radius: 6px;
		font-weight: 500;
		cursor: pointer;
		transition: background-color 0.2s;
	}

	.retry-button:hover {
		background: #2563eb;
	}

	.spinner,
	.button-spinner {
		width: 16px;
		height: 16px;
		border: 2px solid #e5e7eb;
		border-top: 2px solid #3b82f6;
		border-radius: 50%;
		animation: spin 1s linear infinite;
	}

	.button-spinner {
		border-top-color: currentColor;
		border-color: rgba(255, 255, 255, 0.3);
	}

	@keyframes spin {
		0% { transform: rotate(0deg); }
		100% { transform: rotate(360deg); }
	}

	.empty-state {
		text-align: center;
		padding: 32px;
		color: #6b7280;
	}

	.empty-state p {
		margin: 0 0 16px 0;
	}

	/* Responsive adjustments */
	@media (max-width: 640px) {
		.embedding-status {
			padding: 16px;
		}

		.status-overview {
			flex-direction: column;
			align-items: flex-start;
			gap: 8px;
		}

		.progress-info {
			align-items: flex-start;
		}

		.stat-grid {
			grid-template-columns: 1fr;
			gap: 12px;
		}

		.timestamp {
			flex-direction: column;
			gap: 2px;
		}
	}
</style>