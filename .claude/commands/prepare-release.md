---
version: "v0.20.3"
description: Prepare release with PR, merge to main, and tag
argument-hint: [version] [--skip-coverage] [--dry-run] [--help]
---

<!-- EXTENSIBLE: v0.17.0 -->
# /prepare-release

Validate, create PR to main, merge, and tag for deployment.

## Available Extension Points

| Point | Location | Purpose |
|-------|----------|---------|
| `post-analysis` | After Phase 1 | Commit analysis, version recommendation |
| `pre-validation` | Before Phase 2 | Setup, fixtures, containers |
| `post-validation` | After Phase 2 | Coverage gates, build verification |
| `post-prepare` | After Phase 3 | Documentation updates |
| `pre-tag` | Before Phase 4 tagging | Final gate, sign-off |
| `post-tag` | After Phase 4 | Release monitoring, asset verification |

---

## Arguments

| Argument | Description |
|----------|-------------|
| `[version]` | Version to release (e.g., v1.2.0) |
| `--skip-coverage` | Skip coverage gate |
| `--dry-run` | Preview without changes |
| `--help` | Show extension points |

---

## Pre-Checks

### Verify Config

```bash
node .claude/scripts/open-release/verify-config.js
```

**If the script returns `success: false`, STOP and report the error.**

### Verify on Release Branch

```bash
git branch --show-current
```

### Check for Open Sprints

```bash
gh pmu microsprint current
```

Close open sprints before proceeding.

### Check for Incomplete Issues

```bash
gh pmu release current --json issues | jq '.[] | select(.status != "done")'
```

---

## Phase 1: Analysis (Framework-Provided)

### Step 1.1: Analyze Changes

```bash
git log $(git describe --tags --abbrev=0)..HEAD --oneline
```

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

**ASK USER:** Confirm version before proceeding.

---

## Phase 2: Validation (Framework-Provided)

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

### Step 2.1: Run Tests

```bash
go test ./...
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

## Phase 3: Prepare (Framework-Provided)

### Step 3.1: Update CHANGELOG.md

Follow Keep a Changelog format.

### Step 3.2: Commit Preparation

```bash
git add CHANGELOG.md README.md docs/
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

## Phase 4: Git Operations (Framework-Provided)

### Step 4.1: Create PR to Main

```bash
gh pr create --base main --head $(git branch --show-current) \
  --title "Release $VERSION"
```

### Step 4.2: Merge PR

**ASK USER:** Approve and merge.

```bash
gh pr merge --merge
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

### Step 4.3: Tag and Push

**ASK USER:** Confirm ready to tag.

```bash
git tag -a $VERSION -m "Release $VERSION"
git push origin $VERSION
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

## Summary Checklist

**Before tagging:**
- [ ] Config file clean
- [ ] Commits analyzed
- [ ] Lint gate passed
- [ ] Coverage gate passed (or `--skip-coverage`)
- [ ] Version confirmed
- [ ] CI passing
- [ ] CHANGELOG updated
- [ ] PR merged

**After tagging:**
- [ ] All CI jobs completed
- [ ] Release assets uploaded
- [ ] Release notes updated

---

## Next Step

After deployment verified, run `/close-release`.

---

**End of Prepare Release**
