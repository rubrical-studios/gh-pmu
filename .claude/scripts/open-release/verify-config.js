#!/usr/bin/env node
// **Version:** 0.16.0
/**
 * verify-config.js - Verify .gh-pmu.yml is clean
 *
 * Standalone script that outputs JSON to stdout.
 * Run via: node .claude/scripts/open-release/verify-config.js
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

async function main() {
    const configPath = path.join(process.cwd(), '.gh-pmu.yml');

    // Check if config exists
    if (!fs.existsSync(configPath)) {
        console.log(JSON.stringify({
            success: false,
            message: '.gh-pmu.yml not found'
        }));
        process.exit(1);
    }

    // Check if config is modified (dirty)
    try {
        const status = execSync('git status --porcelain .gh-pmu.yml', {
            encoding: 'utf8'
        }).trim();

        if (status) {
            console.log(JSON.stringify({
                success: false,
                message: '.gh-pmu.yml is dirty. Restore or commit changes before release.',
                data: { status }
            }));
            process.exit(1);
        }

        console.log(JSON.stringify({
            success: true,
            message: '.gh-pmu.yml is clean'
        }));
    } catch (err) {
        console.log(JSON.stringify({
            success: false,
            message: `Config verification failed: ${err.message}`
        }));
        process.exit(1);
    }
}

main();
