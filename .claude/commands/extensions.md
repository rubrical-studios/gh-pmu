---
version: "v0.53.1"
description: Discover, view, and manage extension points in release commands
argument-hint: "list|view|edit|validate|matrix|recipes [args]"
---
<!-- MANAGED -->
# /extensions
Unified management of extension points across release commands.
## Subcommands
| Subcommand | Description |
|------------|-------------|
| `list` | Show all extension points |
| `list --command X` | Show extension points for specific command |
| `view X:Y` | Show content of extension point Y in command X |
| `edit X:Y` | Edit extension point Y in command X |
| `validate` | Check all extension blocks are properly formatted |
| `matrix` | Show cross-command extension point comparison |
| `recipes` | Show common patterns for extension points |
| `recipes <category>` | Show recipes for specific category |
---
## Target Commands
Target commands are listed in `.claude/metadata/extension-points.json` (the extension registry).
**Registry path:** `.claude/metadata/extension-points.json`
**Command files:** `.claude/commands/*.md` (for `edit` subcommand)
### Fallback: Registry Missing
If `extension-points.json` does not exist or cannot be parsed:
1. **Warn:** `Extension registry not found. Scanning command files directly (slower).`
2. **Fallback:** Scan `.claude/commands/*.md` for EXTENSIBLE markers and USER-EXTENSION-START/END blocks
---
## Script Delegation
Read-only subcommands (`list`, `view`, `validate`, `matrix`, `recipes`, `help`) are handled by the extensions-cli.js script. Run the script and display its stdout output directly.
**Script path:** `node .claude/scripts/shared/extensions-cli.js`
### Delegated Subcommands
| Subcommand | Script Command |
|------------|---------------|
| `list [--command X]` | `node .claude/scripts/shared/extensions-cli.js list [--command X]` |
| `view X:Y` | `node .claude/scripts/shared/extensions-cli.js view X:Y` |
| `validate` | `node .claude/scripts/shared/extensions-cli.js validate` |
| `matrix` | `node .claude/scripts/shared/extensions-cli.js matrix` |
| `recipes [category]` | `node .claude/scripts/shared/extensions-cli.js recipes [category]` |
| `help` | `node .claude/scripts/shared/extensions-cli.js help` |
### How to Dispatch
1. Parse the user's subcommand and arguments
2. Run the corresponding script command via Bash
3. Display the script's stdout output directly to the user
4. If exit code is non-zero (1 = validation failures, 2 = fatal error), report the error
**Do NOT** re-interpret or reformat the script output — display it as-is.
---
## Subcommand: edit
**Usage:** `/extensions edit <command>:<point>`
**Example:** `/extensions edit prepare-release:post-validation`
The `edit` subcommand is spec-driven (not delegated to the script) because it requires interactive user input and file modification.
### Step 1: Locate Extension Block
Read the live command file (`.claude/commands/{command}.md`). `edit` reads and modifies the command file directly.
Find the `USER-EXTENSION-START:{point}` and `USER-EXTENSION-END:{point}` markers. If not found, report error and **STOP**.
### Step 2: Present Current Content
Show the current content between the START and END markers.
- If empty: `"Extension point '{command}:{point}' is currently empty."`
- If has content: Display the content in a fenced code block
### Step 3: Ask User What They Want Changed (Intent-Based)
**Do NOT ask the user to provide raw replacement text.** Instead, ask what they want:
- If empty: `"What would you like to add to this extension point?"`
- If has content: `"What changes would you like to make?"`
The user describes their intent in natural language (e.g., "add a line that says X", "remove the git status check").
### Step 4: Implement the Edit Directly
The assistant implements the change using the **Edit tool** on the command file. This preserves surrounding formatting exactly.
**Rules:**
- Only modify content between the START and END markers
- Preserve the START and END comment markers exactly
- Preserve all formatting outside the extension block
- For "add" intents: insert the new content between markers
- For "remove" intents: remove the specified content, leave markers (block may become empty)
- For "replace" intents: replace the specified content between markers
### Step 5: Confirm Change
Report the updated extension block content. Show a before/after summary.
**Do NOT rebuild the extension registry.** The registry is a static framework build artifact. The `hasContent` state is computed at runtime by scanning command files.
---
## Extension Point Naming Convention
| Pattern | Purpose |
|---------|---------|
| `pre-*` | Before a workflow phase |
| `post-*` | After a workflow phase |
| `pre-commit` | Generate artifacts before commit |
| `checklist` | Single verification checklist |
| `checklist-before-*` | Pre-action verification items |
| `checklist-after-*` | Post-action verification items |
---
## Help
**Usage:** `/extensions --help` or `/extensions`
When invoked with `--help` or with no subcommand, delegate to the script:
```bash
node .claude/scripts/shared/extensions-cli.js help
```
Display the script's stdout output directly to the user.
---
**End of Extensions Command**
