#!/usr/bin/env node
// **Version:** 0.31.0
/**
 * sprint-status.js
 *
 * Show current sprint status with progress metrics.
 * Used by /sprint-status slash command.
 */

const { execSync } = require('child_process');

function exec(cmd) {
    try {
        return execSync(cmd, { encoding: 'utf-8' }).trim();
    } catch {
        return null;
    }
}

function main() {
    console.log('=== Sprint Status ===\n');

    // Get current microsprint
    const current = exec('gh pmu microsprint current --json');
    if (!current) {
        console.log('No active sprint.');
        console.log('Start one with: gh pmu microsprint start --name "sprint-name"');
        return;
    }

    try {
        const sprint = JSON.parse(current);
        const name = sprint.name || 'current';
        const issues = sprint.issues || [];

        console.log(`Sprint: ${name}`);
        console.log(`Issues: ${issues.length}\n`);

        // Count by status
        let done = 0, inProgress = 0, review = 0, pending = 0;

        issues.forEach(issue => {
            const status = issue.status || issue.fieldValues?.Status || 'Backlog';
            if (status === 'Done') done++;
            else if (status === 'In progress') inProgress++;
            else if (status === 'In review') review++;
            else pending++;
        });

        console.log('Progress:');
        console.log(`  ✓ Done: ${done}`);
        console.log(`  → In Progress: ${inProgress}`);
        console.log(`  ⏸ In Review: ${review}`);
        console.log(`  ○ Pending: ${pending}`);

        const total = issues.length;
        const completed = done;
        const percent = total > 0 ? Math.round((completed / total) * 100) : 0;
        console.log(`\nCompletion: ${percent}% (${completed}/${total})`);

    } catch {
        console.log('Could not parse sprint data.');
        console.log('Use: gh pmu microsprint current');
    }
}

main();
