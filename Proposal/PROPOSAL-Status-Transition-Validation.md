# Proposal: Status Transition Validation

**Date:** 2025-12-20
**Status:** Draft
**Source:** process-docs/Proposal/Release-and-Sprint-Workflow.md (R8)

---

## Executive Summary

Add validation rules to `gh pmu move --status` and `gh pmu create` to enforce IDPF workflow constraints:

1. **Release-gated progression:** Issues cannot move from `backlog` to working statuses (`ready`, `in_progress`) without a Release assignment. Release names must exist in `releases.active[]`.

2. **Body required:** Issues cannot move to `in_review` or `done` without a body (acceptance criteria required).

3. **Checkboxes required for in_review:** Issues cannot move to `in_review` if they have unchecked checkboxes `[ ]` in the issue body.

4. **Checkboxes required for done:** Issues cannot move to `done` if they have unchecked checkboxes `[ ]` in the issue body.

5. **Framework-aware enforcement:** Validation applies by default (missing `framework` field = IDPF). Set `framework: none` to bypass all constraints.

These validations enforce workflow discipline at the gh-pmu CLI level.

**Note:** This is CLI-only enforcement. Direct GitHub API calls or web UI bypass these rules.

**Breaking Change:** Validation enforced immediately on upgrade. No grace period. Use `--force` to bypass checkbox validation only (body and release always required).

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
| `backlog` → `ready` | Release field must be set AND exist in `releases.active[]` |
| `backlog` → `in_progress` | Release field must be set AND exist in `releases.active[]` |

**Error messages:**
```
Error: Issue #45 has no release assignment.
Cannot move from 'backlog' to 'ready' without a release.

Use: gh pmu move 45 --release "release/vX.Y.Z"
 Or: /assign-release #45 release/vX.Y.Z
```

```
Error: Release "hotfix/emergency" not found in active releases.
Available releases: release/v2.0.0, patch/v1.9.1

Use 'gh pmu release start' to create a new release.
```

**Rationale:** All active work should belong to a known release for tracking and branch management.

### Rule 2: Body Required for Review/Done

| Transition | Validation |
|------------|------------|
| Any → `in_review` | Issue body must not be empty |
| Any → `done` | Issue body must not be empty |

**Error message:**
```
Error: Issue #45 has no body.
Issues require acceptance criteria before moving to in_review.
```

**Rationale:** Issues should have acceptance criteria documented in the body. Empty body indicates incomplete issue definition.

### Rule 3: Checkboxes Required for In Review

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

**Note:** All checkboxes are validated equally, including nested/indented ones.

### Rule 4: Checkboxes Required for Done

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

### Combined Validation Errors

When multiple rules fail, all errors are shown at once:

```
Error: Issue #45 has validation failures:
  - No release assignment
  - Unchecked items:
      [ ] Write unit tests
      [ ] Update documentation

Fix all issues or use --force to bypass checkbox validation.
```

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

Add `--force` flag to bypass **checkbox validation only**:

```bash
# Normal - validation applies
gh pmu move 45 --status done
# Error: Issue #45 has unchecked items...

# With force - bypass checkbox validation
gh pmu move 45 --status done --force
# Warning: Bypassing checkbox validation. Issue marked done with unchecked items.
```

**What `--force` bypasses:**

| Validation | Bypassed by `--force`? |
|------------|------------------------|
| Unchecked checkboxes (Rules 3, 4) | ✅ Yes |
| Empty body (Rule 2) | ❌ No - always required |
| Missing release (Rule 1) | ❌ No - always required |
| Release not in `active[]` (Rule 1) | ❌ No - always required |

**Use cases for `--force`:**
- Closing issues that were abandoned (unchecked items no longer relevant)
- Emergency workflow overrides
- Migration/cleanup scripts

**Note:** Body and release requirements cannot be bypassed. These are fundamental workflow constraints.

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

### Default Behavior

**Missing `framework` field defaults to IDPF.** This is a breaking change for existing users:

```go
func (cfg *Config) IsIDPF() bool {
    // Missing or empty framework = IDPF (default)
    if cfg.Framework == "" {
        return true
    }
    return cfg.Framework == "IDPF" || cfg.Framework == "idpf"
}
```

Users who want to opt out must explicitly set `framework: none`.

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

**New project (no existing config):**
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

**Existing config (re-init):**

When `.gh-pmu.yml` already exists, init uses merge behavior:

1. **Preserve** existing `framework` value (no re-prompt if already set)
2. **Preserve** existing `releases.active[]` entries
3. **Re-scan** for active releases from GitHub
4. **Merge** newly discovered releases into `releases.active[]` (additive only, no duplicates)

