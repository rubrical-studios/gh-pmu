---
version: "v0.29.2"
description: Create a branch with tracker issue
argument-hint: <branch-name> (e.g., release/v0.16.0, my-feature, bugfix-123)
---
<!-- EXTENSIBLE -->
# /create-branch
Creates branch and tracker issue for any branch type (release, patch, feature, hotfix).
## Extension Points
| Point | Location | Purpose |
|-------|----------|---------|
| `pre-create` | Before branch | Custom validation |
| `post-create` | After branch | Notifications, CI |
---
## Prerequisites
- `gh pmu` installed, `.gh-pmu.yml` configured, clean working directory
---
## Workflow
### Step 1: Validate Arguments
Branch name must be a valid git branch name (no spaces, no special characters that git rejects).
### Step 2: Check Working Directory
```bash
git status --porcelain
```

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
Creates git branch and tracker issue with `branch` label.
### Step 4: Switch to Branch
```bash
git checkout "$BRANCH"
```
### Step 5: Push Branch
```bash
git push -u origin "$BRANCH"
```
### Step 6: Create Branch Directory
```bash
mkdir -p "Releases/$BRANCH"
```

<!-- USER-EXTENSION-START: post-create -->
<!-- USER-EXTENSION-END: post-create -->

### Step 7: Report Completion
```
Branch created.
Branch: $BRANCH
Tracker: #[issue-number]
Directory: Releases/$BRANCH/
Next: 1. /assign-branch #N #N ...  2. work #N  3. /prepare-release
```
---
**End of Create Branch**
