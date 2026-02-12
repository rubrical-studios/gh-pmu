---
version: "v0.42.2"
description: Review issues with type-specific criteria (project)
argument-hint: "#issue [#issue...]"
---

<!-- EXTENSIBLE -->
# /review-issue
Reviews one or more GitHub issues with type-specific criteria based on labels (bug, enhancement, story, epic). Tracks review count in issue body and posts structured findings as comment. No linked file — operates directly on the issue.
**Extension Points:** See `.claude/metadata/extension-points.json` or run `/extensions list --command review-issue`
---
## Prerequisites
- `gh pmu` extension installed
- `.gh-pmu.yml` configured in repository root
---
## Arguments
| Argument | Required | Description |
|----------|----------|-------------|
| `#issue` | Yes | One or more issue numbers (e.g., `#42` or `42 43 44`) |
| `--with` | No | Comma-separated domain extensions (e.g., `--with security,performance`) or `--with all` |
Accepts multiple issue numbers: `/review-issue #42 #43 #44` — reviews each sequentially.
---
## Execution Instructions
**REQUIRED:** Before executing:
1. **Generate Todo List:** Parse workflow steps, use `TodoWrite` to create todos
2. **Include Extensions:** Add todo item for each non-empty `USER-EXTENSION` block
3. **Track Progress:** Mark todos `in_progress` → `completed` as you work
4. **Post-Compaction:** Re-read spec and regenerate todos after context compaction
**Todo Rules:** One todo per numbered step; one todo per active extension; skip commented-out extensions.
---
## Workflow
**For multiple issues:** Process each issue sequentially through Steps 1–6.
### Step 1: Fetch Issue
```bash
gh issue view $ISSUE --json number,title,body,state,labels
```
**If not found:** `"Issue #$ISSUE not found."` → **STOP** (skip to next if batch)
**If closed:** `"Issue #$ISSUE is closed. Review anyway? (y/n)"` — proceed only if user confirms.
### Step 2: Determine Issue Type
Check issue labels to determine type-specific review criteria:
```bash
gh issue view $ISSUE --json labels --jq '.labels[].name'
```
| Label Found | Review Type |
|-------------|-------------|
| `bug` | Bug review criteria |
| `enhancement` | Enhancement review criteria |
| `story` | Story review criteria |
| `epic` | Epic review criteria |
| `test-plan` | **Redirect** → `/review-test-plan #$ISSUE` |
| `proposal` | **Redirect** → `/review-proposal #$ISSUE` |
| `prd` | **Redirect** → `/review-prd #$ISSUE` |
| None recognized | Generic review criteria |
**If redirect label detected** (`test-plan`, `proposal`, `prd`):
```
Issue #$ISSUE has label '$LABEL' — redirecting to the specialized review command.
```
Invoke the appropriate command using the Skill tool. → **STOP** (the specialized command handles the review)
**If multiple recognized labels:** Use first recognized label as primary type. Redirect labels take precedence over review labels.
**If no recognized labels:**
```
Issue #$ISSUE has no recognized label (bug, enhancement, story, epic). Using generic review criteria.
```
Continue with generic criteria (non-blocking).

<!-- USER-EXTENSION-START: pre-review -->
<!-- USER-EXTENSION-END: pre-review -->

### Step 2b: Extension Loading
**If `--with` is specified:**
1. Read `.claude/metadata/review-extensions.json`
2. Parse the `--with` value:
   - `--with all` → load all registered extensions
   - `--with security,performance` → load specified extensions (trim spaces from comma-separated list)
3. For each requested extension ID:
   - Look up the `source` path in the registry
   - Read the criteria file content
   - Extract the **Issue Review Questions** section
