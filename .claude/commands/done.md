---
version: "v0.54.0"
description: Complete issues with criteria verification and status transitions (project)
argument-hint: "[#issue...] (optional)"
---

<!-- EXTENSIBLE -->
# /done
Complete one or more issues. Moves from `in_review` → `done` with a STOP boundary. Only handles the final transition — `/work` owns `in_progress` → `in_review`.
**Extension Points:** See `.claude/metadata/extension-points.json` or run `/extensions list --command done`
---
## Prerequisites
- `gh pmu` extension installed
- `.gh-pmu.json` configured in repository root
- Issue in `in_review` status (use `/work` first)
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
1. **Generate Todo List:** Parse workflow steps, use `TodoWrite` to create todos
2. **Include Extensions:** Add todo for each non-empty `USER-EXTENSION` block
3. **Track Progress:** Mark todos `in_progress` → `completed` as you work
4. **Post-Compaction:** Re-read spec and regenerate todos after context compaction
---
## Workflow
### Step 1: Context Gathering (Preamble Script)
Run the preamble script for validation, diff verification, status transition, tracker linking, and CI pre-check:
**Single issue:**
```bash
node .claude/scripts/shared/done-preamble.js --issue $ISSUE
```
**Multiple issues:**
```bash
node .claude/scripts/shared/done-preamble.js --issues "$ISSUE1,$ISSUE2"
```
**No arguments (discovery mode):**
```bash
node .claude/scripts/shared/done-preamble.js
```
Parse JSON output and check `ok`:
- **If `ok: false`:** Report errors from `errors[]` (each has `code`, `message`). If error has `suggestion`, include it. → **STOP**
- **If discovery mode** (`discovery` field present): Present `discovery.issues` for user selection. After selection, re-run with `--issue N`.
**If `ok: true` with `diffVerification`:**
- `requiresConfirmation: true` → Report `diffVerification.warnings`, ask "Continue? (yes/no)". If yes, re-run with `--force-move`. If no → **STOP**.
- `requiresConfirmation: false` → Issue already moved to done. Proceed.
**If `ok: true` with `gates.movedToDone: true`:**
- Report: `Issue #$ISSUE: $TITLE → Done`
- If `context.trackerLinked: true`: Report `Linked #$ISSUE to branch tracker #$TRACKER`
- If `context.nextSteps` present: Report `context.nextSteps.guidance` (approval-gate next steps)
**Report any warnings** from `warnings[]` (non-blocking).
**Multiple issues:** Process each sequentially using Steps 1-4.

<!-- USER-EXTENSION-START: pre-done -->
<!-- USER-EXTENSION-END: pre-done -->

<!-- USER-EXTENSION-START: post-done -->
<!-- USER-EXTENSION-END: post-done -->

### Step 2: Push
```bash
git push
```
Report: `Pushed.`
### Step 3: Background CI Monitoring
After push:
1. Get SHA: `sha=$(git rev-parse HEAD)`
2. **Check `context.ci.hasPushWorkflows`** from preamble output:
   - If `false`, skip and report: `"CI skipped (no push-triggered workflows)"`
   - If `true`, continue to step 3.
3. **Pre-check paths-ignore:** Obtain `changedFiles` via `git diff --name-only HEAD~1` and `pathsIgnore` from workflow YAML. If all files match, skip and report: `"CI skipped (paths-ignore)"`
4. **Spawn background:** Bash with `run_in_background: true`:
   ```bash
   node ./.claude/scripts/shared/ci-watch.js --sha $SHA --timeout 300
   ```
5. Report: `"CI monitoring started in background."`
**Exit codes:**
| Code | Report |
|------|--------|
| 0 | `"CI passed for #$ISSUE (duration)"` |
| 1 | `"CI FAILED. Failed step: \"step-name\". Run: gh run view <id> --log-failed"` |
| 2 | `"CI still running after 5m. Check: gh run list --commit $SHA"` |
| 3 | `"No CI run triggered (paths-ignore likely)"` |
| 4 | `"CI cancelled (superseded by newer push)"` |
**Multiple workflows:** Report per-workflow from `workflows` array.
### Step 4: Cleanup
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
---
**End of /done Command**
