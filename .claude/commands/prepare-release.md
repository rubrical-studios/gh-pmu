---
version: "v0.24.1"
description: Prepare framework release with version updates and validation (project)
argument-hint: [options...] (phase:N, skip:*, audit:*, dry-run)
---
<!-- EXTENSIBLE -->
# /prepare-release
Execute the full release preparation workflow.
## Extension Points
| Point | Location | Purpose |
|-------|----------|---------|
| `post-analysis` | After Phase 1 | Custom commit analysis |
| `pre-validation` | Before Phase 2 | Setup test environment |
| `post-validation` | After Phase 2 | Custom validation |
| `post-prepare` | After Phase 3 | Additional updates |
| `pre-tag` | Before Phase 4 | Final gate, sign-off |
| `post-tag` | After Phase 4 | Deployment, notifications |
| `pre-close` | Before Phase 5 | Pre-close validation |
| `post-close` | After Phase 5 | Post-release actions |
---
## Arguments
| Usage | Behavior |
|-------|----------|
| `/prepare-release` | Full process |
| `/prepare-release phase:2` | Start from Phase 2 |
| `/prepare-release skip:X` | Skip sub-phase |
| `/prepare-release dry-run` | Preview only |
---
## Pre-Checks
### Verify Current Branch
```bash
git branch --show-current
```
Record the current branch name as `$BRANCH` for use in subsequent steps.
### Check for Open Work
```bash
gh pmu microsprint current 2>/dev/null
gh pmu release current --json issues | jq '.[] | select(.status != \"done\")'
```
---
## Phase 1: Analysis & Version
### Step 1.1: Identify Last Release
```bash
git describe --tags --abbrev=0
```
### Step 1.2: List Commits Since Last Release
```bash
git log vX.Y.Z..HEAD --oneline
git log vX.Y.Z..HEAD --oneline | wc -l
```
### Step 1.3: Categorize Changes
| Category | Indicators | Impact |
|----------|-----------|--------|
| New Framework | "Add IDPF-*" | MINOR/MAJOR |
| New Skill | "Implement * skill" | MINOR |
| Bug Fix | "Fix *" | PATCH |
### Step 1.4: Determine Version
**ASK USER:** Confirm version.
<!-- USER-EXTENSION-START: post-analysis -->
### Analyze Commits

```bash
node .claude/scripts/framework/analyze-commits.js
```

The script outputs JSON with commit analysis:
- `lastTag`: Previous version
- `commits`: Array of parsed commits
- `summary`: Counts by type (features, fixes, etc.)

### Recommend Version

```bash
node .claude/scripts/framework/recommend-version.js
```

Uses the commit analysis to recommend a version bump.

### Documentation Review

Check if docs need updates based on changes:

- [ ] `docs/commands.md` - if commands/flags changed
- [ ] `docs/configuration.md` - if config options changed
- [ ] `README.md` - if user-facing features changed

**Only update if changes affect documentation.**
<!-- USER-EXTENSION-END: post-analysis -->
---
## Phase 2: Validation
<!-- USER-EXTENSION-START: pre-validation -->
### Handle Incomplete Issues

If incomplete issues exist, prompt user:
- Transfer to next release
- Return to backlog
- Block release (cannot proceed with open issues)

### Lint Gate

```bash
node .claude/scripts/prepare-release/lint.js
```

The script outputs JSON: `{"success": true/false, "message": "..."}`

**If `success` is false, STOP and report the error.**

Runs `golangci-lint run --timeout=5m` to catch lint errors before tagging.
<!-- USER-EXTENSION-END: pre-validation -->
### Step 2.1: Verify Working Directory
```bash
git status --porcelain
```
### Step 2.2: Run Basic Tests
```bash
npm test 2>/dev/null || echo "No test script"
```
<!-- USER-EXTENSION-START: post-validation -->
### Coverage Gate

**If `--skip-coverage` was passed, skip this section.**

```bash
node .claude/scripts/prepare-release/coverage.js
```

The script outputs JSON: `{"success": true/false, "message": "...", "data": {"coverage": 87.5}}`

**If `success` is false, STOP and report the error.**

Coverage metrics include total percentage and threshold comparison.

