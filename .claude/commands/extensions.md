---
version: "v0.33.2"
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
| `recipes <category>` | Show recipes for specific category (ci, coverage, etc.) |

---

## Target Commands

The following commands contain extension points:

| Command | Path |
|---------|------|
| `/create-branch` | `.claude/commands/create-branch.md` |
| `/prepare-release` | `.claude/commands/prepare-release.md` |
| `/merge-branch` | `.claude/commands/merge-branch.md` |
| `/destroy-branch` | `.claude/commands/destroy-branch.md` |
| `/prepare-beta` | `.claude/commands/prepare-beta.md` |

**Note:** Always scan `.claude/commands/` - this is where project commands live after installation.

---

## Subcommand: list

**Usage:** `/extensions list [--command <name>]`

### Step 1: Parse Extension Points

For each target command file, extract extension points using the marker pattern:

```
<!-- USER-EXTENSION-START: {name} -->
...content...
<!-- USER-EXTENSION-END: {name} -->
```

### Step 2: Detect Content

An extension point **has content** if there is non-whitespace, non-comment text between the markers.

**Empty block example:**
```markdown
<!-- USER-EXTENSION-START: post-create -->
<!-- USER-EXTENSION-END: post-create -->
```

**Content block example:**
```markdown
<!-- USER-EXTENSION-START: post-validation -->
### Coverage Gate
...
<!-- USER-EXTENSION-END: post-validation -->
```

### Step 3: Format Output

```
Extension Points (Release Commands)

/create-branch
  ├── pre-create            : Custom validation
  ├── post-create           : Notifications, CI triggers
  └── checklist             : Extension-specific verification items

/prepare-release
  ├── post-analysis         : Commit analysis, version recommendation [HAS CONTENT]
  ├── pre-validation        : Setup, fixtures, containers
  ├── post-validation       : Coverage gates, build verification [HAS CONTENT]
  ├── post-prepare          : Documentation updates [HAS CONTENT]
  ├── pre-tag               : Final gate, sign-off
  ├── post-tag              : Release monitoring, asset verification [HAS CONTENT]
  ├── checklist-before-tag  : Pre-tag verification items [HAS CONTENT]
  └── checklist-after-tag   : Post-tag verification items [HAS CONTENT]

...

Total: N extension points (X with content, Y empty)
```

### Step 4: Filter by Command (if --command specified)

If `--command prepare-release` is passed, show only that command's extension points.

---

## Subcommand: view

**Usage:** `/extensions view <command>:<point>`

**Example:** `/extensions view prepare-release:post-validation`

### Step 1: Parse Arguments

Extract command name and extension point name from the colon-separated argument.

### Step 2: Locate Extension Block

Find the `USER-EXTENSION-START: {point}` marker in the target command file.

### Step 3: Extract Content

Return the content between START and END markers.

### Step 4: Format Output

```
Extension: prepare-release:post-validation

Location: .claude/commands/prepare-release.md (lines 113-127)

Content:
────────────────────────────────────────────────────────
### Coverage Gate

**If `--skip-coverage` was passed, skip this section.**

```bash
node .claude/scripts/prepare-release/coverage.js
```

**If `success` is false, STOP and report the error.**

Coverage metrics include total percentage and threshold comparison.
────────────────────────────────────────────────────────
```

---

## Subcommand: edit

**Usage:** `/extensions edit <command>:<point>`

**Example:** `/extensions edit prepare-release:post-validation`

### Step 1: Locate Extension Block

Same as `view` subcommand.

### Step 2: Present Current Content

Show the current content of the extension block.

### Step 3: ASK USER for New Content

Prompt user for the new content to replace the extension block.

### Step 4: Update File

Replace the content between START and END markers with the new content.

### Step 5: Confirm Change

Report the updated extension block.

---

## Subcommand: validate

**Usage:** `/extensions validate`

### Step 1: Parse All Extension Blocks

Scan all target command files for extension markers.

### Step 2: Check for Issues

Validate each block for:

| Check | Description |
|-------|-------------|
| Matching pairs | Every START has a matching END |
| Valid names | Extension point names match `[a-z][a-z0-9-]*` pattern |
| No nesting | Extension blocks are not nested |
| Documented | Each extension point appears in the "Available Extension Points" table |

### Step 3: Report Results

