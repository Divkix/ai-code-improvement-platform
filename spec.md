# Shadcn-Svelte Migration Specification

## Project Overview
This specification outlines the complete migration of custom UI components in the AI Code Improvement Platform to shadcn-svelte components. The project uses Svelte 5 with Tailwind CSS v4 and will adopt shadcn-svelte's design system for consistency and maintainability.

## Current Technology Stack
- **Frontend**: SvelteKit with Svelte 5
- **Styling**: Tailwind CSS v4
- **Runtime**: Bun
- **Current Components**: 7 custom components with extensive CSS

## Migration Strategy

### Overall Approach
- **Component-by-component migration** in small batches (2-3 related components)
- **In-place replacement** of existing components
- **Element-by-element replacement** within each component (preserve business logic, replace UI elements)
- **Full shadcn-svelte design system adoption** (colors, spacing, typography)

### Key Decisions
1. ✅ **Replace entire components** with shadcn-svelte equivalents where possible
2. ✅ **Adopt shadcn-svelte's default design system** completely
3. ✅ **Component-by-component migration** for faster visible results
4. ✅ **Small batches** of 2-3 related components per iteration
5. ✅ **In-place replacement** without compatibility layers
6. ✅ **Full initialization** of shadcn-svelte system upfront
7. ✅ **Layout first** approach for maximum visual impact

## Component Analysis & Mapping

### Current Custom Components
| Component | File | Primary UI Elements | Complexity |
|-----------|------|-------------------|------------|
| SearchBox | `src/lib/components/SearchBox.svelte` | Input, Button, Select, Spinner | Medium |
| SearchFilters | `src/lib/components/SearchFilters.svelte` | Select, Button, Badge | Medium |
| CodeSnippet | `src/lib/components/CodeSnippet.svelte` | Card, Button, Badge | Medium |
| SearchResults | `src/lib/components/SearchResults.svelte` | Card, Button, ScrollArea | Medium |
| EmbeddingStatus | `src/lib/components/EmbeddingStatus.svelte` | Card, Progress, Button, Badge, Separator | Complex |
| GitHubConnection | `src/lib/components/GitHubConnection.svelte` | Card, Button, Alert | Medium |
| Layout | `src/routes/+layout.svelte` | Button, Separator, Sheet | Medium |

### Shadcn-Svelte Component Mapping
| Custom Element | Shadcn-Svelte Component | Notes |
|----------------|-------------------------|-------|
| Custom input fields | `Input` | Standard text inputs with validation |
| Custom buttons | `Button` | Multiple variants (default, secondary, destructive) |
| Custom select dropdowns | `Select` | With proper keyboard navigation |
| Custom cards | `Card` | Header, content, footer structure |
| Custom badges/tags | `Badge` | Status indicators and filters |
| Custom progress bars | `Progress` | For embedding status |
| Custom alerts | `Alert` | Success/error messages |
| Custom separators | `Separator` | Visual dividers |
| Mobile menu | `Sheet` | Slide-out navigation |
| Scrollable areas | `ScrollArea` | Enhanced scrolling |
| Custom spinners | Built-in loading states | Integrated with components |

## Migration Plan

### Phase 1: Setup & Initialization

#### Step 1: Install Dependencies
```bash
cd frontend
npm install bits-ui clsx tailwind-merge @lucide/svelte
```

#### Step 2: Initialize Shadcn-Svelte
```bash
npx shadcn-svelte@latest init
```

**Configuration Options:**
- Base color: `Slate`
- Global CSS file: `src/app.css` (will be overwritten)
- Import aliases:
  - lib: `$lib`
  - components: `$lib/components`
  - utils: `$lib/utils`
  - hooks: `$lib/hooks`
  - ui: `$lib/components/ui`

#### Step 3: Install Core Components
```bash
npx shadcn-svelte@latest add button input select card badge progress alert separator sheet scroll-area
```

### Phase 2: Batch Migration

#### Batch 1: Layout & Navigation
**Target**: `src/routes/+layout.svelte`
**Priority**: High (Foundation & Immediate Impact)

**Components to Replace:**
- Navigation buttons → `Button` component
- Mobile menu → `Sheet` component
- Visual separators → `Separator` component

