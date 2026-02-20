---
version: "v0.46.1"
description: Review issues with type-specific criteria (project)
argument-hint: "#issue [#issue...]"
---

<!-- EXTENSIBLE -->
# /review-issue
Reviews one or more GitHub issues with type-specific criteria based on labels (bug, enhancement, story, epic). Tracks review count in issue body and posts structured findings as comment.
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
| `--mode` | No | Transient review mode override: `solo`, `team`, or `enterprise`. Does not modify `framework-config.json`. |
Accepts multiple issue numbers: `/review-issue #42 #43 #44` -- reviews each sequentially.
---
## Execution Instructions
**REQUIRED:** Before executing:
1. **Generate Todo List:** Parse workflow steps, use `TodoWrite` to create todos
2. **Include Extensions:** Add todo item for each non-empty `USER-EXTENSION` block
3. **Track Progress:** Mark todos `in_progress` -> `completed` as you work
4. **Post-Compaction:** Re-read spec and regenerate todos after context compaction
**Todo Rules:** One todo per numbered step; one todo per active extension; skip commented-out extensions.
---
## Workflow
**For multiple issues:** Process each issue sequentially through Steps 1-6.
### Step 1: Fetch Issue
```bash
gh issue view $ISSUE --json number,title,body,state,labels
```
**If not found:** `"Issue #$ISSUE not found."` -> **STOP** (skip to next if batch)
**If closed:** `"Issue #$ISSUE is closed. Review anyway? (y/n)"` -- proceed only if user confirms.
### Step 2: Determine Issue Type
Use the shared `getIssueType()` utility from `.claude/scripts/shared/lib/issue-type.js` to determine issue type from labels already fetched in Step 1 (no additional API call needed):
```javascript
const { getIssueType } = require('./.claude/scripts/shared/lib/issue-type.js');
const { type, redirect } = getIssueType(issueData); // issueData from Step 1
```
| Label Found | Review Type |
|-------------|-------------|
| `bug` | Bug review criteria |
| `enhancement` | Enhancement review criteria |
| `story` | Story review criteria |
| `epic` | Epic review criteria |
| `test-plan` | **Redirect** -> `/review-test-plan #$ISSUE` |
| `proposal` | **Redirect** -> `/review-proposal #$ISSUE` |
| `prd` | **Redirect** -> `/review-prd #$ISSUE` |
| None recognized | Generic review criteria |
**If redirect detected:** Invoke appropriate command using Skill tool. -> **STOP**
**If multiple recognized labels:** Use first recognized. Redirect labels take precedence.
**If no recognized labels:** Use generic criteria (non-blocking).

<!-- USER-EXTENSION-START: pre-review -->
<!-- USER-EXTENSION-END: pre-review -->

### Step 2b: Extension Loading
**If `--with` is specified:**
1. Read `.claude/metadata/review-extensions.json`
2. Parse `--with` value: `--with all` loads all; `--with security,performance` loads specified (trim spaces)
3. For each extension ID: look up `source` path, read criteria file, extract **Issue Review Questions** section
4. Unknown extension: warn with available list
5. Store loaded criteria for Step 3
**Error Handling:** Registry not found/malformed -> fall back to standard review. Criteria file not found -> skip domain. All missing -> standard review only.
**If `--with` is not specified:** Skip extension loading.
### Step 3: Perform Review
Evaluate using two-phase approach: **auto-evaluate objective criteria**, then **ask user about subjective criteria** via `AskUserQuestion`. Filter by `reviewMode` from `framework-config.json` (or `--mode` override).
**Step 3a: Load reviewMode Configuration**
Parse `--mode` from arguments if provided. Invalid values produce a clear error with valid options.
```javascript
const { getReviewMode, shouldEvaluate, isObjective } = require('./.claude/scripts/shared/lib/review-mode.js');
// modeOverride is the --mode argument value (null if not provided)
const mode = getReviewMode(process.cwd(), modeOverride); // 'solo', 'team', or 'enterprise'
```
**Hint:** Display the active mode and how to override it:
- Without `--mode`: `Reviewing in {mode} mode (override with --mode solo|team|enterprise)`
- With `--mode`: `Reviewing in {mode} mode (--mode override)`

<!-- USER-EXTENSION-START: criteria-customize -->
<!-- USER-EXTENSION-END: criteria-customize -->

