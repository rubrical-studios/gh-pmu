---
description: Prepare release with PR, merge to main, and tag
argument-hint: [version] (e.g., v1.2.0) [--skip-coverage]
---

# Prepare Release

Validate, create PR to main, merge, and tag for deployment.

---

## Pre-Checks

### 1. Verify Config File

**IMPORTANT:** Ensure `.gh-pmu.yml` is clean (not modified by tests).

```bash
node .claude/scripts/verify-config.js
```

If dirty, fix with:
```bash
node .claude/scripts/verify-config.js --fix
```

### 2. Verify on Release Branch

```bash
git branch --show-current
```

Must be on a release branch (e.g., `release/v1.2.0`), not `main`.

### 3. Check for Open Sprints

```bash
gh pmu microsprint current
```

If sprints are open, close them first:
```bash
gh pmu microsprint close
```

### 4. Check for Incomplete Issues

```bash
gh pmu release current --json issues | jq '.[] | select(.status != "done")'
```

If incomplete issues exist, prompt user:
- Transfer to next release
- Return to backlog
- Block release (cannot close with open issues)

---

## Phase 1: Analysis & Version

### Step 1.1: Analyze Changes Since Last Release

```bash
node .claude/scripts/analyze-commits.js
```

This outputs JSON with:
- `lastTag`: The most recent version tag
- `commits`: Array of parsed commits with type, scope, message, breaking flag
- `summary`: Counts by type (feat, fix, docs, etc.) and breaking changes

**Report to user:**
- Number of commits since last release
- Breakdown by type (feat/fix/docs/chore)
- Any breaking changes (look for `breaking: true`)

### Step 1.2: Coverage Gate (Optional)

**Skip this step if using `--skip-coverage` flag.**

```bash
node .claude/scripts/analyze-coverage.js
```

Options:
- `--since <tag>` - Compare against specific tag (default: latest)
- `--threshold <n>` - Minimum patch coverage % (default: 80)
- `--skip-tests` - Use existing coverage.out file

**If coverage is below threshold:**
1. Review the `addressableGaps` array in output
2. Create issue for coverage improvements
3. **Inform user** with options:
   - Address coverage gaps before releasing
   - Use `--skip-coverage` to proceed anyway
   - Adjust threshold in `.gh-pmu.yml`

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

### Step 1.3: Recommend Version Number

```bash
node .claude/scripts/recommend-version.js
```

Or pipe from analyze-commits:
```bash
node .claude/scripts/analyze-commits.js | node .claude/scripts/recommend-version.js
```

