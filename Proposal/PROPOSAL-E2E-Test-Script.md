# Proposal: Go-based E2E Test Script with Test Project Population

**Version:** 1.1
**Date:** 2026-01-17
**Author:** Backend-Specialist
**Status:** Draft

---

## Executive Summary

This proposal outlines the implementation of a Go-based end-to-end (E2E) test suite that validates complete workflows against a dedicated test project. The tests target commands that are difficult to unit test due to external process calls, complex state management, and API interactions (e.g., issue #551 - batch mutation JSON encoding bug).

### Target Test Infrastructure

| Resource | Details |
|----------|---------|
| **Test Project** | #41 (IDPF-gh-pmu-testing) - Private |
| **Test Repo** | `rubrical-studios/gh-pmu-e2e-test` - Private, empty |

### Key Benefits

| Benefit | Impact |
|---------|--------|
| Coverage contribution | Tests run via `go test -tags=e2e -cover` |
| Windows support | Native Go cross-platform execution |
| Catches API bugs | Would have caught #551 (HTTP 400 batch mutation) |
| Validates workflows | Microsprint, branch, board - 0 integration tests currently |

---

## Scope

### In Scope

- E2E test files with `//go:build e2e` tag in `cmd/e2etest/`
- Test project field creation/verification
- Coverage of untested commands: `microsprint`, `branch`, `board`, `filter`
- Windows, macOS, and Linux compatibility
- Manual execution via `go test -tags=e2e`
- Separate `--cleanup` command for test resource removal

### Out of Scope

- CI/CD integration (future phase)
- Interactive prompt testing (use `--skip-retro` etc.)
- `--config` flag (temp directory approach sufficient)
- Testing `gh pmu init` (write config directly)

---

## Architecture

### Directory Structure

```
cmd/e2etest/
├── e2e_test.go             # Main test file with //go:build e2e
├── setup_test.go           # Project/field setup and config generation
├── cleanup_test.go         # Resource cleanup utilities
├── microsprint_test.go     # Microsprint workflow tests
├── branch_test.go          # Branch/release workflow tests
├── board_test.go           # Board rendering tests
├── filter_test.go          # Filter command tests
├── workflow_test.go        # Multi-command workflow tests
└── testdata/
    └── seed-issues.json    # Issue fixtures (titles, labels, etc.)
```

### Test Execution Flow

```
┌─────────────────────────────────────────────────────────────┐
│                    E2E Test Execution                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  1. Create temp directory                                    │
│     └─> Isolated working directory for config                │
│                                                              │
│  2. Generate .gh-pmu.yml                                     │
│     └─> Write config programmatically (skip init)            │
│                                                              │
│  3. Verify/Create project fields                             │
│     └─> Ensure Status, Priority, Branch, Microsprint exist   │
│                                                              │
│  4. Run test suites                                          │
│     ├─> Microsprint tests (--skip-retro)                    │
│     ├─> Branch tests                                        │
│     ├─> Board tests                                         │
│     ├─> Filter tests                                        │
│     └─> Workflow tests                                      │
│                                                              │
│  5. Report results                                           │
│     └─> Pass/fail summary, error details on failure          │
│                                                              │
│  6. Cleanup (manual, via -cleanup flag)                      │
│     └─> Remove [E2E] prefixed issues                        │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Test Configuration

### Config Generation

Tests generate `.gh-pmu.yml` programmatically in a temp directory:

```go
func setupTestConfig(t *testing.T) string {
    tempDir := t.TempDir()

    config := `project:
  owner: rubrical-studios
  number: 41

repositories:
  - rubrical-studios/gh-pmu-e2e-test

fields:
  status:
    field: Status
    values:
      backlog: Backlog
      ready: Ready
      in_progress: In Progress
      in_review: In Review
      done: Done
  priority:
    field: Priority
    values:
      p0: P0
      p1: P1
      p2: P2
      p3: P3
`
    configPath := filepath.Join(tempDir, ".gh-pmu.yml")
    os.WriteFile(configPath, []byte(config), 0644)

    return tempDir
}
```

### Command Execution

Commands run from temp directory to find correct config:

```go
func runPMU(t *testing.T, workDir string, args ...string) string {
    cmd := exec.Command("gh", append([]string{"pmu"}, args...)...)
    cmd.Dir = workDir  // Config discovery starts here

    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err := cmd.Run()
    if err != nil {
        t.Logf("Command failed: gh pmu %s\nStderr: %s",
            strings.Join(args, " "), stderr.String())
    }

    return stdout.String()
}
```

---

## Project Field Setup

Tests verify/create required fields on first run:

```go
func ensureProjectFields(t *testing.T) {
    requiredFields := []struct {
        Name     string
        Type     string
        Options  []string
    }{
        {"Status", "SINGLE_SELECT", []string{"Backlog", "Ready", "In Progress", "In Review", "Done"}},
        {"Priority", "SINGLE_SELECT", []string{"P0", "P1", "P2", "P3"}},
        {"Branch", "TEXT", nil},
        {"Microsprint", "TEXT", nil},
    }

    for _, field := range requiredFields {
        if !fieldExists(field.Name) {
            createField(t, field)
        }
    }
}
```

---

## E2E Test Categories

### Category 1: Microsprint Workflows

**Commands Tested:** `microsprint start`, `add`, `current`, `list`, `close`

```go
//go:build e2e

func TestMicrosprintLifecycle(t *testing.T) {
    workDir := setupTestConfig(t)

    // 1. Start a new microsprint
    output := runPMU(t, workDir, "microsprint", "start", "--name", "e2e-test-sprint")
    assertContains(t, output, "Started microsprint")

    // 2. Verify current shows the sprint
    output = runPMU(t, workDir, "microsprint", "current")
    assertContains(t, output, "e2e-test-sprint")

    // 3. Add issue to microsprint
    issueNum := createTestIssue(t, "[E2E] Microsprint Test Issue")
    defer markForCleanup(t, issueNum)

    runPMU(t, workDir, "microsprint", "add", fmt.Sprintf("%d", issueNum))

    // 4. List microsprint issues
    output = runPMU(t, workDir, "microsprint", "list")
    assertContains(t, output, "Microsprint Test Issue")

    // 5. Close microsprint (skip retro for automated test)
    output = runPMU(t, workDir, "microsprint", "close", "--skip-retro")
    assertContains(t, output, "Closed microsprint")
}
```

### Category 2: Branch Workflows

**Commands Tested:** `branch start`, `current`, `list`, `close`

```go
//go:build e2e

func TestBranchLifecycle(t *testing.T) {
    workDir := setupTestConfig(t)
    branchName := fmt.Sprintf("release/e2e-test-%d", time.Now().Unix())

    // 1. Start a new branch
    output := runPMU(t, workDir, "branch", "start", "--branch", branchName)
    assertContains(t, output, "Started branch")

    // 2. Verify current shows the branch
    output = runPMU(t, workDir, "branch", "current")
    assertContains(t, output, branchName)

    // 3. Add issue to branch
    issueNum := createTestIssue(t, "[E2E] Branch Test Issue")
    defer markForCleanup(t, issueNum)

    output = runPMU(t, workDir, "move", fmt.Sprintf("%d", issueNum), "--branch", "current")
    assertNotContains(t, output, "Error")

    // 4. List branches
    output = runPMU(t, workDir, "branch", "list")
    assertContains(t, output, branchName)

    // 5. Close branch
    output = runPMU(t, workDir, "branch", "close")
    assertContains(t, output, "Closed branch")
}
```

### Category 3: Board Rendering

**Commands Tested:** `board`

```go
//go:build e2e

func TestBoardRendersAllStatuses(t *testing.T) {
    workDir := setupTestConfig(t)

    // Create issues in different statuses
    issue1 := createTestIssueWithStatus(t, "[E2E] Board Backlog", "Backlog")
    issue2 := createTestIssueWithStatus(t, "[E2E] Board InProgress", "In Progress")
    issue3 := createTestIssueWithStatus(t, "[E2E] Board Done", "Done")
    defer markForCleanup(t, issue1, issue2, issue3)

    output := runPMU(t, workDir, "board")

    // Verify column headers present
    assertContains(t, output, "Backlog")
    assertContains(t, output, "In Progress")
    assertContains(t, output, "Done")

    // Verify issues appear
    assertContains(t, output, "Board Backlog")
    assertContains(t, output, "Board InProgress")
    assertContains(t, output, "Board Done")
}
```

### Category 4: Multi-Command Workflows

**Tests complete user journeys and sub-issue creation.**

```go
//go:build e2e

func TestCreateToCloseWorkflow(t *testing.T) {
    workDir := setupTestConfig(t)
    title := fmt.Sprintf("[E2E] Workflow Test %d", time.Now().Unix())

    // 1. Create new issue
    output := runPMU(t, workDir, "create",
        "--title", title,
        "--status", "backlog",
        "--priority", "p2")

    issueNum := extractIssueNumber(t, output)
    defer markForCleanup(t, issueNum)

    // 2. Move through workflow
    runPMU(t, workDir, "move", fmt.Sprintf("%d", issueNum), "--status", "in_progress")
    runPMU(t, workDir, "move", fmt.Sprintf("%d", issueNum), "--status", "in_review")
    runPMU(t, workDir, "move", fmt.Sprintf("%d", issueNum), "--status", "done")

    // 3. Verify final state
    output = runPMU(t, workDir, "view", fmt.Sprintf("%d", issueNum))
    assertContains(t, output, "Done")
}

func TestSubIssueWorkflow(t *testing.T) {
    workDir := setupTestConfig(t)

    // 1. Create parent issue
    output := runPMU(t, workDir, "create", "--title", "[E2E] Parent Issue")
    parentNum := extractIssueNumber(t, output)
    defer markForCleanup(t, parentNum)

    // 2. Create sub-issue (tests sub create)
    output = runPMU(t, workDir, "sub", "create",
        "--parent", fmt.Sprintf("%d", parentNum),
        "--title", "[E2E] Sub Issue 1")
    subNum := extractIssueNumber(t, output)
    defer markForCleanup(t, subNum)

    // 3. List sub-issues (tests sub list)
    output = runPMU(t, workDir, "sub", "list", fmt.Sprintf("%d", parentNum))
    assertContains(t, output, "Sub Issue 1")

    // 4. Remove sub-issue (tests sub remove)
    output = runPMU(t, workDir, "sub", "remove",
        fmt.Sprintf("%d", parentNum),
        fmt.Sprintf("%d", subNum))
    assertNotContains(t, output, "Error")
}
```

---

## Output Validation

Tests validate **key content presence**, not exact formatting:

```go
func assertContains(t *testing.T, output, expected string) {
    t.Helper()
    if !strings.Contains(output, expected) {
        t.Errorf("Expected output to contain %q\nGot: %s", expected, output)
    }
}

func assertNotContains(t *testing.T, output, unexpected string) {
    t.Helper()
    if strings.Contains(output, unexpected) {
        t.Errorf("Expected output NOT to contain %q\nGot: %s", unexpected, output)
    }
}
```

---

## Cleanup

Test resources persist for inspection. Cleanup is manual:

```bash
# Run cleanup
go test -tags=e2e -run TestCleanup ./cmd/e2etest/
```

```go
//go:build e2e

func TestCleanup(t *testing.T) {
    if os.Getenv("E2E_CLEANUP") != "true" {
        t.Skip("Set E2E_CLEANUP=true to run cleanup")
    }

    // Find all [E2E] prefixed issues
    issues := findE2EIssues(t)

    for _, issue := range issues {
        closeAndDeleteIssue(t, issue.Number)
        t.Logf("Cleaned up issue #%d: %s", issue.Number, issue.Title)
    }

    t.Logf("Cleaned up %d test issues", len(issues))
}
```

---

## Usage

### Prerequisites

1. `gh` CLI authenticated (`gh auth login`)
2. `gh pmu` extension installed
3. Access to test project #41 and test repo

### Running Tests

```powershell
# Run all E2E tests
go test -tags=e2e -v ./cmd/e2etest/

# Run with coverage
go test -tags=e2e -cover -coverprofile=e2e-coverage.out ./cmd/e2etest/

# Run specific test
go test -tags=e2e -v -run TestMicrosprintLifecycle ./cmd/e2etest/

# Run cleanup
E2E_CLEANUP=true go test -tags=e2e -run TestCleanup ./cmd/e2etest/
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `E2E_CLEANUP` | `false` | Set to `true` to enable cleanup test |

---

## Expected Coverage Impact

### Commands Targeted

| Command | Current Integration Tests | LOC |
|---------|--------------------------|-----|
| `microsprint` | 0 | 1,529 |
| `branch` | 0 | 1,164 |
| `board` | 0 | 439 |
| `filter` | 0 | 326 |

### Coverage Estimate

- Current: 68.4%
- Target after E2E: 72-75%

---

## Implementation Phases

### Phase 1: Infrastructure

- [ ] Create `cmd/e2etest/` directory structure
- [ ] Implement test helpers (runPMU, assertContains, etc.)
- [ ] Implement config generation
- [ ] Implement field setup/verification
- [ ] Implement cleanup utilities

### Phase 2: Core Tests

- [ ] Microsprint lifecycle test
- [ ] Branch lifecycle test
- [ ] Board rendering test
- [ ] Filter command tests

### Phase 3: Workflow Tests

- [ ] Create-to-close workflow test
- [ ] Sub-issue workflow test (create, list, remove)
- [ ] Multi-issue move test

---

## Acceptance Criteria

- [ ] E2E tests run successfully on Windows via `go test -tags=e2e`
- [ ] Tests contribute to coverage report
- [ ] All target commands have at least one E2E test
- [ ] Cleanup removes all `[E2E]` prefixed test issues
- [ ] Tests use temp directory for config isolation
- [ ] Tests validate key content presence (not exact formatting)

---

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Config discovery finds wrong file | Tests fail or affect real project | Use temp directory with `cmd.Dir` |
| Rate limiting | Tests slow or fail | Batch operations, add delays |
| Test pollution | Stale test data accumulates | `[E2E]` prefix, manual cleanup |
| Flaky network | Intermittent failures | Retry logic, longer timeouts |

---

## References

- Issue #551: Batch mutation bug (would be caught by E2E)
- Issue #555: This proposal tracking issue
- Existing test utilities: `internal/testutil/testutil.go`
- Test project: https://github.com/users/rubrical-studios/projects/41
- Test repo: https://github.com/rubrical-studios/gh-pmu-e2e-test

---

*Proposal ready for approval.*
