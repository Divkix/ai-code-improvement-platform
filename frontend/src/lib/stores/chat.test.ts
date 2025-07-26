// ABOUTME: Tests for chat store functionality and state management
// ABOUTME: Covers chat sessions, messages, and streaming state handling

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { get } from 'svelte/store';
import { chatStore, chatActions } from './chat';
import type { components } from '$lib/api/types';

type ChatSession = components['schemas']['ChatSession'];
type ChatMessage = components['schemas']['ChatMessage'];

describe('chatStore', () => {
	beforeEach(() => {
		// Reset store to initial state
		chatActions.reset();
	});

	describe('initial state', () => {
		it('should have correct initial state', () => {
			const state = get(chatStore);
			expect(state.currentSession).toBeNull();
			expect(state.sessions).toEqual([]);
			expect(state.loading).toBe(false);
			expect(state.sessionsLoading).toBe(false);
			expect(state.streaming).toBe(false);
			expect(state.error).toBeNull();
		});
	});

	describe('setCurrentSession', () => {
		it('should set current session', () => {
			const mockSession: ChatSession = {
				id: 'session-1',
				title: 'Test Session',
				messages: [],
				repositoryId: 'repo-1',
				createdAt: '2024-01-01T00:00:00Z',
				updatedAt: '2024-01-01T00:00:00Z'
			};

			chatActions.setCurrentSession(mockSession);

			const state = get(chatStore);
			expect(state.currentSession).toEqual(mockSession);
			expect(state.error).toBeNull();
		});

		it('should clear current session when passed null', () => {
			const mockSession: ChatSession = {
				id: 'session-1',
				title: 'Test Session',
				messages: [],
				repositoryId: 'repo-1',
				createdAt: '2024-01-01T00:00:00Z',
				updatedAt: '2024-01-01T00:00:00Z'
			};

			chatActions.setCurrentSession(mockSession);
			chatActions.setCurrentSession(null);

			const state = get(chatStore);
			expect(state.currentSession).toBeNull();
		});
	});

	describe('setSessions', () => {
		it('should set sessions array', () => {
			const mockSessions: ChatSession[] = [
				{
					id: 'session-1',
					title: 'Session 1',
					messages: [],
					repositoryId: 'repo-1',
					createdAt: '2024-01-01T00:00:00Z',
					updatedAt: '2024-01-01T00:00:00Z'
				},
				{
					id: 'session-2',
					title: 'Session 2',
					messages: [],
					repositoryId: 'repo-1',
					createdAt: '2024-01-01T00:00:00Z',
					updatedAt: '2024-01-01T00:00:00Z'
				}
			];

			chatActions.setSessions(mockSessions);

			const state = get(chatStore);
			expect(state.sessions).toEqual(mockSessions);
		});
	});

	describe('addSession', () => {
		it('should add session and set as current', () => {
			const mockSession: ChatSession = {
				id: 'session-1',
				title: 'New Session',
				messages: [],
				repositoryId: 'repo-1',
				createdAt: '2024-01-01T00:00:00Z',
				updatedAt: '2024-01-01T00:00:00Z'
			};

			chatActions.addSession(mockSession);

			const state = get(chatStore);
			expect(state.sessions).toContain(mockSession);
			expect(state.currentSession).toEqual(mockSession);
		});

		it('should add new session at the beginning of sessions array', () => {
			const session1: ChatSession = {
				id: 'session-1',
				title: 'Session 1',
				messages: [],
				repositoryId: 'repo-1',
				createdAt: '2024-01-01T00:00:00Z',
				updatedAt: '2024-01-01T00:00:00Z'
			};

			const session2: ChatSession = {
				id: 'session-2',
				title: 'Session 2',
				messages: [],
				repositoryId: 'repo-1',
				createdAt: '2024-01-01T00:00:00Z',
				updatedAt: '2024-01-01T00:00:00Z'
			};

			chatActions.addSession(session1);
			chatActions.addSession(session2);

			const state = get(chatStore);
			expect(state.sessions[0]).toEqual(session2);
			expect(state.sessions[1]).toEqual(session1);
		});
	});

	describe('updateSession', () => {
		it('should update session in sessions array', () => {
			const mockSession: ChatSession = {
				id: 'session-1',
				title: 'Original Title',
				messages: [],
				repositoryId: 'repo-1',
				createdAt: '2024-01-01T00:00:00Z',
				updatedAt: '2024-01-01T00:00:00Z'
			};

			chatActions.addSession(mockSession);
			chatActions.updateSession('session-1', { title: 'Updated Title' });

			const state = get(chatStore);
			const updatedSession = state.sessions.find((s) => s.id === 'session-1');
			expect(updatedSession?.title).toBe('Updated Title');
		});

		it('should update current session if it matches', () => {
			const mockSession: ChatSession = {
				id: 'session-1',
				title: 'Original Title',
				messages: [],
				repositoryId: 'repo-1',
				createdAt: '2024-01-01T00:00:00Z',
				updatedAt: '2024-01-01T00:00:00Z'
			};

			chatActions.setCurrentSession(mockSession);
			chatActions.addSession(mockSession);
			chatActions.updateSession('session-1', { title: 'Updated Title' });

			const state = get(chatStore);
			expect(state.currentSession?.title).toBe('Updated Title');
		});
	});

	describe('removeSession', () => {
		it('should remove session from sessions array', () => {
			const session1: ChatSession = {
				id: 'session-1',
				title: 'Session 1',
				messages: [],
				repositoryId: 'repo-1',
				createdAt: '2024-01-01T00:00:00Z',
				updatedAt: '2024-01-01T00:00:00Z'
			};

			const session2: ChatSession = {
				id: 'session-2',
				title: 'Session 2',
				messages: [],
				repositoryId: 'repo-1',
				createdAt: '2024-01-01T00:00:00Z',
				updatedAt: '2024-01-01T00:00:00Z'
			};

			chatActions.addSession(session1);
			chatActions.addSession(session2);
			chatActions.removeSession('session-1');

			const state = get(chatStore);
			expect(state.sessions).toHaveLength(1);
			expect(state.sessions[0].id).toBe('session-2');
		});

		it('should update current session when removing current session', () => {
			const session1: ChatSession = {
				id: 'session-1',
				title: 'Session 1',
				messages: [],
				repositoryId: 'repo-1',
				createdAt: '2024-01-01T00:00:00Z',
				updatedAt: '2024-01-01T00:00:00Z'
			};

			const session2: ChatSession = {
				id: 'session-2',
				title: 'Session 2',
				messages: [],
				repositoryId: 'repo-1',
				createdAt: '2024-01-01T00:00:00Z',
				updatedAt: '2024-01-01T00:00:00Z'
			};

			chatActions.addSession(session1);
			chatActions.addSession(session2);
			chatActions.setCurrentSession(session1);
			chatActions.removeSession('session-1');

			const state = get(chatStore);
			expect(state.currentSession?.id).toBe('session-2');
		});

		it('should set current session to null when removing last session', () => {
			const session: ChatSession = {
				id: 'session-1',
				title: 'Session 1',
				messages: [],
				repositoryId: 'repo-1',
				createdAt: '2024-01-01T00:00:00Z',
				updatedAt: '2024-01-01T00:00:00Z'
			};

			chatActions.addSession(session);
			chatActions.removeSession('session-1');

			const state = get(chatStore);
			expect(state.currentSession).toBeNull();
			expect(state.sessions).toHaveLength(0);
		});
	});

	describe('addMessage', () => {
		it('should add message to session', () => {
			const mockSession: ChatSession = {
				id: 'session-1',
				title: 'Test Session',
				messages: [],
				repositoryId: 'repo-1',
				createdAt: '2024-01-01T00:00:00Z',
				updatedAt: '2024-01-01T00:00:00Z'
			};

			const mockMessage: ChatMessage = {
				id: 'message-1',
				role: 'user',
				content: 'Hello',
				timestamp: '2024-01-01T00:00:00Z'
			};

			// Mock Date constructor for consistent timestamp
			const mockDate = new Date('2024-01-01T12:00:00Z');
			const dateSpy = vi.spyOn(global, 'Date').mockImplementation(() => mockDate);

			chatActions.addSession(mockSession);
			chatActions.addMessage('session-1', mockMessage);

			const state = get(chatStore);
			const session = state.sessions.find((s) => s.id === 'session-1');
			expect(session?.messages).toContain(mockMessage);
			expect(session?.updatedAt).toBe(mockDate.toISOString());

			dateSpy.mockRestore();
		});

		it('should add message to current session', () => {
			const mockSession: ChatSession = {
				id: 'session-1',
				title: 'Test Session',
				messages: [],
				repositoryId: 'repo-1',
				createdAt: '2024-01-01T00:00:00Z',
				updatedAt: '2024-01-01T00:00:00Z'
			};

			const mockMessage: ChatMessage = {
				id: 'message-1',
				role: 'user',
				content: 'Hello',
				timestamp: '2024-01-01T00:00:00Z'
			};

			chatActions.setCurrentSession(mockSession);
			chatActions.addSession(mockSession);
			chatActions.addMessage('session-1', mockMessage);

			const state = get(chatStore);
			expect(state.currentSession?.messages).toContain(mockMessage);
		});
	});

	describe('updateLastMessage', () => {
		it('should update last message in session', () => {
			const mockSession: ChatSession = {
				id: 'session-1',
				title: 'Test Session',
				messages: [
					{
						id: 'message-1',
						role: 'user',
						content: 'Hello',
						timestamp: '2024-01-01T00:00:00Z'
					},
					{
						id: 'message-2',
						role: 'assistant',
						content: 'Hi there',
						timestamp: '2024-01-01T00:00:00Z'
					}
				],
				repositoryId: 'repo-1',
				createdAt: '2024-01-01T00:00:00Z',
				updatedAt: '2024-01-01T00:00:00Z'
			};

			chatActions.addSession(mockSession);
			chatActions.updateLastMessage('session-1', { content: 'Updated response' });

			const state = get(chatStore);
			const session = state.sessions.find((s) => s.id === 'session-1');
			const lastMessage = session?.messages[session.messages.length - 1];
			expect(lastMessage?.content).toBe('Updated response');
		});

		it('should handle empty messages array', () => {
			const mockSession: ChatSession = {
				id: 'session-1',
				title: 'Test Session',
				messages: [],
				repositoryId: 'repo-1',
				createdAt: '2024-01-01T00:00:00Z',
				updatedAt: '2024-01-01T00:00:00Z'
			};

			chatActions.addSession(mockSession);
			chatActions.updateLastMessage('session-1', { content: 'Should not crash' });

			const state = get(chatStore);
			const session = state.sessions.find((s) => s.id === 'session-1');
			expect(session?.messages).toHaveLength(0);
		});
	});

	describe('state flags', () => {
		it('should set loading state', () => {
			chatActions.setLoading(true);
			expect(get(chatStore).loading).toBe(true);

			chatActions.setLoading(false);
			expect(get(chatStore).loading).toBe(false);
		});

		it('should set sessions loading state', () => {
			chatActions.setSessionsLoading(true);
			expect(get(chatStore).sessionsLoading).toBe(true);

			chatActions.setSessionsLoading(false);
			expect(get(chatStore).sessionsLoading).toBe(false);
		});

		it('should set streaming state', () => {
			chatActions.setStreaming(true);
			expect(get(chatStore).streaming).toBe(true);

			chatActions.setStreaming(false);
			expect(get(chatStore).streaming).toBe(false);
		});

		it('should set error state', () => {
			chatActions.setError('Test error');
			expect(get(chatStore).error).toBe('Test error');

			chatActions.setError(null);
			expect(get(chatStore).error).toBeNull();
		});
	});

	describe('renameSession', () => {
		it('should rename session in sessions array', () => {
			const mockSession: ChatSession = {
				id: 'session-1',
				title: 'Original Title',
				messages: [],
				repositoryId: 'repo-1',
				createdAt: '2024-01-01T00:00:00Z',
				updatedAt: '2024-01-01T00:00:00Z'
			};

			chatActions.addSession(mockSession);
			chatActions.renameSession('session-1', 'New Title');

			const state = get(chatStore);
			const session = state.sessions.find((s) => s.id === 'session-1');
			expect(session?.title).toBe('New Title');
		});

		it('should rename current session if it matches', () => {
			const mockSession: ChatSession = {
				id: 'session-1',
				title: 'Original Title',
				messages: [],
				repositoryId: 'repo-1',
				createdAt: '2024-01-01T00:00:00Z',
				updatedAt: '2024-01-01T00:00:00Z'
			};

			chatActions.setCurrentSession(mockSession);
			chatActions.addSession(mockSession);
			chatActions.renameSession('session-1', 'New Title');

			const state = get(chatStore);
			expect(state.currentSession?.title).toBe('New Title');
		});
	});

	describe('reset', () => {
		it('should reset store to initial state', () => {
			const mockSession: ChatSession = {
				id: 'session-1',
				title: 'Test Session',
				messages: [],
				repositoryId: 'repo-1',
				createdAt: '2024-01-01T00:00:00Z',
				updatedAt: '2024-01-01T00:00:00Z'
			};

			chatActions.addSession(mockSession);
			chatActions.setLoading(true);
			chatActions.setError('Test error');
			chatActions.reset();

			const state = get(chatStore);
			expect(state.currentSession).toBeNull();
			expect(state.sessions).toEqual([]);
			expect(state.loading).toBe(false);
			expect(state.sessionsLoading).toBe(false);
			expect(state.streaming).toBe(false);
			expect(state.error).toBeNull();
		});
	});
});
