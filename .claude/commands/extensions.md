---
version: "v0.23.3"
description: Discover, view, and manage extension points in release commands
argument-hint: <subcommand> [options]
---
# /extensions
Unified management of extension points across release commands.
## Subcommands
| Subcommand | Description |
|------------|-------------|
| `list` | Show all extension points |
| `list --command X` | Show for specific command |
| `view X:Y` | Show content of extension Y in command X |
| `edit X:Y` | Edit extension Y in command X |
| `validate` | Check all extension blocks are properly formatted |
| `matrix` | Show cross-command comparison |
---
## Target Commands
| Command | File |
|---------|------|
| `/create-branch` | `Templates/commands/create-branch.md` |
| `/prepare-release` | `Templates/commands/prepare-release.md` |
| `/merge-branch` | `Templates/commands/merge-branch.md` |
| `/destroy-branch` | `Templates/commands/destroy-branch.md` |
| `/prepare-beta` | `Templates/commands/prepare-beta.md` |
---
## Subcommand: list
**Usage:** `/extensions list [--command <name>]`
### Step 1: Parse Extension Points
Extract using marker pattern:
```
<!-- USER-EXTENSION-START: {name} -->
...content...
<!-- USER-EXTENSION-END: {name} -->
```
### Step 2: Detect Content
**Empty:** No non-whitespace between markers
**Has content:** Non-whitespace exists
### Step 3: Format Output
```
Extension Points (Release Commands)
/prepare-release
  ├── post-analysis         : Commit analysis [HAS CONTENT]
  ├── pre-validation        : Setup, fixtures
  └── post-validation       : Coverage gates [HAS CONTENT]
Total: N extension points (X with content, Y empty)
```
### Step 4: Filter by Command
If `--command X` passed, show only that command's extension points.
---
## Subcommand: view
**Usage:** `/extensions view <command>:<point>`
1. Parse command:point from argument
2. Find `USER-EXTENSION-START: {point}` marker
3. Extract content between START/END
4. Display with location and line numbers
---
## Subcommand: edit
**Usage:** `/extensions edit <command>:<point>`
1. Locate extension block
2. Present current content
3. **ASK USER** for new content
4. Replace content between markers
5. Confirm change
---
## Subcommand: validate
**Usage:** `/extensions validate`
Check each block for:
| Check | Description |
|-------|-------------|
| Matching pairs | Every START has END |
| Valid names | Match `[a-z][a-z0-9-]*` |
| No nesting | Blocks not nested |
| Documented | Appears in extension table |
Report: `✓ valid` or `✗ issue description`
---
## Subcommand: matrix
**Usage:** `/extensions matrix`
```
Point                | create | prepare | beta |
---------------------|--------|---------|------|
post-analysis        |   -    |    ●    |  ●   |
pre-validation       |   -    |    ○    |  ○   |
Legend: ● = has content, ○ = empty, - = not applicable
```
---
## Naming Convention
| Pattern | Purpose |
|---------|---------|
| `pre-*` | Before workflow phase |
| `post-*` | After workflow phase |
| `checklist` | Single verification |
| `checklist-before-*` | Pre-action verification |
| `checklist-after-*` | Post-action verification |
---
**End of Extensions Command**
