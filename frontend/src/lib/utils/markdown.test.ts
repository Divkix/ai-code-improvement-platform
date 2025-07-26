// ABOUTME: Tests for markdown parsing and sanitization utilities
// ABOUTME: Verifies safe HTML rendering, syntax highlighting, and markdown detection

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { parseMarkdown, hasMarkdownFormatting } from './markdown';

// Mock dependencies
vi.mock('marked', () => {
	const mockRenderer = {
		code: vi.fn(),
		codespan: vi.fn()
	};

	return {
		marked: {
			parse: vi.fn(),
			setOptions: vi.fn(),
			use: vi.fn(),
			Renderer: vi.fn(() => mockRenderer)
		}
	};
});

vi.mock('dompurify', () => ({
	default: {
		sanitize: vi.fn((html) => html) // Return input unchanged for testing
	}
}));

vi.mock('highlight.js', () => ({
	default: {
		highlight: vi.fn(),
		highlightAuto: vi.fn()
	}
}));

describe('Markdown utilities', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	afterEach(() => {
		vi.restoreAllMocks();
	});

	describe('parseMarkdown', () => {
		it('should return empty string for empty content', async () => {
			const result = await parseMarkdown('');
			expect(result).toBe('');
		});

		it('should parse markdown and sanitize HTML', async () => {
			const mockHtml = '<p>Hello <strong>world</strong></p>';
			const sanitizedHtml = '<p>Hello <strong>world</strong></p>';

			const { marked } = await import('marked');
			const mockDOMPurify = await import('dompurify');

			(marked.parse as any).mockResolvedValueOnce(mockHtml);
			(mockDOMPurify.default.sanitize as any).mockReturnValueOnce(sanitizedHtml);

			const result = await parseMarkdown('Hello **world**');

			expect(marked.parse).toHaveBeenCalledWith('Hello **world**');
			expect(mockDOMPurify.default.sanitize).toHaveBeenCalledWith(mockHtml, {
				ALLOWED_TAGS: [
					'p',
					'br',
					'strong',
					'em',
					'code',
					'pre',
					'h1',
					'h2',
					'h3',
					'h4',
					'h5',
					'h6',
					'ul',
					'ol',
					'li',
					'blockquote',
					'a',
					'img',
					'table',
					'thead',
					'tbody',
					'tr',
					'td',
					'th'
				],
				ALLOWED_ATTR: ['href', 'src', 'alt', 'class', 'target', 'rel']
			});
			expect(result).toBe(sanitizedHtml);
		});

		it('should handle markdown with code blocks', async () => {
			const content = '```javascript\nconsole.log("hello");\n```';
			const mockHtml =
				'<pre><code class="language-javascript hljs">console.log("hello");</code></pre>';

			const { marked } = await import('marked');
			const mockDOMPurify = await import('dompurify');

			(marked.parse as any).mockResolvedValueOnce(mockHtml);
			(mockDOMPurify.default.sanitize as any).mockReturnValueOnce(mockHtml);

			const result = await parseMarkdown(content);

			expect(marked.parse).toHaveBeenCalledWith(content);
			expect(result).toBe(mockHtml);
		});

		it('should handle markdown with inline code', async () => {
			const content = 'Use `console.log()` to debug';
			const mockHtml = '<p>Use <code class="inline-code">console.log()</code> to debug</p>';

			const { marked } = await import('marked');
			const mockDOMPurify = await import('dompurify');

			(marked.parse as any).mockResolvedValueOnce(mockHtml);
			(mockDOMPurify.default.sanitize as any).mockReturnValueOnce(mockHtml);

			const result = await parseMarkdown(content);

			expect(result).toBe(mockHtml);
		});

		it('should handle complex markdown content', async () => {
			const content = `# Header
			
**Bold** and *italic* text.

- List item 1
- List item 2

[Link](https://example.com)`;

			const mockHtml = `<h1>Header</h1>
<p><strong>Bold</strong> and <em>italic</em> text.</p>
<ul>
<li>List item 1</li>
<li>List item 2</li>
</ul>
<p><a href="https://example.com">Link</a></p>`;

			const { marked } = await import('marked');
			const mockDOMPurify = await import('dompurify');

			(marked.parse as any).mockResolvedValueOnce(mockHtml);
			(mockDOMPurify.default.sanitize as any).mockReturnValueOnce(mockHtml);

			const result = await parseMarkdown(content);

			expect(result).toBe(mockHtml);
		});

		it('should handle null content gracefully', async () => {
			const result = await parseMarkdown(null as any);
			expect(result).toBe('');
		});

		it('should handle undefined content gracefully', async () => {
			const result = await parseMarkdown(undefined as any);
			expect(result).toBe('');
		});

		it('should call DOMPurify with correct sanitization options', async () => {
			const content = 'Test content';
			const mockHtml = '<p>Test content</p>';

			const { marked } = await import('marked');
			const mockDOMPurify = await import('dompurify');

			(marked.parse as any).mockResolvedValueOnce(mockHtml);

			await parseMarkdown(content);

			expect(mockDOMPurify.default.sanitize).toHaveBeenCalledWith(mockHtml, {
				ALLOWED_TAGS: expect.arrayContaining(['p', 'strong', 'code', 'pre', 'h1', 'a']),
				ALLOWED_ATTR: ['href', 'src', 'alt', 'class', 'target', 'rel']
			});
		});
	});

	describe('hasMarkdownFormatting', () => {
		it('should return false for empty content', () => {
			expect(hasMarkdownFormatting('')).toBe(false);
			expect(hasMarkdownFormatting(null as any)).toBe(false);
			expect(hasMarkdownFormatting(undefined as any)).toBe(false);
		});

		it('should detect bold formatting', () => {
			expect(hasMarkdownFormatting('This is **bold** text')).toBe(true);
			expect(hasMarkdownFormatting('**Bold at start**')).toBe(true);
			expect(hasMarkdownFormatting('End is **bold**')).toBe(true);
		});

		it('should detect italic formatting', () => {
			expect(hasMarkdownFormatting('This is *italic* text')).toBe(true);
			expect(hasMarkdownFormatting('*Italic at start*')).toBe(true);
			expect(hasMarkdownFormatting('End is *italic*')).toBe(true);
		});

		it('should detect inline code formatting', () => {
			expect(hasMarkdownFormatting('Use `console.log()` to debug')).toBe(true);
			expect(hasMarkdownFormatting('`code at start`')).toBe(true);
			expect(hasMarkdownFormatting('End with `code`')).toBe(true);
		});

		it('should detect code block formatting', () => {
			expect(hasMarkdownFormatting('```\ncode block\n```')).toBe(true);
			expect(hasMarkdownFormatting('```javascript\nconsole.log("hello");\n```')).toBe(true);
		});

		it('should detect header formatting', () => {
			expect(hasMarkdownFormatting('# Header 1')).toBe(true);
			expect(hasMarkdownFormatting('## Header 2')).toBe(true);
			expect(hasMarkdownFormatting('### Header 3')).toBe(true);
			expect(hasMarkdownFormatting('#### Header 4')).toBe(true);
			expect(hasMarkdownFormatting('##### Header 5')).toBe(true);
			expect(hasMarkdownFormatting('###### Header 6')).toBe(true);
		});

		it('should detect unordered list formatting', () => {
			expect(hasMarkdownFormatting('- List item')).toBe(true);
			expect(hasMarkdownFormatting('* List item')).toBe(true);
			expect(hasMarkdownFormatting('+ List item')).toBe(true);
			expect(hasMarkdownFormatting('  - Indented list item')).toBe(true);
		});

		it('should detect ordered list formatting', () => {
			expect(hasMarkdownFormatting('1. First item')).toBe(true);
			expect(hasMarkdownFormatting('2. Second item')).toBe(true);
			expect(hasMarkdownFormatting('  1. Indented item')).toBe(true);
		});

		it('should detect link formatting', () => {
			expect(hasMarkdownFormatting('[Link text](https://example.com)')).toBe(true);
			expect(hasMarkdownFormatting('[](https://example.com)')).toBe(true);
			expect(hasMarkdownFormatting('[Link text]()')).toBe(true);
		});

		it('should return false for plain text', () => {
			expect(hasMarkdownFormatting('This is plain text')).toBe(false);
			expect(hasMarkdownFormatting('No formatting here')).toBe(false);
			expect(hasMarkdownFormatting('Just regular words')).toBe(false);
		});

		it('should handle false positives correctly', () => {
			// These should not be detected as markdown
			expect(hasMarkdownFormatting('This * is not italic')).toBe(false);
			expect(hasMarkdownFormatting('This ** is not bold')).toBe(false);
			expect(hasMarkdownFormatting('This ` is not code')).toBe(false);
			expect(hasMarkdownFormatting('#hashtag not header')).toBe(false);
		});

		it('should handle multiline content', () => {
			const multilineContent = `First line
- Second line with list
Third line`;
			expect(hasMarkdownFormatting(multilineContent)).toBe(true);
		});

		it('should handle mixed formatting', () => {
			const mixedContent = 'This has **bold** and `code` and [link](url)';
			expect(hasMarkdownFormatting(mixedContent)).toBe(true);
		});

		it('should handle edge cases', () => {
			expect(hasMarkdownFormatting('***')).toBe(true); // Should match * pattern
			expect(hasMarkdownFormatting('```')).toBe(false); // Incomplete code block
			expect(hasMarkdownFormatting('[]()')).toBe(true); // Empty link
		});
	});
});
