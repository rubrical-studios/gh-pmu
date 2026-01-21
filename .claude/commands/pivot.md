---
version: "v0.29.1"
description: Review stories for direction change (project)
argument-hint: "[epic-number|prd-name]"
---
<!-- MANAGED -->
# /pivot
Review and triage stories when project direction changes. Helps manage scope realignment.
## Arguments
| Argument | Description |
|----------|-------------|
| `[epic-number]` | Epic issue number to pivot (e.g., `42` or `#42`) |
| `[prd-name]` | PRD name to pivot all related stories |
## Usage
```bash
/pivot 42           # Pivot stories under Epic #42
/pivot Auth-System  # Pivot all stories from Auth-System PRD
/pivot              # Interactive - prompts for selection
```
## Workflow
### Phase 1: Identify Scope
**Step 1:** Determine target (number → epic, text → PRD name)
**If no argument:** **ASK USER:** Epic number or PRD name?
**Step 2:** Validate target exists
### Phase 2: Document Pivot Reason
**ASK USER:** What is the new direction or reason for the pivot?
### Phase 3: List Stories
For epic: `gh pmu sub list $epic_num --json number,title,state`
For PRD: `gh issue list --label "story" --search "PRD:$prd_name" --json number,title,state`
Display table of stories with issue number, title, status
### Phase 4: Review Each Story
For each open story, present options:
```
Story #101: User login
Status: In Progress
Options:
1. Keep - Aligns with new direction
2. Archive - May be relevant later (parking lot)
3. Close - No longer needed (not planned)
4. Skip - Decide later
```
Record decision for each story.
### Phase 5: Apply Actions
| Decision | Action |
|----------|--------|
| Keep | No change |
| Archive | `gh pmu move #{num} --status parking_lot` |
| Close | `gh issue close #{num} --reason "not planned" --comment "Closed during pivot: {reason}"` |
| Skip | No change |
Add documentation comment to parent epic/PRD with decisions table.
### Phase 6: Report Summary
```
Pivot complete: Epic #{num} / PRD {name}
Reason: {pivot_reason}
Stories reviewed: {total}
  ✅ Kept: {count}
  📦 Archived: {count}
  ❌ Closed: {count}
  ⏸️ Pending: {count}
Next steps:
1. Review kept stories for priority adjustment
2. Continue work: work #{next_story}
```
## When to Use
**Use for:** Significant requirement changes, market/business shifts, scope realignment
**Not for:** Minor priority changes, single story updates, adding stories (use `/add-story`)
## Error Handling
| Situation | Response |
|-----------|----------|
| Epic not found | "Issue #N not found or not an epic" |
| PRD not found | "No PRD found matching '{name}'" |
| No open stories | "No open stories found to review" |
**End of /pivot Command**
