---
version: "v0.33.3"
description: Prepare release with PR, merge to main, and tag
argument-hint: [version] [--skip-coverage] [--dry-run] [--help]
---

<!-- EXTENSIBLE -->
# /prepare-release

Validate, create PR to main, merge, and tag for deployment.

## Available Extension Points

| Point | Location | Purpose |
|-------|----------|---------|
| `post-analysis` | After Phase 1 | Commit analysis, version recommendation |
| `pre-validation` | Before Phase 2 | Setup, fixtures, containers |
| `post-validation` | After Phase 2 | Coverage gates, build verification |
| `pre-commit` | Before Phase 3 commit | Generate release artifacts |
| `post-prepare` | After Phase 3 | Documentation updates |
| `post-pr-create` | After PR creation | CI wait, PR validation |
| `pre-tag` | Before Phase 4 tagging | Final gate, sign-off |
| `post-tag` | After Phase 4 | Release monitoring, asset verification |
| `checklist-before-tag` | Summary Checklist | Pre-tag verification items |
| `checklist-after-tag` | Summary Checklist | Post-tag verification items |
| `pre-close` | Before Phase 5 | Pre-close validation, notifications |
| `post-close` | After Phase 5 | Post-release actions, announcements |
| `checklist-close` | Summary Checklist | Close verification items |

---

## Arguments

| Argument | Description |
|----------|-------------|
| `[version]` | Version to release (e.g., v1.2.0) |
| `--skip-coverage` | Skip coverage gate |
| `--dry-run` | Preview without changes |
| `--help` | Show extension points |

---

## Execution Instructions

**REQUIRED:** Before executing this command:

1. **Generate Todo List:** Parse the phases and extension points in this spec, then use `TodoWrite` to create todos
2. **Include Extensions:** For each non-empty `USER-EXTENSION` block, add a todo item
3. **Track Progress:** Mark todos `in_progress` → `completed` as you work
4. **Post-Compaction:** If resuming after context compaction, re-read this spec and regenerate todos

**Todo Generation Rules:**
- One todo per numbered phase/step
- One todo per active extension point (non-empty `USER-EXTENSION` blocks)
- Skip commented-out extensions
- Use the phase/step name as the todo content

---

## Pre-Checks

### Verify Current Branch

```bash
git branch --show-current
```

Record the current branch name as `$BRANCH` for use in subsequent steps.

### Auto-Create Release Branch (if on main)

**If `$BRANCH` is `main`:**

1. **Analyze commits to recommend version:**
   ```bash
   git log $(git describe --tags --abbrev=0)..HEAD --oneline
   ```

2. **Recommend version** based on commit analysis (same as Phase 1.1)

3. **ASK USER:** Confirm version (e.g., `v0.26.0`)

4. **If `--dry-run`:** Report "Would create branch: release/v0.26.0" and stop.

5. **Create release branch:**
   ```bash
   gh pmu branch start --name "release/$VERSION"
   git checkout "release/$VERSION"
   git push -u origin "release/$VERSION"
   ```

6. **Update `$BRANCH`** to `release/$VERSION`

7. **Report:** "Created release branch: release/$VERSION. Continuing with release preparation..."

**If `$BRANCH` is NOT `main`:** Continue with normal flow (existing working branch).

### Check for Incomplete Issues

```bash
 gh pmu list --branch current --status backlog,in_progress,in_review
```

---

## Phase 1: Analysis

### Step 1.1: Analyze Changes

```bash
git log $(git describe --tags --abbrev=0)..HEAD --oneline
```

### Analyze Commits

```bash
node .claude/scripts/shared/analyze-commits.js
```

The script outputs JSON with commit analysis:
- `lastTag`: Previous version
- `commits`: Array of parsed commits
- `summary`: Counts by type (features, fixes, etc.)

### Recommend Version

```bash
node .claude/scripts/shared/recommend-version.js
```

Uses the commit analysis to recommend a version bump.

<!-- USER-EXTENSION-START: post-analysis -->

### Documentation Review

Check if docs need updates based on changes:

- [ ] `docs/commands.md` - if commands/flags changed
- [ ] `docs/configuration.md` - if config options changed
- [ ] `README.md` - if user-facing features changed

**Only update if changes affect documentation.**

### E2E Impact Analysis

```bash
node .claude/scripts/e2e/analyze-e2e-impact.js
```

The script analyzes which E2E tests may be impacted by changes:
- `impactedTests`: Test files that cover changed commands
- `newCommandsWithoutTests`: Commands modified without E2E coverage
- `recommendation`: Suggested test review actions

**If `newCommandsWithoutTests` is non-empty, warn user about missing coverage.**

<!-- USER-EXTENSION-END: post-analysis -->

**ASK USER:** Confirm version before proceeding.

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

<!-- USER-EXTENSION-START: post-validation -->

### Step 2.1: Run Tests

```bash
go test ./...
```

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

### E2E Gate

**If `--skip-e2e` was passed, skip this section.**

```bash
node .claude/scripts/e2e/run-e2e-gate.js
```

