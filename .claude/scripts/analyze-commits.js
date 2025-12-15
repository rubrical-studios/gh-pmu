#!/usr/bin/env node
/**
 * analyze-commits.js - Parse and categorize commits since last tag
 *
 * Analyzes commits using conventional commit format and outputs
 * structured data for version recommendation and changelog generation.
 *
 * Usage:
 *   node analyze-commits.js [options]
 *
 * Options:
 *   --since <tag>   Use specified tag instead of latest
 *   --quiet         Suppress non-JSON output
 *   --help          Show this help message
 *
 * Exit codes:
 *   0 - Success
 *   1 - Error (not a git repo, git command failed)
 */

const git = require('./lib/git');
const out = require('./lib/output');

function showHelp() {
    console.log(`
analyze-commits.js - Parse and categorize commits since last tag

Usage:
  node analyze-commits.js [options]

Options:
  --since <tag>   Use specified tag instead of latest
  --quiet         Suppress non-JSON output
  --help          Show this help message

Output format:
  {
    "lastTag": "v0.7.1",
    "commits": [
      { "hash": "abc123", "type": "feat", "scope": "api", "message": "Add endpoint", "breaking": false }
    ],
    "summary": { "total": 5, "feat": 2, "fix": 2, "docs": 1, "breaking": 0 }
  }

Examples:
  node analyze-commits.js                # Analyze since latest tag
  node analyze-commits.js --since v0.6.0 # Analyze since specific tag
  node analyze-commits.js | jq .commits  # Get just commits array
`);
}

function main() {
    const flags = out.parseFlags();

    if (flags.help) {
        showHelp();
        process.exit(0);
    }

    // Check if we're in a git repo
    if (!git.isGitRepo()) {
        out.json({
            error: 'Not a git repository'
        });
        process.exit(1);
    }

    // Get the tag to analyze from
    let tag = out.getFlag(flags.args, '--since');
    if (!tag) {
        tag = git.getLatestTag();
    }

    if (!tag) {
        out.json({
            lastTag: null,
            commits: [],
            summary: {
                total: 0,
                feat: 0,
                fix: 0,
                docs: 0,
                chore: 0,
                refactor: 0,
                test: 0,
                other: 0,
                breaking: 0
            }
        });
        process.exit(0);
    }

    // Get commits since tag
    const rawCommits = git.getCommitsSince(tag);

    // Parse and categorize commits
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

    // Build summary
    const summary = {
        total: commits.length,
        feat: 0,
        fix: 0,
        docs: 0,
        chore: 0,
        refactor: 0,
        test: 0,
        other: 0,
        breaking: 0
    };

    for (const commit of commits) {
        if (commit.breaking) {
            summary.breaking++;
        }

        switch (commit.type) {
            case 'feat':
                summary.feat++;
                break;
            case 'fix':
                summary.fix++;
                break;
            case 'docs':
                summary.docs++;
                break;
            case 'chore':
                summary.chore++;
                break;
            case 'refactor':
                summary.refactor++;
                break;
            case 'test':
                summary.test++;
                break;
            default:
                summary.other++;
        }
    }

    out.json({
        lastTag: tag,
        commits,
        summary
    });

    process.exit(0);
}

main();
