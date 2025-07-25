// ABOUTME: Utility functions for parsing and sanitizing markdown content
// ABOUTME: Provides safe HTML rendering for chat messages with syntax highlighting

import { marked } from 'marked';
import DOMPurify from 'dompurify';

// Configure marked with code highlighting and security options
marked.setOptions({
	breaks: true,
	gfm: true
});

// Custom renderer for code blocks to add syntax highlighting classes
const renderer = new marked.Renderer();
renderer.code = function ({ text, lang }: { text: string; lang?: string }) {
	if (lang) {
		return `<pre><code class="language-${lang} hljs">${text}</code></pre>`;
	}
	return `<pre><code class="hljs">${text}</code></pre>`;
};

renderer.codespan = function ({ text }: { text: string }) {
	return `<code class="inline-code">${text}</code>`;
};

marked.use({ renderer });

/**
 * Parse markdown content and return sanitized HTML
 */
export async function parseMarkdown(content: string): Promise<string> {
	if (!content) return '';

	// Parse markdown to HTML
	const html = await marked.parse(content);

	// Sanitize HTML to prevent XSS attacks
	return DOMPurify.sanitize(html, {
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
}

/**
 * Check if content contains markdown formatting
 */
export function hasMarkdownFormatting(content: string): boolean {
	if (!content) return false;

	// Look for common markdown patterns
	const markdownPatterns = [
		/\*\*.*?\*\*/, // Bold
		/\*.*?\*/, // Italic
		/`.*?`/, // Inline code
		/```[\s\S]*?```/, // Code blocks
		/#{1,6}\s/, // Headers
		/^\s*[-*+]\s/m, // Lists
		/^\s*\d+\.\s/m, // Numbered lists
		/\[.*?\]\(.*?\)/ // Links
	];

	return markdownPatterns.some((pattern) => pattern.test(content));
}
