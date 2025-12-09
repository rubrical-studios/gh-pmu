# Integration Testing Backlog

**Revision:** 1
**Last Updated:** 2025-12-04
**Source:** PROPOSAL-Automated-Testing.md v1.3
**Goal:** Achieve 80% test coverage through integration tests for `run*` functions and UAT scenarios

---

## Current State

| Package | Coverage | Target | Gap |
|---------|----------|--------|-----|
| `internal/api` | 96.6% | 80% | **Exceeded** |
| `internal/config` | 97.0% | 80% | **Exceeded** |
| `internal/ui` | 96.9% | 80% | **Exceeded** |
| `cmd` | 51.2% | 80% | **28.8%** |
| **Total** | **63.6%** | **80%** | **16.4%** |

---

## Definition of Done (Integration Tests)

All integration test stories must meet these criteria:
- [ ] Tests use `//go:build integration` tag
- [ ] Tests run against test fixtures (not production)
- [ ] All created resources cleaned up after test
- [ ] Tests can run independently (no ordering dependencies)
- [ ] Coverage for target functions reaches 80%+
- [ ] Tests pass in CI environment

---

## Phase 1: Foundation

**Phase Goal:** Set up integration test infrastructure

### Story IT-1.1: Create Test Fixtures on GitHub

**As a** developer running integration tests
**I want** dedicated test resources on GitHub
**So that** tests run in isolation from production data

**Acceptance Criteria:**
- [ ] Create test repository `gh-pmu-test` under `rubrical-studios`
- [ ] Create test project with Status, Priority, Sprint, Estimate fields
- [ ] Create seed issues with known states for read-only tests
- [ ] Document test fixture setup in TESTING.md

**Story Points:** 3
**Priority:** High
**Status:** Backlog

---

### Story IT-1.2: Implement internal/testutil Package

**As a** developer writing integration tests
**I want** shared test utilities
**So that** I can write tests without boilerplate

**Acceptance Criteria:**
- [ ] `SetupTestClient(t)` - creates authenticated API client
- [ ] `RequireTestEnv(t)` - skips if env vars not set
- [ ] `CreateTestIssue(t, title)` - creates issue and returns cleanup func
- [ ] `DeleteTestIssue(t, num)` - removes test issue
- [ ] `GetProjectItem(t, issueNum)` - fetches item with fields
- [ ] `RunCommand(t, args...)` - executes gh pmu command
- [ ] `ExtractIssueNumber(t, output)` - parses issue # from output

**Story Points:** 8
**Priority:** High
**Status:** Backlog

---

### Story IT-1.3: Create GitHub Actions Workflow

**As a** developer running integration tests
**I want** a CI workflow for integration tests
**So that** I can run them on demand

**Acceptance Criteria:**
- [ ] `.github/workflows/integration-tests.yml` created
- [ ] `workflow_dispatch` trigger with test_type input
- [ ] Environment secrets configured (TEST_GH_TOKEN)
- [ ] Environment variables configured (TEST_PROJECT_*, TEST_REPO_*)
- [ ] Cleanup step runs on success or failure

**Story Points:** 3
**Priority:** High
**Status:** Backlog

---

### Story IT-1.4: Document Test Environment Setup

**As a** new contributor
**I want** clear documentation for running integration tests locally
**So that** I can contribute to test coverage

**Acceptance Criteria:**
- [ ] TESTING.md created with setup instructions
- [ ] Environment variable documentation
- [ ] How to run tests locally
- [ ] How to run tests in CI
- [ ] Troubleshooting common issues

**Story Points:** 2
**Priority:** Medium
**Status:** Backlog

---

## Phase 2: Command Integration Tests (0% Coverage Functions)

**Phase Goal:** Test all `run*` functions that are currently at 0% coverage

### Story IT-2.1: runList Integration Tests

**As a** developer verifying list command
**I want** integration tests for `runList`
**So that** I can ensure listing works with real API

