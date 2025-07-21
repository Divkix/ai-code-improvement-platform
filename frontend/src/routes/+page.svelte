<script lang="ts">
	// Mock data for demo - this will be replaced with API calls later
	const stats = {
		totalLinesAnalyzed: 1250000,
		activeRepositories: 5,
		monthlySavings: 8250,
		issuesFoundWeek: 23
	};

	const trendData = [72, 75, 78, 83]; // Code quality scores
	const recentActivity = [
		{
			id: 1,
			type: 'optimization',
			severity: 'medium',
			message: 'Found 3 redundant database queries in UserService',
			repository: 'backend-api',
			timestamp: '2 hours ago'
		},
		{
			id: 2,
			type: 'security',
			severity: 'high',
			message: 'Detected potential SQL injection vulnerability',
			repository: 'web-app',
			timestamp: '4 hours ago'
		},
		{
			id: 3,
			type: 'performance',
			severity: 'low',
			message: 'Suggested caching optimization for API endpoints',
			repository: 'backend-api',
			timestamp: '1 day ago'
		}
	];
</script>

<svelte:head>
	<title>Dashboard - GitHub Analyzer</title>
</svelte:head>

<div class="space-y-6">
	<!-- Hero Metrics -->
	<div class="overflow-hidden rounded-lg bg-white shadow">
		<div class="p-5">
			<div class="grid grid-cols-1 gap-5 md:grid-cols-4">
				<div class="text-center">
					<dt class="truncate text-sm font-medium text-gray-500">Lines Analyzed</dt>
					<dd class="mt-1 text-3xl font-semibold text-gray-900">
						{(stats.totalLinesAnalyzed / 1000000).toFixed(1)}M
					</dd>
				</div>
				<div class="text-center">
					<dt class="truncate text-sm font-medium text-gray-500">Active Repositories</dt>
					<dd class="mt-1 text-3xl font-semibold text-gray-900">{stats.activeRepositories}</dd>
				</div>
				<div class="text-center">
					<dt class="truncate text-sm font-medium text-gray-500">Monthly Savings</dt>
					<dd class="mt-1 text-3xl font-semibold text-green-600">
						${stats.monthlySavings.toLocaleString()}
					</dd>
				</div>
				<div class="text-center">
					<dt class="truncate text-sm font-medium text-gray-500">Issues Found (7d)</dt>
					<dd class="mt-1 text-3xl font-semibold text-blue-600">{stats.issuesFoundWeek}</dd>
				</div>
			</div>
		</div>
	</div>

	<!-- Content Grid -->
	<div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
		<!-- Code Quality Trend -->
		<div class="overflow-hidden rounded-lg bg-white shadow">
			<div class="p-5">
				<h3 class="mb-4 text-lg font-medium text-gray-900">Code Quality Trend</h3>
				<div class="space-y-3">
					{#each trendData as score, index}
						<div class="flex items-center justify-between">
							<span class="text-sm text-gray-500">Week {index + 1}</span>
							<div class="flex items-center space-x-2">
								<div class="h-2 w-32 rounded-full bg-gray-200">
									<div
										class="h-2 rounded-full bg-green-600 transition-all duration-300"
										style="width: {score}%"
									></div>
								</div>
								<span class="text-sm font-medium text-gray-900">{score}%</span>
							</div>
						</div>
					{/each}
				</div>
			</div>
		</div>

		<!-- Recent Activity -->
		<div class="overflow-hidden rounded-lg bg-white shadow">
			<div class="p-5">
				<h3 class="mb-4 text-lg font-medium text-gray-900">Recent Activity</h3>
				<div class="space-y-4">
					{#each recentActivity as activity}
						<div class="flex items-start space-x-3">
							<div class="flex-shrink-0">
								{#if activity.severity === 'high'}
									<div class="mt-2 h-2 w-2 rounded-full bg-red-400"></div>
								{:else if activity.severity === 'medium'}
									<div class="mt-2 h-2 w-2 rounded-full bg-yellow-400"></div>
								{:else}
									<div class="mt-2 h-2 w-2 rounded-full bg-green-400"></div>
								{/if}
							</div>
							<div class="min-w-0 flex-1">
								<p class="text-sm text-gray-900">{activity.message}</p>
								<div class="mt-1 flex items-center space-x-2 text-xs text-gray-500">
									<span>{activity.repository}</span>
									<span>•</span>
									<span>{activity.timestamp}</span>
								</div>
							</div>
						</div>
					{/each}
				</div>
			</div>
		</div>
	</div>

	<!-- Call to Action -->
	<div class="rounded-lg bg-blue-50 p-6">
		<div class="flex items-center justify-between">
			<div>
				<h3 class="text-lg font-medium text-blue-900">Ready to analyze a new repository?</h3>
				<p class="mt-1 text-sm text-blue-700">
					Import your GitHub repositories and start getting intelligent insights about your
					codebase.
				</p>
			</div>
			<div class="flex space-x-3">
				<a
					href="/repositories"
					class="inline-flex items-center rounded-md border border-transparent bg-blue-200 px-4 py-2 text-sm font-medium text-blue-700 hover:bg-blue-300"
				>
					View Repositories
				</a>
				<a
					href="/repositories/import"
					class="inline-flex items-center rounded-md border border-transparent bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
				>
					Import Repository
				</a>
			</div>
		</div>
	</div>
</div>
