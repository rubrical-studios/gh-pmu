#!/usr/bin/env node
/**
 * run-e2e-gate.js - Run E2E tests and report results
 *
 * Runs E2E tests via `go test -tags=e2e` and reports pass/fail status.
 * Sets GOCOVERDIR for coverage collection and merges with unit test coverage.
 *
 * Run via: node .claude/scripts/e2e/run-e2e-gate.js
 *
 * Output: JSON with success, testsRun, testsPassed, duration, coveragePercentage
 * Exit code: 0 on success, 1 on failure
 *
 * Progress output (stderr):
 *   [1/45] RUN  TestMicrosprintLifecycle
 *   [1/45] PASS TestMicrosprintLifecycle (12.3s)
 */

const { spawn, spawnSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const os = require('os');

function main() {
    const startTime = Date.now();
    const coverDir = path.join(os.tmpdir(), 'e2e-coverage-' + Date.now());
    const e2eCoverageFile = path.join(process.cwd(), 'e2e-coverage.out');
    const unitCoverageFile = path.join(process.cwd(), 'coverage.out');
    const htmlReportFile = path.join(process.cwd(), 'coverage.html');

    // Create coverage directory
    fs.mkdirSync(coverDir, { recursive: true });

    // Count total tests first
    const totalTests = countTests();
    if (totalTests > 0) {
        process.stderr.write(`\nE2E Test Suite: ${totalTests} tests found\n\n`);
    }

    // Run E2E tests with streaming output
    runTestsWithProgress(coverDir, totalTests, startTime, e2eCoverageFile, unitCoverageFile, htmlReportFile);
}

function countTests() {
    // Use go test -list to count test functions
    const result = spawnSync('go', [
        'test',
        '-tags=e2e',
        '-list', '.*',
        './test/e2e/...'
    ], {
        encoding: 'utf8',
        cwd: process.cwd(),
        stdio: ['pipe', 'pipe', 'pipe']
    });

    if (result.status !== 0) {
        return 0;
    }

    // Count lines that look like test function names (start with Test)
    const lines = result.stdout.split('\n').filter(l => l.match(/^Test[A-Z]/));
    return lines.length;
}

function runTestsWithProgress(coverDir, totalTests, startTime, e2eCoverageFile, unitCoverageFile, htmlReportFile) {
    let output = '';
    let currentTest = 0;
    let passed = 0;
    let failed = 0;
    let currentTestName = '';

    const proc = spawn('go', [
        'test',
        '-tags=e2e',
        '-v',
        '-count=1',
        './test/e2e/...'
    ], {
        env: {
            ...process.env,
            GOCOVERDIR: coverDir
        },
        cwd: process.cwd(),
        stdio: ['pipe', 'pipe', 'pipe']
    });

    proc.stdout.on('data', (data) => {
        const chunk = data.toString();
        output += chunk;
        processOutputChunk(chunk, totalTests, (info) => {
            if (info.type === 'run') {
                // Only increment counter for top-level tests
                if (!info.isSubtest) {
                    currentTest++;
                }
                currentTestName = info.name;
                const progress = totalTests > 0 ? `[${currentTest}/${totalTests}]` : `[${currentTest}]`;
                const indent = info.isSubtest ? '  ' : '';
                process.stderr.write(`${progress} ${indent}RUN  ${info.name}\n`);
            } else if (info.type === 'pass') {
                // Only count top-level passes
                if (!info.isSubtest) {
                    passed++;
                }
                const progress = totalTests > 0 ? `[${currentTest}/${totalTests}]` : `[${currentTest}]`;
                const indent = info.isSubtest ? '  ' : '';
                process.stderr.write(`${progress} ${indent}PASS ${info.name} (${info.duration})\n`);
            } else if (info.type === 'fail') {
                // Only count top-level failures
                if (!info.isSubtest) {
                    failed++;
                }
                const progress = totalTests > 0 ? `[${currentTest}/${totalTests}]` : `[${currentTest}]`;
                const indent = info.isSubtest ? '  ' : '';
                process.stderr.write(`${progress} ${indent}FAIL ${info.name} (${info.duration})\n`);
            } else if (info.type === 'skip') {
                const progress = totalTests > 0 ? `[${currentTest}/${totalTests}]` : `[${currentTest}]`;
                const indent = info.isSubtest ? '  ' : '';
                process.stderr.write(`${progress} ${indent}SKIP ${info.name} (${info.duration})\n`);
            }
        });
    });

    proc.stderr.on('data', (data) => {
        output += data.toString();
    });

    proc.on('close', (code) => {
        const duration = ((Date.now() - startTime) / 1000).toFixed(2);
        const success = code === 0;

        // Final summary to stderr
        process.stderr.write(`\n${'─'.repeat(50)}\n`);
        process.stderr.write(`Results: ${passed} passed, ${failed} failed (${duration}s)\n\n`);

        finishWithCoverage(success, output, passed, failed, duration, coverDir, e2eCoverageFile, unitCoverageFile, htmlReportFile);
    });

    proc.on('error', (err) => {
        const duration = ((Date.now() - startTime) / 1000).toFixed(2);

        // Clean up temp coverage directory on error
        try {
            fs.rmSync(coverDir, { recursive: true, force: true });
        } catch (cleanupErr) {
            // Ignore cleanup errors
        }

        console.log(JSON.stringify({
            success: false,
            message: `E2E test execution failed: ${err.message}`,
            testsRun: 0,
            testsPassed: 0,
            testsFailed: 0,
            duration: parseFloat(duration),
            coveragePercentage: null,
            coverageMerged: false,
            data: {
                error: err.message
            }
        }, null, 2));

        process.exit(1);
    });
}

function processOutputChunk(chunk, totalTests, callback) {
    const lines = chunk.split('\n');
    for (const line of lines) {
        // Match: === RUN   TestName
        // Only count top-level tests (no "/" in name) for the counter
        const runMatch = line.match(/^=== RUN\s+(\S+)/);
        if (runMatch) {
            const isSubtest = runMatch[1].includes('/');
            callback({ type: 'run', name: runMatch[1], isSubtest });
            continue;
        }

        // Match: --- PASS: TestName (0.00s)
        const passMatch = line.match(/^--- PASS: (\S+) \(([^)]+)\)/);
        if (passMatch) {
            const isSubtest = passMatch[1].includes('/');
            callback({ type: 'pass', name: passMatch[1], duration: passMatch[2], isSubtest });
            continue;
        }

        // Match: --- FAIL: TestName (0.00s)
        const failMatch = line.match(/^--- FAIL: (\S+) \(([^)]+)\)/);
        if (failMatch) {
            const isSubtest = failMatch[1].includes('/');
            callback({ type: 'fail', name: failMatch[1], duration: failMatch[2], isSubtest });
            continue;
        }

        // Match: --- SKIP: TestName (0.00s)
        const skipMatch = line.match(/^--- SKIP: (\S+) \(([^)]+)\)/);
        if (skipMatch) {
            const isSubtest = skipMatch[1].includes('/');
            callback({ type: 'skip', name: skipMatch[1], duration: skipMatch[2], isSubtest });
            continue;
        }
    }
}

