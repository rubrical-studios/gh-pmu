# Product Backlog: Issue #70 - Increase Test Coverage to 80%

**Epic:** Tech Story: Increase Test Coverage to 80%
**Current Coverage:** 18.8%
**Target Coverage:** 80%
**Prioritization:** Risk/Complexity (highest risk first)

---

## Coverage Summary by File

| File | Current | Functions at 0% | Risk Level |
|------|---------|-----------------|------------|
| `cmd/triage.go` | 0% | 11 functions | **Critical** |
| `cmd/move.go` | 1.9% | 2 core functions | **Critical** |
| `cmd/sub.go` | 0% (run functions) | 4 run functions | **Critical** |
| `internal/api/mutations.go` | 0% | 11 functions | **High** |
| `internal/api/queries.go` | 0% | 12 functions | **High** |
| `cmd/list.go` | 0% (run function) | 6 functions | **High** |
| `cmd/create.go` | 10.2% | 1 core function | **Medium** |
| `cmd/view.go` | 0% (run function) | 3 functions | **Medium** |
| `cmd/intake.go` | 0% | 4 functions | **Medium** |
| `cmd/split.go` | 0% (run function) | 3 functions | **Medium** |
| `cmd/init.go` | ~50% | 3 functions | **Low** |
| `internal/ui/ui.go` | 71.9% | Spinner methods | **Low** |

---

## Backlog (Prioritized by Risk/Complexity)

### Story 1: Test `triage.go` Core Logic

**As a** developer maintaining gh-pmu
**I want** comprehensive tests for triage command logic
**So that** I can safely refactor and extend triage functionality without regressions

**Story Points:** 8
**Priority:** P0 - Critical
**Risk:** Highest (11 untested functions, complex query matching, state mutations)

**Acceptance Criteria:**

```gherkin
Scenario: matchesTriageQuery correctly filters by label exclusion
  Given an issue with label "bug"
  When matching against query "-label:bug"
  Then the issue should NOT match

Scenario: matchesTriageQuery correctly filters by label inclusion
  Given an issue with label "enhancement"
  When matching against query "label:enhancement"
  Then the issue should match

Scenario: matchesTriageQuery correctly filters by state
  Given an issue with state "OPEN"
  When matching against query "is:open"
  Then the issue should match

Scenario: describeActions summarizes triage actions
  Given a triage config with labels ["pm-tracked"] and fields {priority: p1}
  When describeActions is called
  Then it returns "labels: pm-tracked; priority: p1"

Scenario: listTriageConfigs outputs table format
  Given a config with 2 triage rules
  When listTriageConfigs is called with json=false
  Then it outputs a formatted table with NAME, QUERY, ACTIONS columns
```

**Tests to Create:**

1. **`TestMatchesTriageQuery`** (table-driven)
   - Test cases: label inclusion, label exclusion, state filtering, combined queries
   - Mock: None needed (pure function)
   - Pattern: Input issue + query string → boolean result

2. **`TestDescribeActions`** (table-driven)
   - Test cases: labels only, fields only, combined, interactive only, empty
   - Mock: None needed (pure function on config struct)

3. **`TestDescribeTriageActions`**
   - Test: Verify output formatting for various triage configs
   - Mock: Use `bytes.Buffer` as cmd output

4. **`TestListTriageConfigs`**
   - Test: Table and JSON output formats
   - Mock: Use `bytes.Buffer`, provide test config

5. **`TestSearchIssuesForTriage`**
   - Test: Repository parsing, state detection from query
   - Mock: Create mock API client interface

6. **`TestOutputTriageTable`** and **`TestOutputTriageJSON`**
   - Test: Output formatting for various issue lists
   - Mock: Use `bytes.Buffer`

---

### Story 2: Test `move.go` Core Logic

**As a** developer maintaining gh-pmu
**I want** comprehensive tests for the move command
**So that** recursive sub-issue updates work correctly without data corruption

**Story Points:** 5
**Priority:** P0 - Critical
**Risk:** High (recursive tree traversal, state mutations on multiple issues)

