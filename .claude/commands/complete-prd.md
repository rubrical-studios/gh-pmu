---
version: "v0.42.2"
description: Verify and close PRD tracker (project)
argument-hint: "<issue-number> (e.g., 151)"
---

<!-- MANAGED -->
# /complete-prd

Verify all epics and stories derived from a PRD are complete, then close the PRD tracker issue.

---

## Arguments

| Argument | Description |
|----------|-------------|
| `<prd-issue-number>` | PRD tracker issue number (e.g., `151` or `#151`) |

---

## Usage

```bash
/complete-prd 151
/complete-prd #151
```

---

## Prerequisites

- PRD tracker issue exists with `prd` label
- PRD was created via `/create-backlog` (has linked epics)

---

## Workflow

### Step 1: Validate PRD Issue

Parse issue number (accept `151` or `#151`):
```bash
issue_num="${1#\#}"
```

Verify PRD label:
```bash
gh issue view $issue_num --json labels -q '.labels[].name' | grep -q "prd"
```

**If not a PRD issue:**
```
Error: Issue #$issue_num does not have the 'prd' label.
This command requires a PRD tracker issue.
```

### Step 2: Find Linked Epics

Find all epics that reference this PRD tracker:

```bash
# Search for epics containing "PRD Tracker: #{issue_num}"
gh issue list --label "epic" --state all --json number,title,body,state
```

Filter to epics whose body contains `**PRD Tracker:** #{issue_num}`.

**If no epics found:**
```
Warning: No epics found linked to PRD #{issue_num}.
Check that epics have "PRD Tracker: #{issue_num}" in their body.
```

### Step 3: Check Epic Completion

For each linked epic:

```bash
gh issue view #{epic} --json state -q '.state'
```

**Collect:**
- Total epics count
- Closed epics count
- Open epics list

### Step 4: Check Story Completion

For each linked epic, get sub-issues (stories):

```bash
gh pmu sub list #{epic} --json
```

**Collect:**
- Total stories count
- Closed stories count
- Open stories list

### Step 5: Report Status

**If all complete:**

```
✅ PRD #{issue_num} Verification Complete

## Summary

Epics: {closed_epics}/{total_epics} complete
Stories: {closed_stories}/{total_stories} complete

All work items are complete. Closing PRD tracker.
```

Then close:
```bash
gh pmu move $issue_num --status done
gh issue comment $issue_num --body "## PRD Complete ✅

All epics and stories have been completed.

**Final Summary:**
- Epics: {closed_epics}
- Stories: {closed_stories}

PRD closed by \`/complete-prd\` on {date}."
```

**If incomplete:**

```
⚠️ PRD #{issue_num} Not Ready for Closure

## Incomplete Work

### Open Epics ({open_epic_count})
- #{epic1}: {title}
- #{epic2}: {title}

### Open Stories ({open_story_count})
- #{story1}: {title} (Epic: #{epic})
- #{story2}: {title} (Epic: #{epic})

## Summary

Epics: {closed_epics}/{total_epics} complete
Stories: {closed_stories}/{total_stories} complete

Complete the above items before running /complete-prd again.
```

**Do NOT close** if any items are incomplete.

---

## Verification Logic

```
PRD Complete = (ALL epics CLOSED) AND (ALL stories CLOSED)
```

| Epics | Stories | Result |
|-------|---------|--------|
| All closed | All closed | ✅ Close PRD |
| Any open | Any state | ❌ Report incomplete |
| All closed | Any open | ❌ Report incomplete |

---

## Error Handling

| Situation | Response |
|-----------|----------|
| PRD issue not found | "Issue #N not found. Check the issue number?" |
| Issue missing prd label | "Issue #N does not have 'prd' label." |
| No linked epics found | Warning + suggest manual verification |
| PRD already closed | "PRD #{N} is already closed." |

---

## Related Commands

| Command | Purpose |
|---------|---------|
| `/create-backlog` | Creates PRD tracker with linked epics/stories |
| `/add-story` | Adds stories (updates PRD tracker) |
| `/split-story` | Splits stories (updates PRD tracker) |

---

**End of /complete-prd Command**
