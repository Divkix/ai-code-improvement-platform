<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Chart, registerables } from 'chart.js';
	import { apiClient } from '$lib/api';
	import { authStore } from '$lib/stores/auth';
	import type { DashboardStats, ActivityItem, TrendDataPoint } from '$lib/api';

	Chart.register(...registerables);

	let stats: DashboardStats | null = null;
	let activities: ActivityItem[] = [];
	let trends: TrendDataPoint[] = [];
	let loading = true;
	let error: string | null = null;
	let chartCanvas: HTMLCanvasElement;
	let chart: Chart | null = null;

	async function loadDashboardData() {
		try {
			loading = true;
			error = null;

			// Check authentication first
			if (!$authStore.isAuthenticated) {
				goto('/auth/login');
				return;
			}

			const [statsData, activitiesData, trendsData] = await Promise.all([
				apiClient.getDashboardStats(),
				apiClient.getDashboardActivity(6), // Get 6 recent activities
				apiClient.getDashboardTrends(14) // Get 2 weeks of trend data
			]);

			stats = statsData;
			activities = activitiesData;
			trends = trendsData;

			// Create chart after data is loaded
			if (chartCanvas && trends.length > 0) {
				createChart();
			}
		} catch (err) {
			console.error('Failed to load dashboard data:', err);
			error = err instanceof Error ? err.message : 'Failed to load dashboard data';
			
			// If it's an auth error, redirect to login
			if (err instanceof Error && (err.message.includes('authorization') || err.message.includes('Unauthorized'))) {
				authStore.logout();
				goto('/auth/login');
			}
		} finally {
			loading = false;
		}
	}

	function createChart() {
		if (chart) {
			chart.destroy();
		}

		const ctx = chartCanvas.getContext('2d');
		if (!ctx) return;

		chart = new Chart(ctx, {
			type: 'line',
			data: {
				labels: trends.map((t) => new Date(t.date).toLocaleDateString()),
				datasets: [
					{
						label: 'Code Quality',
						data: trends.map((t) => t.codeQuality),
						borderColor: '#10b981',
						backgroundColor: '#10b981',
						borderWidth: 2,
						fill: false,
						tension: 0.4
					},
					{
						label: 'Performance Score',
						data: trends.map((t) => t.performanceScore),
						borderColor: '#3b82f6',
						backgroundColor: '#3b82f6',
						borderWidth: 2,
						fill: false,
						tension: 0.4
					}
				]
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				plugins: {
					legend: {
						position: 'bottom'
					}
				},
				scales: {
					y: {
						beginAtZero: false,
						min: 50,
						max: 100
					}
				}
			}
		});
	}

	function getSeverityColor(severity: string): string {
		switch (severity) {
			case 'error':
				return 'bg-red-400';
			case 'warning':
				return 'bg-yellow-400';
			case 'success':
				return 'bg-green-400';
			default:
				return 'bg-blue-400';
		}
	}

	function getTypeIcon(type: string): string {
		switch (type) {
			case 'repository_imported':
				return '📁';
			case 'analysis_completed':
				return '✅';
			case 'issue_detected':
				return '⚠️';
			case 'optimization_found':
				return '⚡';
			default:
				return '📊';
		}
	}

	function formatTimeAgo(timestamp: string): string {
		const date = new Date(timestamp);
		const now = new Date();
		const diffInMinutes = Math.floor((now.getTime() - date.getTime()) / (1000 * 60));

		if (diffInMinutes < 60) {
			return `${diffInMinutes}m ago`;
		} else if (diffInMinutes < 1440) {
			return `${Math.floor(diffInMinutes / 60)}h ago`;
		} else {
			return `${Math.floor(diffInMinutes / 1440)}d ago`;
		}
	}

	onMount(() => {
		// Wait for auth initialization before loading data
		const unsubscribe = authStore.subscribe((auth) => {
			if (!auth.isLoading) {
				if (auth.isAuthenticated) {
					loadDashboardData();
				} else {
					goto('/auth/login');
				}
				unsubscribe(); // Only run once
			}
		});
	});
</script>

<svelte:head>
	<title>Dashboard - GitHub Analyzer</title>
