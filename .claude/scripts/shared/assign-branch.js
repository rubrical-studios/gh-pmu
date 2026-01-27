#!/usr/bin/env node
/**
 * @framework-script 0.33.3
 * assign-branch.js
 *
 * Interactive script to assign issues to branches.
 * Used by /assign-branch slash command.
 *
 * Implements: REQ-007 (Assign-Release Command)
 * Source: PRD/PRD-Release-and-Sprint-Workflow.md
 *
 * Usage:
 *   node assign-branch.js                     # Interactive mode
 *   node assign-branch.js release/v1.0 123    # Direct assignment (fast - skips epic check)
 *   node assign-branch.js release/v1.0 --all  # Assign all backlog issues
 *   node assign-branch.js --add-ready         # Assign all ready issues (immediate)
 *
 * Flags:
 *   --all         Assign all unassigned backlog issues (lists first without flag)
 *   --add-ready   Assign all unassigned ready issues (immediate, no listing)
 *   --check-epic  Force epic detection for single issues (slower)
 *   --timing      Show timing information for performance debugging
 *
 * Note: --all and --add-ready are mutually exclusive.
 *
 * Performance notes:
 *   - gh pmu commands use local caching in .gh-pmu.yml (~50ms when cached)
 *   - Parallel sub-issue count fetching
 *   - Single issue assignment skips epic check (~500ms faster)
 */

const { exec, execSync } = require('child_process');
const { promisify } = require('util');

const execAsync = promisify(exec);

// Timing helper
let showTiming = false;
const timers = {};

function startTimer(name) {
    if (showTiming) timers[name] = Date.now();
}

function endTimer(name) {
    if (showTiming && timers[name]) {
        console.log(`  ⏱ ${name}: ${Date.now() - timers[name]}ms`);
    }
}

// Configuration constants
const LARGE_SELECTION_THRESHOLD = 20;  // Warn when assigning more issues than this
const PARALLEL_BATCH_SIZE = 5;         // Number of issues to assign in parallel
const PARALLEL_THRESHOLD = 3;          // Use parallel assignment above this count

function execSyncSafe(cmd) {
    try {
        return execSync(cmd, { encoding: 'utf-8' }).trim();
    } catch {
        return null;
    }
}

async function execAsyncSafe(cmd) {
    try {
        const { stdout } = await execAsync(cmd, { encoding: 'utf-8' });
        return stdout.trim();
    } catch {
        return null;
    }
}

/**
 * Get the last version from git tags
 */
function getLastVersion() {
    try {
        const tag = execSyncSafe('git describe --tags --abbrev=0 2>/dev/null');
        if (tag) {
            const match = tag.match(/v?(\d+)\.(\d+)\.(\d+)/);
            if (match) {
                return {
                    major: parseInt(match[1], 10),
                    minor: parseInt(match[2], 10),
                    patch: parseInt(match[3], 10),
                    raw: tag
                };
            }
        }
    } catch (err) {
        if (showTiming) console.log(`  ⚠ getLastVersion failed: ${err.message}`);
    }
    return null;
}

/**
 * Get issue labels for a given issue number
 */
async function getIssueLabels(issueNumber) {
    try {
        const result = await execAsyncSafe(`gh issue view ${issueNumber} --json labels -q ".labels[].name"`);
        return result ? result.split('\n').filter(l => l.trim()) : [];
    } catch (err) {
        if (showTiming) console.log(`  ⚠ getIssueLabels(${issueNumber}) failed: ${err.message}`);
    }
    return [];
}

/**
 * Generate branch name suggestions based on context
 */
