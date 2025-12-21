#!/usr/bin/env node
/**
 * verify-config.js - Verify .gh-pmu.yml is clean before release
 *
 * Checks if .gh-pmu.yml matches the committed version.
 * Also detects test corruption patterns (e.g., testowner values).
 * Tests or manual runs of `gh pmu init` can modify this file.
 *
 * Usage:
 *   node verify-config.js [options]
 *
 * Options:
 *   --fix     Restore config to committed version if dirty
 *   --quiet   Suppress non-JSON output
 *   --help    Show this help message
 *
 * Exit codes:
 *   0 - Config is clean (or fixed with --fix)
 *   1 - Config is dirty or corrupted (or error occurred)
 */

const fs = require('fs');
const git = require('./lib/git');
const out = require('./lib/output');

const CONFIG_FILE = '.gh-pmu.yml';

// Test patterns that indicate config was corrupted by tests
const TEST_CORRUPTION_PATTERNS = [
    { pattern: /owner:\s*testowner/i, description: 'test owner value' },
    { pattern: /testowner\/testrepo/i, description: 'test repository value' },
    { pattern: /number:\s*1\s*$/m, description: 'test project number' },
];

function showHelp() {
    console.log(`
verify-config.js - Verify .gh-pmu.yml is clean before release

Usage:
  node verify-config.js [options]

Options:
  --fix     Restore config to committed version if dirty
  --quiet   Suppress non-JSON output
  --help    Show this help message

Exit codes:
  0 - Config is clean (or fixed with --fix)
  1 - Config is dirty (or error occurred)

Examples:
  node verify-config.js              # Check if config is clean
  node verify-config.js --fix        # Fix dirty config
  node verify-config.js | jq .       # Parse JSON output
`);
}

/**
 * Check if config file contains test corruption patterns
 * @returns {Array<string>} List of detected corruption patterns
 */
function detectTestCorruption() {
    try {
        const content = fs.readFileSync(CONFIG_FILE, 'utf8');
        const detected = [];

        for (const { pattern, description } of TEST_CORRUPTION_PATTERNS) {
            if (pattern.test(content)) {
                detected.push(description);
            }
        }

        return detected;
    } catch (err) {
        // File doesn't exist or can't be read
        return [];
    }
}

function main() {
    const flags = out.parseFlags();

    if (flags.help) {
        showHelp();
        process.exit(0);
    }

    const shouldFix = out.hasFlag(flags.args, '--fix');

    // Check if we're in a git repo
    if (!git.isGitRepo()) {
        out.json({
            status: 'error',
            file: CONFIG_FILE,
            message: 'Not a git repository'
        });
        process.exit(1);
    }

    // Check for test corruption patterns first (even if not dirty in git)
    const corruptions = detectTestCorruption();
    if (corruptions.length > 0) {
        if (shouldFix) {
            try {
                git.checkout(CONFIG_FILE);
                out.json({
                    status: 'fixed',
                    file: CONFIG_FILE,
                    message: 'Config restored from test corruption',
                    corruptions: corruptions
                });
                process.exit(0);
            } catch (err) {
                out.json({
                    status: 'error',
                    file: CONFIG_FILE,
                    message: `Failed to restore config: ${err.message}`,
                    corruptions: corruptions
                });
                process.exit(1);
            }
        }

        out.json({
            status: 'corrupted',
            file: CONFIG_FILE,
            message: 'Config contains test values (likely corrupted by tests)',
            corruptions: corruptions,
            action: `Run: node verify-config.js --fix`
        });
        process.exit(1);
    }

    // Check if config file is dirty
    const dirty = git.isDirty(CONFIG_FILE);

    if (!dirty) {
        out.json({
            status: 'clean',
            file: CONFIG_FILE,
            message: 'Config matches committed version'
        });
        process.exit(0);
    }

    // Config is dirty
    const diff = git.getDiff(CONFIG_FILE);

    if (shouldFix) {
        try {
            git.checkout(CONFIG_FILE);
            out.json({
                status: 'fixed',
                file: CONFIG_FILE,
                message: 'Config restored to committed version'
            });
            process.exit(0);
        } catch (err) {
            out.json({
                status: 'error',
                file: CONFIG_FILE,
                message: `Failed to restore config: ${err.message}`
            });
            process.exit(1);
        }
    }

    // Report dirty status
    out.json({
        status: 'dirty',
        file: CONFIG_FILE,
        message: 'Config has uncommitted changes',
        diff: diff,
        action: `Run: git checkout ${CONFIG_FILE}`
    });
    process.exit(1);
}

main();
