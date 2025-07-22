<script lang="ts">
	import { onMount } from 'svelte';

	// The layout now handles the auth check.
	onMount(() => {
		// You can add other initialization logic here if needed
	});

	let selectedRepo = 'backend-api';
	let messages = [
		{
			role: 'assistant',
			content:
				"Hello! I'm ready to help you analyze your code. What would you like to know about your repository?",
			timestamp: new Date()
		}
	];
	let inputText = '';
	let loading = false;

	// Mock repositories for selector
	const repositories = [
		{ id: 'backend-api', name: 'backend-api' },
		{ id: 'frontend-web', name: 'frontend-web' },
		{ id: 'mobile-app', name: 'mobile-app' }
	];
	// Suggested questions
	const suggestedQuestions = [
		'Explain the authentication flow',
		'What does the main API handler do?',
		'Find potential improvements in the database queries',
		'Show me the error handling patterns',
		'What are the main security concerns?'
	];

	async function sendMessage() {
		if (!inputText.trim()) return;

		const userMessage = {
			role: 'user',
			content: inputText,
			timestamp: new Date()
		};

		messages = [...messages, userMessage];
		const query = inputText;
		inputText = '';
		loading = true;
		try {
			// TODO: Replace with actual API call
			await new Promise((resolve) => setTimeout(resolve, 2000));
			// Simulate API call

			// Mock response
			const assistantMessage = {
				role: 'assistant',
				content: `Based on my analysis of the ${selectedRepo} repository, here's what I found regarding "${query}":

\`\`\`go
// Example code snippet that relates to your question
func HandleAuthentication(c *gin.Context) {
    token := c.GetHeader("Authorization")
    if token == "" {
        c.JSON(401, gin.H{"error": "Missing authorization token"})
        return
    }
    // Validate token logic here
}
\`\`\`

The authentication flow follows a JWT-based approach where:
1. Users provide credentials via POST /auth/login
2. Server validates credentials against the database
3. A JWT token is generated and returned
4. Subsequent requests include the token in the Authorization header
5. Middleware validates the token on protected routes

This implementation is secure but could be enhanced with refresh tokens for better user experience.`,
				timestamp: new Date(),
				analyzingFiles: ['src/auth/middleware.go', 'src/handlers/auth.go', 'src/models/user.go']
			};
			messages = [...messages, assistantMessage];
		} catch { // FIX: Removed unused '_err' variable
			const errorMessage = {
				role: 'assistant',
				content: 'Sorry, I encountered an error processing your request. Please try again.',
				timestamp: new Date()
			};
			messages = [...messages, errorMessage];
		} finally {
			loading = false;
		}
	}

	function askSuggested(question: string) {
		inputText = question;
		sendMessage();
	}
</script>

<svelte:head>
	<title>AI Chat - GitHub Analyzer</title>
</svelte:head>

<div class="flex h-[calc(100vh-12rem)]">
	<div class="flex flex-1 flex-col rounded-lg bg-white shadow">
		<div class="flex items-center justify-between border-b border-gray-200 p-4">
			<div class="flex items-center space-x-3">
				<div class="h-3 w-3 rounded-full bg-green-500"></div>
				<h2 class="text-lg font-medium text-gray-900">AI Code Assistant</h2>
			</div>
			<div class="flex items-center space-x-2">
				<label for="repo-select" class="text-sm font-medium text-gray-700">Repository:</label>
				<select
					id="repo-select"
					bind:value={selectedRepo}
					class="rounded-md border border-gray-300 px-3 py-1 text-sm focus:ring-2 focus:ring-blue-500 focus:outline-none"
				>
					{#each repositories as repo (repo.id)}
						<option value={repo.id}>{repo.name}</option>
					{/each}
				</select>
			</div>
		</div>

		<div class="flex-1 space-y-4 overflow-y-auto p-4">
			{#each messages as message, i (i)}
				<div class="flex {message.role === 'user' ? 'justify-end' : 'justify-start'}">
					<div
						class="max-w-3xl {message.role === 'user'
							? 'bg-blue-600 text-white'
							: 'bg-gray-100 text-gray-900'} rounded-lg px-4 py-2"
					>
						{#if message.role === 'assistant'}
							<div class="prose prose-sm max-w-none">
								// @eslint-disable-next-line svelte/no-at-html-tags
								{@html message.content.replace(
									/```(\w+)?\n([\s\S]*?)```/g,
									'<pre class="bg-gray-800 text-green-400 p-3 rounded mt-2 mb-2 overflow-x-auto"><code>$2</code></pre>'
								)}
							</div>
						{:else}
							<p class="text-sm">{message.content}</p>
						{/if}
						<div class="mt-1 text-xs opacity-70">
							{message.timestamp.toLocaleTimeString()}
						</div>
					</div>
				</div>
			{/each}

			{#if loading}
				<div class="flex justify-start">
					<div class="rounded-lg bg-gray-100 px-4 py-2">
						<div class="flex items-center space-x-2">
							<div class="h-4 w-4 animate-spin rounded-full border-b-2 border-blue-600"></div>
							<span class="text-sm text-gray-600">Analyzing code...</span>
						</div>
					</div>
				</div>
			{/if}
		</div>

		{#if messages.length === 1}
			<div class="border-t border-gray-200 p-4">
				<p class="mb-2 text-sm font-medium text-gray-700">Try asking:</p>
				<div class="flex flex-wrap gap-2">
					{#each suggestedQuestions as question (question)}
						<button
							on:click={() => askSuggested(question)}
							class="inline-flex items-center rounded-full bg-gray-100 px-3 py-1 text-sm text-gray-700 hover:bg-gray-200"
						>
							{question}
						</button>
					{/each}
				</div>
			</div>
		{/if}

		<div class="border-t border-gray-200 p-4">
			<form on:submit|preventDefault={sendMessage} class="flex space-x-2">
				<input
					bind:value={inputText}
					placeholder="Ask about the code..."
					class="flex-1 rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-transparent focus:ring-2 focus:ring-blue-500 focus:outline-none"
					disabled={loading}
				/>
				<button
					type="submit"
					disabled={loading || !inputText.trim()}
					class="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
				>
					Send
				</button>
			</form>
		</div>
	</div>
</div>