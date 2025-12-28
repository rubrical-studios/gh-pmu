#!/usr/bin/env node
// **Version:** 0.15.2
/**
 * assign-release.js
 *
 * Interactive script to assign issues to releases.
 * Used by /assign-release slash command.
 *
 * Implements: REQ-007 (Assign-Release Command)
 * Source: PRD/PRD-Release-and-Sprint-Workflow.md
 *
 * Usage:
 *   node assign-release.js                     # Interactive mode
 *   node assign-release.js release/v1.0 123    # Direct assignment (fast - skips epic check)
 *   node assign-release.js release/v1.0 --all  # Assign all unassigned to release
 *
 * Flags:
 *   --all         Assign all unassigned backlog issues
 *   --check-epic  Force epic detection for single issues (slower)
 *   --timing      Show timing information for performance debugging
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

function execSyncSafe(cmd) {
    try {
        return execSync(cmd, { encoding: 'utf-8' }).trim();
    } catch (e) {
        return null;
    }
}

async function execAsyncSafe(cmd) {
    try {
        const { stdout } = await execAsync(cmd, { encoding: 'utf-8' });
        return stdout.trim();
    } catch (e) {
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
    } catch {}
    return null;
}

/**
 * Get issue labels for a given issue number
 */
async function getIssueLabels(issueNumber) {
    try {
        const result = await execAsyncSafe(`gh issue view ${issueNumber} --json labels -q ".labels[].name"`);
        return result ? result.split('\n').filter(l => l.trim()) : [];
    } catch {}
    return [];
}

/**
 * Generate branch name suggestions based on context
 */
function generateBranchSuggestions(lastVersion, userInput, labels) {
    const suggestions = [];
    const hasBugLabel = labels.some(l => l.toLowerCase() === 'bug' || l.toLowerCase() === 'hotfix');
    const hasFeatureLabel = labels.some(l => ['enhancement', 'feature', 'epic', 'story'].includes(l.toLowerCase()));

    let userVersionHint = null;
    if (userInput) {
        const match = userInput.match(/v?(\d+)\.(\d+)(?:\.(\d+))?/);
        if (match) {
            userVersionHint = {
                major: parseInt(match[1], 10),
                minor: parseInt(match[2], 10),
                patch: match[3] ? parseInt(match[3], 10) : 0
            };
        }
    }

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
    } catch {
        return 0;
    }
}

/**
 * Get open releases (gh pmu handles caching internally)
 */
async function getOpenReleases() {
    startTimer('getOpenReleases');
    try {
        const result = await execAsyncSafe('gh pmu release list');
        if (result) {
            const lines = result.split('\n').slice(2);
            const releases = [];
            for (const line of lines) {
                if (!line.trim()) continue;
                const parts = line.trim().split(/\s+/);
                if (parts.length >= 4) {
                    const version = parts[0];
                    const status = parts[parts.length - 1];
                    if (status === 'Active') {
                        releases.push({ version, name: version });
                    }
                }
            }
            endTimer('getOpenReleases');
            return releases;
        }
    } catch {}
    endTimer('getOpenReleases');
    return [];
}

/**
 * Get backlog issues (gh pmu handles caching internally)
 */
async function getBacklogIssues() {
    startTimer('getBacklogIssues');
    try {
        const result = await execAsyncSafe('gh pmu list --status backlog --json');
        if (result) {
            const data = JSON.parse(result);
            const items = data.items || data || [];
            const filtered = items.filter(i => !i.fieldValues?.Release);
            endTimer('getBacklogIssues');
            return filtered;
        }
    } catch {
        console.error('Error fetching backlog issues');
    }
    endTimer('getBacklogIssues');
    return [];
}

async function assignToRelease(issueNumber, release, useCurrent = false) {
    try {
        const releaseArg = useCurrent ? 'current' : `"${release}"`;
        const displayRelease = useCurrent ? `${release} (current)` : release;
        console.log(`  → Assigning #${issueNumber} to ${displayRelease}`);
        const result = await execAsyncSafe(`gh pmu move ${issueNumber} --release ${releaseArg} 2>&1`);
        if (result && result.includes('unknown flag')) {
            console.log(`    (Note: gh pmu --release not yet supported, manual assignment needed)`);
            return false;
        }
        return true;
    } catch {
        return false;
    }
}