**Acceptance Criteria:**

```gherkin
Scenario: runMove requires at least one field flag
  Given no --status or --priority flag provided
  When runMove is called
  Then it returns error "at least one of --status or --priority is required"

Scenario: collectSubIssuesRecursive respects depth limit
  Given an issue tree with depth 5
  When collecting with maxDepth=2
  Then only 2 levels of sub-issues are collected

Scenario: collectSubIssuesRecursive handles cross-repo sub-issues
  Given a parent in repo A with sub-issue in repo B
  When collecting sub-issues
  Then both repos are correctly identified in the result
```

**Tests to Create:**

1. **`TestRunMove_ValidationErrors`** (table-driven)
   - Test cases: no flags, invalid issue number, missing config
   - Mock: None needed for validation path

2. **`TestCollectSubIssuesRecursive`**
   - Test: depth limiting, cross-repo handling, empty sub-issues
   - Mock: Create mock client that returns predefined sub-issue trees
   - Pattern: Setup mock → call function → assert collected issues

3. **`TestMoveCommand_FlagParsing`**
   - Test: All flags parse correctly (--status, --priority, --recursive, --depth, --dry-run, --yes)
   - Pattern: Create command, set args, verify opts struct

---

### Story 3: Test `sub.go` Command Functions

**As a** developer maintaining gh-pmu
**I want** tests for sub-issue management commands
**So that** parent-child relationships are created and removed correctly

**Story Points:** 8
**Priority:** P0 - Critical
**Risk:** High (4 distinct operations: add, create, list, remove - all mutating)

**Acceptance Criteria:**

```gherkin
Scenario: outputSubListTable formats cross-repo sub-issues correctly
  Given sub-issues in different repositories
  When outputSubListTable is called
  Then repository info is shown for each sub-issue

Scenario: outputSubListJSON includes summary counts
  Given 3 sub-issues (2 open, 1 closed)
  When outputSubListJSON is called
  Then summary shows total=3, open=2, closed=1

Scenario: subCreateOptions inherit flags have correct defaults
  Given a new subCreateOptions struct
  Then inheritLabels=true, inheritAssign=false, inheritMilestone=true
```

**Tests to Create:**

1. **`TestOutputSubListTable`**
   - Test: Empty list, single repo, cross-repo display, progress calculation
   - Mock: Use `bytes.Buffer` for output capture

2. **`TestOutputSubListJSON`**
   - Test: JSON structure, summary counts, repository field formatting
   - Mock: None (pure function)

3. **`TestSubCreateOptions_Defaults`**
   - Test: Verify default values for inherit flags
   - Pattern: Create struct, assert field values

4. **`TestNewSubAddCommand`**, **`TestNewSubCreateCommand`**, **`TestNewSubListCommand`**, **`TestNewSubRemoveCommand`**
   - Test: Flag existence, required flags marked, command structure
   - Pattern: Similar to existing command tests

---

### Story 4: Test `internal/api/mutations.go`

**As a** developer maintaining gh-pmu
**I want** unit tests for all GraphQL mutation functions
**So that** API interactions are verified without hitting real GitHub API

**Story Points:** 8
**Priority:** P1 - High
**Risk:** High (all functions at 0%, critical for data integrity)

**Acceptance Criteria:**

```gherkin
Scenario: CreateIssue returns error when GraphQL client is nil
  Given a client with nil gql
  When CreateIssue is called
  Then it returns "GraphQL client not initialized" error

Scenario: SetProjectItemField routes to correct setter by field type
  Given a SINGLE_SELECT field
  When SetProjectItemField is called
  Then setSingleSelectField is invoked

Scenario: setSingleSelectField returns error for unknown option
  Given a field with options ["P0", "P1", "P2"]
  When setting value "P3"
  Then it returns error 'option "P3" not found'
```

**Tests to Create:**

1. **`TestCreateIssue_NilClient`**
   - Test: Error when gql is nil
   - Mock: Create Client with nil gql field

