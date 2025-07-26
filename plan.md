# Shadcn-Svelte Migration Plan

## Overview

This document provides a comprehensive, step-by-step plan for migrating the AI Code Improvement Platform from custom UI components to shadcn-svelte components. The migration follows vertical slicing principles, ensuring each step delivers a complete, testable feature while maintaining all existing functionality.

## Current State Analysis

### Existing Components
- **Layout** (`src/routes/+layout.svelte`): Navigation with custom buttons, mobile menu, auth handling
- **SearchBox** (`src/lib/components/SearchBox.svelte`): Complex input with mode selector, debouncing, keyboard shortcuts
- **SearchFilters** (`src/lib/components/SearchFilters.svelte`): Filter dropdowns and active filter badges
- **CodeSnippet** (`src/lib/components/CodeSnippet.svelte`): Code display with copy functionality and syntax highlighting
- **SearchResults** (`src/lib/components/SearchResults.svelte`): Result cards with scrollable areas
- **EmbeddingStatus** (`src/lib/components/EmbeddingStatus.svelte`): Complex status tracking with progress bars
- **GitHubConnection** (`src/lib/components/GitHubConnection.svelte`): OAuth connection management with alerts

### Technology Stack
- **Frontend**: SvelteKit with Svelte 5
- **Styling**: Tailwind CSS v4
- **Runtime**: Bun
- **Package Manager**: Bun (with npm fallback for some scripts)

## Migration Strategy

### Principles
1. **Vertical Slicing**: Each phase delivers complete, testable functionality
2. **Incremental Progress**: Small, safe changes that build on each other
3. **Business Logic Preservation**: Replace UI elements, preserve all functionality
4. **Component-by-Component**: Migrate related components in logical batches
5. **Test-Driven**: Verify functionality at each step

### Risk Mitigation
- Component-level rollback capability
- Comprehensive testing before merging
- Preserved API contracts for all components

---

## Phase 1: Foundation Setup

### Prerequisites Verification
- [x] Verify current Tailwind CSS v4 setup
- [x] Confirm Svelte 5 compatibility
- [x] Backup current component implementations

### Step 1.1: Install Core Dependencies

**Task**: Install shadcn-svelte dependencies

```bash
cd frontend
npm install bits-ui clsx tailwind-merge @lucide/svelte mode-watcher
```

**Expected Outcome**: Core shadcn-svelte dependencies available
**Test Criteria**: Dependencies appear in package.json, no build errors

### Step 1.2: Initialize Shadcn-Svelte

**Task**: Run shadcn-svelte initialization

```bash
npx shadcn-svelte@latest init
```

**Configuration Choices**:
- Base color: `Slate`
- Global CSS file: `src/app.css` (will be overwritten)
- Import aliases:
  - lib: `$lib`
  - components: `$lib/components`
  - utils: `$lib/utils`
  - hooks: `$lib/hooks`
  - ui: `$lib/components/ui`

**Expected Outcome**: 
- `components.json` created
- `src/app.css` updated with shadcn-svelte styles
- `src/lib/utils.ts` created with `cn` utility

**Test Criteria**: 
- [x] Build succeeds without errors
- [x] Existing pages still render correctly
- [x] No visual regressions

### Step 1.3: Install Required Components

**Task**: Install all shadcn-svelte components needed for migration

```bash
npx shadcn-svelte@latest add button input select card badge progress alert separator sheet scroll-area
```

**Expected Outcome**: UI components available in `src/lib/components/ui/`
**Test Criteria**: 
- [x] All components installed successfully
- [x] Components can be imported without errors
- [x] Sample usage works in development

---

## Phase 2: Batch Migration

### Batch 1: Layout Foundation

**Target**: `src/routes/+layout.svelte`
**Priority**: Critical (Foundation for all other components)

#### Step 2.1.1: Pre-Migration Analysis

**Task**: Document current layout functionality

