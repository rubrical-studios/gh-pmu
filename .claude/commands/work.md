---
version: "v0.42.2"
description: Start working on issues with validation and auto-TODO (project)
argument-hint: "#issue [#issue...] | all in <status>"
---

<!-- EXTENSIBLE -->
# /work
Start working on one or more issues. Validates issue existence, branch assignment, and issue type, then moves to `in_progress`, extracts auto-TODO, and dispatches to framework methodology.
## Available Extension Points
| Point | Location | Purpose |
|-------|----------|---------|
| `pre-work` | Before any work begins | Environment setup, dependency check |
| `post-work-start` | After issue moved to in_progress | Notifications, time tracking |
| `pre-framework-dispatch` | Before loading framework methodology | Custom methodology injection |
---
## Prerequisites
- `gh pmu` extension installed
- `.gh-pmu.yml` configured in repository root
- Issue assigned to a branch (use `/assign-branch` first)
---
## Arguments
| Argument | Required | Description |
|----------|----------|-------------|
| `#issue` | Yes (one of) | Single issue number (e.g., `#42` or `42`) |
| `#issue #issue...` | | Multiple issue numbers (e.g., `#42 #43 #44`) |
| `all in <status>` | | All issues in given status (e.g., `all in backlog`) |
---
## Execution Instructions
**REQUIRED:** Before executing:
1. **Generate Todo List:** Parse workflow steps, use `TodoWrite` to create todos
2. **Include Extensions:** Add todo item for each non-empty `USER-EXTENSION` block
3. **Track Progress:** Mark todos `in_progress` → `completed` as you work
4. **Post-Compaction:** Re-read spec and regenerate todos after context compaction
**Todo Rules:** One todo per numbered step; one todo per active extension; skip commented-out extensions.
---
## Workflow
### Step 0: Conditional - Clear Todo List
If not working on an epic, clear todo list.
### Step 1: Parse Arguments
Accept these formats:
- `#42` or `42` — single issue
- `#42 #43 #44` — multiple issues
- `all in backlog` or `all in ready` — query by status
**For `all in <status>`:**
```bash
gh pmu list --status <status> --json number,title
```
Collect all matching issue numbers into a list.
**For multiple issues:** Process each sequentially using Steps 2–8.
### Step 2: Validate Issue Exists
```bash
gh issue view $ISSUE --json number,title,labels,body,state
```
**If not found:** "Error: Issue #$ISSUE not found. Check the issue number?" → **STOP** (skip this issue, continue to next if batch)
### Step 3: Validate Branch Assignment
```bash
gh pmu view $ISSUE --json=status,branch
```
**If no branch assigned:** "Issue #$ISSUE is not assigned to a branch. Run `/assign-branch #$ISSUE` first." → **STOP** (skip this issue, continue to next if batch)

<!-- USER-EXTENSION-START: pre-work -->
<!-- USER-EXTENSION-END: pre-work -->

### Step 4: Detect Issue Type
```bash
gh issue view $ISSUE --json labels --jq '.labels[].name'
```
| Label Found | Flow |
|-------------|------|
| `epic` | **Epic Flow** — load sub-issues via `gh pmu sub list $ISSUE --json` |
| Any other | **Standard Flow** — parse acceptance criteria from body |
### Step 5: Move to in_progress
```bash
gh pmu move $ISSUE --status in_progress
```
**Idempotent:** If already `in_progress`, continue silently.

<!-- USER-EXTENSION-START: post-work-start -->
<!-- USER-EXTENSION-END: post-work-start -->

### Step 6: PRD Tracker Auto-Move
Check issue body for PRD tracker reference: `**PRD Tracker:** #NNN`
**If found:**
1. Query: `gh pmu view $PRD_NUM --json=status`
2. If `Backlog` or `Ready`: `gh pmu move $PRD_NUM --status in_progress`
3. Report: `PRD tracker #$PRD_NUM moved to In Progress`
4. If already `in_progress`/`in_review`/`done`: no action (silent)
**If not found:** Continue silently (not an error).
**If PRD tracker issue doesn't exist:** Warn and continue (non-blocking).
### Step 7: Extract Auto-TODO
Generate a todo list based on issue type:
**Epic Flow:**
```
[AUTO-TODO: EPIC]
Create a todo list from these sub-issues:
- #N: Sub-issue title
- #M: Sub-issue title
...
```
**Standard Flow:**
Parse issue body for acceptance criteria checkboxes (`- [ ] ...`):
```
[AUTO-TODO: ACCEPTANCE CRITERIA]
Create a todo list from these acceptance criteria:
- Criterion text 1
- Criterion text 2
...
```
**Batch Flow** (for `all in <status>`):
```
[AUTO-TODO: BATCH ISSUES]
Create a todo list with these issues:
- #N: Issue title
- #M: Issue title
...
```
**If no acceptance criteria found:**
```
No acceptance criteria found in issue body. Create todos manually as needed.
```

<!-- USER-EXTENSION-START: pre-framework-dispatch -->
<!-- USER-EXTENSION-END: pre-framework-dispatch -->

### Step 8: Framework Methodology Dispatch
Read `framework-config.json` for `processFramework` field:
| Framework | Action |
|-----------|--------|
| `IDPF-Agile` | Load `.min-mirror/IDPF-Agile/Agile-Core.md` — follow TDD RED-GREEN-REFACTOR cycle |
| `IDPF-Vibe` | Load `.min-mirror/IDPF-Vibe/Vibe-Core.md` — follow rapid iteration methodology |
| Not set / missing | Skip methodology dispatch — no framework enforced |
**If framework file not found:** "Warning: Framework {name} not found. Proceeding without methodology." Continue (non-blocking).
### Step 9: Work the Issue
Perform implementation work according to framework methodology.
### Step 10: Verify Acceptance Criteria
**IMPORTANT — Ground in file state:** Before evaluating each AC, re-read the actual file content using the Read tool. Do NOT evaluate from memory — re-read to confirm the criterion is met in current code. This prevents batch fatigue hallucination.
After work, verify each AC:
- **Can verify** → Mark `[x]`, continue
- **Cannot verify** (manual, external) → **STOP**, report to user, wait for disposition
Export and update issue body with checked criteria.
### Step 11: Move to in_review
```bash
gh pmu move $ISSUE --status in_review
```
### Step 12: STOP Boundary — Report and Wait
```
Issue #$ISSUE: $TITLE — In Review
Say "done" or run /done #$ISSUE to close this issue.
```
**STOP.** Wait for user to say "done". Do NOT close the issue.
**Epic sub-issue processing:** Each sub-issue follows Steps 9–12 individually. After user says "done" (which invokes `/done` to close), proceed to next sub-issue. Never batch-close or skip per-sub-issue STOP boundary.
---
## Error Handling
| Situation | Response |
|-----------|----------|
| Issue not found | "Issue #N not found. Check the issue number?" → STOP |
| No branch assignment | "Issue #N is not assigned to a branch. Run `/assign-branch #N` first." → STOP |
| `gh pmu` command fails | "Failed to update issue status: {error}" → STOP |
| PRD tracker not found | Continue silently (non-blocking) |
| Framework file missing | Warn and continue without methodology |
| No acceptance criteria | Report empty auto-TODO, continue |
| Issue already in_progress | Continue silently (idempotent) |
---
**End of /work Command**