2. **`TestSetProjectItemField_FieldTypeRouting`** (table-driven)
   - Test cases: SINGLE_SELECT → setSingleSelectField, TEXT → setTextField, NUMBER → setNumberField, unknown type → error
   - Mock: Create mock client that returns predefined fields

3. **`TestSetSingleSelectField_OptionLookup`**
   - Test: Valid option found, option not found error
   - Mock: Provide field with known options

4. **`TestAddIssueToProject_NilClient`**, **`TestAddSubIssue_NilClient`**, **`TestRemoveSubIssue_NilClient`**
   - Test: All mutation functions properly check for nil client
   - Pattern: Quick nil-check coverage

5. **Interface Extraction for Mocking:**
   - Extract `GraphQLClient` interface from concrete type
   - Create `MockGraphQLClient` for testing Query/Mutate calls

---

### Story 5: Test `internal/api/queries.go`

**As a** developer maintaining gh-pmu
**I want** unit tests for all GraphQL query functions
**So that** data fetching is verified without hitting real GitHub API

**Story Points:** 8
**Priority:** P1 - High
**Risk:** High (12 functions at 0%, foundation for all commands)

**Acceptance Criteria:**

```gherkin
Scenario: GetProject tries user first, then organization
  Given an owner that is an organization
  When GetProject is called
  Then getUserProject fails and getOrgProject succeeds

Scenario: splitRepoName correctly parses owner/repo
  Given "rubrical-studios/gh-pmu"
  When splitRepoName is called
  Then it returns ["rubrical-studios", "gh-pmu"]

Scenario: GetProjectItems filters by repository when filter provided
  Given items from multiple repositories
  When GetProjectItems is called with repository filter
  Then only matching items are returned
```

**Tests to Create:**

1. **`TestSplitRepoName`** (table-driven)
   - Test cases: valid "owner/repo", no slash, empty string
   - Mock: None (pure function)

2. **`TestGetProject_NilClient`**
   - Test: Error when gql is nil
   - Pattern: Quick coverage for guard clause

3. **`TestGetProject_UserVsOrg`**
   - Test: Fallback from user to org when user query fails
   - Mock: Mock client that fails for user, succeeds for org

4. **`TestGetProjectFields_FieldTypeParsing`**
   - Test: SINGLE_SELECT fields include options, regular fields do not
   - Mock: Return varied field types

5. **`TestGetRepositoryIssues_StateMapping`** (table-driven)
   - Test cases: "open"→OPEN, "closed"→CLOSED, "all"→[OPEN, CLOSED], ""→[OPEN, CLOSED]
   - Pattern: Verify state parameter transformation

6. **`TestGetParentIssue_NoParent`**
   - Test: Returns nil, nil when issue has no parent
   - Mock: Return empty parent struct

---

### Story 6: Test `cmd/list.go` Filtering and Output

**As a** developer maintaining gh-pmu
**I want** tests for list command filtering logic
**So that** users can reliably filter and view project items

**Story Points:** 5
**Priority:** P1 - High
**Risk:** Medium (6 functions, filtering logic needs validation)

**Acceptance Criteria:**

```gherkin
Scenario: filterByFieldValue is case-insensitive
  Given items with Status "In Progress"
  When filtering by "in_progress" (lowercase)
  Then matching items are returned

Scenario: getFieldValue returns empty string for missing field
  Given an item without Priority field
  When getFieldValue is called for "Priority"
  Then it returns ""

Scenario: outputTable truncates long titles
  Given an issue with 60-character title
  When outputTable is called
  Then title is truncated to 50 chars with "..."
```

**Tests to Create:**

1. **`TestFilterByFieldValue`** (table-driven)
   - Test cases: exact match, case-insensitive match, no match, empty items
   - Mock: None (pure function on slice)

2. **`TestGetFieldValue`**
   - Test: Field exists, field missing, multiple fields
   - Mock: None (pure function)

3. **`TestOutputTable`**
   - Test: Empty items, title truncation, assignee formatting
   - Mock: Use `bytes.Buffer`

4. **`TestOutputJSON`**
   - Test: JSON structure matches JSONOutput type, empty items
   - Mock: Use `bytes.Buffer`, unmarshal to verify

