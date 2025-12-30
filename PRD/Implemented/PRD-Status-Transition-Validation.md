# Product Requirements Document: Status Transition Validation

**Version:** 1.0
**Date:** 2025-12-20
**Author:** User
**Status:** Draft
**Epic:** #364
**Source:** Proposal/PROPOSAL-Status-Transition-Validation.md

---

## Product Vision

### Vision Statement
Enforce IDPF workflow discipline at the CLI level by validating status transitions, ensuring issues have proper release assignments, body content, and completed acceptance criteria before progressing through the workflow.

### Target Users
Developers and teams using gh-pmu with IDPF workflows who need guardrails to prevent incomplete or untracked work.

### Key Value Proposition
- Prevent issues from being marked complete with unfinished acceptance criteria
- Ensure all active work is assigned to a known release for tracking
- Enforce documentation requirements (body content) before review/completion
- Framework-aware: users not using IDPF can opt out with `framework: none`

---

## Feature Areas

### Epic 1: Framework Configuration

**Description:** Add framework awareness to gh-pmu, allowing IDPF validation to be enabled/disabled.

**Capabilities:**

1. **Add `framework` field to `.gh-pmu.yml`**
   - Values: `IDPF` (default) or `none`
   - `gh pmu init` prompts "Are you using the IDPF framework? (y/n)" for new projects
   - Existing configs without `framework` field are normalized on load (writes `IDPF` to file)

2. **Skip validation when `framework: none`**
   - All status transition rules bypassed
   - Release/Microsprint fields not auto-created on init
   - Release start does not require `--branch`

**Success Criteria:**
- `gh pmu init` prompts for framework choice (new projects only)
- Missing `framework` field triggers write of `IDPF` to config
- `framework: none` bypasses all validation rules

---

### Epic 2: Release-Gated Progression

**Description:** Issues cannot move from `backlog` to working statuses without a release assignment.

**Capabilities:**

1. **Validate release on transition**
   - `backlog → ready` requires Release field set
   - `backlog → in_progress` requires Release field set
   - Release name must exist in `releases.active[]` (typo prevention)

2. **Error messaging**
   ```
   Error: Issue #45 has no release assignment.
   Cannot move from 'backlog' to 'ready' without a release.

   Use: gh pmu move 45 --release "release/vX.Y.Z"
   ```

**Success Criteria:**
- `gh pmu move X --status ready` blocked without release
- `gh pmu move X --status in_progress` blocked without release
- Invalid release name (not in `active[]`) produces clear error
- Error message includes actionable fix command

---

### Epic 3: Body Required for Review/Done

**Description:** Issues cannot move to `in_review` or `done` without body content.

**Capabilities:**

1. **Validate body on transition**
   - `any → in_review` requires non-empty body
   - `any → done` requires non-empty body
   - Empty = empty string or whitespace-only (template headers pass)

2. **Error messaging**
   ```
   Error: Issue #45 has no body.
   Issues require acceptance criteria before moving to in_review.
   ```

**Success Criteria:**
- `gh pmu move X --status in_review` blocked if body is empty/whitespace
- `gh pmu move X --status done` blocked if body is empty/whitespace
- Issues with template headers (non-whitespace content) pass validation

---

### Epic 4: Checkbox Validation

**Description:** Issues cannot move to `in_review` or `done` with unchecked checkboxes.

**Capabilities:**

1. **Validate checkboxes on transition**
   - `any → in_review` blocked if `[ ]` exists in body
   - `any → done` blocked if `[ ]` exists in body
   - All checkboxes validated equally (including nested/indented)
   - Simple regex: accepts rare false positives in code blocks

2. **Error messaging**
   ```
   Error: Issue #45 has unchecked items:
     [ ] Write unit tests
     [ ] Update documentation

   Complete these items before moving to review.
   ```

3. **Force override**
   - `--force` flag bypasses checkbox validation only
   - Body and release requirements cannot be bypassed

**Success Criteria:**
- Unchecked checkboxes block transition to `in_review`/`done`
- Error lists all unchecked items
- `--force` allows bypass with warning
- `--force` does NOT bypass body or release requirements

---

### Epic 5: Batch and Recursive Validation

**Description:** Validation applies to bulk operations and recursive moves.

**Capabilities:**

1. **All-or-nothing batch validation**
   - `gh pmu move 1 2 3 --status done` validates ALL issues first
   - If any fail, entire operation blocked (no partial moves)
   - All issues batch-fetched in single GraphQL query for efficiency

2. **Recursive validation**
   - `gh pmu move X --status done --recursive` validates parent AND all sub-issues
   - Entire operation blocked if ANY issue fails
   - Error lists all failing issues

3. **Dry-run support**
   - `gh pmu move --dry-run` shows validation results without making changes

**Success Criteria:**
- Batch moves validate all before executing any
- Recursive moves validate entire tree
- Single validation failure blocks entire operation
- `--dry-run` shows pass/fail without changes

---

### Epic 6: Create Validation

**Description:** Validation rules apply to `gh pmu create` when creating issues in working statuses.

**Capabilities:**

1. **Validate on create**
   - `--status backlog` - no validation (default)
   - `--status ready` - requires `--release`
   - `--status in_progress` - requires `--release`
   - `--status in_review` - requires `--release` and body
   - `--status done` - requires `--release` and body

