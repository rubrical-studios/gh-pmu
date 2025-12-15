#!/usr/bin/env node
/**
 * recommend-version.js - Calculate semver bump based on commit analysis
 *
 * Analyzes commits or accepts piped input from analyze-commits.js
 * to recommend the appropriate version bump.
 *
 * Usage:
 *   node recommend-version.js [options]
 *   node analyze-commits.js | node recommend-version.js
 *
 * Options:
 *   --current <version>  Use specified version as base (default: from git tag)
 *   --quiet              Suppress non-JSON output
 *   --help               Show this help message
 *
 * Exit codes:
 *   0 - Success
 *   1 - Error
 */

const semver = require('semver');
const git = require('./lib/git');
const out = require('./lib/output');

function showHelp() {
    console.log(`
recommend-version.js - Calculate semver bump based on commit analysis

Usage:
  node recommend-version.js [options]
  node analyze-commits.js | node recommend-version.js

Options:
  --current <version>  Use specified version as base (default: from git tag)
  --quiet              Suppress non-JSON output
  --help               Show this help message

Semver rules:
  - Breaking changes (feat!:, BREAKING CHANGE) → MAJOR bump
  - New features (feat:) → MINOR bump
  - Bug fixes only (fix:) → PATCH bump

Examples:
  node recommend-version.js                     # Analyze and recommend
  node recommend-version.js --current v1.2.3   # Use specific base version
  node analyze-commits.js | node recommend-version.js  # Piped input
`);
}

async function readStdin() {
    return new Promise((resolve) => {
        let data = '';

        // Check if stdin is a TTY (no piped input)
        if (process.stdin.isTTY) {
            resolve(null);
            return;
        }

        process.stdin.setEncoding('utf8');
        process.stdin.on('data', chunk => data += chunk);
        process.stdin.on('end', () => {
            try {
                resolve(JSON.parse(data));
            } catch {
                resolve(null);
            }
        });

        // Timeout after 100ms if no data
        setTimeout(() => {
            if (!data) resolve(null);
        }, 100);
    });
}

function calculateBump(summary) {
    if (summary.breaking > 0) {
        return {
            bump: 'major',
            reason: `${summary.breaking} breaking change(s) detected`
        };
    }

    if (summary.feat > 0) {
        return {
            bump: 'minor',
            reason: `${summary.feat} new feature(s), no breaking changes`
        };
    }

    if (summary.fix > 0) {
        return {
            bump: 'patch',
            reason: `${summary.fix} bug fix(es) only`
        };
    }

    // Default to patch for docs, chore, etc.
    return {
        bump: 'patch',
        reason: `${summary.total} commit(s), no features or fixes`
    };
}

async function main() {
    const flags = out.parseFlags();

    if (flags.help) {
        showHelp();
        process.exit(0);
    }

    // Try to read piped input first
    let analysis = await readStdin();

    // If no piped input, run analyze-commits logic
    if (!analysis) {
        if (!git.isGitRepo()) {
            out.json({ error: 'Not a git repository' });
            process.exit(1);
        }

        const tag = git.getLatestTag();
        if (!tag) {
            out.json({
                current: '0.0.0',
                recommended: '0.1.0',
                bump: 'minor',
                reason: 'No previous version found, starting at 0.1.0'
            });
            process.exit(0);
        }

        const rawCommits = git.getCommitsSince(tag);
        const commits = rawCommits.map(commit => {
            const parsed = git.parseConventionalCommit(commit.message);
            return { ...parsed, hash: commit.hash };
        });

        // Build summary
        const summary = {
            total: commits.length,
            feat: commits.filter(c => c.type === 'feat').length,
            fix: commits.filter(c => c.type === 'fix').length,
            breaking: commits.filter(c => c.breaking).length
        };

        analysis = { lastTag: tag, summary };
    }

    // Get current version
    let currentVersion = out.getFlag(flags.args, '--current');
    if (!currentVersion) {
        currentVersion = analysis.lastTag || '0.0.0';
    }

    // Clean version (remove 'v' prefix if present)
    const cleanVersion = currentVersion.replace(/^v/, '');

    // Validate version
    if (!semver.valid(cleanVersion)) {
        out.json({
            error: `Invalid version: ${currentVersion}`,
            current: currentVersion
        });
        process.exit(1);
    }

    // Calculate bump
    const { bump, reason } = calculateBump(analysis.summary);

    // Calculate new version
    const recommended = semver.inc(cleanVersion, bump);
    const vPrefix = currentVersion.startsWith('v') ? 'v' : '';

    out.json({
        current: currentVersion,
        recommended: vPrefix + recommended,
        bump,
        reason
    });

    process.exit(0);
}

main();
