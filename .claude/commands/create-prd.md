---
version: "v0.26.3"
description: Transform proposal into Agile PRD
argument-hint: "<proposal-path> | extract [<directory>]"
---

<!-- EXTENSIBLE -->
# /create-prd

Transform a proposal document into an Agile PRD (Product Requirements Document) with user stories, acceptance criteria, and epic groupings.

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

- Proposal exists in `Proposal/` directory
- (Recommended) Charter exists: `CHARTER.md` + `Inception/` artifacts

---

## Arguments

| Argument | Description |
|----------|-------------|
| `<proposal-path>` | Path to proposal file (e.g., `Proposal/Feature-X.md`) |
| `extract` | Extract PRD from existing codebase (requires `/charter` first) |
| `extract <directory>` | Extract from specific directory |

---

## Modes

| Mode | Invocation | Description |
|------|------------|-------------|
| **Promote** | `/create-prd Proposal/Feature.md` | Transform proposal to PRD (default) |
| **Extract** | `/create-prd extract` | Extract PRD from codebase |
| **Interactive** | `/create-prd` | Prompt for mode selection |

---

## Workflow (Promote Mode)

### Phase 1: Load Context

Load and validate context files:

| File | Required | Purpose |
|------|----------|---------|
| `<proposal-path>` | Yes | Source proposal |
| `CHARTER.md` | Recommended | Project scope validation |
| `Inception/Scope-Boundaries.md` | Recommended | In/out of scope |
| `Inception/Constraints.md` | Optional | Technical constraints |
| `Inception/Architecture.md` | Optional | System architecture |

**Load Anti-Hallucination Rules:**

| Context | Rules Path |
|---------|------------|
| Framework development (selfHosted) | `.min-mirror/Assistant/Anti-Hallucination-Rules-for-PRD-Work.md` |
| User projects | `{frameworkPath}/Assistant/Anti-Hallucination-Rules-for-PRD-Work.md` |

<!-- USER-EXTENSION-START: pre-analysis -->
<!-- USER-EXTENSION-END: pre-analysis -->

### Phase 2: Validate Against Charter

Compare proposal against charter scope:

| Finding | Action |
|---------|--------|
| Aligned | Proceed |
| Possibly misaligned | Ask for confirmation |
| Conflicts with out-of-scope | Flag conflict, offer resolution |

**Resolution Options:**
1. Expand charter scope
2. Defer to future release
3. Proceed anyway (creates drift)
4. Revise proposal

### Phase 3: Analyze Proposal Gaps

Parse proposal to identify present/missing elements:

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

**Question Rules:**
1. Reference specific proposal details
2. Only ask what's truly missing
3. Allow "skip" or "not sure" responses
4. Present 3-5 questions at a time

<!-- USER-EXTENSION-START: pre-transform -->
<!-- USER-EXTENSION-END: pre-transform -->

### Phase 4.5: Story Transformation

Transform proposal requirements into Agile user stories.

**Transformation Process:**
1. Identify USER (who benefits?)
2. Identify CAPABILITY (what can they do?)
3. Identify BENEFIT (why does it matter?)
4. Transform to story format

**Anti-Pattern Detection:**
Flag implementation details (file operations, internal changes, code-level details) and move to Technical Notes section.

<!-- USER-EXTENSION-START: post-transform -->
<!-- USER-EXTENSION-END: post-transform -->

### Phase 5: Priority Validation

Validate priority distribution before generation:

| Priority | Required Distribution |
|----------|----------------------|
| P0 (Must Have) | ≤40% of stories |
| P1 (Should Have) | 30-40% of stories |
| P2 (Could Have) | ≥20% of stories |

**Small PRD Exemption:** Skip validation for PRDs with <6 stories.

<!-- USER-EXTENSION-START: pre-diagram -->
<!-- USER-EXTENSION-END: pre-diagram -->

### Phase 5.5: Diagram Generation (Optional)

