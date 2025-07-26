// ABOUTME: Tests for CodeSnippet component functionality
// ABOUTME: Covers code syntax highlighting, copy functionality, and line number display

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/svelte';
import CodeSnippet from './CodeSnippet.svelte';

// Mock clipboard API
const mockClipboard = {
	writeText: vi.fn().mockResolvedValue(undefined)
};

Object.defineProperty(global.navigator, 'clipboard', {
	value: mockClipboard,
	writable: true
});

describe('CodeSnippet', () => {
	const mockCode = `function greet(name: string): string {
  return \`Hello, \${name}!\`;
}

export default greet;`;

	beforeEach(() => {
		vi.clearAllMocks();
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	it('should render code content', () => {
		render(CodeSnippet, {
			props: {
				code: mockCode,
				language: 'typescript'
			}
		});

		expect(screen.getByText(/function greet/)).toBeInTheDocument();
		expect(screen.getByText(/export default greet/)).toBeInTheDocument();
	});

	it('should display language badge', () => {
		render(CodeSnippet, {
			props: {
				code: mockCode,
				language: 'typescript'
			}
		});

		expect(screen.getByText('typescript')).toBeInTheDocument();
	});

	it('should show line numbers when enabled', () => {
		render(CodeSnippet, {
			props: {
				code: mockCode,
				language: 'typescript',
				showLineNumbers: true,
				startLine: 10
			}
		});

		// Check that line numbers are displayed
		expect(screen.getByText('10')).toBeInTheDocument();
		expect(screen.getByText('11')).toBeInTheDocument();
		expect(screen.getByText('12')).toBeInTheDocument();
	});

	it('should handle copy to clipboard', async () => {
		const { component } = render(CodeSnippet, {
			props: {
				code: mockCode,
				language: 'typescript',
				showCopyButton: true
			}
		});

		let copyEventFired = false;

		component.$on('copy', () => {
			copyEventFired = true;
		});

		const copyButton = screen.getByTestId('copy-button');
		await fireEvent.click(copyButton);

		expect(mockClipboard.writeText).toHaveBeenCalledWith(mockCode);
		expect(copyEventFired).toBe(true);
	});

	it('should show copy success feedback', async () => {
		render(CodeSnippet, {
			props: {
				code: mockCode,
				language: 'typescript',
				showCopyButton: true
			}
		});

		const copyButton = screen.getByTestId('copy-button');
		await fireEvent.click(copyButton);

		// Should show success state temporarily
		expect(screen.getByText('Copied!')).toBeInTheDocument();
	});

	it('should handle copy failure gracefully', async () => {
		// Mock clipboard failure
		mockClipboard.writeText.mockRejectedValueOnce(new Error('Clipboard failed'));

		render(CodeSnippet, {
			props: {
				code: mockCode,
				language: 'typescript',
				showCopyButton: true
			}
		});

		const copyButton = screen.getByTestId('copy-button');
		await fireEvent.click(copyButton);

		// Should not throw error and still emit event
		expect(mockClipboard.writeText).toHaveBeenCalled();
	});

	it('should highlight specific lines', () => {
		render(CodeSnippet, {
			props: {
				code: mockCode,
				language: 'typescript',
				highlightLines: [1, 3]
			}
		});

		const highlightedLines = screen.getAllByTestId(/highlighted-line-/);
		expect(highlightedLines).toHaveLength(2);
	});

	it('should wrap long lines when enabled', () => {
		const longCode =
			'const veryLongVariableName = "This is a very long string that should wrap when word wrap is enabled and the container is not wide enough to fit it all on one line";';

		render(CodeSnippet, {
			props: {
				code: longCode,
				language: 'javascript',
				wrapLines: true
			}
		});

		const codeContainer = screen.getByTestId('code-container');
		expect(codeContainer).toHaveClass('whitespace-pre-wrap');
	});

	it('should not wrap lines by default', () => {
		render(CodeSnippet, {
			props: {
				code: mockCode,
				language: 'typescript'
			}
		});

		const codeContainer = screen.getByTestId('code-container');
		expect(codeContainer).toHaveClass('whitespace-pre');
	});

	it('should handle empty code', () => {
		render(CodeSnippet, {
			props: {
				code: '',
				language: 'typescript'
			}
		});

		const emptyMessage = screen.getByText(/No code to display/i);
		expect(emptyMessage).toBeInTheDocument();
	});

	it('should display filename when provided', () => {
		render(CodeSnippet, {
			props: {
				code: mockCode,
				language: 'typescript',
				filename: 'greet.ts'
			}
		});

		expect(screen.getByText('greet.ts')).toBeInTheDocument();
	});

	it('should show download option when enabled', async () => {
		const { component } = render(CodeSnippet, {
			props: {
				code: mockCode,
				language: 'typescript',
				filename: 'greet.ts',
				showDownload: true
			}
		});

		let downloadEventFired = false;

		component.$on('download', () => {
			downloadEventFired = true;
		});

		const downloadButton = screen.getByTestId('download-button');
		await fireEvent.click(downloadButton);

		expect(downloadEventFired).toBe(true);
	});

	it('should apply custom theme', () => {
		render(CodeSnippet, {
			props: {
				code: mockCode,
				language: 'typescript',
				theme: 'dark'
			}
		});

		const codeContainer = screen.getByTestId('code-container');
		expect(codeContainer).toHaveClass('theme-dark');
	});

	it('should handle syntax highlighting for different languages', () => {
		const pythonCode = 'def hello_world():\n    print("Hello, World!")';

		render(CodeSnippet, {
			props: {
				code: pythonCode,
				language: 'python'
			}
		});

		expect(screen.getByText('python')).toBeInTheDocument();
		expect(screen.getByText(/def hello_world/)).toBeInTheDocument();
	});

	it('should emit line click events when interactive', async () => {
		const { component } = render(CodeSnippet, {
			props: {
				code: mockCode,
				language: 'typescript',
				interactive: true,
				showLineNumbers: true,
				startLine: 1
			}
		});

		let lineClickEventFired = false;
		let clickedLineNumber = 0;

		component.$on('line-click', (event) => {
			lineClickEventFired = true;
			clickedLineNumber = event.detail.lineNumber;
		});

		const lineElement = screen.getByTestId('line-1');
		await fireEvent.click(lineElement);

		expect(lineClickEventFired).toBe(true);
		expect(clickedLineNumber).toBe(1);
	});

	it('should handle loading state', () => {
		render(CodeSnippet, {
			props: {
				code: '',
				language: 'typescript',
				loading: true
			}
		});

		const loadingIndicator = screen.getByTestId('code-loading');
		expect(loadingIndicator).toBeInTheDocument();
	});

	it('should display error state', () => {
		render(CodeSnippet, {
			props: {
				code: '',
				language: 'typescript',
				error: 'Failed to load code'
			}
		});

		expect(screen.getByText('Failed to load code')).toBeInTheDocument();
	});

	it('should handle read-only mode', () => {
		render(CodeSnippet, {
			props: {
				code: mockCode,
				language: 'typescript',
				readOnly: true
			}
		});

		// Copy button should not be present in read-only mode
		const copyButton = screen.queryByTestId('copy-button');
		expect(copyButton).not.toBeInTheDocument();
	});

	it('should support custom max height', () => {
		render(CodeSnippet, {
			props: {
				code: mockCode,
				language: 'typescript',
				maxHeight: '200px'
			}
		});

		const codeContainer = screen.getByTestId('code-container');
		expect(codeContainer.style.maxHeight).toBe('200px');
	});
});
