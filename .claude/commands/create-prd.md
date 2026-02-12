---
version: "v0.42.0"
description: Transform proposal into Agile PRD
argument-hint: "<issue-number> | extract [<directory>]"
---
<!-- EXTENSIBLE -->
# /create-prd
Transform a proposal document into an Agile PRD with user stories, acceptance criteria, and epic groupings.
## Available Extension Points
| Point | Location | Purpose |
|-------|----------|---------|
| `pre-analysis` | Before proposal analysis | Custom validation, preprocessing |
| `post-analysis` | After gap analysis | Custom analysis steps |
| `pre-transform` | Before story transformation | Pre-transformation hooks |
| `post-transform` | After story transformation | Story validation, enrichment |
| `pre-diagram` | Before diagram generation | Diagram setup |
| `post-diagram` | After diagram generation | Diagram post-processing |
| `pre-generation` | Before PRD generation | Template customization |
| `post-generation` | After PRD generation | Custom finalization |
| `diagram-generator` | Custom diagram generation | Replace default draw.io generator |
| `quality-checklist` | Custom quality checks | Additional validation rules |
---
## Prerequisites
- Proposal issue with `proposal` label
- Issue body contains link to `Proposal/[Name].md`
- Proposal document exists in `Proposal/`
- (Recommended) Charter: `CHARTER.md` + `Inception/` artifacts
---
## Arguments
| Argument | Description |
|----------|-------------|
| `<issue-number>` | Proposal issue number (`123` or `#123`) |
| `extract` | Extract PRD from existing codebase |
| `extract <directory>` | Extract from specific directory |
---
## Modes
| Mode | Invocation | Description |
|------|------------|-------------|
| **Issue-Driven** | `/create-prd 123` | Transform proposal to PRD |
| **Extract** | `/create-prd extract` | Extract PRD from codebase |
| **Interactive** | `/create-prd` | Prompt for mode selection |
---
## Execution Instructions
**REQUIRED:** Before executing:
1. **Generate Todo List:** Parse phases and extension points, use `TodoWrite` to create todos
2. **Include Extensions:** Add todo item for each non-empty `USER-EXTENSION` block
3. **Track Progress:** Mark todos `in_progress` → `completed` as you work
4. **Post-Compaction:** Re-read spec and regenerate todos after context compaction
**Todo Rules:** One todo per numbered phase/step; one todo per active extension; skip commented-out extensions.
---
## Workflow (Issue-Driven Mode)
### Phase 1: Fetch Proposal from Issue
**Step 1: Parse issue number**
```bash
issue_num="${1#\#}"
```
**Step 2: Fetch and validate issue**
```bash
gh issue view $issue_num --json labels,body --jq '.labels[].name' | grep -q "proposal"
```
**If not proposal:** Error: Issue #$issue_num does not have 'proposal' label.
**Step 3: Extract proposal document path**
Pattern: `/Proposal\/[A-Za-z0-9_-]+\.md/`
**If not found:** Error: Could not find proposal document link in issue body.
**Step 4: Load context files**
| File | Required | Purpose |
|------|----------|---------|
| `<extracted-proposal-path>` | Yes | Source proposal |
| `CHARTER.md` | Recommended | Project scope |
| `Inception/Scope-Boundaries.md` | Recommended | In/out of scope |
| `Inception/Constraints.md` | Optional | Technical constraints |
| `Inception/Architecture.md` | Optional | System architecture |
**Load Anti-Hallucination Rules:**
| Context | Rules Path |
|---------|------------|
| Framework (selfHosted) | `.min-mirror/Assistant/Anti-Hallucination-Rules-for-PRD-Work.md` |
| User projects | `{frameworkPath}/Assistant/Anti-Hallucination-Rules-for-PRD-Work.md` |

<!-- USER-EXTENSION-START: pre-analysis -->
<!-- USER-EXTENSION-END: pre-analysis -->

