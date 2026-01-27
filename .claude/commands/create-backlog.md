---
version: "v0.33.2"
description: Create GitHub epics/stories from PRD (project)
argument-hint: "<issue-number> (e.g., 151)"
---

<!-- MANAGED -->
# /create-backlog

Create GitHub epics and stories from an approved PRD with embedded TDD test cases.

---

## Arguments

| Argument | Description |
|----------|-------------|
| `<prd-issue-number>` | PRD tracking issue number (e.g., `151` or `#151`) |

---

## Execution Instructions

**REQUIRED:** Before executing this command:

1. **Create Todo List:** Use `TodoWrite` to create todos from the steps below
2. **Track Progress:** Mark todos `in_progress` → `completed` as you work
3. **Resume Point:** If interrupted, todos show where to continue

**Example todo structure:**
```
- [ ] Phase 1: Fetch and validate PRD issue
- [ ] Phase 2: Check test plan approval gate
- [ ] Phase 3: Parse PRD for epics and stories
- [ ] Phase 4: Load test cases from approved test plan
- [ ] Phase 5: Create epic issues
- [ ] Phase 6: Create story issues with test cases
- [ ] Phase 7: Update PRD status
- [ ] Phase 8: Skill suggestions (optional)
```

---

## Prerequisites

- PRD tracking issue exists with `prd` label
- PRD issue body contains link to `PRD/PRD-[Name].md`
- Test plan exists: `PRD/[Name]/Test-Plan-[Name].md`
- Test plan approval issue is **closed** (approved)

---

## Phase 1: Fetch and Validate PRD Issue

**Step 1: Parse issue number**

Accept `151` or `#151` format:
```bash
# Strip leading # if present
issue_num="${1#\#}"
```

**Step 2: Fetch and validate issue**

```bash
gh issue view $issue_num --json labels,body --jq '.labels[].name' | grep -q "prd"
```

**If not a PRD issue:**
```
Error: Issue #$issue_num does not have the 'prd' label.
This command requires a PRD tracking issue created by /create-prd.
```

**Step 3: Extract PRD document path**

Parse issue body for PRD document reference:
```
Pattern: /PRD\/[A-Za-z0-9_-]+\/PRD-[A-Za-z0-9_-]+\.md/
```

---

## Phase 2: Test Plan Approval Gate

**BLOCKING:** Backlog creation is blocked until test plan is approved.

**Step 1: Find test plan approval issue**

Search for test plan approval issue linked to this PRD:
```bash
gh issue list --label "test-plan" --label "approval-required" --state open --json number,title
```

**Step 2: Check approval status**

| Approval Issue State | Action |
|----------------------|--------|
| **Open** | BLOCK - Show message and exit |
| **Closed** | PROCEED - Continue with backlog creation |
| **Not found** | WARN - Proceed but note missing test plan |

**If blocked (approval issue open):**
```
⚠️ Test plan not yet approved.

Test Plan: PRD/{name}/Test-Plan-{name}.md
Approval Issue: #{number} (OPEN)

Please review and close the approval issue before creating backlog items.
The test plan ensures all acceptance criteria have corresponding test cases.

To approve: Review test plan, then close #{number}
```

---

## Phase 3: Parse PRD for Epics and Stories

**Step 1: Load PRD document**

Read `PRD/{name}/PRD-{name}.md` and extract:
- Epics (Feature Areas)
- Stories (Capabilities under each Epic)
- Acceptance criteria for each story

**Step 2: Structure extraction**

| PRD Section | Maps To |
|-------------|---------|
| `## Epics` / `### Epic N:` | GitHub issue with `epic` label |
| User stories under epic | GitHub issues with `story` label |
| Acceptance criteria | Story body checkboxes |
| Priority (P0/P1/P2) | Priority field |

---

## Phase 4: Load Test Cases from Test Plan

**Step 1: Load test plan**

Read `PRD/{name}/Test-Plan-{name}.md`

**Step 2: Match test cases to stories**

For each story, find corresponding test cases in test plan:
- Match by story title
- Match by acceptance criteria text

**Step 3: Detect tech stack**

