# Prepare Release Workflow

This command guides you through preparing a new release for gh-pmu.

---

## Step 1: Analyze Changes Since Last Release

Run these commands to understand what changed:

```bash
# Get the latest release tag
git tag --sort=-v:refname | head -1

# View commits since last release
git log $(git tag --sort=-v:refname | head -1)..HEAD --oneline

# Count by type
git log $(git tag --sort=-v:refname | head -1)..HEAD --oneline | grep -c "^[a-f0-9]* feat"
git log $(git tag --sort=-v:refname | head -1)..HEAD --oneline | grep -c "^[a-f0-9]* fix"
```

**Report to user:**
- Number of commits since last release
- Breakdown by type (feat/fix/docs/chore)
- Any breaking changes (look for `!:` or `BREAKING CHANGE`)

---

## Step 2: Recommend Version Number

Based on [Semantic Versioning](https://semver.org/):

| Change Type | Version Bump | Example |
|-------------|--------------|---------|
| Breaking changes (`feat!:`, `BREAKING CHANGE`) | MAJOR | 0.2.x → 1.0.0 |
| New features (`feat:`) | MINOR | 0.2.13 → 0.3.0 |
| Bug fixes only (`fix:`) | PATCH | 0.2.13 → 0.2.14 |

**Present recommendation to user and wait for confirmation.**

---

## Step 3: Wait for CI to Pass

**CRITICAL: Do not proceed until CI passes.**

```bash
# Check latest workflow run status
gh run list --limit 1

# If in progress, wait and check again
gh run list --limit 1 --json status,conclusion,name
```

**CI Wait Logic:**
1. Check if any workflow is running
2. If `status: "in_progress"`, inform user and wait
3. Poll every 30 seconds until complete
4. If `conclusion: "failure"`, STOP and report the failure
5. Only proceed if `conclusion: "success"`

**Report CI status to user before continuing.**

---

## Step 4: Update CHANGELOG.md

Follow [Keep a Changelog](https://keepachangelog.com/) format:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added
- New features (from feat: commits)

### Changed
- Changes to existing functionality

### Fixed
- Bug fixes (from fix: commits)

### Removed
- Removed features

### Security
- Security fixes
```

**Steps:**
1. Read current CHANGELOG.md
2. Move `[Unreleased]` section content to new version section
3. Add today's date
4. Group commits by type
5. Write clear, user-facing descriptions (not commit messages verbatim)

**Show the proposed CHANGELOG entry to user for approval.**

---

## Step 5: Check README.md

Review if README needs updates:

- [ ] Version badge updated (if applicable)
- [ ] New commands documented
- [ ] New flags documented
- [ ] Installation instructions current
- [ ] Examples updated

**Only update README if there are user-facing changes to document.**

---

## Step 6: Review Documentation Freshness

Check if documentation files need updates based on changes in this release:

### docs/gh-comparison.md
Compare `gh pmu` commands with base `gh` CLI. Update if:
- [ ] New commands added to gh pmu
- [ ] New flags added that overlap with `gh` functionality
- [ ] Workflow examples need updating

```bash
# Quick check: list current gh pmu commands
go run . --help

# Compare with documented commands in gh-comparison.md
```

### Other docs to review
- [ ] `docs/testing.md` - if test patterns changed
- [ ] `coverage/README.md` - coverage report auto-updates, verify accuracy

**Only update documentation if changes in this release affect it.**

---

## Step 7: Commit Release Preparation

```bash
git add CHANGELOG.md README.md docs/
git commit -m "chore: prepare release vX.Y.Z"
git push
```

**Wait for CI to pass on this commit before proceeding.**

---

## Step 8: Create Tag

**Ask user for confirmation before creating tag.**

```bash
# Create and push tag
git tag vX.Y.Z
git push origin vX.Y.Z
```

**DO NOT STOP HERE. Proceed immediately to Step 9.**

---

## Step 9: Monitor Release Pipeline to Completion

**CRITICAL: The release is NOT complete until all CI jobs finish successfully.**

Pushing a tag triggers the release workflow. You MUST monitor it to completion:

```bash
# Get the run ID for the tag push
gh run list --limit 1 --json databaseId,status,headBranch

# Monitor job progress
gh run view <run-id> --json status,conclusion,jobs
```

**Required Monitoring:**
1. Poll every 30 seconds until `status: "completed"`
2. Verify ALL jobs pass:
   - test (all matrix combinations)
   - lint
   - build (all matrix combinations)
   - **release** (GoReleaser - creates binaries)
   - **coverage** (updates coverage report)
3. If ANY job fails, report immediately and stop
4. Verify release assets were uploaded:
   ```bash
   gh release view vX.Y.Z --json tagName,assets
   ```

**Only after verifying:**
- All jobs completed successfully
- Release assets are uploaded (binaries for all platforms)
- Coverage report was committed

**THEN report to user:**
```
✅ Release vX.Y.Z complete!

CI Pipeline:
  ✅ test (1.22, 1.23)
  ✅ lint
  ✅ build (ubuntu, macos × go 1.22, 1.23)
  ✅ release (GoReleaser)
  ✅ coverage report updated

Assets uploaded:
  • darwin-amd64, darwin-arm64
  • linux-amd64, linux-arm64
  • windows-amd64.exe, windows-arm64.exe
  • checksums.txt

https://github.com/rubrical-studios/gh-pmu/releases/tag/vX.Y.Z
```

---

## Summary Checklist

Before tagging, verify:

- [ ] All commits analyzed and categorized
- [ ] Version number confirmed with user
- [ ] CI passing on main branch
- [ ] CHANGELOG.md updated with new version
- [ ] README.md updated (if needed)
- [ ] docs/gh-comparison.md reviewed (if new commands/flags)
- [ ] Release preparation committed and pushed
- [ ] CI passing on release preparation commit
- [ ] User confirmed ready to tag

After tagging, verify:

- [ ] All CI jobs completed successfully (test, lint, build, release, coverage)
- [ ] Release assets uploaded (all platform binaries + checksums)
- [ ] `gh release view vX.Y.Z` confirms release exists with assets

---

## Example Session

```
Assistant: Analyzing commits since v0.2.13...

Found 3 commits:
- feat: Add gh pmu close command with reason aliases (#249)
- feat: Add gh pmu board command for terminal board view (#250)
- fix(ci): Disable integration tests workflow

Breakdown: 2 feat, 1 fix, 0 breaking

Recommended version: v0.3.0 (MINOR bump for new features)

Checking CI status...
✓ CI passed (workflow: CI, conclusion: success)

Shall I prepare the CHANGELOG entry for v0.3.0?
```

---

## Important Rules

1. **NEVER skip CI verification** - Always wait for green CI
2. **NEVER auto-create tags** - Always get user confirmation
3. **NEVER guess version numbers** - Base on actual commit analysis
4. **ALWAYS show changes before committing** - User must approve
5. **NEVER declare release complete after pushing tag** - Tag push only triggers the pipeline; you MUST monitor until GoReleaser finishes and assets are uploaded
6. **ALWAYS verify release assets exist** - Run `gh release view` to confirm binaries were uploaded
