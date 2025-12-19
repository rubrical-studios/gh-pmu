# Prepare-Release Scripts - Product Requirements Document

**Version:** 1.0
**Date:** 2025-12-14
**Author:** PRD generated from Proposal/PROPOSAL-Prepare-Release-Scripts.md
**Status:** Backlog Created
**Framework:** IDPF-Agile (Lightweight)

---

## Executive Summary

### Problem Statement

The `/prepare-release` Claude command involves repetitive manual steps: parsing git logs, calculating version bumps, polling CI status, and monitoring release pipelines. These tasks are error-prone, time-consuming, and could be automated with standalone scripts.

### Solution Overview

Create a suite of JavaScript scripts in `.claude/scripts/` that automate the mechanical parts of the release workflow:

1. **Commit Analysis** - Parse and categorize commits since last tag
2. **Version Recommendation** - Calculate semver bump from commit analysis
3. **CI Polling** - Wait for CI with exponential backoff
4. **Release Monitoring** - Monitor release pipeline and verify assets
5. **Config Verification** - Ensure `.gh-pmu.yml` is clean before release
6. **Changelog Generation** - Generate CHANGELOG entry from commits

### Target Users

- Developers using `/prepare-release` command
- CI/CD automation pipelines
- Release managers

### Success Criteria

- Scripts output valid JSON for machine consumption
- Scripts support `--help` flag
- Scripts exit with appropriate codes (0=success, 1=failure)
- Scripts work standalone (no Claude required)
- `/prepare-release.md` updated to reference scripts

---

## Stakeholders

| Role | Responsibility |
|------|----------------|
| Developer | Implementation, testing |
| Release Manager | Validation, feedback |

---

## Scope

### In Scope

- 6 JavaScript scripts with shared library utilities
- `package.json` for dependency management
- Integration with existing `/prepare-release` command
- JSON output format for all scripts

### Out of Scope

- Go subcommands (decided against)
- CI integration (manual execution only)
- GUI or web interface
- Cross-platform installers

---

## Feature Areas

### Feature Area 1: Core Scripts

**Epic:** Release Automation Scripts

| ID | Story | Priority | Points |
|----|-------|----------|--------|
| STORY-001 | As a developer, I can run `verify-config.js` to check if `.gh-pmu.yml` is clean | Must Have | 2 |
| STORY-002 | As a developer, I can run `analyze-commits.js` to see commits since last tag | Must Have | 3 |
| STORY-003 | As a developer, I can run `recommend-version.js` to get semver recommendation | Must Have | 2 |
| STORY-004 | As a developer, I can run `wait-for-ci.js` to poll CI until complete | Must Have | 5 |
| STORY-005 | As a developer, I can run `monitor-release.js` to watch release pipeline | Must Have | 5 |
| STORY-006 | As a developer, I can run `generate-changelog.js` to create CHANGELOG entry | Should Have | 3 |

### Feature Area 2: Shared Library

**Epic:** Script Infrastructure

| ID | Story | Priority | Points |
|----|-------|----------|--------|
| STORY-007 | As a developer, I have `lib/git.js` for git command utilities | Must Have | 2 |
| STORY-008 | As a developer, I have `lib/gh.js` for GitHub CLI wrapper | Must Have | 2 |
| STORY-009 | As a developer, I have `lib/poll.js` for async polling with backoff | Must Have | 3 |
| STORY-010 | As a developer, I have `lib/output.js` for consistent output formatting | Must Have | 1 |

### Feature Area 3: Integration

**Epic:** Command Integration

| ID | Story | Priority | Points |
|----|-------|----------|--------|
| STORY-011 | As a developer, `/prepare-release.md` references the new scripts | Must Have | 1 |
| STORY-012 | As a developer, I can run `npm install` in `.claude/scripts/` to set up | Must Have | 1 |

---

## Functional Requirements

### STORY-001: Verify Config Script

**Description:** `verify-config.js` checks if `.gh-pmu.yml` matches the committed version.

**Acceptance Criteria:**
- AC-001-1: Given clean config, When running script, Then output `{"status": "clean", "file": ".gh-pmu.yml"}`
- AC-001-2: Given dirty config, When running script, Then output includes diff and suggested action
- AC-001-3: Given `--fix` flag, When config is dirty, Then run `git checkout .gh-pmu.yml` automatically
- AC-001-4: Given `--help` flag, Then display usage information
- AC-001-5: Script exits 0 on clean, 1 on dirty (unless --fix succeeds)

---

### STORY-002: Analyze Commits Script

**Description:** `analyze-commits.js` parses commits since the last tag and categorizes by type.

**Acceptance Criteria:**
- AC-002-1: Given commits since last tag, When running script, Then output JSON with commits array
- AC-002-2: Given conventional commit format, Then extract type, scope, message, breaking flag
- AC-002-3: Given `feat!:` or `BREAKING CHANGE`, Then mark as breaking
- AC-002-4: Given output, Then include summary counts (total, feat, fix, docs, breaking)
- AC-002-5: Given `--since <tag>` flag, Then use specified tag instead of latest
- AC-002-6: Script exits 0 on success, 1 on git error

