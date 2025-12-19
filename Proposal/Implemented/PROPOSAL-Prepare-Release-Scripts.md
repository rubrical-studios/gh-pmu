# Proposal: Prepare-Release Script Automation

**Issue:** #326
**Date:** 2025-12-14
**Status:** ✅ Implemented
**PRD:** [PRD/Implemented/PRD-Prepare-Release-Scripts.md](../../PRD/Implemented/PRD-Prepare-Release-Scripts.md)
**Implemented:** 2025-12-14

---

## Executive Summary

Evaluate extracting parts of `/prepare-release` Claude command into standalone scripts. Analysis shows **5 high-value candidates** for automation, with a recommendation to implement in **JavaScript** to maintain consistency with existing hooks.

---

## Current State Analysis

### `/prepare-release` Workflow Steps

| Step | Name | Current Implementation | Automation Potential |
|------|------|----------------------|---------------------|
| 1 | Analyze Changes | Git log parsing, manual categorization | **HIGH** |
| 2 | Version Recommendation | Semver logic from commit types | **HIGH** |
| 3 | CI Wait | Polling `gh run list` | **HIGH** |
| 4 | CHANGELOG Update | Manual markdown editing | **MEDIUM** |
| 5 | README Check | Manual review | LOW |
| 6 | Documentation Review | Manual checklist | LOW |
| 7 | Commit Preparation | Git commands | LOW |
| 8 | Tag Creation | Git commands | LOW |
| 9 | Monitor Release | Polling CI, verify assets | **HIGH** |
| - | Config Verification | Verify `.gh-pmu.yml` clean | **HIGH** |

---

## Proposed Scripts

### 1. `analyze-commits.js`

**Purpose:** Parse commits since last tag, categorize by type, detect breaking changes.

**Input:**
- None (reads from git)

**Output (JSON):**
```json
{
  "lastTag": "v0.7.1",
  "commits": [
    { "hash": "abc123", "type": "feat", "scope": "api", "message": "Add new endpoint", "breaking": false }
  ],
  "summary": {
    "total": 5,
    "feat": 2,
    "fix": 2,
    "docs": 1,
    "breaking": 0
  }
}
```

**Commands replaced:**
```bash
git tag --sort=-v:refname | head -1
git log $(git tag...)..HEAD --oneline
git log ... | grep -c "feat"
```

**Complexity:** Low
**Dependencies:** None (pure git)

---

### 2. `recommend-version.js`

**Purpose:** Calculate semver bump based on commit analysis.

**Input:** Output from `analyze-commits.js` or `--last-tag v0.7.1`

**Output (JSON):**
```json
{
  "current": "v0.7.1",
  "recommended": "v0.8.0",
  "bump": "minor",
  "reason": "2 new features (feat:), no breaking changes"
}
```

**Logic:**
- Breaking changes (`feat!:`, `BREAKING CHANGE`) → MAJOR
- Features (`feat:`) → MINOR
- Fixes only (`fix:`) → PATCH

**Complexity:** Low
**Dependencies:** `analyze-commits.js` output (optional)

---

### 3. `wait-for-ci.js`

**Purpose:** Poll CI status with exponential backoff until complete.

**Input:**
- `--timeout 300` (seconds, default 300)
- `--interval 30` (seconds, default 30)

**Output (JSON):**
```json
{
  "status": "success",
  "workflow": "CI",
  "runId": 12345,
  "duration": "2m 34s",
  "jobs": [
    { "name": "test (1.22)", "status": "success" },
    { "name": "test (1.23)", "status": "success" },
    { "name": "lint", "status": "success" },
    { "name": "build", "status": "success" }
  ]
}
```

**Commands replaced:**
```bash
gh run list --limit 1
gh run list --limit 1 --json status,conclusion,name
# + manual polling loop
```

**Complexity:** Medium (async polling, backoff)
**Dependencies:** `gh` CLI

---

### 4. `monitor-release.js`

**Purpose:** Monitor tag-triggered release pipeline to completion, verify assets.