```markdown
Current Layout Elements:
- Logo/brand link (text-xl font-semibold)
- Navigation buttons (px-3 py-2 text-sm font-medium)
- Mobile menu button (hamburger icon)
- User welcome text
- Logout button
- Login button (rounded-md bg-blue-600)
- Mobile menu overlay
- Loading spinner (animate-spin)

Business Logic to Preserve:
- Authentication state management
- Route-based navigation highlighting
- Mobile menu toggle functionality
- OAuth flow handling
- Loading state management
```

**Expected Outcome**: Complete documentation of current functionality
**Test Criteria**: All UI elements and interactions documented

#### Step 2.1.2: Replace Navigation Buttons

**Task**: Replace custom navigation buttons with shadcn-svelte Button component

```svelte
<!-- Before -->
<a href="/" class="px-3 py-2 text-sm font-medium text-gray-500 hover:text-gray-700">Dashboard</a>

<!-- After -->
<Button variant="ghost" size="sm" href="/">Dashboard</Button>
```

**Implementation Steps**:
1. Import Button component: `import { Button } from "$lib/components/ui/button/index.js";`
2. Replace each navigation link with Button component
3. Use `variant="ghost"` for navigation items
4. Use `size="sm"` for compact navigation
5. Preserve all href attributes and onclick handlers

**Expected Outcome**: Navigation uses shadcn-svelte buttons
**Test Criteria**:
- [x] Navigation buttons render correctly
- [x] Hover states work as expected
- [x] Routing functionality preserved
- [x] Responsive behavior maintained

#### Step 2.1.3: Replace Mobile Menu with Sheet

**Task**: Replace custom mobile menu with shadcn-svelte Sheet component

```svelte
<!-- Before -->
{#if $authStore.isAuthenticated && mobileMenuOpen}
<div class="md:hidden" id="mobile-menu">
  <div class="space-y-1 border-t border-gray-200 bg-white px-2 pt-2 pb-3 sm:px-3">
    <!-- menu items -->
  </div>
</div>
{/if}

<!-- After -->
<Sheet.Root bind:open={mobileMenuOpen}>
  <Sheet.Trigger asChild let:builder>
    <Button builders={[builder]} variant="ghost" size="icon" class="md:hidden">
      <Menu class="h-6 w-6" />
    </Sheet.Trigger>
  </Sheet.Trigger>
  <Sheet.Content side="left">
    <Sheet.Header>
      <Sheet.Title>Navigation</Sheet.Title>
    </Sheet.Header>
    <div class="flex flex-col space-y-2 mt-4">
      <Button variant="ghost" href="/" on:click={() => mobileMenuOpen = false}>Dashboard</Button>
      <!-- other menu items -->
    </div>
  </Sheet.Content>
</Sheet.Root>
```

**Implementation Steps**:
1. Import Sheet components and Menu icon from lucide-svelte
2. Replace mobile menu button with Sheet.Trigger
3. Convert menu overlay to Sheet.Content
4. Maintain menu item functionality
5. Preserve auto-close behavior

**Expected Outcome**: Mobile menu uses shadcn-svelte Sheet
**Test Criteria**:
- [x] Mobile menu opens/closes correctly
- [x] Menu items navigate properly
- [x] Auto-close functionality works
- [x] Accessibility improvements

#### Step 2.1.4: Replace Loading Spinner

**Task**: Use shadcn-svelte loading patterns

```svelte
<!-- Before -->
<div class="inline-block h-8 w-8 animate-spin rounded-full border-4 border-solid border-blue-600 border-r-transparent"></div>

<!-- After -->
<div class="flex items-center justify-center">
  <Loader2 class="h-8 w-8 animate-spin" />
  <span class="ml-2">Authenticating...</span>
</div>
```

**Expected Outcome**: Consistent loading indicators
**Test Criteria**: Loading states display correctly

#### Step 2.1.5: Integration Testing

**Task**: Comprehensive testing of migrated layout

**Test Scenarios**:
- [x] Desktop navigation functionality
- [x] Mobile menu open/close behavior
- [x] Authentication state changes
- [x] Route highlighting
- [x] Loading states
- [x] Responsive behavior
- [x] Keyboard navigation
- [x] Screen reader compatibility

