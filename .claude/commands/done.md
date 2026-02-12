---
version: "v0.42.0"
description: Complete issues with criteria verification and status transitions (project)
argument-hint: "[#issue...] (optional)"
---

<!-- EXTENSIBLE -->
# /done
Complete one or more issues. Moves from `in_review` → `done` with a STOP boundary. Only handles the final transition — `/work` owns `in_progress` → `in_review`.
## Available Extension Points
| Point | Location | Purpose |
|-------|----------|---------|
| `pre-done` | Before moving to done | Tests, lint, build verification |
| `post-done` | After moving to done | Deploy trigger, time tracking stop |
---
## Prerequisites
- `gh pmu` extension installed
- `.gh-pmu.yml` configured in repository root
- Issue in `in_review` status (use `/work` to move issues through `in_progress` → `in_review` first)
---
## Arguments
| Argument | Required | Description |
|----------|----------|-------------|
| `#issue` | No | Single issue number (e.g., `#42` or `42`) |
| `#issue #issue...` | | Multiple issue numbers (e.g., `#42 #43 #44`) |
| *(none)* | | Queries `in_review` issues for selection |
---
## Execution Instructions
**REQUIRED:** Before executing:
1. **ALWAYS Generate Todo List:** Parse workflow steps, use `TodoWrite` to create todos
2. **Include Extensions:** Add todo item for each non-empty `USER-EXTENSION` block
3. **Track Progress:** Mark todos `in_progress` → `completed` as you work
4. **Post-Compaction:** Re-read spec and regenerate todos after context compaction
**Todo Rules:** One todo per numbered step; one todo per active extension; skip commented-out extensions.
---
## Workflow
### Step 1: Parse Arguments
Accept these formats:
- `#42` or `42` — single issue
- `#42 #43 #44` — multiple issues
- *(no arguments)* — query for in_review issues
**For no arguments:**
```bash
gh pmu list --status in_review
```
| Result | Action |
|--------|--------|
| No issues found | "No issues in review. Specify issue number or complete work first." → **STOP** |
| Single issue | Confirm: `Issue #N: $TITLE is in review. Complete it? (yes/no)` |
| Multiple issues | Present numbered list for user selection |
**For multiple issues:** Process each sequentially using Steps 2–5.
### Step 2: Validate Issue
```bash
gh issue view $ISSUE --json number,title,labels,body,state
```
**If not found:** "Error: Issue #$ISSUE not found." → **STOP** (skip this issue, continue to next if batch)
**If already closed:** "Issue #$ISSUE is already closed. Skipping." → Skip to next issue or **STOP**
### Step 3: Detect Current Status
```bash
gh pmu view $ISSUE --json=status
```
| Status | Path |
|--------|------|
| `in_review` | Proceed to Step 3b |
| `in_progress` | Report: `Issue #$ISSUE is still in progress. Complete work first via /work.` → **STOP** |
| `done` | Already done — report and skip |
| Other | "Issue #$ISSUE is in $STATUS. Move to in_progress first via /work." → **STOP** |
**Note:** `/done` only handles the `in_review → done` transition. The `in_progress → in_review` transition is owned by `/work` (which verifies acceptance criteria and moves to in_review as part of its workflow).

<!-- USER-EXTENSION-START: pre-done -->
<!-- USER-EXTENSION-END: pre-done -->

### Step 3b: Diff Verification
Run the verification helper to detect hallucinated completions:
```bash
node .claude/scripts/shared/done-verify.js --issue $ISSUE
```
**If no commits found** (warnings contains "No commits found"):
```
Warning: No commits reference #$ISSUE. Was work committed with the issue reference?
```
→ **STOP.** Wait for user.
**If warnings present** (comment-only files, EOF-only appends):
```
Diff verification found concerns:
  [list each warning]
Review these before closing. Continue? (yes/no)
```
→ **STOP.** Wait for user confirmation.
**If new files detected** (newFiles array non-empty):
Report informational (not blocking): `New files: [list files]`
New files are expected to be all-insertions — NOT warnings.
**If clean** (no warnings, all substantive):
```
Diff verified: N commits, M files changed (all substantive) ✓
```
Include new file count if any: `(N new files, M modified)`
→ Proceed to Step 4.
### Step 4: Move to Done
```bash
gh pmu move $ISSUE --status done
```
Report: `Issue #$ISSUE: $TITLE → Done ✓`

