# Proposal: Automated Non-Destructive Integration Tests & UAT

**Version:** 1.3
**Date:** 2025-12-04
**Author:** PRD-Analyst, API-Integration-Specialist
**Status:** Active

---

## Executive Summary

### Current State (v0.2.10)

The gh-pmu CLI extension has achieved **63.6% total test coverage** through comprehensive unit testing efforts. Current coverage by package:

| Package | Coverage | Status |
|---------|----------|--------|
| `internal/api` | 96.6% | ✅ Excellent |
| `internal/config` | 97.0% | ✅ Excellent |
| `internal/ui` | 96.9% | ✅ Excellent |
| `cmd` | 51.2% | 🔄 Needs integration tests |
| **Total** | **63.6%** | 🎯 Target: 80% |

### Remaining Gap Analysis

The existing test suite has strong unit test coverage but lacks:

1. **Integration tests** for `run*` functions that make GitHub API calls (12 functions at 0%)
2. **End-to-end workflow tests** covering complete user scenarios
3. **UAT scenarios** validating acceptance criteria from a user perspective
4. **Non-destructive test patterns** that can run safely against real GitHub projects

### Proposed Solution

Implement integration and UAT testing using **non-destructive patterns** that:
- Test against dedicated test fixtures (projects, repos, issues)
- Use read-only operations where possible
- Clean up any created resources after tests
- Support both CI/CD automation and local development

### Key Benefits

| Benefit | Impact |
|---------|--------|
| Reach 80% coverage target | Close the 16.4% gap via integration tests |
| Increased confidence in releases | Catch API integration bugs before users |
| Faster development cycles | Automated regression testing |
| Living documentation | Tests serve as usage examples |
| Safer refactoring | Full coverage enables bold improvements |

---

## Scope

### In Scope

- Integration tests for all GraphQL API operations
- End-to-end tests for all CLI commands
- UAT scenarios covering Epics 1-3 acceptance criteria
- Test fixture management (setup/teardown)
- CI/CD pipeline integration
- Test documentation and runbooks

### Out of Scope

- Load/performance testing
- Security penetration testing
- Template ecosystem testing (Epic 4 - future)
- Cross-platform compatibility testing (Windows/macOS/Linux variations)

---

## Non-Destructive Testing Strategy

### Core Principles

1. **Isolation**: Tests use dedicated test projects/repos, never production data
2. **Idempotency**: Tests can run multiple times without side effects
3. **Cleanup**: Any created resources are removed after test completion
4. **Read-First**: Prefer read operations; only write when necessary to validate
5. **Atomic Transactions**: Each test is independent and self-contained

### Test Environment Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    GitHub Test Organization                  │
│                    (gh-pmu-test-org)                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────────┐  ┌──────────────────┐                │
│  │  Test Repo 1     │  │  Test Repo 2     │                │
│  │  (primary)       │  │  (cross-repo)    │                │
│  │                  │  │                  │                │
│  │  - Seed issues   │  │  - Seed issues   │                │
│  │  - Labels        │  │  - Labels        │                │
│  └──────────────────┘  └──────────────────┘                │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Test Project (GitHub Projects v2)        │  │
│  │                                                        │  │
│  │  Fields: Status, Priority, Sprint, Estimate           │  │
│  │  Views: Kanban, Table                                 │  │
│  │  Items: Pre-seeded test issues                        │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Resource Management Strategy

| Resource Type | Strategy | Cleanup Method |
|--------------|----------|----------------|
| Test Project | Persistent fixture | Reset to known state |
| Test Issues | Created per-suite | Delete after suite |
| Project Items | Created per-test | Remove from project |
| Sub-Issues | Created per-test | Unlink and delete |
| Field Values | Modified per-test | Reset to defaults |

---

## Test Categories

### Category 1: API Integration Tests

**Purpose:** Verify GraphQL queries and mutations work correctly with real GitHub API

**Approach:**
- Use `//go:build integration` tag
- Run against test fixtures
- Validate response structure and data integrity

**Coverage Areas:**

| Module | Operations | Test Count (Est.) |
|--------|-----------|-------------------|
| Queries | GetProject, GetFields, GetIssue, GetItems, etc. | 15 |
| Mutations | CreateIssue, AddToProject, SetField, LinkSubIssue, etc. | 20 |
| Error Handling | Auth failures, Not found, Rate limits | 10 |
| Edge Cases | Empty results, Large payloads, Special characters | 8 |

**Example Test Pattern:**

