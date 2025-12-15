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

## Step 2: Recommend Version Number

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

## Step 3: Wait for CI to Pass

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

- [ ] Installation instructions current
- [ ] Quick Start examples work correctly
- [ ] Feature descriptions accurate
- [ ] Documentation links valid

**Only update README if there are user-facing changes to document.**

---

## Step 6: Review Documentation Freshness

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

---

## Summary Checklist

Before tagging, verify:

- [ ] Config file clean (`node .claude/scripts/verify-config.js`)
- [ ] All commits analyzed and categorized
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
