# Proposal: Status Transition Validation

**Date:** 2025-12-20
**Status:** Draft
**Source:** process-docs/Proposal/Release-and-Sprint-Workflow.md (R8)

---

## Executive Summary

Add validation rules to `gh pmu move --status` to enforce IDPF workflow constraints:

1. **Release-gated progression:** Issues cannot move from `backlog` to working statuses (`ready`, `in_progress`) without a Release assignment.

2. **Review-gated in_review:** Issues cannot move to `in_review` if they have unchecked checkboxes `[ ]` in the issue body.

3. **Completion-gated done:** Issues cannot move to `done` if they have unchecked checkboxes `[ ]` in the issue body.

4. **Framework-aware enforcement:** Validation only applies when `framework: IDPF` is set in `.gh-pmu.yml`. Users not using IDPF can bypass all constraints.

These validations enforce workflow discipline at the gh-pmu CLI level.

**Note:** This is CLI-only enforcement. Direct GitHub API calls or web UI bypass these rules.

---

## Current State

Currently, `gh pmu move --status` accepts any valid status value without validation:

```bash
gh pmu move 45 --status done   # Works even if issue has unchecked boxes
gh pmu move 45 --status ready  # Works even if no release assigned
```

This allows:
- Issues to be marked complete with unfinished acceptance criteria
- Work to start on issues not assigned to any release (untracked work)

---

## Proposed Validation Rules

### Rule 1: Release Required for Work

| Transition | Validation |
|------------|------------|
| `backlog` → `ready` | Release field must be set |
| `backlog` → `in_progress` | Release field must be set |

**Error message:**
```
Error: Issue #45 has no release assignment.
Cannot move from 'backlog' to 'ready' without a release.

Use: gh pmu move 45 --release "release/vX.Y.Z"
 Or: /assign-release #45 release/vX.Y.Z
```

**Rationale:** All active work should belong to a release for tracking and branch management.

### Rule 2: Checkboxes Required for In Review

| Transition | Validation |
|------------|------------|
| Any → `in_review` | No unchecked boxes `[ ]` in issue body |

**Error message:**
```
Error: Issue #45 has unchecked items:
  [ ] Write unit tests
  [ ] Update documentation

Complete these items before moving to review.
```

**Rationale:** Moving to review implies work is complete. Unchecked boxes indicate incomplete acceptance criteria.

### Rule 3: Checkboxes Required for Done

| Transition | Validation |
|------------|------------|
| Any → `done` | No unchecked boxes `[ ]` in issue body |

**Error message:**
```
Error: Issue #45 has unchecked items:
  [ ] Write unit tests
  [ ] Update documentation

Complete these items or remove them before marking done.
```

**Rationale:** Checkboxes typically represent acceptance criteria. Marking an issue done with unchecked boxes indicates incomplete work.

---

## Implementation

### Validation Function

```go
func validateStatusTransition(cfg *Config, issue *Issue, fromStatus, toStatus string) error {
    // Skip validation if not using IDPF
    if !cfg.IsIDPF() {
        return nil
    }

    // Rule 1: Release required for work
    if fromStatus == "backlog" && (toStatus == "ready" || toStatus == "in_progress") {
        if issue.Release == "" {
            return fmt.Errorf(
                "Issue #%d has no release assignment.\n"+
                "Cannot move from '%s' to '%s' without a release.\n\n"+
                "Use: gh pmu move %d --release \"release/vX.Y.Z\"",
                issue.Number, fromStatus, toStatus, issue.Number,
            )
        }
    }

    // Rule 2: Checkboxes required for in_review
    if toStatus == "in_review" {
        unchecked := findUncheckedBoxes(issue.Body)
        if len(unchecked) > 0 {
            return fmt.Errorf(
                "Issue #%d has unchecked items:\n%s\n\n"+
                "Complete these items before moving to review.",
                issue.Number, formatUncheckedList(unchecked),
            )
        }
    }

    // Rule 3: Checkboxes required for done
    if toStatus == "done" {
        unchecked := findUncheckedBoxes(issue.Body)
        if len(unchecked) > 0 {
            return fmt.Errorf(
                "Issue #%d has unchecked items:\n%s\n\n"+
                "Complete these items or remove them before marking done.",
                issue.Number, formatUncheckedList(unchecked),
            )
        }
    }

    return nil
}

func findUncheckedBoxes(body string) []string {
    // Simple regex - accepts rare false positives in code blocks
    re := regexp.MustCompile(`\[ \] (.+)`)
    matches := re.FindAllStringSubmatch(body, -1)

    var items []string
    for _, m := range matches {
        items = append(items, m[1])
    }
    return items
}
```

