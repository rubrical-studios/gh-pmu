---
version: "v0.46.1"
description: Complete issues with criteria verification and status transitions (project)
argument-hint: "[#issue...] [--no-docs] (optional)"
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
- Issue in `in_review` status (use `/work` first)
---
## Arguments
| Argument | Required | Description |
|----------|----------|-------------|
| `#issue` | No | Single issue number (e.g., `#42` or `42`) |
| `#issue #issue...` | | Multiple issue numbers (e.g., `#42 #43 #44`) |
| *(none)* | | Queries `in_review` issues for selection |
| `--no-docs` | No | Skip design decisions documentation offer (Step 5) |
---
## Execution Instructions
**REQUIRED:** Before executing:
1. **Generate Todo List:** Parse workflow steps, use `TodoWrite` to create todos
2. **Include Extensions:** Add todo for each non-empty `USER-EXTENSION` block
3. **Track Progress:** Mark todos `in_progress` → `completed` as you work
4. **Post-Compaction:** Re-read spec and regenerate todos after context compaction
---
## Workflow
### Step 1: Parse Arguments
Formats: `#42` (single), `#42 #43 #44` (multiple), *(none)* (query)
**For no arguments:**
```bash
gh pmu list --status in_review
```
| Result | Action |
|--------|--------|
| No issues found | "No issues in review. Specify issue number or complete work first." → **STOP** |
| Single issue | Confirm: `Issue #N: $TITLE is in review. Complete it? (yes/no)` |
| Multiple issues | Present numbered list for user selection |
**Multiple issues:** Process each sequentially using Steps 2-5.
### Step 2: Validate Issue
```bash
gh issue view $ISSUE --json number,title,labels,body,state
```
**Not found:** "Error: Issue #$ISSUE not found." → **STOP** (continue to next if batch)
**Already closed:** "Issue #$ISSUE is already closed. Skipping." → skip or **STOP**
### Step 2b: PRD Label Redirect
Check labels from Step 2 for `prd` label.
**If `prd` label:** "Use `/complete-prd #$ISSUE` to close PRD trackers." → **STOP** (skip to next if batch)
**No `prd` label:** Continue to Step 3.
### Step 3: Detect Current Status
```bash
gh pmu view $ISSUE --json=status
```
| Status | Path |
|--------|------|
| `in_review` | Proceed to Step 3b |
| `in_progress` | "Issue #$ISSUE is still in progress. Complete work first via /work." → **STOP** |
| `done` | Already done — report and skip |
| Other | "Issue #$ISSUE is in $STATUS. Move to in_progress first via /work." → **STOP** |

<!-- USER-EXTENSION-START: pre-done -->
<!-- USER-EXTENSION-END: pre-done -->

### Step 3b: Diff Verification
```bash
node .claude/scripts/shared/done-verify.js --issue $ISSUE
```
Parse JSON output:
- **No commits found:** "Warning: No commits reference #$ISSUE." → **STOP.** Wait for user.
- **Warnings present:** List concerns, ask "Continue? (yes/no)" → **STOP.** Wait for user.
- **New files detected:** Report informational (not blocking): `New files: [list]`
- **Clean:** `Diff verified: N commits, M files changed (all substantive)` → Proceed to Step 4.
### Step 4: Move to Done
```bash
gh pmu move $ISSUE --status done
```
Report: `Issue #$ISSUE: $TITLE → Done`

<!-- USER-EXTENSION-START: post-done -->
<!-- USER-EXTENSION-END: post-done -->

### Step 5: Design Decisions Offer
**If `--no-docs`:** Skip this step.
**If the ASSISTANT** is aware of decisions required to complete work, offer:
```
Would you like me to document the design decisions/issues encountered in Construction/Design-Decisions/?
```
**If accepted:**
1. Check/create `Construction/Design-Decisions/`
2. Derive `{topic}` from issue title (kebab-case, max 50 chars). If exists, append `-2`, `-3`, etc.
3. Create `Construction/Design-Decisions/YYYY-MM-DD-{topic}.md`:
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
**If declined:** Proceed without documenting.
### Step 6: Git Add, Commit and Push
**Conditional:** Only commit if changes exist. Check `git status --porcelain`.
**If output is empty:** Push unpushed commits: `git push`. Report: `No new changes to commit. Pushed.`
**If output is non-empty:** Stage, `git commit -m "docs: add design decision for #$ISSUE"`, and push.
### Step 6b: Background CI Monitoring
After push:
1. Get SHA: `sha=$(git rev-parse HEAD)`
2. **Pre-check paths-ignore:** Use `ci-watch.js`'s `shouldSkipMonitoring()`. If all files match → skip, report: `"CI skipped (paths-ignore)"`
3. **Spawn background:** Bash with `run_in_background: true`:
   ```bash
   node .claude/scripts/shared/ci-watch.js --sha $SHA --timeout 300
   ```
4. Report: `"CI monitoring started in background."`
**Exit codes:**
| Code | Report |
|------|--------|
| 0 | `"CI passed for #$ISSUE (duration)"` |
| 1 | `"CI FAILED. Failed step: \"step-name\". Run: gh run view <id> --log-failed"` |
| 2 | `"CI still running after 5m. Check: gh run list --commit $SHA"` |
| 3 | `"No CI run triggered (paths-ignore likely)"` |
| 4 | `"CI cancelled (superseded by newer push)"` |
**Multiple workflows:** Report per-workflow from `workflows` array.
### Step 7: Cleanup
**MUST DO:** Clear todo list.
---
## Error Handling
| Situation | Response |
|-----------|----------|
| Issue not found | "Issue #N not found." → STOP |
| Issue already closed | "Issue #N is already closed." → skip |
| Issue still in_progress | "Complete work first via /work." → STOP |
| Issue in other status | "Move to in_progress first via /work." → STOP |
| No issues in review | "No issues in review." → STOP |
| `gh pmu` fails | "Failed to update issue: {error}" → STOP |
| Construction/ missing | Warn and create |
---
**End of /done Command**
