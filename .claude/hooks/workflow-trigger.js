#!/usr/bin/env node
/**
 * workflow-trigger.js
 *
 * UserPromptSubmit hook that:
 * 1. Detects workflow trigger prefixes and injects reminders
 * 2. Responds to 'commands' with available triggers and slash commands
 *
 * Trigger prefixes: bug:, enhancement:, finding:, idea:, proposal:, prd:
 */

const fs = require('fs');
const path = require('path');

let input = '';

process.stdin.on('data', chunk => input += chunk);
process.stdin.on('end', () => {
    try {
        const data = JSON.parse(input);
        const prompt = (data.prompt || '').trim();

        // Check for 'commands' request
        if (prompt.toLowerCase() === 'commands') {
            const helpText = generateCommandsHelp();
            const output = {
                systemMessage: `⚡ Commands`,
                hookSpecificOutput: {
                    hookEventName: "UserPromptSubmit",
                    additionalContext: `[COMMANDS HELP: Display the following commands to the user in a clean formatted way.]\n\n${helpText}`
                }
            };
            console.log(JSON.stringify(output));
            process.exit(0);
        }

        // Check for 'List-Commands' request (full detailed list)
        if (prompt.toLowerCase() === 'list-commands' || prompt.toLowerCase() === 'list-cmds') {
            const detailedCommands = generateDetailedCommands();
            const output = {
                systemMessage: `⚡ List-Commands`,
                hookSpecificOutput: {
                    hookEventName: "UserPromptSubmit",
                    additionalContext: `[LIST-COMMANDS: Display the following detailed command list to the user in a clean formatted way.]\n\n${detailedCommands}`
                }
            };
            console.log(JSON.stringify(output));
            process.exit(0);
        }

        // Check for workflow triggers
        const match = prompt.match(/^(bug|enhancement|finding|idea|proposal|prd):/i);
        if (match) {
            const triggerType = match[1].toLowerCase();

            // Special handling for prd: trigger
            if (triggerType === 'prd') {
                const output = {
                    systemMessage: `⚡ PRD conversion trigger detected`,
                    hookSpecificOutput: {
                        hookEventName: "UserPromptSubmit",
                        additionalContext: "[PRD TRIGGER: Invoke Proposal-to-PRD workflow (Section 8 of GitHub-Workflow.md). Identify proposal from name or issue number, then run IDPF-PRD phases.]"
                    }
                };
                console.log(JSON.stringify(output));
            } else {
                const output = {
                    systemMessage: `⚡ Workflow trigger detected: "${triggerType}"`,
                    hookSpecificOutput: {
                        hookEventName: "UserPromptSubmit",
                        additionalContext: "[WORKFLOW TRIGGER: Create GitHub issue first. Wait for 'work' instruction before implementing.]"
                    }
                };
                console.log(JSON.stringify(output));
            }
        }

        process.exit(0);
    } catch (e) {
        // Parse error - allow prompt to proceed
        process.exit(0);
    }
});

/**
 * Generate help text for all available commands
 */
