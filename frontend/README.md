# GitHub Analyzer Frontend

[![vite](https://img.shields.io/badge/SvelteKit-%23ff3e00.svg?logo=svelte&logoColor=white)](https://kit.svelte.dev/)  
[![bun](https://img.shields.io/badge/Bun-1.0-blue?logo=bun)](https://bun.sh/)

## Table of Contents

- [Features](#features)
- [Development Setup](#development-setup)
  - [Prerequisites](#prerequisites)
  - [Installation](#installation)
  - [Environment Configuration](#environment-configuration)
  - [Development Server](#development-server)
  - [API Type Generation](#api-type-generation)
- [Building](#building)
- [Project Structure](#project-structure)
- [Key Technologies](#key-technologies)
- [API Integration](#api-integration)
- [Testing](#testing)

## Testing

```bash
# Unit tests (Vitest)
bun run test:unit

# Component tests (Playwright component mode)
bun run test:ct

# End-to-end tests (Playwright)
bun run test:e2e
```

AI-powered code analysis platform frontend built with SvelteKit and TypeScript. Features type-safe API integration with OpenAPI code generation.

## Features

- 🔐 **JWT Authentication** with automatic token management
- 📊 **Dashboard** with repository analytics and metrics
- 🔗 **GitHub Integration** with OAuth authentication
- 💬 **AI Chat Interface** for code analysis (backend integration pending)
- 📱 **Responsive Design** with TailwindCSS
- ⚡ **Type Safety** with OpenAPI-generated types
- 🔧 **Real-time Data** from backend API

## Development Setup

### Prerequisites

- Node.js 18+ or Bun
- Backend API server running (typically on port 8080)

### Installation

```bash
# Install dependencies
bun install

# Copy environment file and configure
cp .env.example .env.local
# Edit .env.local with your backend API URL
```

### Environment Configuration

Copy `.env.example` to `.env.local` and configure:

```env
# Backend API URL
VITE_API_URL=http://localhost:8080
```

### Development Server

```bash
# Start development server
bun run dev

# Or with auto-open browser
bun run dev -- --open
```

### API Type Generation

The frontend uses OpenAPI-generated types for complete type safety:

```bash
# Regenerate API types from backend OpenAPI spec
bun run generate-api
```

## Building

```bash
# Build for production
bun run build

# Preview production build
bun run preview
```

## Project Structure

```
src/
├── lib/
│   ├── api/           # Generated OpenAPI client and types
│   ├── components/    # Reusable Svelte components
│   └── stores/        # Svelte stores for state management
├── routes/            # SvelteKit file-based routing
│   ├── auth/         # Authentication pages
│   ├── chat/         # AI chat interface
│   └── repositories/ # Repository management
└── app.html          # Main HTML template
```

## Key Technologies

- **SvelteKit** - Frontend framework with file-based routing
- **TypeScript** - Type safety throughout the application
- **TailwindCSS** - Utility-first CSS framework
- **openapi-fetch** - Type-safe HTTP client
- **openapi-typescript** - API type generation
- **Chart.js** - Data visualization

## API Integration

This frontend is designed to work with the backend API and features:

- Complete OpenAPI integration with generated types
- Automatic JWT token handling
- Error handling with proper HTTP status codes
- Real-time data synchronization