Generate UML diagrams as `.drawio.svg` files:

| Diagram Type | Default | When Appropriate |
|--------------|---------|------------------|
| Use Case | ON | User-facing features |
| Activity | ON | Multi-step workflows |
| Sequence | OFF | API/service interactions |
| Class | OFF | Data models, entities |
| Component | OFF | System architecture |
| State | OFF | State machines |

<!-- USER-EXTENSION-START: diagram-generator -->
<!-- USER-EXTENSION-END: diagram-generator -->

<!-- USER-EXTENSION-START: post-diagram -->
<!-- USER-EXTENSION-END: post-diagram -->

<!-- USER-EXTENSION-START: pre-generation -->
<!-- USER-EXTENSION-END: pre-generation -->

### Phase 6: Generate PRD

Create PRD in directory structure:

```
PRD/
└── {PRD-Name}/
    ├── PRD-{PRD-Name}.md
    └── Diagrams/
        └── {Epic-Name}/
            └── {type}-{description}.drawio.svg
```

**Note:** Existing flat PRDs (`PRD/PRD-{name}.md`) are grandfathered.

Create PRD document at `PRD/{name}/PRD-{name}.md`:

```markdown
# PRD: <Feature Name>

**Status:** Draft
**Created:** <date>
**Source Proposal:** <proposal-path>

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
> Do not create issues from this section.

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

### Phase 7: Next Steps

```
PRD created: PRD/{name}/PRD-{name}.md

Diagrams (if generated):
  PRD/{name}/Diagrams/{epic}/use-case-{desc}.drawio.svg
  PRD/{name}/Diagrams/{epic}/activity-{desc}.drawio.svg

Next steps:
1. Review and edit the PRD
2. Run Create-Backlog to generate issues
3. Assign to release

Run Create-Backlog now? (yes/no)
```

---

## Interactive Mode

For `/create-prd` (no arguments):

```
How would you like to create the PRD?

1. From a proposal document
2. From existing code (extraction)

> [user selects]
```

---

## Workflow (Extract Mode)

For `/create-prd extract` or `/create-prd extract <directory>`:

### Step 1: Check Inception/ Artifacts

```bash
test -d Inception
```

**If missing:**
```
No Inception/ artifacts found.

For best results, run /charter first to establish project context.

Options:
1. Run /charter now (recommended)
2. Proceed without charter context
3. Cancel
```

### Step 2: Load Analysis Skill

Load `Skills/codebase-analysis/SKILL.md` for analysis capabilities.

### Step 3: Analyze Codebase

Run analysis on target (entire project or specified directory):

| Analysis | Output |
|----------|--------|
| Tech stack detection | Languages, frameworks, dependencies |
| Architecture inference | Structure, layers, patterns |
| Test parsing | Features from test descriptions |
| NFR detection | NFRs from code patterns |

### Step 4: User Validation

Present extracted features with confidence levels for user selection.

### Step 5: Diagram Generation

Same diagram workflow as promote mode (Phase 5.5).

### Step 6: Generate PRD

Same Phase 6 output format, with additions:

- Confidence levels for each story
- Extraction metadata section
- Evidence citations for each feature

---

## Error Handling

| Situation | Response |
|-----------|----------|
| Proposal not found | "Proposal not found at <path>. Check the path?" |
| No Inception/ artifacts | "No charter context. Proceeding with limited validation." |
| User skips all questions | "Insufficient detail. Add more to proposal first?" |
| Empty proposal | "Proposal needs more detail. Minimum: problem + solution." |

---

## Quality Checklist

Before finalizing PRD:

- [ ] All user stories have acceptance criteria
- [ ] Requirements are prioritized (P0-P2)
- [ ] Priority distribution is valid (or <6 stories)
- [ ] Technical Notes separated from stories
- [ ] Out of scope explicitly stated
- [ ] Open questions flagged
- [ ] PRD is Create-Backlog compatible

---

**End of /create-prd Command**
