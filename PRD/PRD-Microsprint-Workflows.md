# Microsprint and Deployment Workflows - Product Requirements Document

**Version:** 1.0
**Date:** 2025-12-13
**Author:** PRD generated from Proposal/Microsprint-Support-gh-pmu.md
**Status:** Ready
**Framework:** IDPF-Structured (Moderate)

---

## Executive Summary

### Problem Statement

GitHub Projects Iterations are designed for traditional sprint cycles (1-4 weeks minimum). With AI-assisted development compressing work into hours rather than weeks, teams need a way to batch related work at a higher cadence. Additionally, IDPF-Structured and IDPF-LTS projects need version-based batching (releases and patches) rather than time-boxed iterations.

### Solution Overview

Add three workflow types to gh-pmu:

1. **Microsprints** - Hour-scale work batches for AI-assisted development (IDPF-Agile)
2. **Release workflows** - Version-based deployment batching (IDPF-Structured)
3. **Patch workflows** - Hotfix deployment batching (IDPF-LTS)

Each workflow uses tracker issues for state management and Text fields as the source of truth for issue assignment.

### Target Users

- Solo developers using AI-assisted development
- Small teams needing higher-cadence work batching
- Projects using IDPF-Structured or IDPF-LTS frameworks

### Success Criteria

- All Phase 1 acceptance criteria passing
- All Phase 2 acceptance criteria passing
- Code coverage targets met (50% for `cmd`, 80% for other new packages)
- Commands integrate seamlessly with existing gh-pmu patterns

---

## Stakeholders

| Role | Responsibility |
|------|----------------|
| Developer | Implementation, testing |
| Framework User | Feedback, validation |

---

## Scope

### In Scope

**Phase 1 (Milestone 1):**
- Microsprint command group (start, current, close, list, add, resolve)
- Tracker issue lifecycle management
- Auto-generated naming (YYYY-MM-DD-a pattern)
- Text field updates for issue assignment
- Review and retrospective artifact generation
- Team-wide microsprint model
- Integration with existing move/create commands
- Conflict resolution

**Phase 2 (Milestone 2):**
- Release command group (start, current, add, close, list)
- Patch command group (start, current, add, close, list)
- Framework detection and command restrictions
- Git tag integration
- Version validation (semver)
- Release notes and patch notes generation

### Out of Scope (Future)

- Cross-repo microsprints (multiple repositories in one project)
- Parallel microsprints (multiple concurrent per team)
- Velocity metrics and analytics
- Time tracking integration
- CI/CD automation hooks
- Daily/weekly rollup reports

---

## Milestones

### Milestone 1: Microsprint Core (Phase 1)

**Requirements:** REQ-001 through REQ-016, REQ-038

**Deliverables:**
- `cmd/microsprint.go` - Command group implementation
- `internal/tracker/tracker.go` - Tracker issue management
- Artifact generation in `Microsprints/` directory
- Unit tests for all commands

### Milestone 2: Deployment Workflows (Phase 2)

**Requirements:** REQ-017 through REQ-037, REQ-039, REQ-040

**Deliverables:**
- `cmd/release.go` - Release command group
- `cmd/patch.go` - Patch command group
- `internal/artifacts/release.go` - Release notes generation
- `internal/artifacts/patch.go` - Patch notes generation
- Framework detection logic
- Unit tests for all commands

---

## Functional Requirements

### Milestone 1: Microsprint Core

#### REQ-001: Start Microsprint

**Description:** `gh pmu microsprint start` creates a tracker issue with `microsprint` label.

**Source:** Proposal AC-1

**Priority:** Must Have

**Acceptance Criteria:**
- AC-001-1: Given no active microsprint, When user runs `microsprint start`, Then tracker issue created with title "Microsprint: YYYY-MM-DD-a" and label `microsprint`
- AC-001-2: Given no active microsprint, When user runs `microsprint start --name "auth"`, Then tracker issue created with title "Microsprint: YYYY-MM-DD-a-auth"
- AC-001-3: Given tracker issue created, Then it is assigned to the current authenticated user
- AC-001-4: Given tracker issue created, Then status is set to In Progress

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-002: Auto-Generated Naming