function finishWithCoverage(success, output, testsPassed, testsFailed, duration, coverDir, e2eCoverageFile, unitCoverageFile, htmlReportFile) {
    const testsRun = testsPassed + testsFailed;

    // Process coverage data
    let coveragePercentage = null;
    let coverageMerged = false;

    try {
        // Convert E2E coverage to text format
        const convertResult = spawnSync('go', [
            'tool', 'covdata', 'textfmt',
            '-i=' + coverDir,
            '-o=' + e2eCoverageFile
        ], {
            encoding: 'utf8',
            cwd: process.cwd(),
            stdio: ['pipe', 'pipe', 'pipe']
        });

        if (convertResult.status === 0 && fs.existsSync(e2eCoverageFile)) {
            // Check if unit coverage exists
            if (fs.existsSync(unitCoverageFile)) {
                // Merge E2E coverage with unit coverage
                const mergeResult = spawnSync('go', [
                    'tool', 'cover',
                    '-merge',
                    unitCoverageFile,
                    e2eCoverageFile,
                    '-o', unitCoverageFile
                ], {
                    encoding: 'utf8',
                    cwd: process.cwd(),
                    stdio: ['pipe', 'pipe', 'pipe']
                });

                if (mergeResult.status === 0) {
                    coverageMerged = true;
                }
            } else {
                // No unit coverage, use E2E coverage as base
                fs.copyFileSync(e2eCoverageFile, unitCoverageFile);
                coverageMerged = true;
            }

            // Calculate coverage percentage
            coveragePercentage = calculateCoverage(unitCoverageFile);

            // Regenerate HTML report
            if (coverageMerged) {
                spawnSync('go', [
                    'tool', 'cover',
                    '-html=' + unitCoverageFile,
                    '-o', htmlReportFile
                ], {
                    encoding: 'utf8',
                    cwd: process.cwd(),
                    stdio: ['pipe', 'pipe', 'pipe']
                });
            }

            // Clean up E2E-specific coverage file
            if (fs.existsSync(e2eCoverageFile)) {
                fs.unlinkSync(e2eCoverageFile);
            }
        }
    } catch (coverErr) {
        // Coverage processing failed, but tests may still have passed
        process.stderr.write(`Coverage processing warning: ${coverErr.message}\n`);
    }

    // Clean up temp coverage directory
    try {
        fs.rmSync(coverDir, { recursive: true, force: true });
    } catch (cleanupErr) {
        // Ignore cleanup errors
    }

    const response = {
        success,
        message: success
            ? `E2E tests passed (${testsPassed}/${testsRun})`
            : `E2E tests failed (${testsPassed}/${testsRun} passed, ${testsFailed} failed)`,
        testsRun,
        testsPassed,
        testsFailed,
        duration: parseFloat(duration),
        coveragePercentage,
        coverageMerged,
        data: {
            exitCode: success ? 0 : 1,
            output: truncateOutput(output, 2000)
        }
    };

    console.log(JSON.stringify(response, null, 2));

    if (!success) {
        process.exit(1);
    }
}

function calculateCoverage(coverageFile) {
    // Parse coverage.out file to calculate total coverage percentage
    try {
        const data = fs.readFileSync(coverageFile, 'utf8');
        const lines = data.split('\n').filter(l => l && !l.startsWith('mode:'));

        let covered = 0;
        let total = 0;

        for (const line of lines) {
            // Format: file:startLine.startCol,endLine.endCol statements count
            const match = line.match(/(\d+)\.(\d+),(\d+)\.(\d+) (\d+) (\d+)/);
            if (match) {
                const statements = parseInt(match[5]);
                const count = parseInt(match[6]);
                total += statements;
                if (count > 0) covered += statements;
            }
        }

        if (total === 0) return null;
        return Math.round((covered / total) * 100 * 10) / 10;
    } catch (err) {
        return null;
    }
}

function truncateOutput(output, maxLength) {
    if (output.length <= maxLength) {
        return output;
    }
    return output.substring(0, maxLength) + '\n... (truncated)';
}

main();
