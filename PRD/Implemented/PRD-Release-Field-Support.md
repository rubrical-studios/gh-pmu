# Release Field Support - Product Requirements Document

**Version:** 1.0
**Date:** 2025-12-14
**Author:** Claude (from Proposal)
**Status:** ✅ Implemented
**Source:** process-docs/Proposal/Implemented/Release-Field-Support.md
**Implemented:** 2025-12-14

---

## Executive Summary

### Problem Statement

The current `gh pmu release` implementation uses tracker issues to represent releases. While this provides a place for release documentation, it lacks visibility in GitHub Project views:

- Cannot filter Table view by release
- Cannot group issues by release
- Cannot create release-specific Kanban boards
- No at-a-glance release assignment visibility

Both Agile and Structured teams need to see which issues belong to which release directly in their project views.

### Solution Overview

Enhance release management by adding a **Release field** (text type) to GitHub Projects, complementing the existing tracker-issue approach. This hybrid model uses:

- **Tracker issues** for documentation/ceremony (release notes, changelog)
- **Release field** as source of truth for membership, filtering, grouping, and visibility

### Target Users

- Development teams using gh-pmu for project management
- Teams using IDPF-Agile or IDPF-Structured frameworks
- Teams managing multiple concurrent releases

### Success Criteria

- Release field visible in GitHub Project Table/Board views
- Multiple concurrent releases supported
- `gh pmu init` auto-creates required fields and labels
- Patch releases work via alias with constraints enforced

---

## Stakeholders

| Role | Responsibility |
|------|----------------|
| gh-pmu maintainers | Implementation in CLI tool |
| process-docs maintainers | Documentation and workflow updates |
| Framework users | Feedback, validation |

---

## Scope

### In Scope

- Release text field in GitHub Projects
- Multiple concurrent releases with tracks
- Track prefix naming convention
- `gh pmu release` commands (start, add, remove, close, current, list)
- `gh pmu patch` as alias for `release --track patch`
- `gh pmu init` enhancements (field/label creation)
- Artifact generation on release close
- Git tag creation matching Release field value
- Multi-repo project support (single primary repo)

### Out of Scope (v1.0)

- Single-select field type (text only)
- Field-only mode without tracker issues
- Full multi-repo support (labels in all repos)
- Automated release notes from AI summarization
- Integration with GitHub Releases API

---

## Functional Requirements

### REQ-001: Release Field Creation

**Description:** `gh pmu init` checks for and creates a Release text field in the GitHub Project if missing.

**Rationale:** Automated setup reduces manual configuration.

**Priority:** High

**Source:** Proposal - GitHub Project Setup section

**Acceptance Criteria:**
- AC-001-1: Given a project without Release field, When `gh pmu init` runs, Then Release text field is created
- AC-001-2: Given a project with existing Release field, When `gh pmu init` runs, Then field is not duplicated
- AC-001-3: Given field creation, Then `.gh-pmu.yml` is updated with field configuration

**Testing:** TDD

---

### REQ-002: Release Label Creation

**Description:** `gh pmu init` checks for and creates `release` and `microsprint` labels if missing.

**Rationale:** Labels identify tracker issues.

**Priority:** High

**Source:** Proposal - GitHub Project Setup section

**Acceptance Criteria:**
- AC-002-1: Given repository without `release` label, When `gh pmu init` runs, Then label is created
- AC-002-2: Given repository without `microsprint` label, When `gh pmu init` runs, Then label is created
- AC-002-3: Given existing labels, When `gh pmu init` runs, Then labels are not duplicated

**Testing:** TDD

---

### REQ-003: Release Start

**Description:** Users can start a new release with version and optional track.

**Rationale:** Initiates the release lifecycle.

**Priority:** High

**Source:** Proposal - Workflow Changes section

**Acceptance Criteria:**
- AC-003-1: Given valid semver version, When `gh pmu release start --version 1.2.0` runs, Then tracker issue created with `release` label
- AC-003-2: Given tracker issue created, Then title follows pattern `Release: {prefix}{version}`
- AC-003-3: Given release started, Then `release.active[]` in `.gh-pmu.yml` is updated
- AC-003-4: Given `--track` flag, Then appropriate prefix applied (e.g., `patch/1.1.1`)
- AC-003-5: Given duplicate version (active or history), Then error returned

**Testing:** TDD

---

### REQ-004: Release Start Interactive

**Description:** Users can start a release interactively with version suggestions from git tags.

**Rationale:** Simplified workflow for common cases.

**Priority:** Medium

**Source:** Proposal - Workflow Changes section