**Description:** Microsprint names follow `YYYY-MM-DD-a` pattern with auto-increment.

**Source:** Proposal AC-2

**Priority:** Must Have

**Acceptance Criteria:**
- AC-002-1: Given no microsprints today, When starting microsprint, Then name is `YYYY-MM-DD-a`
- AC-002-2: Given microsprint `2025-12-13-a` exists, When starting new microsprint, Then name is `2025-12-13-b`
- AC-002-3: Given microsprint `2025-12-13-z` exists, When starting new microsprint, Then name is `2025-12-13-aa` (continues through `zz` for 702 additional slots)

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-003: Add Issue to Microsprint

**Description:** `gh pmu microsprint add <issue>` assigns issue to current microsprint.

**Source:** Proposal AC-3

**Priority:** Must Have

**Acceptance Criteria:**
- AC-003-1: Given active microsprint, When user runs `microsprint add 42`, Then Microsprint Text field on issue #42 is set to microsprint name
- AC-003-2: Given active microsprint, When field updated, Then output confirms "Added #42 to microsprint YYYY-MM-DD-a"
- AC-003-3: Given Text field update, Then tracker issue body is NOT updated (avoid race conditions)

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-004: Close Microsprint with Artifacts

**Description:** `gh pmu microsprint close` generates review.md and retro.md, stages to git.

**Source:** Proposal AC-4

**Priority:** Must Have

**Acceptance Criteria:**
- AC-004-1: Given active microsprint with issues, When user runs `microsprint close`, Then `Microsprints/{name}/review.md` generated with issue summary
- AC-004-2: Given `microsprint close`, When user completes retro prompts, Then `Microsprints/{name}/retro.md` generated
- AC-004-3: Given `microsprint close --skip-retro`, Then retro.md generated with empty template (no prompts)
- AC-004-4: Given artifacts generated, Then files staged to git (`git add`)
- AC-004-5: Given `microsprint close --commit`, Then artifacts committed with standard message
- AC-004-6: Given close complete, Then tracker issue body updated with artifact links and closed

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-005: Integration with Move Command

**Description:** `--microsprint` flag works with `gh pmu move`.

**Source:** Proposal AC-5

**Priority:** Must Have

**Acceptance Criteria:**
- AC-005-1: Given `gh pmu move 42 --microsprint 2025-12-13-a`, Then Microsprint field set to specified value
- AC-005-2: Given `gh pmu move 42 --microsprint current`, Then Microsprint field set to active microsprint name
- AC-005-3: Given `--microsprint current` with no active microsprint, Then error with suggestion to run `microsprint start`

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-006: Integration with Create Command

**Description:** `--microsprint` flag works with `gh pmu create`.

**Source:** Proposal AC-5

**Priority:** Must Have

**Acceptance Criteria:**
- AC-006-1: Given `gh pmu create --title "Bug fix" --microsprint current`, Then issue created and Microsprint field set to active microsprint
- AC-006-2: Given no active microsprint, When using `--microsprint current`, Then error with suggestion

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-007: Tracker Issue Per Microsprint

**Description:** Each microsprint gets its own tracker issue, closed on microsprint close.

**Source:** Proposal AC-6

**Priority:** Must Have

**Acceptance Criteria:**
- AC-007-1: Given `microsprint start`, Then new tracker issue created (not reused)
- AC-007-2: Given `microsprint close`, Then tracker issue is closed
- AC-007-3: Given tracker issue, Then it has `microsprint` label for filtering

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-008: List Microsprint History

**Description:** `gh pmu microsprint list` queries tracker issues by label.

**Source:** Proposal AC-7

**Priority:** Must Have

**Acceptance Criteria:**
- AC-008-1: Given `microsprint list`, Then table displayed with: Microsprint, Tracker#, Issues, Done, Duration, Status
- AC-008-2: Given multiple microsprints, Then sorted by date descending (most recent first)
- AC-008-3: Given `--limit N` flag, Then only N most recent shown

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-009: Team-Wide Model