---

### Story 7: Test `cmd/create.go` Core Logic

**As a** developer maintaining gh-pmu
**I want** tests for issue creation flow
**So that** issues are created with correct defaults and project fields

**Story Points:** 3
**Priority:** P2 - Medium
**Risk:** Medium (currently 10.2%, main flow untested)

**Acceptance Criteria:**

```gherkin
Scenario: runCreate requires --title flag
  Given no title provided
  When runCreate is called
  Then it returns error "--title is required"

Scenario: runCreate merges default labels with command line labels
  Given config defaults with ["pm-tracked"]
  And command line labels ["bug", "urgent"]
  When creating issue
  Then all 3 labels are applied
```

**Tests to Create:**

1. **`TestRunCreate_TitleRequired`**
   - Test: Error when title is empty
   - Pattern: Call with empty opts.title

2. **`TestRunCreate_LabelMerging`**
   - Test: Config defaults + CLI labels combined correctly
   - Mock: Mock API client, verify CreateIssue receives merged labels

---

### Story 8: Test `cmd/view.go` Output Functions

**As a** developer maintaining gh-pmu
**I want** tests for view command output formatting
**So that** issue details display correctly in all formats

**Story Points:** 3
**Priority:** P2 - Medium
**Risk:** Medium (output formatting, parseIssueNumber/Reference already tested)

**Acceptance Criteria:**

```gherkin
Scenario: outputViewTable displays sub-issue progress
  Given an issue with 5 sub-issues (3 closed)
  When outputViewTable is called
  Then progress bar shows 60% complete

Scenario: outputViewJSON includes all issue fields
  Given an issue with labels, assignees, milestone
  When outputViewJSON is called
  Then JSON includes all fields correctly structured
```

**Tests to Create:**

1. **`TestOutputViewTable`**
   - Test: Issue with/without sub-issues, progress bar display
   - Mock: Use `bytes.Buffer`

2. **`TestOutputViewJSON`**
   - Test: All fields present, nested structures correct
   - Mock: Use `bytes.Buffer`, unmarshal to verify

---

### Story 9: Test `cmd/intake.go` Output Functions

**As a** developer maintaining gh-pmu
**I want** tests for intake command output
**So that** untracked issue discovery displays correctly

**Story Points:** 3
**Priority:** P2 - Medium
**Risk:** Medium (output formatting primarily)

**Acceptance Criteria:**

```gherkin
Scenario: outputIntakeTable truncates long titles
  Given issues with 60-character titles
  When outputIntakeTable is called
  Then titles are truncated to 50 chars

Scenario: outputIntakeJSON includes correct status field
  Given status "dry-run"
  When outputIntakeJSON is called
  Then JSON status field is "dry-run"
```

**Tests to Create:**

1. **`TestOutputIntakeTable`**
   - Test: Title truncation, repository display
   - Mock: Use `bytes.Buffer`

2. **`TestOutputIntakeJSON`**
   - Test: Status field, count matches issues length
   - Mock: Use `bytes.Buffer`, unmarshal to verify

---

### Story 10: Test `cmd/split.go` Output Functions

**As a** developer maintaining gh-pmu
**I want** tests for split command output functions
**So that** sub-issue creation results display correctly

**Story Points:** 2
**Priority:** P2 - Medium
**Risk:** Low (parseChecklist already well-tested, only output functions remain)

**Acceptance Criteria:**

```gherkin
Scenario: outputSplitJSON includes parent issue info
  Given a parent issue and list of tasks
  When outputSplitJSON is called
  Then JSON includes parent number, title, url

Scenario: outputSplitJSONCreated tracks created vs failed
  Given 3 created issues and 1 failed
  When outputSplitJSONCreated is called
  Then createdCount=3 and failedCount=1
```

**Tests to Create:**

1. **`TestOutputSplitJSON`**
   - Test: Parent info, task list, status field
   - Mock: Use `bytes.Buffer`

2. **`TestOutputSplitJSONCreated`**
   - Test: Counts match, created issues have correct structure
   - Mock: Use `bytes.Buffer`

