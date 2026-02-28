---
version: "v0.53.1"
description: Start working on issues with validation and auto-TODO (project)
argument-hint: "#issue [#issue...] | all in <status>"
---

<!-- EXTENSIBLE -->
# /work
Start working on one or more issues. Validates issue existence, branch assignment, and issue type, then moves to `in_progress`, extracts auto-TODO, and dispatches to framework methodology.
**Extension Points:** See `.claude/metadata/extension-points.json` or run `/extensions list --command work`
---
## Prerequisites
- `gh pmu` extension installed
- `.gh-pmu.json` configured in repository root
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

<!-- USER-EXTENSION-START: pre-work -->
<!-- USER-EXTENSION-END: pre-work -->

### Step 1: Context Gathering (Preamble Script)
Run the preamble script to consolidate all deterministic setup work into a single invocation:
**Single issue:**
```bash
node .claude/scripts/shared/work-preamble.js --issue $ISSUE
```
**Multiple issues:**
```bash
node .claude/scripts/shared/work-preamble.js --issues "$ISSUE1,$ISSUE2,$ISSUE3"
```
**All in status:**
```bash
node .claude/scripts/shared/work-preamble.js --status $STATUS
```
Parse the JSON output and check `ok`:
- **If `ok: false`:** Report errors from `errors[]` array (each has `code` and `message`). If error has `suggestion`, include it. → **STOP**
- **If `ok: true`:** Extract `context`, `gates`, `autoTodo`, and `warnings` from the envelope.
**Report gate results from `gates`:**
- `movedToInProgress: true` → Issue moved to in_progress
- `prdTrackerMoved: true` → PRD tracker moved to in_progress
**Report any warnings** from `warnings[]` array (non-blocking).
**Extract context for subsequent steps:**
- `context.issue` — issue data (number, title, labels, body, state)
- `context.branch` — branch/status data
- `context.type` — `"epic"` or `"standard"`
- `context.tracker` — branch tracker number
- `context.framework` / `context.frameworkPath` — framework config
- `context.subIssues`, `context.skipped`, `context.processingOrder` — epic-only fields
**Extract autoTodo for Step 4:**
- Standard: `{ source: "acceptance_criteria", items: [{ text, checked }] }`
- Epic: `{ source: "sub_issues", items: [{ number, title }] }` — sub-issues sorted in ascending numeric order by default, or custom `**Processing Order:**` from epic body. Sub-issues already in `in_review` or `done` are skipped (listed in `context.skipped`).

<!-- USER-EXTENSION-START: post-work-start -->
<!-- USER-EXTENSION-END: post-work-start -->

### Step 2: QA Extraction — Manual Test AC Detection
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

### Step 3: Framework Methodology Dispatch
Read `framework-config.json` for `processFramework` and `frameworkPath` fields:
| Framework | Action |
|-----------|--------|
| `IDPF-Agile` | Load `{frameworkPath}/IDPF-Agile/Agile-Core.md` — follow TDD RED-GREEN-REFACTOR cycle |
| `IDPF-Vibe` | Load `{frameworkPath}/IDPF-Vibe/Vibe-Core.md` — follow rapid iteration methodology |
| Not set / missing | Skip methodology dispatch — no framework enforced |
**If framework file not found:** "Warning: Framework {name} not found. Proceeding without methodology." Continue (non-blocking).
### Step 4: Work the Issue
Iterate through the auto-TODO from Step 1. For each acceptance criterion:
1. **Mark TODO `in_progress`**
2. **Execute TDD cycle** (RED → GREEN → REFACTOR) per framework methodology
   - **RED:** Write a failing test for the target behavior, verify it fails
   - **GREEN:** Write minimal implementation to pass the test, verify it passes
   - **REFACTOR:** Analyze for duplication, naming, complexity, structure. Report decision (refactor or skip with reason). If refactoring, keep tests passing.
   - If no framework loaded, implement directly with appropriate testing
