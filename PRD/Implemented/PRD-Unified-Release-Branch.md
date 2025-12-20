# Product Requirements Document: Unified Release with Branch Flag

**Version:** 1.0
**Date:** 2025-12-19
**Author:** User
**Status:** Draft
**Issue:** #346

---

## Product Vision

### Vision Statement
Simplify release management by consolidating `release` and `patch` commands into a unified interface with explicit branch control.

### Target Users
Developers using gh-pmu for project management with release-based workflows.

### Key Value Proposition
- Eliminate ~1750 lines of duplicate code
- Reduce CLI surface area (6 commands removed)
- Enable flexible branching strategies without parsing logic
- Support framework-level `Transfer-Issue` command

---

## Feature Areas

### Epic 1: Unified Release Command

**Description:** Replace `release start --version` and entire `patch` command group with `release start --branch`.

**Capabilities:**

1. **Add `--branch` flag to `release start`**
   - Accept branch name as literal value (no parsing)
   - Create git branch on start: `git checkout -b {branch}`
   - Use branch value for: tracker title, Release field, artifact directory
   - Examples: `release/v2.0.0`, `patch/v1.9.1`, `hotfix-auth-bypass`

2. **Remove `--version` flag from `release start`**
   - Breaking change: `--branch` required, no default
   - Migration: `--version 2.0.0` → `--branch release/v2.0.0`

3. **Remove entire `patch` command group**
   - Delete `cmd/patch.go` (~850 lines)
   - Delete `cmd/patch_test.go` (~900 lines)
   - Migration: `patch start --version 1.9.1` → `release start --branch patch/v1.9.1`

4. **Update artifact path generation**
   - Use branch value directly: `Releases/{branch}/`
   - Examples: `Releases/release/v2.0.0/`, `Releases/patch/v1.9.1/`

**Success Criteria:**
- `gh pmu release start --branch X` creates branch and tracker
- `gh pmu patch` commands return "unknown command" error
- All existing release tests pass with updated syntax
- ~1750 lines of code removed

---

### Epic 2: Extended Move Command

**Description:** Add flags to move issues between releases and sprints, supporting framework-level `Transfer-Issue`.

**Capabilities:**

1. **Add `--release` flag**
   - Set Release field to specified value
   - Example: `gh pmu move 45 --release "release/v2.0.0"`

2. **Add `--microsprint` flag (alias: `--sprint`)**
   - Set Microsprint field to specified value
   - Example: `gh pmu move 45 --microsprint "auth-work"`

3. **Add `--backlog` flag**
   - Clear Release and Microsprint fields
   - Example: `gh pmu move 45 --backlog`
   - Validation: Cannot combine with `--release` or `--microsprint`

4. **Combine with existing flags**
   - All flags remain optional
   - Example: `gh pmu move 45 --status in_progress --release "release/v2.0.0"`

**Success Criteria:**
- `--release` sets Release field correctly
- `--microsprint` sets Microsprint field correctly
- `--backlog` clears both fields
- `--backlog` errors if combined with `--release` or `--microsprint`
- Existing move functionality unchanged

---

## Non-Functional Expectations

| Category | Expectation |
|----------|-------------|
| Performance | No change - same API call patterns |
| Backward Compatibility | Breaking change - requires migration |
| Code Quality | Net reduction of ~1700 lines |

---

## Constraints

- Breaking change release (v0.8.0 or v1.0.0)
- Requires documentation updates
- Requires `.claude/rules/02-github-workflow.md` update in process-docs

---

## Dependencies

| Dependency | Status |
|------------|--------|
| #345 - Remove `--name` and `--track` flags | Done (v0.7.4) |
| #347 - Capture stderr in git subprocess calls | Done |

---

## Testing Approach

| Approach | Status | Notes |
|----------|--------|-------|
| TDD | Required | Update existing release tests, remove patch tests |
| ATDD | [ ] Not Used | |
| BDD | [ ] Not Used | |

---

## Migration Guide

| Before | After |
|--------|-------|
| `gh pmu release start --version 2.0.0` | `gh pmu release start --branch release/v2.0.0` |
| `gh pmu patch start --version 1.9.1` | `gh pmu release start --branch patch/v1.9.1` |
| `gh pmu patch add 42` | `gh pmu release add 42` |
| `gh pmu patch close --tag` | `gh pmu release close --tag` |

---

## Implementation Order

1. Add `--branch` flag to `release start` with git branch creation
2. Add `--release`, `--microsprint`, `--backlog` flags to `move`
3. Remove `--version` flag from `release start`
4. Remove `cmd/patch.go` and `cmd/patch_test.go`
5. Update documentation

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-12-19 | User | Initial draft from PROPOSAL-Unified-Release-Branch.md |