### Phase 2: Validate Against Charter
| Finding | Action |
|---------|--------|
| Aligned | Proceed |
| Possibly misaligned | Ask for confirmation |
| Conflicts with out-of-scope | Flag conflict, offer resolution |
**Resolution Options:** 1) Expand charter 2) Defer to future 3) Proceed anyway 4) Revise proposal
### Phase 3: Analyze Proposal Gaps
| Element | Detection Patterns | Gap Action |
|---------|-------------------|------------|
| Problem statement | "Problem:", "Issue:", first paragraph | Ask if missing |
| Proposed solution | "Solution:", "Approach:" | Ask if missing |
| User stories | "As a...", "User can..." | Generate questions |
| Acceptance criteria | "- [ ]", "Done when" | Generate questions |
| Priority | "P0-P3", "High/Medium/Low" | Ask if missing |

<!-- USER-EXTENSION-START: post-analysis -->
<!-- USER-EXTENSION-END: post-analysis -->

### Phase 4: Dynamic Question Generation
Generate context-aware questions for missing elements.
**Rules:** Reference proposal details; only ask what's missing; allow "skip"/"not sure"; present 3-5 questions at a time.

<!-- USER-EXTENSION-START: pre-transform -->
<!-- USER-EXTENSION-END: pre-transform -->

### Phase 4.5: Story Transformation
**Process:** 1) Identify USER 2) Identify CAPABILITY 3) Identify BENEFIT 4) Transform to story format
**Anti-Pattern Detection:** Flag implementation details, move to Technical Notes.

<!-- USER-EXTENSION-START: post-transform -->
<!-- USER-EXTENSION-END: post-transform -->

### Phase 5: Priority Validation
| Priority | Required Distribution |
|----------|----------------------|
| P0 (Must Have) | ≤40% |
| P1 (Should Have) | 30-40% |
| P2 (Could Have) | ≥20% |
**Small PRD Exemption:** Skip for <6 stories.

<!-- USER-EXTENSION-START: pre-diagram -->
<!-- USER-EXTENSION-END: pre-diagram -->

### Phase 5.5: Diagram Generation (Optional)
Generate UML diagrams as `.drawio.svg`:
| Diagram Type | Default | When Appropriate |
|--------------|---------|------------------|
| Use Case | ON | User-facing features |
| Activity | ON | Multi-step workflows |
| Sequence | OFF | API/service interactions |
| Class | OFF | Data models |
| Component | OFF | System architecture |
| State | OFF | State machines |

<!-- USER-EXTENSION-START: diagram-generator -->
<!-- USER-EXTENSION-END: diagram-generator -->

<!-- USER-EXTENSION-START: post-diagram -->
<!-- USER-EXTENSION-END: post-diagram -->

<!-- USER-EXTENSION-START: pre-generation -->
<!-- USER-EXTENSION-END: pre-generation -->

### Phase 6: Generate PRD
**Directory structure:**
```
PRD/
└── {PRD-Name}/
    ├── PRD-{PRD-Name}.md
    └── Diagrams/
        └── {Epic-Name}/
            └── {type}-{description}.drawio.svg
```
**Note:** Existing flat PRDs (`PRD/PRD-{name}.md`) are grandfathered.
**PRD Template:** `PRD/{name}/PRD-{name}.md`
```markdown
# PRD: <Feature Name>
**Status:** Draft
**Created:** <date>
**Source Proposal:** <proposal-path>
**Proposal Issue:** #<issue-number> (closed)
---
## Overview
<From proposal>
---
## Epics
### Epic 1: <Theme>
Stories: 1.1, 1.2, 1.3
---
## User Stories
### Story 1.1: <Title>
**As a** <user type>
**I want** <capability>
**So that** <benefit>
**Acceptance Criteria:**
- [ ] <criterion>
**Priority:** P0 - Must Have
---
## Diagrams
| Epic | Diagram | Description |
|------|---------|-------------|
| Epic 1 | `Diagrams/Epic-1/use-case-desc.drawio.svg` | Actor interactions |
> **Traceability:** Diagram elements cite source (story ID, AC number).
---
## Technical Notes
> Implementation hints, not requirements.
---
## Out of Scope
<Explicit exclusions>
---
## Dependencies
<Cross-references>
---
## Open Questions
<Unresolved items>
---
*Generated by create-prd skill*
*Ready for Create-Backlog*
```

<!-- USER-EXTENSION-START: post-generation -->
<!-- USER-EXTENSION-END: post-generation -->

<!-- USER-EXTENSION-START: quality-checklist -->
<!-- USER-EXTENSION-END: quality-checklist -->

