#!/usr/bin/env node
/**
 * analyze-e2e-impact.js - Analyze E2E test impact from commits
 *
 * Maps command directories to E2E test files and identifies which tests
 * may be impacted by recent changes.
 *
 * Run via: node .claude/scripts/e2e/analyze-e2e-impact.js [base-ref]
 *
 * Output: JSON with impactedTests, newCommandsWithoutTests, recommendation
 */

const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

// E2E coverage map: command directories to test files
const E2E_COVERAGE_MAP = {
    'cmd/microsprint': 'test/e2e/microsprint_test.go',
    'cmd/branch': 'test/e2e/branch_test.go',
    'cmd/board': 'test/e2e/board_test.go',
    'cmd/filter': 'test/e2e/filter_test.go',
    'cmd/move': 'test/e2e/workflow_test.go',
    'cmd/sub': 'test/e2e/workflow_test.go',
    'cmd/create': 'test/e2e/workflow_test.go',
    'cmd/list': 'test/e2e/filter_test.go',
};

// Commands that should have E2E tests
const TESTABLE_COMMANDS = [
    'microsprint', 'branch', 'board', 'filter', 'move', 'sub', 'create', 'list'
];

function main() {
    const baseRef = process.argv[2] || 'main';

    try {
        // Get list of changed files since base ref
        const changedFiles = getChangedFiles(baseRef);

        // Analyze which E2E tests are impacted
        const impactedTests = new Set();
        const changedCommands = new Set();

        for (const file of changedFiles) {
            // Check if file is in a command directory
            for (const [cmdDir, testFile] of Object.entries(E2E_COVERAGE_MAP)) {
                if (file.startsWith(cmdDir + '/') || file === cmdDir) {
                    impactedTests.add(testFile);
                    // Extract command name from path
                    const cmdName = cmdDir.replace('cmd/', '');
                    changedCommands.add(cmdName);
                }
            }
        }

        // Check for new commands without tests
        const newCommandsWithoutTests = findNewCommandsWithoutTests(changedFiles);

        // Generate recommendation
        const recommendation = generateRecommendation(
            impactedTests.size,
            newCommandsWithoutTests.length,
            changedCommands
        );

        const result = {
            success: true,
            impactedTests: Array.from(impactedTests),
            changedCommands: Array.from(changedCommands),
            newCommandsWithoutTests,
            recommendation,
            data: {
                baseRef,
                changedFilesCount: changedFiles.length,
                impactedTestsCount: impactedTests.size
            }
        };

        console.log(JSON.stringify(result, null, 2));

    } catch (err) {
        console.log(JSON.stringify({
            success: false,
            message: `E2E impact analysis failed: ${err.message}`,
            impactedTests: [],
            newCommandsWithoutTests: [],
            recommendation: 'Unable to analyze impact. Review all E2E tests manually.'
        }));
        process.exit(1);
    }
}

function getChangedFiles(baseRef) {
    try {
        const output = execSync(`git diff --name-only ${baseRef}...HEAD`, {
            encoding: 'utf8',
            stdio: ['pipe', 'pipe', 'pipe']
        });
        return output.trim().split('\n').filter(Boolean);
    } catch {
        // Fallback: compare to main directly
        try {
            const output = execSync(`git diff --name-only ${baseRef}`, {
                encoding: 'utf8',
                stdio: ['pipe', 'pipe', 'pipe']
            });
            return output.trim().split('\n').filter(Boolean);
        } catch {
            return [];
        }
    }
}

function findNewCommandsWithoutTests(changedFiles) {
    const newCommands = [];
    const cmdDir = path.join(process.cwd(), 'cmd');

    // Look for new command directories in changed files
    for (const file of changedFiles) {
        if (!file.startsWith('cmd/')) continue;

        const parts = file.split('/');
        if (parts.length < 2) continue;

        const cmdName = parts[1].replace('.go', '');

        // Skip if already in testable commands or is a test file
        if (cmdName.endsWith('_test')) continue;
        if (TESTABLE_COMMANDS.includes(cmdName)) continue;

        // Check if it's a new command (file exists but no E2E test)
        const cmdPath = `cmd/${cmdName}`;
        if (!Object.keys(E2E_COVERAGE_MAP).some(k => k.includes(cmdName))) {
            if (!newCommands.includes(cmdName)) {
                newCommands.push(cmdName);
            }
        }
    }

    return newCommands;
}

function generateRecommendation(impactedCount, newCommandsCount, changedCommands) {
    const parts = [];

    if (impactedCount === 0 && newCommandsCount === 0) {
        return 'No E2E test impact detected. Changes do not affect tested commands.';
    }

    if (impactedCount > 0) {
        const cmds = Array.from(changedCommands).join(', ');
        parts.push(`Run E2E tests for ${impactedCount} impacted test file(s). Changed commands: ${cmds}.`);
    }

    if (newCommandsCount > 0) {
        parts.push(`WARNING: ${newCommandsCount} command(s) modified without E2E test coverage.`);
    }

    return parts.join(' ');
}

main();