```go
//go:build integration

func TestGetProjectFields_ReturnsAllFieldTypes(t *testing.T) {
    // Arrange: Use test project with known fields
    client := setupTestClient(t)
    projectID := os.Getenv("TEST_PROJECT_ID")

    // Act: Query fields
    fields, err := client.GetProjectFields(projectID)

    // Assert: Verify expected fields exist
    require.NoError(t, err)
    assert.Contains(t, fieldNames(fields), "Status")
    assert.Contains(t, fieldNames(fields), "Priority")

    // Cleanup: None needed (read-only)
}
```

### Category 2: Command Integration Tests

**Purpose:** Verify CLI commands work end-to-end with real API

**Approach:**
- Execute actual `gh pmu` commands via subprocess
- Capture stdout/stderr
- Validate output format and content
- Verify side effects via API queries

**Coverage Areas:**

| Command | Scenarios | Test Count (Est.) |
|---------|-----------|-------------------|
| `init` | New config, existing config, auto-detect | 4 |
| `list` | Filter by status, JSON output, empty results | 6 |
| `view` | Valid issue, invalid issue, with fields | 4 |
| `create` | With fields, minimal, validation errors | 5 |
| `move` | Change status, change priority, multiple fields | 5 |
| `intake` | Find untracked, add to project, dry-run | 4 |
| `triage` | Apply rules, skip processed, dry-run | 5 |
| `sub add` | Link existing, cross-repo, invalid parent | 4 |
| `sub create` | New sub-issue, with fields, inherit parent | 4 |
| `sub list` | With children, no children, recursive | 3 |
| `sub remove` | Unlink, delete, not found | 3 |
| `split` | From checklist, from args, validation | 4 |

**Example Test Pattern:**

```go
//go:build integration

func TestListCommand_FiltersByStatus(t *testing.T) {
    // Arrange: Ensure test issues exist with known statuses
    setupTestIssues(t, []TestIssue{
        {Title: "Test-Todo", Status: "Todo"},
        {Title: "Test-Done", Status: "Done"},
    })
    defer cleanupTestIssues(t)

    // Act: Run list command with filter
    output, err := runCommand("gh", "pmu", "list", "--status", "Todo")

    // Assert: Only Todo issues in output
    require.NoError(t, err)
    assert.Contains(t, output, "Test-Todo")
    assert.NotContains(t, output, "Test-Done")
}
```

### Category 3: User Acceptance Tests (UAT)

**Purpose:** Validate complete user workflows and acceptance criteria from PRD

**Approach:**
- Scenario-based tests using Given-When-Then structure
- Cover complete user journeys (multi-command workflows)
- Validate business value delivery
- Human-readable test names and documentation

**UAT Scenarios by Epic:**

#### Epic 1: Core Unification

| ID | Scenario | Acceptance Criteria |
|----|----------|---------------------|
| UAT-1.1 | Initialize new project | User can run `gh pmu init` and get working config |
| UAT-1.2 | List and filter issues | User can list issues filtered by status/priority |
| UAT-1.3 | Create issue with fields | User can create issue with status/priority pre-set |
| UAT-1.4 | Move issue through workflow | User can update issue status/priority via CLI |
| UAT-1.5 | Intake untracked issues | User can find and add issues not yet in project |
| UAT-1.6 | Triage with rules | User can apply configurable rules to categorize issues |
| UAT-1.7 | Manage sub-issues | User can create, link, list, and remove sub-issues |
| UAT-1.8 | Split issue into tasks | User can split issue into sub-issues from checklist |

#### Epic 2: Project Templates

| ID | Scenario | Acceptance Criteria |
|----|----------|---------------------|
| UAT-2.1 | Create from template | User can create project from YAML template |
| UAT-2.2 | Export project | User can export existing project to YAML |
| UAT-2.3 | Validate template | User can validate template before use |
| UAT-2.4 | List templates | User can discover available templates |

#### Epic 3: Enhanced Integration

| ID | Scenario | Acceptance Criteria |
|----|----------|---------------------|
| UAT-3.1 | Cross-repo sub-issues | User can link issues across repositories |
| UAT-3.2 | Progress tracking | User can see sub-issue completion percentage |
| UAT-3.3 | Recursive operations | User can bulk update issue hierarchies |

**Example UAT Test:**

