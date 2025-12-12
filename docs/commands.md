# Command Reference

Complete reference for all gh-pmu commands.

## Overview

```
gh pmu [command]

Project Management:
  init        Initialize configuration
  list        List issues with project metadata
  view        View issue with project fields
  create      Create issue with project fields
  move        Update issue project fields
  close       Close issue with optional reason
  board       View project board in terminal
  field       Manage custom project fields

Sub-Issue Management:
  sub add     Link existing issue as sub-issue
  sub create  Create new sub-issue under parent
  sub list    List sub-issues of a parent
  sub remove  Unlink sub-issue from parent

Batch Operations:
  intake      Find and add untracked issues to project
  triage      Bulk update issues based on config rules
  split       Create sub-issues from checklist or arguments

Utilities:
  filter      Filter piped issue JSON by project fields
  history     Show git commit history with issue references

Flags:
  -h, --help      help for gh-pmu
  -v, --version   version for gh-pmu
```

---

## Project Management Commands

### init

Initialize or refresh project configuration.

```bash
# Interactive setup
gh pmu init

# Refresh metadata only
gh pmu init --refresh
```

**Output:**
```
? Select a project: my-project (#5)
? Select repositories to track: myorg/frontend, myorg/backend
✓ Configuration saved to .gh-pmu.yml
✓ Fetched 8 project fields
```

### list

List issues with project metadata.

```bash
# List all issues in project
gh pmu list

# Filter by status
gh pmu list --status in_progress

# Filter by priority
gh pmu list --priority p0

# Combine filters
gh pmu list --status ready --priority p1

# JSON output
gh pmu list --json

# Specify repository
gh pmu list --repo owner/other-repo
```

**Output:**
```
#   Title                          Status        Priority
42  Add user authentication        In progress   P1
43  Fix login redirect             Ready         P0
45  Update documentation           Backlog       P2
```

### view

View issue with project fields and sub-issue progress.

```bash
# View issue
gh pmu view 42

# JSON output
gh pmu view 42 --json

# Specify repository
gh pmu view 42 --repo owner/other-repo
```

**Output:**
```
#42 Add user authentication
Status: In progress | Priority: P1 | Size: M

Labels: enhancement, backend
Assignees: @developer

Sub-issues: [████████░░] 80% (4/5)
  ✓ #46 Design auth flow
  ✓ #47 Implement JWT tokens
  ✓ #48 Add login endpoint
  ✓ #49 Add logout endpoint
  ○ #50 Write tests

https://github.com/myorg/app/issues/42
```

### create

Create issue with project fields set in one command.

```bash
# Basic creation
gh pmu create --title "New feature"

# With project fields
gh pmu create --title "Fix bug" --status ready --priority p0

# With body
gh pmu create --title "Add caching" --body "Implement Redis caching for API"

# With labels
gh pmu create --title "Security fix" --label bug --label security
```

**Output:**
```
✓ Created issue #51: New feature
  • Status → Backlog
  • Priority → P2
🔗 https://github.com/myorg/app/issues/51
```

### move

Update issue project fields.

```bash
# Update status
gh pmu move 42 --status in_review

# Update multiple fields
gh pmu move 42 --status done --priority p0

# Recursive update (includes sub-issues)
gh pmu move 42 --status done --recursive

# Specify repository
gh pmu move 42 --status done --repo owner/other-repo
```

**Output:**
```
✓ Updated issue #42: Add user authentication
  • Status → In review
🔗 https://github.com/myorg/app/issues/42
```

### close

Close issue with optional state reason.

```bash
# Close as completed (default)
gh pmu close 42

# Close as not planned
gh pmu close 42 --reason not_planned

# Close as duplicate
gh pmu close 42 --reason duplicate

# Specify repository
gh pmu close 42 --repo owner/other-repo
```

**Reason aliases:**
| Alias | GitHub State Reason |
|-------|---------------------|
| `completed` | completed |
| `not_planned`, `wontfix` | not_planned |
| `duplicate`, `dupe` | duplicate |

**Output:**
```
✓ Closed issue #42: Add user authentication (completed)
🔗 https://github.com/myorg/app/issues/42
```

### board

View project board in terminal.

```bash
# Show board
gh pmu board

# Compact view
gh pmu board --compact
```