</svelte:head>

{#if $authStore.isLoading}
	<div class="flex h-96 items-center justify-center">
		<div class="text-center">
			<div
				class="inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-blue-600 border-r-transparent"
			></div>
			<p class="mt-4 text-gray-600">Initializing...</p>
		</div>
	</div>
{:else if !$authStore.isAuthenticated}
	<div class="flex h-96 items-center justify-center">
		<div class="text-center">
			<div class="mx-auto mb-4 h-12 w-12 rounded-full bg-blue-100 p-3">
				<svg class="h-6 w-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"></path>
				</svg>
			</div>
			<h3 class="text-lg font-medium text-gray-900 mb-2">Authentication Required</h3>
			<p class="text-gray-600 mb-4">Please log in to access your dashboard.</p>
			<a
				href="/auth/login"
				class="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700"
			>
				Go to Login
			</a>
		</div>
	</div>
{:else if loading}
	<div class="flex h-96 items-center justify-center">
		<div class="text-center">
			<div
				class="inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-blue-600 border-r-transparent"
			></div>
			<p class="mt-4 text-gray-600">Loading dashboard data...</p>
		</div>
	</div>
{:else if error}
	<div class="rounded-md bg-red-50 p-4">
		<div class="flex">
			<div class="ml-3">
				<h3 class="text-sm font-medium text-red-800">Error loading dashboard</h3>
				<p class="mt-2 text-sm text-red-700">{error}</p>
				<button
					on:click={loadDashboardData}
					class="mt-3 inline-flex items-center rounded-md bg-red-100 px-3 py-2 text-sm font-medium text-red-800 hover:bg-red-200"
				>
					Try again
				</button>
			</div>
		</div>
	</div>
{:else if stats}
	<div class="space-y-6">
		<!-- Hero Metrics -->
		<div class="overflow-hidden rounded-lg bg-gradient-to-r from-blue-600 to-purple-600 shadow-xl">
			<div class="p-8">
				<div class="grid grid-cols-1 gap-6 md:grid-cols-3 lg:grid-cols-6">
					<div class="text-center">
						<dt class="truncate text-sm font-medium text-blue-100">Repositories</dt>
						<dd class="mt-2 text-4xl font-bold text-white">
							{stats.totalRepositories}
						</dd>
					</div>
					<div class="text-center">
						<dt class="truncate text-sm font-medium text-blue-100">Code Chunks</dt>
						<dd class="mt-2 text-4xl font-bold text-white">
							{(stats.codeChunksProcessed / 1000).toFixed(1)}K
						</dd>
					</div>
					<div class="text-center">
						<dt class="truncate text-sm font-medium text-green-100">Monthly Savings</dt>
						<dd class="mt-2 text-4xl font-bold text-green-200">
							${Math.round(stats.costSavingsMonthly / 1000)}K
						</dd>
					</div>
					<div class="text-center">
						<dt class="truncate text-sm font-medium text-blue-100">Hours Reclaimed</dt>
						<dd class="mt-2 text-4xl font-bold text-white">
							{Math.round(stats.developerHoursReclaimed)}h
						</dd>
					</div>
					<div class="text-center">
						<dt class="truncate text-sm font-medium text-blue-100">Issues Prevented</dt>
						<dd class="mt-2 text-4xl font-bold text-yellow-200">
							{stats.issuesPreventedMonthly}
						</dd>
					</div>
					<div class="text-center">
						<dt class="truncate text-sm font-medium text-blue-100">Avg Response</dt>
						<dd class="mt-2 text-4xl font-bold text-white">
							{stats.avgResponseTime.toFixed(1)}s
						</dd>
					</div>
				</div>
			</div>
		</div>

		<!-- ROI Highlight -->
		<div class="rounded-lg border border-green-200 bg-green-50 p-6">
			<div class="flex items-center justify-between">
				<div class="flex items-center space-x-3">
					<div class="rounded-full bg-green-100 p-3">
						<svg
							class="h-6 w-6 text-green-600"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"
							></path>
						</svg>
					</div>
					<div>
						<h3 class="text-lg font-semibold text-green-900">Impressive Cost Savings</h3>
						<p class="text-sm text-green-700">
							Your team is saving an average of <strong
								>${Math.round(stats.costSavingsMonthly).toLocaleString()}</strong
							>
							per month by using AI-powered code analysis. That's
							<strong>{Math.round(stats.developerHoursReclaimed)} developer hours</strong> reclaimed
							for building features instead of understanding code.
						</p>
					</div>
				</div>
				<div class="text-right">
					<div class="text-3xl font-bold text-green-600">
						${Math.round((stats.costSavingsMonthly * 12) / 1000)}K
					</div>
					<div class="text-sm text-green-600">Annual Savings</div>
				</div>
			</div>
		</div>

		<!-- Content Grid -->
		<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
			<!-- Code Quality Trend Chart -->
			<div class="overflow-hidden rounded-lg bg-white shadow">
				<div class="p-6">
					<h3 class="mb-4 text-lg font-medium text-gray-900">Performance Trends (14 days)</h3>
					{#if trends.length > 0}
						<div class="h-64">
							<canvas bind:this={chartCanvas}></canvas>
						</div>
					{:else}
						<div class="flex h-64 items-center justify-center text-gray-500">
							<p>No trend data available</p>
						</div>
					{/if}
				</div>
			</div>

			<!-- Recent Activity -->
			<div class="overflow-hidden rounded-lg bg-white shadow">
				<div class="p-6">
					<div class="mb-4 flex items-center justify-between">
						<h3 class="text-lg font-medium text-gray-900">Recent Activity</h3>
						<span class="text-sm text-gray-500">{activities.length} items</span>
					</div>
					<div class="space-y-4">
						{#each activities as activity (activity.id)}
							<div class="flex items-start space-x-4">
								<div class="flex-shrink-0">
									<div
										class="flex h-8 w-8 items-center justify-center rounded-full {getSeverityColor(
											activity.severity
										)}"
									>
										<span class="text-sm text-white">{getTypeIcon(activity.type)}</span>
									</div>
								</div>
								<div class="min-w-0 flex-1">
									<p class="text-sm font-medium text-gray-900">{activity.message}</p>
									<div class="mt-1 flex items-center space-x-2 text-xs text-gray-500">
										{#if activity.repositoryName}
											<span class="font-medium">{activity.repositoryName}</span>
											<span>•</span>
										{/if}
										<span>{formatTimeAgo(activity.timestamp)}</span>
									</div>
								</div>
							</div>
						{/each}
						{#if activities.length === 0}
							<div class="py-8 text-center text-gray-500">
								<p>No recent activity</p>
							</div>
						{/if}
					</div>
				</div>
			</div>
		</div>

		<!-- Call to Action -->
		<div class="rounded-lg border border-indigo-200 bg-gradient-to-r from-indigo-50 to-blue-50 p-8">
			<div class="flex items-center justify-between">
				<div class="flex items-center space-x-4">
					<div class="rounded-full bg-indigo-100 p-3">
						<svg
							class="h-8 w-8 text-indigo-600"
							fill="none"
							stroke="currentColor"
							viewBox="0 0 24 24"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
							></path>
						</svg>
					</div>
					<div>
						<h3 class="text-xl font-semibold text-indigo-900">
							Ready to analyze a new repository?
						</h3>
						<p class="mt-1 text-sm text-indigo-700">
							Connect your GitHub repositories and unlock AI-powered insights to accelerate your
							team's productivity.
						</p>
					</div>
				</div>
				<div class="flex space-x-3">
					<a
						href="/repositories"
						class="inline-flex items-center rounded-lg border border-indigo-300 bg-white px-6 py-3 text-sm font-medium text-indigo-700 shadow-sm transition-colors hover:bg-indigo-50"
					>
						View Repositories
					</a>
					<a
						href="/repositories"
						class="inline-flex items-center rounded-lg bg-indigo-600 px-6 py-3 text-sm font-medium text-white shadow-sm transition-colors hover:bg-indigo-700"
					>
						<svg class="mr-2 -ml-1 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M12 4v16m8-8H4"
							></path>
						</svg>
						Import Repository
					</a>
				</div>
			</div>
		</div>
	</div>
{/if}
