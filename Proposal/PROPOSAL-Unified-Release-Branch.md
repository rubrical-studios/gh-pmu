# Proposal: Unified Release with Branch Flag

**Issue:** #346
**Date:** 2025-12-19
**Status:** Draft

---

## Executive Summary

Two changes to simplify gh pmu:

1. **Unified Release Command:** Consolidate `release` and `patch` command groups into a single `release` command with a `--branch` flag. The branch name determines release type, validation rules, and artifact paths. This eliminates redundant code while preserving semantic versioning discipline.

2. **Extended `move` Command:** Add `--release`, `--microsprint`, and `--backlog` flags to move issues between releases and sprints. Supports the framework-level `Transfer-Issue` command in IDPF-Agile.

---

## Current State

### Duplicate Command Groups

| Command | `release` | `patch` |
|---------|-----------|---------|
| `start` | Yes | Yes |
| `add` | Yes | Yes |
| `remove` | Yes | Yes |
| `current` | Yes | Yes |
| `close` | Yes | Yes |
| `list` | Yes | Yes |

**Lines of code:** ~850 in `release.go`, ~850 in `patch.go` = **~1700 lines** with significant duplication.

### Key Differences (Current)

| Aspect | `release` | `patch` |
|--------|-----------|---------|
| Label validation | None | Blocks `breaking-change`, warns if no `bug/fix/hotfix` |
| Tracker label | `release` | `patch` |
| Artifact path | `Releases/v1.2.0/` | `Releases/patch/v1.1.5/` |
| Framework | IDPF-Structured | IDPF-LTS |

---

## Proposed Design

### New Command Syntax

```bash
# Standard release
gh pmu release start --branch "release/v2.0.0"

# Pre-release versions
gh pmu release start --branch "release/v2.0.0-beta.1"

# Patch release (enables label validation)
gh pmu release start --branch "patch/v1.9.1"

# Hotfix - versioned or named (enables label validation)
gh pmu release start --branch "hotfix/v1.9.2"
gh pmu release start --branch "hotfix-auth-bypass"
```

Branch names are not validated - users choose their own naming conventions.

### Branch Name Parsing

```
--branch "track/vX.Y.Z"           ->  track=track, version=X.Y.Z
--branch "track/vX.Y.Z-suffix"    ->  track=track, version=X.Y.Z-suffix
```

| Input | Track | Version | Creates Tag |
|-------|-------|---------|-------------|
| `release/v2.0.0` | `release` | `2.0.0` | Yes |
| `release/v2.0.0-beta.1` | `release` | `2.0.0-beta.1` | Yes |
| `patch/v1.9.1` | `patch` | `1.9.1` | Yes |
| `hotfix/v1.9.2` | `hotfix` | `1.9.2` | Yes |

### Track-Based Behavior

| Track | Label Validation | Artifact Path |
|-------|------------------|---------------|
| `release` (default) | None | `Releases/{branch}/` |
| `patch/*` | Block `breaking-change`, warn if no `bug/fix/hotfix` | `Releases/{branch}/` |
| `hotfix*` | Block `breaking-change`, warn if no `bug/fix/hotfix` | `Releases/{branch}/` |

### Validation Rules by Track

```go
type trackConfig struct {
    blockLabels  []string  // Error if present
    warnLabels   []string  // Warn if ALL missing
}

var tracks = map[string]trackConfig{
    "release": {
        blockLabels: nil,
        warnLabels:  nil,
    },
    "patch": {
        blockLabels: []string{"breaking-change"},
        warnLabels:  []string{"bug", "fix", "hotfix"},
    },
    "hotfix": {
        blockLabels: []string{"breaking-change"},
        warnLabels:  []string{"bug", "fix", "hotfix"},
    },
}
```

All tracks create git branches on start and tags on close (with `--tag` flag).

---

## Commands to Remove

### Entire `patch` Command Group

| Command | Replacement |
|---------|-------------|
| `gh pmu patch start --version X.Y.Z` | `gh pmu release start --branch "patch/vX.Y.Z"` |
| `gh pmu patch add <issue>` | `gh pmu release add <issue>` (track from active release) |
| `gh pmu patch remove <issue>` | `gh pmu release remove <issue>` |
| `gh pmu patch current` | `gh pmu release current` |
| `gh pmu patch close [--tag]` | `gh pmu release close [--tag]` |
| `gh pmu patch list` | `gh pmu release list --track patch` |

