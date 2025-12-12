# Sub-Issues Guide

Manage parent-child issue relationships with gh-pmu. Create epic/story hierarchies, track progress, and work across repositories.

## Quick Start

```bash
# Link existing issue as sub-issue
gh pmu sub add 10 15    # Issue 15 becomes child of 10

# Create new sub-issue
gh pmu sub create --parent 10 --title "Subtask"

# View hierarchy and progress
gh pmu sub list 10
```

## Commands

### sub add

Link an existing issue as a sub-issue of a parent.

```bash
gh pmu sub add <parent> <child>
```

**Example:**
```bash
$ gh pmu sub add 10 15
✓ Linked #15 as sub-issue of #10
```

### sub create

Create a new issue directly under a parent.

```bash
gh pmu sub create --parent <number> --title "Title" [flags]
```

**Flags:**
- `--parent` - Parent issue number (required)
- `--title` - Sub-issue title (required)
- `--body` - Issue body
- `--status` - Set status field
- `--priority` - Set priority field
- `--repo` - Create in different repository

**Examples:**
```bash
# Basic sub-issue
gh pmu sub create --parent 10 --title "Implement login"

# With project fields
gh pmu sub create --parent 10 --title "Add tests" --status ready --priority p1

# In another repository
gh pmu sub create --parent 10 --title "Backend API" --repo myorg/backend
```

### sub list

Show sub-issues and progress for a parent issue.

```bash
gh pmu sub list <parent>
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

**JSON output:**
```bash
gh pmu sub list 10 --json
```

### sub remove

Unlink a sub-issue from its parent (does not delete the issue).

```bash
gh pmu sub remove <parent> <child>
```

**Example:**
```bash
$ gh pmu sub remove 10 15
✓ Unlinked #15 from parent #10
```

## Workflows

### Epic/Story Pattern

Organize work into epics containing multiple stories:

```bash
# Create epic
gh pmu create --title "Epic: User Authentication" --label epic

# Add stories as sub-issues
gh pmu sub create --parent 20 --title "Story: Login flow" --label story
gh pmu sub create --parent 20 --title "Story: Password reset" --label story
gh pmu sub create --parent 20 --title "Story: OAuth integration" --label story

# Track progress
gh pmu view 20
```

**Output:**
```
#20 Epic: User Authentication
Status: In progress | Priority: P1

Sub-issues: [██████░░░░] 33% (1/3)
  ✓ #21 Story: Login flow
  ○ #22 Story: Password reset
  ○ #23 Story: OAuth integration
```

### Breaking Down Issues

Use `split` to convert a checklist into sub-issues:

```bash
# Issue #30 body contains:
# - [ ] Design API schema
# - [ ] Implement endpoints
# - [ ] Add validation
# - [ ] Write tests

gh pmu split 30 --from body
```

**Output:**
```
Created 4 sub-issues for #30:
  #31 Design API schema
  #32 Implement endpoints
  #33 Add validation
  #34 Write tests
```

Or create from command line:

```bash
gh pmu split 30 "Design API" "Implement endpoints" "Add tests"
```

### Cross-Repository Sub-Issues

Create sub-issues in different repositories while maintaining the hierarchy:

```bash
# Parent in myorg/planning
gh pmu create --title "Feature: Analytics Dashboard"
# Created #50

# Sub-issues in different repos
gh pmu sub create --parent 50 --title "Backend API" --repo myorg/backend
gh pmu sub create --parent 50 --title "Frontend components" --repo myorg/frontend
gh pmu sub create --parent 50 --title "Update docs" --repo myorg/docs
```

### Recursive Status Updates

Update parent and all sub-issues at once:

```bash
# Move entire hierarchy to done
gh pmu move 20 --status done --recursive
```

**Output:**
```
✓ Updated #20 Epic: User Authentication → Done
✓ Updated #21 Story: Login flow → Done
✓ Updated #22 Story: Password reset → Done
✓ Updated #23 Story: OAuth integration → Done
```

## Progress Tracking

### In View Output

`gh pmu view` shows sub-issue progress automatically:

```
#20 Epic: User Authentication
...
Sub-issues: [████████░░] 80% (4/5)
```

### In Board View

`gh pmu board` shows issues with sub-issue counts:

```
┌──────────┬─────────────────────┐
│ Backlog  │ In Progress         │
├──────────┼─────────────────────┤
│ #25 (3)  │ #20 Epic (4/5)      │
│ #26      │ #21                 │
└──────────┴─────────────────────┘
```

### GitHub UI

Sub-issue relationships appear in GitHub's native sub-issues UI, including:
- Progress bars on parent issues
- Hierarchy view in project boards
- Quick navigation between parent/child

## Best Practices

1. **Use clear naming conventions**
   - Epics: "Epic: Feature Name"
   - Stories: "Story: User action" or just descriptive titles

2. **Keep hierarchies shallow**
   - Parent → Sub-issues (2 levels max)
   - Avoid deep nesting

3. **Label appropriately**
   - Add `epic` label to parent issues
   - Add `story` label to sub-issues (optional)

4. **Track at the right level**
   - Update sub-issue status as work progresses
   - Parent status reflects overall state

5. **Use split for consistency**
   - Convert checklists to sub-issues for better tracking
   - Each sub-issue gets its own status and assignee

## See Also

- [Commands Reference](commands.md) - All sub commands
- [Batch Operations](batch-operations.md) - Using split
- [Configuration](configuration.md) - Field aliases