**Success Criteria:**
- Creating in `ready`/`in_progress` requires `--release`
- Creating in `in_review`/`done` requires `--release` and `--body`
- No requirement to start in `backlog` (direct creation allowed if requirements met)

---

### Epic 7: Active Release Discovery

**Description:** Init scans for active releases and stores them in config.

**Capabilities:**

1. **Discover releases on init**
   - Scan for open issues with "release" label
   - Parse release name from title ("Release: release/v2.0.0")
   - Store in `releases.active[]` with name and tracker issue number

2. **Config sync on release lifecycle**
   - `release start` adds to `releases.active[]`
   - `release close` removes from `releases.active[]`
   - `release reopen` adds back to `releases.active[]`

3. **Init merge behavior**
   - Re-running init preserves existing `framework` and `releases.active[]`
   - Newly discovered releases merged (additive, no duplicates)

**Success Criteria:**
- Init displays found active releases
- `releases.active[]` populated in `.gh-pmu.yml`
- Release lifecycle operations sync config atomically

---

### Epic 8: Release Close Enhancements

**Description:** Enhanced release close with confirmation and incomplete issue handling.

**Capabilities:**

1. **Close verification**
   - Shows release summary (issue count, done/not done)
   - Warns if issues are not done
   - Requires explicit release name (no implicit "current")

2. **Incomplete issue handling**
   - Auto-moves incomplete issues to backlog
   - Clears both Release AND Microsprint fields
   - `--yes` flag skips confirmation

3. **Reopen support**
   - `release reopen <name>` reopens tracker issue
   - Adds release back to `releases.active[]`

**Success Criteria:**
- `release close` shows confirmation with summary
- Incomplete issues moved to backlog with fields cleared
- `release reopen` restores release to active state

---

### Epic 9: Microsprint Release Context

**Description:** Microsprints require explicit release context.

**Capabilities:**

1. **Require release on start**
   - `microsprint start --release <name>` required
   - Release stored in tracker title: `[release/v2.0.0]`
   - Release stored in tracker body metadata

2. **Auto-update on add**
   - `microsprint add X` updates issue's release to match microsprint
   - Silent update (no confirmation)

**Success Criteria:**
- `microsprint start` without `--release` produces error
- Issue release auto-updated when added to microsprint

---

### Epic 10: List Filter Enhancement

**Description:** Add `--no-release` filter to `gh pmu list`.

**Capabilities:**

1. **Filter unassigned issues**
   - `gh pmu list --no-release` shows issues without Release field
   - Works with `--json` output
   - Combines with other filters (`--status`, `--priority`)

**Success Criteria:**
- `--no-release` filters to issues with empty Release field
- Works in combination with existing filters

---

## Non-Functional Expectations

| Category | Expectation |
|----------|-------------|
| Performance | Batch-fetch issues for bulk validation (single GraphQL query) |
| Backward Compatibility | Breaking change - validation enforced immediately on upgrade |
| Code Quality | Clear error messages with actionable fix commands |
| Scope | CLI-only enforcement - GitHub web UI and direct API bypass these rules |

---

## Constraints

- Breaking change for existing users without `framework` field (defaults to IDPF)
- No grace period - validation enforced immediately
- CLI-only enforcement (cannot prevent GitHub web UI or API bypasses)
- `--force` can only bypass checkbox validation (body/release always required)

---

## Dependencies

| Dependency | Status |
|------------|--------|
| `gh pmu move --release` flag | Implemented (PRD-Unified-Release-Branch) |
| Issue body access in move command | Existing (uses GitHub API) |
| Project item field values | Existing (GetProjectItems) |

---

## Testing Approach

| Approach | Status | Notes |
|----------|--------|-------|
| TDD | Required | Unit tests for validation functions |
| ATDD | [ ] Not Used | |
| BDD | [ ] Not Used | |

### Key Test Cases

1. **Release validation**
   - Issue without release blocked from `ready`/`in_progress`
   - Invalid release name (not in `active[]`) blocked
   - Issue with valid release passes

2. **Body validation**
   - Empty body blocked from `in_review`/`done`
   - Whitespace-only body blocked
   - Template headers (non-whitespace) pass

3. **Checkbox validation**
   - Unchecked `[ ]` blocks `in_review`/`done`
   - All checked `[x]` passes
   - `--force` bypasses checkbox only

4. **Batch/recursive**
   - Single failure blocks entire batch
   - Recursive validates all sub-issues

---

## Migration Guide

| Scenario | Action |
|----------|--------|
| Existing config without `framework` | Auto-set to `IDPF` on load (breaking) |
| Want to disable validation | Set `framework: none` in `.gh-pmu.yml` |
| Move to `ready` without release | Add `--release` flag or assign release first |
| Move to `done` with unchecked boxes | Complete items, remove them, or use `--force` |
| Move to `in_review` without body | Add body content to issue first |

---

## Implementation Order

1. Framework configuration (`framework` field, init prompt, normalize on load)
2. Active release discovery on init
3. Release-gated progression (Rule 1)
4. Body required validation (Rule 2)
5. Checkbox validation (Rules 3, 4)
6. `--force` flag for checkbox bypass
7. Batch validation (all-or-nothing)
8. Recursive validation
9. Create validation
10. `--no-release` list filter
11. Microsprint release context
12. Release close enhancements
13. `--dry-run` support

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-12-20 | User | Initial PRD from PROPOSAL-Status-Transition-Validation.md |
