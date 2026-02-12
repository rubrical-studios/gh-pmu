#!/usr/bin/env node
/**
 * @framework-script 0.42.0
 * @description Recommend semver bump based on commit types
 * @checksum sha256:placeholder
 *
 * This script is provided by the framework and may be updated.
 * To customize: copy to .claude/scripts/shared/ and modify.
 */

const { execSync } = require('child_process');

function parseVersion(tag) {
    const match = tag.replace(/^v/, '').match(/^(\d+)\.(\d+)\.(\d+)/);
    if (!match) return null;
    return { major: parseInt(match[1]), minor: parseInt(match[2]), patch: parseInt(match[3]) };
}

function incVersion(v, type) {
    if (type === 'major') return `${v.major + 1}.0.0`;
    if (type === 'minor') return `${v.major}.${v.minor + 1}.0`;
    return `${v.major}.${v.minor}.${v.patch + 1}`;
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
                message: 'No previous version. Recommend v0.1.0 or v1.0.0.',
                data: { recommendedVersion: 'v0.1.0', reason: 'initial-release' }
            }));
            return;
        }

        const current = parseVersion(lastTag);
        if (!current) {
            console.log(JSON.stringify({
                success: false,
                message: `Could not parse version from tag: ${lastTag}`
            }));
            process.exit(1);
        }

        const rawLog = execSync(`git log ${lastTag}..HEAD --pretty=format:"%s"`, {
            encoding: 'utf8'
        }).trim();

        const lines = rawLog ? rawLog.split('\n') : [];
        let hasBreaking = false, hasFeatures = false;

        for (const line of lines) {
            if (line.includes('BREAKING CHANGE') || line.match(/^\w+!:/)) hasBreaking = true;
            if (line.match(/^feat[(:]/)) hasFeatures = true;
        }

        let bump, reason;
        if (hasBreaking) {
            bump = 'major';
            reason = 'breaking change(s)';
        } else if (hasFeatures) {
            bump = 'minor';
            reason = 'new feature(s)';
        } else {
            bump = 'patch';
            reason = 'fixes and maintenance';
        }

        const next = incVersion(current, bump);
        const recommendedVersion = `v${next}`;

        console.log(JSON.stringify({
            success: true,
            message: `Recommend ${recommendedVersion} (${bump}): ${reason}`,
            data: { recommendedVersion, bumpType: bump, reason, lastTag }
        }));

    } catch (err) {
        console.log(JSON.stringify({
            success: false,
            message: `Version recommendation failed: ${err.message}`
        }));
        process.exit(1);
    }
}

main();
