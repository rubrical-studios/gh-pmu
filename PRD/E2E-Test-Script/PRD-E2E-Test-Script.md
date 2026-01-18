# PRD: E2E Test Script

**Status:** Draft
**Created:** 2026-01-17
**Source Proposal:** Proposal/PROPOSAL-E2E-Test-Script.md

---

## Overview

This PRD defines the implementation of a Go-based end-to-end (E2E) test suite that validates complete workflows against a dedicated test project. The tests target commands that are difficult to unit test due to external process calls, complex state management, and API interactions.

**Key Benefits:**
- Coverage contribution via `go test -tags=e2e -cover`
- Windows, macOS, Linux support (native Go cross-platform)
- Catches API bugs (would have caught issue #551 batch mutation bug)
- Validates untested workflow commands (microsprint, branch, board, filter)
- Pre-merge gate integration with release commands

**Test Infrastructure:**
- Test Project: #41 (IDPF-gh-pmu-testing) - Private
- Test Repo: `rubrical-studios/gh-pmu-e2e-test` - Private

---

## Epics

### Epic 1: E2E Test Infrastructure
Foundation for running E2E tests including directory structure, test helpers, config generation, and cleanup utilities.

**Stories:** 1.1, 1.2, 1.3, 1.4

### Epic 2: Command Test Coverage
E2E tests for untested commands: microsprint, branch, board, filter.

**Stories:** 2.1, 2.2, 2.3, 2.4

### Epic 3: Workflow Tests
Multi-command integration tests covering complete user journeys.

**Stories:** 3.1, 3.2, 3.3

### Epic 4: Release Command Extensions
Integration of E2E tests with release workflow commands via extension points.

**Stories:** 4.1, 4.2, 4.3

---

## User Stories

### Story 1.1: E2E Test Directory and Build Infrastructure

**As a** developer
**I want** an E2E test directory with TestMain that builds a local binary
**So that** tests run against the current code rather than an installed extension

**Acceptance Criteria:**
- [ ] `test/e2e/` directory structure created
- [ ] `e2e_test.go` with `//go:build e2e` tag exists
- [ ] `TestMain` builds binary with coverage instrumentation
- [ ] Binary path handles Windows (.exe) and Unix platforms
- [ ] Build runs from project root via relative path

**Priority:** P0 - Must Have

**Source:** Proposal Section "Directory Structure", "Binary Build (TestMain)"

---

### Story 1.2: Test Configuration Generation

**As a** developer
**I want** tests to generate `.gh-pmu.yml` in a temp directory
**So that** tests are isolated and don't affect real project configuration

**Acceptance Criteria:**
- [ ] `setupTestConfig()` creates temp directory via `t.TempDir()`
- [ ] Config file written with test project #41 settings
- [ ] Config includes Status, Priority field mappings
- [ ] Tests use temp directory as working directory for command execution

**Priority:** P0 - Must Have

**Source:** Proposal Section "Config Generation"

---

### Story 1.3: Test Helper Functions

**As a** developer
**I want** reusable test helper functions
**So that** tests are consistent and maintainable

**Acceptance Criteria:**
- [ ] `runPMU()` executes local binary with args and returns stdout
- [ ] `assertContains()` validates expected content in output
- [ ] `assertNotContains()` validates absence of content in output
- [ ] `createTestIssue()` creates issues with `[E2E]` prefix
- [ ] `extractIssueNumber()` parses issue number from command output
- [ ] Helper functions log stderr on command failure

**Priority:** P0 - Must Have

**Source:** Proposal Sections "Command Execution", "Output Validation"

---

### Story 1.4: Test Cleanup Utilities

**As a** developer
**I want** a cleanup mechanism for test resources
**So that** test issues don't accumulate in the test project

**Acceptance Criteria:**
- [ ] Cleanup runs only when `E2E_CLEANUP=true` environment variable set
- [ ] Cleanup finds all issues with `[E2E]` prefix in test project
- [ ] Cleanup closes and deletes found issues
- [ ] Cleanup logs each issue removed with number and title

**Priority:** P0 - Must Have

**Source:** Proposal Section "Cleanup"

---

### Story 2.1: Microsprint Lifecycle E2E Test

**As a** developer
**I want** E2E tests for microsprint commands
**So that** microsprint start/add/current/list/close workflows are validated

**Acceptance Criteria:**
- [ ] Test starts a new microsprint with `--name` flag
- [ ] Test verifies `microsprint current` shows the sprint
- [ ] Test adds an issue to the microsprint
- [ ] Test verifies `microsprint list` shows the issue
- [ ] Test closes microsprint with `--skip-retro` flag
- [ ] Test issues cleaned up after test

**Priority:** P0 - Must Have

**Source:** Proposal Section "Category 1: Microsprint Workflows"

---

### Story 2.2: Branch Lifecycle E2E Test

**As a** developer
**I want** E2E tests for branch commands
**So that** branch start/current/list/close workflows are validated

**Acceptance Criteria:**
- [ ] Test starts a new branch with unique timestamped name
- [ ] Test verifies `branch current` shows the branch
- [ ] Test adds an issue to the branch via `move --branch current`
- [ ] Test verifies `branch list` shows the branch
- [ ] Test closes the branch
- [ ] Branch name uses `release/e2e-test-{timestamp}` format

**Priority:** P0 - Must Have

**Source:** Proposal Section "Category 2: Branch Workflows"

---

### Story 2.3: Board Rendering E2E Test

**As a** developer
**I want** E2E tests for board command
**So that** board visualization is validated across statuses

**Acceptance Criteria:**
- [ ] Test creates issues in different statuses (Backlog, In Progress, Done)
- [ ] Test runs `board` command
- [ ] Test verifies all status column headers appear in output
- [ ] Test verifies issues appear under correct columns
- [ ] Validates key content presence, not exact formatting

**Priority:** P0 - Must Have

**Source:** Proposal Section "Category 3: Board Rendering"

---

### Story 2.4: Filter Command E2E Test

**As a** developer
**I want** E2E tests for filter command
**So that** filter functionality is validated

**Acceptance Criteria:**
- [ ] Test creates issues with different field values
- [ ] Test runs filter command with field-based criteria
- [ ] Test verifies filtered results match expected issues
- [ ] Test verifies excluded issues are not in output

**Priority:** P0 - Must Have

**Source:** Proposal Section "In Scope" - Coverage of filter command

---

### Story 3.1: Create-to-Close Workflow E2E Test

**As a** developer
**I want** E2E tests for complete issue lifecycle
**So that** the full create-to-done workflow is validated

**Acceptance Criteria:**
- [ ] Test creates new issue with `--title`, `--status`, `--priority`
- [ ] Test moves issue through workflow: backlog -> in_progress -> in_review -> done
- [ ] Test verifies final state shows "Done" status
- [ ] Test uses sequential moves (not parallel)

**Priority:** P0 - Must Have

**Source:** Proposal Section "Category 4: Multi-Command Workflows" - TestCreateToCloseWorkflow

---

### Story 3.2: Sub-Issue Workflow E2E Test

**As a** developer
**I want** E2E tests for sub-issue operations
**So that** sub create/list/remove workflows are validated

**Acceptance Criteria:**
- [ ] Test creates parent issue
- [ ] Test creates sub-issue via `sub create --parent`
- [ ] Test verifies `sub list` shows the sub-issue
- [ ] Test removes sub-issue via `sub remove`
- [ ] Test verifies removal succeeded (no error in output)

**Priority:** P0 - Must Have

**Source:** Proposal Section "Category 4: Multi-Command Workflows" - TestSubIssueWorkflow

---

### Story 3.3: Multi-Issue Move E2E Test

**As a** developer
**I want** E2E tests for batch issue moves
**So that** multi-issue operations are validated

**Acceptance Criteria:**
- [ ] Test creates multiple issues
- [ ] Test moves multiple issues in single command
- [ ] Test verifies all issues have updated status
- [ ] Validates batch operations work correctly

**Priority:** P0 - Must Have

**Source:** Proposal Section "In Scope" - Coverage validation of untested commands

---

### Story 4.1: E2E Impact Analysis Script

**As a** release manager
**I want** automated analysis of E2E test impact from commits
**So that** I know which tests need attention before release

**Acceptance Criteria:**
- [ ] `analyze-e2e-impact.js` script created in `.claude/scripts/e2e/`
- [ ] Script maps command directories to E2E test files
- [ ] Script outputs JSON with `impactedTests` array
- [ ] Script identifies `newCommandsWithoutTests` for warning
- [ ] Script provides `recommendation` for test review

**Priority:** P0 - Must Have

**Source:** Proposal Section "Script: analyze-e2e-impact.js"

---

### Story 4.2: E2E Gate Runner Script

**As a** release manager
**I want** a script that runs E2E tests and reports results
**So that** E2E tests can gate releases

**Acceptance Criteria:**
- [ ] `run-e2e-gate.js` script created in `.claude/scripts/e2e/`
- [ ] Script runs E2E tests via `go test -tags=e2e`
- [ ] Script outputs JSON with `success`, `testsRun`, `testsPassed`, `duration`
- [ ] Script exits with code 1 on test failure
- [ ] Script sets `GOCOVERDIR` for coverage collection

**Priority:** P0 - Must Have

**Source:** Proposal Section "Script: run-e2e-gate.js"

---

### Story 4.3: Release Command E2E Extensions

**As a** release manager
**I want** E2E tests integrated into release commands
**So that** releases are gated by E2E test success

**Acceptance Criteria:**
- [ ] `/prepare-release` `post-analysis` runs E2E impact analysis
- [ ] `/prepare-release` `post-validation` runs E2E gate
- [ ] `/prepare-beta` `post-analysis` runs E2E impact analysis
- [ ] `/prepare-beta` `post-validation` runs E2E gate
- [ ] `/merge-branch` `gates` includes E2E test check
- [ ] `--skip-e2e` flag bypasses E2E gate when needed

**Priority:** P0 - Must Have

**Source:** Proposal Section "Release Command Extensions"

---

## Technical Notes

> Implementation hints, not requirements. Do not create issues from this section.

### Test Execution
- Sequential execution only (no `t.Parallel()`) to avoid shared resource conflicts
- Binary built once in `TestMain`, reused across all tests
- Temp directories for config isolation per test

### Project Field Verification
- Tests verify required fields exist (Status, Priority, Release, Microsprint)
- Fields must be created via `gh pmu init` on test project beforehand
- Tests fail fast if fields missing

### Coverage Integration
- Binary built with `-cover` flag for instrumentation
- Coverage data written to `GOCOVERDIR` directory
- Merge with unit test coverage via `go tool covdata textfmt`

### E2E Coverage Map (for impact analysis)
```
cmd/microsprint → test/e2e/microsprint_test.go
cmd/branch     → test/e2e/branch_test.go
cmd/board      → test/e2e/board_test.go
cmd/filter     → test/e2e/filter_test.go
cmd/move       → test/e2e/workflow_test.go
cmd/sub        → test/e2e/workflow_test.go
cmd/create     → test/e2e/workflow_test.go
```

---

## Out of Scope

| Item | Reason |
|------|--------|
| CI/CD integration | Future phase |
| Interactive prompt testing | Use `--skip-retro` and similar flags |
| `--config` flag | Temp directory approach sufficient |
| Testing `gh pmu init` | Write config directly in tests |

---

## Dependencies

| Dependency | Type | Notes |
|------------|------|-------|
| Issue #551 | Reference | Bug that E2E would have caught |
| Issue #555 | Tracking | This proposal's tracking issue |
| Test Project #41 | Infrastructure | Must exist with fields configured |
| Test Repo | Infrastructure | `rubrical-studios/gh-pmu-e2e-test` |
| Go 1.22+ | Technology | Required for build and coverage |

---

## Open Questions

None - proposal is comprehensive and ready for implementation.

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Config discovery finds wrong file | Tests fail or affect real project | Use temp directory with `cmd.Dir` |
| Rate limiting | Tests slow or fail | Batch operations, add delays |
| Test pollution | Stale test data accumulates | `[E2E]` prefix, manual cleanup |
| Flaky network | Intermittent failures | Retry logic, longer timeouts |
| E2E gate slows releases | Longer release cycle | `--skip-e2e` flag for emergencies |

---

*Generated by create-prd skill*
*Ready for Create-Backlog*
