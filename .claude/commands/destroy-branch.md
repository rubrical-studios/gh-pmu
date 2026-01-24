---
version: "v0.31.0"
description: Safely delete branch with confirmation
argument-hint: [branch-name] [--force]
---
<!-- EXTENSIBLE -->
# /destroy-branch
Safely abandon and delete a branch. **Destructive operation requiring confirmation.**
## Extension Points
| Point | Location | Purpose |
|-------|----------|---------|
| `pre-destroy` | Before confirmation | Pre-destruction validation |
| `post-confirm` | After confirmation | Actions before deletion |
| `post-destroy` | After deletion | Cleanup, notifications |
---
## Arguments
| Argument | Description |
|----------|-------------|
| `[branch-name]` | Branch to destroy (defaults to current) |
| `--force` | Skip confirmation (dangerous) |
---
## Pre-Checks
### Identify Target Branch
```bash
BRANCH=${1:-$(git branch --show-current)}
```
### Cannot Destroy Main
```bash
if [ "$BRANCH" = "main" ] || [ "$BRANCH" = "master" ]; then
  echo "ERROR: Cannot destroy main/master branch"
  exit 1
fi
```
### Check Branch Exists
```bash
git rev-parse --verify "$BRANCH" 2>/dev/null
```
**FAIL if branch does not exist.**

---

<!-- USER-EXTENSION-START: pre-destroy -->
<!-- USER-EXTENSION-END: pre-destroy -->

## Phase 1: Confirmation
**⚠️ DESTRUCTIVE OPERATION**
Will permanently delete:
- Local branch: `$BRANCH`
- Remote branch: `origin/$BRANCH`
- Release artifacts: `Releases/[prefix]/[identifier]/`
- Tracker issue (closed as "not planned")
### Step 1.1: Show What Will Be Destroyed
```bash
git log main..$BRANCH --oneline 2>/dev/null || echo "No unmerged commits"
ls -la Releases/*/$BRANCH/ 2>/dev/null || echo "No release artifacts found"
```
### Step 1.2: Require Explicit Confirmation
**If `--force` NOT passed:**
**ASK USER:** Type the full branch name to confirm destruction.
User must type exactly: `$BRANCH`
**If input does not match, ABORT.**

<!-- USER-EXTENSION-START: post-confirm -->
<!-- USER-EXTENSION-END: post-confirm -->

---
## Phase 2: Close Tracker
### Step 2.1: Find Tracker Issue
```bash
gh pmu branch current --json tracker 2>/dev/null
```
### Step 2.2: Close as Not Planned
```bash
gh issue close [TRACKER_NUMBER] \
  --reason "not planned" \
  --comment "Branch destroyed via /destroy-branch. Work abandoned."
```
### Step 2.3: Close Branch in Project
```bash
gh pmu branch close 2>/dev/null || echo "No branch to close"
```
---
## Phase 3: Delete Artifacts
### Step 3.1: Identify Artifact Directory
Parse branch name: `release/vX.Y.Z` → `Releases/release/vX.Y.Z/`
### Step 3.2: Delete Artifact Directory
```bash
ARTIFACT_DIR="Releases/${BRANCH_PREFIX}/${BRANCH_ID}"
if [ -d "$ARTIFACT_DIR" ]; then
  rm -rf "$ARTIFACT_DIR"
  git add -A
  git commit -m "chore: remove artifacts for destroyed branch $BRANCH"
fi
```
---
## Phase 4: Delete Branch
### Step 4.1: Switch to Main (if on target branch)
```bash
if [ "$(git branch --show-current)" = "$BRANCH" ]; then
  git checkout main
  git pull origin main
fi
```
### Step 4.2: Delete Remote Branch
```bash
git push origin --delete "$BRANCH" 2>/dev/null || echo "Remote branch not found"
```
### Step 4.3: Delete Local Branch
```bash
git branch -D "$BRANCH"
```

<!-- USER-EXTENSION-START: post-destroy -->
<!-- USER-EXTENSION-END: post-destroy -->

---
## Completion
- ✅ User confirmed destruction
- ✅ Tracker closed (not planned)
- ✅ Artifacts deleted
- ✅ Remote branch deleted
- ✅ Local branch deleted
**This action cannot be undone.**
---
## Recovery
1. **If pushed before deletion:** Check if team member has branch
2. **If only local:** Use `git reflog` within ~30 days
3. **Artifacts:** Check backups or git history
---
**End of Destroy Branch**
