#!/usr/bin/env node
/**
 * @framework-script v0.16.0
 * @description Parse commits since last tag, categorize by type
 * @checksum sha256:placeholder
 *
 * This script is provided by the framework and may be updated.
 * To customize: copy to .claude/scripts/shared/ and modify.
 */

const { execSync } = require('child_process');

function parseConventionalCommit(message) {
    // Match: type(scope)!: message or type!: message or type: message
    const match = message.match(/^(\w+)(\([\w-]+\))?(!)?: (.+)$/);

    if (!match) {
        return { type: 'other', scope: null, message, breaking: false };
    }

    const [, type, scopeWithParens, bang, msg] = match;
    const scope = scopeWithParens ? scopeWithParens.slice(1, -1) : null;
    const breaking = !!bang || message.includes('BREAKING CHANGE');

    return { type, scope, message: msg, breaking };
}

async function main() {
    try {
        // Get the latest tag
        let lastTag;
        try {
            lastTag = execSync('git describe --tags --abbrev=0', {
                encoding: 'utf8'
            }).trim();
        } catch {
            // No tags yet
            console.log(JSON.stringify({
                success: true,
                message: 'No previous tags found',
                data: {
                    lastTag: null,
                    commits: [],
                    summary: { total: 0, feat: 0, fix: 0, docs: 0, chore: 0, breaking: 0 }
                }
            }));
            return;
        }

        // Get commits since tag
        const rawLog = execSync(`git log ${lastTag}..HEAD --pretty=format:"%H|%s"`, {
            encoding: 'utf8'
        }).trim();

        if (!rawLog) {
            console.log(JSON.stringify({
                success: true,
                message: `No commits since ${lastTag}`,
                data: {
                    lastTag,
                    commits: [],
                    summary: { total: 0, feat: 0, fix: 0, docs: 0, chore: 0, breaking: 0 }
                }
            }));
            return;
        }

        // Parse commits
        const commits = rawLog.split('\n').map(line => {
            const [hash, ...rest] = line.split('|');
            const message = rest.join('|');
            const parsed = parseConventionalCommit(message);
            return { hash: hash.substring(0, 7), ...parsed };
        });

        // Build summary
        const summary = {
            total: commits.length,
            feat: commits.filter(c => c.type === 'feat').length,
            fix: commits.filter(c => c.type === 'fix').length,
            docs: commits.filter(c => c.type === 'docs').length,
            chore: commits.filter(c => c.type === 'chore').length,
            refactor: commits.filter(c => c.type === 'refactor').length,
            test: commits.filter(c => c.type === 'test').length,
            breaking: commits.filter(c => c.breaking).length
        };

        console.log(JSON.stringify({
            success: true,
            message: `Analyzed ${summary.total} commits since ${lastTag}`,
            data: { lastTag, commits, summary }
        }));

    } catch (err) {
        console.log(JSON.stringify({
            success: false,
            message: `Commit analysis failed: ${err.message}`
        }));
        process.exit(1);
    }
}

main();