4. If an extension ID is not found in the registry, warn: `"Unknown extension: {id}. Available: security, accessibility, performance, chaos, contract, qa"`
5. Store loaded criteria for use in Step 3
**Error Handling (Extension Loading):**
- **Registry not found:** `"Review extensions registry not found. Run hub update or check installation."` → fall back to standard review only
- **Registry malformed:** `"Review extensions registry is malformed. Run hub update or check installation."` → fall back to standard review only
- **Criteria file not found:** `"Warning: Review criteria file not found for '{domain}'. Skipping domain. Update hub to resolve."` → continue with remaining domains
- **All criteria files missing:** `"No review criteria files found. Running standard review only."` → fall back to standard review only
**If `--with` is not specified:** Skip extension loading (standard review only).
### Step 3: Perform Review
Evaluate the issue using a two-phase approach: **auto-evaluate objective criteria** by reading the issue, then **ask the user only about subjective criteria** via `AskUserQuestion`. Filter criteria by `reviewMode` from `framework-config.json`.
**Step 3a: Load reviewMode Configuration**
```javascript
const { getReviewMode, shouldEvaluate, isObjective } = require('./.claude/scripts/shared/lib/review-mode.js');
const mode = getReviewMode(); // 'solo', 'team', or 'enterprise'
```

<!-- USER-EXTENSION-START: criteria-customize -->
<!-- USER-EXTENSION-END: criteria-customize -->

