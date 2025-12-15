#!/usr/bin/env node
/**
 * monitor-release.js - Monitor release pipeline and verify assets
 *
 * Monitors a tag-triggered release workflow and verifies that all
 * expected assets are uploaded to the release.
 *
 * Usage:
 *   node monitor-release.js --tag <version> [options]
 *
 * Options:
 *   --tag <version>       Tag to monitor (required)
 *   --timeout <seconds>   Max wait time (default: 600)
 *   --interval <seconds>  Initial polling interval (default: 30)
 *   --repo <owner/repo>   Repository (default: current)
 *   --quiet               Suppress progress output
 *   --help                Show this help message
 *
 * Exit codes:
 *   0 - Release complete with all assets
 *   1 - Failed, timeout, or missing assets
 */

const gh = require('./lib/gh');
const { poll, formatDuration, sleep } = require('./lib/poll');
const out = require('./lib/output');

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

function showHelp() {
    console.log(`
monitor-release.js - Monitor release pipeline and verify assets

Usage:
  node monitor-release.js --tag <version> [options]

Options:
  --tag <version>       Tag to monitor (required)
  --timeout <seconds>   Max wait time (default: 600)
  --interval <seconds>  Initial polling interval (default: 30)
  --repo <owner/repo>   Repository (default: current)
  --quiet               Suppress progress output
  --help                Show this help message

Expected assets:
  - darwin-amd64, darwin-arm64 (macOS)
  - linux-amd64, linux-arm64 (Linux)
  - windows-amd64.exe, windows-arm64.exe (Windows)
  - checksums.txt

Examples:
  node monitor-release.js --tag v0.8.0
  node monitor-release.js --tag v0.8.0 --timeout 900
`);
}

async function main() {
    const flags = out.parseFlags();

    if (flags.help) {
        showHelp();
        process.exit(0);
    }

    // Parse options
    const tag = out.getFlag(flags.args, '--tag');
    const timeout = parseInt(out.getFlag(flags.args, '--timeout', '600')) * 1000;
    const interval = parseInt(out.getFlag(flags.args, '--interval', '30')) * 1000;
    const repo = out.getFlag(flags.args, '--repo') || gh.getCurrentRepo();

    if (!tag) {
        out.json({
            status: 'error',
            message: 'Tag is required. Use --tag <version>'
        });
        process.exit(1);
    }

    // Check gh availability
    if (!gh.isAvailable()) {
        out.json({
            status: 'error',
            message: 'GitHub CLI (gh) is not available or not authenticated'
        });
        process.exit(1);
    }

    out.info(`Monitoring release for tag ${tag}...`);

    try {
        // Step 1: Wait for release workflow to complete
        const workflowResult = await waitForWorkflow(tag, repo, timeout, interval);

        if (workflowResult.status !== 'success') {
            out.json(workflowResult);
            process.exit(1);
        }

        // Step 2: Wait for release to appear and verify assets
        out.info('Workflow complete. Verifying release assets...');
        const releaseResult = await verifyRelease(tag, repo, timeout, interval);

        out.json(releaseResult);
        process.exit(releaseResult.status === 'success' ? 0 : 1);

    } catch (err) {
        out.json({
            status: 'error',
            tag,
            message: err.message
        });
        process.exit(1);
    }
}

async function waitForWorkflow(tag, repo, timeout, interval) {
    const startTime = Date.now();

    // Find the release workflow run
    // It might take a moment for the tag push to trigger a workflow
    let run = null;

    try {
        await poll(
            () => {
                const runs = gh.getRuns({ limit: 20, repo });
                // Look for the tag-triggered run
                for (const r of runs) {
                    // The run might be triggered by the tag
                    if (r.headBranch === tag || r.name === 'CI') {
                        // Check if this is recent (within last 5 minutes)
                        const created = new Date(r.createdAt);
                        if (Date.now() - created < 5 * 60 * 1000) {
                            run = r;
                            return true;
                        }
                    }
                }
                return false;
            },
            (found) => found,
            {
                interval: 5000,
                timeout: 60000,
                backoff: 1,
                onPoll: (_, elapsed) => {
                    out.progress(`⏳ Waiting for workflow to start... ${formatDuration(elapsed)}`);
                }
            }
        );
    } catch {
        // If we timeout finding the workflow, check if release already exists
        const release = gh.getRelease(tag, repo);
        if (release) {
            out.clearProgress();
            return {
                status: 'success',
                message: 'Release already exists',
                jobs: []
            };
        }

        return {
            status: 'error',
            tag,
            message: 'Could not find release workflow run'
        };
    }

    out.clearProgress();
    out.info(`Found workflow run #${run.databaseId}`);

    // Wait for workflow to complete
    if (run.status !== 'completed') {
        try {
            const { result, elapsed } = await poll(
                () => gh.getRun(run.databaseId, repo),
                (r) => r.status === 'completed',
                {
                    interval,
                    timeout: timeout - (Date.now() - startTime),
                    backoff: 1.5,
                    maxInterval: 60000,
                    onPoll: (r, elapsed) => {
                        const completedJobs = (r.jobs || []).filter(j => j.status === 'completed').length;
                        const totalJobs = (r.jobs || []).length;
                        out.progress(`⏳ ${r.status} - ${completedJobs}/${totalJobs} jobs - ${formatDuration(elapsed)}`);
                    }
                }
            );

            out.clearProgress();
            run = result;
        } catch (err) {
            out.clearProgress();
            if (err.message.includes('Timeout')) {
                return {
                    status: 'timeout',
                    tag,
                    message: `Workflow timed out after ${formatDuration(timeout)}`,
                    runId: run.databaseId
                };
            }
            throw err;
        }
    }

    // Check workflow result
    const jobs = (run.jobs || []).map(job => ({
        name: job.name,
        status: job.conclusion || job.status
    }));

    if (run.conclusion !== 'success') {
        const failedJobs = jobs.filter(j => j.status === 'failure');
        return {
            status: 'failure',
            tag,
            message: `Workflow failed`,
            runId: run.databaseId,
            jobs,
            failedJobs
        };
    }

    return {
        status: 'success',
        runId: run.databaseId,
        jobs
    };
}

async function verifyRelease(tag, repo, timeout, interval) {
    const startTime = Date.now();

    // Poll for release to appear
    let release = null;

    try {
        const { result } = await poll(
            () => gh.getRelease(tag, repo),
            (r) => r !== null,
            {
                interval: 5000,
                timeout: Math.min(60000, timeout - (Date.now() - startTime)),
                backoff: 1,
                onPoll: (_, elapsed) => {
                    out.progress(`⏳ Waiting for release... ${formatDuration(elapsed)}`);
                }
            }
        );
        release = result;
    } catch {
        out.clearProgress();
        return {
            status: 'error',
            tag,
            message: 'Release not found after workflow completed'
        };
    }

    out.clearProgress();

    // Verify assets
    const assets = (release.assets || []).map(a => a.name);
    const missing = [];

    for (const expected of EXPECTED_ASSETS) {
        const found = assets.some(a => a.includes(expected));
        if (!found) {
            missing.push(expected);
        }
    }

    if (missing.length > 0) {
        return {
            status: 'incomplete',
            tag,
            message: `Missing ${missing.length} expected asset(s)`,
            releaseUrl: release.url,
            assets,
            missing
        };
    }

    return {
        status: 'success',
        tag,
        message: 'Release complete with all assets',
        releaseUrl: release.url,
        assets
    };
}

main();