**Configuration** (`.gh-pmu.yml`):
```yaml
release:
  coverage:
    enabled: true
    threshold: 80
    skip_patterns:
      - "*_test.go"
      - "mock_*.go"
```
<!-- USER-EXTENSION-END: post-validation -->
**ASK USER:** Confirm validation passed.
---
## Phase 3: Prepare
### Step 3.1: Update Version Files
| File | Action |
|------|--------|
| `framework-manifest.json` | Update version |
| `CHANGELOG.md` | Add new section |
| `README.md` | Update version line |
### Step 3.2: Generate Release Artifacts
```bash
mkdir -p "Releases/$TRACK/$VERSION"
```
Create `release-notes.md` and `changelog.md`.
<!-- USER-EXTENSION-START: post-prepare -->
### Wait for CI

```bash
node .claude/scripts/framework/wait-for-ci.js
```

The script polls CI status every 60 seconds (5-minute timeout).

**If CI fails, STOP and report the error.**
<!-- USER-EXTENSION-END: post-prepare -->
---
## Phase 4: Git Operations
### Step 4.1: Commit Release
```bash
git add -A
git commit -m "Release vX.Y.Z"
git push origin release/vX.Y.Z
```
### Step 4.2: Update Issue Criteria
Update acceptance criteria on release issues before PR.
### Step 4.3: Create PR to Main
```bash
gh pr create --base main --head release/vX.Y.Z --title "Release vX.Y.Z"
```
### Step 4.4: Merge PR
**ASK USER:** Approve and merge.
```bash
gh pr merge --merge
```
### Step 4.5: Close Branch Tracker
```bash
gh pmu branch close
```
### Step 4.6: Switch to Main
```bash
git checkout main
git pull origin main
```
<!-- USER-EXTENSION-START: pre-tag -->
### Important Rules

1. **NEVER skip CI verification** - Always wait for green CI
2. **NEVER auto-create tags** - Always get user confirmation
3. **NEVER guess version numbers** - Base on actual commit analysis
4. **ALWAYS show changes before committing** - User must approve
5. **NEVER declare release complete after pushing tag** - Monitor until assets uploaded
6. **ALWAYS verify release assets exist** - Run `gh release view` to confirm
<!-- USER-EXTENSION-END: pre-tag -->
### Step 4.7: Tag Main
**ASK USER:** Confirm ready to tag.
```bash
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push origin vX.Y.Z
```
### Step 4.8: Verify Deployment
```bash
gh run list --limit 1
gh run watch
```
<!-- USER-EXTENSION-START: post-tag -->
### Monitor Release Workflow

```bash
node .claude/scripts/close-release/monitor-release.js
```

The script monitors the tag-triggered workflow and verifies all platform binaries are uploaded:
- darwin-amd64, darwin-arm64
- linux-amd64, linux-arm64
- windows-amd64.exe, windows-arm64.exe
- checksums.txt

**If assets are missing after timeout, report warning but continue.**

### Update Release Notes

```bash
node .claude/scripts/framework/update-release-notes.js
```

Updates GitHub Release with formatted notes from CHANGELOG.

### Clean Up Old Release Assets (Optional)

```bash
node .claude/scripts/shared/cleanup-release-assets.js --keep 3 --dry-run
```

Options:
- `--keep <n>` - Releases to keep assets for (default: 3)
- `--dry-run` - Preview without deleting

**Preview first with `--dry-run`.**

### Post-Release Reminder

**Releasing code does NOT close related issues.**

Issues included in this release still require explicit user approval ("Done") to close.
Do NOT auto-close issues just because they shipped.
<!-- USER-EXTENSION-END: post-tag -->
---
<!-- USER-EXTENSION-START: pre-close -->
<!-- USER-EXTENSION-END: pre-close -->
## Phase 5: Close & Cleanup
**ASK USER:** Confirm deployment verified.
### Step 5.1: Close Tracker Issue
```bash
gh issue close [TRACKER_NUMBER] --comment "Release vX.Y.Z deployed successfully"
```
### Step 5.2: Delete Release Branch
```bash
git push origin --delete release/vX.Y.Z
git branch -d release/vX.Y.Z
```
### Step 5.3: Create GitHub Release
```bash
gh release create vX.Y.Z --title "Release vX.Y.Z" --notes-file "Releases/release/vX.Y.Z/release-notes.md"
```
<!-- USER-EXTENSION-START: post-close -->
<!-- USER-EXTENSION-END: post-close -->
---
## Completion
- ✅ Code merged to main
- ✅ Tag created and pushed
- ✅ Deployment verified
- ✅ Tracker closed
- ✅ Branch deleted
- ✅ GitHub Release created
---
**End of Prepare Release**
