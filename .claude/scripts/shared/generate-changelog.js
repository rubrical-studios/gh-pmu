#!/usr/bin/env node
// **Version:** 0.20.0
/**
 * generate-changelog.js - Generate CHANGELOG entry from commits
 *
 * Analyzes commits and generates a Keep a Changelog formatted entry.
 * Can accept piped input from analyze-commits.js.
 *
 * Usage:
 *   node generate-changelog.js [options]
 *   node analyze-commits.js | node generate-changelog.js --version v0.8.0
 *
 * Options:
 *   --version <version>   Version for header (required)
 *   --date <date>         Date for header (default: today)
 *   --quiet               Suppress non-output messages
 *   --help                Show this help message
 *
 * Exit codes:
 *   0 - Success
 *   1 - Error
 */

const git = require('./lib/git');
const out = require('./lib/output');

function showHelp() {
    console.log(`
generate-changelog.js - Generate CHANGELOG entry from commits

Usage:
  node generate-changelog.js [options]
  node analyze-commits.js | node generate-changelog.js --version v0.8.0

Options:
  --version <version>   Version for header (required)
  --date <date>         Date for header (default: today YYYY-MM-DD)
  --quiet               Suppress non-output messages
  --help                Show this help message

Output format (Keep a Changelog):
  ## [0.8.0] - 2025-12-14

  ### Added
  - New feature description

  ### Fixed
  - Bug fix description

Examples:
  node generate-changelog.js --version v0.8.0
  node analyze-commits.js | node generate-changelog.js --version v0.8.0
`);
}

async function readStdin() {
    return new Promise((resolve) => {
        let data = '';

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

        setTimeout(() => {
            if (!data) resolve(null);
        }, 100);
    });
}

function formatDate(dateStr) {
    if (dateStr) return dateStr;
    const now = new Date();
    return now.toISOString().split('T')[0];
}

function generateChangelog(commits, version, date) {
    const sections = {
        added: [],
        changed: [],
        deprecated: [],
        removed: [],
        fixed: [],
        security: []
    };

    // Categorize commits
    for (const commit of commits) {
        const entry = `- ${capitalizeFirst(commit.message)}`;

        switch (commit.type) {
            case 'feat':
                sections.added.push(entry);
                break;
            case 'fix':
                sections.fixed.push(entry);
                break;
            case 'docs':
            case 'style':
            case 'refactor':
            case 'perf':
            case 'chore':
                sections.changed.push(entry);
                break;
            case 'security':
                sections.security.push(entry);
                break;
            default:
                // Put other commits in changed
                if (commit.message) {
                    sections.changed.push(entry);
                }
        }

        // Breaking changes get special mention
        if (commit.breaking) {
            const breakingEntry = `- **BREAKING:** ${capitalizeFirst(commit.message)}`;
            if (!sections.changed.includes(breakingEntry)) {
                sections.changed.unshift(breakingEntry);
            }
        }
    }

    // Build markdown
    const cleanVersion = version.replace(/^v/, '');
    let md = `## [${cleanVersion}] - ${date}\n`;

    if (sections.added.length > 0) {
        md += `\n### Added\n${sections.added.join('\n')}\n`;
    }

    if (sections.changed.length > 0) {
        md += `\n### Changed\n${sections.changed.join('\n')}\n`;
    }

    if (sections.deprecated.length > 0) {
        md += `\n### Deprecated\n${sections.deprecated.join('\n')}\n`;
    }

    if (sections.removed.length > 0) {
        md += `\n### Removed\n${sections.removed.join('\n')}\n`;
    }

    if (sections.fixed.length > 0) {
        md += `\n### Fixed\n${sections.fixed.join('\n')}\n`;
    }

    if (sections.security.length > 0) {
        md += `\n### Security\n${sections.security.join('\n')}\n`;
    }

    return md;
}

function capitalizeFirst(str) {
    if (!str) return '';
    return str.charAt(0).toUpperCase() + str.slice(1);
}

async function main() {
    const flags = out.parseFlags();

    if (flags.help) {
        showHelp();
        process.exit(0);
    }

    // Parse options
    const version = out.getFlag(flags.args, '--version');
    const date = formatDate(out.getFlag(flags.args, '--date'));

    if (!version) {
        out.error('Version is required. Use --version <version>');
        process.exit(1);
    }

    // Try to read piped input
    let analysis = await readStdin();

    // If no piped input, analyze commits
    if (!analysis) {
        if (!git.isGitRepo()) {
            out.error('Not a git repository');
            process.exit(1);
        }

        const tag = git.getLatestTag();
        const rawCommits = git.getCommitsSince(tag);
        const commits = rawCommits.map(commit => {
            const parsed = git.parseConventionalCommit(commit.message);
            return {
                hash: commit.hash,
                type: parsed.type,
                scope: parsed.scope,
                message: parsed.message,
                breaking: parsed.breaking
            };
        });

        analysis = { commits };
    }

    // Generate changelog
    const changelog = generateChangelog(analysis.commits, version, date);

    // Output markdown directly (not JSON)
    console.log(changelog);

    process.exit(0);
}

main();
