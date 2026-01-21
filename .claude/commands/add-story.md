---
version: "v0.29.1"
description: Add story to epic with charter compliance (project)
argument-hint: "[epic-number]"
---
<!-- MANAGED -->
# /add-story
Add a new story to an epic with charter compliance validation and automatic test plan updates.
## Arguments
| Argument | Description |
|----------|-------------|
| `[epic-number]` | Parent epic issue number (e.g., `42` or `#42`). Optional - prompts if not specified. |
## Execution Instructions
**REQUIRED:** Before executing:
1. **Create Todo List:** Use `TodoWrite` to create todos from steps below
2. **Track Progress:** Mark todos `in_progress` → `completed`
3. **Resume Point:** If interrupted, todos show where to continue
## Phase 1: Validate Epic and Gather Story Details
**Step 1:** Parse epic number (strip leading # if present)
**Step 2:** If no epic specified:
```bash
gh issue list --label "epic" --state open --json number,title
```
**ASK USER:** Which epic should this story be added to?
**Step 3:** Validate epic has `epic` label
**Step 4:** **ASK USER:** Describe the new story:
- What should the user be able to do?
- What is the benefit/value?
- Key acceptance criteria?
**Step 5:** Transform to story format:
| User Input | Story Field |
|------------|-------------|
| User action | **I want** clause |
| Benefit | **So that** clause |
| Criteria | Checkbox list |
## Phase 2: Charter Compliance Check
**Step 1:** Load charter context:
| File | Required | Purpose |
|------|----------|---------|
| `CHARTER.md` | Recommended | Vision, goals, scope |
| `Inception/Scope-Boundaries.md` | Optional | In/out boundaries |
| `Inception/Constraints.md` | Optional | Constraints |
**If no charter:** `⚠️ No CHARTER.md found. Skipping compliance check.`
**Step 2:** Validate story against charter (vision, goals, scope, constraints)
**Step 3:** Report compliance - **ASK USER** if concerns found
## Phase 3: Create Story Issue
**Step 1:** **ASK USER:** Priority? (P0=must-have, P1=important, P2=nice-to-have)
**Step 2:** Create issue:
```bash
gh pmu create --repo {repository} \
  --title "Story: {Story Title}" \
  --label "story" \
  --body "{story_body}" \
  --status backlog \
  --priority {priority} \
  --assignee @me
```
### Story Body Template
```markdown
## Story: {Title}
### Description
As a {user type}, I want {capability} so that {benefit}.
### Acceptance Criteria
- [ ] {Criterion 1}
- [ ] {Criterion 2}
### TDD Test Cases
**Note:** Test cases added when story work begins.
### Definition of Done
- [ ] All acceptance criteria met
- [ ] TDD test cases pass
- [ ] Code reviewed
- [ ] No regressions
**Priority:** {P0|P1|P2}
**Parent Epic:** #{epic_num}
```
**Step 3:** Link to parent: `gh pmu sub add {epic_num} {story_num}`
## Phase 4: Update Test Plan
**Step 1:** Find test plan via epic's PRD reference:
```bash
gh issue view $epic_num --json body --jq '.body' | grep -oE "PRD/[A-Za-z0-9_-]+/PRD-[A-Za-z0-9_-]+\.md"
```
Derive: `PRD/{name}/PRD-{name}.md → PRD/{name}/Test-Plan-{name}.md`
**If no test plan:** Skip to Phase 5 with note
**Step 2:** Generate test cases from acceptance criteria
**Step 3:** Update test plan with new story section
**Step 4:** Commit test plan changes
## Phase 5: Report Completion
```
Story created: #{story_num}
Story: {Title}
Epic: #{epic_num} - {Epic Title}
Priority: {P0|P1|P2}
Charter compliance: ✅ Aligned (or ⚠️ Proceeded with warning)
Test plan: {Updated|Not applicable}
Next steps:
1. Work the story: work #{story_num}
2. View epic progress: gh pmu sub list #{epic_num}
```
## Error Handling
| Situation | Response |
|-----------|----------|
| Epic not found | "Issue #N not found" |
| Issue not an epic | "Issue #N does not have 'epic' label" |
| Charter concern, user declines | "Story creation cancelled" |
**End of /add-story Command**
