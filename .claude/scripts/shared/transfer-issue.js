#!/usr/bin/env node
// **Version:** 0.31.0
/**
 * transfer-issue.js
 *
 * Transfer an issue between releases or sprints.
 * Used by /transfer-issue slash command.
 *
 * Implements: REQ-009 (Sprint-Release Integration)
 * Source: PRD/PRD-Release-and-Sprint-Workflow.md
 *
 * Usage:
 *   node transfer-issue.js #123                       # Show current and options
 *   node transfer-issue.js #123 --branch release/v2.0 # Move to different branch
 *   node transfer-issue.js #123 --sprint auth-work    # Move to different sprint
 *   node transfer-issue.js #123 --remove-sprint       # Remove from sprint
 *   node transfer-issue.js #123 --remove-branch       # Remove from branch
 */

const { execSync } = require('child_process');

function exec(cmd) {
    try {
        return execSync(cmd, { encoding: 'utf-8' }).trim();
    } catch (_e) {
        return null;
    }
}

function getIssueDetails(issueNumber) {
    try {
        const result = exec(`gh pmu view ${issueNumber} --json`);
        if (result) {
            return JSON.parse(result);
        }
    } catch {
        // Intentionally ignored
    }
    return null;
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

function main() {
    const args = process.argv.slice(2);

    // Parse arguments
    const issueArg = args.find(a => a.match(/^#?\d+$/));
    const issueNumber = issueArg ? parseInt(issueArg.replace('#', ''), 10) : null;

    const newBranch = args.find((a, i) => args[i - 1] === '--branch');
    const newSprint = args.find((a, i) => args[i - 1] === '--sprint');
    const removeFromSprint = args.includes('--remove-sprint');
    const removeFromBranch = args.includes('--remove-branch');

    console.log('=== Transfer Issue ===\n');

    if (!issueNumber) {
        console.log('Usage: /transfer-issue #123 [options]');
        console.log('\nOptions:');
        console.log('  --branch <name>     Transfer to different branch');
        console.log('  --sprint <name>     Transfer to different sprint');
        console.log('  --remove-sprint     Remove from current sprint');
        console.log('  --remove-branch     Remove from current branch (back to backlog)');
        return;
    }

    // Get current issue details
    const issue = getIssueDetails(issueNumber);
    if (!issue) {
        console.log(`Issue #${issueNumber} not found or error fetching.`);
        return;
    }

    const currentBranch = issue.fieldValues?.Branch || '(none)';
    const currentSprint = issue.fieldValues?.Microsprint || issue.fieldValues?.Sprint || '(none)';

    console.log(`Issue #${issueNumber}: ${issue.title}`);
    console.log(`Current branch: ${currentBranch}`);
    console.log(`Current sprint: ${currentSprint}`);
    console.log('');

    // Handle actions
    if (removeFromBranch) {
        console.log('Removing from branch...');
        // gh pmu may not support removing branch field directly
        // This would need to be implemented in gh pmu
        console.log('Note: Use gh pmu move to update branch assignment.');
        console.log('Example: gh pmu move ' + issueNumber + ' --branch ""');
        return;
    }

    if (removeFromSprint) {
        console.log('Removing from sprint...');
        const result = exec(`gh pmu microsprint remove ${issueNumber}`);
        if (result !== null) {
            console.log(`✓ Issue #${issueNumber} removed from sprint`);
        } else {
            console.log('Note: Use gh pmu microsprint remove to update sprint assignment.');
        }
        return;
    }

    if (newBranch) {
        console.log(`Transferring to branch: ${newBranch}...`);
        const branchName = newBranch.startsWith('release/') || newBranch.startsWith('patch/')
            ? newBranch
            : `release/${newBranch}`;
        const result = exec(`gh pmu move ${issueNumber} --branch "${branchName}"`);
        if (result !== null) {
            console.log(`✓ Issue #${issueNumber} transferred to ${branchName}`);
        } else {
            console.log('Note: Branch transfer may require gh pmu --branch support.');
            console.log(`Manual: gh pmu move ${issueNumber} --branch "${branchName}"`);
        }
        return;
    }

    if (newSprint) {
        console.log(`Transferring to sprint: ${newSprint}...`);
        const result = exec(`gh pmu microsprint add ${issueNumber}`);
        if (result !== null) {
            console.log(`✓ Issue #${issueNumber} added to current sprint`);
        } else {
            console.log('Note: Sprint transfer requires active microsprint.');
            console.log('Use gh pmu microsprint current first, then add the issue.');
        }
        return;
    }

    // No action specified - show transfer options
    console.log('--- Transfer Options ---');
    const branches = getOpenBranches();
    if (branches.length > 0) {
        console.log('\nAvailable branches:');
        branches.forEach(b => {
            const name = b.name || b.version || b;
            const marker = currentBranch === name ? ' ← current' : '';
            console.log(`  - ${name}${marker}`);
        });
    }

    console.log('\nTo transfer:');
    console.log(`  /transfer-issue #${issueNumber} --branch release/vX.Y.Z`);
    console.log(`  /transfer-issue #${issueNumber} --remove-branch`);
    console.log(`  /transfer-issue #${issueNumber} --remove-sprint`);
}

main();