async function main() {
    const args = process.argv.slice(2);

    // Parse args - order-independent parsing
    const assignAll = args.includes('--all');
    const checkEpic = args.includes('--check-epic');
    showTiming = args.includes('--timing');

    // Auto-detect: arguments starting with # are issues, release/patch/hotfix prefixes are releases
    let release = args.find(a => a.startsWith('release/') || a.startsWith('patch/') || a.startsWith('hotfix/'));
    let issueNumbers = args.filter(a => a.match(/^#?\d+$/)).map(a => parseInt(a.replace('#', ''), 10));
    const userInput = args.find(a => !a.startsWith('-') && !a.startsWith('release/') && !a.startsWith('patch/') && !a.startsWith('hotfix/') && !a.match(/^#?\d+$/));

    console.log('=== Assign-Release ===\n');
    startTimer('total');

    // Step 1: Get releases (gh pmu handles caching internally)
    const releases = await getOpenReleases();
    const currentRelease = releases.length === 1 ? releases[0].version : null;

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
        console.log('  gh pmu release start --branch "<selected-branch>"');
        endTimer('total');
        return;
    }

    // Step 3: Single-release shortcut - if only one release and issues provided, use current
    if (!release && currentRelease && issueNumbers.length > 0) {
        release = currentRelease;
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
            console.log('  /assign-release #issue              # Assign to current release');
            console.log('  /assign-release #issue #issue ...   # Assign multiple issues');
        }
        console.log('  /assign-release #issue release/...  # Assign to specific release');
        console.log('  /assign-release release/... #issue  # Same (order doesn\'t matter)');
        console.log('  /assign-release release/... --all   # Assign all backlog issues');
        console.log('');
        console.log('Examples:');
        if (currentRelease) {
            console.log(`  /assign-release #123`);
            console.log(`  /assign-release #123 #124 #125`);
        }
        console.log(`  /assign-release ${releases[0].version} #123`);
        console.log(`  /assign-release ${releases[0].version} --all\n`);
        endTimer('total');
        return;
    }

    // Step 5: Get issues to assign
    if (issueNumbers.length === 0) {
        const backlog = await getBacklogIssues();

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
            console.log(`To assign specific issues: /assign-release ${release} #N #M ...`);
            console.log(`To assign all: /assign-release ${release} --all`);
            endTimer('total');
            return;
        }
    }

    // Step 6: Confirm if large selection
    if (issueNumbers.length >= 20) {
        console.log(`WARNING: About to assign ${issueNumbers.length} issues to ${release}.`);
        console.log('If this is too many, cancel and specify individual issue numbers.\n');
    }

    // Step 7: Assign issues (use --release current when single release)
    let totalAssigned = 0;
    let epicCount = 0;
    let subIssueCount = 0;

    const shouldCheckEpic = checkEpic || issueNumbers.length > 1 || assignAll;
    const useCurrent = (currentRelease && release === currentRelease);

    console.log(`Assigning to ${release}${useCurrent ? ' (current)' : ''}:\n`);

    // For parallel assignment (when multiple issues)
    if (issueNumbers.length > 3) {
        startTimer('parallelAssign');
        // Batch assignments in groups of 5 for parallelization
        const batchSize = 5;
        for (let i = 0; i < issueNumbers.length; i += batchSize) {
            const batch = issueNumbers.slice(i, i + batchSize);
            const results = await Promise.all(
                batch.map(async (num) => {
                    let isEpic = false;
                    if (shouldCheckEpic) {
                        const labels = await getIssueLabels(num);
                        isEpic = labels.includes('epic');
                    }

                    let subAssigned = 0;
                    if (isEpic) {
                        const subResult = await execAsyncSafe(`gh pmu sub list ${num} --json`);
                        if (subResult) {
                            try {
                                const subData = JSON.parse(subResult);
                                const children = subData.children || [];
                                const subResults = await Promise.all(
                                    children.map(sub => assignToRelease(sub.number || sub, release, useCurrent))
                                );
                                subAssigned = subResults.filter(Boolean).length;
                            } catch {}
                        }
                    }

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
                const subResult = await execAsyncSafe(`gh pmu sub list ${num} --json`);
                if (subResult) {
                    try {
                        const subData = JSON.parse(subResult);
                        const children = subData.children || [];
                        for (const sub of children) {
                            const subNum = sub.number || sub;
                            if (await assignToRelease(subNum, release, useCurrent)) {
                                subIssueCount++;
                                totalAssigned++;
                            }
                        }
                    } catch {}
                }
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

main().catch(console.error);