### Integration Point

Add validation call in `runMove()` before updating status:

```go
func runMove(opts moveOptions) error {
    cfg, err := config.LoadFromDirectory(mustGetwd())
    if err != nil {
        return err
    }

    // Batch-fetch all issues upfront for efficiency
    issues, err := batchFetchIssues(opts.issueNumbers)
    if err != nil {
        return err
    }

    // Validate ALL issues before making any changes (all-or-nothing)
    if opts.status != "" && cfg.IsIDPF() {
        for _, issue := range issues {
            if err := validateStatusTransition(cfg, issue, issue.Status, opts.status); err != nil {
                return err  // Block entire operation if any fails
            }
        }
    }

    // All validations passed - proceed with move...
}
```

### Batch-Fetch Optimization

For bulk moves (`gh pmu move 1 2 3 4 5 --status ready`), fetch all issue data in a single GraphQL query:

```go
func batchFetchIssues(numbers []int) ([]*Issue, error) {
    // Single GraphQL query fetches: number, status, release, body
    // Avoids N+1 API calls for validation
    query := buildBatchIssueQuery(numbers)
    return executeQuery(query)
}
```

### Recursive Validation

When `--recursive` is used, validation applies to parent AND all sub-issues:

```bash
gh pmu move 100 --status done --recursive
# Epic #100 has 5 sub-issues
```

**Behavior:** If ANY issue (parent or sub-issue) fails validation, the entire operation is blocked.

```
Error: Validation failed for recursive move:
  - Issue #102 has unchecked items:
      [ ] Write unit tests
  - Issue #104 has unchecked items:
      [ ] Update documentation

Fix all issues or use --force to bypass.
```

---

## Override Flag

Add `--force` flag to bypass validation when necessary:

```bash
# Normal - validation applies
gh pmu move 45 --status done
# Error: Issue #45 has unchecked items...

# With force - bypass validation
gh pmu move 45 --status done --force
# Warning: Bypassing validation. Issue marked done with unchecked items.
```

**Use cases for `--force`:**
- Closing issues that were abandoned (unchecked items no longer relevant)
- Emergency workflow overrides
- Migration/cleanup scripts

---

## Framework-Aware Validation

### Configuration

`gh pmu init` asks whether the user is using the IDPF framework:

```
? Are you using the IDPF framework? (y/n): y
```

This is stored in `.gh-pmu.yml`:

```yaml
project:
  owner: rubrical-studios
  number: 11
framework: IDPF  # or "none"
```

### Bypass Rules

When `framework: none`, the following constraints are bypassed:

| Category | Constraint | IDPF Mode | None Mode |
|----------|------------|-----------|-----------|
| **Status Validation** | Release required for `backlog → ready/in_progress` | Enforced | Bypassed |
| | Checkboxes required for `→ in_review` | Enforced | Bypassed |
| | Checkboxes required for `→ done` | Enforced | Bypassed |
| **Sprint Constraints** | Sprint must belong to a release | Enforced | Bypassed |
| | All issues in sprint must share same release | Enforced | Bypassed |
| **Release Constraints** | `gh pmu release start` requires `--branch` | Required | Optional |
| **Field Creation** | Create `Release` field on init | Auto-created | Skipped |
| | Create `Microsprint` field on init | Auto-created | Skipped |

