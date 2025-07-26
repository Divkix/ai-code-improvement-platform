import { defineConfig } from 'vitest/config';

export default defineConfig({
	test: {
		projects: [
			{
				extends: './vite.config.ts',
				test: {
					name: 'server',
					environment: 'node',
					include: ['src/**/*.{test,spec}.{js,ts}'],
					exclude: [
						'src/**/*.svelte.{test,spec}.{js,ts}',
						'src/lib/components/**/*.{test,spec}.{js,ts}',
						'src/routes/**/*.svelte.{test,spec}.{js,ts}'
					],
					setupFiles: ['./vitest-setup-server.ts']
				}
			}
		]
	}
});
