#!/usr/bin/env node
// **Version:** 0.16.0
/**
 * coverage.js - Validate test coverage meets threshold
 *
 * Standalone script that outputs JSON to stdout.
 * Run via: node .claude/scripts/prepare-release/coverage.js
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

async function main() {
    const threshold = 80;
    const coveragePath = path.join(process.cwd(), 'coverage.out');

    try {
        // Run tests with coverage if file doesn't exist
        if (!fs.existsSync(coveragePath)) {
            execSync('go test -coverprofile=coverage.out ./...', {
                encoding: 'utf8',
                stdio: 'pipe'
            });
        }

        // Parse coverage
        const coverageData = fs.readFileSync(coveragePath, 'utf8');
        const lines = coverageData.split('\n').filter(l => l && !l.startsWith('mode:'));

        let covered = 0, total = 0;

        for (const line of lines) {
            const match = line.match(/(\d+)\.(\d+),(\d+)\.(\d+) (\d+) (\d+)/);
            if (match) {
                const statements = parseInt(match[5]);
                const count = parseInt(match[6]);
                total += statements;
                if (count > 0) covered += statements;
            }
        }

        const totalCoverage = total > 0 ? Math.round((covered / total) * 100 * 10) / 10 : 0;
        const passed = totalCoverage >= threshold;

        if (!passed) {
            console.log(JSON.stringify({
                success: false,
                message: `Coverage ${totalCoverage}% is below threshold ${threshold}%`,
                data: { coverage: totalCoverage, threshold, passed: false }
            }));
            process.exit(1);
        }

        console.log(JSON.stringify({
            success: true,
            message: `Coverage ${totalCoverage}% meets threshold ${threshold}%`,
            data: { coverage: totalCoverage, threshold, passed: true }
        }));

    } catch (err) {
        console.log(JSON.stringify({
            success: false,
            message: `Coverage analysis failed: ${err.message}`
        }));
        process.exit(1);
    }
}

main();
