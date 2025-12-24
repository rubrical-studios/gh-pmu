# Prepare Release Workflow

This command guides you through preparing a new release for gh-pmu.

---

## Pre-Step: Verify Config File

**IMPORTANT:** Before starting, ensure `.gh-pmu.yml` is clean (not modified by tests).

```bash
node .claude/scripts/verify-config.js
```

If dirty, fix with:
```bash
node .claude/scripts/verify-config.js --fix
```

---

## Step 1: Analyze Changes Since Last Release

Run the analysis script:

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

---

## Step 2: Coverage Gate (Optional)

**Skip this step if using `--skip-coverage` flag.**

Run the coverage analysis script:

```bash
node .claude/scripts/analyze-coverage.js
```

Options:
- `--since <tag>` - Compare against specific tag (default: latest)
- `--threshold <n>` - Minimum patch coverage % (default: 80, configurable in `.gh-pmu.yml`)
- `--skip-tests` - Use existing coverage.out file

The script will:
1. Run `go test -coverprofile=coverage.out ./...`
2. Parse coverage results and compare against changed lines since last tag
3. Calculate patch coverage percentage
4. Categorize uncovered lines as addressable vs. non-addressable
5. Exit with code 2 if coverage is below threshold

**If coverage is below threshold:**
1. Review the `addressableGaps` array in the output
2. Create an issue for coverage improvements using `gh pmu create`:
   ```bash
   gh pmu create --repo rubrical-studios/gh-pmu --title "Test coverage: Address gaps before release" --label enhancement --status backlog
   ```
3. Include the uncovered files and line numbers in the issue body
4. **Inform user**: "Coverage at X% is below the Y% threshold. Created issue #Z for coverage work. Options:"
   - Address the coverage gaps before releasing
   - Use `--skip-coverage` to proceed anyway (not recommended)
   - Adjust threshold in `.gh-pmu.yml` if appropriate

**Configuration** (`.gh-pmu.yml`):
```yaml
release:
  coverage:
    enabled: true          # Enable coverage gate (default: true)
    threshold: 80          # Minimum patch coverage % (default: 80)
    skip_patterns:         # Patterns to exclude
      - "*_test.go"
      - "mock_*.go"
```

**If coverage passes or is skipped, proceed to Step 3.**

---

## Step 3: Recommend Version Number

Run the version recommendation script:

```bash
node .claude/scripts/recommend-version.js
```

Or pipe from analyze-commits:
```bash
node .claude/scripts/analyze-commits.js | node .claude/scripts/recommend-version.js
```