---

### Story 11: Test `cmd/init.go` Remaining Functions

**As a** developer maintaining gh-pmu
**I want** tests for init command helper functions
**So that** repository detection and config writing are reliable

**Story Points:** 3
**Priority:** P3 - Low
**Risk:** Low (partially tested, mostly I/O operations)

**Acceptance Criteria:**

```gherkin
Scenario: detectRepository returns empty for non-git directory
  Given no git remote configured
  When detectRepository is called
  Then it returns ""

Scenario: splitRepository handles invalid input
  Given "noslash"
  When splitRepository is called
  Then owner="" and name=""
```

**Tests to Create:**

1. **`TestDetectRepository`**
   - Test: Requires git environment (may need to skip in CI)
   - Pattern: Integration test with actual git command

2. **`TestSplitRepository`** (table-driven)
   - Test cases: valid "owner/repo", no slash, empty
   - Mock: None (pure function)

3. **`TestWriteConfig`**, **`TestWriteConfigWithMetadata`**
   - Test: File creation, YAML structure
   - Mock: Use temp directory

---

### Story 12: Test `internal/ui/ui.go` Spinner Methods

**As a** developer maintaining gh-pmu
**I want** tests for Spinner functionality
**So that** loading indicators work correctly across terminals

**Story Points:** 2
**Priority:** P3 - Low
**Risk:** Low (UI feedback, not critical path)

**Acceptance Criteria:**

```gherkin
Scenario: Spinner starts and stops without panic
  Given a new Spinner
  When Start then Stop are called
  Then no error or panic occurs

Scenario: UpdateMessage changes spinner text
  Given a running Spinner with "Loading..."
  When UpdateMessage("Processing...") is called
  Then the message updates
```

**Tests to Create:**

1. **`TestSpinner_StartStop`**
   - Test: Start/Stop lifecycle doesn't panic
   - Mock: Use `bytes.Buffer` as writer

2. **`TestSpinner_UpdateMessage`**
   - Test: Message can be changed after start
   - Pattern: Start spinner, update, verify no crash

---

## Mock Client Strategy

For Stories 4-5 (API layer), create a reusable mock:

```go
// internal/api/mock_client.go (test file)

type MockGraphQLClient struct {
    QueryFunc  func(name string, q interface{}, vars map[string]interface{}) error
    MutateFunc func(name string, m interface{}, vars map[string]interface{}) error
}

func (m *MockGraphQLClient) Query(name string, q interface{}, vars map[string]interface{}) error {
    if m.QueryFunc != nil {
        return m.QueryFunc(name, q, vars)
    }
    return nil
}

func (m *MockGraphQLClient) Mutate(name string, m interface{}, vars map[string]interface{}) error {
    if m.MutateFunc != nil {
        return m.MutateFunc(name, m, vars)
    }
    return nil
}
```

---

## Summary

| Story | File(s) | Points | Priority | Coverage Gain |
|-------|---------|--------|----------|---------------|
| 1 | cmd/triage.go | 8 | P0 | ~15% |
| 2 | cmd/move.go | 5 | P0 | ~5% |
| 3 | cmd/sub.go | 8 | P0 | ~8% |
| 4 | internal/api/mutations.go | 8 | P1 | ~10% |
| 5 | internal/api/queries.go | 8 | P1 | ~12% |
| 6 | cmd/list.go | 5 | P1 | ~5% |
| 7 | cmd/create.go | 3 | P2 | ~3% |
| 8 | cmd/view.go | 3 | P2 | ~3% |
| 9 | cmd/intake.go | 3 | P2 | ~3% |
| 10 | cmd/split.go | 2 | P2 | ~2% |
| 11 | cmd/init.go | 3 | P3 | ~3% |
| 12 | internal/ui/ui.go | 2 | P3 | ~2% |

**Total Story Points:** 58
**Estimated Coverage After Completion:** ~85%+ (exceeds 80% target)

---

*Generated for Issue #70: Tech Story: Increase Test Coverage to 80%*