**Acceptance Criteria:**
- AC-004-1: Given `gh pmu release start` without version, When run, Then latest git tag is fetched
- AC-004-2: Given latest tag v1.1.0, Then menu displays Patch (v1.1.1), Minor (v1.2.0), Major (v2.0.0), Custom
- AC-004-3: Given user selects option, Then release started with selected version
- AC-004-4: Given track prompt, Then user can specify track (stable, beta, patch, etc.)

**Testing:** TDD

---

### REQ-005: Release Add

**Description:** Users can add issues to a release by setting the Release field.

**Rationale:** Associates issues with releases for tracking.

**Priority:** High

**Source:** Proposal - Workflow Changes section

**Acceptance Criteria:**
- AC-005-1: Given active release, When `gh pmu release add 123` runs, Then Release field set on issue #123
- AC-005-2: Given multiple active releases, When no `--version` specified, Then prompt for selection
- AC-005-3: Given `--version` flag, Then issue added to specified release
- AC-005-4: Given Release field set, Then value matches `{prefix}{version}` format

**Testing:** TDD

---

### REQ-006: Release Remove

**Description:** Users can remove issues from a release by clearing the Release field.

**Rationale:** Allows reassignment or removal from release.

**Priority:** High

**Source:** Proposal - Workflow Changes section

**Acceptance Criteria:**
- AC-006-1: Given issue with Release field set, When `gh pmu release remove 123` runs, Then Release field cleared
- AC-006-2: Given issue not in any release, When remove runs, Then no error (idempotent)

**Testing:** TDD

---

### REQ-007: Release Current

**Description:** Users can view all active releases.

**Rationale:** Visibility into concurrent release state.

**Priority:** High

**Source:** Proposal - Workflow Changes section

**Acceptance Criteria:**
- AC-007-1: Given multiple active releases, When `gh pmu release current` runs, Then all listed with version, track, start date, issue count
- AC-007-2: Given no active releases, Then "No active releases" displayed
- AC-007-3: Given `--track` filter, Then only releases of that track shown

**Testing:** TDD

---

### REQ-008: Release Close

**Description:** Users can close a release, generating artifacts and optionally creating a git tag.

**Rationale:** Completes the release lifecycle.

**Priority:** High

**Source:** Proposal - Workflow Changes section

**Acceptance Criteria:**
- AC-008-1: Given active release, When `gh pmu release close` runs, Then issues with Release field queried
- AC-008-2: Given issues not in Done status, Then warning displayed with option to continue
- AC-008-3: Given `--tag` flag, Then git tag created matching Release field value
- AC-008-4: Given close completes, Then artifacts generated in `Releases/{prefix}{version}/`
- AC-008-5: Given close completes, Then tracker issue closed
- AC-008-6: Given close completes, Then release removed from `release.active[]`

**Testing:** TDD

---

### REQ-009: Release List

**Description:** Users can list all releases (active and historical).

**Rationale:** View release history.

**Priority:** Medium

**Source:** Proposal - Workflow Changes section

**Acceptance Criteria:**
- AC-009-1: Given closed releases exist, When `gh pmu release list` runs, Then all releases shown (from tracker issues)
- AC-009-2: Given `--track` filter, Then only releases of that track shown
- AC-009-3: Given results, Then sorted by version descending

**Testing:** TDD

---

### REQ-010: Patch Command Alias

**Description:** `gh pmu patch` is an alias for `gh pmu release --track patch` with constraints.

**Rationale:** Simplified workflow for patch releases.

**Priority:** Medium

**Source:** Proposal - Patch Command section

**Acceptance Criteria:**
- AC-010-1: Given `gh pmu patch start --version 1.1.1`, Then equivalent to `release start --version 1.1.1 --track patch`
- AC-010-2: Given patch version not a PATCH increment, Then error (e.g., 1.2.0 not allowed)
- AC-010-3: Given issue with `breaking-change` label, When adding to patch, Then error
- AC-010-4: Given issue without bug/fix/hotfix label, When adding to patch, Then warning

**Testing:** TDD

---

### REQ-011: Track Prefix Configuration

**Description:** Release tracks use configurable prefixes in the Release field value.

**Rationale:** Enables filtering and visual distinction in project views.

**Priority:** High

**Source:** Proposal - Release Track Naming Convention section

**Acceptance Criteria:**
- AC-011-1: Given track configuration in `.gh-pmu.yml`, When release started, Then prefix applied to field value
- AC-011-2: Given stable track with prefix `v`, Then field value is `v1.2.0`
- AC-011-3: Given patch track with prefix `patch/`, Then field value is `patch/1.1.1`
- AC-011-4: Given track constraints configured, Then constraints enforced on release operations

**Testing:** TDD

---

### REQ-012: Init Sync Releases

**Description:** `gh pmu init` reconstructs `release.active[]` from open tracker issues.

**Rationale:** Enables recovery if `.gh-pmu.yml` is lost or out of sync.

**Priority:** Medium