function generateBranchSuggestions(lastVersion, userInput, labels) {
    const suggestions = [];
    const hasBugLabel = labels.some(l => l.toLowerCase() === 'bug' || l.toLowerCase() === 'hotfix');
    const hasFeatureLabel = labels.some(l => ['enhancement', 'feature', 'epic', 'story'].includes(l.toLowerCase()));

    if (lastVersion) {
        const nextPatch = `v${lastVersion.major}.${lastVersion.minor}.${lastVersion.patch + 1}`;
        suggestions.push({
            branch: `patch/${nextPatch}`,
            description: `Next patch version (bug fixes only)`,
            recommended: hasBugLabel && !hasFeatureLabel
        });

        const nextMinor = `v${lastVersion.major}.${lastVersion.minor + 1}.0`;
        suggestions.push({
            branch: `release/${nextMinor}`,
            description: `Next minor version (new features)`,
            recommended: hasFeatureLabel && !hasBugLabel
        });

        const nextMajor = `v${lastVersion.major + 1}.0.0`;
        suggestions.push({
            branch: `release/${nextMajor}`,
            description: `Next major version (breaking changes)`
        });
    }

    if (userInput && !userInput.startsWith('release/') && !userInput.startsWith('patch/')) {
        const cleanName = userInput.replace(/^v/, '').replace(/[^a-zA-Z0-9.-]/g, '-');
        if (hasBugLabel) {
            suggestions.push({ branch: `patch/${cleanName}`, description: `Your input with patch prefix` });
        } else {
            suggestions.push({ branch: `release/${cleanName}`, description: `Your input with release prefix` });
        }
    }

    if (userInput && (userInput.startsWith('release/') || userInput.startsWith('patch/'))) {
        suggestions.unshift({ branch: userInput, description: `Your specified branch`, recommended: true });
    }

    if (suggestions.length > 0 && !suggestions.some(s => s.recommended)) {
        suggestions[0].recommended = true;
    }

    return suggestions;
}

/**
 * Get sub-issue count - async version for parallel fetching
 */
async function getSubIssueCountAsync(issueNumber) {
    try {
        const result = await execAsyncSafe(`gh pmu sub list ${issueNumber} --json`);
        if (result) {
            const data = JSON.parse(result);
            if (data.children) return data.children.length;
            if (Array.isArray(data)) return data.length;
        }
        return 0;
    } catch (err) {
        if (showTiming) console.log(`  ⚠ getSubIssueCountAsync(${issueNumber}) failed: ${err.message}`);
        return 0;
    }
}

/**
 * Get open branches (gh pmu handles caching internally)
 */
async function getOpenBranches() {
    startTimer('getOpenBranches');
    try {
        const result = await execAsyncSafe('gh pmu branch list');
        if (result) {
            const lines = result.split('\n').slice(2);
            const branches = [];
            for (const line of lines) {
                if (!line.trim()) continue;
                const parts = line.trim().split(/\s+/);
                if (parts.length >= 4) {
                    const version = parts[0];
                    const status = parts[parts.length - 1];
                    if (status === 'Active') {
                        branches.push({ version, name: version });
                    }
                }
            }
            endTimer('getOpenBranches');
            return branches;
        }
    } catch (err) {
        if (showTiming) console.log(`  ⚠ getOpenBranches failed: ${err.message}`);
    }
    endTimer('getOpenBranches');
    return [];
}

/**
 * Get issues by status (gh pmu handles caching internally)
 * @param {string} status - Status to query (default: 'backlog')
 */
async function getIssuesByStatus(status = 'backlog') {
    startTimer(`getIssuesByStatus(${status})`);
    try {
        const result = await execAsyncSafe(`gh pmu list --status ${status} --json`);
        if (result) {
            const data = JSON.parse(result);
            const items = data.items || data || [];
            const filtered = items.filter(i => !i.fieldValues?.Branch);
            endTimer(`getIssuesByStatus(${status})`);
            return filtered;
        }
    } catch {
        console.error(`Error fetching ${status} issues`);
    }
    endTimer(`getIssuesByStatus(${status})`);
    return [];
}

async function assignToRelease(issueNumber, release, useCurrent = false) {
    try {
        const releaseArg = useCurrent ? 'current' : `"${release}"`;
        const displayRelease = useCurrent ? `${release} (current)` : release;
        console.log(`  → Assigning #${issueNumber} to ${displayRelease}`);
        const result = await execAsyncSafe(`gh pmu move ${issueNumber} --branch ${releaseArg} 2>&1`);
        if (result && result.includes('unknown flag')) {
            console.log(`    (Note: gh pmu --branch not supported, manual assignment needed)`);
            return false;
        }
        return true;
    } catch (err) {
        if (showTiming) console.log(`  ⚠ assignToRelease(${issueNumber}) failed: ${err.message}`);
        return false;
    }
}

