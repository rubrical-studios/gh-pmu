---
version: "v0.29.3"
description: Split story into smaller stories (project)
argument-hint: "<story-number>"
---
<!-- MANAGED -->
# /split-story
Split a story into smaller, more manageable stories while maintaining charter compliance and test plan integrity.
## Arguments
| Argument | Description |
|----------|-------------|
| `<story-number>` | Story issue number to split (e.g., `123` or `#123`) |
## Execution Instructions
**REQUIRED:** Before executing:
1. **Create Todo List:** Use `TodoWrite` to create todos from steps below
2. **Track Progress:** Mark todos `in_progress` → `completed`
3. **Resume Point:** If interrupted, todos show where to continue
## Phase 1: Fetch and Validate Original Story
**Step 1:** Parse story number (strip leading #)
**Step 2:** Validate story label: `gh issue view $story_num --json labels,body,title --jq '.labels[].name' | grep -q "story"`
**Step 3:** Extract story details: `gh pmu view $story_num --body-stdout > .tmp-story.md`
Parse: Title, Description, Acceptance criteria, Priority, Parent epic
**Step 4:** Find parent epic: `gh pmu sub list --child $story_num --json parent`
## Phase 2: Determine Split Criteria
**ASK USER:** How should this story be split?
| Pattern | Description |
|---------|-------------|
| By acceptance criteria | Each criterion becomes a story |
| By user workflow | Split by distinct user actions |
| By technical component | Frontend/backend/API |
| By priority | Must-have vs nice-to-have |
| Custom | User defines split |
Gather: Title, acceptance criteria, priority for each new story. **Minimum:** 2 stories required.
## Phase 3: Charter Compliance Check
**Step 1:** Load charter context (CHARTER.md, Inception/Scope-Boundaries.md, Inception/Constraints.md)
**If no charter:** `⚠️ No CHARTER.md found. Skipping compliance check.`
**Step 2:** Validate split stories for scope creep, constraint violations, goal alignment
**Step 3:** Report compliance - **ASK USER** if concerns found
## Phase 4: Create New Stories
For each new story:
**Step 1:** Create issue:
```bash
gh pmu create --repo {repository} \
  --title "Story: {New Story Title}" \
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
- [ ] {Assigned criterion 1}
- [ ] {Assigned criterion 2}
### Origin
Split from #{original_story_num}: {Original Story Title}
### TDD Test Cases
Test cases inherited from original story (see test plan).
### Definition of Done
- [ ] All acceptance criteria met
- [ ] TDD test cases pass
- [ ] Code reviewed
- [ ] No regressions
**Priority:** {P0|P1|P2}
**Parent Epic:** #{epic_num}
```
**Step 2:** Link to parent: `gh pmu sub add {epic_num} {new_story_num}`
**Step 3:** Track created story numbers
## Phase 5: Update Original Story
**Step 1:** Export body: `gh pmu view $story_num --body-stdout > .tmp-original.md`
**Step 2:** Add split notice section with new story references
**Step 3:** Save: `gh pmu edit $story_num -F .tmp-original.md && rm .tmp-original.md`
**Step 4:** Close original: `gh issue close $story_num --comment "Split into: #{new_1}, #{new_2}..."`
## Phase 6: Update Test Plan
**Step 1:** Find test plan via epic's PRD reference
**If no test plan:** Skip to Phase 7 with note
**Step 2:** Find original story section in test plan
**Step 3:** Replace with split story sections, redistributing test cases
**Step 4:** Commit test plan changes
## Phase 7: Report Completion
```
Story split complete: #{original_num} → {count} stories
Original story: #{original_num} - {Original Title} (CLOSED)
New stories created:
  • #{new_story_1}: {Title 1} (Priority: {P})
  • #{new_story_2}: {Title 2} (Priority: {P})
Parent epic: #{epic_num}
Charter compliance: ✅ All stories aligned (or ⚠️ Proceeded with warning)
Test plan: {Updated|Not applicable}
Next steps:
1. Work a split story: work #{new_story_1}
2. View epic progress: gh pmu sub list #{epic_num}
```
## Error Handling
| Situation | Response |
|-----------|----------|
| Story not found | "Issue #N not found" |
| Issue not a story | "Issue #N does not have 'story' label" |
| No parent epic | "Could not find parent epic. Link manually after split." |
| Less than 2 stories | "Split requires at least 2 new stories" |
| Original already closed | "Cannot split closed stories" |
**End of /split-story Command**
