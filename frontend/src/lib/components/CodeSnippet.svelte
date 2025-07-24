<!-- ABOUTME: Code snippet display component with syntax highlighting and search term highlighting -->
<!-- ABOUTME: Supports line number display and copy-to-clipboard functionality -->

<script lang="ts">
	import { createEventDispatcher } from 'svelte';

	export let content: string;
	export let language: string = '';
	export let searchTerm: string = '';
	export let maxLines: number = 0;
	export let showLineNumbers: boolean = false;
	export let startLine: number = 1;
	export let fileName: string = '';

	const dispatch = createEventDispatcher<{
		copy: string;
	}>();

	let copySuccess = false;

	// Language color mapping
	const languageColors: Record<string, string> = {
		javascript: '#f7df1e',
		typescript: '#3178c6',
		python: '#3776ab',
		go: '#00add8',
		java: '#ed8b00',
		php: '#777bb4',
		cpp: '#00599c',
		csharp: '#239120',
		rust: '#000000',
		ruby: '#cc342d',
		html: '#e34f26',
		css: '#1572b6',
		json: '#000000',
		yaml: '#cb171e',
		shell: '#4eaa25',
		sql: '#336791'
	};

	function getLanguageColor(lang: string): string {
		return languageColors[lang.toLowerCase()] || '#6b7280';
	}

	function highlightSearchTerm(text: string, term: string): string {
		if (!term) return escapeHtml(text);

		const escapedText = escapeHtml(text);
		const escapedTerm = escapeHtml(term);
		const regex = new RegExp(`(${escapedTerm})`, 'gi');

		return escapedText.replace(regex, '<mark>$1</mark>');
	}

	function escapeHtml(text: string): string {
		const div = document.createElement('div');
		div.textContent = text;
		return div.innerHTML;
	}

	function processContent(content: string): string[] {
		const lines = content.split('\n');

		if (maxLines > 0 && lines.length > maxLines) {
			return lines.slice(0, maxLines);
		}

		return lines;
	}

	async function copyToClipboard() {
		try {
			await navigator.clipboard.writeText(content);
			copySuccess = true;
			dispatch('copy', content);

			setTimeout(() => {
				copySuccess = false;
			}, 2000);
		} catch (err) {
			console.warn('Failed to copy to clipboard:', err);
			// Fallback for older browsers
			const textArea = document.createElement('textarea');
			textArea.value = content;
			document.body.appendChild(textArea);
			textArea.select();
			try {
				document.execCommand('copy');
				copySuccess = true;
				dispatch('copy', content);
				setTimeout(() => {
					copySuccess = false;
				}, 2000);
			} catch (fallbackErr) {
				console.error('Fallback copy failed:', fallbackErr);
			}
			document.body.removeChild(textArea);
		}
	}

	$: processedLines = processContent(content);
	$: truncated = maxLines > 0 && content.split('\n').length > maxLines;
</script>

