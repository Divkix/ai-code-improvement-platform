<script lang="ts">
	import { Toaster as Sonner, type ToasterProps as SonnerProps } from 'svelte-sonner';
	import { mode } from 'mode-watcher';

	let { ...restProps }: SonnerProps = $props();
</script>

<Sonner
	theme={mode.current}
	richColors
	class="toaster group"
	toastOptions={{
		class: 'bg-popover text-popover-foreground border border-border',
		descriptionClass: 'text-muted-foreground'
	}}
	{...restProps}
/>