```go
//go:build uat

func TestUAT_1_3_CreateIssueWithFields(t *testing.T) {
    /*
    Scenario: Create issue with project fields pre-populated

    Given I have initialized gh-pmu with a valid configuration
    And the project has Status and Priority fields
    When I run: gh pmu create --title "New Feature" --status "Todo" --priority "High"
    Then a new issue should be created
    And it should be added to the project
    And the Status field should be "Todo"
    And the Priority field should be "High"
    */

    // Given
    cfg := setupTestConfig(t)

    // When
    output, err := runCommand("gh", "pmu", "create",
        "--title", "UAT-Test-Issue-"+randomSuffix(),
        "--status", "Todo",
        "--priority", "High",
    )
    require.NoError(t, err)

    // Then
    issueNum := extractIssueNumber(output)
    defer deleteTestIssue(t, issueNum)

    issue := getIssueWithFields(t, issueNum)
    assert.Equal(t, "Todo", issue.Fields["Status"])
    assert.Equal(t, "High", issue.Fields["Priority"])
}
```

---

## Test Infrastructure

### Test Fixtures

#### Required GitHub Resources

1. **Test Organization:** `gh-pmu-test-org` (or user account)
2. **Test Repository 1:** `gh-pmu-test-repo` (primary)
3. **Test Repository 2:** `gh-pmu-test-repo-2` (cross-repo tests)
4. **Test Project:** GitHub Projects v2 with standard fields

#### Fixture Configuration

```yaml
# testdata/fixtures/test-project.yml
project:
  title: "gh-pmu Integration Test Project"
  owner: "gh-pmu-test-org"
  fields:
    - name: Status
      type: single_select
      options: [Todo, In Progress, In Review, Done]
    - name: Priority
      type: single_select
      options: [Low, Medium, High, Critical]
    - name: Sprint
      type: iteration
    - name: Estimate
      type: number
  views:
    - name: Kanban
      type: board
      group_by: Status
    - name: All Items
      type: table

seed_issues:
  - title: "[TEST] Seed Issue 1"
    status: Todo
    priority: Medium
  - title: "[TEST] Seed Issue 2"
    status: In Progress
    priority: High
  - title: "[TEST] Seed Issue 3 (Parent)"
    status: Todo
    priority: Low
    sub_issues:
      - title: "[TEST] Sub-Issue 3.1"
      - title: "[TEST] Sub-Issue 3.2"
```

### Test Utilities Package

Create `internal/testutil/` package with:

```go
// internal/testutil/testutil.go
package testutil

// Client setup
func SetupTestClient(t *testing.T) *api.Client
func GetTestProjectID() string
func GetTestRepoOwner() string
func GetTestRepoName() string

// Issue management
func CreateTestIssue(t *testing.T, title string, opts ...IssueOption) int
func DeleteTestIssue(t *testing.T, issueNum int)
func CleanupTestIssues(t *testing.T)

// Project item management
func AddIssueToProject(t *testing.T, issueNum int) string
func RemoveItemFromProject(t *testing.T, itemID string)
func ResetProjectItem(t *testing.T, itemID string)

// Assertions
func AssertIssueHasField(t *testing.T, issueNum int, field, value string)
func AssertIssueInProject(t *testing.T, issueNum int)

// Command execution
func RunCommand(t *testing.T, args ...string) (string, error)
func RunCommandWithConfig(t *testing.T, cfg *config.Config, args ...string) (string, error)
```

### CI/CD Integration

#### GitHub Actions Workflow

```yaml
# .github/workflows/integration-tests.yml
name: Integration Tests

on:
  workflow_dispatch:  # Manual trigger only
    inputs:
      test_type:
        description: 'Type of tests to run'
        required: true
        default: 'all'
        type: choice
        options:
          - all
          - integration
          - uat

jobs:
  integration:
    runs-on: ubuntu-latest
    environment: integration-tests

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Run Integration Tests
        if: ${{ inputs.test_type == 'all' || inputs.test_type == 'integration' }}
        env:
          GH_TOKEN: ${{ secrets.TEST_GH_TOKEN }}
          TEST_PROJECT_ID: ${{ vars.TEST_PROJECT_ID }}
          TEST_REPO_OWNER: ${{ vars.TEST_REPO_OWNER }}
          TEST_REPO_NAME: ${{ vars.TEST_REPO_NAME }}
        run: |
          go test -v -tags=integration ./...

      - name: Run UAT Tests
        if: ${{ inputs.test_type == 'all' || inputs.test_type == 'uat' }}
        env:
          GH_TOKEN: ${{ secrets.TEST_GH_TOKEN }}
          TEST_PROJECT_ID: ${{ vars.TEST_PROJECT_ID }}
          TEST_REPO_OWNER: ${{ vars.TEST_REPO_OWNER }}
          TEST_REPO_NAME: ${{ vars.TEST_REPO_NAME }}
        run: |
          go test -v -tags=uat ./...

      - name: Cleanup Test Resources
        if: always()
        env:
          GH_TOKEN: ${{ secrets.TEST_GH_TOKEN }}
        run: |
          go run ./cmd/testcleanup/main.go
```