**Output Format:**
```json
{
  "lastTag": "v0.7.1",
  "commits": [
    { "hash": "abc123", "type": "feat", "scope": "api", "message": "Add endpoint", "breaking": false }
  ],
  "summary": { "total": 5, "feat": 2, "fix": 2, "docs": 1, "breaking": 0 }
}
```

---

### STORY-003: Recommend Version Script

**Description:** `recommend-version.js` calculates semver bump based on commit analysis.

**Acceptance Criteria:**
- AC-003-1: Given breaking changes, When running script, Then recommend MAJOR bump
- AC-003-2: Given features (no breaking), When running script, Then recommend MINOR bump
- AC-003-3: Given fixes only, When running script, Then recommend PATCH bump
- AC-003-4: Given `--current <version>` flag, Then use specified version as base
- AC-003-5: Given piped input from `analyze-commits.js`, Then parse JSON input
- AC-003-6: Script exits 0 on success

**Output Format:**
```json
{
  "current": "v0.7.1",
  "recommended": "v0.8.0",
  "bump": "minor",
  "reason": "2 new features (feat:), no breaking changes"
}
```

---

### STORY-004: Wait for CI Script

**Description:** `wait-for-ci.js` polls CI status with exponential backoff until complete.

**Acceptance Criteria:**
- AC-004-1: Given CI in progress, When running script, Then poll until complete
- AC-004-2: Given `--timeout <seconds>` flag, Then timeout after specified duration (default 300)
- AC-004-3: Given `--interval <seconds>` flag, Then use as initial polling interval (default 30)
- AC-004-4: Given CI failure, Then output failure details and exit 1
- AC-004-5: Given CI success, Then output job summary and exit 0
- AC-004-6: Given timeout reached, Then output timeout message and exit 1
- AC-004-7: Given polling, Then use exponential backoff (1.5x multiplier, max 60s)

**Output Format:**
```json
{
  "status": "success",
  "workflow": "CI",
  "runId": 12345,
  "duration": "2m 34s",
  "jobs": [
    { "name": "test (1.22)", "status": "success" },
    { "name": "lint", "status": "success" }
  ]
}
```

---

### STORY-005: Monitor Release Script

**Description:** `monitor-release.js` monitors tag-triggered release pipeline and verifies assets.

**Acceptance Criteria:**
- AC-005-1: Given `--tag <version>` flag, Then find and monitor the tag-triggered workflow
- AC-005-2: Given release workflow complete, Then verify expected assets uploaded
- AC-005-3: Given missing assets, Then report which assets are missing
- AC-005-4: Given `--timeout <seconds>` flag, Then timeout after specified duration (default 600)
- AC-005-5: Given all jobs pass and assets present, Then exit 0
- AC-005-6: Given any failure, Then exit 1 with details

**Expected Assets:**
- darwin-amd64, darwin-arm64
- linux-amd64, linux-arm64
- windows-amd64.exe, windows-arm64.exe
- checksums.txt

**Output Format:**
```json
{
  "status": "success",
  "tag": "v0.8.0",
  "jobs": [...],
  "assets": ["gh-pmu_0.8.0_darwin_amd64.tar.gz", ...],
  "releaseUrl": "https://github.com/rubrical-studios/gh-pmu/releases/tag/v0.8.0"
}
```

---

### STORY-006: Generate Changelog Script

**Description:** `generate-changelog.js` generates a CHANGELOG.md entry from commit analysis.

**Acceptance Criteria:**
- AC-006-1: Given commit analysis input, When running script, Then output markdown changelog entry
- AC-006-2: Given features, Then group under "### Added" section
- AC-006-3: Given fixes, Then group under "### Fixed" section
- AC-006-4: Given `--version <version>` flag, Then use in header
- AC-006-5: Given `--date <date>` flag, Then use specified date (default today)
- AC-006-6: Script outputs markdown to stdout

**Output Format:**
```markdown
## [0.8.0] - 2025-12-14

### Added
- Add new endpoint for API integration
- Add dark mode support

### Fixed
- Fix GraphQL field name casing
```

---

### STORY-007: Git Library

**Description:** Shared `lib/git.js` for git command utilities.

**Acceptance Criteria:**
- AC-007-1: Export `getLatestTag()` - returns latest semver tag
- AC-007-2: Export `getCommitsSince(tag)` - returns commit array since tag
- AC-007-3: Export `parseConventionalCommit(message)` - extracts type, scope, message, breaking
- AC-007-4: Export `isDirty(file)` - returns true if file has uncommitted changes
- AC-007-5: Export `checkout(file)` - runs `git checkout <file>`

---

### STORY-008: GitHub CLI Library

**Description:** Shared `lib/gh.js` for GitHub CLI wrapper.

