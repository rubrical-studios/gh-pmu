#!/usr/bin/env node
// **Version:** 0.20.0
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
 *   node transfer-issue.js #123                      # Show current and options
 *   node transfer-issue.js #123 --release v2.0.0    # Move to different release
 *   node transfer-issue.js #123 --sprint auth-work  # Move to different sprint
 *   node transfer-issue.js #123 --remove-sprint     # Remove from sprint
 *   node transfer-issue.js #123 --remove-release    # Remove from release
 */

const { execSync } = require('child_process');

function exec(cmd) {
    try {
        return execSync(cmd, { encoding: 'utf-8' }).trim();
    } catch (e) {
        return null;
    }
}

function getIssueDetails(issueNumber) {
    try {
        const result = exec(`gh pmu view ${issueNumber} --json`);
        if (result) {
            return JSON.parse(result);
        }
    } catch {}
    return null;
}

function getOpenReleases() {
    try {
        const result = exec('gh pmu release list --open --json');
        if (result) {
            const data = JSON.parse(result);
            return data.releases || data.items || data || [];
        }
    } catch {}
    return [];
}

function main() {
    const args = process.argv.slice(2);

    // Parse arguments
    const issueArg = args.find(a => a.match(/^#?\d+$/));
    const issueNumber = issueArg ? parseInt(issueArg.replace('#', ''), 10) : null;

    const newRelease = args.find((a, i) => args[i - 1] === '--release');
    const newSprint = args.find((a, i) => args[i - 1] === '--sprint');
    const removeFromSprint = args.includes('--remove-sprint');
    const removeFromRelease = args.includes('--remove-release');

    console.log('=== Transfer Issue ===\n');

    if (!issueNumber) {
        console.log('Usage: /transfer-issue #123 [options]');
        console.log('\nOptions:');
        console.log('  --release <name>    Transfer to different release');
        console.log('  --sprint <name>     Transfer to different sprint');
        console.log('  --remove-sprint     Remove from current sprint');
        console.log('  --remove-release    Remove from current release (back to backlog)');
        return;
    }

    // Get current issue details
    const issue = getIssueDetails(issueNumber);
    if (!issue) {
        console.log(`Issue #${issueNumber} not found or error fetching.`);
        return;
    }

    const currentRelease = issue.fieldValues?.Release || '(none)';
    const currentSprint = issue.fieldValues?.Microsprint || issue.fieldValues?.Sprint || '(none)';

    console.log(`Issue #${issueNumber}: ${issue.title}`);
    console.log(`Current release: ${currentRelease}`);
    console.log(`Current sprint: ${currentSprint}`);
    console.log('');

    // Handle actions
    if (removeFromRelease) {
        console.log('Removing from release...');
        // gh pmu may not support removing release field directly
        // This would need to be implemented in gh pmu
        console.log('Note: Use gh pmu move to update release assignment.');
        console.log('Example: gh pmu move ' + issueNumber + ' --release ""');
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

    if (newRelease) {
        console.log(`Transferring to release: ${newRelease}...`);
        const releaseName = newRelease.startsWith('release/') || newRelease.startsWith('patch/')
            ? newRelease
            : `release/${newRelease}`;
        const result = exec(`gh pmu move ${issueNumber} --release "${releaseName}"`);
        if (result !== null) {
            console.log(`✓ Issue #${issueNumber} transferred to ${releaseName}`);
        } else {
            console.log('Note: Release transfer may require gh pmu --release support.');
            console.log(`Manual: gh pmu move ${issueNumber} --release "${releaseName}"`);
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
    const releases = getOpenReleases();
    if (releases.length > 0) {
        console.log('\nAvailable releases:');
        releases.forEach(r => {
            const name = r.name || r.version || r;
            const marker = currentRelease === name ? ' ← current' : '';
            console.log(`  - ${name}${marker}`);
        });
    }

    console.log('\nTo transfer:');
    console.log(`  /transfer-issue #${issueNumber} --release release/vX.Y.Z`);
    console.log(`  /transfer-issue #${issueNumber} --remove-release`);
    console.log(`  /transfer-issue #${issueNumber} --remove-sprint`);
}

main();
