<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { getRepositories } from '$lib/api/hooks';
	import { chatStore, chatActions } from '$lib/stores/chat';
	import { chatClient, ChatAPIError } from '$lib/api/chat-client';
	import type { Repository } from '$lib/api';
	import type { components } from '$lib/api/types';

	type ChatSession = components['schemas']['ChatSession'];
	type ChatMessage = components['schemas']['ChatMessage'];

	// Component state
	let selectedRepo = $state('');
	let inputText = $state('');
	let repositories = $state<Repository[]>([]);
	let repositoriesLoading = $state(true);
	let showSessionSidebar = $state(false);
	let messagesContainer: HTMLElement;

	// Subscribe to chat store
	let chatState = $state($chatStore);
	chatStore.subscribe((value) => (chatState = value));

	// Suggested questions
	const suggestedQuestions = [
		'Explain the authentication flow',
		'What does the main API handler do?',
		'Find potential improvements in the database queries',
		'Show me the error handling patterns',
		'What are the main security concerns?'
	];

	// Load data when component mounts
	onMount(async () => {
		await Promise.all([loadRepositories(), loadChatSessions()]);
	});

	onDestroy(() => {
		chatActions.reset();
	});

	async function loadRepositories() {
		try {
			repositoriesLoading = true;
			const response = await getRepositories({ limit: 50 });
			const repoArray = response.repositories ?? [];
			repositories = repoArray.map((repo) => ({
				...repo,
				isPrivate: repo.isPrivate ?? false
			}));

			// Auto-select first repository if available
			if (repositories.length > 0 && !selectedRepo) {
				selectedRepo = repositories[0].id;
			}
		} catch (error) {
			console.error('Failed to load repositories:', error);
			chatActions.setError('Failed to load repositories');
		} finally {
			repositoriesLoading = false;
		}
	}

	async function loadChatSessions() {
		try {
			chatActions.setSessionsLoading(true);
			const response = await chatClient.listSessions({ limit: 20 });
			chatActions.setSessions(response.sessions);
		} catch (error) {
			console.error('Failed to load chat sessions:', error);
			chatActions.setError('Failed to load chat sessions');
		} finally {
			chatActions.setSessionsLoading(false);
		}
	}

	async function createNewSession() {
		try {
			chatActions.setLoading(true);
			const session = await chatClient.createSession({
				repositoryId: selectedRepo || undefined
			});
			chatActions.addSession(session);
			showSessionSidebar = false;
		} catch (error) {
			console.error('Failed to create session:', error);
			chatActions.setError(
				error instanceof ChatAPIError ? error.message : 'Failed to create session'
			);
		} finally {
			chatActions.setLoading(false);
		}
	}

	async function selectSession(session: ChatSession) {
		try {
			chatActions.setLoading(true);
			const fullSession = await chatClient.getSession(session.id);
			chatActions.setCurrentSession(fullSession);
			showSessionSidebar = false;
		} catch (error) {
			console.error('Failed to load session:', error);
			chatActions.setError(
				error instanceof ChatAPIError ? error.message : 'Failed to load session'
			);
		} finally {
			chatActions.setLoading(false);
		}
	}

	async function deleteSession(sessionId: string) {
		try {
			await chatClient.deleteSession(sessionId);
			chatActions.removeSession(sessionId);
		} catch (error) {
			console.error('Failed to delete session:', error);
			chatActions.setError(
				error instanceof ChatAPIError ? error.message : 'Failed to delete session'
			);
		}
	}

	async function sendMessage(event: Event) {
		event.preventDefault();
		if (!inputText.trim()) return;

		// Create session if none exists
		if (!chatState.currentSession) {
			await createNewSession();
			if (!chatState.currentSession) return;
		}

		const sessionId = chatState.currentSession.id;
		const messageContent = inputText.trim();
		inputText = '';

		// Add user message to UI immediately
		const userMessage: ChatMessage = {
			id: `temp-${Date.now()}`,
			role: 'user',
			content: messageContent,
			timestamp: new Date().toISOString()
		};
		chatActions.addMessage(sessionId, userMessage);

		// Add empty assistant message for streaming
		const assistantMessage: ChatMessage = {
			id: `temp-assistant-${Date.now()}`,
			role: 'assistant',
			content: '',
			timestamp: new Date().toISOString()
		};
		chatActions.addMessage(sessionId, assistantMessage);
		chatActions.setStreaming(true);

		try {
			let fullContent = '';
			await chatClient.sendMessageStream(sessionId, messageContent, (chunk) => {
				if (chunk.type === 'content') {
					fullContent = chunk.content;
					chatActions.updateLastMessage(sessionId, { content: fullContent });
					scrollToBottom();
				} else if (chunk.type === 'error') {
					chatActions.updateLastMessage(sessionId, {
						content: `Error: ${chunk.content}`
					});
				} else if (chunk.type === 'done') {
					// Update with final content and proper ID
					chatActions.updateLastMessage(sessionId, {
						id: `assistant-${Date.now()}`,
						content: fullContent || 'No response received'
					});
				}
			});
		} catch (error) {
			console.error('Failed to send message:', error);
			chatActions.updateLastMessage(sessionId, {
				content: `Error: ${error instanceof ChatAPIError ? error.message : 'Failed to send message'}`
			});
		} finally {
			chatActions.setStreaming(false);
			scrollToBottom();
		}
	}

	function askSuggested(question: string) {
		inputText = question;
		sendMessage(new Event('submit'));
	}

	function scrollToBottom() {
		if (messagesContainer) {
			setTimeout(() => {
				messagesContainer.scrollTop = messagesContainer.scrollHeight;
			}, 100);
		}
	}

	// Format timestamp for display
	function formatTime(timestamp: string) {
		return new Date(timestamp).toLocaleTimeString([], {
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	// Get current messages to display
	let currentMessages = $derived(chatState.currentSession?.messages || []);
	let showSuggestedQuestions = $derived(currentMessages.length === 0);
	let canSendMessage = $derived(!chatState.streaming && !chatState.loading);
</script>

<svelte:head>
	<title>AI Chat - GitHub Analyzer</title>
</svelte:head>

<div class="flex h-[calc(100vh-12rem)]">
	<!-- Session Sidebar -->
	{#if showSessionSidebar}
		<div class="w-80 border-r border-gray-200 bg-white">
			<div class="border-b border-gray-200 p-4">
				<div class="flex items-center justify-between">
					<h3 class="text-lg font-medium text-gray-900">Chat Sessions</h3>
					<button
						onclick={() => (showSessionSidebar = false)}
						class="text-gray-400 hover:text-gray-600"
						aria-label="Close session sidebar"
					>
						<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M6 18L18 6M6 6l12 12"
							/>
						</svg>
					</button>
				</div>
				<button
					onclick={createNewSession}
					class="mt-2 w-full rounded-md bg-blue-600 px-3 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
					disabled={chatState.loading}
				>
					{#if chatState.loading}
						Creating...
					{:else}
						New Chat
					{/if}
				</button>
			</div>
			<div class="flex-1 overflow-y-auto p-4">
				{#if chatState.sessionsLoading}
					<div class="flex items-center justify-center py-8">
						<div class="h-6 w-6 animate-spin rounded-full border-b-2 border-blue-600"></div>
					</div>
				{:else if chatState.sessions.length === 0}
					<p class="py-8 text-center text-gray-500">No chat sessions yet</p>
				{:else}
					<div class="space-y-2">
						{#each chatState.sessions as session (session.id)}
							<div
								class="flex items-center justify-between rounded-lg border p-3 hover:bg-gray-50 {chatState
									.currentSession?.id === session.id
									? 'border-blue-500 bg-blue-50'
									: 'border-gray-200'}"
							>
								<button onclick={() => selectSession(session)} class="flex-1 text-left">
									<div class="truncate text-sm font-medium text-gray-900">{session.title}</div>
									<div class="text-xs text-gray-500">{formatTime(session.updatedAt)}</div>
								</button>
								<button
									onclick={() => deleteSession(session.id)}
									class="ml-2 text-gray-400 hover:text-red-600"
									aria-label="Delete session"
								>
									<svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
										/>
									</svg>
								</button>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>
	{/if}

	<!-- Main Chat Area -->
	<div class="flex flex-1 flex-col rounded-lg bg-white shadow">
		<div class="flex items-center justify-between border-b border-gray-200 p-4">
			<div class="flex items-center space-x-3">
				<button
					onclick={() => (showSessionSidebar = !showSessionSidebar)}
					class="rounded-md p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
					aria-label="Toggle session sidebar"
				>
					<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							stroke-width="2"
							d="M4 6h16M4 12h16M4 18h16"
						/>
					</svg>
				</button>
				<div
					class="h-3 w-3 rounded-full {chatState.streaming ? 'bg-orange-500' : 'bg-green-500'}"
				></div>
				<h2 class="text-lg font-medium text-gray-900">
					{chatState.currentSession?.title || 'AI Code Assistant'}
				</h2>
			</div>
			<div class="flex items-center space-x-2">
				<label for="repo-select" class="text-sm font-medium text-gray-700">Repository:</label>
				{#if repositoriesLoading}
					<div class="flex items-center space-x-2">
						<div class="h-4 w-4 animate-spin rounded-full border-b-2 border-blue-600"></div>
						<span class="text-sm text-gray-500">Loading...</span>
					</div>
				{:else if repositories.length === 0}
					<span class="text-sm text-gray-500">No repositories found</span>
				{:else}
					<select
						id="repo-select"
						bind:value={selectedRepo}
						class="rounded-md border border-gray-300 px-3 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
					>
						<option value="">All repositories</option>
						{#each repositories as repo (repo.id)}
							<option value={repo.id}>{repo.fullName}</option>
						{/each}
					</select>
				{/if}
			</div>
		</div>

		<div bind:this={messagesContainer} class="flex-1 space-y-4 overflow-y-auto p-4">
			{#if chatState.error}
				<div class="rounded-lg border border-red-200 bg-red-50 p-4">
					<div class="flex">
						<svg class="h-5 w-5 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"
							/>
						</svg>
						<div class="ml-3">
							<p class="text-sm text-red-800">{chatState.error}</p>
							<button
								onclick={() => chatActions.setError(null)}
								class="mt-1 text-xs text-red-600 underline hover:text-red-800"
							>
								Dismiss
							</button>
						</div>
					</div>
				</div>
			{/if}

			{#if currentMessages.length === 0 && !chatState.loading}
				<div class="flex h-full items-center justify-center">
					<div class="text-center">
						<svg
							class="mx-auto h-12 w-12 text-gray-400"
							fill="none"
							viewBox="0 0 24 24"
							stroke="currentColor"
						>
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-3.582 8-8 8a8.955 8.955 0 01-4.126-.98L3 20l1.98-5.126A8.955 8.955 0 013 12c0-4.418 3.582-8 8-8s8 3.582 8 8z"
							/>
						</svg>
						<h3 class="mt-2 text-sm font-medium text-gray-900">Start a conversation</h3>
						<p class="mt-1 text-sm text-gray-500">Ask me anything about your code!</p>
					</div>
				</div>
			{:else}
				{#each currentMessages as message, i (message.id || i)}
					<div class="flex {message.role === 'user' ? 'justify-end' : 'justify-start'}">
						<div
							class="max-w-3xl {message.role === 'user'
								? 'bg-blue-600 text-white'
								: 'bg-gray-100 text-gray-900'} rounded-lg px-4 py-2"
						>
							<div class="prose prose-sm max-w-none">
								{#if message.content}
									<div class="whitespace-pre-wrap text-sm">{message.content}</div>
								{:else if message.role === 'assistant'}
									<div class="flex items-center space-x-2">
										<div class="h-4 w-4 animate-spin rounded-full border-b-2 border-blue-600"></div>
										<span class="text-sm text-gray-600">Thinking...</span>
									</div>
								{/if}
							</div>
							{#if message.retrievedChunks && message.retrievedChunks.length > 0}
								<div class="mt-2 text-xs opacity-70">
									Analyzed {message.retrievedChunks.length} code chunks
								</div>
							{/if}
							<div class="mt-1 flex items-center justify-between text-xs opacity-70">
								<span>{formatTime(message.timestamp)}</span>
								{#if message.tokensUsed}
									<span>{message.tokensUsed} tokens</span>
								{/if}
							</div>
						</div>
					</div>
				{/each}

				{#if chatState.streaming}
					<div class="flex justify-start">
						<div class="rounded-lg bg-gray-100 px-4 py-2">
							<div class="flex items-center space-x-2">
								<div class="h-4 w-4 animate-spin rounded-full border-b-2 border-blue-600"></div>
								<span class="text-sm text-gray-600">Analyzing code...</span>
							</div>
						</div>
					</div>
				{/if}
			{/if}
		</div>

		{#if showSuggestedQuestions}
			<div class="border-t border-gray-200 p-4">
				<p class="mb-2 text-sm font-medium text-gray-700">Try asking:</p>
				<div class="flex flex-wrap gap-2">
					{#each suggestedQuestions as question (question)}
						<button
							onclick={() => askSuggested(question)}
							class="inline-flex items-center rounded-full bg-gray-100 px-3 py-1 text-sm text-gray-700 hover:bg-gray-200 disabled:opacity-50"
							disabled={!canSendMessage}
						>
							{question}
						</button>
					{/each}
				</div>
			</div>
		{/if}

		<div class="border-t border-gray-200 p-4">
			<form onsubmit={sendMessage} class="flex space-x-2">
				<input
					bind:value={inputText}
					placeholder="Ask about the code..."
					class="flex-1 rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-transparent focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-50"
					disabled={!canSendMessage}
					autocomplete="off"
				/>
				<button
					type="submit"
					disabled={!canSendMessage || !inputText.trim()}
					class="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
				>
					{#if chatState.streaming}
						Stop
					{:else if chatState.loading}
						Sending...
					{:else}
						Send
					{/if}
				</button>
			</form>
		</div>
	</div>
</div>
