#!/usr/bin/env node
// **Version:** 0.16.0
/**
 * monitor-release.js - Monitor release pipeline and verify assets
 *
 * Standalone script that outputs JSON to stdout.
 * Run via: node .claude/scripts/close-release/monitor-release.js [version]
 *
 * Polls every 60 seconds with a 10-minute timeout.
 */

const { execSync } = require('child_process');

// Expected release assets for gh-pmu
const EXPECTED_ASSETS = [
    'darwin-amd64',
    'darwin-arm64',
    'linux-amd64',
    'linux-arm64',
    'windows-amd64',
    'windows-arm64',
    'checksums.txt'
];

const TIMEOUT = 600000; // 10 minutes
const POLL_INTERVAL = 60000; // 60 seconds

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

async function main() {
    // Get version from args or latest tag
    let version = process.argv[2];

    if (!version) {
        try {
            version = execSync('git describe --tags --abbrev=0', {
                encoding: 'utf8'
            }).trim();
        } catch {
            console.log(JSON.stringify({
                success: false,
                message: 'Version not provided and no tags found'
            }));
            process.exit(1);
        }
    }

    try {
        // Step 1: Wait for workflow to complete
        const workflowResult = await waitForWorkflow(version);
        if (!workflowResult.success) {
            console.log(JSON.stringify(workflowResult));
            process.exit(1);
        }

        // Step 2: Verify release assets
        const releaseResult = await verifyRelease(version);
        console.log(JSON.stringify(releaseResult));
        if (!releaseResult.success) process.exit(1);

    } catch (err) {
        console.log(JSON.stringify({
            success: false,
            message: `Release monitoring failed: ${err.message}`
        }));
        process.exit(1);
    }
}

async function waitForWorkflow(version) {
    const startTime = Date.now();

    while (Date.now() - startTime < TIMEOUT) {
        try {
            const runsJson = execSync('gh run list --limit 10 --json databaseId,status,conclusion,headBranch', {
                encoding: 'utf8'
            });

            const runs = JSON.parse(runsJson);
            const tagRun = runs.find(r => r.headBranch === version);

            if (tagRun) {
                if (tagRun.status === 'completed') {
                    if (tagRun.conclusion === 'success') {
                        return { success: true };
                    }
                    return {
                        success: false,
                        message: `Workflow failed: ${tagRun.conclusion}`,
                        data: { status: 'workflow-failed' }
                    };
                }
            }
        } catch {
            // gh command failed, continue polling
        }

        console.error(`Waiting for release workflow... (${Math.round((Date.now() - startTime) / 1000)}s)`);
        await sleep(POLL_INTERVAL);
    }

    return {
        success: false,
        message: 'Workflow did not complete in time',
        data: { status: 'timeout' }
    };
}

async function verifyRelease(version) {
    const startTime = Date.now();
    const assetPollInterval = 10000; // Check assets more frequently

    while (Date.now() - startTime < TIMEOUT) {
        try {
            const releaseJson = execSync(`gh release view ${version} --json assets`, {
                encoding: 'utf8'
            });

            const release = JSON.parse(releaseJson);
            const assets = (release.assets || []).map(a => a.name);

            // Check for missing assets
            const missing = EXPECTED_ASSETS.filter(expected =>
                !assets.some(a => a.includes(expected))
            );

            if (missing.length === 0) {
                return {
                    success: true,
                    message: `Release complete with ${assets.length} assets`,
                    data: { version, assets, missing: [] }
                };
            }

            // Assets still uploading, continue polling
        } catch {
            // Release not found yet, continue polling
        }

        await sleep(assetPollInterval);
    }

    return {
        success: false,
        message: 'Release assets not fully uploaded',
        data: { status: 'incomplete', version }
    };
}

main();
