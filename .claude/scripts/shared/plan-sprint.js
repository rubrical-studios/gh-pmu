#!/usr/bin/env node
// **Version:** 0.31.0
/**
 * plan-sprint.js
 *
 * Plan a sprint by selecting epics to include.
 * Used by /plan-sprint slash command.
 *
 * Implements: Sprint Commands
 * Source: PRD/PRD-Release-and-Sprint-Workflow.md
 *
 * Usage:
 *   node plan-sprint.js                     # Interactive mode
 *   node plan-sprint.js --name "auth-work"  # Named sprint
 */

const { execSync } = require('child_process');

function exec(cmd) {
    try {
        return execSync(cmd, { encoding: 'utf-8' }).trim();
    } catch (_e) {
        return null;
    }
}

function getEpicsNotInSprint() {
    try {
        // Get all epics in backlog or ready status
        const result = exec('gh pmu list --label epic --json');
        if (result) {
            const data = JSON.parse(result);
            const items = data.items || data || [];
            // Filter to epics not in a sprint
            return items.filter(i =>
                !i.fieldValues?.Microsprint &&
                ['Backlog', 'Ready'].includes(i.fieldValues?.Status)
            );
        }
    } catch {
        // Intentionally ignored
    }
    return [];
}

function getSubIssueCount(issueNumber) {
    try {
        const result = exec(`gh pmu sub list ${issueNumber} --json`);
        if (result) {
            const data = JSON.parse(result);
            if (data.children) return data.children.length;
            if (Array.isArray(data)) return data.length;
        }
    } catch {
        // Intentionally ignored
    }
    return 0;
}

function getCurrentBranch() {
    try {
        const result = exec('gh pmu branch current --json');
        if (result) {
            const data = JSON.parse(result);
            return data.name || data.version || data;
        }
    } catch {
        // Intentionally ignored
    }
    return null;
}

function main() {
    const args = process.argv.slice(2);
    const nameArg = args.find((a, i) => args[i - 1] === '--name');
    const sprintName = nameArg || null;

    console.log('=== Plan Sprint ===\n');

    // Check current branch context
    const branch = getCurrentBranch();
    if (!branch) {
        console.log('No active branch context.');
        console.log('Start a branch first: gh pmu branch start --name "release/vX.Y.Z"');
        console.log('Then use /switch-branch to set the branch.');
        return;
    }
    console.log(`Branch context: ${branch}\n`);

    // Get available epics
    const epics = getEpicsNotInSprint();

    if (epics.length === 0) {
        console.log('No available epics to plan.');
        console.log('All epics are either already in a sprint or in progress.');
        return;
    }

    console.log('Available epics for sprint:\n');

    epics.forEach((epic, i) => {
        const subCount = getSubIssueCount(epic.number);
        const priority = epic.fieldValues?.Priority || 'P2';
        console.log(`  [${i + 1}] #${epic.number} - ${epic.title}`);
        console.log(`      ${subCount} stories | ${priority}`);
    });

    console.log('\n--- To start a sprint ---');
    console.log('1. Start microsprint: gh pmu microsprint start --name "sprint-name"');
    console.log('2. Add epics: gh pmu microsprint add #<epic-number>');
    console.log('\nOr use: /plan-sprint --name "sprint-name"');

    if (sprintName) {
        console.log(`\nCreating sprint: ${sprintName}...`);
        const result = exec(`gh pmu microsprint start --name "${sprintName}"`);
        if (result !== null) {
            console.log(`✓ Sprint "${sprintName}" created`);
            console.log('\nNow add epics with: gh pmu microsprint add #<epic-number>');
        } else {
            console.log('Note: Sprint creation may require gh pmu microsprint support.');
        }
    }
}

main();
