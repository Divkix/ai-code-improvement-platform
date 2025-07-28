# Frontend – AI Code Fixing Platform

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

## Overview

AI-powered automated code fixing platform frontend built with SvelteKit and TypeScript. Transforms from "smart text search" into an automated code fixing engine that generates complete, validated solutions for technical debt and code issues.

## Features

- 🔐 **JWT Authentication** with automatic token management
- 📊 **Analytics Dashboard** with fix generation metrics and cost savings
- 🔗 **GitHub Integration** with OAuth authentication and repository import
- 🤖 **AI Fix Generation Interface** for automated code fixing with AST-powered analysis
- 🛠️ **Code Issue Detection** with structural and semantic analysis
- 💻 **Fix Validation Display** showing syntax, compilation, and test results
- 📱 **Responsive Design** with TailwindCSS
- ⚡ **Type Safety** with OpenAPI-generated types
- 🔧 **Real-time Progress** tracking for repository analysis and fix generation

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

This frontend interfaces with the AI Code Fixing Platform backend API:

- **Complete OpenAPI integration** with generated types for type-safe API communication
- **Automatic JWT token handling** for secure authentication
- **Error handling** with proper HTTP status codes and user feedback
- **Real-time progress tracking** for repository analysis and fix generation
- **Fix validation results** displaying syntax, compilation, and test validation
- **AST-based analysis** results with code structure visualization
- **Multi-modal context** showing code, comments, tests, and documentation

## Testing

```bash
# Unit tests (Vitest)
bun run test:unit

# Component tests (Playwright component mode)
bun run test:ct

# End-to-end tests (Playwright)
bun run test:e2e
```