**Target Function:** `cmd/list.go:46` - `runList` (0% → 80%+)

**Acceptance Criteria:**
- [ ] Test listing all items from project
- [ ] Test filtering by --status flag
- [ ] Test filtering by --priority flag
- [ ] Test --json output format
- [ ] Test empty result handling
- [ ] Test with --has-sub-issues filter

**Story Points:** 5
**Priority:** High
**Status:** Backlog

---

### Story IT-2.2: runView Integration Tests

**As a** developer verifying view command
**I want** integration tests for `runView`
**So that** I can ensure viewing works with real API

**Target Function:** `cmd/view.go:42` - `runView` (0% → 80%+)

**Acceptance Criteria:**
- [ ] Test viewing issue by number
- [ ] Test viewing issue by URL
- [ ] Test viewing issue with sub-issues
- [ ] Test viewing sub-issue (shows parent)
- [ ] Test --json output format
- [ ] Test issue not found error

**Story Points:** 5
**Priority:** High
**Status:** Backlog

---

### Story IT-2.3: runCreate Integration Tests

**As a** developer verifying create command
**I want** integration tests for `runCreate`
**So that** I can ensure creation works with real API

**Target Function:** `cmd/create.go:48` - `runCreate` (44.9% → 80%+)

**Acceptance Criteria:**
- [ ] Test creating issue with --title and --body
- [ ] Test setting --status and --priority
- [ ] Test applying --label flags
- [ ] Test applying defaults from config
- [ ] Test merging CLI labels with config defaults
- [ ] Test field value aliases

**Story Points:** 5
**Priority:** High
**Status:** Backlog

---

### Story IT-2.4: runMove Integration Tests

**As a** developer verifying move command
**I want** integration tests for `runMove`
**So that** I can ensure field updates work with real API

**Target Function:** `cmd/move.go:76` - `runMove` (16.7% → 80%+)

