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

Workflow Commands:
  microsprint Manage microsprints for AI-assisted development
  release     Manage releases for version-based deployment
  patch       Manage patches for hotfix deployment

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

# Filter by status (🆕 unique to gh-pmu)
gh pmu list --status in_progress

# Filter by priority (🆕 unique to gh-pmu)
gh pmu list --priority p0

# Filter to parent issues only (🆕 unique to gh-pmu)
gh pmu list --has-sub-issues

# Combine filters
gh pmu list --status ready --priority p1

# JSON output
gh pmu list --json

# Specify repository
gh pmu list --repo owner/other-repo
```

**Flags unique to gh-pmu:**
| Flag | Purpose |
|------|---------|
| `--status` | Filter by project status field |
| `--priority` | Filter by project priority field |
| `--has-sub-issues` | Show only parent issues |

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

# With project fields (🆕 unique to gh-pmu)
gh pmu create --title "Fix bug" --status ready --priority p0

# Create from YAML/JSON file (🆕 unique to gh-pmu)
gh pmu create --from-file issue.yml

# Interactive mode with prompts (🆕 unique to gh-pmu)
gh pmu create --interactive

# With body
gh pmu create --title "Add caching" --body "Implement Redis caching for API"

# With body from file
gh pmu create --title "Add caching" --body-file description.md

# Open editor to compose body
gh pmu create --title "Add caching" --editor

# Use issue template
gh pmu create --title "Bug report" --template bug

# Open in browser after creation
gh pmu create --title "New feature" --web

# With labels
gh pmu create --title "Security fix" --label bug --label security
```

**Flags unique to gh-pmu:**
| Flag | Purpose |
|------|---------|
| `--status` | Set project status field on create |
| `--priority` | Set project priority field on create |
| `--microsprint` | Assign to microsprint (use 'current' for active) |
| `--from-file` | Create issue from YAML/JSON file |
| `--interactive` | Prompt for all fields interactively |

**Flags matching `gh issue create`:**
| Flag | Purpose |
|------|---------|
| `--body-file` / `-F` | Read body text from file (use "-" for stdin) |
| `--editor` / `-e` | Open editor to compose body |
| `--template` / `-T` | Use issue template from `.github/ISSUE_TEMPLATE/` |
| `--web` / `-w` | Open browser after creating issue |

**Output:**
```
✓ Created issue #51: New feature
  • Status → Backlog
  • Priority → P2
🔗 https://github.com/myorg/app/issues/51
```

### move

Update issue project fields. **🆕 This entire command is unique to gh-pmu.**

```bash
# Update status
gh pmu move 42 --status in_review

# Update multiple fields
gh pmu move 42 --status done --priority p0

# Recursive update - cascade to all sub-issues (🆕 unique)
gh pmu move 42 --status done --recursive

# Preview changes without applying (🆕 unique)
gh pmu move 42 --status done --recursive --dry-run

# Limit recursion depth (🆕 unique)
gh pmu move 42 --status in_progress --recursive --depth 2

# Skip confirmation prompt (🆕 unique)
gh pmu move 42 --status done --recursive --yes

# Specify repository
gh pmu move 42 --status done --repo owner/other-repo
```

**Flags unique to gh-pmu:**
| Flag | Purpose |
|------|---------|
| `--recursive` | Apply changes to all sub-issues |
| `--dry-run` | Preview what would change |
| `--depth` | Limit recursion depth (default 10) |
| `--yes` | Skip confirmation for recursive ops |

**Output:**
```
✓ Updated issue #42: Add user authentication
  • Status → In review
🔗 https://github.com/myorg/app/issues/42
```

**Recursive output:**
```
✓ Updated #42 Epic: Auth → Done
✓ Updated #43 Login flow → Done
✓ Updated #44 Password reset → Done
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

# Update project status to 'done' before closing (🆕 unique to gh-pmu)
gh pmu close 42 --update-status

# Combine: update status and close with reason
gh pmu close 42 --reason completed --update-status

# Specify repository
gh pmu close 42 --repo owner/other-repo
```

**Flags unique to gh-pmu:**
| Flag | Purpose |
|------|---------|
| `--update-status` | Move issue to 'done' status before closing |

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

