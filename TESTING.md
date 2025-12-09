# Integration Testing Guide

This guide covers running integration tests for gh-pmu against the GitHub API.

## Test Fixtures

### Test Repository

- **Repository:** `rubrical-studios/gh-pmu-test`
- **Visibility:** Private
- **Purpose:** Isolated environment for integration tests

### Test Project

- **Project:** `gh-pmu-test-prj` (Project #29)
- **Owner:** `rubrical-studios`
- **URL:** https://github.com/users/rubrical-studios/projects/29

#### Project Fields

| Field | Type | Options |
|-------|------|---------|
| Status | Single Select | Backlog, Ready, In progress, In review, Done |
| Priority | Single Select | P0, P1, P2 |
| Size | Single Select | XS, S, M, L, XL |
| Estimate | Number | - |

### Seed Issues

These read-only issues exist for testing list, view, and filter operations:

| Issue | Title | Status | Priority | Notes |
|-------|-------|--------|----------|-------|
| #1 | Seed Issue 1: Backlog P0 | Backlog | P0 | Basic filtering test |
| #2 | Seed Issue 2: In Progress P1 | In progress | P1 | Basic filtering test |
| #3 | Seed Issue 3: Done P2 | Done | P2 | Basic filtering test |
| #4 | Seed Issue 4: Parent with Sub-issues | In progress | P1 | Has sub-issue #5 |
| #5 | Seed Issue 5: Sub-issue of #4 | Backlog | P1 | Child of #4 |
| #6 | Seed Issue 6: With Checklist for Split | Backlog | P2 | Has task checklist |

**Important:** Do not modify seed issues #1-6. They are used for read-only tests.

---

## Environment Variables

Set these environment variables before running integration tests:

```bash
# Required
export TEST_PROJECT_OWNER="rubrical-studios"
export TEST_PROJECT_NUMBER="29"
export TEST_REPO_OWNER="rubrical-studios"
export TEST_REPO_NAME="gh-pmu-test"

# Optional (uses gh auth token if not set)
export TEST_GH_TOKEN="ghp_your_token_here"
```

### Environment Variable Reference

| Variable | Required | Description |
|----------|----------|-------------|
| `TEST_PROJECT_OWNER` | Yes | GitHub username or org owning the test project |
| `TEST_PROJECT_NUMBER` | Yes | Project number (29 for gh-pmu-test-prj) |
| `TEST_REPO_OWNER` | Yes | Repository owner |
| `TEST_REPO_NAME` | Yes | Repository name |
| `TEST_GH_TOKEN` | No | GitHub token (defaults to `gh auth token`) |

---

## Running Tests Locally

### Prerequisites

1. Go 1.21 or later
2. `gh` CLI authenticated with access to test fixtures
3. Environment variables configured

### Run All Integration Tests

```bash
go test -v -tags=integration ./...
```

### Run Specific Test

```bash
go test -v -tags=integration ./cmd/... -run "TestRunList_Integration"
```

### Run UAT Tests

```bash
go test -v -tags=uat ./test/uat/...
```

### Skip Integration Tests

Integration tests are excluded by default (no build tag). Standard unit tests run with:

```bash
go test ./...
```

---

## Running Tests in CI

### Manual Trigger

```bash
# Run all tests
gh workflow run integration-tests.yml -f test_type=all

# Run only integration tests
gh workflow run integration-tests.yml -f test_type=integration

# Run only UAT tests
gh workflow run integration-tests.yml -f test_type=uat
```

### CI Environment

The GitHub Actions workflow uses repository secrets:

| Secret | Description |
|--------|-------------|
| `TEST_GH_TOKEN` | GitHub PAT with repo and project access |

---

## Writing Integration Tests

### Build Tag

All integration tests must include the build tag:

```go
//go:build integration

package cmd

import "testing"

func TestRunList_Integration(t *testing.T) {
    // Test implementation
}
```

### Test Utilities

Use the `internal/testutil` package (see IT-1.2) for common operations:

```go
func TestExample_Integration(t *testing.T) {
    testutil.RequireTestEnv(t)
    client := testutil.SetupTestClient(t)

    // Create test issue (cleaned up automatically)
    issueNum, cleanup := testutil.CreateTestIssue(t, "Test Issue")
    defer cleanup()

    // Run command
    output := testutil.RunCommand(t, "list", "--status", "backlog")

    // Assertions
}
```

### Test Guidelines

1. **Isolation:** Each test should be independent and not rely on other tests
2. **Cleanup:** Always clean up created resources (use `defer cleanup()`)
3. **Seed Data:** Use seed issues #1-6 for read-only tests only
4. **Created Resources:** Create new issues for write tests, then delete them
5. **Timeouts:** Allow for API latency in assertions

---

## Troubleshooting

### Test Skipped

If tests are skipped with "TEST_PROJECT_OWNER not set":
- Ensure all required environment variables are set
- Check variable names match exactly (case-sensitive)

### Authentication Errors

If you see 401/403 errors:
- Run `gh auth status` to verify authentication
- Ensure token has `repo` and `project` scopes
- For CI, verify the secret is configured correctly

### Rate Limiting

If you see rate limit errors:
- Wait for the rate limit window to reset
- Use a token with higher rate limits
- Reduce test parallelism

### Stale Seed Data

If seed issues have unexpected states:
- Check the test project board manually
- Reset seed issues to documented states
- Do not run write tests against seed issues

---

## Recreating Test Fixtures

If test fixtures need to be recreated:

### 1. Create Repository

```bash
gh repo create rubrical-studios/gh-pmu-test --private --description "gh-pmu-test for automated testing"
```

### 2. Create Project

Create project manually at https://github.com/users/rubrical-studios/projects with:
- Status: Backlog, Ready, In progress, In review, Done
- Priority: P0, P1, P2
- Size: XS, S, M, L, XL
- Estimate: Number

### 3. Create Seed Issues

See the seed issues table above for required issues and their states.

---

## Reference

- Epic: #88 - Integration Testing
- Backlog: `backlog/integration-testing-backlog.md`
- Proposal: `Proposal/PROPOSAL-Automated-Testing.md`
