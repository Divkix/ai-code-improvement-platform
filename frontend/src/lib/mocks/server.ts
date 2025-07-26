// ABOUTME: MSW server setup for Node.js testing environment
// ABOUTME: Configures mock service worker for intercepting API calls during server-side tests

import { setupServer } from 'msw/node';
import { handlers } from './handlers';

// Setup server with default handlers
export const server = setupServer(...handlers);
