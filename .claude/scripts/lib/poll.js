/**
 * lib/poll.js - Async polling with exponential backoff
 *
 * Provides utilities for polling async operations until a condition is met.
 */

/**
 * Sleep for a specified duration
 * @param {number} ms - Milliseconds to sleep
 * @returns {Promise<void>}
 */
function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

/**
 * Format milliseconds as human-readable duration
 * @param {number} ms - Milliseconds
 * @returns {string} Formatted duration (e.g., "2m 34s")
 */
function formatDuration(ms) {
    const seconds = Math.floor(ms / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);

    if (hours > 0) {
        return `${hours}h ${minutes % 60}m ${seconds % 60}s`;
    }
    if (minutes > 0) {
        return `${minutes}m ${seconds % 60}s`;
    }
    return `${seconds}s`;
}

/**
 * Poll an async function until a condition is met
 *
 * @param {Function} fn - Async function to poll (should return current state)
 * @param {Function} condition - Function that receives fn result, returns true when done
 * @param {object} options - Polling options
 * @param {number} options.interval - Initial interval in ms (default: 5000)
 * @param {number} options.timeout - Max total duration in ms (default: 300000 = 5 min)
 * @param {number} options.backoff - Multiplier for exponential backoff (default: 1.5)
 * @param {number} options.maxInterval - Cap on interval growth in ms (default: 60000)
 * @param {Function} options.onPoll - Callback on each poll (receives result, elapsed)
 * @returns {Promise<{result: any, elapsed: number, polls: number}>}
 * @throws {Error} On timeout
 */
async function poll(fn, condition, options = {}) {
    const {
        interval = 5000,
        timeout = 300000,
        backoff = 1.5,
        maxInterval = 60000,
        onPoll = null
    } = options;

    const startTime = Date.now();
    let currentInterval = interval;
    let polls = 0;

    while (true) {
        const elapsed = Date.now() - startTime;

        // Check timeout
        if (elapsed >= timeout) {
            throw new Error(`Timeout after ${formatDuration(elapsed)}`);
        }

        // Poll the function
        polls++;
        const result = await fn();

        // Call onPoll callback if provided
        if (onPoll) {
            onPoll(result, elapsed, polls);
        }

        // Check if condition is met
        if (condition(result)) {
            return {
                result,
                elapsed,
                polls
            };
        }

        // Wait before next poll (with backoff)
        await sleep(currentInterval);

        // Increase interval with backoff, capped at maxInterval
        currentInterval = Math.min(currentInterval * backoff, maxInterval);
    }
}

/**
 * Poll until a value changes
 *
 * @param {Function} fn - Async function to poll
 * @param {any} initialValue - Initial value to compare against
 * @param {object} options - Polling options (same as poll())
 * @returns {Promise<{result: any, elapsed: number, polls: number}>}
 */
async function pollUntilChanged(fn, initialValue, options = {}) {
    return poll(fn, result => result !== initialValue, options);
}

/**
 * Poll with retries on error
 *
 * @param {Function} fn - Async function that might throw
 * @param {object} options - Options
 * @param {number} options.retries - Max retries (default: 3)
 * @param {number} options.delay - Delay between retries in ms (default: 1000)
 * @returns {Promise<any>} Result from fn
 * @throws {Error} After all retries exhausted
 */
async function retry(fn, options = {}) {
    const { retries = 3, delay = 1000 } = options;

    let lastError;
    for (let i = 0; i <= retries; i++) {
        try {
            return await fn();
        } catch (err) {
            lastError = err;
            if (i < retries) {
                await sleep(delay);
            }
        }
    }

    throw lastError;
}

module.exports = {
    sleep,
    formatDuration,
    poll,
    pollUntilChanged,
    retry
};