**Input:**
- `--tag v0.8.0`
- `--timeout 600` (seconds)

**Output (JSON):**
```json
{
  "status": "success",
  "tag": "v0.8.0",
  "jobs": [
    { "name": "test", "status": "success" },
    { "name": "lint", "status": "success" },
    { "name": "build", "status": "success" },
    { "name": "release", "status": "success" },
    { "name": "coverage", "status": "success" }
  ],
  "assets": [
    "gh-pmu_0.8.0_darwin_amd64.tar.gz",
    "gh-pmu_0.8.0_darwin_arm64.tar.gz",
    "gh-pmu_0.8.0_linux_amd64.tar.gz",
    "gh-pmu_0.8.0_linux_arm64.tar.gz",
    "gh-pmu_0.8.0_windows_amd64.zip",
    "gh-pmu_0.8.0_windows_arm64.zip",
    "checksums.txt"
  ],
  "releaseUrl": "https://github.com/rubrical-studios/gh-pmu/releases/tag/v0.8.0"
}
```

**Commands replaced:**
```bash
gh run list --limit 1 --json databaseId,status,headBranch
gh run view <run-id> --json status,conclusion,jobs
gh release view vX.Y.Z --json tagName,assets
# + manual polling loop
```

**Complexity:** Medium-High (polling, asset verification)
**Dependencies:** `gh` CLI

---

### 5. `verify-config.js`

**Purpose:** Verify `.gh-pmu.yml` matches committed version (not modified by tests).

**Input:** None

**Output (JSON):**
```json
{
  "status": "clean",
  "file": ".gh-pmu.yml",
  "message": "Config matches committed version"
}
```

Or if dirty:
```json
{
  "status": "dirty",
  "file": ".gh-pmu.yml",
  "diff": "+ Release field added\n+ Microsprint field added",
  "action": "Run: git checkout .gh-pmu.yml"
}
```

**Commands replaced:**
```bash
git diff .gh-pmu.yml
git checkout .gh-pmu.yml  # if --fix flag
```

**Complexity:** Low
**Dependencies:** None (pure git)

---

### 6. `generate-changelog.js` (Lower Priority)

**Purpose:** Generate CHANGELOG entry from commit analysis.

**Input:** Output from `analyze-commits.js`

**Output (Markdown):**
```markdown
## [0.8.0] - 2025-12-14

### Added
- Add new endpoint for API integration
- Add dark mode support

### Fixed
- Fix GraphQL field name casing
- Fix config file parsing
```

**Complexity:** Medium (grouping, formatting)
**Dependencies:** `analyze-commits.js` output

---

## Shared Utilities

Common functionality to extract into `lib/`:

| Utility | Used By | Purpose |
|---------|---------|---------|
| `git.js` | 1, 2, 5 | Git command execution, tag parsing |
| `gh.js` | 3, 4 | GitHub CLI wrapper, JSON parsing |
| `poll.js` | 3, 4 | Async polling with backoff |
| `semver.js` | 2 | Version parsing and bumping |
| `output.js` | All | Consistent JSON/text output formatting |

---

## Language Evaluation: JS vs Go

| Factor | JavaScript | Go |
|--------|-----------|-----|
| **Existing pattern** | ✅ `workflow-trigger.js` exists | Project is Go |
| **Setup required** | None (Node.js assumed) | Already have Go |
| **Claude hooks compatibility** | ✅ Native (stdin/stdout JSON) | Requires build step |
| **Async/polling** | ✅ Native async/await | Goroutines (overkill) |
| **Portability** | Needs Node.js | Single binary |
| **Development speed** | ✅ Faster iteration | Slower |
| **Integration with gh-pmu** | Separate scripts | Could be subcommands |

### Recommendation: **JavaScript**

Reasons:
1. **Existing pattern** - `workflow-trigger.js` already demonstrates the hook approach
2. **Claude integration** - Scripts can be called from Claude commands easily
3. **No build step** - Edit and run immediately
4. **Async-friendly** - Polling is natural with async/await
5. **Separation of concerns** - Release tooling != core CLI functionality

