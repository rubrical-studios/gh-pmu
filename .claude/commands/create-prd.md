---
version: "v0.31.0"
description: Transform proposal into Agile PRD
argument-hint: "<issue-number> | extract [<directory>]"
---
<!-- EXTENSIBLE -->
# /create-prd
Transform a proposal document into an Agile PRD with user stories, acceptance criteria, and epic groupings.
## Available Extension Points
| Point | Location | Purpose |
|-------|----------|---------|
| `pre-analysis` | Before proposal analysis | Custom validation |
| `post-analysis` | After gap analysis | Custom analysis |
| `pre-transform` | Before story transformation | Pre-transformation hooks |
| `post-transform` | After story transformation | Story validation |
| `pre-diagram` | Before diagram generation | Diagram setup |
| `post-diagram` | After diagram generation | Post-processing |
| `pre-generation` | Before PRD generation | Template customization |
| `post-generation` | After PRD generation | Finalization |
| `diagram-generator` | Custom diagram generation | Replace default |
| `quality-checklist` | Custom quality checks | Additional validation |
## Prerequisites
- Proposal issue with `proposal` label
- Proposal body contains link to `Proposal/[Name].md`
- (Recommended) Charter exists: `CHARTER.md` + `Inception/`
## Arguments
| Argument | Description |
|----------|-------------|
| `<issue-number>` | Proposal issue number (e.g., `123` or `#123`) |
| `extract` | Extract PRD from existing codebase |
| `extract <directory>` | Extract from specific directory |
## Modes
| Mode | Invocation | Description |
|------|------------|-------------|
| **Issue-Driven** | `/create-prd 123` | Transform proposal to PRD |
| **Extract** | `/create-prd extract` | Extract PRD from codebase |
| **Interactive** | `/create-prd` | Prompt for mode selection |
## Workflow (Issue-Driven Mode)
### Phase 1: Fetch Proposal from Issue
**Step 1:** Parse issue number (strip leading #)
**Step 2:** Validate `proposal` label
**Step 3:** Extract proposal path: `/Proposal\/[A-Za-z0-9_-]+\.md/`
**Step 4:** Load context (proposal, CHARTER.md, Inception/*)
**Load Anti-Hallucination Rules:** `.min-mirror/Assistant/Anti-Hallucination-Rules-for-PRD-Work.md`

<!-- USER-EXTENSION-START: pre-analysis -->
<!-- USER-EXTENSION-END: pre-analysis -->

### Phase 2: Validate Against Charter
| Finding | Action |
|---------|--------|
| Aligned | Proceed |
| Possibly misaligned | Ask confirmation |
| Conflicts with out-of-scope | Flag, offer resolution |
### Phase 3: Analyze Proposal Gaps
| Element | Detection | Gap Action |
|---------|-----------|------------|
| Problem statement | "Problem:", "Issue:" | Ask if missing |
| Proposed solution | "Solution:", "Approach:" | Ask if missing |
| User stories | "As a...", "User can..." | Generate questions |
| Acceptance criteria | "- [ ]", "Done when" | Generate questions |
| Priority | "P0-P3", "High/Medium/Low" | Ask if missing |

<!-- USER-EXTENSION-START: post-analysis -->
<!-- USER-EXTENSION-END: post-analysis -->

### Phase 4: Dynamic Question Generation
Generate context-aware questions for missing elements. Only ask what's missing, 3-5 at a time.

<!-- USER-EXTENSION-START: pre-transform -->
<!-- USER-EXTENSION-END: pre-transform -->

### Phase 4.5: Story Transformation
Transform requirements into Agile stories (USER, CAPABILITY, BENEFIT).
Flag implementation details → move to Technical Notes.

<!-- USER-EXTENSION-START: post-transform -->
<!-- USER-EXTENSION-END: post-transform -->

### Phase 5: Priority Validation
| Priority | Distribution |
|----------|-------------|
| P0 (Must Have) | ≤40% |
| P1 (Should Have) | 30-40% |
| P2 (Could Have) | ≥20% |
Skip for PRDs with <6 stories.

<!-- USER-EXTENSION-START: pre-diagram -->
<!-- USER-EXTENSION-END: pre-diagram -->

### Phase 5.5: Diagram Generation (Optional)
| Type | Default | When Appropriate |
|------|---------|------------------|
| Use Case | ON | User-facing features |
| Activity | ON | Multi-step workflows |
| Sequence | OFF | API interactions |
| Class | OFF | Data models |

<!-- USER-EXTENSION-START: diagram-generator -->
<!-- USER-EXTENSION-END: diagram-generator -->

<!-- USER-EXTENSION-START: post-diagram -->
<!-- USER-EXTENSION-END: post-diagram -->

<!-- USER-EXTENSION-START: pre-generation -->
<!-- USER-EXTENSION-END: pre-generation -->

### Phase 6: Generate PRD
Create: `PRD/{name}/PRD-{name}.md` with Overview, Epics, User Stories, Diagrams, Technical Notes, Out of Scope, Dependencies, Open Questions.

<!-- USER-EXTENSION-START: post-generation -->
<!-- USER-EXTENSION-END: post-generation -->

<!-- USER-EXTENSION-START: quality-checklist -->
<!-- USER-EXTENSION-END: quality-checklist -->

### Phase 6.5: Generate TDD Test Plan
Create: `PRD/{name}/Test-Plan-{name}.md`
- Test Strategy Overview (Unit, Integration, E2E)
- Epic Test Coverage table (Story → AC → Test Cases)
- Integration Test Points
- E2E Scenarios
- Coverage Targets
**Derivation:** Parse each AC → generate 2-3 test cases (valid, invalid, edge)
### Phase 6.6: Create Test Plan Approval Issue
```bash
gh pmu create --label test-plan --label approval-required --assignee @me \
  --title "Approve Test Plan: {Name}" \
  --body "..." --status backlog
```
**⚠️ Create-Backlog is blocked until this issue is closed.**
### Phase 7: Proposal Lifecycle Completion
**Step 1:** Move proposal: `git mv Proposal/{Name}.md Proposal/Implemented/{Name}.md`
**Step 2:** Close proposal issue with PRD reference
**Step 3:** Create PRD tracking issue with `prd` label
**Step 4:** Report completion with next steps
## Interactive Mode
Prompt: "1. From proposal issue, 2. From existing code"
## Workflow (Extract Mode)
**Step 1:** Check Inception/ artifacts (recommend /charter first)
**Step 2:** Load codebase-analysis skill
**Step 3:** Analyze codebase (tech, architecture, tests, NFRs)
**Step 4:** User validation with confidence levels
**Step 5:** Diagram generation
**Step 6:** Generate PRD with extraction metadata
## Error Handling
| Situation | Response |
|-----------|----------|
| Issue not found | "Issue #N not found" |
| Missing proposal label | "Issue #N does not have 'proposal' label" |
| Proposal path not in body | "Could not find proposal document link" |
| Empty proposal | "Proposal needs more detail" |
## Quality Checklist
- [ ] All stories have acceptance criteria
- [ ] Priority distribution valid
- [ ] Technical Notes separated
- [ ] Out of scope stated
- [ ] PRD is Create-Backlog compatible
**End of /create-prd Command**