### Phase 6.5: Generate TDD Test Plan
**Step 1: Load test configuration**
| Source File | Data to Extract |
|-------------|-----------------|
| `Inception/Test-Strategy.md` | Test framework, coverage targets |
| `Inception/Tech-Stack.md` | Language (for test syntax) |
**Fallback:** Check `IDPF-Agile/Agile-Core.md` TDD section; warn if missing; use defaults (80% unit, "TBD" framework).
**Step 2: Generate test plan at `PRD/{name}/Test-Plan-{name}.md`**
```markdown
# TDD Test Plan: {Feature Name}
## Source
- **PRD:** PRD-{name}.md
- **Created:** {date}
- **Approval Issue:** #{to-be-created}
- **Test Strategy:** Inception/Test-Strategy.md
## Test Strategy Overview
| Level | Scope | Framework |
|-------|-------|-----------|
| Unit | Individual functions | {from Test-Strategy.md} |
| Integration | Cross-component flows | {from Test-Strategy.md} |
| E2E | Critical user journeys | {from Test-Strategy.md or "TBD"} |
## Epic Test Coverage
### Epic 1: {Name}
| Story | Acceptance Criteria | Test Cases |
|-------|--------------------| -----------|
| {Story title} | {AC from PRD} | ✓ Test valid input |
|               |               | ✓ Test invalid input |
|               |               | ✓ Test edge case |
## Integration Test Points
| Components | Test Scenario | Priority |
|------------|---------------|----------|
| [Component A] ↔ [Component B] | Data flows correctly | P0 |
## E2E Scenarios
| Scenario | User Journey | Expected Outcome |
|----------|--------------|------------------|
| Happy path | User completes primary flow | Success state |
| Error recovery | User encounters error | Graceful handling |
## Coverage Targets
| Type | Target | Rationale |
|------|--------|-----------|
| Unit | {default: 80%+} | Core logic |
| Integration | {default: Key flows} | Boundary verification |
| E2E | {default: Critical paths} | Journey validation |
## Approval Checklist
- [ ] All PRD acceptance criteria have test cases
- [ ] Edge cases and error conditions identified
- [ ] Integration points between epics mapped
- [ ] E2E scenarios cover critical journeys
- [ ] Coverage targets are realistic
```
**Derivation Rules:** Parse AC from PRD; generate 2-3 test cases per criterion; identify integration points; extract E2E from user journeys.
### Phase 6.6: Create Test Plan Approval Issue
```bash
gh pmu create --label test-plan --label approval-required --assignee @me \
  --title "Approve Test Plan: {Name}" \
  --body "## Test Plan Review
A TDD test plan has been generated for **{Name}**.
**Test Plan:** PRD/{name}/Test-Plan-{name}.md
**PRD:** PRD/{name}/PRD-{name}.md
## Review Checklist
- [ ] Test cases cover all acceptance criteria
- [ ] Edge cases and error scenarios included
- [ ] Integration test points are complete
- [ ] E2E scenarios cover critical paths
- [ ] Coverage targets are appropriate
## Instructions
1. Review the test plan document
2. Check all boxes above when satisfied
3. Comment with any required changes
4. Close this issue to approve
**⚠️ Create-Backlog is blocked until this issue is closed.**" \
  --status backlog
```
**Update test plan with issue number after creation.**
### Phase 7: Proposal Lifecycle Completion
**Only for Issue-Driven Mode**
**Step 1: Move proposal**
```bash
git mv Proposal/{Name}.md Proposal/Implemented/{Name}.md
```
**Step 2: Close proposal issue**
```bash
gh issue close $issue_num --comment "Transformed to PRD: PRD/{name}/PRD-{name}.md"
gh pmu move $issue_num --status done
```
**Step 3: Create PRD tracking issue**
```bash
gh pmu create --label prd --assignee @me \
  --title "PRD: {Name}" \
  --body "## PRD Document
**File:** PRD/{name}/PRD-{name}.md
**Test Plan:** PRD/{name}/Test-Plan-{name}.md
**Source Proposal:** #$issue_num (closed)
## Status
- [ ] PRD reviewed
- [ ] Test plan approved (see #{test_plan_issue})
- [ ] Ready for backlog creation
## Next Step
1. Review and close test plan approval issue: #{test_plan_issue}
2. Run: \`/create-backlog {this-issue-number}\`" \
  --status backlog
```
**Step 4: Report completion**
```
PRD created: PRD/{name}/PRD-{name}.md
Test Plan created: PRD/{name}/Test-Plan-{name}.md
Proposal lifecycle completed:
  ✓ Proposal archived: Proposal/Implemented/{Name}.md
  ✓ Proposal issue closed: #{issue_num}
  ✓ PRD tracking issue created: #{prd_issue_num}
  ✓ Test plan approval issue created: #{test_plan_issue_num}
Diagrams (if generated):
  PRD/{name}/Diagrams/{epic}/use-case-{desc}.drawio.svg
⚠️ APPROVAL REQUIRED before Create-Backlog:
  Review and close: #{test_plan_issue_num}
Next steps:
1. Review PRD and test plan
2. Approve test plan by closing #{test_plan_issue_num}
3. Run /create-backlog {prd_issue_num}
```
---
## Interactive Mode
For `/create-prd` (no arguments):
```
How would you like to create the PRD?
1. From a proposal issue (enter issue number)
2. From existing code (extraction)
> [user selects]
```
**If 1:** Prompt for issue number, proceed with Issue-Driven Mode.
---
## Workflow (Extract Mode)
### Step 1: Check Inception/ Artifacts
```bash
test -d Inception
```
**If missing:**
```
No Inception/ artifacts found.
Options:
1. Run /charter now (recommended)
2. Proceed without charter context
3. Cancel
```
### Step 2: Load Analysis Skill
Load `Skills/codebase-analysis/SKILL.md`
### Step 3: Analyze Codebase
| Analysis | Output |
|----------|--------|
| Tech stack detection | Languages, frameworks, dependencies |
| Architecture inference | Structure, layers, patterns |
| Test parsing | Features from test descriptions |
| NFR detection | NFRs from code patterns |
### Step 4: User Validation
Present extracted features with confidence levels.
### Step 5: Diagram Generation
Same as Phase 5.5.
### Step 6: Generate PRD
Same Phase 6 format with: confidence levels, extraction metadata, evidence citations.
---
## Error Handling
| Situation | Response |
|-----------|----------|
| Issue not found | "Issue #N not found." |
| Missing proposal label | "Issue #N does not have 'proposal' label." |
| Proposal path not in body | "Could not find proposal document link." |
| Proposal file not found | "Proposal not found at <path>." |
| No Inception/ artifacts | "No charter context. Proceeding with limited validation." |
| User skips all questions | "Insufficient detail. Add more to proposal?" |
| Empty proposal | "Proposal needs more detail. Minimum: problem + solution." |
---
## Quality Checklist
- [ ] All stories have acceptance criteria
- [ ] Requirements prioritized (P0-P2)
- [ ] Priority distribution valid (or <6 stories)
- [ ] Technical Notes separated
- [ ] Out of scope stated
- [ ] Open questions flagged
- [ ] PRD is Create-Backlog compatible
---
## Technical Skills Mapping
### Step 1: Load Skill Registry and Config
```bash
cat .claude/metadata/skill-registry.json
cat framework-config.json
```
### Step 2: Extract Technical Requirements
Parse PRD for: Tech Requirements, acceptance criteria technical terms, NFR mentions (CI/CD, security, performance).
### Step 3: Match Against Skill Triggers
| PRD Content | Matched Skill |
|-------------|---------------|
| "CI/CD pipeline" | `ci-cd-pipeline-design` |
| "E2E tests" | `playwright-setup` |
| "API versioning" | `api-versioning` |
### Step 4: Identify New Skills
```
newSkills = matchedSkills.filter(s => !projectSkills.includes(s))
```
### Step 5: Present New Skills
**ASK USER:**
```
PRD mentions technical requirements suggesting additional skills:
- ci-cd-pipeline-design (CI/CD pipeline in NFRs)
- api-versioning (API versioning for service integration)
Add to project skills? (yes/no/edit)
```
### Step 6: Update framework-config.json
```javascript
config.projectSkills = [...existingSkills, ...newSkills];
fs.writeFileSync('framework-config.json', JSON.stringify(config, null, 2));
```
Report: `Added skills: ci-cd-pipeline-design, api-versioning`
---
**End of /create-prd Command**