**Expected Outcome**: All layout functionality preserved with improved UI

---

### Batch 2: Search Components

**Targets**: `SearchBox.svelte` + `SearchFilters.svelte`
**Priority**: High (Core application functionality)

#### Step 2.2.1: Pre-Migration Analysis - SearchBox

**Task**: Document SearchBox functionality

```markdown
Current SearchBox Elements:
- Mode selector buttons (text/semantic/hybrid)
- Search input with debouncing (300ms)
- Search icon / loading spinner
- Clear button (X icon)
- Keyboard shortcuts (Enter, Escape)

Business Logic to Preserve:
- Debounced search dispatch
- Mode switching functionality
- Event dispatching (search, clear, focus, blur, modeChange)
- Keyboard shortcut handling
- Loading state management
- Accessibility features
```

#### Step 2.2.2: Replace Search Input

**Task**: Replace custom input with shadcn-svelte Input component

```svelte
<!-- Before -->
<input
  type="text"
  bind:value
  on:input={handleInput}
  on:keydown={handleKeydown}
  class="search-input"
  {placeholder}
  {disabled}
/>

<!-- After -->
<div class="relative">
  <Input
    type="text"
    bind:value
    on:input={handleInput}
    on:keydown={handleKeydown}
    {placeholder}
    {disabled}
    class="pl-10 pr-10"
  />
  <Search class="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
  {#if value}
    <Button
      variant="ghost"
      size="sm"
      class="absolute right-1 top-1/2 transform -translate-y-1/2 h-8 w-8 p-0"
      on:click={handleClear}
    >
      <X class="h-4 w-4" />
    </Button>
  {/if}
</div>
```

**Expected Outcome**: Search input uses shadcn-svelte styling
**Test Criteria**:
- [x] Input functionality preserved
- [x] Icons positioned correctly
- [x] Clear button works
- [x] Focus states work

#### Step 2.2.3: Replace Mode Selector Buttons

**Task**: Replace custom mode buttons with shadcn-svelte Button group

```svelte
<!-- Before -->
<div class="mode-selector" role="radiogroup">
  <button class="mode-button" class:active={searchMode === 'text'}>Text</button>
  <!-- other buttons -->
</div>

<!-- After -->
<div class="flex gap-1 p-1 bg-muted rounded-lg mb-3">
  <Button
    variant={searchMode === 'text' ? 'default' : 'ghost'}
    size="sm"
    on:click={() => handleModeChange('text')}
    {disabled}
  >
    <BookOpen class="h-4 w-4 mr-2" />
    Text
  </Button>
  <!-- other buttons -->
</div>
```

**Expected Outcome**: Mode selector uses shadcn-svelte button styling
**Test Criteria**:
- [x] Mode switching works
- [x] Active state highlighting
- [x] Icons display correctly
- [x] Accessibility preserved

#### Step 2.2.4: Pre-Migration Analysis - SearchFilters

**Task**: Document SearchFilters functionality

```markdown
Current SearchFilters Elements:
- Language filter dropdown
- Repository filter dropdown
- Active filter badges
- Clear filters button

Business Logic to Preserve:
- Filter state management
- Event dispatching
- Badge creation/removal
- Clear all functionality
```

#### Step 2.2.5: Replace Filter Dropdowns

**Task**: Replace custom selects with shadcn-svelte Select component

```svelte
<!-- Before -->
<select bind:value={selectedLanguage} on:change={handleLanguageChange}>
  <option value="">All Languages</option>
  {#each languages as language}
    <option value={language}>{language}</option>
  {/each}
</select>

<!-- After -->
<Select.Root bind:value={selectedLanguage} onValueChange={handleLanguageChange}>
  <Select.Trigger class="w-48">
    <Select.Value placeholder="All Languages" />
  </Select.Trigger>
  <Select.Content>
    <Select.Item value="">All Languages</Select.Item>
    {#each languages as language}
      <Select.Item value={language}>{language}</Select.Item>
    {/each}
  </Select.Content>
</Select.Root>
```