```
Extension Validation Report

/create-branch
  ✓ pre-create: valid
  ✓ post-create: valid
  ✓ checklist: valid

/prepare-release
  ✓ post-analysis: valid
  ✓ pre-validation: valid
  ✓ post-validation: valid
  ✓ post-prepare: valid
  ✓ pre-tag: valid
  ✓ post-tag: valid
  ✓ checklist-before-tag: valid
  ✓ checklist-after-tag: valid

...

Summary: N extension points validated, 0 issues found
```

If issues are found:

```
/some-command
  ✗ bad-block: START marker found but no matching END
  ✗ Another_Bad: Invalid name format (must be lowercase with hyphens)
```

---

## Subcommand: matrix

**Usage:** `/extensions matrix`

### Step 1: Collect All Extension Points

Build a union of all extension point names across all target commands.

### Step 2: Build Matrix

For each command, check if each extension point exists and has content.

### Step 3: Format Output

```
Extension Point Matrix

Point                | create | prepare | beta |
---------------------|--------|---------|------|
pre-create           |   ●    |    -    |  -   |
post-create          |   ○    |    -    |  -   |
checklist            |   ●    |    -    |  -   |
post-analysis        |   -    |    ●    |  ●   |
pre-validation       |   -    |    ○    |  ○   |
post-validation      |   -    |    ●    |  ●   |
post-prepare         |   -    |    ●    |  ○   |
pre-tag              |   -    |    ○    |  ○   |
post-tag             |   -    |    ●    |  ●   |
checklist-before-tag |   -    |    ●    |  ●   |
checklist-after-tag  |   -    |    ●    |  ●   |
pre-close            |   -    |    ●    |  -   |
post-close           |   -    |    ○    |  -   |
checklist-close      |   -    |    ●    |  -   |

Legend: ● = has content, ○ = empty, - = not applicable
```

---

## Subcommand: recipes

**Usage:** `/extensions recipes [category]`

Browse common patterns for extension points.

### Step 1: Load Recipe Data

Read recipes from `.claude/metadata/extension-recipes.json`.

### Step 2: Display Based on Arguments

**If no category specified:** Show category list

```
Extension Recipes

Available categories:

  ci            CI build verification and merge gates
  coverage      Test coverage threshold enforcement
  notifications Slack, Discord, email alerts on release events
  docs          Auto-generate changelog, release notes
  security      Dependency audit, secret scanning

Usage:
  /extensions recipes <category>     Show recipes in category
  /extensions recipes ci             Show CI integration recipes

Tip: Use /extensions edit <command>:<point> to add a recipe to your workflow.
```

**If category specified:** Show recipes in that category

```
CI Recipes

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Recipe: ci-gate-before-merge
Purpose: Block merge/tag until CI passes
Extension Points: post-pr-create, pre-tag, post-prepare

Template:
┌─────────────────────────────────────────────────────────────────────────────┐
│ ### Wait for CI                                                             │
│                                                                             │
│ ```bash                                                                     │
│ node .claude/scripts/shared/wait-for-ci.js                                  │
│ ```                                                                         │
│                                                                             │
│ **If CI fails, STOP and report the error.**                                 │
└─────────────────────────────────────────────────────────────────────────────┘

Prerequisites:
  ✓ wait-for-ci.js (included with framework)

Apply to:
  /extensions edit prepare-release:post-pr-create
  /extensions edit prepare-release:pre-tag

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

[Additional recipes in category...]

N recipes in 'ci' category
```

### Step 3: Handle Unknown Category

If category not found:

```
Unknown category: 'foo'

Available categories: ci, coverage, notifications, docs, security
```

---

## Extension Point Naming Convention

| Pattern | Used By | Purpose |
|---------|---------|---------|
| `pre-*` | All | Before a workflow phase |
| `post-*` | All | After a workflow phase |
| `pre-commit` | prepare-release, prepare-beta | Generate artifacts before commit |
| `checklist` | create-branch | Single verification checklist |
| `checklist-before-*` | prepare-release, prepare-beta | Pre-action verification items |
| `checklist-after-*` | prepare-release, prepare-beta | Post-action verification items |

---

## Help

**Usage:** `/extensions --help` or `/extensions`

Display this command's subcommands and usage information.

---

**End of Extensions Command**
