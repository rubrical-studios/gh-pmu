# Proposal: Unified Release with Branch Flag

**Issue:** #346
**Date:** 2025-12-19
**Status:** Draft

---

## Executive Summary

Two changes to simplify gh pmu:

1. **Unified Release Command:** Consolidate `release` and `patch` command groups into a single `release` command with a `--branch` flag. The branch name is used literally (no parsing). This eliminates redundant code and simplifies the API.

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

Branch names are not validated or parsed - users choose their own naming conventions. The `--branch` value is used literally for:
- Git branch name
- Tracker issue title
- Release field value
- Artifact directory path

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
2. Use branch value literally (no parsing)
3. Create git branch on start
4. Deprecation warning when using `patch` commands

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
# Creates: git branch "release/v2.0.0", tracker "Release: release/v2.0.0"

gh pmu release add 123
# Sets Release field to "release/v2.0.0"

gh pmu release close --tag
# Creates: tag v2.0.0, artifacts in Releases/release/v2.0.0/
```

### Patch Workflow

```bash
gh pmu release start --branch "patch/v1.9.1"
# Creates: git branch "patch/v1.9.1", tracker "Release: patch/v1.9.1"

gh pmu release add 456
# Sets Release field to "patch/v1.9.1"

gh pmu release close --tag
# Creates: tag v1.9.1, artifacts in Releases/patch/v1.9.1/
```

### Hotfix Workflow

```bash
gh pmu release start --branch "hotfix-auth-bypass"
# Creates: git branch "hotfix-auth-bypass", tracker "Release: hotfix-auth-bypass"

gh pmu release add 789
# Sets Release field to "hotfix-auth-bypass"

gh pmu release close
# No tag, artifacts in Releases/hotfix-auth-bypass/
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

## Decisions

| Question | Decision |
|----------|----------|
| Interactive mode | **No** - Require `--branch` flag |
| `--version` flag | **Remove** - Only `--branch` flag, cleaner API |
| `--track` flag | **Remove** - Already done (#345) |
| Branch parsing | **None** - Branch value used literally, no track detection |
| Track-based validation | **Remove** - No automatic label validation |
| Git branch creation | **Always** - `git checkout -b {branch}` on start |
| Branch validation | **None** - User controls naming, conventions in docs |

---

## Branch Naming (Documentation Only)

No validation enforced. Recommended conventions documented in `docs/workflows.md`:

| Type | Branch Pattern | Tag Pattern | Example |
|------|----------------|-------------|---------|
| Release | `release/vX.Y.Z` | `vX.Y.Z` | `release/v2.0.0` |
| Pre-release | `release/vX.Y.Z-suffix` | `vX.Y.Z-suffix` | `release/v2.0.0-beta.1` |
| Patch | `patch/vX.Y.Z` | `vX.Y.Z` | `patch/v1.9.1` |
| Hotfix (versioned) | `hotfix/vX.Y.Z` | `vX.Y.Z` | `hotfix/v1.9.2` |
| Hotfix (named) | `hotfix-name` | - | `hotfix-auth-bypass` |

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

- ~~#347 - Capture stderr in git subprocess calls~~ ✅ Done (commit `00db4b4`)
- #345 - Remove --name and --track flags ✅ Done (code complete, docs pending)

---

## References

- Current `patch` implementation: `cmd/patch.go` (~850 lines to remove)
- Current `release` implementation: `cmd/release.go`
- Commit removing `--track`: `cc82a48`
- Commit removing `--name`: (part of #345)
