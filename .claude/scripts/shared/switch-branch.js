#!/usr/bin/env node
// **Version:** 0.31.0
/**
 * switch-branch.js
 *
 * Switch between branch and sprint contexts.
 * Used by /switch-branch slash command.
 *
 * Implements: REQ-009 (Sprint-Release Integration)
 * Source: PRD/PRD-Release-and-Sprint-Workflow.md
 *
 * Usage:
 *   node switch-branch.js                    # Interactive mode
 *   node switch-branch.js release/v1.0       # Direct branch switch
 */

const { execSync } = require('child_process');

function exec(cmd) {
    try {
        return execSync(cmd, { encoding: 'utf-8' }).trim();
    } catch (_e) {
        return null;
    }
}

function getOpenBranches() {
    try {
        const result = exec('gh pmu branch list --open --json');
        if (result) {
            const data = JSON.parse(result);
            return data.branches || data.items || data || [];
        }
    } catch {
        // Intentionally ignored
    }
    return [];
}

function getSprints(_release) {
    try {
        const result = exec('gh pmu microsprint list --json');
        if (result) {
            const data = JSON.parse(result);
            const sprints = data.sprints || data.items || data || [];
            // Filter to sprints for this release (if release field exists)
            return sprints;
        }
    } catch {
        // Intentionally ignored
    }
    return [];
}

function getCurrentBranch() {
    return exec('git branch --show-current');
}

function branchExists(branch) {
    const result = exec(`git rev-parse --verify ${branch} 2>/dev/null`);
    return result !== null;
}

function main() {
    const args = process.argv.slice(2);
    let release = args.find(a => a.startsWith('release/') || a.startsWith('patch/') || a.startsWith('hotfix/'));

    console.log('=== Switch Branch ===\n');

    const currentBranch = getCurrentBranch();
    console.log(`Current branch: ${currentBranch}\n`);

    // Step 1: Get branches
    const branches = getOpenBranches();

    if (!release) {
        if (branches.length === 0) {
            console.log('No open branches found.');
            console.log('\nCreate one with: gh pmu branch start --name "release/vX.Y.Z"');
            return;
        }

        console.log('Available Branches:');
        branches.forEach((b, i) => {
            const name = b.name || b.version || b;
            const branch = b.branch || name;
            const marker = currentBranch === branch ? ' ← current' : '';
            console.log(`  [${i + 1}] ${name}${marker}`);
        });

        console.log('\nUsage: /switch-branch <branch>');
        console.log('Example: /switch-branch release/v2.0.0\n');
        return;
    }

    // Step 2: Switch to release branch
    const releaseBranch = release.startsWith('release/') || release.startsWith('patch/') || release.startsWith('hotfix/')
        ? release
        : `release/${release}`;

    if (!branchExists(releaseBranch)) {
        console.log(`Branch '${releaseBranch}' does not exist.`);
        console.log('\nAvailable release branches:');
        const branches = exec('git branch -a | grep -E "(release|patch|hotfix)"');
        if (branches) {
            console.log(branches);
        } else {
            console.log('  (none found)');
        }
        return;
    }

    if (currentBranch === releaseBranch) {
        console.log(`Already on branch '${releaseBranch}'.`);
    } else {
        console.log(`Switching to branch '${releaseBranch}'...`);
        const switchResult = exec(`git checkout ${releaseBranch}`);
        if (switchResult !== null) {
            console.log(`✓ Switched to ${releaseBranch}`);
        } else {
            console.log('✗ Failed to switch branch. Check for uncommitted changes.');
            return;
        }
    }

    // Step 3: Show available sprints
    console.log('\n--- Sprint Context ---');
    const sprints = getSprints(release);

    if (sprints.length === 0) {
        console.log('No active sprints for this release.');
        console.log('\nStart one with: gh pmu microsprint start --name "sprint-name"');
    } else {
        console.log('Active sprints:');
        sprints.forEach((s, i) => {
            const name = s.name || s;
            console.log(`  [${i + 1}] ${name}`);
        });
        console.log('\nJoin a sprint with: gh pmu microsprint current');
    }

    console.log('\n✓ Context switched to branch: ' + release);
}

main();