3. **Run full test suite** — confirm no regressions
4. **Mark TODO `completed`**
5. **Commit** with `Refs #$ISSUE — <brief description>`
**Commit granularity:** One commit per AC default. Closely related ACs may share a commit if separating creates broken intermediate states. Never combine unrelated ACs.
**If no auto-TODO was generated:** Work as a single unit and commit with `Refs #$ISSUE`.
**Post-compaction recovery:** Re-read this spec and check TODO list to determine completed vs pending ACs. Resume from first incomplete AC.
### Step 4b: Documentation Judgment
After implementation, evaluate whether documentation is warranted before verifying acceptance criteria.
**Assess each category:**
| Category | Warranted When | Target Path |
|----------|---------------|-------------|
| Design Decision | Chose between alternatives, non-obvious approach, or architectural trade-off | `Construction/Design-Decisions/YYYY-MM-DD-{topic}.md` |
| Tech Debt | Took a shortcut, deferred work, or introduced known limitations | `Construction/Tech-Debt/YYYY-MM-DD-{topic}.md` |
**If warranted:**
1. Determine category (may be both)
2. Check/create target directory
3. Derive `{topic}` from issue title (kebab-case, max 50 chars). If exists, append `-2`, `-3`, etc.
4. Create document using Design Decision or Tech Debt template
5. Commit: `Refs #$ISSUE — add {design decision|tech debt} documentation`
6. Check documentation AC checkbox with inline note: `[x] Design decisions documented — {topic}`
**If not warranted:** Check AC checkbox with note: `[x] No design decisions warranted — implementation was straightforward`

<!-- USER-EXTENSION-START: post-documentation -->
<!-- USER-EXTENSION-END: post-documentation -->

<!-- USER-EXTENSION-START: post-implementation -->
<!-- USER-EXTENSION-END: post-implementation -->

### Step 5: Verify Acceptance Criteria
**IMPORTANT — Ground in file state:** Before evaluating each AC, re-read the actual file content using the Read tool. Do NOT evaluate from memory — re-read to confirm the criterion is met in current code. This prevents batch fatigue hallucination.
For each AC checkbox in the issue body:
- **Can verify** → Mark `[x]`, continue
- **Cannot verify** (manual, external) → **STOP**, present options, wait for user disposition
After all ACs resolved, export and update issue body:
```bash
gh pmu view $ISSUE --body-stdout > .tmp-$ISSUE.md
# Update checkboxes to [x]
gh pmu edit $ISSUE -F .tmp-$ISSUE.md && rm .tmp-$ISSUE.md
```

<!-- USER-EXTENSION-START: post-ac-verification -->
<!-- USER-EXTENSION-END: post-ac-verification -->

### Step 6: Move to in_review
```bash
gh pmu move $ISSUE --status in_review
```
### Step 7: STOP Boundary — Report and Wait
```
Issue #$ISSUE: $TITLE — In Review
Say "done" or run /done #$ISSUE to close this issue.
```
**STOP.** Wait for user to say "done". Do NOT close the issue.
**CRITICAL — Autonomous Epic Sub-Issue Processing:**
When working an epic, process sub-issues autonomously in the order determined by Step 1 (`context.processingOrder` from preamble output). Sub-issues already in `in_review` or `done` are skipped (`context.skipped`).
For each sub-issue in processing order:
1. Move sub-issue to `in_progress`
2. Work through TDD cycles (Steps 4–5)
3. Move sub-issue to `in_review`
4. **STOP** — report and wait for user to say "done"
5. After user says "done" (invokes `/done` to move sub-issue to done), proceed to next sub-issue

<!-- USER-EXTENSION-START: post-sub-issue-done -->
<!-- USER-EXTENSION-END: post-sub-issue-done -->

**After all sub-issues reach `in_review` or `done`:**
1. Evaluate the epic's own acceptance criteria (Step 5)
2. Move epic to `in_review`
3. **STOP** — report epic status and wait for user to say "done"
**Never batch-close sub-issues or skip the per-sub-issue STOP boundary.** Sequential means work order, not bypassing STOP boundaries.
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