**Description:** One active microsprint at a time, with join/wait prompt for concurrent starts.

**Source:** Proposal AC-8

**Priority:** Must Have

**Acceptance Criteria:**
- AC-009-1: Given active microsprint by alice, When bob runs `microsprint start`, Then prompt displayed: "Microsprint active (started by alice). Join / Work without / Cancel"
- AC-009-2: Given bob selects "Join", Then bob continues without creating new microsprint
- AC-009-3: Given bob selects "Work without", Then bob works without microsprint assignment
- AC-009-4: Given bob tries to close alice's microsprint, Then confirmation prompt: "Close microsprint started by alice? [y/N]"

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-010: Review Generation

**Description:** Auto-generate `Microsprints/{name}/review.md` from issue data.

**Source:** Proposal AC-9

**Priority:** Must Have

**Acceptance Criteria:**
- AC-010-1: Given microsprint with issues, When closing, Then review.md contains: elapsed time, tracker issue number, issue counts by status
- AC-010-2: Given review.md, Then issues grouped by status (Done, In Review, other)
- AC-010-3: Given issue data, Then each issue shows: number, title, status

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-011: Retrospective Prompts

**Description:** Prompt user for retrospective input, generate `Microsprints/{name}/retro.md`.

**Source:** Proposal AC-10

**Priority:** Must Have

**Acceptance Criteria:**
- AC-011-1: Given `microsprint close`, Then user prompted for: What Went Well, What Could Be Improved, Action Items
- AC-011-2: Given user input, Then retro.md generated with responses
- AC-011-3: Given `--skip-retro` flag, Then empty template generated without prompts

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-012: Tracker Naming Validation

**Description:** Tracker issues must match naming pattern; non-matching issues ignored.

**Source:** Proposal AC-11

**Priority:** Must Have

**Acceptance Criteria:**
- AC-012-1: Given issue with `microsprint` label but title "Random Issue", Then ignored by microsprint commands
- AC-012-2: Given issue with title "Microsprint: 2025-12-13-a", Then recognized as tracker issue
- AC-012-3: Given pattern `^Microsprint: \d{4}-\d{2}-\d{2}-[a-z]+(-.*)?$`, Then validation applied

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-013: Multiple Active Detection

**Description:** Commands fail with error if multiple open tracker issues exist.

**Source:** Proposal AC-12

**Priority:** Must Have

**Acceptance Criteria:**
- AC-013-1: Given 2+ open tracker issues, When running `microsprint add`, Then error: "Multiple active microsprints detected. Run `microsprint resolve`"
- AC-013-2: Given 2+ open tracker issues, When running `microsprint close`, Then same error
- AC-013-3: Given error, Then user directed to `microsprint resolve` command

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-014: Resolve Conflicts

**Description:** `gh pmu microsprint resolve` prompts user to keep one, merge, or abandon.

**Source:** Proposal AC-13

**Priority:** Must Have

**Acceptance Criteria:**
- AC-014-1: Given 2 open tracker issues, When running `microsprint resolve`, Then prompt: Keep #1, Keep #2, Close both (new microsprint), Abandon both
- AC-014-2: Given "Keep #1" selected, Then #2 closed, its issues reassigned to #1's microsprint
- AC-014-3: Given "Abandon both" selected, Then both closed without reassignment
- AC-014-4: Given only 1 open tracker issue, When running `resolve`, Then message: "No conflict to resolve"

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-015: No Active Microsprint Error

**Description:** `microsprint add` without active microsprint returns helpful error.

**Source:** Proposal AC-14

**Priority:** Must Have

**Acceptance Criteria:**
- AC-015-1: Given no active microsprint, When running `microsprint add 42`, Then error: "No active microsprint. Run `gh pmu microsprint start` first."

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-016: Empty Microsprint Close

**Description:** `microsprint close` with no issues generates review noting "No issues completed".

**Source:** Proposal AC-15

**Priority:** Must Have

