// ABOUTME: Chat API client with Server-Sent Events streaming support
// ABOUTME: Handles chat sessions, messages, and real-time streaming responses

import { authStore } from '$lib/stores/auth';
import type { components } from './types';

type ChatSession = components['schemas']['ChatSession'];
type ChatSessionListResponse = components['schemas']['ChatSessionListResponse'];
type CreateChatSessionRequest = components['schemas']['CreateChatSessionRequest'];
type UpdateChatSessionRequest = components['schemas']['UpdateChatSessionRequest'];
type SendMessageRequest = components['schemas']['SendMessageRequest'];

// Use the same base URL mechanism as the rest of the frontend so that requests
// always target the backend instead of the SvelteKit dev server. We append
// `/api` here so that all existing endpoint paths continue to work unchanged.
const API_BASE_URL = `${(import.meta.env.VITE_API_URL || 'http://localhost:8080').replace(/\/$/, '')}/api`;

class ChatAPIError extends Error {
    constructor(
        message: string,
        public status?: number,
        public code?: string
    ) {
        super(message);
        this.name = 'ChatAPIError';
    }
}

function getAuthHeaders(): Record<string, string> {
    return {
        'Content-Type': 'application/json',
        ...(authStore.current.token ? { Authorization: `Bearer ${authStore.current.token}` } : {})
    };
}

export interface ChatStreamChunk {
    type: 'content' | 'done' | 'error';
    content: string;
    delta?: string;
}

export class ChatClient {
    async listSessions(params?: {
        limit?: number;
        offset?: number;
    }): Promise<ChatSessionListResponse> {
        // Build an absolute URL that points to the backend instead of relying on
        // the current window origin.
        const url = new URL(`${API_BASE_URL}/chat/sessions`);
        if (params?.limit) url.searchParams.set('limit', params.limit.toString());
        if (params?.offset) url.searchParams.set('offset', params.offset.toString());

        const response = await fetch(url.toString(), {
            headers: getAuthHeaders()
        });

        if (!response.ok) {
            const error = await response.json().catch(() => ({ message: 'Failed to fetch sessions' }));
            throw new ChatAPIError(
                error.message || 'Failed to fetch chat sessions',
                response.status,
                error.error
            );
        }

        return response.json();
    }

    async createSession(request?: CreateChatSessionRequest): Promise<ChatSession> {
        const response = await fetch(`${API_BASE_URL}/chat/sessions`, {
            method: 'POST',
            headers: getAuthHeaders(),
            body: JSON.stringify(request || {})
        });

        if (!response.ok) {
            const error = await response.json().catch(() => ({ message: 'Failed to create session' }));
            throw new ChatAPIError(
                error.message || 'Failed to create chat session',
                response.status,
                error.error
            );
        }

        return response.json();
    }

    async getSession(sessionId: string): Promise<ChatSession> {
        const response = await fetch(`${API_BASE_URL}/chat/sessions/${sessionId}`, {
            headers: getAuthHeaders()
        });

        if (!response.ok) {
            const error = await response.json().catch(() => ({ message: 'Failed to fetch session' }));
            throw new ChatAPIError(
                error.message || 'Failed to fetch chat session',
                response.status,
                error.error
            );
        }

        return response.json();
    }

    async deleteSession(sessionId: string): Promise<void> {
        const response = await fetch(`${API_BASE_URL}/chat/sessions/${sessionId}`, {
            method: 'DELETE',
            headers: getAuthHeaders()
        });

        if (!response.ok) {
            const error = await response.json().catch(() => ({ message: 'Failed to delete session' }));
            throw new ChatAPIError(
                error.message || 'Failed to delete chat session',
                response.status,
                error.error
            );
        }
    }

    async updateSession(sessionId: string, request: UpdateChatSessionRequest): Promise<ChatSession> {
        const response = await fetch(`${API_BASE_URL}/chat/sessions/${sessionId}`, {
            method: 'PATCH',
            headers: getAuthHeaders(),
            body: JSON.stringify(request)
        });

        if (!response.ok) {
            const error = await response.json().catch(() => ({ message: 'Failed to update session' }));
            throw new ChatAPIError(
                error.message || 'Failed to update chat session',
                response.status,
                error.error
            );
        }

        return response.json();
    }

    async sendMessage(sessionId: string, content: string): Promise<ChatSession> {
        const request: SendMessageRequest = { content };

        const response = await fetch(`${API_BASE_URL}/chat/sessions/${sessionId}/message`, {
            method: 'POST',
            headers: getAuthHeaders(),
            body: JSON.stringify(request)
        });

        if (!response.ok) {
            const error = await response.json().catch(() => ({ message: 'Failed to send message' }));
            throw new ChatAPIError(
                error.message || 'Failed to send message',
                response.status,
                error.error
            );
        }

        return response.json();
    }

