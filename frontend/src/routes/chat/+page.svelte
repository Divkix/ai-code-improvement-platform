<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { getRepositories } from '$lib/api/hooks';
	import { chatStore, chatActions } from '$lib/stores/chat';
	import { chatClient, ChatAPIError, type ChatStreamChunk } from '$lib/api/chat-client';
	import { parseMarkdown, hasMarkdownFormatting } from '$lib/utils/markdown';
	import type { Repository } from '$lib/api';
	import type { components } from '$lib/api/types';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Label } from '$lib/components/ui/label/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import * as Alert from '$lib/components/ui/alert/index.js';
	import { Loader2, MessageCircle, Edit, Trash2, Plus, MoreHorizontal } from '@lucide/svelte';

	type ChatSession = components['schemas']['ChatSession'];
	type ChatMessage = components['schemas']['ChatMessage'];

	// Component state
	let selectedRepo = $state('');
	let inputText = $state('');
	let repositories = $state<Repository[]>([]);
	let repositoriesLoading = $state(true);
	let messagesContainer: HTMLElement;
	let renamingSessionId = $state<string | null>(null);
	let renameInputValue = $state('');

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
			// Sidebar is always visible in fullscreen mode
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
			// Sidebar is always visible in fullscreen mode
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

	function startRenaming(sessionId: string, currentTitle: string) {
		renamingSessionId = sessionId;
		renameInputValue = currentTitle;
	}

	function cancelRenaming() {
		renamingSessionId = null;
		renameInputValue = '';
	}

	async function saveRename(sessionId: string) {
		if (!renameInputValue.trim()) {
			cancelRenaming();
			return;
		}

		try {
			const updatedSession = await chatClient.updateSession(sessionId, {
				title: renameInputValue.trim()
			});
			chatActions.updateSession(sessionId, { title: updatedSession.title });
			cancelRenaming();
		} catch (error) {
			console.error('Failed to rename session:', error);
			chatActions.setError(
				error instanceof ChatAPIError ? error.message : 'Failed to rename session'
			);
		}
	}

	function handleRenameKeydown(event: KeyboardEvent, sessionId: string) {
		if (event.key === 'Enter') {
			event.preventDefault();
			saveRename(sessionId);
		} else if (event.key === 'Escape') {
			event.preventDefault();
			cancelRenaming();
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
			await chatClient.sendMessageStream(sessionId, messageContent, (chunk: ChatStreamChunk) => {
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

<!-- Fullscreen chat container that bypasses layout constraints -->
<div class="fixed inset-0 z-50 bg-white">
	<!-- Custom header for fullscreen mode -->
	<div class="flex h-16 items-center justify-between border-b border-gray-200 bg-white px-4">
		<div class="flex items-center space-x-3">
			<a href="/" class="text-xl font-semibold text-gray-900">GitHub Analyzer</a>
			<span class="text-gray-400">|</span>
			<h1 class="text-lg font-medium text-gray-700">AI Chat</h1>
		</div>
		<div class="flex items-center space-x-4">
			<div class="flex items-center space-x-2">
				<Label class="text-sm font-medium">Repository:</Label>
				{#if repositoriesLoading}
					<div class="flex items-center space-x-2">
						<Loader2 class="h-4 w-4 animate-spin" />
						<span class="text-sm text-muted-foreground">Loading...</span>
					</div>
				{:else if repositories.length === 0}
					<span class="text-sm text-muted-foreground">No repositories found</span>
				{:else}
					<Select.Root
						type="single"
						bind:value={selectedRepo}
						onValueChange={(v) => (selectedRepo = v || '')}
					>
						<Select.Trigger class="w-48">
							{selectedRepo
								? repositories.find((r) => r.id === selectedRepo)?.fullName ||
									'Repository not found'
								: 'All repositories'}
						</Select.Trigger>
						<Select.Content>
							<Select.Item value="">All repositories</Select.Item>
							{#each repositories as repo (repo.id)}
								<Select.Item value={repo.id}>{repo.fullName}</Select.Item>
							{/each}
						</Select.Content>
					</Select.Root>
				{/if}
			</div>
			<a href="/" class="text-sm text-gray-500 hover:text-gray-700">← Back to Dashboard</a>
		</div>
	</div>

	<!-- Main chat interface -->
	<div class="flex h-[calc(100vh-4rem)]">
		<!-- Persistent Session Sidebar -->
		<div class="w-80 border-r border-gray-200 bg-gray-50">
			<div class="border-b border-gray-200 p-4">
				<div class="flex items-center justify-between">
					<h3 class="text-lg font-medium text-gray-900">Chat Sessions</h3>
				</div>
				<Button onclick={createNewSession} class="mt-2 w-full" disabled={chatState.loading}>
					{#if chatState.loading}
						<Loader2 class="mr-2 h-4 w-4 animate-spin" />
						Creating...
					{:else}
						<Plus class="mr-2 h-4 w-4" />
						New Chat
					{/if}
				</Button>
			</div>
			<div class="flex-1 overflow-y-auto p-4">
				{#if chatState.sessionsLoading}
					<div class="flex items-center justify-center py-8">
						<Loader2 class="h-6 w-6 animate-spin" />
					</div>
				{:else if chatState.sessions.length === 0}
					<p class="py-8 text-center text-gray-500">No chat sessions yet</p>
				{:else}
					<div class="space-y-2">
						{#each chatState.sessions as session (session.id)}
							<div
								class="rounded-lg border p-3 hover:bg-gray-50 {chatState.currentSession?.id ===
								session.id
									? 'border-blue-500 bg-blue-50'
									: 'border-gray-200'}"
							>
								{#if renamingSessionId === session.id}
									<!-- Rename mode -->
									<div class="space-y-2">
										<Input
											bind:value={renameInputValue}
											onkeydown={(e) => handleRenameKeydown(e, session.id)}
											placeholder="Enter new name..."
											class="text-sm"
										/>
										<div class="flex justify-end space-x-2">
											<Button variant="ghost" size="sm" onclick={cancelRenaming}>Cancel</Button>
											<Button size="sm" onclick={() => saveRename(session.id)}>Save</Button>
										</div>
									</div>
								{:else}
									<!-- Normal mode -->
									<div class="flex items-center">
										<button
											onclick={() => selectSession(session)}
											class="min-w-0 flex-1 pr-2 text-left"
										>
											<div class="truncate text-sm font-medium text-gray-900">{session.title}</div>
											<div class="text-xs text-gray-500">{formatTime(session.updatedAt)}</div>
										</button>
										<div class="relative flex-shrink-0">
											<DropdownMenu.Root>
												<DropdownMenu.Trigger>
													<Button variant="ghost" size="sm" class="h-8 w-8 p-0">
														<MoreHorizontal class="h-4 w-4" />
													</Button>
												</DropdownMenu.Trigger>
												<DropdownMenu.Content>
													<DropdownMenu.Item
														onclick={() => startRenaming(session.id, session.title)}
													>
														<Edit class="mr-2 h-4 w-4" />
														Edit
													</DropdownMenu.Item>
													<DropdownMenu.Item
														onclick={() => deleteSession(session.id)}
														class="text-destructive"
													>
														<Trash2 class="mr-2 h-4 w-4" />
														Delete
													</DropdownMenu.Item>
												</DropdownMenu.Content>
											</DropdownMenu.Root>
										</div>
									</div>
								{/if}
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>

		<!-- Main Chat Area -->
		<div class="flex flex-1 flex-col bg-white">
			<div class="flex items-center justify-between border-b border-gray-200 p-4">
				<div class="flex items-center space-x-3">
					<div
						class="h-3 w-3 rounded-full {chatState.streaming ? 'bg-orange-500' : 'bg-green-500'}"
					></div>
					<h2 class="text-lg font-medium text-gray-900">
						{chatState.currentSession?.title || 'AI Code Assistant'}
					</h2>
				</div>
			</div>

			<div bind:this={messagesContainer} class="flex-1 space-y-4 overflow-y-auto p-4">
				{#if chatState.error}
					<Alert.Root variant="destructive">
						<Alert.Description>
							<p>{chatState.error}</p>
							<Button
								variant="link"
								size="sm"
								onclick={() => chatActions.setError(null)}
								class="mt-1 h-auto p-0 text-xs underline"
							>
								Dismiss
							</Button>
						</Alert.Description>
					</Alert.Root>
				{/if}

				{#if currentMessages.length === 0 && !chatState.loading}
					<div class="flex h-full items-center justify-center">
						<div class="text-center">
							<MessageCircle class="mx-auto h-12 w-12 text-muted-foreground" />
							<h3 class="mt-2 text-sm font-medium">Start a conversation</h3>
							<p class="mt-1 text-sm text-muted-foreground">Ask me anything about your code!</p>
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
										{#if hasMarkdownFormatting(message.content)}
											{#await parseMarkdown(message.content)}
												<div class="text-sm whitespace-pre-wrap">{message.content}</div>
											{:then parsedContent}
												<!-- eslint-disable-next-line svelte/no-at-html-tags -->
												<div class="markdown-content text-sm">{@html parsedContent}</div>
											{:catch}
												<div class="text-sm whitespace-pre-wrap">{message.content}</div>
											{/await}
										{:else}
											<div class="text-sm whitespace-pre-wrap">{message.content}</div>
										{/if}
									{:else if message.role === 'assistant'}
										<div class="flex items-center space-x-2">
											<Loader2 class="h-4 w-4 animate-spin" />
											<span class="text-sm text-muted-foreground">Analyzing code...</span>
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
				{/if}
			</div>

			{#if showSuggestedQuestions}
				<div class="border-t border-gray-200 p-4">
					<p class="mb-2 text-sm font-medium">Try asking:</p>
					<div class="flex flex-wrap gap-2">
						{#each suggestedQuestions as question (question)}
							<Button
								variant="secondary"
								size="sm"
								onclick={() => askSuggested(question)}
								disabled={!canSendMessage}
								class="rounded-full"
							>
								{question}
							</Button>
						{/each}
					</div>
				</div>
			{/if}

			<div class="border-t border-gray-200 p-4">
				<form onsubmit={sendMessage} class="flex space-x-2">
					<Input
						bind:value={inputText}
						placeholder="Ask about the code..."
						disabled={!canSendMessage}
						autocomplete="off"
						class="flex-1"
					/>
					<Button type="submit" disabled={!canSendMessage || !inputText.trim()}>
						{#if chatState.streaming}
							Stop
						{:else if chatState.loading}
							<Loader2 class="mr-2 h-4 w-4 animate-spin" />
							Sending...
						{:else}
							Send
						{/if}
					</Button>
				</form>
			</div>
		</div>
	</div>
</div>
