#!/usr/bin/env node
/**
 * analyze-coverage.js - Analyze code coverage for release readiness
 *
 * Runs tests with coverage, parses results, and compares against
 * changes since the last tag to calculate patch coverage.
 *
 * Usage:
 *   node analyze-coverage.js [options]
 *
 * Options:
 *   --since <tag>      Compare against specific tag (default: latest)
 *   --threshold <n>    Minimum patch coverage % (default: 80)
 *   --skip-tests       Skip running tests (use existing coverage.out)
 *   --quiet            Suppress non-JSON output
 *   --help             Show this help message
 *
 * Exit codes:
 *   0 - Success (coverage meets threshold)
 *   1 - Error (test failure, parse error)
 *   2 - Coverage below threshold
 */

const { execSync, spawnSync } = require('child_process');
const fs = require('fs');
const path = require('path');
const git = require('./lib/git');
const out = require('./lib/output');

function showHelp() {
    console.log(`
analyze-coverage.js - Analyze code coverage for release readiness

Usage:
  node analyze-coverage.js [options]

Options:
  --since <tag>      Compare against specific tag (default: latest)
  --threshold <n>    Minimum patch coverage % (default: 80)
  --skip-tests       Skip running tests (use existing coverage.out)
  --quiet            Suppress non-JSON output
  --help             Show this help message

Output format:
  {
    "patchCoverage": 85.5,
    "threshold": 80,
    "passed": true,
    "totalLines": 100,
    "coveredLines": 85,
    "uncoveredLines": 15,
    "files": [
      {
        "path": "cmd/release.go",
        "patchCoverage": 75.0,
        "changedLines": 20,
        "coveredLines": 15,
        "uncoveredLines": [123, 125, 130, 145, 150]
      }
    ],
    "addressableGaps": [
      { "file": "cmd/release.go", "line": 123, "reason": "error handling" }
    ],
    "nonAddressableGaps": [
      { "file": "internal/config/config.go", "line": 440, "reason": "os.Getwd error" }
    ]
  }

Examples:
  node analyze-coverage.js                    # Analyze since latest tag
  node analyze-coverage.js --threshold 85    # Require 85% coverage
  node analyze-coverage.js --skip-tests      # Use existing coverage.out
`);
}

/**
 * Run go test with coverage
 * @returns {boolean} True if tests passed
 */
function runTests() {
    out.info('Running tests with coverage...');
    const result = spawnSync('go', ['test', '-coverprofile=coverage.out', '-covermode=atomic', './...'], {
        encoding: 'utf8',
        stdio: ['inherit', 'pipe', 'pipe'],
        timeout: 300000 // 5 minute timeout
    });

    if (result.error) {
        out.error(`Failed to run tests: ${result.error.message}`);
        return false;
    }

    if (result.status !== 0) {
        out.error('Tests failed');
        if (result.stderr) {
            console.error(result.stderr);
        }
        return false;
    }

    out.success('Tests passed');
    return true;
}

/**
 * Parse coverage.out file using go tool cover
 * @returns {Map<string, Set<number>>} Map of file -> covered line numbers
 */
function parseCoverage() {
    if (!fs.existsSync('coverage.out')) {
        throw new Error('coverage.out not found. Run tests first or use --skip-tests with existing file.');
    }

    const result = spawnSync('go', ['tool', 'cover', '-func=coverage.out'], {
        encoding: 'utf8'
    });

    if (result.error || result.status !== 0) {
        throw new Error('Failed to parse coverage: ' + (result.error?.message || result.stderr));
    }

    // Parse the -func output to get function-level coverage
    // This is a simplified approach - for line-level we'd need to parse coverage.out directly
    const coverage = new Map();
    const lines = result.stdout.split('\n');

    for (const line of lines) {
        if (!line.trim() || line.includes('total:')) continue;
        const match = line.match(/^(.+):(\d+):\s+(\S+)\s+([\d.]+)%$/);
        if (match) {
            const [, file, lineNum, funcName, percent] = match;
            if (!coverage.has(file)) {
                coverage.set(file, { functions: [], totalPercent: 0 });
            }
            coverage.get(file).functions.push({
                line: parseInt(lineNum),
                name: funcName,
                percent: parseFloat(percent)
            });
        }
    }

    return coverage;
}

/**
 * Parse coverage.out to get line-level coverage data
 * Format: file:startLine.startCol,endLine.endCol statements count
 * @returns {Map<string, {covered: Set<number>, uncovered: Set<number>}>}
 */
