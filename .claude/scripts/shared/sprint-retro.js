#!/usr/bin/env node
// **Version:** 0.29.3
/**
 * sprint-retro.js
 *
 * Run a retrospective for the current sprint.
 * Used by /sprint-retro slash command.
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
    console.log('=== Sprint Retrospective ===\n');

    // Get current microsprint
    const current = exec('gh pmu microsprint current --json');
    if (!current) {
        console.log('No active sprint to retrospect.');
        return;
    }

    try {
        const sprint = JSON.parse(current);
        const name = sprint.name || 'current';

        console.log(`Sprint: ${name}\n`);

        console.log('Retrospective Questions:\n');
        console.log('1. What went well? (Keep doing)');
        console.log('   - Practices that worked');
        console.log('   - Tools that helped');
        console.log('   - Collaboration successes\n');

        console.log('2. What could be improved? (Start doing)');
        console.log('   - Bottlenecks encountered');
        console.log('   - Missing processes');
        console.log('   - Communication gaps\n');

        console.log('3. What should we stop? (Stop doing)');
        console.log('   - Wasteful practices');
        console.log('   - Ineffective processes');
        console.log('   - Unnecessary overhead\n');

        console.log('4. Action items for next sprint:');
        console.log('   - Specific improvements to implement');
        console.log('   - Experiments to try\n');

        console.log('---');
        console.log('To close sprint with retro: gh pmu microsprint close');
        console.log('To skip retro: gh pmu microsprint close --skip-retro');

    } catch {
        console.log('Could not parse sprint data.');
    }
}

main();