**Acceptance Criteria:**
- AC-016-1: Given active microsprint with 0 issues, When running `microsprint close`, Then review.md generated with "No issues completed in this microsprint"
- AC-016-2: Given empty microsprint, Then close succeeds (not an error)

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-038: Remove Issue from Microsprint

**Description:** `gh pmu microsprint remove <issue>` clears the Microsprint Text field on the issue.

**Source:** PRD evaluation (gap analysis)

**Priority:** Must Have

**Acceptance Criteria:**
- AC-038-1: Given issue #42 assigned to microsprint, When running `microsprint remove 42`, Then Microsprint Text field cleared (set to empty)
- AC-038-2: Given field cleared, Then output confirms "Removed #42 from microsprint YYYY-MM-DD-a"
- AC-038-3: Given issue not in any microsprint, When running `microsprint remove 42`, Then warning: "Issue #42 is not assigned to a microsprint"

**Testing:** TDD (unit tests, no integration tests)

---

### Milestone 2: Deployment Workflows

#### REQ-017: Start Release

**Description:** `gh pmu release start --version X.Y.Z` creates tracker issue with `release` label.

**Source:** Proposal AC-17

**Priority:** Must Have

**Acceptance Criteria:**
- AC-017-1: Given `release start --version 1.2.0`, Then tracker issue created: "Release: v1.2.0"
- AC-017-2: Given `release start --version 1.2.0 --name "Phoenix"`, Then tracker issue: "Release: v1.2.0 (Phoenix)"
- AC-017-3: Given tracker issue created, Then has `release` label
- AC-017-4: Given active release exists, When running `release start`, Then error: "Active release exists"

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-018: Version Validation

**Description:** Semver format enforced; duplicate versions rejected.

**Source:** Proposal AC-18

**Priority:** Must Have

**Acceptance Criteria:**
- AC-018-1: Given `release start --version 1.2.0`, Then accepted (valid semver)
- AC-018-2: Given `release start --version 1.2`, Then error: "Invalid version format. Use semver: X.Y.Z"
- AC-018-3: Given `release start --version v1.2.0`, Then accepted (v prefix allowed)
- AC-018-4: Given closed release v1.2.0 exists, When starting v1.2.0, Then error: "Version v1.2.0 already released"

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-019: Add Issue to Release

**Description:** `gh pmu release add <issue>` sets Release Text field.

**Source:** Proposal AC-19

**Priority:** Must Have

**Acceptance Criteria:**
- AC-019-1: Given active release v1.2.0, When running `release add 42`, Then Release field on #42 set to "v1.2.0"
- AC-019-2: Given issue added, Then output: "Added #42 to release v1.2.0"

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-020: Release Artifacts

**Description:** `gh pmu release close` generates release-notes.md and changelog.md.

**Source:** Proposal AC-20

**Priority:** Must Have

**Acceptance Criteria:**
- AC-020-1: Given `release close`, Then `Releases/v1.2.0/release-notes.md` generated
- AC-020-2: Given release-notes.md, Then contains: date, codename (if set), tracker issue, issues grouped by label (Features, Bug Fixes, etc.)
- AC-020-3: Given `release close`, Then `Releases/v1.2.0/changelog.md` generated (append-friendly format)
- AC-020-4: Given artifacts, Then staged to git

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-021: Release Git Tag

**Description:** `gh pmu release close --tag` creates annotated git tag.

**Source:** Proposal AC-21

**Priority:** Must Have

**Acceptance Criteria:**
- AC-021-1: Given `release close --tag`, Then `git tag -a v1.2.0 -m "Release v1.2.0"` executed
- AC-021-2: Given tag created, Then NOT pushed (user controls push timing)
- AC-021-3: Given `release close` (no --tag), Then no tag created

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-022: List Releases

**Description:** `gh pmu release list` shows release history from tracker issues.

**Source:** Proposal AC-22

**Priority:** Must Have

**Acceptance Criteria:**
- AC-022-1: Given `release list`, Then table: Version, Codename, Tracker#, Issues, Date, Status
- AC-022-2: Given multiple releases, Then sorted by version descending

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-023: Start Patch