**Running Integration Tests:**
```bash
# Via GitHub CLI
gh workflow run integration-tests.yml -f test_type=all

# Or from GitHub Actions UI: Actions → Integration Tests → Run workflow
```

### Required Secrets/Variables

| Name | Type | Purpose |
|------|------|---------|
| `TEST_GH_TOKEN` | Secret | GitHub PAT with repo, project, admin:org scopes |
| `TEST_PROJECT_ID` | Variable | Project node ID for test project |
| `TEST_REPO_OWNER` | Variable | Owner of test repositories |
| `TEST_REPO_NAME` | Variable | Primary test repository name |
| `TEST_REPO_NAME_2` | Variable | Secondary test repository (cross-repo) |

---

## Implementation Roadmap

### Phase 1: Foundation (Story Points: 13)

**Deliverables:**
- [ ] Create test organization/repos/project on GitHub
- [ ] Implement `internal/testutil/` package
- [ ] Create fixture seed scripts
- [ ] Set up CI/CD workflow skeleton
- [ ] Document test environment setup

**Dependencies:** None

### Phase 2: API Integration Tests (Story Points: 21)

**Deliverables:**
- [ ] Query tests (GetProject, GetFields, GetIssue, etc.)
- [ ] Mutation tests (CreateIssue, SetField, LinkSubIssue, etc.)
- [ ] Error handling tests
- [ ] Edge case coverage

**Dependencies:** Phase 1

### Phase 3: Command Integration Tests (Story Points: 21)

**Deliverables:**
- [ ] Init command tests
- [ ] List/View command tests
- [ ] Create/Move command tests
- [ ] Sub-issue command tests
- [ ] Intake/Triage command tests
- [ ] Split command tests

**Dependencies:** Phase 2

### Phase 4: UAT Implementation (Story Points: 13)

**Deliverables:**
- [ ] Epic 1 UAT scenarios (8 tests)
- [ ] Epic 2 UAT scenarios (4 tests)
- [ ] Epic 3 UAT scenarios (3 tests)
- [ ] UAT documentation and runbook

**Dependencies:** Phase 3

### Phase 5: Polish & Documentation (Story Points: 5)

**Deliverables:**
- [ ] Test coverage report integration
- [ ] Test documentation (TESTING.md)
- [ ] Local development testing guide
- [ ] CI/CD optimization (parallelization, caching)

**Dependencies:** Phase 4

---

## Success Metrics

| Metric | Current | Target | Measurement |
|--------|---------|--------|-------------|
| Overall test coverage | 63.6% | 80% | `go test -cover` |
| `cmd/` package coverage | 51.2% | 80% | Per-package coverage |
| Integration test coverage | 0% | 80% of `run*` functions | Query/mutation coverage |
| Command test coverage | Partial | 100% of commands | At least 1 test per command |
| UAT coverage | 0% | 100% of PRD acceptance criteria | Traced to PRD |
| Test execution time | ~2s | < 5 minutes (with integration) | CI/CD duration |
| Test reliability | N/A | < 1% flaky rate | Failure tracking |
| Cleanup success rate | N/A | 100% | No orphaned resources |

---

## Risks & Mitigations

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| GitHub API rate limiting | Tests fail or slow down | Medium | Use caching, batch operations, retry logic |
| Test fixture corruption | Tests become unreliable | Low | Reset fixtures before each run |
| Flaky tests from network | False failures | Medium | Retry logic, longer timeouts |
| Token exposure | Security breach | Low | Use GitHub secrets, rotate regularly |
| Test org costs | Unexpected charges | Low | Monitor usage, use free tier limits |

---

## Open Questions

| # | Question | Impact | Owner |
|---|----------|--------|-------|
| 1 | Use existing org or create dedicated test org? | Setup complexity | TBD |
| 2 | Should UAT tests run on every PR or just main? | CI time vs coverage | TBD |
| 3 | How to handle GitHub API preview feature changes? | Test stability | TBD |
| 4 | Should we mock any API calls for speed? | Test fidelity vs speed | TBD |

---

## Appendix A: Test File Structure

