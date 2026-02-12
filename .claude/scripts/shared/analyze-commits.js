#!/usr/bin/env node
/**
 * @framework-script 0.42.0
 * @description Parse commits since last tag, categorize by type
 * @checksum sha256:placeholder
 *
 * This script is provided by the framework and may be updated.
 * To customize: copy to .claude/scripts/shared/ and modify.
 */

const { execSync } = require('child_process');

function parseConventionalCommit(message) {
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
        // Get last semantic version tag (v*.*.*), ignoring branch names
        let lastTag;
        try {
            lastTag = execSync('git tag -l "v*" --sort=-v:refname', {
                encoding: 'utf8'
            }).trim().split('\n')[0];

            if (!lastTag) {
                throw new Error('No version tags found');
            }
        } catch {
            console.log(JSON.stringify({
                success: true,
                message: 'No previous tags found',
                data: { lastTag: null, commits: [], summary: { total: 0 } }
            }));
            return;
        }

        const rawLog = execSync(`git log ${lastTag}..HEAD --pretty=format:"%H|%s"`, {
            encoding: 'utf8'
        }).trim();

        if (!rawLog) {
            console.log(JSON.stringify({
                success: true,
                message: `No commits since ${lastTag}`,
                data: { lastTag, commits: [], summary: { total: 0 } }
            }));
            return;
        }

        const commits = rawLog.split('\n').map(line => {
            const [hash, ...rest] = line.split('|');
            const message = rest.join('|');
            const parsed = parseConventionalCommit(message);
            return { hash: hash.substring(0, 7), ...parsed };
        });

        const summary = {
            total: commits.length,
            feat: commits.filter(c => c.type === 'feat').length,
            fix: commits.filter(c => c.type === 'fix').length,
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