From `Inception/Tech-Stack.md` or PRD, determine test syntax:

| Tech Stack | Test Syntax |
|------------|-------------|
| TypeScript/JavaScript | Jest/Vitest `test()` |
| Python | pytest `def test_*():` |
| Go | `func Test*(t *testing.T)` |
| Rust | `#[test] fn test_*()` |
| Unknown | Generic pseudocode |

---

## Phase 5: Create Epic Issues

For each epic in PRD:

**Note:** Use `gh pmu create` to automatically add issues to the project board.

```bash
gh pmu create \
  --title "Epic: {Epic Name}" \
  --label "epic" \
  --status backlog \
  -F .tmp-epic-body.md
```

**Epic body template** (write to `.tmp-epic-body.md`):

```markdown
## Epic: {Epic Name}

**PRD:** PRD/{name}/PRD-{name}.md
**PRD Tracker:** #{prd_issue_number}
**Test Plan:** PRD/{name}/Test-Plan-{name}.md

## Description

{Epic description from PRD}

## Success Criteria

{Success criteria from PRD}

## Stories

Stories will be linked as sub-issues.
```

Clean up temp file after creation: `rm .tmp-epic-body.md`

---

## Phase 6: Create Story Issues with Test Cases

For each story under an epic:

**Note:** Use `gh pmu create` to automatically add issues to the project board.

```bash
gh pmu create \
  --title "Story: {Story Title}" \
  --label "story" \
  --status backlog \
  -F .tmp-story-body.md
```

Clean up temp file after creation: `rm .tmp-story-body.md`

Then link to parent epic:
```bash
gh pmu sub add {epic_number} {story_number}
```

### Story Body Template

```markdown
## Story: {Title}

### Description

As a {user}, I want {goal} so that {benefit}.

### Relevant Skills

<!-- Read from framework-config.json projectSkills array -->
<!-- For each skill, lookup description from .claude/metadata/skill-registry.json -->

**If projectSkills configured:**
- {skill-name} - {description from skill-registry.json}
- {skill-name} - {description}

Load skill: `read Skills/{skill-name}/SKILL.md`

**If no projectSkills:**
No project skills configured. Run `/charter` to set up project-specific skills.

### Acceptance Criteria

- [ ] {Criterion 1}
- [ ] {Criterion 2}
- [ ] {Criterion 3}

### Documentation (if applicable)

- [ ] Design decisions documented (update existing or create `Construction/Design-Decisions/YYYY-MM-DD-{topic}.md`)
- [ ] Tech debt logged (update existing or create `Construction/Tech-Debt/YYYY-MM-DD-{topic}.md`)

**Guidelines:** Skip trivial findings. Update existing docs rather than duplicating. For significant tech debt, create an enhancement issue.

### TDD Test Cases

**Source:** [Test-Plan-{name}.md](PRD/{name}/Test-Plan-{name}.md#epic-story-section)

Write these tests BEFORE implementation:

#### Unit Tests

```{language}
// Test: {criterion 1 - valid case}
test('{criterion 1} succeeds with valid input', () => {
  // Arrange: set up test data
  // Act: call function under test
  // Assert: verify expected outcome
});

// Test: {criterion 1 - invalid case}
test('{criterion 1} rejects invalid input', () => {
  // Arrange: set up invalid data
  // Act: call function under test
  // Assert: expect error/rejection
});

// Additional tests for criteria 2, 3...
```

#### Edge Cases

- [ ] Empty/null input handling
- [ ] Boundary values
- [ ] Error recovery

### Definition of Done

- [ ] All TDD test cases pass
- [ ] Code coverage ≥ 80%
- [ ] No skipped tests
- [ ] Edge cases handled
- [ ] Acceptance criteria verified

**Priority:** {P0|P1|P2}
```

---

## Phase 7: Update PRD Status

**Step 1: Update PRD document**

Change status from "Draft" to "Backlog Created":
```markdown
**Status:** Backlog Created
```

**Step 2: Update PRD tracking issue body**

Prepend instruction banner to PRD tracker issue body:

```markdown
> **📋 PRD In Progress** — When all stories are complete, run `/complete-prd {issue_number}` to verify and close.

## Backlog Summary

✅ Epics: {count}
✅ Stories: {count}
✅ Test cases embedded in each story

---

{original PRD issue body}
```

Use `gh pmu edit` with temp file to update the body.

**Step 3: Add summary comment**

```bash
gh issue comment $issue_num --body "## Backlog Created

✅ Epics: {count}
✅ Stories: {count}

**Next:** Work stories via \`work #N\`
**When complete:** Run \`/complete-prd $issue_num\`"
```

**Step 4: Move PRD to In Progress**

```bash
gh pmu move $issue_num --status in_progress
```

**Note:** PRD remains open until `/complete-prd` verifies all stories are Done.

---

## Phase 8: Skill Suggestions (Optional)

**Purpose:** Suggest relevant skills based on technologies mentioned in created stories.

**Step 1: Check opt-out setting**

Read `framework-config.json`:
```json
{
  "skillSuggestions": false  // If present and false, skip this phase
}
```

**If skillSuggestions is false:** Skip to Output Summary.

**Step 2: Load skill keywords registry**

Read `.claude/metadata/skill-keywords.json` for keyword-to-skill mappings.

**Step 3: Collect story content for analysis**

Combine all story titles and acceptance criteria text:
```
{story_1_title} {story_1_criteria}
{story_2_title} {story_2_criteria}
...
```

**Step 4: Match keywords (case-insensitive)**

For each skill in `skillKeywords`:
- Check if any keywords appear in story content
- Track which stories triggered each match

Also check `groupKeywords` for broader matches (e.g., "tdd" → multiple TDD skills).

**Step 5: Filter already-installed skills**

Read `framework-config.json` `projectSkills` array.
Remove skills already in `projectSkills` from matches.

**Step 6: If matches found, display and prompt**

```
Detected technologies in your backlog:
┌──────────────────────────┬─────────────────────────┬──────────────┐
│ Skill                    │ Matched Keywords        │ Stories      │
├──────────────────────────┼─────────────────────────┼──────────────┤
│ playwright-setup         │ "e2e tests"             │ #{N}, #{M}   │
│ electron-development     │ "Electron", "IPC"       │ #{N}, #{O}   │
│ error-handling-patterns  │ "retry logic"           │ #{P}         │
└──────────────────────────┴─────────────────────────┴──────────────┘
```

**ASK USER:** Install suggested skills? (y/n/select)

| Response | Action |
|----------|--------|
| `y` / `yes` | Install all matched skills |
| `n` / `no` | Skip skill installation |
| `select` | Present numbered list for individual selection |

**Step 7: Install selected skills**

```bash
node .claude/scripts/shared/install-skill.js {skill-names...}
```

Report results:
```
Installing skills...
✓ playwright-setup - Installed (5 resources)
✓ electron-development - Installed (8 resources)
⊘ error-handling-patterns - Already installed (skipped)
```

---

## Output Summary

```
Backlog created from PRD: PRD/{name}/PRD-{name}.md

Epics created:
  • #{N}: Epic: {Name} ({X} stories)
  • #{N}: Epic: {Name} ({Y} stories)

Total: {E} epics, {S} stories

Test cases embedded:
  ✓ {T} test cases pulled from Test-Plan-{name}.md
  ✓ Test skeletons generated ({language} syntax)

Skills suggested: {count} (installed: {installed_count})

PRD status: Backlog Created

Next steps:
1. Assign stories to release: gh pmu move #{story} --branch current
2. Start work: work #{story}
```

---

## Error Handling

| Situation | Response |
|-----------|----------|
| PRD issue not found | "Issue #N not found. Check the issue number?" |
| Issue missing prd label | "Issue #N does not have 'prd' label." |
| PRD path not in issue body | "Could not find PRD document link in issue body." |
| PRD file not found | "PRD not found at <path>. Check the file exists?" |
| Test plan not found | "Warning: No test plan found. Stories created without embedded test cases." |
| Test plan not approved | BLOCK with approval instructions |
| No epics in PRD | "PRD contains no epics. Add epics before creating backlog." |

---

**End of /create-backlog Command**
