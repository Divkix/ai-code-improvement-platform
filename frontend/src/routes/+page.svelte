<script lang="ts">
	import { onMount } from 'svelte';
	import { Chart, registerables } from 'chart.js';
	import { getDashboardStats, getDashboardActivity, getDashboardTrends } from '$lib/api/hooks';
	import { authStore } from '$lib/stores/auth';
	import type { DashboardStats, ActivityItem, TrendDataPoint } from '$lib/api';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Alert from '$lib/components/ui/alert/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Skeleton } from '$lib/components/ui/skeleton/index.js';
	import { ChartContainer } from '$lib/components/ui/chart/index.js';
	import { TrendingUp } from '@lucide/svelte';

	Chart.register(...registerables);

	let stats = $state<DashboardStats | null>(null);

	// Format large numbers: <1k -> exact, 1k-999,999 -> #.#K, >=1M -> #.#M
	function formatMetric(n: number): string {
		if (n < 1000) return n.toLocaleString();
		if (n < 1_000_000) {
			const k = n / 1000;
			return `${k % 1 === 0 ? k.toFixed(0) : k.toFixed(1)}K`;
		}
		const m = n / 1_000_000;
		return `${m % 1 === 0 ? m.toFixed(0) : m.toFixed(1)}M`;
	}
	let activities = $state<ActivityItem[]>([]);
	let trends = $state<TrendDataPoint[]>([]);
	let loading = $state(true);
	let error = $state<string | null>(null);
	let chartCanvas = $state<HTMLCanvasElement | undefined>(undefined);
	let chart: Chart | null = null;

	async function loadDashboardData() {
		try {
			loading = true;
			error = null;

			const [statsData, activitiesData, trendsData] = await Promise.all([
				getDashboardStats(),
				getDashboardActivity(6),
				getDashboardTrends(14)
			]);

			console.log('Received stats data from API:', statsData);

			stats = statsData;
			activities = activitiesData;
			trends = trendsData;
		} catch (err) {
			console.error('Failed to load dashboard data:', err);
			error = err instanceof Error ? err.message : 'Failed to load dashboard data';

			if (
				err instanceof Error &&
				(err.message.includes('authorization') || err.message.includes('Unauthorized'))
			) {
				authStore.logout();
			}
		} finally {
			loading = false;
		}
	}

	function createChart() {
		if (chart) {
			chart.destroy();
		}
		if (!chartCanvas) return;

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
		loadDashboardData();
	});

	$effect(() => {
		if (chartCanvas && trends.length > 0) {
			createChart();
		}
	});
</script>

<svelte:head>
	<title>Dashboard - GitHub Analyzer</title>
</svelte:head>

