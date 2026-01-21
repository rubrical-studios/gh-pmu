---
version: "v0.29.2"
allowed-tools: Bash, AskUserQuestion
description: Assign issues to a branch (project)
---
<!-- MANAGED -->
Assign issues to a branch.
```bash
node .claude/scripts/shared/assign-branch.js "$ARGUMENTS"
```
## Handling "NO_RELEASE_FOUND"
If script outputs `NO_RELEASE_FOUND`, no open branches exist. Script also outputs:
1. **CONTEXT:** - Info about last version, labels, input
2. **SUGGESTIONS:** - Formatted as `number|branch|description`
When you see this:
1. Parse SUGGESTIONS for branch options
2. Use `AskUserQuestion` to let user select (recommended first, descriptions included, Other for custom)
3. Create branch: `gh pmu branch start --name "<selected>"`
4. Re-run original assign-branch command
## Normal Output
If branches exist, report result directly.