**Current Elements:**
```svelte
<!-- Replace these custom elements -->
<button class="px-3 py-2 text-sm font-medium text-gray-500 hover:text-gray-700">
<button class="inline-flex items-center justify-center rounded-md p-2 text-gray-400">
<button class="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white">
```

**New Elements:**
```svelte
<!-- With shadcn-svelte components -->
<Button variant="ghost" size="sm">Dashboard</Button>
<Button variant="ghost" size="icon"><!-- mobile menu --></Button>
<Button>Login</Button>
<Sheet><!-- mobile navigation --></Sheet>
```

**Testing Requirements:**
- [ ] Desktop navigation functionality
- [ ] Mobile menu open/close
- [ ] Authentication state handling
- [ ] Route highlighting
- [ ] Responsive behavior

#### Batch 2: Core Search Components
**Targets**: `SearchBox.svelte` + `SearchFilters.svelte`
**Priority**: High (Core Functionality)

**SearchBox Migration:**
- Search input → `Input` component
- Mode selector buttons → `Button` with variants
- Clear button → `Button` with icon
- Loading spinner → Built-in loading state

**SearchFilters Migration:**
- Filter dropdowns → `Select` component
- Active filter tags → `Badge` component
- Clear filters button → `Button` component

**Current Logic to Preserve:**
- Debounced search functionality
- Mode switching (text/vector/hybrid)
- Event dispatching
- Keyboard shortcuts (Enter, Escape)
- Filter state management

**Testing Requirements:**
- [ ] Search input debouncing
- [ ] Mode selector functionality
- [ ] Filter selection and clearing
- [ ] Keyboard shortcuts
- [ ] Loading states
- [ ] Event dispatching

#### Batch 3: Data Display Components
**Targets**: `CodeSnippet.svelte` + `SearchResults.svelte`
**Priority**: Medium (Content Presentation)

**CodeSnippet Migration:**
- Container → `Card` component
- Copy button → `Button` with icon
- Language badge → `Badge` component
- Code content → `ScrollArea` if needed

**SearchResults Migration:**
- Result containers → `Card` component
- Action buttons → `Button` component
- Scrollable area → `ScrollArea` component

**Current Logic to Preserve:**
- Syntax highlighting
- Copy to clipboard functionality
- Search term highlighting
- Line number display
- Truncation logic
- Result pagination/loading

**Testing Requirements:**
- [ ] Code syntax highlighting
- [ ] Copy functionality
- [ ] Search term highlighting
- [ ] Responsive layout
- [ ] Loading and empty states

#### Batch 4: Status & Connection Components
**Targets**: `EmbeddingStatus.svelte` + `GitHubConnection.svelte`
**Priority**: Medium (Complex State Management)

**EmbeddingStatus Migration:**
- Main container → `Card` component
- Progress indicator → `Progress` component
- Action buttons → `Button` component
- Status badges → `Badge` component
- Section dividers → `Separator` component

**GitHubConnection Migration:**
- Container → `Card` component
- Connect/disconnect buttons → `Button` component
- Success/error messages → `Alert` component

**Current Logic to Preserve:**
- Real-time status polling
- Progress calculations
- OAuth flow handling
- Error state management
- Auto-refresh functionality

**Testing Requirements:**
- [ ] Status polling and updates
- [ ] Progress bar accuracy
- [ ] Button state management
- [ ] OAuth flow functionality
- [ ] Error handling and display
- [ ] Auto-refresh behavior

## Implementation Guidelines

### Code Structure
```
frontend/src/
├── lib/
│   ├── components/
│   │   ├── ui/                    # Shadcn-svelte components
│   │   │   ├── button/
│   │   │   ├── input/
│   │   │   ├── select/
│   │   │   └── ...
│   │   ├── SearchBox.svelte       # Migrated components
│   │   ├── SearchFilters.svelte
│   │   └── ...
│   └── utils/
│       └── cn.ts                  # Class name utility
├── routes/
│   └── +layout.svelte            # Migrated layout
└── app.css                       # Shadcn-svelte styles
```

### Migration Checklist Per Component

