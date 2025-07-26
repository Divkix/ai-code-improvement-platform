// ABOUTME: MSW worker setup for browser testing environment
// ABOUTME: Configures mock service worker for intercepting API calls during browser tests

import { setupWorker } from 'msw/browser';
import { handlers } from './handlers';

// Setup worker with default handlers
export const worker = setupWorker(...handlers);