**Description:** `gh pmu patch start --version X.Y.Z` creates tracker issue with `patch` label.

**Source:** Proposal AC-23

**Priority:** Must Have

**Acceptance Criteria:**
- AC-023-1: Given `patch start --version 1.1.5`, Then tracker issue created: "Patch: v1.1.5"
- AC-023-2: Given tracker issue created, Then has `patch` label
- AC-023-3: Given active patch exists, When running `patch start`, Then error

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-024: LTS Constraints

**Description:** Warning if non-bug issues added; error if breaking-change label detected.

**Source:** Proposal AC-24

**Priority:** Must Have

**Acceptance Criteria:**
- AC-024-1: Given issue without `bug` or `hotfix` label, When running `patch add`, Then warning: "Issue #42 is not labeled bug/hotfix"
- AC-024-2: Given issue with `breaking-change` label, When running `patch add`, Then error: "Breaking changes not allowed in patches"
- AC-024-3: Given bug-labeled issue, When running `patch add`, Then no warning

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-025: Add Issue to Patch

**Description:** `gh pmu patch add <issue>` sets Patch Text field.

**Source:** Proposal AC-25

**Priority:** Must Have

**Acceptance Criteria:**
- AC-025-1: Given active patch v1.1.5, When running `patch add 42`, Then Patch field set to "v1.1.5"
- AC-025-2: Given warning condition (REQ-024), Then warning displayed but field still set

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-026: Patch Artifacts

**Description:** `gh pmu patch close` generates patch-notes.md.

**Source:** Proposal AC-26

**Priority:** Must Have

**Acceptance Criteria:**
- AC-026-1: Given `patch close`, Then `Patches/v1.1.5/patch-notes.md` generated
- AC-026-2: Given patch-notes.md, Then contains: date, tracker issue, severity, fixes list, affected components
- AC-026-3: Given artifacts, Then staged to git

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-027: Patch Git Tag

**Description:** `gh pmu patch close --tag` creates annotated git tag.

**Source:** Proposal AC-27

**Priority:** Must Have

**Acceptance Criteria:**
- AC-027-1: Given `patch close --tag`, Then `git tag -a v1.1.5 -m "Patch v1.1.5"` executed
- AC-027-2: Given tag created, Then NOT pushed

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-028: List Patches

**Description:** `gh pmu patch list` shows patch history.

**Source:** Proposal AC-28

**Priority:** Must Have

**Acceptance Criteria:**
- AC-028-1: Given `patch list`, Then table: Version, Tracker#, Issues, Date, Status
- AC-028-2: Given multiple patches, Then sorted by version descending

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-039: Remove Issue from Release

**Description:** `gh pmu release remove <issue>` clears the Release Text field on the issue.

**Source:** PRD evaluation (gap analysis)

**Priority:** Must Have

**Acceptance Criteria:**
- AC-039-1: Given issue #42 assigned to release, When running `release remove 42`, Then Release Text field cleared (set to empty)
- AC-039-2: Given field cleared, Then output confirms "Removed #42 from release vX.Y.Z"
- AC-039-3: Given issue not in any release, When running `release remove 42`, Then warning: "Issue #42 is not assigned to a release"

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-040: Remove Issue from Patch

**Description:** `gh pmu patch remove <issue>` clears the Patch Text field on the issue.

**Source:** PRD evaluation (gap analysis)

**Priority:** Must Have

**Acceptance Criteria:**
- AC-040-1: Given issue #42 assigned to patch, When running `patch remove 42`, Then Patch Text field cleared (set to empty)
- AC-040-2: Given field cleared, Then output confirms "Removed #42 from patch vX.Y.Z"
- AC-040-3: Given issue not in any patch, When running `patch remove 42`, Then warning: "Issue #42 is not assigned to a patch"

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-029: Framework Detection

**Description:** Auto-detect framework from config files.

**Source:** Proposal AC-29

**Priority:** Must Have

**Acceptance Criteria:**
- AC-029-1: Given `.gh-pmu.yml` with `framework: IDPF-Agile`, Then framework detected as Agile
- AC-029-2: Given `framework-config.json` with `processFramework: "IDPF-Structured"`, Then framework detected
- AC-029-3: Given no framework config, Then no framework restriction applied

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-030: Command Restrictions

