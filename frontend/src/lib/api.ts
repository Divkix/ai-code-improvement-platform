// ABOUTME: Legacy API client - now re-exports from generated client
// ABOUTME: Maintained for backward compatibility during migration

// Re-export everything from the new generated API
export * from './api/index';
export { apiClient as default, apiClient } from './api/client';
export * from './api/hooks';
