// **Version:** 0.29.2
/**
 * lib/output.js - Output formatting utilities
 *
 * Provides consistent output formatting for all scripts.
 */

// Check if we're in quiet mode (set via --quiet flag)
let quietMode = false;

/**
 * Set quiet mode
 * @param {boolean} quiet
 */
function setQuiet(quiet) {
    quietMode = quiet;
}

/**
 * Check if stdout is a TTY (interactive terminal)
 * @returns {boolean}
 */
function isTTY() {
    return process.stdout.isTTY === true;
}

/**
 * Output JSON to stdout
 * @param {object} data - Data to output
 */
function json(data) {
    console.log(JSON.stringify(data, null, 2));
}

/**
 * Output compact JSON to stdout (single line)
 * @param {object} data - Data to output
 */
function jsonCompact(data) {
    console.log(JSON.stringify(data));
}

/**
 * Output error message to stderr
 * @param {string} message - Error message
 */
function error(message) {
    console.error(`Error: ${message}`);
}

/**
 * Output success message (respects quiet mode)
 * @param {string} message - Success message
 */
function success(message) {
    if (!quietMode) {
        console.log(`✓ ${message}`);
    }
}

/**
 * Output info message (respects quiet mode)
 * @param {string} message - Info message
 */
function info(message) {
    if (!quietMode) {
        console.log(`ℹ ${message}`);
    }
}

/**
 * Output warning message (respects quiet mode)
 * @param {string} message - Warning message
 */
function warn(message) {
    if (!quietMode) {
        console.log(`⚠ ${message}`);
    }
}

/**
 * Output progress message (only if TTY, respects quiet mode)
 * @param {string} message - Progress message
 */
function progress(message) {
    if (!quietMode && isTTY()) {
        process.stdout.write(`\r${message}`);
    }
}

/**
 * Clear progress line (only if TTY)
 */
function clearProgress() {
    if (isTTY()) {
        process.stdout.write('\r\x1b[K');
    }
}

/**
 * Output a table-like format
 * @param {Array<object>} rows - Array of objects
 * @param {Array<string>} columns - Column keys to display
 */
function table(rows, columns) {
    if (quietMode || rows.length === 0) return;

    // Calculate column widths
    const widths = {};
    columns.forEach(col => {
        widths[col] = col.length;
        rows.forEach(row => {
            const val = String(row[col] || '');
            widths[col] = Math.max(widths[col], val.length);
        });
    });

    // Print header
    const header = columns.map(col => col.padEnd(widths[col])).join('  ');
    console.log(header);
    console.log(columns.map(col => '-'.repeat(widths[col])).join('  '));

    // Print rows
    rows.forEach(row => {
        const line = columns.map(col => String(row[col] || '').padEnd(widths[col])).join('  ');
        console.log(line);
    });
}

/**
 * Parse common CLI flags from process.argv
 * @returns {{help: boolean, quiet: boolean, args: string[]}}
 */
function parseFlags() {
    const args = process.argv.slice(2);
    const flags = {
        help: false,
        quiet: false,
        args: []
    };

    for (const arg of args) {
        if (arg === '--help' || arg === '-h') {
            flags.help = true;
        } else if (arg === '--quiet' || arg === '-q') {
            flags.quiet = true;
            setQuiet(true);
        } else {
            flags.args.push(arg);
        }
    }

    return flags;
}

/**
 * Get a named flag value from args
 * @param {string[]} args - Arguments array
 * @param {string} flag - Flag name (e.g., '--timeout')
 * @param {string} defaultValue - Default value if not found
 * @returns {string}
 */
function getFlag(args, flag, defaultValue = '') {
    const idx = args.indexOf(flag);
    if (idx !== -1 && idx + 1 < args.length) {
        return args[idx + 1];
    }
    return defaultValue;
}

/**
 * Check if a flag exists in args
 * @param {string[]} args - Arguments array
 * @param {string} flag - Flag name
 * @returns {boolean}
 */
function hasFlag(args, flag) {
    return args.includes(flag);
}

module.exports = {
    setQuiet,
    isTTY,
    json,
    jsonCompact,
    error,
    success,
    info,
    warn,
    progress,
    clearProgress,
    table,
    parseFlags,
    getFlag,
    hasFlag
};