**Description:** Inapplicable commands error with helpful suggestions when framework detected.

**Source:** Proposal AC-30

**Priority:** Must Have

**Acceptance Criteria:**
- AC-030-1: Given IDPF-Structured framework, When running `microsprint start`, Then error: "Microsprint commands not applicable for IDPF-Structured. Use `gh pmu release start --version X.Y.Z`"
- AC-030-2: Given IDPF-LTS framework, When running `microsprint start`, Then error suggesting `patch` command
- AC-030-3: Given IDPF-Agile framework, When running `patch start`, Then error suggesting `microsprint` command

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-031: No Framework Fallback

**Description:** If no framework detected, all commands available without restrictions.

**Source:** Proposal AC-31

**Priority:** Must Have

**Acceptance Criteria:**
- AC-031-1: Given no framework config, When running any workflow command, Then command executes without restriction
- AC-031-2: Given no framework, Then microsprint, release, and patch commands all available

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-032: Unit Tests

**Description:** All command handlers have unit tests with mocked API calls.

**Source:** Proposal AC-32

**Priority:** Must Have

**Acceptance Criteria:**
- AC-032-1: Given each command handler, Then unit test file exists
- AC-032-2: Given tests, Then API calls are mocked (no real GitHub API)
- AC-032-3: Given tests, Then table-driven pattern used where appropriate

**Testing:** Test coverage verification

---

#### REQ-033: Error Path Coverage

**Description:** All error conditions have tests.

**Source:** Proposal AC-34

**Priority:** Must Have

**Acceptance Criteria:**
- AC-033-1: Given "no active microsprint" error, Then test exists
- AC-033-2: Given "multiple active" error, Then test exists
- AC-033-3: Given "invalid version" error, Then test exists
- AC-033-4: Given "breaking change in patch" error, Then test exists

**Testing:** Test coverage verification

---

#### REQ-034: Coverage Target

**Description:** Minimum code coverage targets: 50% for `cmd` package, 80% for other new packages.

**Source:** Proposal AC-35

**Priority:** Must Have

**Acceptance Criteria:**
- AC-034-1: Given `go test -cover`, Then `cmd` package shows >= 50% coverage
- AC-034-2: Given `go test -cover`, Then other new packages show >= 80% coverage
- AC-034-3: Given coverage report, Then critical paths covered

**Testing:** Coverage measurement

---

#### REQ-035: View Current Microsprint

**Description:** `gh pmu microsprint current` shows active microsprint details.

**Source:** Proposal (implied from command list)

**Priority:** Must Have

**Acceptance Criteria:**
- AC-035-1: Given active microsprint, When running `microsprint current`, Then displays: name, started time, issue count, tracker issue number
- AC-035-2: Given no active microsprint, Then message: "No active microsprint"
- AC-035-3: Given `--refresh` flag, Then tracker issue body updated with current issue list

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-036: View Current Release

**Description:** `gh pmu release current` shows active release details.

**Source:** Technical Architecture (implied from command list)

**Priority:** Must Have

**Acceptance Criteria:**
- AC-036-1: Given active release, When running `release current`, Then displays: version, codename (if set), started time, issue count, tracker issue number
- AC-036-2: Given no active release, Then message: "No active release"
- AC-036-3: Given `--refresh` flag, Then tracker issue body updated with current issue list

**Testing:** TDD (unit tests, no integration tests)

---

#### REQ-037: View Current Patch

**Description:** `gh pmu patch current` shows active patch details.

**Source:** Technical Architecture (implied from command list)

**Priority:** Must Have

**Acceptance Criteria:**
- AC-037-1: Given active patch, When running `patch current`, Then displays: version, started time, issue count, tracker issue number
- AC-037-2: Given no active patch, Then message: "No active patch"
- AC-037-3: Given `--refresh` flag, Then tracker issue body updated with current issue list

**Testing:** TDD (unit tests, no integration tests)

---