function generateCommandsHelp() {
    let help = `📋 **Available Commands**

**Workflow Triggers** (prefix your message):
- \`bug:\` - Report a bug → creates issue, wait for 'work' to implement fix
- \`enhancement:\` - Request enhancement → creates issue, wait for 'work' to implement
- \`finding:\` - Document a finding → creates issue for discovered issues
- \`idea:\` - Quick idea → creates lightweight proposal + issue
- \`proposal:\` - Formal proposal → creates proposal document + issue
- \`prd: [name]\` - Convert proposal to PRD → invokes IDPF-PRD workflow

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

    // Detect and show IDPF framework commands
    const frameworkCommands = getFrameworkCommands();
    if (frameworkCommands) {
        help += frameworkCommands;
    }

    return help;
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
            const content = fs.readFileSync(filePath, 'utf8');

            // Extract description from frontmatter
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
    } catch (e) {
        // Silently fail if can't read commands
    }

    return commands;
}

/**
 * Detect active IDPF framework and return relevant commands.
 * Returns full detailed commands (same as List-Commands) for consistency.
 */
function getFrameworkCommands() {
    const cwd = process.cwd();

    // Check for framework-config.json (user projects)
    const configPath = path.join(cwd, 'framework-config.json');
    let activeFramework = null;

    try {
        if (fs.existsSync(configPath)) {
            const config = JSON.parse(fs.readFileSync(configPath, 'utf8'));
            // Support both old (config.framework) and new (config.projectType.processFramework) formats
            activeFramework = config.projectType?.processFramework || config.framework;
        }
    } catch (e) {}

    // Check for IDPF directories (framework dev or direct usage)
    if (!activeFramework) {
        if (fs.existsSync(path.join(cwd, 'IDPF-Agile'))) {
            activeFramework = 'IDPF-Agile';
        } else if (fs.existsSync(path.join(cwd, 'IDPF-Structured'))) {
            activeFramework = 'IDPF-Structured';
        } else if (fs.existsSync(path.join(cwd, 'IDPF-Vibe'))) {
            activeFramework = 'IDPF-Vibe';
        }
    }

    if (!activeFramework) return null;

    // Return framework-specific detailed commands (same as List-Commands)
    if (activeFramework === 'IDPF-Agile' || activeFramework === 'agile') {
        return '\n' + getAgileDetailedCommands();
    }

    if (activeFramework === 'IDPF-Structured' || activeFramework === 'structured') {
        return '\n' + getStructuredDetailedCommands();
    }

    if (activeFramework === 'IDPF-Vibe' || (activeFramework && activeFramework.startsWith('vibe'))) {
        return '\n' + getVibeDetailedCommands();
    }

    if (activeFramework === 'IDPF-PRD' || activeFramework === 'prd') {
        return '\n' + getPRDDetailedCommands();
    }

    if (activeFramework === 'IDPF-LTS' || activeFramework === 'lts') {
        return '\n' + getLTSDetailedCommands();
    }

    return null;
}

/**
 * Generate detailed command list based on active framework
 */
function generateDetailedCommands() {
    const cwd = process.cwd();
    let activeFramework = null;

    // Check for framework-config.json (user projects)
    const configPath = path.join(cwd, 'framework-config.json');
    try {
        if (fs.existsSync(configPath)) {
            const config = JSON.parse(fs.readFileSync(configPath, 'utf8'));
            // Support both old (config.framework) and new (config.projectType.processFramework) formats
            activeFramework = config.projectType?.processFramework || config.framework;
        }
    } catch (e) {}

    // Check for IDPF directories (framework dev or direct usage)
    if (!activeFramework) {
        if (fs.existsSync(path.join(cwd, 'IDPF-Agile'))) {
            activeFramework = 'IDPF-Agile';
        } else if (fs.existsSync(path.join(cwd, 'IDPF-Structured'))) {
            activeFramework = 'IDPF-Structured';
        } else if (fs.existsSync(path.join(cwd, 'IDPF-Vibe'))) {
            activeFramework = 'IDPF-Vibe';
        } else if (fs.existsSync(path.join(cwd, 'IDPF-PRD'))) {
            activeFramework = 'IDPF-PRD';
        } else if (fs.existsSync(path.join(cwd, 'IDPF-LTS'))) {
            activeFramework = 'IDPF-LTS';
        }
    }

    // Return framework-specific detailed commands
    if (activeFramework === 'IDPF-Agile' || activeFramework === 'agile') {
        return getAgileDetailedCommands();
    }

    if (activeFramework === 'IDPF-Structured' || activeFramework === 'structured') {
        return getStructuredDetailedCommands();
    }

    if (activeFramework === 'IDPF-Vibe' || (activeFramework && activeFramework.startsWith('vibe'))) {
        return getVibeDetailedCommands();
    }

    if (activeFramework === 'IDPF-PRD' || activeFramework === 'prd') {
        return getPRDDetailedCommands();
    }

    if (activeFramework === 'IDPF-LTS' || activeFramework === 'lts') {
        return getLTSDetailedCommands();
    }

    // Fallback: show all frameworks available
    return getFrameworkSelectionHelp();
}

function getAgileDetailedCommands() {
    return `## IDPF-Agile Commands - Full List

### Backlog Management

| Command | Description |
|---------|-------------|
| \`Create-Backlog\` | Generate initial product backlog from project vision |
| \`Show-Backlog\` | Display current product backlog |
| \`Add-Story\` | User describes a new story to add to backlog |
| \`Refine-Story [ID]\` | Update/clarify an existing story |
| \`Estimate-Story [ID]\` | Re-estimate story points |
| \`Prioritize-Backlog\` | Re-order stories by priority |
| \`Split-Story [ID]\` | Break a large story into smaller stories |
| \`Archive-Story [ID]\` | Move story to icebox |

---

### GitHub Issue Commands

| Command | Description |
|---------|-------------|
| \`Create-Issues\` | Create GitHub issues from PRD/backlog (auto-detects framework) |
| \`Create-Issues-Agile\` | Explicit Agile issue creation |

---

### Sprint Commands

| Command | Description |
|---------|-------------|
| \`Plan-Sprint\` | Select stories for next sprint |
| \`Show-Sprint\` | Display current sprint backlog |
| \`Start-Story [ID]\` | Begin development on a story |
| \`Story-Status\` | Check progress on current story |
| \`Story-Complete [ID]\` | Mark story as done |
| \`Sprint-Progress\` | Show sprint burndown/progress |
| \`Sprint-Review\` | Review completed sprint |
| \`Sprint-Retro\` | Conduct retrospective |
| \`End-Sprint\` | Close current sprint (runs Review + Retro) |

---

### Development Commands

| Command | Description |
|---------|-------------|
| \`Done-Next-Step\` | Current TDD iteration successful, proceed |
| \`Rollback-Previous-Step\` | Undo last iteration |
| \`Run-Tests\` | Execute full test suite |
| \`Show-Coverage\` | Display test coverage report |
| \`Refactor-Now\` | Dedicated refactoring session |

**TDD Skills:** \`tdd-red-phase\`, \`tdd-green-phase\`, \`tdd-refactor-phase\`, \`tdd-failure-recovery\`, \`test-writing-patterns\`

---

### Project Commands

| Command | Description |
|---------|-------------|
| \`Project-Status\` | Overall project health dashboard |
| \`Velocity-Report\` | Show velocity trends across sprints |
| \`Push-Changes\` | Commit and push current work |
| \`Create-Release\` | Tag a release version |
| \`Project-Complete\` | Finalize project and create final PR |

---

### Special Scenario Commands

| Command | Description |
|---------|-------------|
| \`Story-Blocked [ID] [reason]\` | Mark story as blocked |
| \`Story-Growing [ID]\` | Flag story scope creep |
| \`Emergency-Bug [description]\` | Create unplanned bug fix story |
| \`Pivot [new direction]\` | Change project direction |

---

### Utility Commands

| Command | Description |
|---------|-------------|
| \`List-Commands\` | Show all available commands with descriptions |
| \`List-Cmds\` | Show all commands without descriptions |
| \`Help [command]\` | Get detailed help for specific command |
| \`Review-Last\` | Review ASSISTANT's most recent reply for accuracy |`;
}

