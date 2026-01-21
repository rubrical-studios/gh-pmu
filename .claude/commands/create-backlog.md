---
version: "v0.29.1"
description: Create GitHub epics/stories from PRD (project)
argument-hint: "<prd-issue-number>"
---
<!-- MANAGED -->
# /create-backlog
Create GitHub epics and stories from an approved PRD with embedded TDD test cases.
## Arguments
| Argument | Description |
|----------|-------------|
| `<prd-issue-number>` | PRD tracking issue number (e.g., `151` or `#151`) |
## Execution Instructions
**REQUIRED:** Before executing:
1. **Create Todo List:** Use `TodoWrite` to create todos from steps below
2. **Track Progress:** Mark todos `in_progress` → `completed`
3. **Resume Point:** If interrupted, todos show where to continue
## Prerequisites
- PRD tracking issue with `prd` label
- PRD issue body contains link to `PRD/PRD-[Name].md`
- Test plan exists: `PRD/[Name]/Test-Plan-[Name].md`
- Test plan approval issue is **closed** (approved)
## Phase 1: Fetch and Validate PRD Issue
**Step 1:** Parse issue number (accept `151` or `#151`):
```bash
issue_num="${1#\#}"
```
**Step 2:** Validate PRD label:
```bash
gh issue view $issue_num --json labels,body --jq '.labels[].name' | grep -q "prd"
```
**Step 3:** Extract PRD path from issue body using pattern: `/PRD\/[A-Za-z0-9_-]+\/PRD-[A-Za-z0-9_-]+\.md/`
## Phase 2: Test Plan Approval Gate
**BLOCKING:** Backlog creation blocked until test plan approved.
```bash
gh issue list --label "test-plan" --label "approval-required" --state open --json number,title
```
| Approval Issue State | Action |
|----------------------|--------|
| **Open** | BLOCK - Show message and exit |
| **Closed** | PROCEED |
| **Not found** | WARN - Proceed but note missing test plan |
## Phase 3: Parse PRD for Epics and Stories
Load `PRD/{name}/PRD-{name}.md` and extract:
| PRD Section | Maps To |
|-------------|---------|
| `## Epics` / `### Epic N:` | GitHub issue with `epic` label |
| User stories under epic | GitHub issues with `story` label |
| Acceptance criteria | Story body checkboxes |
| Priority (P0/P1/P2) | Priority field |
## Phase 4: Load Test Cases from Test Plan
Load `PRD/{name}/Test-Plan-{name}.md`, match test cases to stories by title/acceptance criteria.
**Tech stack detection:**
| Tech Stack | Test Syntax |
|------------|-------------|
| TypeScript/JavaScript | Jest/Vitest `test()` |
| Python | pytest `def test_*():` |
| Go | `func Test*(t *testing.T)` |
| Rust | `#[test] fn test_*()` |
| Unknown | Generic pseudocode |
## Phase 5: Create Epic Issues
```bash
gh issue create --repo {repository} \
  --title "Epic: {Epic Name}" \
  --label "epic" \
  --body "## Epic: {Epic Name}
**PRD:** PRD/{name}/PRD-{name}.md
**Test Plan:** PRD/{name}/Test-Plan-{name}.md
## Description
{Epic description from PRD}
## Stories
Stories will be linked as sub-issues."
```
## Phase 6: Create Story Issues with Test Cases
```bash
gh issue create --repo {repository} \
  --title "Story: {Story Title}" \
  --label "story" \
  --body "{story_body_with_test_cases}"
```
Link to parent: `gh pmu sub add {epic_number} {story_number}`
### Story Body Template
```markdown
## Story: {Title}
### Description
As a {user}, I want {goal} so that {benefit}.
### Acceptance Criteria
- [ ] {Criterion 1}
- [ ] {Criterion 2}
### TDD Test Cases
**Source:** [Test-Plan-{name}.md](PRD/{name}/Test-Plan-{name}.md#epic-story-section)
Write these tests BEFORE implementation:
#### Unit Tests
```{language}
test('{criterion 1} succeeds with valid input', () => {
  // Arrange: set up test data
  // Act: call function under test
  // Assert: verify expected outcome
});
```
### Definition of Done
- [ ] All TDD test cases pass
- [ ] Code coverage ≥ 80%
- [ ] Edge cases handled
**Priority:** {P0|P1|P2}
```
## Phase 7: Update PRD Status
Update PRD document status to "Backlog Created", add comment to tracking issue:
```bash
gh issue comment $issue_num --body "## Backlog Created
✅ Epics: {count}
✅ Stories: {count}
✅ Test cases embedded
**Next:** Work stories via \`work #N\`"
```
## Error Handling
| Situation | Response |
|-----------|----------|
| PRD issue not found | "Issue #N not found" |
| Issue missing prd label | "Issue #N does not have 'prd' label" |
| PRD path not in body | "Could not find PRD document link" |
| Test plan not approved | BLOCK with approval instructions |
| No epics in PRD | "PRD contains no epics" |
**End of /create-backlog Command**
