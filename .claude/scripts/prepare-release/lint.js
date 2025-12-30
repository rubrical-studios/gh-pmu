#!/usr/bin/env node
// **Version:** 0.17.1
/**
 * lint.js - Run golangci-lint before release
 *
 * Standalone script that outputs JSON to stdout.
 * Run via: node .claude/scripts/prepare-release/lint.js
 *
 * Catches lint errors before tagging, preventing failed releases
 * that require tag deletion and re-release.
 */

const { execSync } = require('child_process');

async function main() {
    const timeout = '5m';

    try {
        // Check if golangci-lint is installed
        try {
            execSync('golangci-lint --version', {
                encoding: 'utf8',
                stdio: 'pipe'
            });
        } catch {
            console.log(JSON.stringify({
                success: false,
                message: 'golangci-lint not installed. Install via: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest'
            }));
            process.exit(1);
        }

        // Run golangci-lint
        const output = execSync(`golangci-lint run --timeout=${timeout}`, {
            encoding: 'utf8',
            stdio: 'pipe'
        });

        console.log(JSON.stringify({
            success: true,
            message: 'Lint check passed',
            data: { timeout }
        }));

    } catch (err) {
        // execSync throws on non-zero exit code
        const lintOutput = err.stdout || err.stderr || err.message;

        // Count issues if possible
        const issueCount = (lintOutput.match(/\n/g) || []).length;

        console.log(JSON.stringify({
            success: false,
            message: `Lint check failed with ${issueCount} issue(s). Fix lint errors before release.`,
            data: {
                timeout,
                output: lintOutput.substring(0, 500) // Truncate for readability
            }
        }));
        process.exit(1);
    }
}

main();