**Acceptance Criteria:**
- AC-008-1: Export `getLatestRun()` - returns latest workflow run
- AC-008-2: Export `getRun(runId)` - returns specific run details with jobs
- AC-008-3: Export `getRelease(tag)` - returns release info with assets
- AC-008-4: All functions return parsed JSON objects
- AC-008-5: Handle gh CLI errors gracefully

---

### STORY-009: Polling Library

**Description:** Shared `lib/poll.js` for async polling with backoff.

**Acceptance Criteria:**
- AC-009-1: Export `poll(fn, options)` - polls async function until condition met
- AC-009-2: Support `options.interval` - initial polling interval in ms
- AC-009-3: Support `options.timeout` - max total duration in ms
- AC-009-4: Support `options.backoff` - multiplier for exponential backoff
- AC-009-5: Support `options.maxInterval` - cap on interval growth
- AC-009-6: Return promise that resolves with final result or rejects on timeout

---

### STORY-010: Output Library

**Description:** Shared `lib/output.js` for consistent output formatting.

**Acceptance Criteria:**
- AC-010-1: Export `json(data)` - outputs JSON to stdout
- AC-010-2: Export `error(message)` - outputs error to stderr
- AC-010-3: Export `success(message)` - outputs success message
- AC-010-4: Export `progress(message)` - outputs progress indicator (if TTY)
- AC-010-5: Respect `--quiet` flag when set

---

### STORY-011: Update Prepare-Release Command

**Description:** Update `/prepare-release.md` to reference the new scripts.

**Acceptance Criteria:**
- AC-011-1: Step 1 references `node .claude/scripts/analyze-commits.js`
- AC-011-2: Step 2 references `node .claude/scripts/recommend-version.js`
- AC-011-3: Step 3 references `node .claude/scripts/wait-for-ci.js`
- AC-011-4: Step 9 references `node .claude/scripts/monitor-release.js`
- AC-011-5: Pre-release step added for `verify-config.js`

---

### STORY-012: Package Setup

**Description:** Create `package.json` and document setup process.

**Acceptance Criteria:**
- AC-012-1: `package.json` created in `.claude/scripts/`
- AC-012-2: Dependencies declared: `semver`, `chalk` (if needed)
- AC-012-3: Scripts section includes `"test": "node --test"`
- AC-012-4: README or comments explain `npm install` requirement

---

## Technical Architecture

### Directory Structure

```
.claude/scripts/
├── package.json
├── lib/
│   ├── git.js
│   ├── gh.js
│   ├── poll.js
│   └── output.js
├── analyze-commits.js
├── recommend-version.js
├── wait-for-ci.js
├── monitor-release.js
├── verify-config.js
└── generate-changelog.js
```

### Dependencies

| Package | Purpose | Required |
|---------|---------|----------|
| `semver` | Version parsing and comparison | Yes |
| `chalk` | Colored terminal output | Optional |

### Design Decisions

| Decision | Rationale |
|----------|-----------|
| JavaScript over Go | Matches existing `workflow-trigger.js` pattern, easier iteration |
| `package.json` included | Allows npm packages (semver, chalk) for richer functionality |
| No Go subcommands | Keep release tooling separate from core CLI |
| Manual execution only | Avoid CI complexity, scripts run during `/prepare-release` |

---

## Constraints

| ID | Constraint | Impact |
|----|------------|--------|
| CON-001 | Node.js required | Users must have Node.js installed |
| CON-002 | `gh` CLI required | CI/release scripts depend on GitHub CLI |
| CON-003 | Git repository required | All scripts assume git context |

---

## Testing Approach

- **Unit Tests:** Node.js built-in test runner (`node --test`)
- **Mocking:** Mock `child_process.execSync` for git/gh calls
- **Coverage:** Aim for 80% coverage on library functions
- **Integration:** Manual testing with `/prepare-release` command

---

## Implementation Order

| Priority | Story | Rationale |
|----------|-------|-----------|
| 1 | STORY-007 | Git library - foundation for other scripts |
| 2 | STORY-010 | Output library - needed by all scripts |
| 3 | STORY-001 | verify-config.js - immediate value |
| 4 | STORY-002 | analyze-commits.js - foundation for versioning |
| 5 | STORY-003 | recommend-version.js - uses commit analysis |
| 6 | STORY-008 | GitHub CLI library - needed for CI scripts |
| 7 | STORY-009 | Polling library - needed for CI scripts |
| 8 | STORY-004 | wait-for-ci.js - automates waiting |
| 9 | STORY-005 | monitor-release.js - ensures complete releases |
| 10 | STORY-006 | generate-changelog.js - optional enhancement |
| 11 | STORY-012 | package.json setup |
| 12 | STORY-011 | Update prepare-release.md |

---

## Approval

| Role | Name | Date | Approved |
|------|------|------|----------|
| Stakeholder | | | [ ] |

---

**PRD Status: READY - Pending approval for IDPF-Agile development**
