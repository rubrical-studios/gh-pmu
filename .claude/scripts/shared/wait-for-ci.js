#!/usr/bin/env node
/**
 * @framework-script 0.33.2
 * @description Poll GitHub Actions workflow status with timeout
 * @checksum sha256:placeholder
 *
 * This script is provided by the framework and may be updated.
 * To customize: copy to .claude/scripts/shared/ and modify.
 */

const { execSync } = require('child_process');

const TIMEOUT = 300000; // 5 minutes
const POLL_INTERVAL = 60000; // 60 seconds

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

async function main() {
    const startTime = Date.now();

    try {
        // Get latest run
        const runsJson = execSync('gh run list --limit 1 --json databaseId,status,conclusion,name', {
            encoding: 'utf8'
        });

        const runs = JSON.parse(runsJson);
        if (!runs.length) {
            console.log(JSON.stringify({
                success: false,
                message: 'No workflow runs found'
            }));
            process.exit(1);
        }

        let run = runs[0];

        // If already completed, return immediately
        if (run.status === 'completed') {
            outputResult(run);
            return;
        }

        // Poll until complete or timeout
        while (Date.now() - startTime < TIMEOUT) {
            await sleep(POLL_INTERVAL);

            // Check if run is still running
            const checkJson = execSync(`gh run view ${run.databaseId} --json status,conclusion,jobs`, {
                encoding: 'utf8'
            });

            const check = JSON.parse(checkJson);

            if (check.status === 'completed') {
                outputResult({ ...run, ...check });
                return;
            }

            // Still running - continue polling
            console.error(`CI still running... (${Math.round((Date.now() - startTime) / 1000)}s elapsed)`);
        }

        // Timeout
        console.log(JSON.stringify({
            success: false,
            message: `CI timed out after ${Math.round(TIMEOUT / 1000)}s`,
            data: {
                status: 'timeout',
                workflow: run.name,
                runId: run.databaseId
            }
        }));
        process.exit(1);

    } catch (err) {
        console.log(JSON.stringify({
            success: false,
            message: `CI check failed: ${err.message}`
        }));
        process.exit(1);
    }
}

function outputResult(run) {
    const jobs = (run.jobs || []).map(j => ({
        name: j.name,
        status: j.conclusion || j.status
    }));

    const passed = run.conclusion === 'success';

    console.log(JSON.stringify({
        success: passed,
        message: passed ? 'CI passed' : `CI failed: ${run.conclusion}`,
        data: {
            status: run.conclusion,
            workflow: run.name,
            runId: run.databaseId,
            jobs
        }
    }));

    if (!passed) process.exit(1);
}

main();
