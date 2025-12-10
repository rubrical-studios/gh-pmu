# Feature Roadmap Proposal: gh pmu v0.5+

**Status:** Draft
**Created:** 2025-12-10
**Author:** Claude Code

---

## Overview

This proposal outlines potential features for gh pmu based on a gap analysis comparing `gh pmu` with the base GitHub CLI (`gh`). Features are organized into three categories:

1. **Complement** - Fill workflow gaps where gh + gh pmu are used together
2. **Extend** - Add project-aware versions of existing gh commands
3. **New Capabilities** - Features not available in gh at all

---

## Category 1: Complement gh

These features streamline common multi-command workflows.

### 1.1 `gh pmu comment`

**Problem:** Adding a comment and updating status requires two commands:
```bash
gh issue comment 123 --body "Starting work"
gh pmu move 123 --status in_progress
```

**Solution:**
```bash
gh pmu comment 123 --body "Starting work" --status in_progress
```

**Flags:**
- `--body` - Comment text (required)
- `--status` - Update status after commenting
- `--priority` - Update priority after commenting
- `--body-file` - Read comment from file

**Complexity:** Low
**Value:** High (daily use pattern)

---

### 1.2 `gh pmu assign`

**Problem:** Assigning an issue often means work is starting, but requires separate status update.

**Solution:**
```bash
gh pmu assign 123 @username --status in_progress
gh pmu assign 123 --me --status in_progress
```

**Flags:**
- `@username` or `--me` - Assignee
- `--status` - Set status when assigning
- `--remove` - Unassign instead

**Complexity:** Low
**Value:** Medium

---

### 1.3 `gh pmu label`

**Problem:** Adding labels often triggers workflow changes (e.g., `bug` label → priority bump).

**Solution:**
```bash
gh pmu label 123 bug --priority p1
gh pmu label 123 --remove enhancement
```

**Flags:**
- Label name(s) as arguments
- `--remove` - Remove instead of add
- `--status`, `--priority` - Set fields when labeling

**Auto-triage integration:**
```yaml
# .gh-pmu.yml
label_rules:
  bug:
    priority: p1
  security:
    priority: p0
    status: in_progress
```

**Complexity:** Medium (config integration)
**Value:** Medium

---

### 1.4 `gh pmu link`

**Problem:** `gh issue develop` links PRs to issues, but there's no way to link related issues.

**Solution:**
```bash
gh pmu link 123 456 --type blocks
gh pmu link 123 456 --type related
gh pmu link 123 456 --type duplicates
```

**Link types:**
- `blocks` / `blocked-by` (inverse)
- `related`
- `duplicates` / `duplicated-by`

**Storage:** Task list in issue body or dedicated section:
```markdown
## Related Issues
- blocks #456
- related #789
```

**Complexity:** Medium
**Value:** High (dependency tracking)

---

## Category 2: Extend gh

Project-aware versions of existing gh commands.

### 2.1 `gh pmu search`

**Problem:** `gh search issues` cannot filter by project field values.

**Solution:**
```bash
gh pmu search "auth bug" --status in_progress --priority p0
gh pmu search --author @me --status in_review
gh pmu search --label bug --no-assignee --status backlog
```

**Flags:**
- Query string (searches title/body)
- `--status`, `--priority` - Project field filters
- `--author`, `--assignee`, `--label` - Standard filters
- `--no-assignee`, `--no-label` - Negative filters
- `--created`, `--updated` - Date filters
- `--json` - JSON output

**Complexity:** Medium
**Value:** High (common need)

---

### 2.2 `gh pmu edit`

**Problem:** Editing issue content and project fields requires two commands.

**Solution:**
```bash
gh pmu edit 123 --title "New title" --status in_progress
gh pmu edit 123 --body-file updated.md --priority p1
gh pmu edit 123 --add-label bug --remove-label enhancement --status backlog
```

**Flags:**
- `--title` - Update title
- `--body` / `--body-file` - Update body
- `--add-label`, `--remove-label` - Modify labels
- `--assignee`, `--remove-assignee` - Modify assignees
- `--status`, `--priority` - Project fields

