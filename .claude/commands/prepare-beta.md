---
version: "v0.31.0"
description: Tag beta from feature branch (no merge to main)
argument-hint: [--skip-coverage] [--dry-run] [--help]
---
<!-- EXTENSIBLE -->
# /prepare-beta
Tag a beta release from feature branch without merging to main.
## Available Extension Points
| Point | Location | Purpose |
|-------|----------|---------|
| `post-analysis` | After Phase 1 | Commit analysis |
| `pre-validation` | Before Phase 2 | Setup test environment |
| `post-validation` | After Phase 2 | Beta validation |
| `post-prepare` | After Phase 3 | Additional updates |
| `pre-tag` | Before Phase 4 | Final gate |
| `post-tag` | After Phase 4 | Beta monitoring |
| `checklist-before-tag` | Summary Checklist | Pre-tag items |
| `checklist-after-tag` | Summary Checklist | Post-tag items |
## Arguments
| Argument | Description |
|----------|-------------|
| `--skip-coverage` | Skip coverage gate |
| `--dry-run` | Preview without changes |
| `--help` | Show extension points |
## Execution Instructions
**REQUIRED:** Before executing:
1. **Create Todo List:** Use `TodoWrite` for phases
2. **Track Progress:** Mark `in_progress` → `completed`
3. **Resume Point:** Todos show where to continue
## Pre-Checks
### Verify NOT on Main
```bash
BRANCH=$(git branch --show-current)
if [ "$BRANCH" = "main" ]; then echo "Error: Cannot create beta from main."; exit 1; fi
```
## Phase 1: Analysis
```bash
git log $(git describe --tags --abbrev=0)..HEAD --oneline
```

<!-- USER-EXTENSION-START: post-analysis -->
### Analyze Commits

```bash
node .claude/scripts/framework/analyze-commits.js
```

### Recommend Version

```bash
node .claude/scripts/framework/recommend-version.js
```

Recommend beta version (e.g., `v1.0.0-beta.1`).

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

**ASK USER:** Confirm beta version.
## Phase 2: Validation

<!-- USER-EXTENSION-START: pre-validation -->
### Lint Gate

```bash
node .claude/scripts/prepare-release/lint.js
```

The script outputs JSON: `{"success": true/false, "message": "..."}`

**If `success` is false, STOP and report the error.**

Runs `golangci-lint run --timeout=5m` to catch lint errors before tagging.
<!-- USER-EXTENSION-END: pre-validation -->

```bash
go test ./...
```

<!-- USER-EXTENSION-START: post-validation -->
### Coverage Gate (Optional for Beta)

**If `--skip-coverage` was passed, skip this section.**

```bash
node .claude/scripts/prepare-release/coverage.js
```

**If `success` is false, STOP and report the error.**

### E2E Gate (Optional for Beta)

**If `--skip-e2e` was passed, skip this section.**

```bash
node .claude/scripts/e2e/run-e2e-gate.js
```

The script outputs JSON: `{"success": true/false, "testsRun": N, "testsPassed": N, "duration": N}`

**If `success` is false, STOP and report the error.**

E2E tests validate complete workflows against the test project.
<!-- USER-EXTENSION-END: post-validation -->

**ASK USER:** Confirm validation passed.
## Phase 3: Prepare
Update CHANGELOG.md with beta section.

<!-- USER-EXTENSION-START: post-prepare -->
### Wait for CI

```bash
node .claude/scripts/framework/wait-for-ci.js
```

The script polls CI status every 60 seconds (5-minute timeout).

**If CI fails, STOP and report the error.**

<!-- USER-EXTENSION-END: post-prepare -->

## Phase 4: Tag (No Merge)
### Step 4.1: Commit Changes
```bash
git add -A
git commit -m "chore: prepare beta $VERSION"
git push origin $(git branch --show-current)
```

<!-- USER-EXTENSION-START: pre-tag -->

### Important Rules

1. **NEVER skip CI verification** - Always wait for green CI
2. **NEVER auto-create tags** - Always get user confirmation
3. **NEVER guess version numbers** - Base on actual commit analysis
4. **ALWAYS show changes before committing** - User must approve
5. **NEVER declare release complete after pushing tag** - Monitor until assets uploaded

### Beta Tag Authorization

The pre-push hook blocks version tags. For beta tags, authorize before pushing:

```bash
echo 'beta-authorized' > .release-authorized
git push origin $VERSION
rm .release-authorized
```
<!-- USER-EXTENSION-END: pre-tag -->

### Step 4.2: Create Beta Tag
**ASK USER:** Confirm ready to tag.
```bash
git tag -a $VERSION -m "Beta $VERSION"
git push origin $VERSION
```
**Note:** Beta tags feature branch. No merge to main.
### Step 4.3: Wait for CI Workflow
```bash
node .claude/scripts/framework/wait-for-ci.js
```
**If CI fails, STOP.**
### Step 4.4: Update Release Notes
```bash
node .claude/scripts/framework/update-release-notes.js
```

<!-- USER-EXTENSION-START: post-tag -->
### Monitor Beta Build

```bash
node .claude/scripts/close-release/monitor-release.js
```

Monitor beta build and asset upload.

### Update Release Notes

```bash
node .claude/scripts/framework/update-release-notes.js
```

Updates GitHub Release with formatted notes from CHANGELOG.

### Post-Release Reminder

**Releasing a beta does NOT close related issues.**

Issues included in this beta still require explicit user approval ("Done") to close.
Do NOT auto-close issues just because a beta shipped.
<!-- USER-EXTENSION-END: post-tag -->

## Next Step
When ready for full release:
1. Merge feature branch to main
2. Run `/prepare-release`
## Summary Checklist
**Before tagging:**
- [ ] Not on main
- [ ] Commits analyzed
- [ ] Beta version confirmed
- [ ] Tests passing
- [ ] CHANGELOG updated

<!-- USER-EXTENSION-START: checklist-before-tag -->
- [ ] Coverage gate passed (or `--skip-coverage`)
<!-- USER-EXTENSION-END: checklist-before-tag -->

**After tagging:**
- [ ] Beta tag pushed
- [ ] CI workflow completed
- [ ] Release notes updated

<!-- USER-EXTENSION-START: checklist-after-tag -->
- [ ] Beta build monitored
<!-- USER-EXTENSION-END: checklist-after-tag -->

**End of Prepare Beta**