function getStructuredDetailedCommands() {
    return `## IDPF-Structured Commands - Full List

### TDD Workflow Commands

| Command | Description |
|---------|-------------|
| \`Done-Next-Step\` | Current TDD iteration successful, proceed to next |
| \`Rollback-Previous-Step\` | Undo last iteration |
| \`Run-Tests\` | Execute full test suite |
| \`Show-Coverage\` | Display test coverage report |
| \`Refactor-Now\` | Dedicated refactoring session |

---

### TDD Phase Skills

| Skill | Description |
|-------|-------------|
| \`tdd-red-phase\` | RED phase - write failing test first |
| \`tdd-green-phase\` | GREEN phase - minimal code to pass |
| \`tdd-refactor-phase\` | REFACTOR phase - improve without changing behavior |
| \`tdd-failure-recovery\` | When tests behave unexpectedly |
| \`test-writing-patterns\` | Test structure and assertion guidance |

---

### Requirement Commands

| Command | Description |
|---------|-------------|
| \`Show-Requirements\` | Display all requirements from PRD |
| \`Start-Requirement [REQ-XXX]\` | Begin work on specific requirement |
| \`Requirement-Complete [REQ-XXX]\` | Mark requirement as implemented |
| \`Show-Progress\` | Display implementation progress |

---

### GitHub Issue Commands

| Command | Description |
|---------|-------------|
| \`Create-Issues\` | Create GitHub issues from PRD (auto-detects framework) |
| \`Create-Issues-Structured\` | Explicit Structured issue creation |

---

### Project Commands

| Command | Description |
|---------|-------------|
| \`Project-Status\` | Overall project health dashboard |
| \`Push-Changes\` | Commit and push current work |
| \`Create-Release\` | Tag a release version |
| \`Project-Complete\` | Finalize project and create final PR |

---

### Utility Commands

| Command | Description |
|---------|-------------|
| \`List-Commands\` | Show all available commands with descriptions |
| \`Help [command]\` | Get detailed help for specific command |
| \`Review-Last\` | Review ASSISTANT's most recent reply for accuracy |`;
}