---

## Proposed Directory Structure

```
.claude/
├── hooks/
│   └── workflow-trigger.js      # Existing
├── commands/
│   └── prepare-release.md       # Updated to use scripts
└── scripts/
    ├── package.json             # Dependencies (semver, chalk, etc.)
    ├── lib/
    │   ├── git.js               # Git utilities
    │   ├── gh.js                # GitHub CLI wrapper
    │   ├── poll.js              # Polling utilities
    │   └── output.js            # Output formatting
    ├── analyze-commits.js       # Step 1
    ├── recommend-version.js     # Step 2
    ├── wait-for-ci.js           # Step 3
    ├── monitor-release.js       # Step 9
    ├── verify-config.js         # Pre-release check
    └── generate-changelog.js    # Step 4 (optional)
```

---

## Updated `/prepare-release` Flow

```markdown
## Step 1: Analyze Changes
Run: `node .claude/scripts/analyze-commits.js`
[Script outputs JSON with commit breakdown]

## Step 2: Recommend Version
Run: `node .claude/scripts/recommend-version.js`
[Script outputs recommended version with reasoning]

## Step 3: Wait for CI
Run: `node .claude/scripts/wait-for-ci.js`
[Script polls and reports when CI passes]

... (Steps 4-8 remain manual with Claude guidance) ...

## Step 9: Monitor Release
Run: `node .claude/scripts/monitor-release.js --tag vX.Y.Z`
[Script monitors pipeline and verifies assets]

## Pre-release: Verify Config
Run: `node .claude/scripts/verify-config.js`
[Script ensures .gh-pmu.yml is clean]
```

---

## Implementation Priority

| Priority | Script | Effort | Value |
|----------|--------|--------|-------|
| 1 | `verify-config.js` | Low | High (prevents dirty releases) |
| 2 | `analyze-commits.js` | Low | High (used by multiple steps) |
| 3 | `recommend-version.js` | Low | High (eliminates guesswork) |
| 4 | `wait-for-ci.js` | Medium | High (automates waiting) |
| 5 | `monitor-release.js` | Medium | High (ensures complete releases) |
| 6 | `generate-changelog.js` | Medium | Medium (still needs review) |

---

## Acceptance Criteria

- [x] Scripts output valid JSON for machine consumption
- [x] Scripts support `--help` flag
- [x] Scripts exit with appropriate codes (0=success, 1=failure)
- [x] Scripts work standalone (no Claude required)
- [x] `/prepare-release.md` updated to reference scripts
- [x] `.gh-pmu.yml` verification added to workflow

### Implementation Notes (2025-12-14)

All 6 scripts implemented in `.claude/scripts/`:
- `verify-config.js` (2.9 KB) - Config file validation
- `analyze-commits.js` (3.9 KB) - Commit parsing/categorization
- `recommend-version.js` (5.2 KB) - Semver bump recommendation
- `wait-for-ci.js` (4.5 KB) - CI polling with backoff
- `monitor-release.js` (9.1 KB) - Release pipeline monitoring
- `generate-changelog.js` (6.0 KB) - Changelog generation

Shared libraries in `.claude/scripts/lib/`:
- `git.js`, `gh.js`, `poll.js`, `output.js`

---

## Decisions

| Question | Decision |
|----------|----------|
| **Package.json?** | **Yes** - Add `package.json` for npm packages (semver, chalk, etc.) |
| **Go subcommands?** | **No** - JS scripts only, keep separate from core CLI |
| **CI integration?** | **No** - Manual execution during `/prepare-release` only |

---

## Next Steps

1. Review and approve this proposal
2. Implement `verify-config.js` first (immediate value)
3. Implement `analyze-commits.js` + `recommend-version.js` together
4. Implement `wait-for-ci.js` + `monitor-release.js` together
5. Update `/prepare-release.md` to use scripts