```
gh-pm-unified/
├── internal/
│   ├── api/
│   │   ├── queries_integration_test.go     # //go:build integration
│   │   └── mutations_integration_test.go   # //go:build integration
│   ├── testutil/
│   │   ├── testutil.go                     # Test utilities
│   │   ├── fixtures.go                     # Fixture management
│   │   └── cleanup.go                      # Resource cleanup
│   └── config/
│       └── config_integration_test.go      # //go:build integration
├── cmd/
│   ├── list_integration_test.go            # //go:build integration
│   ├── create_integration_test.go          # //go:build integration
│   ├── move_integration_test.go            # //go:build integration
│   └── ...
├── test/
│   ├── uat/
│   │   ├── epic1_test.go                   # //go:build uat
│   │   ├── epic2_test.go                   # //go:build uat
│   │   └── epic3_test.go                   # //go:build uat
│   └── fixtures/
│       ├── test-project.yml                # Project fixture definition
│       └── seed-issues.yml                 # Issue seed data
├── testdata/
│   └── configs/
│       ├── valid-config.yml                # Test configs
│       └── invalid-config.yml
└── TESTING.md                              # Test documentation
```

---

## Appendix B: Makefile Targets

```makefile
# Test targets
.PHONY: test test-unit test-integration test-uat test-all

test: test-unit                              ## Run unit tests only (default)

test-unit:                                   ## Run unit tests
	go test -v ./...

test-integration:                            ## Run integration tests (requires TEST_* env vars)
	go test -v -tags=integration ./...

test-uat:                                    ## Run UAT tests (requires TEST_* env vars)
	go test -v -tags=uat ./...

test-all: test-unit test-integration test-uat ## Run all tests

test-coverage:                               ## Run tests with coverage
	go test -coverprofile=coverage.out ./...
	go test -tags=integration -coverprofile=coverage-integration.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-setup:                                  ## Setup test fixtures
	go run ./cmd/testsetup/main.go

test-cleanup:                                ## Cleanup test resources
	go run ./cmd/testcleanup/main.go
```

---

## Appendix C: runCreate Integration Test Requirements

The `runCreate` function in `cmd/create.go` currently has 44.9% unit test coverage. The uncovered portions (lines 93-139) involve actual GitHub API calls that cannot be tested without authenticated access. This section documents the integration test requirements for achieving full coverage.

### Current Coverage Gap

The following code paths require integration testing:

```go
// Line 93: CreateIssue API call
issue, err := client.CreateIssue(owner, repo, title, body, labels)

// Line 99: GetProject API call
project, err := client.GetProject(cfg.Project.Owner, cfg.Project.Number)

// Line 104: AddIssueToProject mutation
itemID, err := client.AddIssueToProject(project.ID, issue.ID)

// Lines 110-135: SetProjectItemField mutations for status/priority
client.SetProjectItemField(project.ID, itemID, "Status", statusValue)
client.SetProjectItemField(project.ID, itemID, "Priority", priorityValue)

// Lines 138-139: Success output
fmt.Printf("Created issue #%d: %s\n", issue.Number, issue.Title)
```

### Proposed Integration Test