```
$ gh pmu init

Configuration found: .gh-pmu.yml
  Framework: IDPF (preserved)

Scanning for active releases...
  Found 1 new release:
    - patch/v2.0.1 (5 issues)
  Existing releases preserved:
    - release/v2.0.0 (12 issues)

✓ Configuration updated
✓ 1 new release added to config
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

**Note:** Release history is NOT stored locally. Query GitHub for closed releases:
```bash
gh issue list --label release --state closed
```

**Lifecycle:**
- `release start --branch X` → adds to `active[]`
- `release close <name>` → removes from `active[]`, closes tracker issue, auto-moves incomplete issues
- `release close <name> --tag` → same as above + creates git tag
- `release reopen <name>` → reopens tracker issue, adds back to `active[]`

**Note:** `release close` always requires explicit release name. No implicit "current" release.

**Close Verification:**

`release close` requires confirmation before executing:

```
$ gh pmu release close release/v2.0.0

Closing release: release/v2.0.0
  Tracker issue: #350
  Issues in release: 12 (8 done, 4 in_progress)
  Tag: (none - use --tag to create)

⚠️  4 issues are not done. They will be moved to backlog.
Proceed? (y/n): y

Moving incomplete issues to backlog...
  #45 - Feature X (was: in_progress)
  #46 - Feature Y (was: in_progress)
  #47 - Bug fix (was: ready)
  #48 - Enhancement (was: ready)

✓ Release closed
✓ 4 issues moved to backlog (Release and Microsprint cleared)
```

**Incomplete Issue Handling:**

When closing a release with incomplete issues:
1. Issues not in `done` status are auto-moved to `backlog`
2. Both `Release` and `Microsprint` fields are cleared
3. Issues get a clean slate for future release assignment

Use `--yes` to skip confirmation (for scripts):
```bash
gh pmu release close release/v2.0.0 --yes
gh pmu release close release/v2.0.0 --tag --yes
```

**Abandoned releases:** Use `release close` without `--tag` to close a release that won't ship (cancelled, abandoned, superseded).

**Reopening releases:**
```bash
# Reopen an accidentally closed or resumed release
gh pmu release reopen release/v2.0.0

# This will:
# 1. Reopen the tracker issue (#350)
# 2. Add release back to releases.active[]
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

## Microsprint Release Context

### Requirement

`microsprint start` requires explicit release context:

```bash
# Must specify release
gh pmu microsprint start --release release/v2.0.0

# Optional name
gh pmu microsprint start --release release/v2.0.0 --name "API refactor"
```

Error if no release specified:
```
Error: Microsprint requires a release context.
Use: gh pmu microsprint start --release <release-name>
```

### Release Context Storage

The microsprint's release is stored in both title and body:

**Title format:**
```
Microsprint: 2025-12-20-api-refactor [release/v2.0.0]
```

**Body metadata:**
```markdown
## Metadata
- Release: release/v2.0.0
- Started: 2025-12-20T10:00:00Z
- Theme: API refactor

## Issues
...
```

### Auto-Update on Add

When adding an issue to a microsprint, the issue's release is auto-updated to match:

```bash
gh pmu microsprint add 45
# Issue #45's release changed from "patch/v1.9.0" to "release/v2.0.0"
```

This ensures all issues in a microsprint share the same release.

---

## Validation on Create

Validation rules also apply to `gh pmu create`:

```bash
# Creating directly in in_progress requires release
gh pmu create --title "Urgent fix" --status in_progress --release release/v2.0.0

# Error if no release when creating in working status
gh pmu create --title "Urgent fix" --status in_progress
# Error: Cannot create issue in 'in_progress' without a release.
```

**Rules for create:**

| Status | Requirements |
|--------|--------------|
| `backlog` | None (default) |
| `ready` | Must specify `--release` |
| `in_progress` | Must specify `--release` |
| `in_review` | Must specify `--release`, must have body |
| `done` | Must specify `--release`, must have body |

**Note:** Can create directly in any status if requirements are met. No requirement to start in `backlog`.

---

## Acceptance Criteria

### Framework Configuration
- [ ] `gh pmu init` prompts "Are you using the IDPF framework? (y/n)" (new projects only)
- [ ] `framework: IDPF` or `framework: none` stored in `.gh-pmu.yml`
- [ ] **Missing `framework` field defaults to IDPF** (breaking change)
- [ ] When `framework: none`, skip creating Release field
- [ ] When `framework: none`, skip creating Microsprint field
- [ ] All validation rules bypassed when `framework: none`

