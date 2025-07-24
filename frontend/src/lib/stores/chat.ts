// ABOUTME: Chat state management with Svelte stores
// ABOUTME: Handles chat sessions, messages, and streaming state for the chat interface

import { writable } from 'svelte/store';
import type { components } from '$lib/api/types';

type ChatSession = components['schemas']['ChatSession'];
type ChatMessage = components['schemas']['ChatMessage'];

export interface ChatState {
    currentSession: ChatSession | null;
    sessions: ChatSession[];
    loading: boolean;
    sessionsLoading: boolean;
    streaming: boolean;
    error: string | null;
}

const initialState: ChatState = {
    currentSession: null,
    sessions: [],
    loading: false,
    sessionsLoading: false,
    streaming: false,
    error: null
};

export const chatStore = writable<ChatState>(initialState);

// Helper functions to update the store
export const chatActions = {
    setCurrentSession: (session: ChatSession | null) => {
        chatStore.update(state => ({
            ...state,
            currentSession: session,
            error: null
        }));
    },

    setSessions: (sessions: ChatSession[]) => {
        chatStore.update(state => ({
            ...state,
            sessions
        }));
    },

    addSession: (session: ChatSession) => {
        chatStore.update(state => ({
            ...state,
            sessions: [session, ...state.sessions],
            currentSession: session
        }));
    },

    updateSession: (sessionId: string, updates: Partial<ChatSession>) => {
        chatStore.update(state => {
            const updatedSessions = state.sessions.map(session =>
                session.id === sessionId ? { ...session, ...updates } : session
            );
            const updatedCurrentSession = state.currentSession?.id === sessionId
                ? { ...state.currentSession, ...updates }
                : state.currentSession;

            return {
                ...state,
                sessions: updatedSessions,
                currentSession: updatedCurrentSession
            };
        });
    },

    removeSession: (sessionId: string) => {
        chatStore.update(state => {
            const filteredSessions = state.sessions.filter(session => session.id !== sessionId);
            const newCurrentSession = state.currentSession?.id === sessionId
                ? (filteredSessions.length > 0 ? filteredSessions[0] : null)
                : state.currentSession;

            return {
                ...state,
                sessions: filteredSessions,
                currentSession: newCurrentSession
            };
        });
    },

    addMessage: (sessionId: string, message: ChatMessage) => {
        chatStore.update(state => {
            const updatedSessions = state.sessions.map(session =>
                session.id === sessionId
                    ? { 
                        ...session, 
                        messages: [...session.messages, message],
                        updatedAt: new Date().toISOString()
                    }
                    : session
            );
            const updatedCurrentSession = state.currentSession?.id === sessionId
                ? {
                    ...state.currentSession,
                    messages: [...state.currentSession.messages, message],
                    updatedAt: new Date().toISOString()
                }
                : state.currentSession;

            return {
                ...state,
                sessions: updatedSessions,
                currentSession: updatedCurrentSession
            };
        });
    },

    updateLastMessage: (sessionId: string, updates: Partial<ChatMessage>) => {
        chatStore.update(state => {
            const updateMessages = (messages: ChatMessage[]) => {
                if (messages.length === 0) return messages;
                const lastIndex = messages.length - 1;
                return [
                    ...messages.slice(0, lastIndex),
                    { ...messages[lastIndex], ...updates }
                ];
            };

            const updatedSessions = state.sessions.map(session =>
                session.id === sessionId
                    ? { ...session, messages: updateMessages(session.messages) }
                    : session
            );
            const updatedCurrentSession = state.currentSession?.id === sessionId
                ? { ...state.currentSession, messages: updateMessages(state.currentSession.messages) }
                : state.currentSession;

            return {
                ...state,
                sessions: updatedSessions,
                currentSession: updatedCurrentSession
            };
        });
    },

    setLoading: (loading: boolean) => {
        chatStore.update(state => ({
            ...state,
            loading
        }));
    },

    setSessionsLoading: (sessionsLoading: boolean) => {
        chatStore.update(state => ({
            ...state,
            sessionsLoading
        }));
    },

    setStreaming: (streaming: boolean) => {
        chatStore.update(state => ({
            ...state,
            streaming
        }));
    },

    setError: (error: string | null) => {
        chatStore.update(state => ({
            ...state,
            error
        }));
    },

    reset: () => {
        chatStore.set(initialState);
    }
};