**Complexity:** Low (wraps gh issue edit + gh pmu move)
**Value:** Medium

---

### 2.3 `gh pmu reopen`

**Problem:** Reopening an issue usually means re-triaging (setting status back).

**Solution:**
```bash
gh pmu reopen 123 --status backlog
gh pmu reopen 123 --status in_progress --comment "Reopening to address feedback"
```

**Flags:**
- `--status` - Status after reopening (default: backlog)
- `--comment` - Add comment explaining reopen

**Complexity:** Low
**Value:** Low-Medium

---

### 2.4 `gh pmu transfer`

**Problem:** `gh issue transfer` moves issue but loses project association.

**Solution:**
```bash
gh pmu transfer 123 other-repo --keep-project
gh pmu transfer 123 other-repo --project other-project
```

**Flags:**
- Destination repo (required)
- `--keep-project` - Re-add to same project after transfer
- `--project` - Add to different project
- `--status`, `--priority` - Set fields in destination

**Complexity:** Medium (cross-repo operations)
**Value:** Low (infrequent use)

---

## Category 3: New Capabilities

Features not available in gh at all.

### 3.1 `gh pmu stale`

**Problem:** No way to find issues stuck in a status for too long.

**Solution:**
```bash
gh pmu stale
gh pmu stale --status in_progress --days 7
gh pmu stale --status in_review --days 3 --json
```

**Output:**
```
Stale Issues (default: >7 days in status)

in_progress (3 issues):
  #123  14 days  Fix authentication bug
  #456   9 days  Add export feature
  #789   8 days  Update documentation

in_review (2 issues):
  #234   5 days  Refactor API client
  #567   4 days  Add unit tests
```

**Flags:**
- `--status` - Filter by status (default: all non-done)
- `--days` - Threshold (default: 7)
- `--json` - JSON output

**Configuration:**
```yaml
# .gh-pmu.yml
stale:
  in_progress: 7
  in_review: 3
  backlog: 30
```

**Complexity:** Medium (need to track status change dates)
**Value:** High (common pain point)

---

### 3.2 `gh pmu standup`

**Problem:** Generating daily standup reports is manual.

**Solution:**
```bash
gh pmu standup
gh pmu standup --user @me --days 1
gh pmu standup --team --days 1
```

**Output:**
```
Standup Report - 2025-12-10

Yesterday:
  ✓ #123 Fix authentication bug (Done)
  ✓ #456 Add export feature (In Review)

Today:
  → #789 Update documentation (In Progress)
  → #234 Refactor API client (In Progress)

Blocked:
  ⚠ #567 Waiting on API access (#890)
```

**Flags:**
- `--user` - Filter by assignee (default: @me)
- `--team` - Show all assignees
- `--days` - Lookback period (default: 1)
- `--format` - Output format (text, markdown, json)
- `--output` - Write to file

**Complexity:** Medium
**Value:** High (daily use)

---

### 3.3 `gh pmu bulk`

**Problem:** Updating multiple issues requires loops or repetition.

**Solution:**
```bash
gh pmu bulk 123 456 789 --status done
gh pmu bulk --label bug --status in_progress --priority p1
gh pmu bulk --query "is:open label:stale" --close --reason not_planned
```

**Flags:**
- Issue numbers as arguments, OR
- `--label`, `--status`, `--query` - Select issues by filter
- `--set-status`, `--set-priority` - Fields to update
- `--add-label`, `--remove-label` - Label changes
- `--close`, `--reason` - Close issues
- `--dry-run` - Preview changes
- `--confirm` - Skip confirmation prompt

**Output:**
```
Bulk operation: Set status=done

Affected issues (3):
  #123 Fix authentication bug
  #456 Add export feature
  #789 Update documentation

Proceed? [y/N]
```

**Complexity:** Medium
**Value:** High (saves significant time)

---

### 3.4 `gh pmu sprint`

**Problem:** GitHub Projects v2 has Iteration fields but no CLI support.

