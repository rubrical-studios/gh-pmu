---
version: "v0.31.0"
description: Verify and close PRD tracker (project)
argument-hint: "<prd-issue-number>"
---
<!-- MANAGED -->
# /complete-prd
Verify all epics and stories derived from a PRD are complete, then close the PRD tracker issue.
## Arguments
| Argument | Description |
|----------|-------------|
| `<prd-issue-number>` | PRD tracker issue number (e.g., `151` or `#151`) |
## Prerequisites
- PRD tracker issue with `prd` label
- PRD created via `/create-backlog` (has linked epics)
## Workflow
### Step 1: Validate PRD Issue
Parse issue number, verify `prd` label:
```bash
gh issue view $issue_num --json labels -q '.labels[].name' | grep -q "prd"
```
### Step 2: Find Linked Epics
Find epics containing `**PRD Tracker:** #{issue_num}` in body:
```bash
gh issue list --label "epic" --state all --json number,title,body,state
```
### Step 3: Check Epic Completion
For each linked epic, check state:
```bash
gh issue view #{epic} --json state -q '.state'
```
Collect: total epics, closed epics, open epics list
### Step 4: Check Story Completion
For each epic, get sub-issues:
```bash
gh pmu sub list #{epic} --json
```
Collect: total stories, closed stories, open stories list
### Step 5: Report Status
**If all complete:**
```
✅ PRD #{issue_num} Verification Complete
Epics: {closed}/{total} complete
Stories: {closed}/{total} complete
All work items complete. Closing PRD tracker.
```
Then close:
```bash
gh pmu move $issue_num --status done
gh issue comment $issue_num --body "## PRD Complete ✅
All epics and stories completed.
PRD closed by \`/complete-prd\` on {date}."
```
**If incomplete:**
```
⚠️ PRD #{issue_num} Not Ready for Closure
## Incomplete Work
### Open Epics ({count})
- #{epic}: {title}
### Open Stories ({count})
- #{story}: {title} (Epic: #{epic})
Complete above items before running /complete-prd again.
```
**Do NOT close** if incomplete.
## Verification Logic
| Epics | Stories | Result |
|-------|---------|--------|
| All closed | All closed | ✅ Close PRD |
| Any open | Any state | ❌ Report incomplete |
| All closed | Any open | ❌ Report incomplete |
## Error Handling
| Situation | Response |
|-----------|----------|
| PRD not found | "Issue #N not found" |
| Missing prd label | "Issue #N does not have 'prd' label" |
| No linked epics | Warning + suggest manual verification |
| PRD already closed | "PRD #{N} is already closed" |
**End of /complete-prd Command**