This applies [Semantic Versioning](https://semver.org/):

| Change Type | Version Bump | Example |
|-------------|--------------|---------|
| Breaking changes (`feat!:`, `BREAKING CHANGE`) | MAJOR | 0.2.x → 1.0.0 |
| New features (`feat:`) | MINOR | 0.2.13 → 0.3.0 |
| Bug fixes only (`fix:`) | PATCH | 0.2.13 → 0.2.14 |

**Present recommendation to user and wait for confirmation.**

---

## Step 4: Wait for CI to Pass

**CRITICAL: Do not proceed until CI passes.**

Run the CI waiting script:

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

---

## Step 5: Update CHANGELOG.md

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

## Step 6: Check README.md

Review if README needs updates:

- [ ] Installation instructions current
- [ ] Quick Start examples work correctly
- [ ] Feature descriptions accurate
- [ ] Documentation links valid

**Only update README if there are user-facing changes to document.**

---

## Step 7: Review Documentation Freshness

Check if documentation files need updates based on changes in this release:

```bash
# Quick check: list current gh pmu commands
go run . --help
```

### docs/commands.md
Complete command reference. Update if:
- [ ] New commands added
- [ ] New flags added to existing commands
- [ ] Command output format changed
- [ ] Examples still accurate

### docs/configuration.md
Configuration file reference. Update if:
- [ ] New config options added
- [ ] Field alias format changed
- [ ] Triage rule syntax changed
- [ ] Default values changed

### docs/sub-issues.md
Sub-issue hierarchy guide. Update if:
- [ ] Sub-issue commands changed
- [ ] Progress tracking behavior changed
- [ ] Cross-repo functionality changed

### docs/batch-operations.md
Intake, triage, split workflows. Update if:
- [ ] Batch command behavior changed
- [ ] Triage rule options changed
- [ ] Split functionality changed

### docs/gh-comparison.md
Compare `gh pmu` with base `gh` CLI. Update if:
- [ ] New commands added to gh pmu
- [ ] New flags that overlap with `gh` functionality
- [ ] Workflow examples need updating

### docs/development.md
Development guide. Update if:
- [ ] Build process changed
- [ ] Test commands changed
- [ ] Project structure changed

### Other files to review
- [ ] `CONTRIBUTING.md` - contribution guidelines current
- [ ] `coverage/README.md` - auto-updates, verify accuracy

**Only update documentation if changes in this release affect it.**

---

## Step 8: Commit Release Preparation

```bash
git add CHANGELOG.md README.md docs/
git commit -m "chore: prepare release vX.Y.Z"
git push
```

**Wait for CI to pass on this commit before proceeding.**

---

## Step 9: Create Tag

**Ask user for confirmation before creating tag.**

```bash
# Create and push tag
git tag vX.Y.Z
git push origin vX.Y.Z
```

**DO NOT STOP HERE. Proceed immediately to Step 10.**

---

## Step 10: Monitor Release Pipeline to Completion

**CRITICAL: The release is NOT complete until all CI jobs finish successfully.**

Run the release monitoring script:

```bash
node .claude/scripts/monitor-release.js --tag vX.Y.Z
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

**Report the output to user when complete.**

**DO NOT STOP HERE. Proceed immediately to Step 11.**

---

## Step 11: Update GitHub Release Notes

After the release is created by CI, update the release notes with content from CHANGELOG.md:

```bash
node .claude/scripts/update-release-notes.js --version vX.Y.Z
```

Options:
- `--version <version>` - Version to update (required)
- `--dry-run` - Show what would be updated without making changes

The script will:
1. Parse CHANGELOG.md for the specified version
2. Transform content into GitHub release format with sections:
   - "What's New" (features, changes, performance)
   - "Bug Fixes"
   - "Security" (if applicable)
   - Full changelog comparison link
3. Update the GitHub release with the formatted notes

**Preview first with `--dry-run`, then run without the flag to update.**

---

## Step 12: Clean Up Old Release Assets

After the release is complete, clean up assets from older releases to save storage:

```bash
node .claude/scripts/cleanup-release-assets.js
```

Options:
- `--keep <n>` - Number of recent releases to keep assets for (default: 3)
- `--dry-run` - Show what would be deleted without making changes

The script will:
1. List all GitHub releases with assets
2. Keep assets for the 3 most recent tagged releases
3. Delete assets from older releases (preserves release entries)

**Preview first with `--dry-run`, then run without the flag to clean up.**

**Report cleanup results to user when complete.**

---

## Summary Checklist

Before tagging, verify:

- [ ] Config file clean (`node .claude/scripts/verify-config.js`)
- [ ] All commits analyzed and categorized
- [ ] Coverage gate passed (or `--skip-coverage` confirmed)
- [ ] Version number confirmed with user
- [ ] CI passing on main branch
- [ ] CHANGELOG.md updated with new version
- [ ] README.md reviewed (if user-facing changes)
- [ ] Documentation reviewed (if commands/config changed):
  - [ ] docs/commands.md
  - [ ] docs/configuration.md
  - [ ] docs/sub-issues.md
  - [ ] docs/batch-operations.md
  - [ ] docs/gh-comparison.md
  - [ ] docs/development.md
  - [ ] CONTRIBUTING.md
- [ ] Release preparation committed and pushed
- [ ] CI passing on release preparation commit
- [ ] User confirmed ready to tag

After tagging, verify:

- [ ] All CI jobs completed successfully (test, lint, build, release, coverage)
- [ ] Release assets uploaded (all platform binaries + checksums)
- [ ] `gh release view vX.Y.Z` confirms release exists with assets
- [ ] Release notes updated from CHANGELOG.md (Step 11)
- [ ] Old release assets cleaned up (Step 12)

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