**Solution:**
```bash
gh pmu sprint list
gh pmu sprint current
gh pmu sprint add 123 456 --sprint "Sprint 5"
gh pmu sprint burndown
gh pmu sprint report
```

**Subcommands:**
- `list` - List all sprints/iterations
- `current` - Show current sprint issues
- `add` - Add issues to sprint
- `remove` - Remove from sprint
- `burndown` - ASCII burndown chart
- `report` - Sprint summary report

**Output (burndown):**
```
Sprint 5 Burndown (Dec 2-13)

Points │
   20  │██
   15  │████
   10  │██████        ·····
    5  │████████      ·········
    0  │──────────────────────────
       │ M  T  W  T  F  M  T  W  T  F
         ▲ Today

Remaining: 8 points | Velocity: 2.4/day | On track: ✓
```

**Complexity:** High (iteration field support, calculations)
**Value:** Medium-High (teams using sprints)

---

### 3.5 `gh pmu blocked`

**Problem:** No native dependency/blocker tracking in GitHub.

**Solution:**
```bash
gh pmu blocked 123 --by 456
gh pmu blocked 123 --unblock
gh pmu blocked list
gh pmu blocked graph
```

**Subcommands:**
- `gh pmu blocked 123 --by 456` - Mark 123 blocked by 456
- `gh pmu blocked 123 --unblock` - Remove blocker
- `gh pmu blocked list` - Show all blocked issues
- `gh pmu blocked graph` - Dependency graph (ASCII)

**Storage:** Task list or keywords in body:
```markdown
## Blockers
- [ ] Blocked by #456
- [ ] Waiting on #789
```

**Output (graph):**
```
Dependency Graph

#123 Fix auth
  └── blocked by #456 Add API endpoint
        └── blocked by #789 Database migration

#234 Update docs
  └── blocked by #123 Fix auth
```

**Complexity:** Medium
**Value:** Medium (useful for complex projects)

---

### 3.6 `gh pmu report`

**Problem:** Generating status reports is manual.

**Solution:**
```bash
gh pmu report weekly
gh pmu report release v0.5.0
gh pmu report custom --since 2025-12-01 --until 2025-12-10
```

**Report types:**
- `weekly` - Issues completed this week, in progress, upcoming
- `release` - Changes since tag, grouped by type
- `custom` - Configurable date range

**Output (weekly):**
```markdown
# Weekly Report - Dec 4-10, 2025

## Completed (5)
- #259 Add --browser flag to history command
- #260 Show affected files in directory history
- #258 Show commit body in detailed view
- #257 Search parent directories for config
- #256 Add detailed single-file view

## In Progress (2)
- #261 Add sprint support
- #262 Implement bulk operations

## Blockers
- #263 Waiting on upstream API fix

## Metrics
- Completed: 5 issues
- Velocity: 1.0 issues/day
- Avg cycle time: 2.3 days
```

**Flags:**
- `--format` - Output format (markdown, text, json)
- `--output` - Write to file
- `--since`, `--until` - Date range

**Complexity:** Medium
**Value:** Medium-High (weekly use)

---

### 3.7 `gh pmu watch`

**Problem:** `gh pmu board` is a static snapshot.

**Solution:**
```bash
gh pmu watch
gh pmu watch --refresh 30
gh pmu watch --status in_progress,in_review
```

**Behavior:**
- Live-updating terminal board view
- Refreshes on interval or webhook
- Highlights changes since last refresh
- Keyboard controls: q=quit, r=refresh, arrows=navigate

**Complexity:** High (requires bubbletea TUI)
**Value:** Low-Medium (niche use case)

---

### 3.8 `gh pmu template`

**Problem:** Issue templates don't set project fields.

**Solution:**
```bash
gh pmu template list
gh pmu template use bug --title "Login fails"
gh pmu template create --name "feature-request"
```

**Configuration:**
```yaml
# .gh-pmu.yml
templates:
  bug:
    labels: [bug]
    priority: p1
    status: backlog
    body_template: .github/ISSUE_TEMPLATE/bug.md
  feature:
    labels: [enhancement]
    priority: p2
    status: backlog
```