**Expected Outcome**: Filter dropdowns use shadcn-svelte Select
**Test Criteria**:
- [x] Filter selection works
- [x] Options display correctly
- [x] Event dispatching preserved

#### Step 2.2.6: Replace Filter Badges

**Task**: Replace custom badges with shadcn-svelte Badge component

```svelte
<!-- Before -->
<span class="filter-badge">
  {filterName}
  <button on:click={() => removeFilter(filterName)}>×</button>
</span>

<!-- After -->
<Badge variant="secondary" class="gap-1">
  {filterName}
  <Button
    variant="ghost"
    size="sm"
    class="h-4 w-4 p-0 hover:bg-destructive hover:text-destructive-foreground"
    on:click={() => removeFilter(filterName)}
  >
    <X class="h-3 w-3" />
  </Button>
</Badge>
```

**Expected Outcome**: Filter badges use shadcn-svelte Badge component
**Test Criteria**:
- [x] Badges display correctly
- [x] Remove functionality works
- [x] Hover states appropriate

#### Step 2.2.7: Integration Testing - Search Components

**Task**: Test complete search functionality

**Test Scenarios**:
- [x] Search input debouncing (300ms delay)
- [x] Mode selector functionality
- [x] Filter selection and clearing
- [x] Keyboard shortcuts (Enter, Escape)
- [x] Loading states during search
- [x] Event dispatching to parent components
- [x] Responsive behavior
- [x] Screen reader support

**Expected Outcome**: All search functionality preserved with improved UI

---

### Batch 3: Data Display Components

**Targets**: `CodeSnippet.svelte` + `SearchResults.svelte`
**Priority**: Medium (Content presentation)

#### Step 2.3.1: Pre-Migration Analysis - CodeSnippet

**Task**: Document CodeSnippet functionality

```markdown
Current CodeSnippet Elements:
- Code container card
- Language badge
- Copy button with success feedback
- Syntax highlighted content
- Line numbers (optional)

Business Logic to Preserve:
- Syntax highlighting (highlight.js)
- Copy to clipboard functionality
- Language detection
- Line number display
- Responsive text sizing
```

#### Step 2.3.2: Replace Code Container

**Task**: Replace custom container with shadcn-svelte Card

```svelte
<!-- Before -->
<div class="code-snippet-container">
  <div class="code-header">
    <span class="language-badge">{language}</span>
    <button class="copy-button" on:click={copyToClipboard}>Copy</button>
  </div>
  <div class="code-content">
    <pre><code>{@html highlightedCode}</code></pre>
  </div>
</div>

<!-- After -->
<Card.Root class="overflow-hidden">
  <Card.Header class="flex flex-row items-center justify-between space-y-0 pb-2">
    <Badge variant="outline" class="text-xs">
      {language || 'text'}
    </Badge>
    <Button
      variant="outline"
      size="sm"
      on:click={copyToClipboard}
      disabled={copying}
    >
      {#if copying}
        <Check class="h-4 w-4 mr-2" />
        Copied!
      {:else}
        <Copy class="h-4 w-4 mr-2" />
        Copy
      {/if}
    </Button>
  </Card.Header>
  <Card.Content class="p-0">
    <ScrollArea class="h-full max-h-96">
      <pre class="p-4 text-sm"><code>{@html highlightedCode}</code></pre>
    </ScrollArea>
  </Card.Content>
</Card.Root>
```

**Expected Outcome**: Code snippets use shadcn-svelte Card layout
**Test Criteria**:
- [x] Card structure displays correctly
- [x] Copy functionality works with feedback
- [x] Scrolling works for long code
- [x] Syntax highlighting preserved

#### Step 2.3.3: Pre-Migration Analysis - SearchResults

**Task**: Document SearchResults functionality

```markdown
Current SearchResults Elements:
- Results container with scrolling
- Individual result cards
- File path display
- Code preview
- Action buttons (view, open)
- Loading and empty states

Business Logic to Preserve:
- Result rendering and virtualization
- Action button functionality
- Loading state management
- Empty state handling
- Infinite scroll (if implemented)
```

#### Step 2.3.4: Replace Results Container

