#!/usr/bin/env node
/**
 * workflow-trigger.js
 *
 * UserPromptSubmit hook that detects workflow trigger prefixes
 * and injects a reminder into Claude's context.
 *
 * Trigger prefixes: bug:, enhancement:, finding:, idea:, proposal:
 */

let input = '';

process.stdin.on('data', chunk => input += chunk);
process.stdin.on('end', () => {
    try {
        const data = JSON.parse(input);
        const prompt = data.prompt || '';

        const match = prompt.match(/^(bug|enhancement|finding|idea|proposal):/i);
        if (match) {
            const triggerType = match[1].toLowerCase();
            // Output JSON format for proper context injection with visual feedback
            const output = {
                systemMessage: `⚡ Workflow trigger detected: "${triggerType}"`,
                hookSpecificOutput: {
                    hookEventName: "UserPromptSubmit",
                    additionalContext: "[WORKFLOW TRIGGER: Create GitHub issue first. Wait for 'work' instruction before implementing.]"
                }
            };
            console.log(JSON.stringify(output));
        }

        process.exit(0);
    } catch (e) {
        // Parse error - allow prompt to proceed
        process.exit(0);
    }
});