**Output:**
```
┌──────────┬───────────────┬─────────────┬────────┐
│ Backlog  │ Ready         │ In Progress │ Done   │
├──────────┼───────────────┼─────────────┼────────┤
│ #52 Add  │ #43 Fix login │ #42 Auth    │ #41    │
│ #53 Docs │               │ #44 API     │ #40    │
│          │               │             │ #39    │
└──────────┴───────────────┴─────────────┴────────┘
```

### field

Manage project fields.

```bash
# List all fields
gh pmu field list

# Create a new single-select field
gh pmu field create --name "Sprint" --type single_select --options "Sprint 1,Sprint 2,Sprint 3"
```

**Output:**
```
Name        Type           Options
Status      SINGLE_SELECT  Backlog, Ready, In progress, In review, Done
Priority    SINGLE_SELECT  P0, P1, P2
Size        SINGLE_SELECT  XS, S, M, L, XL
Estimate    NUMBER         -
Start date  DATE           -
```

---

## Sub-Issue Commands

See [Sub-Issues Guide](sub-issues.md) for detailed workflows.

### sub add

Link existing issue as sub-issue.

```bash
gh pmu sub add 10 15    # Issue 15 becomes sub-issue of 10

# Specify repository
gh pmu sub add 10 15 --repo owner/other-repo
```

### sub create

Create new sub-issue under parent.

```bash
gh pmu sub create --parent 10 --title "Subtask 1"

# With project fields
gh pmu sub create --parent 10 --title "Subtask" --status ready --priority p1

# Cross-repository
gh pmu sub create --parent 10 --title "Backend task" --repo owner/backend
```

### sub list

List sub-issues of a parent.

```bash
gh pmu sub list 10

# JSON output
gh pmu sub list 10 --json
```

**Output:**
```
Parent: #10 Epic: User Management

Sub-issues:
  ✓ #11 Design user schema
  ✓ #12 Implement CRUD endpoints
  ○ #13 Add validation
  ○ #14 Write tests

Progress: [████████░░] 50% (2/4)
```

### sub remove

Unlink sub-issue from parent.

```bash
gh pmu sub remove 10 15

# Specify repository
gh pmu sub remove 10 15 --repo owner/other-repo
```

---

## Batch Operations

See [Batch Operations Guide](batch-operations.md) for detailed workflows.

### intake

Find and add untracked issues to project.

```bash
# Preview untracked issues
gh pmu intake --dry-run

# Add to project with defaults
gh pmu intake --apply
```

### triage

Bulk update issues based on config rules.

```bash
# Preview rule effects
gh pmu triage untracked --dry-run

# Apply rule
gh pmu triage untracked --apply

# Interactive mode (prompts for each issue)
gh pmu triage untracked --interactive
```

### split

Create sub-issues from checklist or arguments.

```bash
# From checklist in issue body
gh pmu split 42 --from body

# From arguments
gh pmu split 42 "Task 1" "Task 2" "Task 3"

# With status for new sub-issues
gh pmu split 42 --from body --status ready

# Specify repository
gh pmu split 42 --from body --repo owner/other-repo
```

---

## Utilities

### filter

Filter piped issue JSON by project fields.

```bash
# Filter by status
gh issue list --json number,title | gh pmu filter --status ready

# Filter with JSON output
gh issue list --json number,title | gh pmu filter --status in_progress --json

# From another repository
gh issue list -R owner/repo --json number,title | gh pmu filter --priority p0
```

### history

Show git commit history with issue references.

```bash
# Current directory
gh pmu history

# Specific path
gh pmu history src/

# Limit results
gh pmu history --limit 20
```

**Output:**
```
abc1234 feat: Add login endpoint (#42)
def5678 fix: Handle null user (#43)
ghi9012 docs: Update API reference
```

---

## Global Flags

These flags work with most commands:

| Flag | Description |
|------|-------------|
| `--repo owner/repo` | Specify repository (overrides config) |
| `--json` | Output in JSON format |
| `--help` | Show command help |

## See Also

- [Configuration Guide](configuration.md) - Setup and field aliases
- [Sub-Issues Guide](sub-issues.md) - Hierarchy management
- [Batch Operations](batch-operations.md) - Intake, triage, split
- [gh vs gh pmu](gh-comparison.md) - When to use which
