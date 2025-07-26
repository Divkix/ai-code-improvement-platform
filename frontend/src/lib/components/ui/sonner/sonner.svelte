<script lang="ts">
	import { Toaster as Sonner, type ToasterProps as SonnerProps } from 'svelte-sonner';
	import { mode } from 'mode-watcher';

	let { ...restProps }: SonnerProps = $props();
</script>

<Sonner
	theme={mode.current}
	richColors
	class="toaster group"
	style="--normal-bg: var(--color-popover); --normal-text: var(--color-popover-foreground); --normal-border: var(--color-border); --error-bg: var(--color-destructive); --error-text: var(--color-popover-foreground); --success-bg: var(--color-secondary); --success-text: var(--color-popover-foreground);"
	{...restProps}
/>
