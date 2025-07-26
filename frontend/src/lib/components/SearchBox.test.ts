// ABOUTME: Tests for SearchBox component functionality
// ABOUTME: Covers search input, debouncing, event handling, and keyboard shortcuts

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { page } from '@vitest/browser/context';
import { render, waitFor } from 'vitest-browser-svelte';
import SearchBox from './SearchBox.svelte';

describe('SearchBox', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('should render search input with placeholder', () => {
		render(SearchBox);

		const searchInput = page.getByPlaceholderText('Search your code...');
		expect.element(searchInput).toBeInTheDocument();
	});

	it('should emit search event when input changes', async () => {
		const { component } = render(SearchBox);

		let searchEventFired = false;
		let searchQuery = '';

		component.$on('search', (event) => {
			searchEventFired = true;
			searchQuery = event.detail.query;
		});

		const searchInput = page.getByPlaceholderText('Search your code...');
		await searchInput.fill('test query');

		// Wait for debounce
		await waitFor(
			() => {
				expect(searchEventFired).toBe(true);
				expect(searchQuery).toBe('test query');
			},
			{ timeout: 1000 }
		);
	});

	it('should debounce search input', async () => {
		const { component } = render(SearchBox);

		let searchEventCount = 0;

		component.$on('search', () => {
			searchEventCount++;
		});

		const searchInput = page.getByPlaceholderText('Search your code...');

		// Type multiple characters quickly
		await searchInput.fill('t');
		await searchInput.fill('te');
		await searchInput.fill('test');

		// Should only fire once after debounce delay
		await waitFor(
			() => {
				expect(searchEventCount).toBe(1);
			},
			{ timeout: 1000 }
		);
	});

	it('should handle empty search query', async () => {
		const { component } = render(SearchBox);

		let searchEventFired = false;
		let searchQuery = '';

		component.$on('search', (event) => {
			searchEventFired = true;
			searchQuery = event.detail.query;
		});

		const searchInput = page.getByPlaceholderText('Search your code...');
		await searchInput.fill('   ');

		await waitFor(
			() => {
				expect(searchEventFired).toBe(true);
				expect(searchQuery).toBe('');
			},
			{ timeout: 1000 }
		);
	});

	it('should handle keyboard shortcuts', async () => {
		const { component } = render(SearchBox);

		let clearEventFired = false;

		component.$on('clear', () => {
			clearEventFired = true;
		});

		const searchInput = page.getByPlaceholderText('Search your code...');
		await searchInput.fill('test query');

		// Simulate Escape key
		await searchInput.press('Escape');

		expect(clearEventFired).toBe(true);
		expect(await searchInput.inputValue()).toBe('');
	});

	it('should submit search on Enter key', async () => {
		const { component } = render(SearchBox);

		let submitEventFired = false;
		let submitQuery = '';

		component.$on('submit', (event) => {
			submitEventFired = true;
			submitQuery = event.detail.query;
		});

		const searchInput = page.getByPlaceholderText('Search your code...');
		await searchInput.fill('test query');
		await searchInput.press('Enter');

		expect(submitEventFired).toBe(true);
		expect(submitQuery).toBe('test query');
	});

	it('should display search results count when provided', () => {
		render(SearchBox, { props: { resultsCount: 42 } });

		const resultsText = page.getByText('42 results');
		expect.element(resultsText).toBeInTheDocument();
	});

	it('should show loading state', () => {
		render(SearchBox, { props: { loading: true } });

		const loadingIndicator = page.getByTestId('search-loading');
		expect.element(loadingIndicator).toBeInTheDocument();
	});

	it('should focus input when focused prop is true', async () => {
		render(SearchBox, { props: { focused: true } });

		const searchInput = page.getByPlaceholderText('Search your code...');
		await waitFor(() => {
			expect.element(searchInput).toBeFocused();
		});
	});

	it('should allow initial value to be set', () => {
		render(SearchBox, { props: { value: 'initial search' } });

		const searchInput = page.getByPlaceholderText('Search your code...');
		expect.element(searchInput).toHaveValue('initial search');
	});

	it('should emit clear event when clear button is clicked', async () => {
		const { component } = render(SearchBox, { props: { value: 'test query' } });

		let clearEventFired = false;

		component.$on('clear', () => {
			clearEventFired = true;
		});

		const clearButton = page.getByTestId('clear-search');
		await clearButton.click();

		expect(clearEventFired).toBe(true);
	});

	it('should not show clear button when input is empty', () => {
		render(SearchBox, { props: { value: '' } });

		const clearButton = page.getByTestId('clear-search');
		expect.element(clearButton).not.toBeInTheDocument();
	});

	it('should handle disabled state', () => {
		render(SearchBox, { props: { disabled: true } });

		const searchInput = page.getByPlaceholderText('Search your code...');
		expect.element(searchInput).toBeDisabled();
	});
});