**Task**: Replace custom container with shadcn-svelte ScrollArea and Cards

```svelte
<!-- Before -->
<div class="search-results">
  {#each results as result}
    <div class="result-card">
      <div class="result-header">
        <span class="file-path">{result.filePath}</span>
        <button class="view-button">View</button>
      </div>
      <div class="result-content">{result.content}</div>
    </div>
  {/each}
</div>

<!-- After -->
<ScrollArea class="h-[600px] w-full rounded-md border">
  <div class="space-y-4 p-4">
    {#each results as result}
      <Card.Root>
        <Card.Header class="pb-3">
          <div class="flex items-center justify-between">
            <div class="flex items-center space-x-2">
              <FileText class="h-4 w-4 text-muted-foreground" />
              <span class="text-sm font-medium truncate">{result.filePath}</span>
            </div>
            <Button variant="outline" size="sm" on:click={() => viewResult(result)}>
              <ExternalLink class="h-4 w-4 mr-2" />
              View
            </Button>
          </div>
        </Card.Header>
        <Card.Content>
          <pre class="text-xs text-muted-foreground whitespace-pre-wrap">{result.content}</pre>
        </Card.Content>
      </Card.Root>
    {/each}
  </div>
</ScrollArea>
```

**Expected Outcome**: Search results use shadcn-svelte Cards and ScrollArea
**Test Criteria**:
- [x] Results display in card format
- [x] Scrolling works properly
- [x] Action buttons function correctly
- [x] File paths display with proper truncation

#### Step 2.3.5: Integration Testing - Display Components

**Test Scenarios**:
- [x] Code syntax highlighting accuracy
- [x] Copy functionality with success feedback
- [x] Search result rendering
- [x] Scrolling performance with many results
- [x] Responsive layout on mobile
- [x] Loading states during search
- [x] Empty state handling

**Expected Outcome**: All display functionality preserved with improved UI

---

### Batch 4: Status & Connection Components

**Targets**: `EmbeddingStatus.svelte` + `GitHubConnection.svelte`
**Priority**: Medium (Complex state management)

#### Step 2.4.1: Pre-Migration Analysis - EmbeddingStatus

**Task**: Document EmbeddingStatus functionality

```markdown
Current EmbeddingStatus Elements:
- Status card container
- Progress bar
- Status badges (pending, processing, completed, failed)
- Action buttons (trigger, retry)
- Statistics display
- Auto-refresh functionality

Business Logic to Preserve:
- Real-time status polling (5s intervals)
- Progress calculation
- Status state management
- Auto-refresh lifecycle
- Event dispatching
- Error handling
```

#### Step 2.4.2: Replace Status Container

**Task**: Replace custom container with shadcn-svelte Card and Progress

```svelte
<!-- Before -->
<div class="embedding-status-card">
  <div class="status-header">
    <h3>Embedding Status</h3>
    <span class="status-badge">{status.status}</span>
  </div>
  <div class="progress-container">
    <div class="progress-bar" style="width: {status.progress}%"></div>
  </div>
  <div class="status-actions">
    <button on:click={triggerEmbedding}>Re-embed</button>
  </div>
</div>

<!-- After -->
<Card.Root>
  <Card.Header>
    <Card.Title class="flex items-center justify-between">
      Embedding Status
      <Badge variant={getStatusVariant(status.status)}>
        {getStatusIcon(status.status)}
        {status.status}
      </Badge>
    </Card.Title>
  </Card.Header>
  <Card.Content class="space-y-4">
    <div class="space-y-2">
      <div class="flex justify-between text-sm">
        <span>Progress</span>
        <span>{status.processedChunks || 0} / {status.totalChunks || 0}</span>
      </div>
      <Progress value={status.progress} class="h-2" />
    </div>
    <Separator />
    <div class="flex justify-between items-center">
      <div class="text-sm text-muted-foreground">
        {#if status.estimatedTimeRemaining}
          {Math.ceil(status.estimatedTimeRemaining / 60)}m remaining
        {/if}
      </div>
      <Button
        variant="outline"
        size="sm"
        on:click={triggerEmbedding}
        disabled={loading || status.status === 'processing'}
      >
        {#if loading}
          <Loader2 class="h-4 w-4 mr-2 animate-spin" />
        {:else}
          <RefreshCw class="h-4 w-4 mr-2" />
        {/if}
        Re-embed
      </Button>
    </div>
  </Card.Content>
</Card.Root>
```