**Complexity:** Low-Medium
**Value:** Medium

---

### 3.9 `gh pmu clone`

**Problem:** No way to duplicate an issue with modifications.

**Solution:**
```bash
gh pmu clone 123
gh pmu clone 123 --title "Clone of #123" --status backlog
gh pmu clone 123 --repo other-repo
```

**Behavior:**
- Copies title, body, labels
- Optionally copies assignees, milestone
- Sets project fields as specified
- Adds reference to original: "Cloned from #123"

**Complexity:** Low
**Value:** Low-Medium

---

### 3.10 `gh pmu merge`

**Problem:** Merging duplicate issues is manual.

**Solution:**
```bash
gh pmu merge 123 456 --into 123
gh pmu merge 123 456 789 --into 123 --close-duplicates
```

**Behavior:**
- Combines bodies (appends or merges checklists)
- Copies comments to target (optional)
- Links duplicates
- Closes source issues with "duplicate" reason
- Transfers sub-issues to target

**Complexity:** Medium
**Value:** Low (infrequent)

---

## Implementation Priority

### Tier 1: High Value, Low-Medium Complexity
| Feature | Complexity | Value | Rationale |
|---------|------------|-------|-----------|
| `stale` | Medium | High | Common pain point, actionable output |
| `standup` | Medium | High | Daily use, immediate value |
| `bulk` | Medium | High | Saves significant repetitive work |
| `search` | Medium | High | Project-aware search is missing |

### Tier 2: Medium Value, Low Complexity
| Feature | Complexity | Value | Rationale |
|---------|------------|-------|-----------|
| `comment` | Low | High | Streamlines daily workflow |
| `edit` | Low | Medium | Wraps existing commands |
| `link` | Medium | High | Dependency tracking needed |
| `report` | Medium | Medium-High | Weekly/release reporting |

### Tier 3: Specialized Features
| Feature | Complexity | Value | Rationale |
|---------|------------|-------|-----------|
| `sprint` | High | Medium-High | Only for iteration users |
| `blocked` | Medium | Medium | Subset of `link` |
| `template` | Low-Medium | Medium | Config-driven |
| `assign` | Low | Medium | Simple wrapper |
| `label` | Medium | Medium | Auto-triage integration |

### Tier 4: Nice to Have
| Feature | Complexity | Value | Rationale |
|---------|------------|-------|-----------|
| `watch` | High | Low-Medium | Niche use case |
| `clone` | Low | Low-Medium | Infrequent need |
| `merge` | Medium | Low | Rare operation |
| `reopen` | Low | Low-Medium | Simple wrapper |
| `transfer` | Medium | Low | Infrequent, complex |

---

## Suggested Roadmap

### v0.5.0 - Workflow Efficiency
- `gh pmu stale` - Find stuck issues
- `gh pmu standup` - Daily reports
- `gh pmu comment` - Comment + status update

### v0.6.0 - Bulk Operations
- `gh pmu bulk` - Multi-issue updates
- `gh pmu search` - Project-aware search

### v0.7.0 - Dependencies & Reporting
- `gh pmu link` - Issue relationships
- `gh pmu blocked` - Blocker tracking
- `gh pmu report` - Status reports

### v0.8.0 - Sprint Support
- `gh pmu sprint` - Iteration management
- `gh pmu template` - Project-aware templates

---

## Open Questions

1. **Storage for relationships:** Should `link`/`blocked` use issue body, comments, or external storage?
2. **Status change tracking:** How to track when status changed for `stale` detection? (git log parsing? API events?)
3. **Sprint/Iteration support:** Does the GraphQL API expose iteration fields adequately?
4. **Standup scope:** Should standup be user-specific or project-wide by default?

---

## Next Steps

1. Review and prioritize features
2. Create individual issues for approved features
3. Design detailed specs for Tier 1 features
4. Begin implementation with `stale` or `standup`
