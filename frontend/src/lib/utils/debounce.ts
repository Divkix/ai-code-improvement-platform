// ABOUTME: Debounce utility function for delaying execution until after a period of inactivity
// ABOUTME: Commonly used for search inputs to avoid making too many API calls

/**
 * Creates a debounced function that delays invoking the provided function
 * until after `wait` milliseconds have elapsed since the last time it was invoked.
 *
 * @param func The function to debounce
 * @param wait The number of milliseconds to delay
 * @returns The debounced function
 */
export function debounce<A extends unknown[], R>(
	func: (...args: A) => R,
	wait: number
): (...args: A) => void {
	let timeoutId: ReturnType<typeof setTimeout> | undefined;

	return (...args: A) => {
		clearTimeout(timeoutId);
		timeoutId = setTimeout(() => func(...args), wait);
	};
}