**Expected Outcome**: Embedding status uses shadcn-svelte Card and Progress
**Test Criteria**:
- [x] Status display updates correctly
- [x] Progress bar reflects actual progress
- [x] Status badges show appropriate variants
- [x] Re-embed button works properly
- [x] Auto-refresh continues working

#### Step 2.4.3: Pre-Migration Analysis - GitHubConnection

**Task**: Document GitHubConnection functionality

```markdown
Current GitHubConnection Elements:
- Connection status card
- Connect/disconnect buttons
- Success/error alerts
- OAuth flow handling
- Loading states

Business Logic to Preserve:
- OAuth callback handling  
- Connection state management
- Success/error messaging
- URL cleanup after OAuth
- User data updates
```

#### Step 2.4.4: Replace Connection Container

**Task**: Replace custom container with shadcn-svelte Card and Alert

```svelte
<!-- Before -->
<div class="github-connection-card">
  <div class="connection-status">
    {#if user.githubConnected}
      <p>Connected as {user.githubUsername}</p>
      <button on:click={disconnectGitHub}>Disconnect</button>
    {:else}
      <p>Connect your GitHub account</p>
      <button on:click={connectGitHub}>Connect</button>
    {/if}
  </div>
  {#if error}
    <div class="error-message">{error}</div>
  {/if}
  {#if success}
    <div class="success-message">{success}</div>
  {/if}
</div>

<!-- After -->
<Card.Root>
  <Card.Header>
    <Card.Title class="flex items-center gap-2">
      <Github class="h-5 w-5" />
      GitHub Connection
    </Card.Title>
  </Card.Header>
  <Card.Content class="space-y-4">
    {#if user.githubConnected}
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <div class="h-2 w-2 bg-green-500 rounded-full"></div>
          <span class="text-sm">Connected as <strong>{user.githubUsername}</strong></span>
        </div>
        <Button
          variant="outline"
          size="sm"
          on:click={disconnectGitHub}
          disabled={connecting}
        >
          {#if connecting}
            <Loader2 class="h-4 w-4 mr-2 animate-spin" />
          {:else}
            <Unlink class="h-4 w-4 mr-2" />
          {/if}
          Disconnect
        </Button>
      </div>
    {:else}
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <div class="h-2 w-2 bg-gray-400 rounded-full"></div>
          <span class="text-sm text-muted-foreground">Not connected</span>
        </div>
        <Button
          on:click={connectGitHub}
          disabled={connecting}
        >
          {#if connecting}
            <Loader2 class="h-4 w-4 mr-2 animate-spin" />
          {:else}
            <Link class="h-4 w-4 mr-2" />
          {/if}
          Connect GitHub
        </Button>
      </div>
    {/if}

    {#if error}
      <Alert.Root variant="destructive">
        <AlertCircle class="h-4 w-4" />
        <Alert.Title>Connection Error</Alert.Title>
        <Alert.Description>{error}</Alert.Description>
      </Alert.Root>
    {/if}

    {#if success}
      <Alert.Root>
        <Check class="h-4 w-4" />
        <Alert.Title>Success</Alert.Title>
        <Alert.Description>{success}</Alert.Description>
      </Alert.Root>
    {/if}
  </Card.Content>
</Card.Root>
```

**Expected Outcome**: GitHub connection uses shadcn-svelte Card and Alert
**Test Criteria**:
- [x] Connection status displays correctly
- [x] Connect/disconnect buttons work
- [x] OAuth flow completes successfully
- [x] Success/error alerts display
- [x] URL cleanup after OAuth

#### Step 2.4.5: Integration Testing - Status Components