    async sendMessageStream(
        sessionId: string,
        content: string,
        onChunk: (chunk: ChatStreamChunk) => void
    ): Promise<void> {
        const request: SendMessageRequest = { content };

        const response = await fetch(`${API_BASE_URL}/chat/sessions/${sessionId}/message`, {
            method: 'POST',
            headers: {
                ...getAuthHeaders(),
                Accept: 'text/event-stream'
            },
            body: JSON.stringify(request)
        });

        if (!response.ok) {
            const error = await response.json().catch(() => ({ message: 'Failed to send message' }));
            throw new ChatAPIError(
                error.message || 'Failed to send message',
                response.status,
                error.error
            );
        }

        if (!response.body) {
            throw new ChatAPIError('No response body received');
        }

        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let buffer = ''; // Buffer to accumulate partial lines

        try {
            while (true) {
                const { done, value } = await reader.read();
                if (done) break;

                // Append decoded chunk to buffer
                buffer += decoder.decode(value, { stream: true });

                // Process complete lines from buffer
                let lineEndIndex: number;
                while ((lineEndIndex = buffer.indexOf('\n')) !== -1) {
                    // Extract complete line
                    const line = buffer.slice(0, lineEndIndex).trim();
                    // Remove processed line from buffer
                    buffer = buffer.slice(lineEndIndex + 1);

                    // Skip empty lines
                    if (!line) continue;

                    try {
                        if (line.startsWith('data: ')) {
                            const data = line.slice(6).trim();
                            if (!data) continue;

                            try {
                                const parsed = JSON.parse(data) as ChatStreamChunk;
                                onChunk(parsed);
                            } catch (parseError) {
                                console.warn('Failed to parse SSE data:', data, parseError);
                            }
                        } else if (line.startsWith('event: ')) {
                            const event = line.slice(7).trim();
                            if (event === 'done') {
                                onChunk({ type: 'done', content: '' });
                                return;
                            } else if (event === 'error') {
                                onChunk({ type: 'error', content: 'Stream error occurred' });
                                return;
                            }
                        }
                    } catch (error) {
                        console.error('Error processing SSE line:', line, error);
                    }
                }
            }

            // Process any remaining data in buffer after stream ends
            if (buffer.trim()) {
                const line = buffer.trim();
                try {
                    if (line.startsWith('data: ')) {
                        const data = line.slice(6).trim();
                        if (data) {
                            const parsed = JSON.parse(data) as ChatStreamChunk;
                            onChunk(parsed);
                        }
                    }
                } catch (error) {
                    console.warn('Failed to parse remaining buffer:', buffer, error);
                }
            }
        } catch (error) {
            console.error('Stream reading error:', error);
            throw new ChatAPIError(
                error instanceof Error ? error.message : 'Stream processing failed',
                undefined,
                'STREAM_ERROR'
            );
        } finally {
            // Always release the reader lock
            try {
                reader.releaseLock();
            } catch (lockError) {
                console.warn('Failed to release reader lock:', lockError);
            }
        }
    }

    async sendMessageWithEventSource(
        sessionId: string,
        content: string,
        onChunk: (chunk: ChatStreamChunk) => void,
        onError: (error: Error) => void
    ): Promise<void> {
        const request: SendMessageRequest = { content };

        // Create a form data for sending the request
        const formData = new FormData();
        formData.append('data', JSON.stringify(request));

        const eventSourceUrl = new URL(
            `${API_BASE_URL}/chat/sessions/${sessionId}/message`,
            window.location.origin
        );

        // For EventSource, we need to handle the authentication differently
        // We'll send the auth token as a query parameter since EventSource doesn't support custom headers
        if (authStore.current.token) {
            eventSourceUrl.searchParams.set('token', authStore.current.token);
        }

        try {
            // First send the message via POST
            const response = await fetch(`${API_BASE_URL}/chat/sessions/${sessionId}/message`, {
                method: 'POST',
                headers: {
                    ...getAuthHeaders(),
                    Accept: 'text/event-stream'
                },
                body: JSON.stringify(request)
            });

            if (!response.ok) {
                const error = await response.json().catch(() => ({ message: 'Failed to send message' }));
                throw new ChatAPIError(
                    error.message || 'Failed to send message',
                    response.status,
                    error.error
                );
            }

            // Handle streaming response
            return this.sendMessageStream(sessionId, content, onChunk);
        } catch (error) {
            onError(error instanceof Error ? error : new Error('Unknown error occurred'));
        }
    }
}

export const chatClient = new ChatClient();
export { ChatAPIError };