```go
//go:build integration

package cmd

import (
    "os"
    "strings"
    "testing"

    "github.com/rubrical-studios/gh-pmu/internal/testutil"
)

func TestRunCreate_Integration_CreatesIssueWithFields(t *testing.T) {
    // ARRANGE: Set up test environment
    testutil.RequireTestEnv(t)

    cfg := testutil.CreateTempConfigForTest(t, testutil.TestConfig{
        ProjectOwner:  os.Getenv("TEST_PROJECT_OWNER"),
        ProjectNumber: os.Getenv("TEST_PROJECT_NUMBER"),
        Repository:    os.Getenv("TEST_REPO_OWNER") + "/" + os.Getenv("TEST_REPO_NAME"),
    })
    defer testutil.CleanupTempConfig(t, cfg)

    uniqueTitle := "Integration-Test-Issue-" + testutil.RandomSuffix()

    // ACT: Run create command
    output, err := testutil.RunCommandInDir(t, cfg.Dir,
        "gh", "pmu", "create",
        "--title", uniqueTitle,
        "--body", "Created by integration test",
        "--status", "Todo",
        "--priority", "Medium",
        "--label", "test-label",
    )

    // ASSERT: Command succeeded
    if err != nil {
        t.Fatalf("create command failed: %v\nOutput: %s", err, output)
    }

    // Extract issue number from output
    issueNum := testutil.ExtractIssueNumber(t, output)
    defer testutil.DeleteIssue(t, issueNum) // Cleanup

    // Verify issue was created
    if !strings.Contains(output, "Created issue #") {
        t.Errorf("Expected success message, got: %s", output)
    }

    // Verify issue is in project with correct fields
    item := testutil.GetProjectItem(t, issueNum)

    if item.Fields["Status"] != "Todo" {
        t.Errorf("Expected Status='Todo', got '%s'", item.Fields["Status"])
    }
    if item.Fields["Priority"] != "Medium" {
        t.Errorf("Expected Priority='Medium', got '%s'", item.Fields["Priority"])
    }

    // Verify labels were applied
    issue := testutil.GetIssue(t, issueNum)
    if !testutil.HasLabel(issue, "test-label") {
        t.Errorf("Expected issue to have 'test-label'")
    }
}

func TestRunCreate_Integration_AppliesDefaultsFromConfig(t *testing.T) {
    testutil.RequireTestEnv(t)

    cfg := testutil.CreateTempConfigForTest(t, testutil.TestConfig{
        ProjectOwner:  os.Getenv("TEST_PROJECT_OWNER"),
        ProjectNumber: os.Getenv("TEST_PROJECT_NUMBER"),
        Repository:    os.Getenv("TEST_REPO_OWNER") + "/" + os.Getenv("TEST_REPO_NAME"),
        Defaults: testutil.Defaults{
            Status:   "Backlog",
            Priority: "Low",
            Labels:   []string{"auto-created"},
        },
    })
    defer testutil.CleanupTempConfig(t, cfg)

    uniqueTitle := "Integration-Test-Defaults-" + testutil.RandomSuffix()

    // ACT: Run create without explicit status/priority
    output, err := testutil.RunCommandInDir(t, cfg.Dir,
        "gh", "pmu", "create",
        "--title", uniqueTitle,
    )

    if err != nil {
        t.Fatalf("create command failed: %v", err)
    }

    issueNum := testutil.ExtractIssueNumber(t, output)
    defer testutil.DeleteIssue(t, issueNum)

    // ASSERT: Defaults were applied
    item := testutil.GetProjectItem(t, issueNum)

    if item.Fields["Status"] != "Backlog" {
        t.Errorf("Expected default Status='Backlog', got '%s'", item.Fields["Status"])
    }
    if item.Fields["Priority"] != "Low" {
        t.Errorf("Expected default Priority='Low', got '%s'", item.Fields["Priority"])
    }

    issue := testutil.GetIssue(t, issueNum)
    if !testutil.HasLabel(issue, "auto-created") {
        t.Errorf("Expected default label 'auto-created'")
    }
}

func TestRunCreate_Integration_MergesLabels(t *testing.T) {
    testutil.RequireTestEnv(t)

    cfg := testutil.CreateTempConfigForTest(t, testutil.TestConfig{
        ProjectOwner:  os.Getenv("TEST_PROJECT_OWNER"),
        ProjectNumber: os.Getenv("TEST_PROJECT_NUMBER"),
        Repository:    os.Getenv("TEST_REPO_OWNER") + "/" + os.Getenv("TEST_REPO_NAME"),
        Defaults: testutil.Defaults{
            Labels: []string{"pm-tracked"},
        },
    })
    defer testutil.CleanupTempConfig(t, cfg)

    uniqueTitle := "Integration-Test-Labels-" + testutil.RandomSuffix()

    // ACT: Run create with additional labels
    output, err := testutil.RunCommandInDir(t, cfg.Dir,
        "gh", "pmu", "create",
        "--title", uniqueTitle,
        "--label", "bug",
        "--label", "urgent",
    )

    if err != nil {
        t.Fatalf("create command failed: %v", err)
    }

    issueNum := testutil.ExtractIssueNumber(t, output)
    defer testutil.DeleteIssue(t, issueNum)

    // ASSERT: Both default and CLI labels present
    issue := testutil.GetIssue(t, issueNum)

    expectedLabels := []string{"pm-tracked", "bug", "urgent"}
    for _, label := range expectedLabels {
        if !testutil.HasLabel(issue, label) {
            t.Errorf("Expected label '%s' to be present", label)
        }
    }
}

func TestRunCreate_Integration_FieldValueAliases(t *testing.T) {
    testutil.RequireTestEnv(t)

    cfg := testutil.CreateTempConfigForTest(t, testutil.TestConfig{
        ProjectOwner:  os.Getenv("TEST_PROJECT_OWNER"),
        ProjectNumber: os.Getenv("TEST_PROJECT_NUMBER"),
        Repository:    os.Getenv("TEST_REPO_OWNER") + "/" + os.Getenv("TEST_REPO_NAME"),
        Fields: map[string]testutil.FieldConfig{
            "status": {
                Field: "Status",
                Values: map[string]string{
                    "todo": "Todo",
                    "wip":  "In Progress",
                    "done": "Done",
                },
            },
            "priority": {
                Field: "Priority",
                Values: map[string]string{
                    "p0": "Critical",
                    "p1": "High",
                    "p2": "Medium",
                },
            },
        },
    })
    defer testutil.CleanupTempConfig(t, cfg)

    uniqueTitle := "Integration-Test-Aliases-" + testutil.RandomSuffix()

    // ACT: Use aliases instead of actual field values
    output, err := testutil.RunCommandInDir(t, cfg.Dir,
        "gh", "pmu", "create",
        "--title", uniqueTitle,
        "--status", "wip",      // Alias for "In Progress"
        "--priority", "p1",     // Alias for "High"
    )

    if err != nil {
        t.Fatalf("create command failed: %v", err)
    }

    issueNum := testutil.ExtractIssueNumber(t, output)
    defer testutil.DeleteIssue(t, issueNum)

    // ASSERT: Aliases resolved to actual values
    item := testutil.GetProjectItem(t, issueNum)

    if item.Fields["Status"] != "In Progress" {
        t.Errorf("Expected Status='In Progress' (from alias 'wip'), got '%s'", item.Fields["Status"])
    }
    if item.Fields["Priority"] != "High" {
        t.Errorf("Expected Priority='High' (from alias 'p1'), got '%s'", item.Fields["Priority"])
    }
}
```