### Implementation

```go
func (cfg *Config) IsIDPF() bool {
    return cfg.Framework == "IDPF" || cfg.Framework == "idpf"
}

func validateStatusTransition(cfg *Config, issue *Issue, fromStatus, toStatus string) error {
    // Skip validation if not using IDPF
    if !cfg.IsIDPF() {
        return nil
    }

    // Apply IDPF validation rules...
}
```

### Init Flow

```
$ gh pmu init

? Project owner: rubrical-studios
? Project number: 11
? Repository: rubrical-studios/gh-pmu
? Are you using the IDPF framework? (y/n): y

Scanning for active releases...
  Found 2 active releases:
    - release/v2.0.0 (12 issues)
    - patch/v1.9.1 (3 issues)

✓ Configuration saved to .gh-pmu.yml
✓ Release and Microsprint fields created
✓ Active releases synced to config
```

**Non-IDPF flow:**
```
$ gh pmu init

? Project owner: rubrical-studios
? Project number: 11
? Repository: rubrical-studios/gh-pmu
? Are you using the IDPF framework? (y/n): n

✓ Configuration saved to .gh-pmu.yml
  Note: IDPF workflow constraints disabled.
  Release and Microsprint fields not created.
```

### Active Release Discovery

When `framework: IDPF`, init scans for active releases:

```go
func discoverActiveReleases(client *api.Client, owner, repo string) ([]Release, error) {
    // Find open issues with "release" label
    issues, err := client.GetOpenIssuesByLabel(owner, repo, "release")
    if err != nil {
        return nil, err
    }

    var releases []Release
    for _, issue := range issues {
        // Parse release name from title: "Release: release/v2.0.0"
        if name := parseReleaseName(issue.Title); name != "" {
            releases = append(releases, Release{
                Name:       name,
                TrackerID:  issue.ID,
                IssueCount: countIssuesInRelease(name),
            })
        }
    }
    return releases, nil
}
```

**Stored in `.gh-pmu.yml`:**

```yaml
project:
  owner: rubrical-studios
  number: 11
framework: IDPF
releases:
  active:
    - name: release/v2.0.0
      tracker: 350
    - name: patch/v1.9.1
      tracker: 355
```

---

## New Flag: `--no-release` Filter

Add `--no-release` flag to `gh pmu list` for querying backlog issues without release assignment:

```bash
# Get all backlog issues without a release (true backlog)
gh pmu list --status backlog --no-release

# JSON output for scripting
gh pmu list --status backlog --no-release --json number,title,labels
```

**Implementation:**

```go
type listOptions struct {
    // Existing fields
    status   string
    priority string
    // ...

    // New field
    noRelease bool
}

// In filter logic
if opts.noRelease {
    query += ` AND release IS NULL`
}
```

---

## Acceptance Criteria

### Framework Configuration
- [ ] `gh pmu init` prompts "Are you using the IDPF framework? (y/n)"
- [ ] `framework: IDPF` or `framework: none` stored in `.gh-pmu.yml`
- [ ] When `framework: none`, skip creating Release field
- [ ] When `framework: none`, skip creating Microsprint field
- [ ] All validation rules bypassed when `framework: none`

### Active Release Discovery (IDPF only)
- [ ] `gh pmu init` scans for open issues with "release" label
- [ ] Parses release name from issue title ("Release: release/v2.0.0")
- [ ] Counts issues assigned to each release
- [ ] Displays found releases during init
- [ ] Stores active releases in `releases.active[]` in `.gh-pmu.yml`
- [ ] Each release entry includes name and tracker issue number