### Init Merge Behavior
- [ ] Re-running `gh pmu init` on existing config preserves `framework` value
- [ ] Re-running `gh pmu init` preserves existing `releases.active[]` entries
- [ ] Re-running `gh pmu init` re-scans for active releases
- [ ] Newly discovered releases merged into `releases.active[]` (additive, no duplicates)

### Active Release Discovery (IDPF only)
- [ ] `gh pmu init` scans for open issues with "release" label
- [ ] Parses release name from issue title ("Release: release/v2.0.0")
- [ ] Counts issues assigned to each release
- [ ] Displays found releases during init
- [ ] Stores active releases in `releases.active[]` in `.gh-pmu.yml`
- [ ] Each release entry includes name and tracker issue number

### Release Lifecycle Config Sync
- [ ] `gh pmu release start` adds release to `releases.active[]`
- [ ] `gh pmu release close <name>` requires explicit release name (no implicit "current")
- [ ] `gh pmu release close` removes release from `releases.active[]`
- [ ] `gh pmu release close` shows confirmation with release summary (issues count, done/not done)
- [ ] `gh pmu release close` warns if issues are not done
- [ ] `gh pmu release close` auto-moves incomplete issues to backlog
- [ ] `gh pmu release close` clears both Release AND Microsprint fields on incomplete issues
- [ ] `gh pmu release close --yes` skips confirmation
- [ ] `gh pmu release close` (without `--tag`) closes tracker issue for abandoned releases
- [ ] `gh pmu release reopen <name>` reopens tracker issue and adds to `releases.active[]`
- [ ] Config file updated atomically with release operations

### Status Transition Validation (IDPF only)
- [ ] `gh pmu move --status` validates transitions before executing
- [ ] Backlog → Ready blocked if Release field empty
- [ ] Backlog → In Progress blocked if Release field empty
- [ ] **Release name must exist in `releases.active[]`** (typo prevention)
- [ ] **Any → In Review blocked if body is empty** (Rule 2)
- [ ] **Any → Done blocked if body is empty** (Rule 2)
- [ ] Any → In Review blocked if unchecked `[ ]` exists in body (Rule 3)
- [ ] Any → Done blocked if unchecked `[ ]` exists in body (Rule 4)
- [ ] **All checkboxes validated equally (including nested/indented)**
- [ ] **Multiple validation errors shown at once** (combined output)
- [ ] Error messages are actionable (include fix command)
- [ ] `--force` flag bypasses **checkbox validation only**
- [ ] `--force` does NOT bypass body or release requirements
- [ ] Uses existing data sources (GetIssue for body, GetProjectItems for fields)

### Recursive Validation
- [ ] `--recursive` validates parent AND all sub-issues
- [ ] **All issues (including sub-issues) must have body**
- [ ] Entire operation blocked if ANY issue fails validation
- [ ] Error message lists all failing issues

### Microsprint Constraints (IDPF only)
- [ ] `microsprint start` requires `--release` flag
- [ ] Microsprint release stored in tracker title: `[release/v2.0.0]`
- [ ] Microsprint release stored in tracker body metadata
- [ ] `microsprint add` auto-updates issue's release to match microsprint's release

### Release Constraints (IDPF only)
- [ ] `gh pmu release start` requires `--branch` flag
- [ ] When `framework: none`, `--branch` is optional

### Create Validation (IDPF only)
- [ ] `gh pmu create --status ready` requires `--release`
- [ ] `gh pmu create --status in_progress` requires `--release`
- [ ] `gh pmu create --status in_review` requires `--release` and body
- [ ] `gh pmu create --status done` requires `--release` and body
- [ ] Can create directly in any status if requirements met (no backlog-first requirement)

### List Filter
- [ ] `--no-release` flag added to `gh pmu list`
- [ ] Filters issues where Release field is empty/null
- [ ] Works with `--json` output
- [ ] Works in combination with other filters (`--status`, `--priority`)

### Pre-flight Validation
- [ ] `gh pmu move --dry-run` shows validation results without making changes

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
3. Move issues to `in_review`/`done` without a body
4. Have existing `.gh-pmu.yml` without `framework` field (now defaults to IDPF)

**Migration:**
- Users must assign releases before moving to working statuses
- Users must add body content before moving to `in_review`/`done`
- Users must complete or remove checkboxes before marking done
- `--force` flag available for checkbox overrides only (body and release always required)
- Set `framework: none` to opt out of all validation

**No grace period.** Validation enforced immediately on upgrade.

---

## Dependencies

| Dependency | Status |
|------------|--------|
| `gh pmu move --release` flag | Implemented (PROPOSAL-Unified-Release-Branch) |
| Issue body access in move command | Existing (uses GitHub API) |
| Project item field values | Existing (GetProjectItems) |

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
