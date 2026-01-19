# Proposal: Go-based E2E Test Script with Test Project Population

**Version:** 1.3
**Date:** 2026-01-17
**Author:** Backend-Specialist
**Status:** Ready for Implementation

---

## Executive Summary

This proposal outlines the implementation of a Go-based end-to-end (E2E) test suite that validates complete workflows against a dedicated test project. The tests target commands that are difficult to unit test due to external process calls, complex state management, and API interactions (e.g., issue #551 - batch mutation JSON encoding bug).

### Target Test Infrastructure

| Resource | Details | Status |
|----------|---------|--------|
| **Test Project** | #41 (IDPF-gh-pmu-testing) - Private | ✅ Verified |
| **Test Repo** | `rubrical-studios/gh-pmu-e2e-test` - Private | ✅ Verified |

> **Note:** Infrastructure verified 2026-01-17. Fields (Release, Microsprint) created via `gh pmu init`.

### Key Benefits

| Benefit | Impact |
|---------|--------|
| Coverage contribution | Tests run via `go test -tags=e2e -cover` |
| Windows support | Native Go cross-platform execution |
| Catches API bugs | Would have caught #551 (HTTP 400 batch mutation) |
| Validates workflows | Microsprint, branch, board - 0 integration tests currently |
| Release integration | Pre-merge gate via `/prepare-release`, `/merge-branch`, `/prepare-beta` |

---

## Scope

### In Scope

- E2E test files with `//go:build e2e` tag in `test/e2e/`
- Test project field verification (created via `gh pmu init`)
- Coverage of untested commands: `microsprint`, `branch`, `board`, `filter`
- Windows, macOS, and Linux compatibility
- Manual execution via `go test -tags=e2e`
- Separate cleanup via `E2E_CLEANUP=true` environment variable
- Release command extensions for pre-merge E2E gates
- E2E impact analysis scripts for change detection

### Out of Scope

- CI/CD integration (future phase)
- Interactive prompt testing (use `--skip-retro` etc.)
- `--config` flag (temp directory approach sufficient)
- Testing `gh pmu init` (write config directly)

---

## Architecture

### Execution Model

**Sequential execution** - Tests run one at a time (no `t.Parallel()`) to avoid conflicts on shared project resources (microsprints, branches, issues).

### Directory Structure

```
test/e2e/
├── e2e_test.go             # Main test file with //go:build e2e
├── setup_test.go           # Project/field verification and config generation
├── cleanup_test.go         # Resource cleanup utilities
├── microsprint_test.go     # Microsprint workflow tests
├── branch_test.go          # Branch/release workflow tests
├── board_test.go           # Board rendering tests
├── filter_test.go          # Filter command tests
├── workflow_test.go        # Multi-command workflow tests
└── testdata/
    └── seed-issues.json    # Issue fixtures (titles, labels, etc.)

.claude/scripts/e2e/
├── analyze-e2e-impact.js   # Commit analysis for E2E test impact
└── run-e2e-gate.js         # E2E test runner for release gates
```

### Test Execution Flow

```
┌─────────────────────────────────────────────────────────────┐
│                    E2E Test Execution                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  1. Build local binary (TestMain)                            │
│     └─> go build -o test/e2e/gh-pmu-test .                   │
│                                                              │
│  2. Create temp directory                                    │
│     └─> Isolated working directory for config                │
│                                                              │
│  3. Generate .gh-pmu.yml                                     │
│     └─> Write config programmatically (skip init)            │
│                                                              │
│  4. Verify project fields                                    │
│     └─> Ensure Status, Priority, Release, Microsprint exist  │
│                                                              │
│  5. Run test suites (sequential, no t.Parallel)              │
│     ├─> Microsprint tests (--skip-retro)                    │
│     ├─> Branch tests                                        │
│     ├─> Board tests                                         │
│     ├─> Filter tests                                        │
│     └─> Workflow tests                                      │
│                                                              │
│  6. Report results                                           │
│     └─> Pass/fail summary, error details on failure          │
│                                                              │
│  7. Cleanup (manual, via E2E_CLEANUP=true)                   │
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

### Binary Build (TestMain)

Tests build a local binary once before all tests run:

```go
var testBinary string

func TestMain(m *testing.M) {
    // Build the binary from project root
    testBinary = filepath.Join(".", "gh-pmu-test")
    if runtime.GOOS == "windows" {
        testBinary += ".exe"
    }

    // Build with coverage instrumentation
    cmd := exec.Command("go", "build", "-cover", "-o", testBinary, ".")
    cmd.Dir = filepath.Join("..", "..")  // Project root from test/e2e/
    if out, err := cmd.CombinedOutput(); err != nil {
        fmt.Fprintf(os.Stderr, "Failed to build test binary: %v\n%s", err, out)
        os.Exit(1)
    }

    code := m.Run()

    // Cleanup binary
    os.Remove(testBinary)
    os.Exit(code)
}
```

### Command Execution

Commands run the local binary from temp directory:

```go
func runPMU(t *testing.T, workDir string, args ...string) string {
    cmd := exec.Command(testBinary, args...)
    cmd.Dir = workDir  // Config discovery starts here

    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    err := cmd.Run()
    if err != nil {
        t.Logf("Command failed: %s %s\nStderr: %s",
            testBinary, strings.Join(args, " "), stderr.String())
    }

    return stdout.String()
}
```

---

## Project Field Setup

Tests verify required fields exist (created via `gh pmu init`):

```go
func verifyProjectFields(t *testing.T) {
    requiredFields := []struct {
        Name     string
        Type     string
        Options  []string
    }{
        {"Status", "SINGLE_SELECT", []string{"Backlog", "Ready", "In Progress", "In Review", "Done"}},
        {"Priority", "SINGLE_SELECT", []string{"P0", "P1", "P2", "P3"}},
        {"Release", "TEXT", nil},
        {"Microsprint", "TEXT", nil},
    }

    for _, field := range requiredFields {
        if !fieldExists(field.Name) {
            t.Fatalf("Required field %q not found. Run 'gh pmu init' on test project first.", field.Name)
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

```powershell
# Run cleanup (PowerShell)
$env:E2E_CLEANUP="true"; go test -tags=e2e -run TestCleanup ./test/e2e/
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
2. Access to test project #41 and test repo
3. Go 1.22+ installed (for building test binary)

### Running Tests

```powershell
# Run all E2E tests
go test -tags=e2e -v ./test/e2e/

# Run specific test
go test -tags=e2e -v -run TestMicrosprintLifecycle ./test/e2e/

# Run cleanup
$env:E2E_CLEANUP="true"; go test -tags=e2e -run TestCleanup ./test/e2e/
```

### Coverage (Merged with Unit Tests)

```powershell
# Set coverage output directory
$env:GOCOVERDIR="coverage"
mkdir -p coverage

# Run unit tests with coverage
go test -cover -coverprofile=coverage/unit.out ./...

# Run E2E tests (binary built with -cover writes to GOCOVERDIR)
go test -tags=e2e ./test/e2e/

# Merge coverage reports
go tool covdata textfmt -i=coverage -o=coverage/merged.out

# View combined coverage
go tool cover -func=coverage/merged.out
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `E2E_CLEANUP` | `false` | Set to `true` to enable cleanup test |
| `GOCOVERDIR` | - | Directory for coverage data from instrumented binary |

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

> **Delivery:** Single PR for all phases.

### Phase 1: Infrastructure

- [ ] Create `test/e2e/` directory structure
- [ ] Implement TestMain with local binary build
- [ ] Implement test helpers (runPMU, assertContains, etc.)
- [ ] Implement config generation
- [ ] Implement field verification (not creation)
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

### Phase 4: Release Command Extensions

- [ ] Create `.claude/scripts/e2e/` directory
- [ ] Implement `analyze-e2e-impact.js` script
- [ ] Implement `run-e2e-gate.js` script
- [ ] Add E2E impact analysis to `/prepare-release` `post-analysis`
- [ ] Add E2E impact analysis to `/prepare-beta` `post-analysis`
- [ ] Add E2E gate to `/prepare-release` `post-validation`
- [ ] Add E2E gate to `/prepare-beta` `post-validation`
- [ ] Add E2E gate to `/merge-branch` `gates`

---

## Acceptance Criteria

### E2E Test Suite
- [ ] E2E tests run successfully on Windows via `go test -tags=e2e ./test/e2e/`
- [ ] Tests use locally-built binary (not installed extension)
- [ ] Tests run sequentially (no `t.Parallel()`)
- [ ] Coverage merges with unit tests via GOCOVERDIR
- [ ] All target commands have at least one E2E test
- [ ] Cleanup removes all `[E2E]` prefixed test issues
- [ ] Tests use temp directory for config isolation
- [ ] Tests validate key content presence (not exact formatting)

### Release Command Extensions
- [ ] `analyze-e2e-impact.js` identifies changes to E2E-tested commands
- [ ] `run-e2e-gate.js` runs E2E tests and reports JSON result
- [ ] `/prepare-release` runs E2E impact analysis and gate
- [ ] `/prepare-beta` runs E2E impact analysis and gate
- [ ] `/merge-branch` runs E2E tests as a gate check
- [ ] `--skip-e2e` flag bypasses E2E gate when needed

---

## Release Command Extensions

E2E tests integrate with release workflows via extension points.

### Extension Point Usage

| Command | Extension Point | Purpose |
|---------|-----------------|---------|
| `/prepare-release` | `post-analysis` | Evaluate E2E test impact |
| `/prepare-release` | `post-validation` | Run E2E tests before prepare |
| `/merge-branch` | `gates` | Run E2E tests as merge gate |
| `/prepare-beta` | `post-analysis` | Evaluate E2E test impact |
| `/prepare-beta` | `post-validation` | Run E2E tests before tag |

### Extension 1: E2E Test Impact Analysis (`post-analysis`)

Analyzes commits to identify changes requiring E2E test updates:

```markdown
<!-- USER-EXTENSION-START: post-analysis -->
### E2E Test Impact Analysis

```bash
node .claude/scripts/e2e/analyze-e2e-impact.js
```

The script checks commits for changes to E2E-tested commands:
- `cmd/microsprint/` → `test/e2e/microsprint_test.go`
- `cmd/branch/` → `test/e2e/branch_test.go`
- `cmd/board/` → `test/e2e/board_test.go`
- `cmd/filter/` → `test/e2e/filter_test.go`

Output:
```json
{
  "impactedTests": ["microsprint", "board"],
  "newCommandsWithoutTests": [],
  "recommendation": "Review microsprint_test.go for coverage of new --flag"
}
```

**If `newCommandsWithoutTests` is non-empty, warn user.**
<!-- USER-EXTENSION-END: post-analysis -->
```

### Extension 2: E2E Gate (`post-validation` / `gates`)

Runs E2E tests as a pre-merge validation gate:

```markdown
<!-- For /prepare-release and /prepare-beta post-validation -->
### E2E Test Gate

**If `--skip-e2e` was passed, skip this section.**

```bash
node .claude/scripts/e2e/run-e2e-gate.js
```

The script:
1. Builds local binary with coverage
2. Runs E2E tests: `go test -tags=e2e ./test/e2e/`
3. Reports pass/fail with summary

Output:
```json
{
  "success": true,
  "testsRun": 12,
  "testsPassed": 12,
  "duration": "45s"
}
```

**If `success` is false, STOP and report failing tests.**
```

```markdown
<!-- For /merge-branch gates -->
#### Gate: E2E Tests Pass

```bash
node .claude/scripts/e2e/run-e2e-gate.js
```

**FAIL if E2E tests fail.**
```

### Script: `analyze-e2e-impact.js`

```javascript
// .claude/scripts/e2e/analyze-e2e-impact.js
const { execSync } = require('child_process');

const E2E_COVERAGE_MAP = {
  'cmd/microsprint': 'microsprint_test.go',
  'cmd/branch': 'branch_test.go',
  'cmd/board': 'board_test.go',
  'cmd/filter': 'filter_test.go',
  'cmd/move': 'workflow_test.go',
  'cmd/sub': 'workflow_test.go',
  'cmd/create': 'workflow_test.go',
};

function getChangedFiles() {
  const lastTag = execSync('git describe --tags --abbrev=0', { encoding: 'utf8' }).trim();
  const files = execSync(`git diff --name-only ${lastTag}..HEAD`, { encoding: 'utf8' });
  return files.split('\n').filter(Boolean);
}

function analyzeImpact() {
  const changed = getChangedFiles();
  const impactedTests = new Set();

  for (const file of changed) {
    for (const [cmdPath, testFile] of Object.entries(E2E_COVERAGE_MAP)) {
      if (file.startsWith(cmdPath)) {
        impactedTests.add(testFile.replace('_test.go', ''));
      }
    }
  }

  // Check for new commands without E2E coverage
  const newCommands = changed
    .filter(f => f.startsWith('cmd/') && f.endsWith('.go'))
    .map(f => f.split('/')[1])
    .filter(cmd => !Object.keys(E2E_COVERAGE_MAP).some(p => p.includes(cmd)));

  console.log(JSON.stringify({
    impactedTests: [...impactedTests],
    newCommandsWithoutTests: [...new Set(newCommands)],
    recommendation: impactedTests.size > 0
      ? `Review ${[...impactedTests].join(', ')} tests for coverage of changes`
      : 'No E2E test impact detected'
  }));
}

analyzeImpact();
```

### Script: `run-e2e-gate.js`

```javascript
// .claude/scripts/e2e/run-e2e-gate.js
const { execSync } = require('child_process');
const path = require('path');

function runE2EGate() {
  const start = Date.now();

  try {
    // Run E2E tests
    const output = execSync('go test -tags=e2e -v ./test/e2e/', {
      encoding: 'utf8',
      cwd: path.resolve(__dirname, '../../..'),
      env: { ...process.env, GOCOVERDIR: 'coverage' }
    });

    // Parse test results
    const passed = (output.match(/--- PASS:/g) || []).length;
    const failed = (output.match(/--- FAIL:/g) || []).length;

    const duration = Math.round((Date.now() - start) / 1000);

    console.log(JSON.stringify({
      success: failed === 0,
      testsRun: passed + failed,
      testsPassed: passed,
      testsFailed: failed,
      duration: `${duration}s`
    }));

    process.exit(failed > 0 ? 1 : 0);
  } catch (err) {
    console.log(JSON.stringify({
      success: false,
      error: err.message,
      output: err.stdout?.toString() || ''
    }));
    process.exit(1);
  }
}

runE2EGate();
```

### Configuration

Add to `.gh-pmu.yml`:

```yaml
e2e:
  enabled: true
  skip_on_flags:
    - --skip-e2e
  timeout: 300  # seconds
```

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

## References

- Issue #551: Batch mutation bug (would be caught by E2E)
- Issue #555: This proposal tracking issue
- Existing test utilities: `internal/testutil/testutil.go`
- Test project: https://github.com/users/rubrical-studios/projects/41
- Test repo: https://github.com/rubrical-studios/gh-pmu-e2e-test

---

*Proposal approved. Ready for implementation.*