The script outputs JSON: `{"success": true/false, "testsRun": N, "testsPassed": N, "duration": N}`

**If `success` is false, STOP and report the error.**

E2E tests validate complete workflows against the test project.
<!-- USER-EXTENSION-END: post-validation -->

**ASK USER:** Confirm validation passed.


---

## Phase 3: Prepare

### Step 3.1: Update Version Files


| File | Action |
|------|--------|
| `CHANGELOG.md` | Add new section following Keep a Changelog format |
| `README.md` | Update version badge or header |
| `README-DIST.md` | Verify skill/specialist counts match actuals, license populated |
| `framework-config.json` | (Self-hosted only) Update `frameworkVersion` and `installedDate` |


<!-- USER-EXTENSION-START: pre-commit -->
<!-- Pre-commit: generate release artifacts -->
<!-- USER-EXTENSION-END: pre-commit -->

### Step 3.2: Commit Preparation

```bash
git add CHANGELOG.md README.md README-DIST.md docs/
git commit -m "chore: prepare release $VERSION"
git push
```

<!-- USER-EXTENSION-START: post-prepare -->
### Wait for CI

```bash
node .claude/scripts/framework/wait-for-ci.js
```

The script polls CI status every 60 seconds (5-minute timeout).

**If CI fails, STOP and report the error.**
<!-- USER-EXTENSION-END: post-prepare -->

**CRITICAL:** Do not proceed until CI passes.

---

## Phase 4: Git Operations

### Step 4.1: Create PR to Main

```bash
gh pr create --base main --head $(git branch --show-current) \
  --title "Release $VERSION"
```

<!-- USER-EXTENSION-START: post-pr-create -->

### Wait for CI
```bash
node .claude/scripts/shared/wait-for-ci.js
```
**If CI fails, STOP and report.**

<!-- USER-EXTENSION-END: post-pr-create -->

### Step 4.2: Merge PR

**ASK USER:** Approve and merge.

```bash
gh pr merge --merge
```

### Step 4.3: Close Branch Tracker

Close the branch tracker while still on the working branch:

```bash
gh pmu branch close
```

### Step 4.4: Switch to Main

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

### Step 4.5: Tag and Push

**ASK USER:** Confirm ready to tag.

```bash
git tag -a $VERSION -m "Release $VERSION"
git push origin $VERSION
```

### Step 4.6: Wait for CI Workflow

```bash
node .claude/scripts/shared/wait-for-ci.js
```

**If CI fails, STOP and report.**

### Step 4.7: Update Release Notes

```bash
node .claude/scripts/shared/update-release-notes.js
```

Updates GitHub Release with formatted notes from CHANGELOG.

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

### Clean Up Old Release Assets

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

## Summary Checklist

**Core (Before tagging):**
- [ ] Commits analyzed
- [ ] Version confirmed
- [ ] CHANGELOG updated
- [ ] PR merged

<!-- USER-EXTENSION-START: checklist-before-tag -->
- [ ] Coverage gate passed (or `--skip-coverage`)
- [ ] CI passing
<!-- USER-EXTENSION-END: checklist-before-tag -->

**Core (After tagging):**

- [ ] Tag pushed
- [ ] CI workflow completed
- [ ] Release notes updated

<!-- USER-EXTENSION-START: checklist-after-tag -->
- [ ] All CI jobs completed
- [ ] Release assets uploaded
- [ ] Release notes updated
<!-- USER-EXTENSION-END: checklist-after-tag -->

---

<!-- USER-EXTENSION-START: pre-close -->
<!-- USER-EXTENSION-END: pre-close -->

## Phase 5: Close & Cleanup

**ASK USER:** Confirm deployment verified and ready to close release.

### Step 5.1: Add Deployment Comment

Add deployment success comment to the tracker issue (already closed by Phase 4.3):

```bash
gh issue comment [TRACKER_NUMBER] --body "Release $VERSION deployed successfully"
```

### Step 5.2: Delete Working Branch

After successful merge and tag, clean up the working branch:

```bash
# Delete remote branch (already on main from Step 4.4)
git push origin --delete $BRANCH
git branch -d $BRANCH
```

### Step 5.3: Create GitHub Release

Create an official GitHub Release with release notes:

```bash
gh release create $VERSION \
  --title "Release $VERSION" \
  --notes-file CHANGELOG.md
```

<!-- USER-EXTENSION-START: post-close -->
<!-- USER-EXTENSION-END: post-close -->

---

## Summary Checklist (Close)

<!-- USER-EXTENSION-START: checklist-close -->
- [ ] Tracker issue closed
- [ ] Release closed in project
- [ ] Working branch deleted
- [ ] GitHub Release created
<!-- USER-EXTENSION-END: checklist-close -->

---

## Completion

Release $VERSION is complete:
- ✅ Code merged to main
- ✅ Tag created and pushed
- ✅ Deployment verified
- ✅ Tracker issue closed
- ✅ Working branch deleted
- ✅ GitHub Release created

---

**End of Prepare Release**