### Files to Delete

- `cmd/patch.go` (~850 lines)
- `cmd/patch_test.go` (~900 lines)

### Estimated Reduction

- **~1750 lines removed**
- **6 commands removed** from CLI surface
- **1 command group removed** from help output

---

## Migration Path

### Phase 1: Add `--branch` Flag (Non-Breaking)

1. Add `--branch` flag to `release start`
2. Implement branch parsing and track detection
3. Add track-based validation to `release add`
4. Update artifact paths based on track
5. Deprecation warning when using `patch` commands

### Phase 2: Remove `patch` Commands (Breaking)

1. Remove `cmd/patch.go` and tests
2. Update documentation
3. Update `.claude/rules/02-github-workflow.md`
4. Major version bump (v0.8.0 or v1.0.0)

---

## Implementation Details

### Options Struct Change

```go
// Before
type releaseStartOptions struct {
    version string
}

// After
type releaseStartOptions struct {
    version string  // Deprecated, use --branch
    branch  string  // New: "release/v2.0.0", "patch/v1.9.1", "hotfix-name"
}
```

### Branch Parsing Function

```go
func parseBranch(branch string) (track, version string, err error) {
    // Handle "prefix/version" format
    if strings.Contains(branch, "/") {
        parts := strings.SplitN(branch, "/", 2)
        track = parts[0]
        version = strings.TrimPrefix(parts[1], "v")
        return track, version, nil
    }

    // Handle bare name (hotfix)
    return "hotfix", branch, nil
}
```

### Git Branch Creation

```go
func (c *Client) CreateBranch(name string) error {
    return exec.Command("git", "checkout", "-b", name).Run()
}
```

---

## Examples

### Standard Release Workflow

```bash
gh pmu release start --branch "release/v2.0.0"
# Creates: branch release/v2.0.0, tracker "Release: v2.0.0"

gh pmu release add 123
# No validation, sets Release field to "v2.0.0"

gh pmu release close --tag
# Creates: tag v2.0.0, artifacts in Releases/v2.0.0/
```

### Patch Workflow

```bash
gh pmu release start --branch "patch/v1.9.1"
# Creates: branch patch/v1.9.1, tracker "Release: patch/v1.9.1"

gh pmu release add 456
# Validates: blocks breaking-change, warns if no bug/fix/hotfix label

gh pmu release close --tag
# Creates: tag v1.9.1, artifacts in Releases/patch/v1.9.1/
```

### Emergency Hotfix (Versioned)

```bash
gh pmu release start --branch "hotfix/v1.9.2"
# Creates: branch hotfix/v1.9.2, tracker "Release: hotfix/v1.9.2"

gh pmu release add 789
# Validates: blocks breaking-change, warns if no bug/fix/hotfix label

gh pmu release close --tag
# Creates: tag v1.9.2, artifacts in Releases/hotfix/v1.9.2/
```

### Emergency Hotfix (Named)

```bash
gh pmu release start --branch "hotfix-auth-bypass"
# Creates: branch hotfix-auth-bypass, tracker "Release: hotfix-auth-bypass"

gh pmu release add 789
# Validates: blocks breaking-change, warns if no bug/fix/hotfix label

gh pmu release close
# No tag (no version), artifacts in Releases/hotfix-auth-bypass/
```

---

## Backward Compatibility

### Breaking Changes (Single Release)

This is a **breaking change** release:

- `gh pmu patch *` commands removed
- `--version` flag removed from `release start`
- `--branch` flag required (no default)

### Migration Guide

| Before | After |
|--------|-------|
| `gh pmu release start --version 2.0.0` | `gh pmu release start --branch release/v2.0.0` |
| `gh pmu patch start --version 1.9.1` | `gh pmu release start --branch patch/v1.9.1` |
| `gh pmu patch add 42` | `gh pmu release add 42` |
| `gh pmu patch close --tag` | `gh pmu release close --tag` |