**Step 3b: Auto-Evaluate Objective Criteria**
For each **objective** criterion applicable to the current reviewMode, evaluate autonomously by reading the issue content. Do NOT ask the user.
**Common Objective Criteria:**
| Criterion | Auto-Check Method |
|-----------|-------------------|
| `title-clear` | Verify title is specific (>3 words), contains actionable context, not generic |
| `ac-present-testable` | Check for `- [ ]` checkbox items; verify each contains measurable/verifiable statement |
| `labels-correct` | Verify issue has at least one type label matching its content |
| `priority-set` | Check via `gh pmu view $ISSUE --json=priority` |
| `dependencies-identified` | Check for `Refs #`, `Depends on #`, `Blocks #`, or `**Refs:**` patterns |
| `effort-estimate` | Check for estimate/effort section or field in body |
**Type-Specific Objective Checks:**
**For Bug — auto-check:**
- Reproduction steps present: look for numbered list or "Steps to Reproduce" section
- Expected vs actual behavior documented: look for "Expected" and "Actual" sections
- Environment info present: look for "Environment", browser, OS, version references
- Severity proportionate to impact: compare severity/priority label to described user impact
**For Story — auto-check:**
- User story format: check for "As a ... I want ... So that ..." pattern
- Acceptance criteria present: check for `- [ ]` items
- AC structured as testable assertions: verify each `- [ ]` item is measurable/verifiable (TDD-compatible)
- Scope supports incremental implementation: check if ACs are small, independent, testable units
- Canonical story body template: compare body structure against expected template. If NOT matching, flag for repair
**For Epic — auto-check:**
- Sub-issues exist: run `gh pmu sub list $ISSUE` and verify non-empty result
- Scope description present: check for description of what the epic covers
- Success criteria for the epic defined: check for epic-level success criteria or definition of done
- Dependencies between sub-issues identified: check sub-issue bodies for cross-references
- **MUST: Iterate over each sub-issue and review** — for each sub-issue, run review criteria recursively
**For Enhancement — auto-check:**
- Motivation/rationale present: check for "Motivation", "Why", or rationale section
- Proposed solution described: check for "Proposed Solution" or solution description
- Scope boundaries defined: check for in-scope / out-of-scope sections
- Success criteria measurable: check for quantitative language, thresholds, or verifiable outcomes
**For Generic — auto-check:**
- Body is non-empty and contains substantive content
- Issue is actionable: check for concrete next steps, clear requirements, or defined AC
Emit ✅ for pass, ⚠️ for partial/uncertain, ❌ for missing/fail. Include brief evidence for each finding.
**Step 3c: Ask Subjective Criteria**
For **subjective** criteria applicable to the current reviewMode, use `AskUserQuestion`:
- `description-sufficient`: "Does the description provide enough context for someone unfamiliar with this area?"
- `story-sizing` (team/enterprise): "Is this appropriately sized for a single iteration?"
- `risk-assessment` (enterprise): "Are technical and business risks adequately identified?"
If no subjective criteria apply (all filtered by reviewMode), skip the question step entirely.
**Step 3d: Extension Criteria** (if `--with` specified)
For each loaded extension domain, evaluate the issue against the extension's Issue Review Questions. Auto-evaluate objective; ask user about subjective.
**Step 3e: Collect All Findings**
Merge auto-evaluated and user-evaluated findings using check marks:
- `✅` — Criterion met
- `⚠️` — Needs attention
- `❌` — Missing or incorrect
Determine a recommendation:
- **Ready for work** — No blocking concerns
- **Needs minor revision** — Small issues, can start with awareness
- **Needs revision** — Should be addressed before starting work
- **Needs major rework** — Fundamental issues with issue definition
**If extensions were loaded (Step 2b):**
For each loaded extension, present findings as a separate section:
```markdown
### Security Review (IDPF-Security)
- [Finding 1]
- [Finding 2]
```
Extension findings can **escalate** the overall recommendation but cannot downgrade it.
**Applicability Filtering:** Omit extension domain sections that produce no applicable findings. Only domains with actual findings appear in the output and in the `**Extensions Applied:**` header. If no domains produce findings, fall back to standard-only review with a warning: `"No domain extensions produced findings. Showing standard review only."` At least one domain section must appear when `--with` is used; otherwise the fallback applies.
### Step 4: Update Issue Body Metadata
Export the current issue body:
```bash
gh pmu view $ISSUE --body-stdout > .tmp-$ISSUE.md
```
**Update `**Reviews:** N` field:**
- If `**Reviews:**` field exists: increment the number
- If `**Reviews:**` field does not exist: add `**Reviews:** 1` at the end of the issue body (after acceptance criteria if present)
Write the updated body:
```bash
gh pmu edit $ISSUE -F .tmp-$ISSUE.md && rm .tmp-$ISSUE.md
```
**If edit fails:** `"Failed to update issue body: {error}"` → **STOP**
### Step 5: Post Issue Comment
Post a structured review comment:
```markdown
## Issue Review #N — YYYY-MM-DD

**Issue:** #$ISSUE — $TITLE
**Type:** [Bug | Enhancement | Story | Epic | Generic]
**Total Reviews:** N
**Extensions Applied:** {list of applied extensions, or "None"}

### Findings

#### Auto-Evaluated
- ✅ [Criterion] — [evidence]
- ❌ [Criterion] — [what's missing]

#### User-Evaluated
- ✅ [Criterion] — [user assessment]
- ⚠️ [Criterion] — [user concern]

### Recommendation
**[Ready for work | Needs minor revision | Needs revision | Needs major rework]** — [Brief explanation]
```
**Backwards compatibility:** The `### Findings` section header and emoji markers (✅ ⚠️ ❌) remain unchanged for `/resolve-review` parser compatibility. The `#### Auto-Evaluated` and `#### User-Evaluated` subsections are additive.
```bash
gh issue comment $ISSUE -F .tmp-review-comment.md
rm .tmp-review-comment.md
```
**If comment post fails:** Warn and continue (non-blocking — body metadata already updated).

<!-- USER-EXTENSION-START: post-review -->
<!-- USER-EXTENSION-END: post-review -->

### Step 6: Report Summary
```
Review #N complete for Issue #$ISSUE: $TITLE
  Type: [Bug | Enhancement | Story | Epic | Generic]
  Recommendation: [recommendation]
  Reviews: N (updated)
  Issue comment: [posted | failed]
```
**For multiple issues:** Report summary for each, then batch summary at end.
**If `--with` is not specified**, append a discoverability tip after the summary:
```
Tip: Use --with security,performance to add domain-specific review criteria.
Available: security, accessibility, performance, chaos, contract, qa (or --with all)
```
---
## Error Handling
| Situation | Response |
|-----------|----------|
| Issue not found | "Issue #N not found." → STOP (skip to next if batch) |
| Issue closed | "Issue #N is closed. Review anyway? (y/n)" → ask user |
| Unknown issue type | "Using generic review criteria." → warn, continue |
| Issue body is empty | Flag as critical finding in review |
| `gh pmu edit` fails | "Failed to update issue body: {error}" → STOP |
| Comment post fails | Warn, continue (body already updated) |
| Multiple recognized labels | Use first recognized label as primary type |
---
**End of /review-issue Command**
