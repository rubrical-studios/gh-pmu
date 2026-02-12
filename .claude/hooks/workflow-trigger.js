#!/usr/bin/env node
/**
 * @framework-script 0.42.0
 * workflow-trigger.js
 *
 * UserPromptSubmit hook that:
 * 1. Detects workflow trigger prefixes and injects reminders
 * 2. Responds to 'commands' with available triggers and slash commands
 * 3. Validates branch assignment for 'work #N' commands
 * 4. Detects analysis keywords and injects STOP reminder (#1056)
 * 5. Auto-moves PRD tracker to in_progress for work commands (#1193)
 * 6. Detects 'done' trigger and routes to /done command (#1200)
 * 7. Detects 'review #N' and routes to /review-issue command (#1210)
 *
 * Trigger prefixes: bug:, enhancement:, idea:, proposal:
 * Work command: work #N (validates branch assignment, provides branch context)
 * Done command: done [#N...] (contextual - only triggers with active issues)
 * Review command: review #N [#N...] (routes to /review-issue)
 * Analysis keywords: evaluate, analyze, assess, investigate, check, verify
 *   - When combined with issue reference, injects STOP reminder
 *   - Prevents analysis requests from drifting into implementation
 *
 * Performance optimizations:
 * - Early exit for non-matching prompts (no I/O)
 * - Single detectFramework() function (no duplication)
 * - Cached command help (regenerated on demand)
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

// Cache file location
const CACHE_FILE = path.join(process.cwd(), '.claude', 'hooks', '.command-cache.json');

// Analysis keywords that trigger STOP-after-report behavior
// When these appear with an issue reference, inject reminder to report only
const ANALYSIS_KEYWORDS = ['evaluate', 'analyze', 'assess', 'investigate', 'check', 'verify'];

let input = '';

process.stdin.on('data', chunk => input += chunk);
process.stdin.on('end', () => {
    try {
        const data = JSON.parse(input);
        const prompt = (data.prompt || '').trim();
        const promptLower = prompt.toLowerCase();

        // FAST PATH: Early exit for non-matching prompts (no I/O needed)
        // Check for --refresh flag
        const hasRefreshFlag = /--refresh/i.test(prompt);
        const basePrompt = promptLower.replace(/\s*--refresh\s*/gi, '').trim();
        const isCommandRequest = ['commands', 'list-commands'].includes(basePrompt);
        const triggerMatch = prompt.match(/^(bug|enhancement|idea|proposal):/i);
        // Broad "work" trigger - matches any prompt starting with "work "
        // We'll opportunistically extract issue numbers or status later
        const workTrigger = prompt.match(/^work\s/i);
        // Done trigger - matches standalone "done" or "done #N" at message start
        // Contextual detection happens later (queries active issues)
        const doneTrigger = promptLower.match(/^done(\s|$)/i);
        // Review trigger - "review" with issue reference routes to /review-issue (#1210)
        // review is a tracked action (not passive analysis like evaluate/assess/etc.)
        const hasReviewKeyword = promptLower.includes('review');
        const hasIssueRef = prompt.match(/#\d+|\bissue\s+\d+|\b\d{2,}\b/i);
        const reviewTrigger = hasReviewKeyword && hasIssueRef && !workTrigger;
        // Analysis trigger - detect analysis keywords with issue references
        // This prevents "evaluate #123" from drifting into implementation
        const hasAnalysisKeyword = ANALYSIS_KEYWORDS.some(kw => promptLower.includes(kw));
        const analysisMatch = hasAnalysisKeyword && hasIssueRef && !workTrigger && !reviewTrigger;

        if (!isCommandRequest && !triggerMatch && !workTrigger && !doneTrigger && !reviewTrigger && !analysisMatch) {
            process.exit(0);
        }

        // Handle 'commands' request
        if (basePrompt === 'commands') {
            const helpText = generateCommandsHelp(hasRefreshFlag);
            const output = {
                systemMessage: `Success`,
                hookSpecificOutput: {
                    hookEventName: "UserPromptSubmit",
                    additionalContext: `[COMMANDS HELP: Display the following commands to the user in a clean formatted way.]\n\n${helpText}`
                }
            };
            console.log(JSON.stringify(output));
            process.exit(0);
        }

        // Handle 'List-Commands' request (full detailed list)
        if (basePrompt === 'list-commands') {
            const detailedCommands = getDetailedCommands(hasRefreshFlag);
            const output = {
                systemMessage: `Success`,
                hookSpecificOutput: {
                    hookEventName: "UserPromptSubmit",
                    additionalContext: `[LIST-COMMANDS: Display the following detailed command list to the user in a clean formatted way.]\n\n${detailedCommands}`
                }
            };
            console.log(JSON.stringify(output));
            process.exit(0);
        }

        // Handle 'work' commands - broad trigger with opportunistic extraction
        // Supports: work #123, work on #123, work issue 45, work all in Ready, etc.
        if (workTrigger) {
            // Try to extract issue number using flexible patterns:
            // 1. Explicit #N reference anywhere: "work on #123", "work #123 now"
            // 2. Number after work [on] [issue]: "work 123", "work on 45", "work issue 789"
            const explicitIssue = prompt.match(/#(\d+)/);
            const implicitIssue = prompt.match(/^work\s+(?:on\s+)?(?:issue\s+)?(\d+)/i);
            const issueNumber = explicitIssue ? explicitIssue[1] : (implicitIssue ? implicitIssue[1] : null);

            // Try to extract batch status: "work all in Ready", "work the issues in backlog"
            const statusMatch = prompt.match(/\bin\s+(\w+[-\w]*)/i);
            const statusFilter = statusMatch ? statusMatch[1] : null;

            // If no issue number and no status, exit silently (let Claude handle naturally)
            if (!issueNumber && !statusFilter) {
                process.exit(0);
            }

            // Single issue work (has issue number)
            if (issueNumber) {

            try {
                // Query issue's Branch field via gh pmu view
                const result = execSync(
                    `gh pmu view ${issueNumber} --json`,
                    { encoding: 'utf-8', timeout: 10000 }
                );

                const issueData = JSON.parse(result);
                const branch = issueData.fieldValues?.Branch;
                if (!branch || branch === '' || branch === 'null') {
                    // No branch assigned - block with actionable message
                    const output = {
                        continue: false,
                        hookSpecificOutput: {
                            permissionDecision: 'deny',
                            message: `Issue #${issueNumber} has no branch assignment.\n\nUse: /assign-branch #${issueNumber} release/vX.Y.Z\n\nOr use: gh pmu move ${issueNumber} --branch "release/vX.Y.Z"`
                        }
                    };
                    console.log(JSON.stringify(output));
                    process.exit(0);
                }

                // Branch assigned - route to /work command
                let contextMessage = `[INVOKE: /work ${issueNumber}]\n`;

                // PRD Tracker Auto-Move (#1193)
                // Check issue body for **PRD Tracker:** #NNN and move to in_progress if needed
                const body = issueData.body || '';
                const prdTrackerMatch = body.match(/\*\*PRD Tracker:\*\*\s*#(\d+)/);
                if (prdTrackerMatch) {
                    const prdNumber = prdTrackerMatch[1];
                    try {
                        const prdResult = execSync(
                            `gh pmu view ${prdNumber} --json=status`,
                            { encoding: 'utf-8', timeout: 10000 }
                        );
                        const prdData = JSON.parse(prdResult);
                        const prdStatus = (prdData.status || '').toLowerCase();
                        if (prdStatus === 'backlog' || prdStatus === 'ready') {
                            execSync(
                                `gh pmu move ${prdNumber} --status in_progress`,
                                { encoding: 'utf-8', timeout: 10000 }
                            );
                            contextMessage += `\nPRD tracker #${prdNumber} moved to In Progress\n`;
                        }
                    } catch (_prdError) {
                        // Fail silently - PRD tracker query is non-blocking
                    }
                }

                // Auto-todo: Extract acceptance criteria or sub-issues for todo list
                const labels = issueData.labels || [];
                const isEpic = labels.some(l => l === 'epic' || l.name === 'epic');

                if (isEpic) {
                    // Epic: get sub-issues for todo list
                    try {
                        const subResult = execSync(
                            `gh pmu sub list ${issueNumber} --json`,
                            { encoding: 'utf-8', timeout: 10000 }
                        );
                        const subIssues = JSON.parse(subResult);
                        if (subIssues && subIssues.length > 0) {
                            contextMessage += `\n[AUTO-TODO: EPIC]\n`;
                            contextMessage += `Create a todo list from these sub-issues:\n`;
                            subIssues.forEach(sub => {
                                contextMessage += `- #${sub.number}: ${sub.title}\n`;
                            });
                        }
                    } catch (_subError) {
                        // Fail silently - sub-issue query is optional
                    }
                } else {
                    // Story/Bug: extract acceptance criteria checkboxes
                    const checkboxes = body.match(/- \[ \] .+/g) || [];
                    if (checkboxes.length > 0) {
                        contextMessage += `\n[AUTO-TODO: ACCEPTANCE CRITERIA]\n`;
                        contextMessage += `Create a todo list from these acceptance criteria:\n`;
                        checkboxes.forEach(cb => {
                            // Clean up the checkbox format for todo
                            const item = cb.replace(/^- \[ \] /, '').trim();
                            contextMessage += `- ${item}\n`;
                        });
                    }
                }

                const output = {
                    systemMessage: `Success`,
                    hookSpecificOutput: {
                        hookEventName: 'UserPromptSubmit',
                        additionalContext: contextMessage
                    }
                };
                console.log(JSON.stringify(output));
                process.exit(0);

            } catch (_error) {
                // Error checking - route to /work and let command handle (fail-open)
                const output = {
                    systemMessage: `Success`,
                    hookSpecificOutput: {
                        hookEventName: 'UserPromptSubmit',
                        additionalContext: `[INVOKE: /work ${issueNumber}]`
                    }
                };
                console.log(JSON.stringify(output));
                process.exit(0);
            }
            }

            // Batch work (has status but no specific issue number)
            // Route to /work command with the full "all in <status>" argument
            if (statusFilter) {
                let contextMessage = `[INVOKE: /work all in ${statusFilter}]\n`;

                try {
                // Opportunistically query issues for AUTO-TODO
                const result = execSync(
                    `gh pmu list --status ${statusFilter} --json`,
                    { encoding: 'utf-8', timeout: 15000 }
                );

                const data = JSON.parse(result);
                const issues = data.items || [];
                if (issues.length > 0) {
                    contextMessage += `\n[AUTO-TODO: BATCH ISSUES]\n`;
                    contextMessage += `Create a todo list with these issues:\n`;
                    issues.forEach(issue => {
                        contextMessage += `- #${issue.number}: ${issue.title}\n`;
                    });
                }

                } catch (_error) {
                    // Fail silently - /work command will query issues itself
                }

                const output = {
                    systemMessage: `Success`,
                    hookSpecificOutput: {
                        hookEventName: 'UserPromptSubmit',
                        additionalContext: contextMessage
                    }
                };
                console.log(JSON.stringify(output));
                process.exit(0);
            }
        }

        // Handle 'done' commands - contextual trigger (#1200, #1218)
        // Only triggers when issues are in in_review status
        // /done only handles in_review→done (not in_progress→in_review, which is /work's job)
        if (doneTrigger) {
            // Extract optional issue numbers: "done #42 #43" or "done 42"
            const issueNumbers = prompt.match(/#?(\d+)/g);
            const args = issueNumbers
                ? issueNumbers.map(n => n.replace('#', '')).slice(0).join(' ')
                : '';

            // If explicit issue numbers given, route directly (no context check needed)
            if (args) {
                const output = {
                    systemMessage: 'Success',
                    hookSpecificOutput: {
                        hookEventName: 'UserPromptSubmit',
                        additionalContext: `[INVOKE: /done ${args}]`
                    }
                };
                console.log(JSON.stringify(output));
                process.exit(0);
            }

            // No arguments - contextual detection: check for in_review issues only
            try {
                const inReviewResult = execSync(
                    'gh pmu list --status in_review',
                    { encoding: 'utf-8', timeout: 10000 }
                );

                const hasReviewIssues =
                    inReviewResult && inReviewResult.trim().length > 0 && !inReviewResult.includes('no items');

                if (hasReviewIssues) {
                    const output = {
                        systemMessage: 'Success',
                        hookSpecificOutput: {
                            hookEventName: 'UserPromptSubmit',
                            additionalContext: '[INVOKE: /done]'
                        }
                    };
                    console.log(JSON.stringify(output));
                    process.exit(0);
                }
                // No in_review issues - fall through (let Claude handle "done" naturally)
            } catch (_error) {
                // gh pmu not available or query failed - fall through silently
            }
            process.exit(0);
        }

        // Handle review commands - route to /review-issue (#1210)
        if (reviewTrigger) {
            // Extract all issue numbers from the prompt
            const issueMatches = prompt.match(/#(\d+)/g) || prompt.match(/\b(\d{2,})\b/g) || [];
            const issueArgs = issueMatches.map(n => n.replace('#', '')).join(' ');
            const output = {
                systemMessage: 'Success',
                hookSpecificOutput: {
                    hookEventName: 'UserPromptSubmit',
                    additionalContext: `[INVOKE: /review-issue ${issueArgs}]`
                }
            };
            console.log(JSON.stringify(output));
            process.exit(0);
        }

        // Handle workflow triggers - route to specific commands
        if (triggerMatch) {
            const triggerType = triggerMatch[1].toLowerCase();
            const title = prompt.slice(triggerMatch[0].length).trim();

            // Map trigger prefixes to slash commands
            const commandMap = {
                'bug': '/bug',
                'enhancement': '/enhancement',
                'proposal': '/proposal',
                'idea': '/proposal'  // alias
            };

            const command = commandMap[triggerType];
            const output = {
                systemMessage: `Success`,
                hookSpecificOutput: {
                    hookEventName: "UserPromptSubmit",
                    additionalContext: `[INVOKE: ${command} ${title}]`
                }
            };
            console.log(JSON.stringify(output));
        }

        // Handle analysis requests - inject STOP reminder
        // Prevents "evaluate #123" from drifting into implementation
        if (analysisMatch) {
            const output = {
                systemMessage: `Success`,
                hookSpecificOutput: {
                    hookEventName: "UserPromptSubmit",
                    additionalContext: "[ANALYSIS MODE: Report findings and STOP. Do NOT implement, fix, or run commands with side effects until explicit 'work' instruction.]"
                }
            };
            console.log(JSON.stringify(output));
        }

        process.exit(0);
    } catch (_e) {
        // Intentionally ignored
        process.exit(0);
    }
});

/**
 * Detect active IDPF framework (single source of truth)
 * @returns {string|null} Framework name or null
 */
function detectFramework() {
    const cwd = process.cwd();

    // Check framework-config.json first (user projects - most specific)
    try {
        const configPath = path.join(cwd, 'framework-config.json');
        if (fs.existsSync(configPath)) {
            const config = JSON.parse(fs.readFileSync(configPath, 'utf8'));
            const framework = config.processFramework || config.projectType?.processFramework || config.framework;
            if (framework) return normalizeFramework(framework);
        }
    } catch (_e) {
        // Intentionally ignored
    }

    // Check for IDPF directories (framework dev or direct usage)
    const frameworks = ['IDPF-Agile', 'IDPF-Vibe'];
    for (const fw of frameworks) {
        if (fs.existsSync(path.join(cwd, fw))) {
            return fw;
        }
    }

    return null;
}

/**
 * Normalize framework name to standard format
 */
function normalizeFramework(name) {
    const lower = name.toLowerCase();
    if (lower === 'agile' || lower === 'idpf-agile') return 'IDPF-Agile';
    if (lower.startsWith('vibe') || lower === 'idpf-vibe') return 'IDPF-Vibe';
    return name;
}

/**
 * Try to load from cache, regenerate if stale or missing
 * @param {string} key - Cache key to retrieve
 * @param {boolean} forceRefresh - If true, skip cache and force regeneration
 */
function getFromCache(key, forceRefresh = false) {
    // Skip cache entirely if refresh requested
    if (forceRefresh) {
        return null;
    }

    try {
        if (fs.existsSync(CACHE_FILE)) {
            const cache = JSON.parse(fs.readFileSync(CACHE_FILE, 'utf8'));

            // Check if cache is from today (daily expiration)
            if (cache.timestamp) {
                const cacheDate = new Date(cache.timestamp).toDateString();
                const today = new Date().toDateString();
                if (cacheDate !== today) {
                    return null; // Cache expired, force regeneration
                }
            }

            // Trust cached framework when cache is fresh (no detectFramework() call)
            // Framework changes are rare within a session
            if (cache[key]) {
                return cache[key];
            }
        }
    } catch (_e) {
        // Intentionally ignored
    }
    return null;
}

/**
 * Save to cache
 */
function saveToCache(key, value) {
    try {
        let cache = {};
        if (fs.existsSync(CACHE_FILE)) {
            cache = JSON.parse(fs.readFileSync(CACHE_FILE, 'utf8'));
        }
        cache.framework = detectFramework();
        cache[key] = value;
        cache.timestamp = Date.now();
        fs.writeFileSync(CACHE_FILE, JSON.stringify(cache, null, 2));
    } catch (_e) {
        // Intentionally ignored
    }
}

/**
 * Generate help text for all available commands
 * @param {boolean} forceRefresh - If true, bypass cache and regenerate
 */
function generateCommandsHelp(forceRefresh = false) {
    // Try cache first (unless refresh requested)
    const cached = getFromCache('commandsHelp', forceRefresh);
    if (cached) return cached;

    let help = `📋 **Available Commands**

**Workflow Triggers** (prefix your message):
- \`bug:\` - Report a bug → creates issue, wait for 'work' to implement fix
- \`enhancement:\` - Request enhancement → creates issue, wait for 'work' to implement
- \`idea:\` - Alias for proposal: → creates proposal document + issue
- \`proposal:\` - Formal proposal → creates proposal document + issue

**Issue Management**:
- \`work #N\` or \`work <issue>\` - Start working on issue (moves to In Progress)
- \`done\` - Complete current issue (moves to Done, closes issue)
`;

    // Get slash commands
    const slashCommands = getSlashCommands();
    if (slashCommands.length > 0) {
        help += `\n**Slash Commands**:\n`;
        for (const cmd of slashCommands) {
            help += `- \`/${cmd.name}\` - ${cmd.description}\n`;
        }
    }

    // Show active framework info
    const framework = detectFramework();
    if (framework) {
        help += `\n**Active Framework:** ${framework}\n`;
        help += `Type \`list-commands\` for the detailed command reference.\n`;
    }

    // Cache the result
    saveToCache('commandsHelp', help);

    return help;
}

/**
 * Get detailed commands (used by list-commands)
 * @param {boolean} forceRefresh - If true, bypass cache and regenerate
 */
function getDetailedCommands(forceRefresh = false) {
    // Try cache first (unless refresh requested)
    const cached = getFromCache('detailedCommands', forceRefresh);
    if (cached) return cached;

    const framework = detectFramework();
    let result = '';

    if (framework) {
        result = getFrameworkDetailedCommands(framework);
    } else {
        // No framework detected — show workflow triggers + dynamic slash commands
        result = `## Available Triggers & Slash Commands

### Workflow Triggers (prefix your message)

| Trigger | Description |
|---------|-------------|
| \`bug: <title>\` | Create a bug issue |
| \`enhancement: <title>\` | Create an enhancement issue |
| \`proposal: <title>\` | Create a proposal document + tracking issue |
| \`idea: <title>\` | Alias for proposal: |
| \`work #N\` | Start working on an issue |
| \`work all in <status>\` | Batch work on issues by status |
| \`done [#N]\` | Complete issue (in_review → done) |
| \`review #N\` | Route to /review-issue |
| \`commands\` | Show available triggers and slash commands |
| \`list-commands\` | Show detailed command list |

### Slash Commands
`;
        const slashCommands = getSlashCommands();
        if (slashCommands.length > 0) {
            result += `\n| Command | Description |\n|---------|-------------|\n`;
            for (const cmd of slashCommands) {
                result += `| \`/${cmd.name}\` | ${cmd.description} |\n`;
            }
        } else {
            result += `\nNo slash commands found in .claude/commands/\n`;
        }

        result += '\n' + getFrameworkSelectionHelp();
    }

    // Cache the result
    saveToCache('detailedCommands', result);

    return result;
}

/**
 * Read slash commands from .claude/commands/ directory
 */
function getSlashCommands() {
    const commands = [];
    const commandsDir = path.join(process.cwd(), '.claude', 'commands');

    try {
        if (!fs.existsSync(commandsDir)) return commands;

        const files = fs.readdirSync(commandsDir).filter(f => f.endsWith('.md'));
        for (const file of files) {
            const filePath = path.join(commandsDir, file);
            // Read only first 500 bytes for frontmatter (optimization)
            const fd = fs.openSync(filePath, 'r');
            const buffer = Buffer.alloc(500);
            fs.readSync(fd, buffer, 0, 500, 0);
            fs.closeSync(fd);
            const content = buffer.toString('utf8');

            const frontmatterMatch = content.match(/^---\s*\n([\s\S]*?)\n---/);
            if (frontmatterMatch) {
                const descMatch = frontmatterMatch[1].match(/description:\s*(.+)/);
                if (descMatch) {
                    commands.push({
                        name: file.replace('.md', ''),
                        description: descMatch[1].trim()
                    });
                }
            }
        }
    } catch (_e) {
        // Intentionally ignored
    }

    return commands;
}

/**
 * Get framework-specific detailed commands
 */
function getFrameworkDetailedCommands(framework) {
    switch (framework) {
        case 'IDPF-Agile': return getAgileDetailedCommands();
        default: return getFrameworkSelectionHelp();
    }
}

function getAgileDetailedCommands() {
    let result = `## IDPF-Agile — Workflow Triggers & Slash Commands

### Workflow Triggers (prefix your message)

| Trigger | Description |
|---------|-------------|
| \`bug: <title>\` | Create a bug issue |
| \`enhancement: <title>\` | Create an enhancement issue |
| \`proposal: <title>\` | Create a proposal document + tracking issue |
| \`idea: <title>\` | Alias for proposal: |
| \`work #N\` | Start working on an issue (validates branch, extracts auto-TODO) |
| \`work all in <status>\` | Batch work on issues by status |
| \`done [#N]\` | Complete issue (in_review → done) |
| \`review #N\` | Route to /review-issue |
| \`commands\` | Show available triggers and slash commands |
| \`list-commands\` | Show detailed command list |

### Slash Commands
`;
    const slashCommands = getSlashCommands();
    if (slashCommands.length > 0) {
        result += `\n| Command | Description |\n|---------|-------------|\n`;
        for (const cmd of slashCommands) {
            result += `| \`/${cmd.name}\` | ${cmd.description} |\n`;
        }
    } else {
        result += `\nNo slash commands found in .claude/commands/\n`;
    }

    result += `
### TDD Methodology

IDPF-Agile uses TDD RED-GREEN-REFACTOR cycles. Workflow checkpoint is story completion (In Review → Done).`;

    return result;
}

function getFrameworkSelectionHelp() {
    return `## No Active Framework Detected

To see framework-specific commands, either:

1. **Set up a project** with \`framework-config.json\`:
   \`\`\`json
   { "framework": "IDPF-Agile" }
   \`\`\`

2. **Available frameworks:**
   - \`IDPF-Agile\` - Story-based development with TDD
   - \`IDPF-Vibe\` - Exploratory development

3. **Quick start:** Type \`commands\` to see workflow triggers and slash commands.`;
}