**Step 3b: Auto-Evaluate Objective Criteria**
For each **objective** criterion applicable to the current reviewMode, evaluate autonomously. Do NOT ask the user.
**Common Objective Criteria:**
| Criterion | Auto-Check Method |
|-----------|-------------------|
| `title-clear` | Verify title is specific (>3 words), contains actionable context, not generic like "Fix bug" or "Update code" |
| `ac-present-testable` | Check for `- [ ]` checkbox items; verify each contains measurable/verifiable statement |
| `labels-correct` | Verify issue has at least one type label (bug/enhancement/story/epic) matching its content |
| `priority-set` | Check via `gh pmu view $ISSUE --json=priority` |
| `dependencies-identified` | Check for `Refs #`, `Depends on #`, `Blocks #`, or `**Refs:**` patterns |
| `effort-estimate` | Check for estimate/effort section or field in body |
| `test-coverage-proportionate` | Read issue body and linked resources. Classify change type. Scan ACs for test keywords. Compare scope against test ACs -- report proportionality with evidence. |
**Type-Specific Objective Checks:**
**For Bug -- auto-check:**
- Reproduction steps present: look for numbered list or "Steps to Reproduce" section
- Expected vs actual behavior documented: look for "Expected" and "Actual" sections
- Environment info present: look for "Environment", browser, OS, version references
- Severity proportionate to impact: compare severity/priority to described user impact
- Test coverage proportionate: single-path fix -> regression test preferred; multi-path refactor -> ACs MUST require tests
**For Story -- auto-check:**
- User story format: "As a ... I want ... So that ..."
- Acceptance criteria present: `- [ ]` items
- AC structured as testable assertions (TDD-compatible)
- Scope supports incremental implementation
- Canonical story body template: compare body structure, flag for repair if not matching
- Test coverage proportionate to scope
**For Epic -- auto-check:**
- Sub-issues exist: `gh pmu sub list $ISSUE`
- Scope description present
- Success criteria defined
- Dependencies between sub-issues identified
- **MUST: Iterate over each sub-issue and review** recursively
**For Enhancement -- auto-check:**
- Motivation/rationale present
- Proposed solution described
- Scope boundaries defined
- Success criteria measurable
- Test coverage proportionate: new feature -> unit tests required; behavior change -> updated tests required
**For Generic -- auto-check:**
- Body is non-empty and substantive
- Issue is actionable with concrete next steps
Emit ✅ for pass, ⚠️ for partial/uncertain, ❌ for missing/fail. Include brief evidence.
**Step 3c: Ask Subjective Criteria**
For **subjective** criteria applicable to the current reviewMode, use `AskUserQuestion`.
**Common Subjective Criteria** (filtered by reviewMode):
```javascript
const subjectiveCriteria = [
  shouldEvaluate('description-sufficient') && { q: "Does the description provide enough context for someone unfamiliar with this area?", h: "Description" },
  shouldEvaluate('story-sizing') && { q: "Is this appropriately sized for a single iteration?", h: "Sizing" },
  shouldEvaluate('risk-assessment') && { q: "Are technical and business risks adequately identified?", h: "Risks" }
].filter(Boolean);
```
**Solo mode:** `description-sufficient` is excluded. Solo mode has NO subjective criteria -- skip entirely.
**Team/enterprise mode -- description preview:** Emit 3-5 line summary before `AskUserQuestion`.
If no subjective criteria apply, skip the question step entirely.
**Step 3d: Extension Criteria** (if `--with` specified)
For each loaded extension domain, evaluate against Issue Review Questions. Auto-evaluate objective; ask about subjective.
**Step 3e: Collect All Findings**
Merge findings using: ✅ met, ⚠️ needs attention, ❌ missing/incorrect.
Determine recommendation:
- **Ready for work** -- No blocking concerns
- **Needs minor revision** -- Small issues, can start with awareness
- **Needs revision** -- Should be addressed before starting work
- **Needs major rework** -- Fundamental issues
**If extensions loaded:** Present as separate sections. Extensions can **escalate** but not downgrade.
**Applicability Filtering:** Omit domains with no findings. At least one domain section must appear when `--with` used; otherwise fallback.
### Step 4: Update Issue Body Metadata
```bash
gh pmu view $ISSUE --body-stdout > .tmp-$ISSUE.md
```
**Update `**Reviews:** N` field:** Increment if exists, add `**Reviews:** 1` if not.
```bash
gh pmu edit $ISSUE -F .tmp-$ISSUE.md && rm .tmp-$ISSUE.md
```
**If edit fails:** -> **STOP**
### Step 5: Post Issue Comment
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
**Backwards compatibility:** `### Findings` header and emoji markers unchanged for `/resolve-review` parser. `#### Auto-Evaluated` and `#### User-Evaluated` are additive.
```bash
gh issue comment $ISSUE -F .tmp-review-comment.md
rm .tmp-review-comment.md
```
**If comment post fails:** Warn and continue (non-blocking).
### Step 5.5: Assign Reviewed Label (Conditional)
If recommendation starts with "Ready for":
```bash
gh issue edit $ISSUE --add-label=reviewed
```
**Epic sub-issue label propagation:** If issue type is `epic`, also apply `reviewed` label to all sub-issues:
1. Retrieve sub-issues: `gh pmu sub list $ISSUE`
2. For each sub-issue: `gh issue edit $SUB_ISSUE --add-label=reviewed`
3. Track count for Step 6 reporting
If not an epic, only the issue itself is labeled.
If not "Ready for": skip.

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
**For epic issues with "Ready for" recommendation:** Include sub-issue reviewed label count:
```
  Sub-issues labeled 'reviewed': M
```
**For multiple issues:** Report summary for each, then batch summary at end.
**If `--with` is not specified**, append a discoverability tip:
```
Tip: Use --with security,performance to add domain-specific review criteria.
Available: security, accessibility, performance, chaos, contract, qa (or --with all)
```
---
## Error Handling
| Situation | Response |
|-----------|----------|
| Issue not found | "Issue #N not found." -> STOP (skip to next if batch) |
| Issue closed | "Issue #N is closed. Review anyway? (y/n)" -> ask user |
| Unknown issue type | "Using generic review criteria." -> warn, continue |
| Issue body is empty | Flag as critical finding in review |
| `gh pmu edit` fails | "Failed to update issue body: {error}" -> STOP |
| Comment post fails | Warn, continue (body already updated) |
| Multiple recognized labels | Use first recognized label as primary type |
---
**End of /review-issue Command**