function getVibeDetailedCommands() {
    return `## IDPF-Vibe Commands - Full List

### Development Flow

IDPF-Vibe uses conversational development. No strict commands required.

| Action | Description |
|--------|-------------|
| Describe what you want | Assistant builds iteratively |
| "Try this..." | Experiment with approaches |
| "Change it to..." | Modify current implementation |
| "That works, next..." | Move to next feature |

---

### Transition Commands

| Command | Description |
|---------|-------------|
| \`Formalize\` | Ready to transition to Structured or Agile |
| \`Extract-PRD\` | Generate PRD from current implementation |
| \`Add-Tests\` | Begin adding test coverage |

---

### When to Transition

- Code is working and stable
- Ready to add comprehensive tests
- Need formal requirements documentation
- Moving to team development

---

### Utility Commands

| Command | Description |
|---------|-------------|
| \`List-Commands\` | Show all available commands |
| \`Help\` | Get guidance on Vibe workflow |`;
}

function getPRDDetailedCommands() {
    return `## IDPF-PRD Commands - Full List

### Discovery Phase

| Command | Description |
|---------|-------------|
| \`Start-PRD\` | Begin PRD creation process |
| \`Define-Stakeholders\` | Identify project stakeholders |
| \`Define-Domain\` | Specify project domain and context |

---

### Elicitation Phase

| Command | Description |
|---------|-------------|
| \`Add-Requirement\` | Add a new functional requirement |
| \`Add-NFR\` | Add non-functional requirement |
| \`Add-Constraint\` | Add project constraint |
| \`Add-Risk\` | Document project risk |

---

### Specification Phase

| Command | Description |
|---------|-------------|
| \`Detail-Requirement [REQ-XXX]\` | Add details to requirement |
| \`Add-Acceptance-Criteria [REQ-XXX]\` | Define acceptance criteria |
| \`Prioritize-Requirements\` | Order requirements by priority |

---

### Generation Phase

| Command | Description |
|---------|-------------|
| \`Generate-PRD\` | Generate final PRD document |
| \`Select-Template\` | Choose PRD template (Comprehensive/Moderate/Lightweight) |
| \`Export-PRD\` | Export to specified format |

---

### Utility Commands

| Command | Description |
|---------|-------------|
| \`List-Commands\` | Show all available commands |
| \`Show-PRD-Progress\` | Display current PRD status |
| \`Review-Last\` | Review ASSISTANT's most recent reply |`;
}

function getLTSDetailedCommands() {
    return `## IDPF-LTS Commands - Full List

### Maintenance Commands

| Command | Description |
|---------|-------------|
| \`Show-Status\` | Display system health and metrics |
| \`Review-Logs\` | Analyze recent log entries |
| \`Check-Dependencies\` | Review dependency versions and updates |

---

### Bug Fix Workflow

| Command | Description |
|---------|-------------|
| \`Report-Bug\` | Document a new bug |
| \`Investigate-Bug [ID]\` | Begin bug investigation |
| \`Fix-Bug [ID]\` | Implement bug fix |
| \`Verify-Fix [ID]\` | Confirm fix resolves issue |

---

### Security Commands

| Command | Description |
|---------|-------------|
| \`Security-Audit\` | Run security assessment |
| \`Apply-Patch [CVE]\` | Apply security patch |
| \`Review-Vulnerabilities\` | Check for known vulnerabilities |

---

### Release Commands

| Command | Description |
|---------|-------------|
| \`Prepare-Patch\` | Prepare patch release |
| \`Create-Hotfix\` | Create emergency hotfix |
| \`Deploy-Release\` | Deploy to production |

---

### Utility Commands

| Command | Description |
|---------|-------------|
| \`List-Commands\` | Show all available commands |
| \`Help [command]\` | Get detailed help for specific command |`;
}

function getFrameworkSelectionHelp() {
    return `## No Active Framework Detected

To see framework-specific commands, either:

1. **Set up a project** with \`framework-config.json\`:
   \`\`\`json
   { "framework": "IDPF-Agile" }
   \`\`\`

2. **Available frameworks:**
   - \`IDPF-Agile\` - Sprint-based development with user stories
   - \`IDPF-Structured\` - TDD with fixed requirements
   - \`IDPF-Vibe\` - Exploratory development
   - \`IDPF-PRD\` - Requirements engineering
   - \`IDPF-LTS\` - Long-term support/maintenance

3. **Quick start:** Type \`commands\` to see workflow triggers and slash commands.`;
}