### Required Test Utilities

The integration tests require the following additions to `internal/testutil/`:

```go
// RequireTestEnv skips the test if integration test environment is not configured
func RequireTestEnv(t *testing.T) {
    required := []string{
        "TEST_PROJECT_OWNER",
        "TEST_PROJECT_NUMBER",
        "TEST_REPO_OWNER",
        "TEST_REPO_NAME",
    }
    for _, env := range required {
        if os.Getenv(env) == "" {
            t.Skipf("Skipping: %s not set", env)
        }
    }
}

// ExtractIssueNumber parses "Created issue #123" output
func ExtractIssueNumber(t *testing.T, output string) int

// DeleteIssue removes a test issue (cleanup)
func DeleteIssue(t *testing.T, issueNum int)

// GetProjectItem fetches project item with field values
func GetProjectItem(t *testing.T, issueNum int) *ProjectItem

// GetIssue fetches issue details including labels
func GetIssue(t *testing.T, issueNum int) *Issue

// HasLabel checks if issue has a specific label
func HasLabel(issue *Issue, label string) bool
```

### Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `TEST_PROJECT_OWNER` | Owner of test project | `rubrical-studios` |
| `TEST_PROJECT_NUMBER` | Test project number | `99` |
| `TEST_REPO_OWNER` | Owner of test repository | `rubrical-studios` |
| `TEST_REPO_NAME` | Test repository name | `gh-pmu-test` |

### Running the Tests

```bash
# Set environment variables
export TEST_PROJECT_OWNER="rubrical-studios"
export TEST_PROJECT_NUMBER="99"
export TEST_REPO_OWNER="rubrical-studios"
export TEST_REPO_NAME="gh-pmu-test"

# Run integration tests only
go test -v -tags=integration ./cmd/... -run "Integration"

# Run specific create integration tests
go test -v -tags=integration ./cmd/... -run "TestRunCreate_Integration"
```

### Expected Coverage After Integration Tests

With these integration tests in place, `runCreate` coverage should increase from 44.9% to approximately 95%, covering:

- Issue creation via API (line 93)
- Project lookup (line 99)
- Adding issue to project (line 104)
- Setting status field (lines 110-122)
- Setting priority field (lines 124-135)
- Success output (lines 138-139)

The only remaining uncovered paths would be error handling for non-fatal warnings (lines 114, 120, 127, 133).

---

## Appendix D: Functions Below 80% Coverage (v0.2.10)

Based on the coverage report from v0.2.10 (63.6% total coverage), the following functions require additional testing:

### Critical Priority (0% Coverage - Require Integration Tests)

These `run*` functions involve API calls and require integration testing infrastructure:

| File | Function | Coverage | Notes |
|------|----------|----------|-------|
| `cmd/init.go:36` | `runInit` | 0.0% | Interactive I/O + API calls |
| `cmd/init.go:283` | `detectRepository` | 0.0% | Requires git environment |
| `cmd/intake.go:55` | `runIntake` | 0.0% | API calls for untracked issues |
| `cmd/list.go:46` | `runList` | 0.0% | API calls for project items |
| `cmd/list.go:126` | `filterByHasSubIssues` | 0.0% | Filters based on sub-issue data |
| `cmd/split.go:61` | `runSplit` | 0.0% | API calls for issue creation |
| `cmd/sub.go:54` | `runSubAdd` | 0.0% | API calls for linking |
| `cmd/sub.go:202` | `runSubCreate` | 0.0% | API calls for sub-issue creation |
| `cmd/sub.go:337` | `runSubList` | 0.0% | API calls for sub-issue queries |
| `cmd/sub.go:521` | `runSubRemove` | 0.0% | API calls for unlinking |
| `cmd/triage.go:68` | `runTriage` | 0.0% | API calls + rule engine |
| `cmd/view.go:42` | `runView` | 0.0% | API calls for issue details |
| `cmd/root.go:36` | `Execute` | 0.0% | Entry point (main wrapper) |
| `internal/ui/ui.go:270` | `PrintMenu` | 0.0% | Simple wrapper (low priority) |
| `main.go:9` | `main` | 0.0% | Entry point (low priority) |