<!-- USER-EXTENSION-START: post-done -->
<!-- USER-EXTENSION-END: post-done -->

### Step 5: Design Decisions Offer
**If the ASSISTANT** is aware of decisions required to complete work on this issue, offer to document:
```
Would you like me to document the design decisions/issues encountered in Construction/Design-Decisions/?
```
**If accepted:**
1. Check if `Construction/Design-Decisions/` exists; create if missing with warning
2. Derive `{topic}` from issue title (kebab-case, truncate to 50 chars). If file exists, append `-2`, `-3`, etc.
3. Create `Construction/Design-Decisions/YYYY-MM-DD-{topic}.md` using template:
```markdown
# Design Decision: [Title]
**Date:** [YYYY-MM-DD]
**Status:** Accepted
**Context:** Issue #[N] — [Issue Title]
## Decision
[What was decided]
## Rationale
[Why this choice was made]
## Alternatives Considered
- [Alternative 1]: [Why rejected]
## Consequences
- [Positive consequence]
- [Negative consequence or trade-off]
## Issues Encountered
[Any blockers, surprises, or lessons learned]
```
4. Reference the issue number in the document
**If declined:** Proceed without documenting (optional).
### Step 6. Git add, commit and push
Perform the steps to commit and push changes associated with this issue.
### Step 6b. Background CI Monitoring
After push completes, spawn a background CI monitor:
1. Get commit SHA: `sha=$(git rev-parse HEAD)`
2. **Pre-check paths-ignore:** Inspect changed files via `git diff --name-only HEAD~1` against workflow `paths-ignore` patterns from `.github/workflows/test.yml`. Use `ci-watch.js`'s `shouldSkipMonitoring()` function. If all changed files match paths-ignore → skip monitoring, report: `"CI skipped (paths-ignore — only ignored files changed)"`
3. **Spawn background agent:** Use Bash tool with `run_in_background: true`:
   ```bash
   node .claude/scripts/shared/ci-watch.js --sha $SHA --timeout 300
   ```
4. Report: `"CI monitoring started in background. You'll be notified when complete."`
**When background agent returns**, report based on exit code:
- **Exit 0 (pass):** `"CI passed for #$ISSUE (duration) ✓"`
- **Exit 1 (fail):** `"CI FAILED for #$ISSUE. Failed step: \"step-name\". Run: gh run view <id> --log-failed"`
- **Exit 2 (timeout):** `"CI still running after 5m. Check manually: gh run list --commit $SHA"`
- **Exit 3 (no run):** `"No CI run triggered for $SHA (paths-ignore likely applies)"`
- **Exit 4 (cancelled):** `"CI cancelled for #$ISSUE (superseded by newer push)"`
**Multiple workflows:** If the JSON output contains a `workflows` array, report per-workflow results.
---
## Error Handling
| Situation | Response |
|-----------|----------|
| Issue not found | "Issue #N not found." → STOP |
| Issue already closed | "Issue #N is already closed. Skipping." → skip |
| Issue still in_progress | "Issue #N is still in progress. Complete work first via /work." → STOP |
| Issue in other status | "Issue #N is in $STATUS. Move to in_progress first via /work." → STOP |
| No issues in review (no-arg) | "No issues in review. Specify issue number." → STOP |
| `gh pmu` command fails | "Failed to update issue: {error}" → STOP |
| Construction/ missing | Warn and create |
---
**End of /done Command**
