---
version: "v0.26.3"
description: Merge branch to main with gated checks
argument-hint: [--skip-gates] [--skip-e2e] [--dry-run]
---
<!-- EXTENSIBLE -->
# /merge-branch
Merge current branch to main with gated validation. Use when merging without version tag (e.g., feature branches). For versioned releases, use `/prepare-release`.
## Extension Points
| Point | Location | Purpose |
|-------|----------|---------|
| `pre-gate` | Before gates | Setup, environment |
| `gates` | Gate checks | **Custom validation** |
| `post-gate` | After gates | Pre-merge actions |
| `post-pr-create` | After PR creation | CI wait, PR validation |
| `post-merge` | After merge | Post-merge actions |
| `post-close` | After cleanup | Notifications |
---
## Arguments
| Argument | Description |
|----------|-------------|
| `--skip-gates` | Emergency bypass (use with caution) |
| `--dry-run` | Preview without changes |
---
## Pre-Checks
### Verify on Feature Branch
```bash
BRANCH=$(git branch --show-current)
```
Must NOT be on `main`.
### Check Tracker Issue
```bash
gh pmu release current --json tracker
```
Will be closed at end if exists.
---
<!-- USER-EXTENSION-START: pre-gate -->
<!-- USER-EXTENSION-END: pre-gate -->
## Phase 1: Gate Checks
**If `--skip-gates`, skip to Phase 2.**
### Default Gates
#### Gate 1.1: No Uncommitted Changes
```bash
git status --porcelain
```
**FAIL if not empty.**
#### Gate 1.2: Tests Pass
```bash
npm test 2>/dev/null || echo "No test script configured"
```
**FAIL if tests fail.**
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
### Gate Summary
- ✅ Passed / ❌ Failed (with details)
**If any fails, STOP.**
<!-- USER-EXTENSION-START: post-gate -->
<!-- USER-EXTENSION-END: post-gate -->
---
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
### Step 2.3: Wait for Approval
**ASK USER:** Review and approve PR.
```bash
gh pr view --json reviewDecision
```
#### Gate 2.4: PR Approved
**FAIL if not approved** (unless `--skip-gates`).
### Step 2.5: Merge PR
```bash
gh pr merge --merge
git checkout main
git pull origin main
```
<!-- USER-EXTENSION-START: post-merge -->
<!-- USER-EXTENSION-END: post-merge -->
---
## Phase 3: Cleanup
### Step 3.1: Close Tracker (if exists)
```bash
gh issue close [TRACKER_NUMBER] --comment "Branch merged to main"
```
### Step 3.2: Close Release (if exists)
```bash
gh pmu release close 2>/dev/null || echo "No release to close"
```
### Step 3.3: Delete Branch
```bash
git push origin --delete $BRANCH
git branch -d $BRANCH
```
<!-- USER-EXTENSION-START: post-close -->
<!-- USER-EXTENSION-END: post-close -->
---
## Completion
- ✅ Gates passed
- ✅ PR merged
- ✅ Tracker closed
- ✅ Branch deleted
---
## /merge-branch vs /prepare-release
| Feature | /merge-branch | /prepare-release |
|---------|---------------|------------------|
| Version bump | No | Yes |
| CHANGELOG | No | Yes |
| Git tag | No | Yes |
| GitHub Release | No | Yes |
**Use /merge-branch:** Feature/fix branches, non-versioned work.
**Use /prepare-release:** Versioned releases with CHANGELOG and tags.
---
**End of Merge Branch**