## Technical Architecture

### New Files

| File | Purpose |
|------|---------|
| `cmd/microsprint.go` | Command group: start, current, add, remove, close, list, resolve |
| `cmd/release.go` | Command group: start, current, add, remove, close, list |
| `cmd/patch.go` | Command group: start, current, add, remove, close, list |
| `internal/tracker/tracker.go` | Tracker issue state management |
| `internal/artifacts/microsprint.go` | Review and retro generation |
| `internal/artifacts/release.go` | Release notes generation |
| `internal/artifacts/patch.go` | Patch notes generation |

### Modified Files

| File | Change |
|------|--------|
| `cmd/root.go` | Add command registrations |
| `cmd/move.go` | Add `--microsprint` flag |
| `cmd/create.go` | Add `--microsprint` flag |
| `internal/config/config.go` | Add workflow config structs |
| `internal/api/mutations.go` | Add `UpdateTextField()` |
| `internal/api/queries.go` | Add tracker issue queries |

### Configuration Extension

```yaml
# .gh-pmu.yml additions
framework: IDPF-Agile  # Optional framework declaration

fields:
  microsprint:
    field: Microsprint
    type: text

microsprint:
  stale_threshold_hours: 24  # Default
```

---

## Constraints

| ID | Constraint | Impact |
|----|------------|--------|
| CON-001 | Git repository required | Artifact commands fail outside git repos |
| CON-002 | Text fields only | Avoids Single Select option pollution |
| CON-003 | One active tracker per workflow | Simplifies state management |
| CON-004 | No integration tests | Per stakeholder decision |

---

## Testing Approach

- **TDD:** Required for all development
- **Unit Tests:** Standard Go `testing` package
- **Table-Driven:** Where appropriate (existing pattern)
- **Mocked API:** No real GitHub API calls in tests
- **Integration Tests:** Not required
- **Coverage Target:** 50% for `cmd` package, 80% for other new packages

---

## Framework Handoff

**Selected Framework:** IDPF-Structured

**Rationale:** Fixed requirements from detailed proposal, clear scope, milestoned delivery.

**TDD Order (Milestone 1):**
1. REQ-001: Start Microsprint
2. REQ-002: Auto-Generated Naming
3. REQ-007: Tracker Issue Per Microsprint
4. REQ-003: Add Issue to Microsprint
5. REQ-038: Remove Issue from Microsprint
6. REQ-035: View Current Microsprint
7. REQ-004: Close Microsprint with Artifacts
8. REQ-010: Review Generation
9. REQ-011: Retrospective Prompts
10. REQ-008: List Microsprint History
11. REQ-005: Integration with Move Command
12. REQ-006: Integration with Create Command
13. REQ-009: Team-Wide Model
14. REQ-012: Tracker Naming Validation
15. REQ-013: Multiple Active Detection
16. REQ-014: Resolve Conflicts
17. REQ-015: No Active Microsprint Error
18. REQ-016: Empty Microsprint Close

**TDD Order (Milestone 2):**
19. REQ-029: Framework Detection
20. REQ-030: Command Restrictions
21. REQ-031: No Framework Fallback
22. REQ-017: Start Release
23. REQ-018: Version Validation
24. REQ-019: Add Issue to Release
25. REQ-039: Remove Issue from Release
26. REQ-036: View Current Release
27. REQ-020: Release Artifacts
28. REQ-021: Release Git Tag
29. REQ-022: List Releases
30. REQ-023: Start Patch
31. REQ-024: LTS Constraints
32. REQ-025: Add Issue to Patch
33. REQ-040: Remove Issue from Patch
34. REQ-037: View Current Patch
35. REQ-026: Patch Artifacts
36. REQ-027: Patch Git Tag
37. REQ-028: List Patches
38. REQ-032: Unit Tests
39. REQ-033: Error Path Coverage
40. REQ-034: Coverage Target

---

## Approval

| Role | Name | Date | Approved |
|------|------|------|----------|
| Stakeholder | | | [ ] |

---

**PRD Status: READY - Pending approval for IDPF-Structured development**