**Source:** Proposal - Reconstruction via gh pmu init section

**Acceptance Criteria:**
- AC-012-1: Given open issues with `release` label, When `gh pmu init` runs, Then `release.active[]` populated
- AC-012-2: Given tracker issue title `Release: v1.2.0`, Then version and track parsed correctly
- AC-012-3: Given tracker issue title `Release: patch/1.1.1`, Then version and track parsed correctly

**Testing:** TDD

---

### REQ-013: Artifact Generation

**Description:** On release close, artifacts are generated in the appropriate directory.

**Rationale:** Automated documentation for releases.

**Priority:** Medium

**Source:** Proposal - Artifact directory structure section

**Acceptance Criteria:**
- AC-013-1: Given `release close`, Then `release-notes.md` generated from commit history
- AC-013-2: Given `release close`, Then `changelog.md` generated
- AC-013-3: Given release `v1.2.0`, Then artifacts in `Releases/v1.2.0/`
- AC-013-4: Given release `patch/1.1.1`, Then artifacts in `Releases/patch/1.1.1/`
- AC-013-5: Given tracker issue, Then artifact checkboxes checked

**Testing:** TDD

---

### REQ-014: Git Tag Format

**Description:** Git tags match the Release field value exactly.

**Rationale:** Consistency between project tracking and version control.

**Priority:** Medium

**Source:** Proposal - Git tag format section

**Acceptance Criteria:**
- AC-014-1: Given release `v1.2.0` with `--tag`, Then git tag `v1.2.0` created
- AC-014-2: Given release `patch/1.1.1` with `--tag`, Then git tag `patch/1.1.1` created
- AC-014-3: Given release `beta/2.0.0` with `--tag`, Then git tag `beta/2.0.0` created

**Testing:** TDD

---

## Non-Functional Requirements

### Usability

| ID | Requirement | Target |
|----|-------------|--------|
| NFR-USE-001 | Interactive prompts | Clear, minimal input required |
| NFR-USE-002 | Error messages | Actionable with suggested fixes |
| NFR-USE-003 | Help text | All commands have `--help` |

### Reliability

| ID | Requirement | Target |
|----|-------------|--------|
| NFR-REL-001 | Idempotent operations | Re-running doesn't cause errors |
| NFR-REL-002 | Config recovery | `init` can reconstruct from tracker issues |
| NFR-REL-003 | Graceful degradation | Works if optional features missing |

### Compatibility

| ID | Requirement | Target |
|----|-------------|--------|
| NFR-COM-001 | GitHub Projects | v2 (new projects) |
| NFR-COM-002 | gh CLI | Latest stable version |
| NFR-COM-003 | Multi-repo projects | Single primary repo supported |

---

## Constraints

| ID | Constraint | Impact |
|----|------------|--------|
| CON-001 | Text field only (not single-select) | Avoids API limitations |
| CON-002 | Tracker issues required | Cannot use field-only mode |
| CON-003 | Primary repo for labels | Multi-repo limited |

---

## Risks

| ID | Risk | Probability | Impact | Mitigation |
|----|------|-------------|--------|------------|
| RISK-001 | GitHub API rate limits | Low | Medium | Cache responses, batch operations |
| RISK-002 | Text field typos | Medium | Low | Commands manage field values, not users |
| RISK-003 | Config file conflicts | Low | Medium | Clear managed vs user-edited sections |

---

## Testing Approach

- **TDD:** Required for all gh-pmu development
- **Integration Tests:** GitHub API interactions
- **E2E Tests:** Full workflow scenarios
- **Coverage Target:** 80%

---

## Implementation Phases

### Phase 1: Wire Up Existing Stubs
1. Connect RunE handlers to WithDeps functions
2. Verify tracker issue creation/management works
3. Verify artifact generation works

### Phase 2: Add Release Field Support
1. Add `release` field configuration to schema
2. Implement field value management in `release start`
3. Update `release add/remove` to set/clear field

### Phase 3: Documentation Updates (process-docs)
1. Update GitHub-Workflow.md with Release field configuration
2. Update Framework Applicability table
3. Update .gh-pmu.yml template
4. Add setup guide

---

## Framework Handoff

**Target Projects:**
- gh-pmu (Project 11) - Implementation
- process-docs (Project 7) - Documentation

**Selected Framework:** IDPF-Structured (for gh-pmu implementation)

**Rationale:** Requirements are well-defined from proposal, fixed scope.

**Starting Point:** REQ-001 (Release Field Creation)

---

## Approval

| Role | Name | Date | Approved |
|------|------|------|----------|
| Stakeholder | | | [ ] |

---

**PRD Status: DRAFT - Awaiting approval**

---

*Generated from Proposal/Release-Field-Support.md using IDPF-PRD Framework*
