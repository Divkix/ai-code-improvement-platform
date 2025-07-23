<!-- ABOUTME: Search input component with debounced input handling and clear functionality -->
<!-- ABOUTME: Provides keyboard shortcuts and loading states for enhanced user experience -->

<script lang="ts">
    import { createEventDispatcher } from 'svelte';
    import { debounce } from '../utils/debounce';
    
    export let value = '';
    export let placeholder = 'Search code...';
    export let disabled = false;
    export let loading = false;
    // Remove autofocus prop to improve accessibility
    
    const dispatch = createEventDispatcher<{
        search: string;
        clear: void;
        focus: void;
        blur: void;
    }>();
    
    // Debounced search function to avoid too many API calls
    const debouncedSearch = debounce((query: string) => {
        if (query.trim()) {
            dispatch('search', query.trim());
        } else {
            dispatch('clear');
        }
    }, 300);
    
    function handleInput(event: Event) {
        const target = event.target as HTMLInputElement;
        value = target.value;
        
        // Always trigger the debounced search
        debouncedSearch(value);
    }
    
    function handleKeydown(event: KeyboardEvent) {
        if (event.key === 'Enter' && value.trim()) {
            // Cancel debounce and search immediately on Enter
            event.preventDefault();
            dispatch('search', value.trim());
        }
        
        if (event.key === 'Escape') {
            value = '';
            dispatch('clear');
            (event.target as HTMLInputElement).blur();
        }
    }
    
    function handleClear() {
        value = '';
        dispatch('clear');
    }
    
    function handleFocus() {
        dispatch('focus');
    }
    
    function handleBlur() {
        dispatch('blur');
    }
</script>

<div class="search-box">
    <div class="search-container">
        <input
            type="text"
            bind:value
            on:input={handleInput}
            on:keydown={handleKeydown}
            on:focus={handleFocus}
            on:blur={handleBlur}
            {placeholder}
            {disabled}
            class="search-input"
            class:loading
            autocomplete="off"
            spellcheck="false"
        />
        
        <!-- Search icon or loading spinner -->
        <div class="search-icon">
            {#if loading}
                <div class="spinner" aria-label="Searching..."></div>
            {:else}
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <circle cx="11" cy="11" r="8" />
                    <path d="M21 21l-4.35-4.35" />
                </svg>
            {/if}
        </div>
        
        <!-- Clear button -->
        {#if value && !disabled}
            <button 
                type="button" 
                class="clear-button" 
                on:click={handleClear}
                aria-label="Clear search"
                tabindex="-1"
            >
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <line x1="18" y1="6" x2="6" y2="18" />
                    <line x1="6" y1="6" x2="18" y2="18" />
                </svg>
            </button>
        {/if}
    </div>
</div>

<style>
    .search-box {
        width: 100%;
        max-width: 600px;
    }
    
    .search-container {
        position: relative;
        display: flex;
        align-items: center;
    }
    
    .search-input {
        width: 100%;
        padding: 12px 48px 12px 48px;
        border: 2px solid #e5e7eb;
        border-radius: 8px;
        font-size: 16px;
        font-family: inherit;
        background: white;
        transition: border-color 0.2s, box-shadow 0.2s;
        outline: none;
    }
    
    .search-input:focus {
        border-color: #3b82f6;
        box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
    }
    
    .search-input:disabled {
        background-color: #f9fafb;
        color: #6b7280;
        cursor: not-allowed;
    }
    
    .search-input.loading {
        padding-right: 56px;
    }
    
    .search-icon {
        position: absolute;
        left: 14px;
        top: 50%;
        transform: translateY(-50%);
        color: #6b7280;
        pointer-events: none;
        display: flex;
        align-items: center;
        justify-content: center;
    }
    
    .clear-button {
        position: absolute;
        right: 12px;
        top: 50%;
        transform: translateY(-50%);
        background: none;
        border: none;
        color: #6b7280;
        cursor: pointer;
        padding: 6px;
        border-radius: 4px;
        transition: color 0.2s, background-color 0.2s;
        display: flex;
        align-items: center;
        justify-content: center;
    }
    
    .clear-button:hover {
        color: #374151;
        background-color: #f3f4f6;
    }
    
    .clear-button:focus {
        outline: 2px solid #3b82f6;
        outline-offset: 2px;
    }
    
    .spinner {
        width: 20px;
        height: 20px;
        border: 2px solid #e5e7eb;
        border-top: 2px solid #3b82f6;
        border-radius: 50%;
        animation: spin 1s linear infinite;
    }
    
    @keyframes spin {
        0% { transform: rotate(0deg); }
        100% { transform: rotate(360deg); }
    }
    
    /* Responsive adjustments */
    @media (max-width: 640px) {
        .search-input {
            font-size: 16px; /* Prevent zoom on iOS */
            padding: 10px 40px 10px 40px;
        }
        
        .search-input.loading {
            padding-right: 48px;
        }
        
        .search-icon {
            left: 12px;
        }
        
        .clear-button {
            right: 10px;
        }
    }
    
    /* High contrast mode support */
    @media (prefers-contrast: high) {
        .search-input {
            border-color: #000;
        }
        
        .search-input:focus {
            border-color: #0066cc;
            box-shadow: 0 0 0 3px rgba(0, 102, 204, 0.3);
        }
    }
    
    /* Reduced motion support */
    @media (prefers-reduced-motion: reduce) {
        .search-input,
        .clear-button,
        .spinner {
            transition: none;
            animation: none;
        }
    }
</style>