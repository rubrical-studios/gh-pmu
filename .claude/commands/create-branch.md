---
version: "v0.44.0"
description: Create a branch with tracker issue (project)
argument-hint: "<branch-name> (e.g., release/v0.16.0, my-feature, bugfix-123)"
---

<!-- EXTENSIBLE -->
# /create-branch
Creates a new branch and associated tracker issue for any branch type.
**Extension Points:** See `.claude/metadata/extension-points.json` or run `/extensions list --command create-branch`
---
## Prerequisites
- `gh pmu` extension installed
- `.gh-pmu.yml` configured in repository root
---
## Arguments
| Argument | Required | Description |
|----------|----------|-------------|
| `$1` | Yes | Branch name (any valid git branch name) |
---
## Workflow
### Step 1: Validate Arguments
Branch name must be valid git branch name (no spaces, no special characters git rejects).
If invalid or empty, report error and stop.
### Step 2: Check Working Directory
```bash
git status --porcelain
```
**If changes exist:**
1. Report: "Uncommitted changes detected. These will be carried to the new branch."
2. Save output for Step 6 reporting
3. Continue with branch creation (do NOT block)

<!-- USER-EXTENSION-START: pre-create -->
### Verify Config
```bash
git status --porcelain .gh-pmu.yml
```
**If `success` is false, STOP.**
<!-- USER-EXTENSION-END: pre-create -->

### Step 3: Create Branch with Tracker
```bash
gh pmu branch start --name "$BRANCH"
```
Creates git branch `$BRANCH` and tracker issue with `branch` label.
### Step 4: Switch to Branch
```bash
git checkout "$BRANCH"
```
### Step 5: Push Branch to Remote
```bash
git push -u origin "$BRANCH"
```
### Step 5.5: Set Active Label
```bash
node .claude/scripts/shared/lib/active-label.js ensure [TRACKER_NUMBER]
```

<!-- USER-EXTENSION-START: post-create -->
<!-- USER-EXTENSION-END: post-create -->

### Step 6: Report Completion
```
Branch created.

Branch: $BRANCH
Tracker: #[tracker-issue-number]
```
**If uncommitted changes detected in Step 2:**
Report carried-over files from saved `git status --porcelain` output.
**Conditional Commit Prompt:**
If any changes exist (staged, unstaged, or untracked):
**ASK USER:** "Stage and commit all changes to new branch? (y/n)"
- **Yes:** Request commit message, run `git add -A && git commit -m "<message>"`, report success
- **No:** Continue without modifying working tree
**Always end with:**
```
Next steps:
1. Assign issues: /assign-branch #N #N ...
2. Work issues: work #N
3. When ready: /prepare-release
```
---
**End of Create Branch**