function parseLineCoverage() {
    const content = fs.readFileSync('coverage.out', 'utf8');
    const lines = content.split('\n');
    const coverage = new Map();

    for (const line of lines) {
        if (line.startsWith('mode:') || !line.trim()) continue;

        // Format: github.com/rubrical-studios/gh-pmu/cmd/release.go:123.5,125.2 2 1
        const match = line.match(/^(.+):(\d+)\.\d+,(\d+)\.\d+\s+(\d+)\s+(\d+)$/);
        if (!match) continue;

        const [, fullPath, startLine, endLine, statements, count] = match;
        // Extract relative path (remove module prefix)
        const file = fullPath.replace(/^github\.com\/[^/]+\/[^/]+\//, '');

        if (!coverage.has(file)) {
            coverage.set(file, { covered: new Set(), uncovered: new Set() });
        }

        const start = parseInt(startLine);
        const end = parseInt(endLine);
        const hitCount = parseInt(count);

        for (let i = start; i <= end; i++) {
            if (hitCount > 0) {
                coverage.get(file).covered.add(i);
                coverage.get(file).uncovered.delete(i);
            } else if (!coverage.get(file).covered.has(i)) {
                coverage.get(file).uncovered.add(i);
            }
        }
    }

    return coverage;
}

/**
 * Get changed lines since a tag using git diff
 * @param {string} tag - Git tag to compare against
 * @returns {Map<string, Set<number>>} Map of file -> changed line numbers
 */
function getChangedLines(tag) {
    const changed = new Map();

    try {
        // Get list of changed files
        const diffFiles = execSync(`git diff --name-only ${tag}...HEAD`, { encoding: 'utf8' })
            .trim()
            .split('\n')
            .filter(f => f.endsWith('.go') && !f.endsWith('_test.go'));

        for (const file of diffFiles) {
            if (!file.trim()) continue;

            // Get line-by-line diff
            try {
                const diff = execSync(`git diff ${tag}...HEAD -U0 -- "${file}"`, { encoding: 'utf8' });
                const lines = parseDiffForAddedLines(diff);
                if (lines.size > 0) {
                    changed.set(file, lines);
                }
            } catch {
                // File might have been deleted or is new
                continue;
            }
        }
    } catch (err) {
        throw new Error(`Failed to get git diff: ${err.message}`);
    }

    return changed;
}

/**
 * Parse a unified diff to extract added line numbers
 * @param {string} diff - Unified diff output
 * @returns {Set<number>} Set of added line numbers
 */
function parseDiffForAddedLines(diff) {
    const added = new Set();
    const lines = diff.split('\n');
    let currentLine = 0;

    for (const line of lines) {
        // Match @@ -oldStart,oldCount +newStart,newCount @@
        const hunkMatch = line.match(/^@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@/);
        if (hunkMatch) {
            currentLine = parseInt(hunkMatch[1]);
            continue;
        }

        if (line.startsWith('+') && !line.startsWith('+++')) {
            added.add(currentLine);
            currentLine++;
        } else if (line.startsWith('-') && !line.startsWith('---')) {
            // Removed line, don't increment
        } else if (!line.startsWith('\\')) {
            currentLine++;
        }
    }

    return added;
}

/**
 * Categorize uncovered lines as addressable or non-addressable
 * @param {string} file - File path
 * @param {number[]} lines - Uncovered line numbers
 * @returns {{addressable: Array, nonAddressable: Array}}
 */
function categorizeGaps(file, lines) {
    const addressable = [];
    const nonAddressable = [];

    // Non-addressable patterns (OS-level errors, external dependencies)
    const nonAddressablePatterns = [
        { pattern: /os\.Getwd\(\)/, reason: 'os.Getwd error path' },
        { pattern: /os\.MkdirAll\(/, reason: 'os.MkdirAll error path' },
        { pattern: /os\.OpenFile\(/, reason: 'file system error path' },
        { pattern: /os\.ReadFile\(/, reason: 'file read error path' },
        { pattern: /exec\.Command\(/, reason: 'external command error' },
        { pattern: /http\.Get\(/, reason: 'network error path' },
        { pattern: /defer\s+\w+\.Close\(\)/, reason: 'deferred close' },
    ];

    let fileContent;
    try {
        fileContent = fs.readFileSync(file, 'utf8').split('\n');
    } catch {
        // Can't read file, assume all addressable
        return {
            addressable: lines.map(line => ({ file, line, reason: 'new code' })),
            nonAddressable: []
        };
    }

    for (const lineNum of lines) {
        const lineContent = fileContent[lineNum - 1] || '';
        let isNonAddressable = false;

        for (const { pattern, reason } of nonAddressablePatterns) {
            if (pattern.test(lineContent)) {
                nonAddressable.push({ file, line: lineNum, reason });
                isNonAddressable = true;
                break;
            }
        }

        if (!isNonAddressable) {
            // Determine reason for addressable gap
            let reason = 'new code';
            if (/if\s+err\s*!=\s*nil/.test(lineContent)) {
                reason = 'error handling';
            } else if (/return\s+.*err/.test(lineContent)) {
                reason = 'error return';
            } else if (/case\s+/.test(lineContent)) {
                reason = 'switch case';
            } else if (/else\s*{/.test(lineContent)) {
                reason = 'else branch';
            }

            addressable.push({ file, line: lineNum, reason });
        }
    }

    return { addressable, nonAddressable };
}

function main() {
    const flags = out.parseFlags();

    if (flags.help) {
        showHelp();
        process.exit(0);
    }

    if (!git.isGitRepo()) {
        out.json({ error: 'Not a git repository' });
        process.exit(1);
    }

    const threshold = parseInt(out.getFlag(flags.args, '--threshold', '80'));
    const skipTests = out.hasFlag(flags.args, '--skip-tests');
    let tag = out.getFlag(flags.args, '--since');

    if (!tag) {
        tag = git.getLatestTag();
    }

    if (!tag) {
        out.json({
            error: 'No tags found. Create a tag first or use --since to specify a commit.',
            patchCoverage: 0,
            threshold,
            passed: false
        });
        process.exit(1);
    }

    // Run tests if needed
    if (!skipTests) {
        if (!runTests()) {
            out.json({
                error: 'Tests failed',
                patchCoverage: 0,
                threshold,
                passed: false
            });
            process.exit(1);
        }
    }

    // Parse coverage data
    let lineCoverage;
    try {
        lineCoverage = parseLineCoverage();
    } catch (err) {
        out.json({
            error: err.message,
            patchCoverage: 0,
            threshold,
            passed: false
        });
        process.exit(1);
    }

    // Get changed lines since tag
    let changedLines;
    try {
        changedLines = getChangedLines(tag);
    } catch (err) {
        out.json({
            error: err.message,
            patchCoverage: 0,
            threshold,
            passed: false
        });
        process.exit(1);
    }

    // Calculate patch coverage
    let totalChangedLines = 0;
    let totalCoveredLines = 0;
    const files = [];
    const allAddressableGaps = [];
    const allNonAddressableGaps = [];

    for (const [file, lines] of changedLines) {
        const fileCoverage = lineCoverage.get(file) || { covered: new Set(), uncovered: new Set() };

        let fileChangedLines = 0;
        let fileCoveredLines = 0;
        const uncoveredLines = [];

        for (const line of lines) {
            fileChangedLines++;
            totalChangedLines++;

            if (fileCoverage.covered.has(line)) {
                fileCoveredLines++;
                totalCoveredLines++;
            } else {
                uncoveredLines.push(line);
            }
        }

        if (fileChangedLines > 0) {
            const { addressable, nonAddressable } = categorizeGaps(file, uncoveredLines);
            allAddressableGaps.push(...addressable);
            allNonAddressableGaps.push(...nonAddressable);

            files.push({
                path: file,
                patchCoverage: fileChangedLines > 0 ? Math.round((fileCoveredLines / fileChangedLines) * 1000) / 10 : 100,
                changedLines: fileChangedLines,
                coveredLines: fileCoveredLines,
                uncoveredLines: uncoveredLines.sort((a, b) => a - b)
            });
        }
    }

    const patchCoverage = totalChangedLines > 0
        ? Math.round((totalCoveredLines / totalChangedLines) * 1000) / 10
        : 100;

    const passed = patchCoverage >= threshold;

    const result = {
        since: tag,
        patchCoverage,
        threshold,
        passed,
        totalLines: totalChangedLines,
        coveredLines: totalCoveredLines,
        uncoveredLines: totalChangedLines - totalCoveredLines,
        files: files.sort((a, b) => a.patchCoverage - b.patchCoverage),
        addressableGaps: allAddressableGaps,
        nonAddressableGaps: allNonAddressableGaps,
        summary: {
            addressableCount: allAddressableGaps.length,
            nonAddressableCount: allNonAddressableGaps.length
        }
    };

    out.json(result);

    if (!passed) {
        process.exit(2);
    }

    process.exit(0);
}

main();
