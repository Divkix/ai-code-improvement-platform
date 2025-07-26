// ABOUTME: Tests for debounce utility function
// ABOUTME: Verifies debouncing behavior, timing, and argument handling

import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { debounce } from './debounce';

describe('debounce', () => {
	beforeEach(() => {
		vi.useFakeTimers();
	});

	afterEach(() => {
		vi.restoreAllMocks();
		vi.useRealTimers();
	});

	it('should delay function execution', () => {
		const mockFn = vi.fn();
		const debouncedFn = debounce(mockFn, 100);

		debouncedFn('test');

		expect(mockFn).not.toHaveBeenCalled();

		vi.advanceTimersByTime(100);

		expect(mockFn).toHaveBeenCalledOnce();
		expect(mockFn).toHaveBeenCalledWith('test');
	});

	it('should cancel previous executions when called again', () => {
		const mockFn = vi.fn();
		const debouncedFn = debounce(mockFn, 100);

		debouncedFn('first');
		vi.advanceTimersByTime(50);

		debouncedFn('second');
		vi.advanceTimersByTime(50);

		expect(mockFn).not.toHaveBeenCalled();

		vi.advanceTimersByTime(50);

		expect(mockFn).toHaveBeenCalledOnce();
		expect(mockFn).toHaveBeenCalledWith('second');
	});

	it('should handle multiple arguments', () => {
		const mockFn = vi.fn();
		const debouncedFn = debounce(mockFn, 100);

		debouncedFn('arg1', 'arg2', 'arg3');
		vi.advanceTimersByTime(100);

		expect(mockFn).toHaveBeenCalledWith('arg1', 'arg2', 'arg3');
	});

	it('should handle different delay times', () => {
		const mockFn = vi.fn();
		const debouncedFn = debounce(mockFn, 250);

		debouncedFn('test');
		vi.advanceTimersByTime(200);

		expect(mockFn).not.toHaveBeenCalled();

		vi.advanceTimersByTime(50);

		expect(mockFn).toHaveBeenCalledOnce();
	});

	it('should work with functions that return values', () => {
		const mockFn = vi.fn().mockReturnValue('result');
		const debouncedFn = debounce(mockFn, 100);

		// Note: debounced functions don't return values as they're async
		const result = debouncedFn('test');
		expect(result).toBeUndefined();

		vi.advanceTimersByTime(100);

		expect(mockFn).toHaveBeenCalledWith('test');
	});

	it('should allow multiple debounced functions to work independently', () => {
		const mockFn1 = vi.fn();
		const mockFn2 = vi.fn();
		const debouncedFn1 = debounce(mockFn1, 100);
		const debouncedFn2 = debounce(mockFn2, 200);

		debouncedFn1('fn1');
		debouncedFn2('fn2');

		vi.advanceTimersByTime(100);

		expect(mockFn1).toHaveBeenCalledWith('fn1');
		expect(mockFn2).not.toHaveBeenCalled();

		vi.advanceTimersByTime(100);

		expect(mockFn2).toHaveBeenCalledWith('fn2');
	});

	it('should handle rapid successive calls correctly', () => {
		const mockFn = vi.fn();
		const debouncedFn = debounce(mockFn, 100);

		for (let i = 0; i < 10; i++) {
			debouncedFn(`call-${i}`);
			vi.advanceTimersByTime(20);
		}

		expect(mockFn).not.toHaveBeenCalled();

		vi.advanceTimersByTime(100);

		expect(mockFn).toHaveBeenCalledOnce();
		expect(mockFn).toHaveBeenCalledWith('call-9');
	});

	it('should work with zero delay', () => {
		const mockFn = vi.fn();
		const debouncedFn = debounce(mockFn, 0);

		debouncedFn('test');

		// Even with 0 delay, it should still be async
		expect(mockFn).not.toHaveBeenCalled();

		vi.advanceTimersByTime(0);

		expect(mockFn).toHaveBeenCalledWith('test');
	});

	it('should preserve function context when called', () => {
		const obj = {
			value: 'test',
			method: function (arg: string) {
				return this.value + arg;
			}
		};

		const mockMethod = vi.fn().mockImplementation(obj.method);
		obj.method = mockMethod;

		const debouncedMethod = debounce(obj.method.bind(obj), 100);

		debouncedMethod('-suffix');
		vi.advanceTimersByTime(100);

		expect(mockMethod).toHaveBeenCalledWith('-suffix');
	});
});