### Status Transition Validation (IDPF only)
- [ ] `gh pmu move --status` validates transitions before executing
- [ ] Backlog → Ready blocked if Release field empty
- [ ] Backlog → In Progress blocked if Release field empty
- [ ] Any → In Review blocked if unchecked `[ ]` exists in body
- [ ] Any → Done blocked if unchecked `[ ]` exists in body
- [ ] Error messages are actionable (include fix command)
- [ ] `--force` flag bypasses validation with warning
- [ ] Batch-fetch optimization for bulk moves (single GraphQL query)

### Recursive Validation
- [ ] `--recursive` validates parent AND all sub-issues
- [ ] Entire operation blocked if ANY issue fails validation
- [ ] Error message lists all failing issues

### Sprint Constraints (IDPF only)
- [ ] Sprint must belong to a release (error if no release context)
- [ ] All issues in sprint must share same release (warn on mismatch)

### Release Constraints (IDPF only)
- [ ] `gh pmu release start` requires `--branch` flag
- [ ] When `framework: none`, `--branch` is optional

### List Filter
- [ ] `--no-release` flag added to `gh pmu list`
- [ ] Filters issues where Release field is empty/null
- [ ] Works with `--json` output
- [ ] Works in combination with other filters (`--status`, `--priority`)

### CLI Enforcement Note
- [ ] Documentation notes this is CLI-only enforcement
- [ ] GitHub web UI and direct API bypass these rules

---

## Testing

### Unit Tests

```go
func TestValidateStatusTransition_ReleaseRequired(t *testing.T) {
    issue := &Issue{Number: 45, Status: "backlog", Release: ""}

    err := validateStatusTransition(issue, "backlog", "ready")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "no release assignment")
}

func TestValidateStatusTransition_UncheckedBoxes(t *testing.T) {
    issue := &Issue{
        Number: 45,
        Status: "in_progress",
        Body:   "## Tasks\n- [x] Done\n- [ ] Not done",
    }

    err := validateStatusTransition(issue, "in_progress", "done")
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "unchecked items")
}

func TestValidateStatusTransition_AllChecked(t *testing.T) {
    issue := &Issue{
        Number: 45,
        Status: "in_progress",
        Body:   "## Tasks\n- [x] Done\n- [x] Also done",
    }

    err := validateStatusTransition(issue, "in_progress", "done")
    assert.NoError(t, err)
}
```

### Integration Tests

```bash
# Test: Release required
gh pmu create --title "Test" --status backlog
gh pmu move 999 --status ready
# Expected: Error about missing release

# Test: Unchecked boxes
gh issue create --title "Test" --body "- [ ] Task"
gh pmu move 999 --status done
# Expected: Error about unchecked items

# Test: Force override
gh pmu move 999 --status done --force
# Expected: Warning, but succeeds
```

---

## Backward Compatibility

This is a **breaking change** for workflows that:
1. Move issues to `ready`/`in_progress` without release assignment
2. Move issues to `done` with unchecked checkboxes

**Migration:**
- Users must assign releases before moving to working statuses
- Users must complete or remove checkboxes before marking done
- `--force` flag available for emergency overrides

---

## Dependencies

| Dependency | Status |
|------------|--------|
| `gh pmu move --release` flag | Implemented (PROPOSAL-Unified-Release-Branch) |
| Issue body access in move command | Existing (uses GitHub API) |

---

## Estimated Effort

| Component | Lines | Complexity |
|-----------|-------|------------|
| Validation function | ~50 | Low |
| Integration in move | ~10 | Low |
| `--force` flag | ~15 | Low |
| `--no-release` filter | ~20 | Low |
| Tests | ~100 | Medium |
| **Total** | ~195 | Low-Medium |

---

## Decision

- [ ] Approved - Proceed to implementation
- [ ] Approved with modifications
- [ ] Rejected
- [ ] Needs more information

---

## References

- Source: `process-docs/Proposal/Release-and-Sprint-Workflow.md` (R8)
- Related: `PROPOSAL-Unified-Release-Branch.md` (Implemented)
- GitHub Workflow: `.claude/rules/02-github-workflow.md`