{#if loading}
	<div class="space-y-6">
		<!-- Stats Cards Skeleton -->
		<div class="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4">
			<Card.Root>
				<Card.Content class="p-6">
					<div class="space-y-2">
						<Skeleton class="h-4 w-20" />
						<Skeleton class="h-8 w-16" />
					</div>
				</Card.Content>
			</Card.Root>
			<Card.Root>
				<Card.Content class="p-6">
					<div class="space-y-2">
						<Skeleton class="h-4 w-20" />
						<Skeleton class="h-8 w-16" />
					</div>
				</Card.Content>
			</Card.Root>
			<Card.Root>
				<Card.Content class="p-6">
					<div class="space-y-2">
						<Skeleton class="h-4 w-20" />
						<Skeleton class="h-8 w-16" />
					</div>
				</Card.Content>
			</Card.Root>
			<Card.Root>
				<Card.Content class="p-6">
					<div class="space-y-2">
						<Skeleton class="h-4 w-20" />
						<Skeleton class="h-8 w-16" />
					</div>
				</Card.Content>
			</Card.Root>
		</div>

		<!-- Chart and Activity Skeleton -->
		<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
			<Card.Root>
				<Card.Header>
					<Skeleton class="h-6 w-32" />
				</Card.Header>
				<Card.Content>
					<Skeleton class="h-64 w-full" />
				</Card.Content>
			</Card.Root>
			<Card.Root>
				<Card.Header>
					<Skeleton class="h-6 w-40" />
				</Card.Header>
				<Card.Content>
					<div class="space-y-4">
						<div class="space-y-2">
							<Skeleton class="h-4 w-full" />
							<Skeleton class="h-3 w-24" />
						</div>
						<div class="space-y-2">
							<Skeleton class="h-4 w-full" />
							<Skeleton class="h-3 w-24" />
						</div>
						<div class="space-y-2">
							<Skeleton class="h-4 w-full" />
							<Skeleton class="h-3 w-24" />
						</div>
					</div>
				</Card.Content>
			</Card.Root>
		</div>
	</div>
{:else if error}
	<Alert.Root variant="destructive">
		<Alert.Title>Error loading dashboard</Alert.Title>
		<Alert.Description>
			<p>{error}</p>
			<div class="mt-3">
				<Button variant="outline" size="sm" onclick={loadDashboardData}>Try again</Button>
			</div>
		</Alert.Description>
	</Alert.Root>
{:else if stats}
	<div class="space-y-6">
		<Card.Root class="overflow-hidden border-0 bg-gradient-to-r from-blue-600 to-purple-600">
			<Card.Content class="p-8">
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
							{formatMetric(stats.codeChunksProcessed)}
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
						<dt class="truncate text-sm font-medium text-blue-100">Knowledge Radius</dt>
						<dd class="mt-2 text-4xl font-bold text-white">
							{formatMetric(stats.knowledgeRadius ?? 0)}
						</dd>
					</div>
				</div>
			</Card.Content>
		</Card.Root>

		<Alert.Root class="border-green-200 bg-green-50">
			<TrendingUp class="h-4 w-4" />
			<Alert.Title>Impressive Cost Savings</Alert.Title>
			<Alert.Description>
				<div class="mt-2 flex items-center justify-between">
					<div>
						<p class="text-sm text-green-700">
							Your team is saving an average of <strong
								>${Math.round(stats.costSavingsMonthly).toLocaleString()}</strong
							>
							per month by using AI-powered code analysis. That's
							<strong>{Math.round(stats.developerHoursReclaimed)} developer hours</strong> reclaimed
							for building features instead of understanding code.
						</p>
					</div>
					<div class="text-right">
						<div class="text-3xl font-bold text-green-600">
							${Math.round((stats.costSavingsMonthly * 12) / 1000)}K
						</div>
						<div class="text-sm text-green-600">Annual Savings</div>
					</div>
				</div>
			</Alert.Description>
		</Alert.Root>

		<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
			<Card.Root>
				<Card.Header>
					<Card.Title>Performance Trends (14 days)</Card.Title>
				</Card.Header>
				<Card.Content>
					{#if trends.length > 0}
						<ChartContainer config={{}} class="h-64">
							<canvas bind:this={chartCanvas}></canvas>
						</ChartContainer>
					{:else}
						<div class="flex h-64 items-center justify-center text-muted-foreground">
							<p>No trend data available</p>
						</div>
					{/if}
				</Card.Content>
			</Card.Root>

			<Card.Root>
				<Card.Header>
					<div class="flex items-center justify-between">
						<Card.Title>Recent Activity</Card.Title>
						<span class="text-sm text-muted-foreground">{activities.length} items</span>
					</div>
				</Card.Header>
				<Card.Content>
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
									<p class="text-sm font-medium">{activity.message}</p>
									<div class="mt-1 flex items-center space-x-2 text-xs text-muted-foreground">
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
							<div class="py-8 text-center text-muted-foreground">
								<p>No recent activity</p>
							</div>
						{/if}
					</div>
				</Card.Content>
			</Card.Root>
		</div>

		<Card.Root class="border-indigo-200 bg-gradient-to-r from-indigo-50 to-blue-50">
			<Card.Content class="p-8">
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
						<Button
							variant="outline"
							href="/repositories"
							class="border-indigo-300 bg-white text-indigo-700 hover:bg-indigo-50"
						>
							View Repositories
						</Button>
						<Button href="/repositories" class="bg-indigo-600 hover:bg-indigo-700">
							<svg class="mr-2 -ml-1 h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M12 4v16m8-8H4"
								></path>
							</svg>
							Import Repository
						</Button>
					</div>
				</div>
			</Card.Content>
		</Card.Root>
	</div>
{:else}
	<div class="py-16 text-center">
		<div class="mx-auto h-12 w-12 text-gray-400">
			<svg fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M9 17v-2m3 2v-4m3 4v-6m2 10H7a2 2 0 01-2-2V7a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
				></path>
			</svg>
		</div>
		<h3 class="mt-2 text-sm font-medium text-gray-900">No Dashboard Data</h3>
		<p class="mt-1 text-sm text-gray-500">
			Could not retrieve dashboard statistics. Please check the browser console for errors.
		</p>
	</div>
{/if}