**Acceptance Criteria:**
- [ ] Test changing --status
- [ ] Test changing --priority
- [ ] Test changing multiple fields
- [ ] Test --recursive flag with sub-issues
- [ ] Test field value aliases
- [ ] Test issue not in project error
- [ ] Test pagination: move issue in project with >100 items (ref: #90)

**Story Points:** 5
**Priority:** High
**Status:** Backlog

**Notes:** Pagination test requires a project with >100 items to verify fix from #90.

---

### Story IT-2.5: runIntake Integration Tests

**As a** developer verifying intake command
**I want** integration tests for `runIntake`
**So that** I can ensure intake works with real API

**Target Function:** `cmd/intake.go:55` - `runIntake` (0% → 80%+)

**Acceptance Criteria:**
- [ ] Test finding untracked issues
- [ ] Test --dry-run shows what would be added
- [ ] Test --apply adds issues to project
- [ ] Test --json output format
- [ ] Test no untracked issues case

**Story Points:** 5
**Priority:** Medium
**Status:** Backlog

---

### Story IT-2.6: runTriage Integration Tests

**As a** developer verifying triage command
**I want** integration tests for `runTriage`
**So that** I can ensure triage works with real API

**Target Function:** `cmd/triage.go:68` - `runTriage` (0% → 80%+)

**Acceptance Criteria:**
- [ ] Test running named triage config
- [ ] Test --dry-run shows what would change
- [ ] Test applying labels
- [ ] Test applying status/priority changes
- [ ] Test --list shows available configs
- [ ] Test config not found error

**Story Points:** 5
**Priority:** Medium
**Status:** Backlog

---

### Story IT-2.7: runSplit Integration Tests

**As a** developer verifying split command
**I want** integration tests for `runSplit`
**So that** I can ensure split works with real API

**Target Function:** `cmd/split.go:61` - `runSplit` (0% → 80%+)

**Acceptance Criteria:**
- [ ] Test splitting from issue body checklist
- [ ] Test splitting from arguments
- [ ] Test --dry-run shows what would be created
- [ ] Test sub-issues linked to parent
- [ ] Test --json output format
- [ ] Test issue with no checklist

**Story Points:** 5
**Priority:** Medium
**Status:** Backlog

---

### Story IT-2.8: runSubAdd Integration Tests

**As a** developer verifying sub add command
**I want** integration tests for `runSubAdd`
**So that** I can ensure linking works with real API

**Target Function:** `cmd/sub.go:54` - `runSubAdd` (0% → 80%+)

**Acceptance Criteria:**
- [ ] Test linking existing issues
- [ ] Test already linked error
- [ ] Test parent not found error
- [ ] Test child not found error

**Story Points:** 3
**Priority:** Medium
**Status:** Backlog

---

### Story IT-2.9: runSubCreate Integration Tests

**As a** developer verifying sub create command
**I want** integration tests for `runSubCreate`
**So that** I can ensure sub-issue creation works with real API

**Target Function:** `cmd/sub.go:202` - `runSubCreate` (0% → 80%+)

**Acceptance Criteria:**
- [ ] Test creating sub-issue with --title
- [ ] Test --inherit-labels flag
- [ ] Test --inherit-assignees flag
- [ ] Test sub-issue linked to parent
- [ ] Test parent not found error

**Story Points:** 5
**Priority:** Medium
**Status:** Backlog

---

### Story IT-2.10: runSubList Integration Tests

**As a** developer verifying sub list command
**I want** integration tests for `runSubList`
**So that** I can ensure listing sub-issues works with real API

**Target Function:** `cmd/sub.go:337` - `runSubList` (0% → 80%+)

**Acceptance Criteria:**
- [ ] Test listing sub-issues
- [ ] Test showing completion count
- [ ] Test --json output format
- [ ] Test no sub-issues case
- [ ] Test parent not found error

**Story Points:** 3
**Priority:** Medium
**Status:** Backlog

---

### Story IT-2.11: runSubRemove Integration Tests

**As a** developer verifying sub remove command
**I want** integration tests for `runSubRemove`
**So that** I can ensure unlinking works with real API

**Target Function:** `cmd/sub.go:521` - `runSubRemove` (0% → 80%+)

**Acceptance Criteria:**
- [ ] Test unlinking sub-issue
- [ ] Test not linked error
- [ ] Test parent not found error
- [ ] Test child not found error

**Story Points:** 3
**Priority:** Medium
**Status:** Backlog

---

### Story IT-2.12: runInit Integration Tests

**As a** developer verifying init command
**I want** integration tests for `runInit`
**So that** I can ensure initialization works with real API

**Target Function:** `cmd/init.go:36` - `runInit` (0% → 80%+)

**Acceptance Criteria:**
- [ ] Test non-interactive init with flags
- [ ] Test project validation
- [ ] Test field metadata fetching
- [ ] Test config file creation
- [ ] Test existing config handling

**Story Points:** 5
**Priority:** Low
**Status:** Backlog

**Notes:** Requires I/O mocking for interactive prompts

---

## Phase 3: Unit Test Gap Filling

**Phase Goal:** Increase coverage on functions between 0-80%

### Story IT-3.1: filterByHasSubIssues Unit Tests

**Target Function:** `cmd/list.go:126` - `filterByHasSubIssues` (0% → 80%+)

**Acceptance Criteria:**
- [ ] Test filtering items with sub-issues
- [ ] Test filtering items without sub-issues
- [ ] Test empty list

**Story Points:** 2
**Priority:** Medium
**Status:** Backlog

---

### Story IT-3.2: newSubAddCommand Flag Tests

**Target Function:** `cmd/sub.go:32` - `newSubAddCommand` (66.7% → 80%+)

**Acceptance Criteria:**
- [ ] Test all flag combinations
- [ ] Test flag validation

**Story Points:** 1
**Priority:** Low
**Status:** Backlog

---

### Story IT-3.3: newSubRemoveCommand Flag Tests

**Target Function:** `cmd/sub.go:499` - `newSubRemoveCommand` (66.7% → 80%+)

**Acceptance Criteria:**
- [ ] Test all flag combinations
- [ ] Test flag validation

**Story Points:** 1
**Priority:** Low
**Status:** Backlog

---

### Story IT-3.4: writeConfig Error Path Tests

**Target Function:** `cmd/init.go:419` - `writeConfig` (75% → 80%+)

**Acceptance Criteria:**
- [ ] Test file write error handling
- [ ] Test directory creation error

**Story Points:** 1
**Priority:** Low
**Status:** Backlog

---

### Story IT-3.5: AddLabelToIssue Error Tests

**Target Function:** `internal/api/mutations.go:382` - `AddLabelToIssue` (66.7% → 80%+)

**Acceptance Criteria:**
- [ ] Test label not found error
- [ ] Test API error handling

**Story Points:** 2
**Priority:** Low
**Status:** Backlog

---

## Phase 4: User Acceptance Tests (UAT)

**Phase Goal:** Validate end-to-end workflows from user perspective

### Story IT-4.1: UAT Epic 1 - Core Unification

**As a** product owner
**I want** UAT tests for Epic 1 acceptance criteria
**So that** I can verify the core functionality works end-to-end

**Acceptance Criteria:**
- [ ] UAT-1.1: Initialize new project (`gh pmu init`)
- [ ] UAT-1.2: List and filter issues (`gh pmu list`)
- [ ] UAT-1.3: Create issue with fields (`gh pmu create`)
- [ ] UAT-1.4: Move issue through workflow (`gh pmu move`)
- [ ] UAT-1.5: Intake untracked issues (`gh pmu intake`)
- [ ] UAT-1.6: Triage with rules (`gh pmu triage`)
- [ ] UAT-1.7: Manage sub-issues (`gh pmu sub *`)
- [ ] UAT-1.8: Split issue into tasks (`gh pmu split`)

**Story Points:** 8
**Priority:** Medium
**Status:** Backlog

---

### Story IT-4.2: UAT Epic 3 - Enhanced Integration

**As a** product owner
**I want** UAT tests for Epic 3 acceptance criteria
**So that** I can verify enhanced integration features

**Acceptance Criteria:**
- [ ] UAT-3.1: Cross-repo sub-issues
- [ ] UAT-3.2: Progress tracking (sub-issue completion %)
- [ ] UAT-3.3: Recursive operations on issue trees

**Story Points:** 5
**Priority:** Low
**Status:** Backlog

---

## Summary

| Phase | Stories | Story Points | Priority |
|-------|---------|--------------|----------|
| Phase 1: Foundation | 4 | 16 | High |
| Phase 2: Command Integration | 12 | 54 | High/Medium |
| Phase 3: Unit Test Gaps | 5 | 7 | Low |
| Phase 4: UAT | 2 | 13 | Medium/Low |
| **Total** | **23** | **90** | - |

### Expected Coverage After Completion

| Package | Current | After Phase 2 | After All Phases |
|---------|---------|---------------|------------------|
| `cmd` | 51.2% | ~80% | ~85% |
| **Total** | **63.6%** | **~80%** | **~85%** |

---

## Running the Tests

### Locally
```bash
# Set environment variables
export TEST_PROJECT_OWNER="rubrical-studios"
export TEST_PROJECT_NUMBER="99"
export TEST_REPO_OWNER="rubrical-studios"
export TEST_REPO_NAME="gh-pmu-test"

# Run all integration tests
go test -v -tags=integration ./...

# Run specific test
go test -v -tags=integration ./cmd/... -run "TestRunCreate_Integration"

# Run UAT tests
go test -v -tags=uat ./test/uat/...
```

### Via GitHub Actions
```bash
# Run all tests
gh workflow run integration-tests.yml -f test_type=all

# Run only integration tests
gh workflow run integration-tests.yml -f test_type=integration

# Run only UAT tests
gh workflow run integration-tests.yml -f test_type=uat
```