<div class="code-snippet">
	<div class="code-header">
		{#if language}
			<span class="language-badge" style="background-color: {getLanguageColor(language)}">
				{language}
			</span>
		{/if}

		{#if fileName}
			<span class="file-name">{fileName}</span>
		{/if}

		<div class="code-actions">
			<button
				class="copy-button"
				on:click={copyToClipboard}
				title="Copy to clipboard"
				aria-label="Copy code to clipboard"
			>
				{#if copySuccess}
					<svg
						width="16"
						height="16"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
					>
						<path d="M20 6L9 17l-5-5" />
					</svg>
					<span class="sr-only">Copied!</span>
				{:else}
					<svg
						width="16"
						height="16"
						viewBox="0 0 24 24"
						fill="none"
						stroke="currentColor"
						stroke-width="2"
					>
						<rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
						<path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
					</svg>
				{/if}
			</button>
		</div>
	</div>

	<div class="code-content">
		<pre class="code-block" class:line-numbers={showLineNumbers}><code class="language-{language}"
				>{#each processedLines as line, index (index)}
					<span class="code-line" data-line={startLine + index}>
						<!-- eslint-disable-next-line svelte/no-at-html-tags -->
						{@html highlightSearchTerm(line, searchTerm)}
					</span>
				{/each}</code
			></pre>

		{#if truncated}
			<div class="truncated-indicator">
				<span class="truncated-text">
					... truncated ({content.split('\n').length - maxLines} more lines)
				</span>
			</div>
		{/if}
	</div>
</div>

<style>
	.code-snippet {
		border: 1px solid #e5e7eb;
		border-radius: 8px;
		background: #fafafa;
		font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
		overflow: hidden;
	}

	.code-header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		padding: 8px 12px;
		background: #f8f9fa;
		border-bottom: 1px solid #e5e7eb;
		gap: 8px;
	}

	.language-badge {
		display: inline-block;
		padding: 2px 8px;
		border-radius: 12px;
		font-size: 12px;
		font-weight: 500;
		color: white;
		text-transform: capitalize;
		min-width: 0;
		flex-shrink: 0;
	}

	.file-name {
		font-size: 12px;
		color: #6b7280;
		font-weight: 500;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		flex: 1;
	}

	.code-actions {
		display: flex;
		align-items: center;
		gap: 4px;
	}

	.copy-button {
		background: none;
		border: none;
		color: #6b7280;
		cursor: pointer;
		padding: 4px;
		border-radius: 4px;
		transition:
			color 0.2s,
			background-color 0.2s;
		display: flex;
		align-items: center;
		justify-content: center;
	}

	.copy-button:hover {
		color: #374151;
		background-color: #e5e7eb;
	}

	.copy-button:focus {
		outline: 2px solid #3b82f6;
		outline-offset: 2px;
	}

	.code-content {
		position: relative;
	}

	.code-block {
		margin: 0;
		padding: 16px;
		background: white;
		overflow-x: auto;
		font-size: 10px; /* slightly smaller text */
		line-height: 1; /* tighter line spacing */
		color: #1f2937;
	}

	.code-block.line-numbers {
		padding-left: 60px;
		position: relative;
	}

	.code-block.line-numbers .code-line::before {
		content: attr(data-line);
		position: absolute;
		left: 16px;
		width: 32px;
		text-align: right;
		color: #9ca3af;
		font-size: 12px;
		user-select: none;
	}

	.code-line {
		display: block;
		position: relative;
		min-height: 1em; /* match new line-height */
	}

	.code-line:empty::before {
		content: ' ';
	}

	.truncated-indicator {
		padding: 8px 16px;
		background: #f8f9fa;
		border-top: 1px solid #e5e7eb;
		text-align: center;
	}

	.truncated-text {
		font-size: 12px;
		color: #6b7280;
		font-style: italic;
	}

	/* Search term highlighting */
	:global(.code-snippet mark) {
		background-color: #fef3c7;
		color: #92400e;
		padding: 1px 2px;
		border-radius: 2px;
		font-weight: 600;
	}

	/* Screen reader only content */
	.sr-only {
		position: absolute;
		width: 1px;
		height: 1px;
		padding: 0;
		margin: -1px;
		overflow: hidden;
		clip: rect(0, 0, 0, 0);
		white-space: nowrap;
		border: 0;
	}

	/* Mobile responsiveness */
	@media (max-width: 640px) {
		.code-header {
			padding: 6px 8px;
		}

		.code-block {
			padding: 12px;
			font-size: 13px;
		}

		.code-block.line-numbers {
			padding-left: 48px;
		}

		.code-block.line-numbers .code-line::before {
			left: 12px;
			width: 28px;
		}

		.file-name {
			font-size: 11px;
		}

		.language-badge {
			font-size: 11px;
			padding: 1px 6px;
		}
	}

	/* High contrast mode support */
	@media (prefers-contrast: high) {
		.code-snippet {
			border-color: #000;
		}

		.code-header {
			background: #f0f0f0;
			border-bottom-color: #000;
		}

		.code-block {
			background: white;
			color: #000;
		}

		:global(.code-snippet mark) {
			background-color: #ffff00;
			color: #000;
		}
	}

	/* Reduced motion support */
	@media (prefers-reduced-motion: reduce) {
		.copy-button {
			transition: none;
		}
	}
</style>