Applies [Semantic Versioning](https://semver.org/):

| Change Type | Version Bump | Example |
|-------------|--------------|---------|
| Breaking changes (`feat!:`, `BREAKING CHANGE`) | MAJOR | 0.2.x → 1.0.0 |
| New features (`feat:`) | MINOR | 0.2.13 → 0.3.0 |
| Bug fixes only (`fix:`) | PATCH | 0.2.13 → 0.2.14 |

**Present recommendation to user and wait for confirmation.**

### Step 1.4: Update CHANGELOG.md

Follow [Keep a Changelog](https://keepachangelog.com/) format:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- New features (from feat: commits)

### Changed
- Changes to existing functionality

### Fixed
- Bug fixes (from fix: commits)
```

**Steps:**
1. Read current CHANGELOG.md
2. Move `[Unreleased]` content to new version section
3. Add today's date
4. Group commits by type
5. Write clear, user-facing descriptions

**Show proposed CHANGELOG entry to user for approval.**

### Step 1.5: Review Documentation (if needed)

Check if docs need updates based on changes:

- [ ] `docs/commands.md` - if commands/flags changed
- [ ] `docs/configuration.md` - if config options changed
- [ ] `README.md` - if user-facing features changed

**Only update if changes affect documentation.**

---

## Phase 2: Git Preparation

### Step 2.1: Commit Release Preparation

```bash
git add CHANGELOG.md README.md docs/
git commit -m "chore: prepare release $VERSION"
git push
```

### Step 2.2: Wait for CI to Pass

**CRITICAL: Do not proceed until CI passes.**

```bash
node .claude/scripts/wait-for-ci.js
```

Options:
- `--timeout <seconds>` - Max wait time (default: 300)
- `--interval <seconds>` - Polling interval (default: 30)

The script will:
1. Find the latest workflow run
2. Poll with exponential backoff until complete
3. Output job-by-job status
4. Exit 0 on success, 1 on failure

**Report CI status to user before continuing.**

### Step 2.3: Create PR to Main

```bash
gh pr create --base main --head $(git branch --show-current) \
  --title "Release $VERSION" \
  --body "## Release $VERSION

See CHANGELOG.md for details."
```

### Step 2.4: Merge PR

**ASK USER:** Approve and merge PR (or wait for CI/review)

After PR merged, switch to main and pull:
```bash
git checkout main
git pull origin main
```

---

## Phase 3: Tag & Release

### Step 3.1: Create Tag on Main

**Ask user for confirmation before creating tag.**

```bash
git tag -a $VERSION -m "Release $VERSION"
git push origin $VERSION
```

**DO NOT STOP HERE. Proceed to monitoring.**

### Step 3.2: Monitor Release Pipeline

**CRITICAL: The release is NOT complete until all CI jobs finish successfully.**

```bash
node .claude/scripts/monitor-release.js --tag $VERSION
```

Options:
- `--timeout <seconds>` - Max wait time (default: 600)
- `--interval <seconds>` - Polling interval (default: 30)

The script will:
1. Find and monitor the tag-triggered workflow
2. Poll until all jobs complete
3. Verify release assets are uploaded:
   - darwin-amd64, darwin-arm64
   - linux-amd64, linux-arm64
   - windows-amd64.exe, windows-arm64.exe
   - checksums.txt
4. Exit 0 only when release is complete with all assets

**Report output to user when complete.**

---

## Phase 4: Post-Release

### Step 4.1: Update GitHub Release Notes

```bash
node .claude/scripts/update-release-notes.js --version $VERSION
```

Options:
- `--version <version>` - Version to update (required)
- `--dry-run` - Preview without making changes

The script will:
1. Parse CHANGELOG.md for the specified version
2. Transform content into GitHub release format
3. Update the GitHub release with formatted notes

**Preview first with `--dry-run`, then run without flag.**

### Step 4.2: Clean Up Old Release Assets (Optional)

```bash
node .claude/scripts/cleanup-release-assets.js
```

Options:
- `--keep <n>` - Releases to keep assets for (default: 3)
- `--dry-run` - Preview without deleting

**Preview first with `--dry-run`.**

### Step 4.3: Verify Release

```bash
gh release view $VERSION
```

Confirm:
- [ ] Release exists with correct tag
- [ ] All platform binaries uploaded
- [ ] Release notes populated
- [ ] Checksums.txt present

---

## Next Step

After deployment is verified, run `/close-release` to:
- Create GitHub Release page (if not auto-created)
- Close gh-pmu tracker issue
- Delete release branch

---

## Post-Release Reminder

**Releasing code does NOT close related issues.**
Issues included in this release still require explicit user approval ("Done") to close.
Do NOT auto-close issues just because they shipped.

---

## Summary Checklist

**Before tagging:**
- [ ] Config file clean
- [ ] All commits analyzed and categorized
- [ ] Coverage gate passed (or `--skip-coverage`)
- [ ] Version confirmed with user
- [ ] CI passing
- [ ] CHANGELOG.md updated
- [ ] Documentation reviewed
- [ ] Release preparation committed
- [ ] PR created and merged

**After tagging:**
- [ ] All CI jobs completed successfully
- [ ] Release assets uploaded (all platforms + checksums)
- [ ] Release notes updated
- [ ] Old assets cleaned up (optional)

---

## Important Rules

1. **NEVER skip CI verification** - Always wait for green CI
2. **NEVER auto-create tags** - Always get user confirmation
3. **NEVER guess version numbers** - Base on actual commit analysis
4. **ALWAYS show changes before committing** - User must approve
5. **NEVER declare release complete after pushing tag** - Monitor until assets uploaded
6. **ALWAYS verify release assets exist** - Run `gh release view` to confirm

---

## Release Lifecycle

```
/open-release v1.2.0
    └── Creates branch + tracker
         │
         ▼
    [Work on release branch]
         │
         ▼
/prepare-release v1.2.0     ◄── YOU ARE HERE
    └── PR → merge → tag → deploy
         │
         ▼
/close-release
    └── GitHub Release → cleanup
```

---

**End of Prepare Release**
