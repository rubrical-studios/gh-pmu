#!/usr/bin/env node
/**
 * wait-for-ci.js - Poll CI status until complete
 *
 * Waits for the latest CI workflow run to complete, with exponential backoff.
 *
 * Usage:
 *   node wait-for-ci.js [options]
 *
 * Options:
 *   --timeout <seconds>   Max wait time (default: 300)
 *   --interval <seconds>  Initial polling interval (default: 30)
 *   --repo <owner/repo>   Repository (default: current)
 *   --quiet               Suppress progress output
 *   --help                Show this help message
 *
 * Exit codes:
 *   0 - CI passed
 *   1 - CI failed or timeout
 */

const gh = require('./lib/gh');
const { poll, formatDuration } = require('./lib/poll');
const out = require('./lib/output');

function showHelp() {
    console.log(`
wait-for-ci.js - Poll CI status until complete

Usage:
  node wait-for-ci.js [options]

Options:
  --timeout <seconds>   Max wait time (default: 300)
  --interval <seconds>  Initial polling interval (default: 30)
  --repo <owner/repo>   Repository (default: current)
  --quiet               Suppress progress output
  --help                Show this help message

Exit codes:
  0 - CI passed
  1 - CI failed or timeout

Examples:
  node wait-for-ci.js                    # Wait for CI with defaults
  node wait-for-ci.js --timeout 600      # Wait up to 10 minutes
  node wait-for-ci.js --interval 10      # Poll every 10 seconds initially
`);
}

async function main() {
    const flags = out.parseFlags();

    if (flags.help) {
        showHelp();
        process.exit(0);
    }

    // Parse options
    const timeout = parseInt(out.getFlag(flags.args, '--timeout', '300')) * 1000;
    const interval = parseInt(out.getFlag(flags.args, '--interval', '30')) * 1000;
    const repo = out.getFlag(flags.args, '--repo') || undefined;

    // Check gh availability
    if (!gh.isAvailable()) {
        out.json({
            status: 'error',
            message: 'GitHub CLI (gh) is not available or not authenticated'
        });
        process.exit(1);
    }

    // Get initial run
    let latestRun = gh.getLatestRun(repo);

    if (!latestRun) {
        out.json({
            status: 'error',
            message: 'No workflow runs found'
        });
        process.exit(1);
    }

    const startTime = Date.now();

    // If already completed, fetch full details and return
    if (latestRun.status === 'completed') {
        const fullRun = gh.getRun(latestRun.databaseId, repo);
        outputResult(fullRun, repo, 0);
        process.exit(fullRun.conclusion === 'success' ? 0 : 1);
    }

    // Poll until complete
    out.info(`Waiting for CI run #${latestRun.databaseId} (${latestRun.name})...`);

    try {
        const { result, elapsed } = await poll(
            () => gh.getRun(latestRun.databaseId, repo),
            (run) => run.status === 'completed',
            {
                interval,
                timeout,
                backoff: 1.5,
                maxInterval: 60000,
                onPoll: (run, elapsed) => {
                    out.progress(`⏳ ${run.status} - ${formatDuration(elapsed)} elapsed`);
                }
            }
        );

        out.clearProgress();
        outputResult(result, repo, elapsed);
        process.exit(result.conclusion === 'success' ? 0 : 1);

    } catch (err) {
        out.clearProgress();

        if (err.message.includes('Timeout')) {
            out.json({
                status: 'timeout',
                workflow: latestRun.name,
                runId: latestRun.databaseId,
                message: `Timed out after ${formatDuration(timeout)}`,
                lastStatus: latestRun.status
            });
        } else {
            out.json({
                status: 'error',
                message: err.message
            });
        }
        process.exit(1);
    }
}

function outputResult(run, repo, elapsed) {
    const jobs = (run.jobs || []).map(job => ({
        name: job.name,
        status: job.conclusion || job.status,
        duration: job.completedAt && job.startedAt
            ? formatDuration(new Date(job.completedAt) - new Date(job.startedAt))
            : null
    }));

    out.json({
        status: run.conclusion === 'success' ? 'success' : 'failure',
        workflow: run.name,
        runId: run.databaseId,
        conclusion: run.conclusion,
        duration: formatDuration(elapsed),
        jobs
    });
}

main();