Create new sub-issue under parent. **🆕 This entire command is unique to gh-pmu.**

```bash
gh pmu sub create --parent 10 --title "Subtask 1"

# With project fields
gh pmu sub create --parent 10 --title "Subtask" --status ready --priority p1

# Cross-repository
gh pmu sub create --parent 10 --title "Backend task" --repo owner/backend

# Control inheritance from parent (🆕 unique flags)
gh pmu sub create --parent 10 --title "Task" --no-inherit-labels
gh pmu sub create --parent 10 --title "Task" --inherit-assignees
```

**Flags unique to gh-pmu:**
| Flag | Purpose |
|------|---------|
| `--inherit-labels` | Copy labels from parent (default: true) |
| `--inherit-milestone` | Copy milestone from parent (default: true) |
| `--inherit-assignees` | Copy assignees from parent (default: false) |

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

Create sub-issues from checklist or arguments. **🆕 This entire command is unique to gh-pmu.**

```bash
# From checklist in issue body (🆕 unique)
gh pmu split 42 --from body

# From external file (🆕 unique)
gh pmu split 42 --from tasks.md

# From arguments
gh pmu split 42 "Task 1" "Task 2" "Task 3"

# Preview without creating (🆕 unique)
gh pmu split 42 --from body --dry-run

# With status for new sub-issues
gh pmu split 42 --from body --status ready

# Specify repository
gh pmu split 42 --from body --repo owner/other-repo
```

**Flags unique to gh-pmu:**
| Flag | Purpose |
|------|---------|
| `--from` | Source: 'body' (issue body) or file path |
| `--dry-run` | Preview what would be created |

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

## Workflow Commands

Commands for managing development workflows at different cadences.

### microsprint

Manage microsprints for AI-assisted development (hour-scale work batches).

```bash
# Start a new microsprint
gh pmu microsprint start
gh pmu microsprint start --name "auth"   # With optional suffix

# Add/remove issues from current microsprint
gh pmu microsprint add 42
gh pmu microsprint remove 42

# View current microsprint
gh pmu microsprint current
gh pmu microsprint current --refresh     # Update tracker issue body

# Close microsprint with artifacts
gh pmu microsprint close                 # Interactive retrospective prompts
gh pmu microsprint close --skip-retro    # Skip prompts, generate empty template
gh pmu microsprint close --commit        # Auto-commit artifacts

# List microsprint history
gh pmu microsprint list

# Resolve multiple active microsprints
gh pmu microsprint resolve
```

**Artifacts generated on close:**
- `Microsprints/{name}/review.md` - Issue summary
- `Microsprints/{name}/retro.md` - Retrospective notes

### release

Manage releases for version-based deployment (IDPF-Structured projects).

```bash
# Start a new release (interactive version prompt)
gh pmu release start
gh pmu release start --version 1.2.0           # Skip prompt
gh pmu release start --version 1.2.0 --track lts  # Specify track

# Manage issues in release
gh pmu release add 42
gh pmu release remove 42

# View current release
gh pmu release current

# Close release (generates artifacts)
gh pmu release close

# List release history
gh pmu release list
```

**Notes:**
- Interactive version prompt shows latest git tag for reference
- Artifact directory controlled by track prefix (see [configuration](configuration.md#release-configuration))

### patch

Manage patches for hotfix deployment (IDPF-LTS projects).

```bash
# Start a new patch
gh pmu patch start --version 1.2.1

# Add issue to patch (with validation)
gh pmu patch add 42
gh pmu patch add 42 --force              # Skip validation warnings

# Remove issue from patch
gh pmu patch remove 42

# View current patch
gh pmu patch current

# Close patch
gh pmu patch close

# List patch history
gh pmu patch list
```

**Validation (patch add):**
- **Error** if issue has `breaking-change` label (incompatible with patches)
- **Warning** if issue lacks `bug`/`fix`/`hotfix` label
- Validates version is a valid patch increment from latest git tag

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
- [Workflows Guide](workflows.md) - Microsprint, release, and patch
- [Sub-Issues Guide](sub-issues.md) - Hierarchy management
- [Batch Operations](batch-operations.md) - Intake, triage, split
- [gh vs gh pmu](gh-comparison.md) - When to use which