**Test Scenarios**:
- [x] Embedding status polling and updates
- [x] Progress bar accuracy during processing
- [x] Status badge variants for different states
- [x] Re-embed functionality
- [x] GitHub OAuth flow end-to-end
- [x] Connection state persistence  
- [x] Error handling and display
- [x] Success message display and timeout

**Expected Outcome**: All status functionality preserved with improved UI

---

## Phase 3: Testing & Integration

### Step 3.1: Cross-Component Integration Testing

**Task**: Test interactions between migrated components

**Test Scenarios**:
- [ ] Search components work with results display
- [ ] Layout navigation updates search filters
- [ ] Status components update after GitHub connection
- [ ] Mobile responsive behavior across all components
- [ ] Keyboard navigation throughout the app
- [ ] Screen reader compatibility

### Step 3.2: Performance Testing

**Task**: Ensure no performance regressions

**Metrics to Monitor**:
- [ ] Bundle size impact
- [ ] Initial page load time
- [ ] Component render performance
- [ ] Memory usage during interactions

### Step 3.3: Accessibility Audit

**Task**: Verify accessibility improvements

**Checks**:
- [ ] WCAG 2.1 AA compliance
- [ ] Screen reader compatibility
- [ ] Keyboard navigation
- [ ] Color contrast ratios
- [ ] Focus management

### Step 3.4: Browser Compatibility

**Task**: Test across supported browsers

**Browser Matrix**:
- [ ] Chrome (latest)
- [ ] Firefox (latest)
- [ ] Safari (latest)
- [ ] Edge (latest)

---

## Phase 4: Documentation & Cleanup

### Step 4.1: Update Component Documentation

**Task**: Document new component APIs

**Updates Required**:
- [ ] Update component prop interfaces
- [ ] Document new event handlers
- [ ] Update usage examples
- [ ] Add migration notes

### Step 4.2: Code Cleanup

**Task**: Remove unused code and styles

**Cleanup Items**:
- [ ] Remove custom CSS styles that are now handled by shadcn-svelte
- [ ] Remove unused utility functions
- [ ] Update import statements
- [ ] Clean up commented code

### Step 4.3: Update Tests

**Task**: Update test suites for new components

**Test Updates**:
- [ ] Update component selectors
- [ ] Adjust interaction patterns
- [ ] Update snapshot tests
- [ ] Add new accessibility tests

---

## Implementation Prompts

### Prompt 1: Phase 1 Setup

```
I need to set up shadcn-svelte in my SvelteKit project. Please help me:

1. Install the required dependencies (bits-ui, clsx, tailwind-merge, @lucide/svelte, mode-watcher)
2. Initialize shadcn-svelte with the following configuration:
   - Base color: Slate
   - Global CSS file: src/app.css
   - Import aliases: $lib, $lib/components, $lib/utils, $lib/hooks, $lib/components/ui
3. Install these components: button, input, select, card, badge, progress, alert, separator, sheet, scroll-area
4. Verify the setup works by testing that components can be imported

Current project uses SvelteKit with Svelte 5, Tailwind CSS v4, and Bun as the package manager.
```

### Prompt 2: Layout Migration

```
I need to migrate the layout component from custom styling to shadcn-svelte components. Please help me:

1. Replace navigation buttons with shadcn-svelte Button components (variant="ghost", size="sm")
2. Replace the mobile menu with shadcn-svelte Sheet component
3. Replace the loading spinner with Lucide icons
4. Preserve all existing functionality:
   - Authentication state management
   - Route-based navigation
   - Mobile menu toggle
   - Loading states

The current layout is in src/routes/+layout.svelte. Preserve all business logic and only replace UI elements.
```

### Prompt 3: Search Components Migration

```
I need to migrate SearchBox and SearchFilters components to use shadcn-svelte. Please help me:

SearchBox migration:
1. Replace the search input with shadcn-svelte Input component
2. Replace mode selector buttons with shadcn-svelte Button components
3. Add appropriate Lucide icons (Search, X for clear)
4. Preserve debounced search functionality (300ms)
5. Preserve keyboard shortcuts (Enter, Escape)

SearchFilters migration:
1. Replace select dropdowns with shadcn-svelte Select components  
2. Replace filter badges with shadcn-svelte Badge components
3. Preserve all filter state management and event dispatching

Maintain all existing business logic and component APIs.
```

