---
version: "v0.23.3"
description: Create a branch with tracker issue
argument-hint: <prefix/identifier> (e.g., release/v1.2.0, feature/new-auth)
---
<!-- EXTENSIBLE: v0.17.0 -->
# /create-branch
Creates branch and tracker issue for any branch type (release, patch, feature, hotfix).
## Extension Points
| Point | Location | Purpose |
|-------|----------|---------|
| `pre-create` | Before branch | Custom validation |
| `post-create` | After branch | Notifications, CI |
| `checklist` | Summary | Extension verification |
---
## Prerequisites
- `gh pmu` installed, `.gh-pmu.yml` configured, clean working directory
---
## Workflow
### Step 1: Validate Arguments
Branch must follow `[prefix]/[name]`: exactly one `/`, both parts non-empty.
**Valid:** `release/v1.2.0`, `patch/v1.9.1`, `idpf/domain-reorg`
**Invalid:** `v1.2.0`, `release/`, `a/b/c`
### Step 2: Check Working Directory
```bash
git status --porcelain
```
<!-- USER-EXTENSION-START: pre-create -->
### Verify Config
```bash
node .claude/scripts/create-branch/verify-config.js
```
**If `success` is false, STOP.**
<!-- USER-EXTENSION-END: pre-create -->
### Step 3: Create Branch with Tracker
```bash
gh pmu release start --branch "$BRANCH"
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
## Summary Checklist
**Core:**
- [ ] Branch name validated
- [ ] Working directory clean
- [ ] Branch created via `gh pmu release start`
- [ ] Switched to branch
- [ ] Branch pushed
- [ ] Directory created
<!-- USER-EXTENSION-START: checklist -->
- [ ] Config verified
<!-- USER-EXTENSION-END: checklist -->
---
**End of Create Branch**
