#!/usr/bin/env node
// **Version:** 0.29.3
/**
 * end-sprint.js
 *
 * End the current sprint with review and retrospective.
 * Used by /end-sprint slash command.
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
    const args = process.argv.slice(2);
    const skipRetro = args.includes('--skip-retro');

    console.log('=== End Sprint ===\n');

    // Get current microsprint
    const current = exec('gh pmu microsprint current --json');
    if (!current) {
        console.log('No active sprint to end.');
        return;
    }

    try {
        const sprint = JSON.parse(current);
        const name = sprint.name || 'current';
        const issues = sprint.issues || [];

        console.log(`Ending sprint: ${name}\n`);

        // Check for incomplete issues
        const incomplete = issues.filter(i => {
            const status = i.status || i.fieldValues?.Status;
            return status !== 'Done';
        });

        if (incomplete.length > 0) {
            console.log(`⚠ ${incomplete.length} incomplete issues:`);
            incomplete.forEach(i => {
                const status = i.status || i.fieldValues?.Status || 'Unknown';
                console.log(`  #${i.number} - ${i.title} (${status})`);
            });
            console.log('\nIncomplete issues will remain in backlog.\n');
        }

        // Run sprint review
        console.log('--- Sprint Review ---');
        const done = issues.filter(i => {
            const status = i.status || i.fieldValues?.Status;
            return status === 'Done';
        });
        console.log(`Completed: ${done.length}/${issues.length} issues\n`);

        if (!skipRetro) {
            console.log('--- Retrospective ---');
            console.log('Consider: What went well? What to improve? What to stop?\n');
        }

        // Close sprint
        console.log('Closing sprint...');
        const closeCmd = skipRetro
            ? 'gh pmu microsprint close --skip-retro'
            : 'gh pmu microsprint close';
        const result = exec(closeCmd);

        if (result !== null) {
            console.log(`✓ Sprint "${name}" closed`);
        } else {
            console.log('Note: Run `gh pmu microsprint close` to complete.');
        }

    } catch {
        console.log('Could not parse sprint data.');
        console.log('Use: gh pmu microsprint close');
    }
}

main();
