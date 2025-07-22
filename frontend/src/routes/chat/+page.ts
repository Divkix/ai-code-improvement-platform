import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	// The root layout now handles all authentication checks.
	// This function can be used for page-specific data loading in the future.
	return {};
};