/**
 * Assign all sub-issues of an epic to a release
 * @returns {number} Count of successfully assigned sub-issues
 */
async function assignSubIssuesToRelease(issueNumber, release, useCurrent) {
    const subResult = await execAsyncSafe(`gh pmu sub list ${issueNumber} --json`);
    if (!subResult) return 0;

    try {
        const subData = JSON.parse(subResult);
        const children = subData.children || [];
        const results = await Promise.all(
            children.map(sub => assignToRelease(sub.number || sub, release, useCurrent))
        );
        return results.filter(Boolean).length;
    } catch (err) {
        if (showTiming) console.log(`  ⚠ Failed to parse sub-issues for #${issueNumber}: ${err.message}`);
        return 0;
    }
}

async function main() {
    let args = process.argv.slice(2);

    // Handle space-separated arguments passed as a single string (from skill invocation)
    // e.g., "#789 #790 #788" gets passed as one arg instead of three separate args
    if (args.length === 1 && args[0].includes(' ')) {
        args = args[0].split(/\s+/).filter(a => a.trim());
    }

    // Parse args - order-independent parsing
    const assignAll = args.includes('--all');
    const addReady = args.includes('--add-ready');
    const checkEpic = args.includes('--check-epic');
    showTiming = args.includes('--timing');

    // Mutual exclusion check
    if (assignAll && addReady) {
        console.log('Error: --all and --add-ready are mutually exclusive.');
        process.exit(1);
    }

    // Auto-detect: arguments starting with # are issues, release/patch/hotfix prefixes are releases
    let release = args.find(a => a.startsWith('release/') || a.startsWith('patch/') || a.startsWith('hotfix/'));
    let issueNumbers = args.filter(a => a.match(/^#?\d+$/)).map(a => parseInt(a.replace('#', ''), 10));
    const userInput = args.find(a => !a.startsWith('-') && !a.startsWith('release/') && !a.startsWith('patch/') && !a.startsWith('hotfix/') && !a.match(/^#?\d+$/));

    console.log('=== Assign-Branch ===\n');
    startTimer('total');

    // Step 1: Get releases (gh pmu handles caching internally)
    const releases = await getOpenBranches();
    const currentBranch = releases.length === 1 ? releases[0].version : null;

    // Step 2: Handle no releases case
    if (releases.length === 0) {
        console.log('NO_RELEASE_FOUND');
        console.log('');

        const lastVersion = getLastVersion();
        const labels = issueNumbers.length > 0 ? await getIssueLabels(issueNumbers[0]) : [];

        console.log(`CONTEXT:`);
        console.log(`  Last version: ${lastVersion ? lastVersion.raw : '(none)'}`);
        if (issueNumbers.length > 0) {
            console.log(`  Issue: #${issueNumbers[0]}`);
            console.log(`  Labels: ${labels.length > 0 ? labels.join(', ') : '(none)'}`);
        }
        if (userInput) {
            console.log(`  User input: ${userInput}`);
        }
        console.log('');

        const suggestions = generateBranchSuggestions(lastVersion, userInput, labels);

        if (suggestions.length > 0) {
            console.log('SUGGESTIONS:');
            suggestions.forEach((s, i) => {
                const recMarker = s.recommended ? ' (recommended)' : '';
                console.log(`${i + 1}|${s.branch}|${s.description}${recMarker}`);
            });
        } else {
            console.log('SUGGESTIONS:');
            console.log('1|release/v1.0.0|Initial release');
            console.log('2|patch/v0.0.1|Initial patch');
        }

        console.log('');
        console.log('ACTION_REQUIRED: Use AskUserQuestion to let user select a branch, then run:');
        console.log('  gh pmu branch start --name "<selected-branch>"');
        endTimer('total');
        return;
    }

    // Step 3: Single-release shortcut - if only one release and issues/flags provided, use current
    if (!release && currentBranch && (issueNumbers.length > 0 || addReady || assignAll)) {
        release = currentBranch;
        console.log(`Using current release: ${release}\n`);
    }

    // Step 4: Show help if no release determined
    if (!release) {
        console.log('Open Releases:');
        releases.forEach((r, i) => {
            const name = r.name || r.version || r;
            const currentMarker = (releases.length === 1) ? ' (current)' : '';
            console.log(`  ${i + 1}. ${name}${currentMarker}`);
        });
        console.log('');
        console.log('Usage:');
        if (releases.length === 1) {
            // Show quick workflow first when there's a current release
            console.log('  /assign-branch --add-ready         # Quick: assign all ready issues');
            console.log('  /assign-branch #issue              # Assign single issue');
            console.log('  /assign-branch #issue #issue ...   # Assign multiple issues');
        } else {
            console.log('  /assign-branch --add-ready         # Assign all ready issues to current');
        }
        console.log('  /assign-branch release/... #issue  # Assign to specific release');
        console.log('  /assign-branch release/... --all   # Assign all backlog issues');
        console.log('');
        console.log('Examples:');
        if (currentBranch) {
            console.log(`  /assign-branch --add-ready          # Most common workflow`);
            console.log(`  /assign-branch #123`);
            console.log(`  /assign-branch #123 #124 #125`);
        }
        console.log(`  /assign-branch ${releases[0].version} #123`);
        console.log(`  /assign-branch ${releases[0].version} --all\n`);
        endTimer('total');
        return;
    }

    // Step 5: Get issues to assign
    if (issueNumbers.length === 0) {
        // Handle --add-ready: assign all Ready issues immediately
        if (addReady) {
            const ready = await getIssuesByStatus('ready');

            if (ready.length === 0) {
                console.log('No unassigned ready issues found.');
                endTimer('total');
                return;
            }

            issueNumbers = ready.map(i => i.number);
            console.log(`Assigning all ${issueNumbers.length} ready issues to ${release}...\n`);
        } else {
            // Handle backlog issues (--all or listing)
            const backlog = await getIssuesByStatus('backlog');

            if (backlog.length === 0) {
                console.log('No unassigned backlog issues found.');
                endTimer('total');
                return;
            }

            if (assignAll) {
                issueNumbers = backlog.map(i => i.number);
                console.log(`Assigning all ${issueNumbers.length} unassigned backlog issues to ${release}...\n`);
            } else {
                // Group by type for display
                const epics = backlog.filter(i => i.labels?.includes('epic'));
                const bugs = backlog.filter(i => i.labels?.includes('bug'));
                const enhancements = backlog.filter(i => i.labels?.includes('enhancement'));
                const stories = backlog.filter(i => i.labels?.includes('story'));
                const other = backlog.filter(i =>
                    !['epic', 'bug', 'enhancement', 'story'].some(l => i.labels?.includes(l))
                );

                console.log(`Unassigned backlog issues (${backlog.length} total):\n`);

                // Fetch sub-issue counts in parallel for epics
                if (epics.length > 0) {
                    console.log('── Epics ──');
                    startTimer('epicSubCounts');
                    const subCounts = await Promise.all(
                        epics.map(i => getSubIssueCountAsync(i.number))
                    );
                    endTimer('epicSubCounts');
                    epics.forEach((i, idx) => {
                        console.log(`  #${i.number} - ${i.title} (${subCounts[idx]} sub-issues)`);
                    });
                }

                if (bugs.length > 0) {
                    console.log('\n── Bugs ──');
                    bugs.forEach(i => console.log(`  #${i.number} - ${i.title}`));
                }

                if (enhancements.length > 0) {
                    console.log('\n── Enhancements ──');
                    enhancements.forEach(i => console.log(`  #${i.number} - ${i.title}`));
                }

                if (stories.length > 0) {
                    console.log('\n── Stories ──');
                    stories.forEach(i => console.log(`  #${i.number} - ${i.title}`));
                }

                if (other.length > 0) {
                    console.log('\n── Other ──');
                    other.forEach(i => console.log(`  #${i.number} - ${i.title}`));
                }

                console.log('\n---');
                console.log(`To assign specific issues: /assign-branch ${release} #N #M ...`);
                console.log(`To assign all: /assign-branch ${release} --all`);
                endTimer('total');
                return;
            }
        }
    }

    // Step 6: Confirm if large selection
    if (issueNumbers.length >= LARGE_SELECTION_THRESHOLD) {
        console.log(`WARNING: About to assign ${issueNumbers.length} issues to ${release}.`);
        console.log('If this is too many, cancel and specify individual issue numbers.\n');
    }

    // Step 7: Assign issues (use --branch current when single release)
    let totalAssigned = 0;
    let epicCount = 0;
    let subIssueCount = 0;

    const shouldCheckEpic = checkEpic || issueNumbers.length > 1 || assignAll || addReady;
    const useCurrent = (currentBranch && release === currentBranch);

    console.log(`Assigning to ${release}${useCurrent ? ' (current)' : ''}:\n`);

    // For parallel assignment (when multiple issues)
    if (issueNumbers.length > PARALLEL_THRESHOLD) {
        startTimer('parallelAssign');
        for (let i = 0; i < issueNumbers.length; i += PARALLEL_BATCH_SIZE) {
            const batch = issueNumbers.slice(i, i + PARALLEL_BATCH_SIZE);
            const results = await Promise.all(
                batch.map(async (num) => {
                    let isEpic = false;
                    if (shouldCheckEpic) {
                        const labels = await getIssueLabels(num);
                        isEpic = labels.includes('epic');
                    }

                    const subAssigned = isEpic ? await assignSubIssuesToRelease(num, release, useCurrent) : 0;

                    const assigned = await assignToRelease(num, release, useCurrent);
                    return { num, isEpic, assigned, subAssigned };
                })
            );

            for (const r of results) {
                if (r.assigned) totalAssigned++;
                if (r.isEpic) {
                    epicCount++;
                    subIssueCount += r.subAssigned;
                    totalAssigned += r.subAssigned;
                }
            }
        }
        endTimer('parallelAssign');
    } else {
        // Sequential for small batches (simpler output)
        for (const num of issueNumbers) {
            let isEpic = false;

            if (shouldCheckEpic) {
                const labels = await getIssueLabels(num);
                isEpic = labels.includes('epic');
            }

            if (isEpic) {
                epicCount++;
                const assigned = await assignSubIssuesToRelease(num, release, useCurrent);
                subIssueCount += assigned;
                totalAssigned += assigned;
            }

            if (await assignToRelease(num, release, useCurrent)) {
                totalAssigned++;
            }
        }
    }

    console.log(`\n✓ ${totalAssigned} issues assigned to ${release}${useCurrent ? ' (current)' : ''}`);
    if (epicCount > 0) {
        console.log(`  (${epicCount} epics with ${subIssueCount} sub-issues)`);
    }
    if (!shouldCheckEpic && issueNumbers.length === 1) {
        console.log(`  (epic check skipped for single issue - use --check-epic if needed)`);
    }
    endTimer('total');
}

// Run main only when executed directly (not when required for testing)
if (require.main === module) {
    main().catch(console.error);
}

// Export functions for testing
module.exports = {
    getLastVersion,
    getIssueLabels,
    generateBranchSuggestions,
    getSubIssueCountAsync,
    getOpenBranches,
    getIssuesByStatus,
    assignToRelease,
    assignSubIssuesToRelease,
    main,
    // Export helpers for testing
    execSyncSafe,
    execAsyncSafe,
    startTimer,
    endTimer,
    // Export configuration constants
    LARGE_SELECTION_THRESHOLD,
    PARALLEL_BATCH_SIZE,
    PARALLEL_THRESHOLD,
    // Export for mocking
    setShowTiming: (value) => { showTiming = value; }
};