### Medium Priority (Below 80% Coverage)

| File | Function | Coverage | Gap Analysis |
|------|----------|----------|--------------|
| `cmd/move.go:76` | `runMove` | 16.7% | Entry point; main logic in `runMoveWithDeps` |
| `cmd/create.go:48` | `runCreate` | 44.9% | API calls (see Appendix C) |
| `cmd/sub.go:32` | `newSubAddCommand` | 66.7% | Flag parsing edge cases |
| `cmd/sub.go:499` | `newSubRemoveCommand` | 66.7% | Flag parsing edge cases |
| `cmd/init.go:419` | `writeConfig` | 75.0% | Error paths for file I/O |

### Low Priority (Near 80% Threshold)

| File | Function | Coverage | Notes |
|------|----------|----------|-------|
| `cmd/view.go:19` | `newViewCommand` | 80.0% | At threshold |
| `cmd/sub.go:311` | `newSubListCommand` | 80.0% | At threshold |
| `internal/api/errors.go:46` | `IsRateLimited` | 83.3% | Error handling edge case |
| `internal/api/errors.go:59` | `IsAuthError` | 83.3% | Error handling edge case |
| `internal/config/config.go:144` | `GetFieldName` | 83.3% | Lookup edge cases |
| `cmd/triage.go:192` | `listTriageConfigs` | 83.3% | Minimal gap |
| `internal/api/mutations.go:382` | `AddLabelToIssue` | 66.7% | API error handling |
| `internal/api/mutations.go:234` | `setNumberField` | 85.7% | Error paths |
| `internal/api/errors.go:73` | `WrapError` | 85.7% | Error wrapping edge cases |
| `cmd/init.go:490` | `writeConfigWithMetadata` | 85.7% | Error paths |
| `cmd/intake.go:21` | `newIntakeCommand` | 85.7% | Flag edge cases |
| `cmd/split.go:22` | `newSplitCommand` | 85.7% | Flag edge cases |
| `cmd/triage.go:33` | `newTriageCommand` | 87.5% | Flag edge cases |
| `cmd/list.go:22` | `newListCommand` | 87.5% | Flag edge cases |

### Summary by Package

| Package | Current Coverage | Functions Below 80% | Functions at 0% |
|---------|-----------------|---------------------|-----------------|
| `cmd/` | 51.2% | 18 | 12 |
| `internal/api/` | 96.6% | 4 | 0 |
| `internal/config/` | 97.0% | 1 | 0 |
| `internal/ui/` | 96.9% | 1 | 1 |
| **Total** | **63.6%** | **24** | **13** |

### Recommended Test Priority Order

1. **Phase 1 - Quick Wins (Unit Tests)**
   - `writeConfig` error paths
   - `newSub*Command` flag edge cases
   - `new*Command` flag edge cases
   - API error handling functions

2. **Phase 2 - Integration Test Infrastructure**
   - Set up test fixtures (see Phase 1 in main proposal)
   - Implement `internal/testutil/` package

3. **Phase 3 - Run Function Integration Tests**
   - `runCreate` (Appendix C)
   - `runList`, `runView`
   - `runMove`, `runTriage`
   - `runSplit`, `runIntake`
   - `runSub*` commands

4. **Phase 4 - Complex Functions**
   - `runInit` (requires interactive I/O mocking)
   - `detectRepository` (requires git environment)

---

## Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2025-12-03 | PRD-Analyst | Initial proposal |
| 1.1 | 2025-12-03 | API-Integration-Specialist | Added Appendix C: runCreate integration test requirements |
| 1.2 | 2025-12-04 | API-Integration-Specialist | Updated executive summary with current coverage (63.6%), added Appendix D with functions below 80% coverage based on v0.2.10 release |
| 1.3 | 2025-12-04 | API-Integration-Specialist | Changed CI workflow to on-demand only (workflow_dispatch), removed cron schedule |