### Prompt 4: Display Components Migration

```
I need to migrate CodeSnippet and SearchResults components to use shadcn-svelte. Please help me:

CodeSnippet migration:
1. Replace container with shadcn-svelte Card component
2. Use Badge for language display
3. Use Button for copy functionality with success feedback
4. Use ScrollArea for long code snippets
5. Preserve syntax highlighting and copy functionality

SearchResults migration:
1. Replace results container with shadcn-svelte ScrollArea
2. Replace result items with shadcn-svelte Card components
3. Use appropriate Lucide icons (FileText, ExternalLink)
4. Preserve all result rendering and action functionality

Maintain all existing business logic and performance characteristics.
```

### Prompt 5: Status Components Migration

```
I need to migrate EmbeddingStatus and GitHubConnection components to use shadcn-svelte. Please help me:

EmbeddingStatus migration:
1. Replace container with shadcn-svelte Card component
2. Use Progress component for embedding progress
3. Use Badge for status indicators with appropriate variants
4. Use Button for action buttons with loading states
5. Preserve auto-refresh functionality and status polling

GitHubConnection migration:
1. Replace container with shadcn-svelte Card component
2. Use Alert components for success/error messages
3. Use Button for connect/disconnect actions
4. Add appropriate Lucide icons (Github, Link, Unlink)
5. Preserve OAuth flow and connection state management

Maintain all existing business logic and state management.
```

### Prompt 6: Integration Testing

```
I need to test the completed shadcn-svelte migration. Please help me:

1. Create comprehensive tests for all migrated components
2. Test cross-component interactions
3. Verify no functionality regressions
4. Check responsive behavior
5. Validate accessibility improvements
6. Test performance impact

Focus on ensuring all business logic is preserved while confirming the UI improvements are working correctly.
```

---

## Success Criteria

### Functional Requirements
- [ ] All existing functionality preserved
- [ ] No regression in user experience  
- [ ] Performance maintained or improved
- [ ] Accessibility compliance maintained or improved

### Design Requirements
- [ ] Consistent visual design using shadcn-svelte
- [ ] Proper responsive design on all screen sizes
- [ ] Modern, professional appearance
- [ ] Improved interaction patterns

### Development Requirements
- [ ] Reduced custom CSS codebase (80%+ reduction expected)
- [ ] Improved component maintainability
- [ ] Better design system consistency
- [ ] Easier future component additions
- [ ] All tests passing

## Timeline Estimate

| Phase | Duration | Deliverables |
|-------|----------|--------------|
| Phase 1: Foundation Setup | 1-2 days | Dependencies installed, shadcn-svelte initialized |
| Batch 1: Layout Migration | 2-3 days | Migrated navigation and layout |
| Batch 2: Search Components | 3-4 days | Migrated SearchBox and SearchFilters |
| Batch 3: Display Components | 3-4 days | Migrated CodeSnippet and SearchResults |
| Batch 4: Status Components | 4-5 days | Migrated EmbeddingStatus and GitHubConnection |
| Phase 3: Testing & Integration | 2-3 days | Comprehensive testing complete |
| Phase 4: Documentation & Cleanup | 1-2 days | Documentation updated, code cleaned |

**Total Estimate: 16-23 days**

## Risk Mitigation

### Identified Risks
1. **Breaking changes in component APIs** - Mitigated by comprehensive testing at each step
2. **Performance regressions** - Mitigated by performance monitoring and bundle analysis
3. **Accessibility issues** - Mitigated by leveraging shadcn-svelte's built-in accessibility
4. **Styling conflicts** - Mitigated by proper CSS cascade management

### Rollback Strategy
- Component-level rollback capability maintained
- No database or API changes required
- Feature flags not needed due to in-place replacement approach

---

*This migration plan provides a comprehensive roadmap for safely migrating from custom components to shadcn-svelte while maintaining all functionality and improving the user experience.*