---

## Configuration

### .gh-pmu.yml Extensions

```yaml
tracks:
  patch:
    prefix: "patch/"
    block_labels: ["breaking-change"]
    warn_labels: ["bug", "fix", "hotfix"]
    create_tag: true
  hotfix:
    prefix: "hotfix-"
    block_labels: ["breaking-change"]
    warn_labels: ["bug", "fix", "hotfix"]
    create_tag: false
  beta:
    prefix: "beta/"
    create_tag: true
```

---

## Decisions

| Question | Decision |
|----------|----------|
| Interactive mode | **No** - Require `--branch` flag |
| `--version` flag | **Remove** - Only `--branch` flag, cleaner API |
| Git branch creation | **Always** - `git checkout -b {branch}` on start |
| Branch validation | **None** - User controls naming, conventions in docs |

---

## Branch Naming (Documentation Only)

No validation enforced. Recommended conventions documented in `docs/workflows.md`:

### Recommended Formats

| Type | Branch Pattern | Tag Pattern | Example |
|------|----------------|-------------|---------|
| Release | `release/vX.Y.Z` | `vX.Y.Z` | `release/v2.0.0` |
| Pre-release | `release/vX.Y.Z-suffix` | `vX.Y.Z-suffix` | `release/v2.0.0-beta.1` |
| Patch | `patch/vX.Y.Z` | `vX.Y.Z` | `patch/v1.9.1` |
| Hotfix | `hotfix/vX.Y.Z` | `vX.Y.Z` | `hotfix/v1.9.2` |
| Hotfix (named) | `hotfix-name` | - | `hotfix-auth-bypass` |

### Track Detection from Branch

Track is derived from branch prefix for label validation:

```go
func parseTrack(branch string) string {
    if strings.HasPrefix(branch, "patch/") {
        return "patch"
    }
    if strings.HasPrefix(branch, "hotfix") {
        return "hotfix"
    }
    return "release"  // default
}
```

Only `patch` and `hotfix` tracks apply label validation (block `breaking-change`, warn if no `bug/fix/hotfix`).

---

## Extended `move` Command

### Purpose

Add flags to move issues between releases and sprints. Supports the framework-level `Transfer-Issue` command.

### New Flags

| Flag | Description |
|------|-------------|
| `--release <value>` | Set the Release field to specified value |
| `--microsprint <value>` | Set the Microsprint field to specified value (alias: `--sprint`) |
| `--backlog` | Clear Release and Microsprint fields (return to backlog) |

### Examples

```bash
# Existing move functionality (unchanged)
gh pmu move 45 --status in_progress --priority p1

# New: Move issue to different release
gh pmu move 45 --release "release/v2.0.0"

# New: Move issue to different sprint within current release
gh pmu move 45 --microsprint "auth-work"

# New: Move issue to different release and sprint
gh pmu move 45 --release "patch/v1.9.1" --microsprint "bugfixes"

# New: Return issue to backlog (clear release and microsprint)
gh pmu move 45 --backlog

# Combine with existing flags
gh pmu move 45 --status in_progress --release "release/v2.0.0"
```

### Validation

- `--backlog` cannot be combined with `--release` or `--microsprint`
- No validation on release/microsprint values — user controls naming
- All flags remain optional (existing behavior preserved)

### Implementation

Extend existing `moveOptions` struct:

```go
type moveOptions struct {
    // Existing fields
    status      string
    priority    string
    // ...

    // New fields
    release     string
    microsprint string
    backlog     bool
}
```

### Estimated Effort

- ~50 lines of code (extends existing infrastructure)
- No new command, just new flags

---

## Decision

- [ ] Approved - Proceed to implementation
- [ ] Approved with modifications
- [ ] Rejected
- [ ] Needs more information

---

## Prerequisites

- #347 - Capture stderr in git subprocess calls (required for good error messages when `git checkout -b` fails)

---

## References

- Current `patch` implementation: `cmd/patch.go`
- Current `release` implementation: `cmd/release.go`
- Track parsing: `cmd/release.go:1017-1045` (`parseReleaseTitle`)
- Commit removing `--track`: `cc82a48`
