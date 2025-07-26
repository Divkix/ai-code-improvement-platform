<!-- ABOUTME: Code snippet display component with syntax highlighting and search term highlighting -->
<!-- ABOUTME: Supports line number display and copy-to-clipboard functionality -->

<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { ScrollArea } from '$lib/components/ui/scroll-area/index.js';
	import { Check, Copy } from 'lucide-svelte';

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

	let copying = false;

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
		if (copying) return;
		copying = true;

		try {
			await navigator.clipboard.writeText(content);
			dispatch('copy', content);
			setTimeout(() => {
				copying = false;
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
				dispatch('copy', content);
				setTimeout(() => {
					copying = false;
				}, 2000);
			} catch (fallbackErr) {
				console.error('Fallback copy failed:', fallbackErr);
				copying = false;
			}
			document.body.removeChild(textArea);
		}
	}

	$: processedLines = processContent(content);
	$: truncated = maxLines > 0 && content.split('\n').length > maxLines;
</script>

<Card.Root class="overflow-hidden font-mono">
	<Card.Header class="flex flex-row items-center justify-between space-y-0 pb-2">
		<div class="flex items-center gap-2">
			{#if language}
				<Badge
					variant="outline"
					class="text-xs"
					style="background-color: {getLanguageColor(language)}; color: white;"
				>
					{language}
				</Badge>
			{/if}

			{#if fileName}
				<span class="truncate text-sm font-medium text-muted-foreground">{fileName}</span>
			{/if}
		</div>

		<Button
			variant="outline"
			size="sm"
			on:click={copyToClipboard}
			disabled={copying}
			title="Copy to clipboard"
			aria-label="Copy code to clipboard"
		>
			{#if copying}
				<Check class="mr-2 h-4 w-4" />
				Copied!
			{:else}
				<Copy class="mr-2 h-4 w-4" />
				Copy
			{/if}
		</Button>
	</Card.Header>
	<Card.Content class="p-0">
		<ScrollArea class="h-full max-h-96">
			<pre class="code-block p-4 text-sm" class:line-numbers={showLineNumbers}><code
					class="language-{language}"
					>{#each processedLines as line, index (index)}
						<span class="code-line" data-line={startLine + index}>
							<!-- eslint-disable-next-line svelte/no-at-html-tags -->
							{@html highlightSearchTerm(line, searchTerm)}
						</span>
					{/each}</code
				></pre>

			{#if truncated}
				<div class="border-t bg-muted/50 px-4 py-2 text-center">
					<span class="text-xs text-muted-foreground italic">
						... truncated ({content.split('\n').length - maxLines} more lines)
					</span>
				</div>
			{/if}
		</ScrollArea>
	</Card.Content>
</Card.Root>

<style>
	.code-block {
		margin: 0;
		background: transparent;
		overflow-x: auto;
		font-size: 13px;
		line-height: 1.4;
		color: hsl(var(--foreground));
		font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
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
		color: hsl(var(--muted-foreground));
		font-size: 12px;
		user-select: none;
	}

	.code-line {
		display: block;
		position: relative;
		min-height: 1.4em;
	}

	.code-line:empty::before {
		content: ' ';
	}

	/* Search term highlighting */
	:global(mark) {
		background-color: #fef3c7;
		color: #92400e;
		padding: 1px 2px;
		border-radius: 2px;
		font-weight: 600;
	}

	/* Mobile responsiveness */
	@media (max-width: 640px) {
		.code-block {
			font-size: 12px;
		}

		.code-block.line-numbers {
			padding-left: 48px;
		}

		.code-block.line-numbers .code-line::before {
			left: 12px;
			width: 28px;
		}
	}

	/* High contrast mode support */
	@media (prefers-contrast: high) {
		:global(mark) {
			background-color: #ffff00;
			color: #000;
		}
	}
</style>
