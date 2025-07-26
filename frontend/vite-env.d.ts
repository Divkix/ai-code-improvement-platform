/// <reference types="vite/client" />
/// <reference types="@sveltejs/kit" />

interface ImportMetaEnv {
	readonly VITE_API_URL?: string;
	// add other VITE_ prefixed env vars here as needed
}

interface ImportMeta {
	readonly env: ImportMetaEnv;
}
