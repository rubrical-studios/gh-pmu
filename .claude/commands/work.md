---
version: "v0.49.1"
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
### Step 7b: QA Extraction — Manual Test AC Detection
Scan AC for manual test indicators. If found, offer to extract into tracked QA sub-issues.
**Detection keywords** (case-insensitive): `manually verify`, `visually confirm`, `visual verification`, `QA:`, `qa-required`, `exploratory test`, `manual test`, `manual check`, `UX walkthrough`
**If manual test AC detected:**
1. **Present candidates** — Use `AskUserQuestion` with multiSelect. Include "Skip all" option.
2. **Create QA sub-issues** — For each confirmed AC:
   ```bash
   gh pmu sub create --parent $ISSUE --title "QA: [AC description]" --label qa-required -F .tmp-qa-body.md
   ```
   QA sub-issue body contains: test description, parent issue context (`Parent: #$ISSUE — $TITLE`), steps to perform, expected result.
3. **Annotate parent AC** — Update parent issue body with QA sub-issue reference. AC remains **unchecked**:
   `- [ ] Manually verify the login flow → QA: #NNN`
4. **Report** extraction count and sub-issue numbers.
**Closure path:** Parent stays `in_review` until all QA sub-issues are closed and AC checked off. Only then can `/done` complete the parent.
**If no manual test AC detected:** Continue silently.
**If user selects "Skip all":** Continue without extraction.

<!-- USER-EXTENSION-START: pre-framework-dispatch -->
<!-- USER-EXTENSION-END: pre-framework-dispatch -->

### Step 8: Framework Methodology Dispatch
Read `framework-config.json` for `processFramework` and `frameworkPath` fields:
| Framework | Action |
|-----------|--------|
| `IDPF-Agile` | Load `{frameworkPath}/IDPF-Agile/Agile-Core.md` — follow TDD RED-GREEN-REFACTOR cycle |
| `IDPF-Vibe` | Load `{frameworkPath}/IDPF-Vibe/Vibe-Core.md` — follow rapid iteration methodology |
| Not set / missing | Skip methodology dispatch — no framework enforced |
**If framework file not found:** "Warning: Framework {name} not found. Proceeding without methodology." Continue (non-blocking).
### Step 9: Work the Issue
Iterate through the auto-TODO from Step 7. For each acceptance criterion:
1. **Mark TODO `in_progress`**
2. **Execute TDD cycle** (RED → GREEN → REFACTOR) per framework methodology
   - **RED:** Write failing test, verify it fails
   - **GREEN:** Write minimal implementation to pass, verify it passes
   - **REFACTOR:** Analyze for duplication, naming, complexity, structure. Report decision (refactor or skip with reason). If refactoring, keep tests passing.
   - If no framework loaded, implement directly with appropriate testing
3. **Run full test suite** — confirm no regressions
4. **Mark TODO `completed`**
5. **Commit** with `Refs #$ISSUE — <brief description>`
**Commit granularity:** One commit per AC is the default. Closely related ACs may share a commit if separating them would create broken intermediate states. Never combine unrelated ACs.
**If no auto-TODO was generated:** Work as a single unit and commit with `Refs #$ISSUE`.
**Post-compaction recovery:** Re-read this spec and check TODO list to determine completed vs pending ACs. Resume from first incomplete AC.
### Step 9b: Documentation Judgment
After implementation, evaluate whether documentation is warranted before verifying acceptance criteria.
**Assess each category:**
| Category | Warranted When | Target Path |
|----------|---------------|-------------|
| Design Decision | Chose between alternatives, non-obvious approach, or architectural trade-off | `Construction/Design-Decisions/YYYY-MM-DD-{topic}.md` |
| Tech Debt | Took a shortcut, deferred work, or introduced known limitations | `Construction/Tech-Debt/YYYY-MM-DD-{topic}.md` |
**If warranted:**
1. Check/create target directory
2. Derive `{topic}` from issue title (kebab-case, max 50 chars). If exists, append `-2`, `-3`, etc.
3. Create document using Design Decision or Tech Debt template
4. Commit: `Refs #$ISSUE — add {design decision|tech debt} documentation`
5. Check documentation AC checkbox with inline note: `[x] Design decisions documented — {topic}`
**If not warranted:** Check AC checkbox with note: `[x] No design decisions warranted — implementation was straightforward`
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
