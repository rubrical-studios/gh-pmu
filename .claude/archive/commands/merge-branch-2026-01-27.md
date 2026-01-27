<!-- Archived: 2026-01-27T22:10:59.208Z -->
<!-- Original: E:\Projects\gh-pmu\.claude\commands\merge-branch.md -->
<!-- Reason: Rogue edits detected outside USER-EXTENSION blocks -->

---
version: "v0.31.0"
description: Merge branch to main with gated checks (project)
argument-hint: [--skip-gates] [--dry-run]
---
<!-- EXTENSIBLE -->
# /merge-branch
Merge the current branch to main with gated validation checks. For versioned releases with tagging, use `/prepare-release` instead.
## Available Extension Points
| Point | Location | Purpose |
|-------|----------|---------|
| `pre-gate` | Before gates | Environment preparation |
| `gates` | Gate checks | **Custom validation gates** |
| `post-gate` | After gates pass | Pre-merge actions |
| `post-pr-create` | After PR creation | CI wait, PR validation |
| `post-merge` | After PR merged | Post-merge actions |
| `post-close` | After cleanup | Notifications |
## Arguments
| Argument | Description |
|----------|-------------|
| `--skip-gates` | Emergency bypass - skip all gates |
| `--dry-run` | Preview actions without executing |
## Execution Instructions
**REQUIRED:** Before executing:
1. **Create Todo List:** Use `TodoWrite` to create todos from steps below
2. **Track Progress:** Mark todos `in_progress` → `completed`
3. **Resume Point:** If interrupted, todos show where to continue
## Pre-Checks
### Verify on Feature Branch
```bash
BRANCH=$(git branch --show-current)
```
Must NOT be on `main`. Typical: `feature/*`, `fix/*`, `idpf/*`, `patch/*`, `release/*`.
### Check for Tracker Issue
```bash
gh pmu branch current --json tracker
```

<!-- USER-EXTENSION-START: pre-gate -->
<!-- USER-EXTENSION-END: pre-gate -->

## Phase 1: Gate Checks
**If `--skip-gates`, skip to Phase 2.**
### Default Gates
#### Gate 1.1: No Uncommitted Changes
```bash
git status --porcelain
```
**FAIL if output not empty.**
#### Gate 1.2: Tests Pass
```bash
npm test 2>/dev/null || echo "No test script configured"
```

<!-- USER-EXTENSION-START: gates -->
#### Gate 1.3: E2E Tests Pass

**If `--skip-e2e` was passed, skip this gate.**

```bash
node .claude/scripts/e2e/run-e2e-gate.js
```

The script outputs JSON: `{"success": true/false, "testsRun": N, "testsPassed": N, "duration": N}`

**FAIL if `success` is false.**

E2E tests validate complete workflows against the test project.
<!-- USER-EXTENSION-END: gates -->

**If any gate fails, STOP and report.**

<!-- USER-EXTENSION-START: post-gate -->
<!-- USER-EXTENSION-END: post-gate -->

## Phase 2: Create and Merge PR
### Step 2.1: Push Branch
```bash
git push origin $(git branch --show-current)
```
### Step 2.2: Create PR
```bash
gh pr create --base main --head $(git branch --show-current) \
  --title "Merge: $(git branch --show-current)"
```

<!-- USER-EXTENSION-START: post-pr-create -->

### Wait for CI
```bash
node .claude/scripts/framework/wait-for-ci.js
```
<!-- USER-EXTENSION-END: post-pr-create -->

### Step 2.3: Wait for PR Approval
**ASK USER:** Review and approve the PR.
### Step 2.4: PR Approved Gate
**FAIL if not approved** (unless `--skip-gates`).
### Step 2.5: Merge PR
```bash
gh pr merge --merge
git checkout main
git pull origin main
```

<!-- USER-EXTENSION-START: post-merge -->
<!-- USER-EXTENSION-END: post-merge -->

## Phase 3: Cleanup
### Step 3.1: Close Tracker Issue
```bash
gh issue close [TRACKER_NUMBER] --comment "Branch merged to main"
```
### Step 3.2: Close Branch in Project
```bash
gh pmu branch close 2>/dev/null || echo "No branch to close"
```
### Step 3.3: Delete Branch
```bash
git push origin --delete $BRANCH
git branch -d $BRANCH
```

<!-- USER-EXTENSION-START: post-close -->
<!-- USER-EXTENSION-END: post-close -->

## Completion
- ✅ All gates passed
- ✅ PR created and merged
- ✅ Tracker issue closed
- ✅ Branch deleted
## Comparison: /merge-branch vs /prepare-release
| Feature | /merge-branch | /prepare-release |
|---------|---------------|------------------|
| Version bump | No | Yes |
| CHANGELOG | No | Yes |
| Git tag | No | Yes |
| Gates | Yes | Yes |
| PR to main | Yes | Yes |
**Use `/merge-branch` for:** Feature/fix branches, non-versioned work.
**Use `/prepare-release` for:** Versioned releases with CHANGELOG and tags.
**End of Merge Branch**