#### Before Migration
- [ ] Document current component API (props, events, slots)
- [ ] Identify all UI elements to replace
- [ ] Note any custom styling or behavior to preserve
- [ ] Create branch: `feat/migrate-[component-name]`

#### During Migration
- [ ] Install required shadcn-svelte components
- [ ] Replace UI elements one by one
- [ ] Preserve all business logic and state management
- [ ] Update import statements
- [ ] Test component in isolation
- [ ] Verify props and events still work

#### After Migration
- [ ] Update component documentation
- [ ] Test integration with parent components
- [ ] Verify responsive behavior
- [ ] Check accessibility compliance
- [ ] Run full test suite
- [ ] Create pull request for review

### Design System Changes

#### Colors
- Primary: From custom blue (`#3b82f6`) to shadcn-svelte primary
- Secondary: From custom gray to shadcn-svelte secondary
- Success: From custom green to shadcn-svelte success
- Error: From custom red to shadcn-svelte destructive

#### Typography
- Font family: Inherits from shadcn-svelte system fonts
- Font sizes: Uses shadcn-svelte typography scale
- Font weights: Standardized shadcn-svelte weights

#### Spacing
- Padding/margins: shadcn-svelte spacing tokens
- Component gaps: Consistent spacing system
- Border radius: shadcn-svelte radius values

#### Animations
- Transitions: shadcn-svelte standard transitions
- Loading states: Built-in component animations
- Hover effects: Consistent interaction patterns

### Testing Strategy

#### Unit Testing
- Test component props and events
- Verify business logic preservation
- Test accessibility compliance
- Validate responsive behavior

#### Integration Testing
- Test component interactions
- Verify event propagation
- Test with real data
- Cross-browser compatibility

#### Visual Testing
- Compare before/after screenshots
- Verify design system consistency
- Test dark mode support (if needed)
- Mobile responsiveness

## Risk Mitigation

### Potential Issues
1. **Breaking changes in component APIs**
   - Solution: Comprehensive testing at each step
   - Rollback plan: Git branch per batch

2. **Performance regressions**
   - Solution: Bundle size monitoring
   - Performance testing before/after each batch

3. **Accessibility issues**
   - Solution: Accessibility audit per component
   - Use shadcn-svelte's built-in accessibility features

4. **Styling conflicts**
   - Solution: CSS specificity management
   - Use shadcn-svelte's design tokens consistently

### Rollback Strategy
- Each batch is a separate branch
- Component-level rollback possible
- Database/API unchanged (no breaking changes)
- Feature flags not needed (in-place replacement)

## Success Criteria

### Functional Requirements
- [ ] All existing functionality preserved
- [ ] No regression in user experience
- [ ] Performance maintained or improved
- [ ] Accessibility compliance maintained

### Design Requirements
- [ ] Consistent visual design across all components
- [ ] Proper dark mode support (if applicable)
- [ ] Responsive design on all screen sizes
- [ ] Modern, professional appearance

### Development Requirements
- [ ] Reduced custom CSS codebase
- [ ] Improved component maintainability
- [ ] Better design system consistency
- [ ] Easier future component additions

## Timeline Estimate

| Phase | Duration | Deliverables |
|-------|----------|--------------|
| Setup & Initialization | 1-2 days | Dependencies installed, shadcn-svelte initialized |
| Batch 1: Layout | 2-3 days | Migrated navigation and layout |
| Batch 2: Search Components | 3-4 days | Migrated SearchBox and SearchFilters |
| Batch 3: Display Components | 3-4 days | Migrated CodeSnippet and SearchResults |
| Batch 4: Status Components | 4-5 days | Migrated EmbeddingStatus and GitHubConnection |
| Testing & Cleanup | 2-3 days | Full integration testing, documentation |

**Total Estimate: 15-21 days**

## Next Steps

1. **Review and approve this specification**
2. **Set up development environment**
3. **Execute Phase 1: Setup & Initialization**
4. **Begin Batch 1: Layout & Navigation migration**
5. **Iterate through remaining batches**
6. **Conduct final testing and cleanup**

---

*This specification provides a comprehensive roadmap for migrating from custom components to shadcn-svelte while maintaining functionality and improving design consistency.*