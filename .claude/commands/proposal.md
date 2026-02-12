---
version: "v0.42.0"
description: Create a proposal document and tracking issue (project)
argument-hint: "<title>"
---

<!-- EXTENSIBLE -->
# /proposal
Creates a proposal document (`Proposal/[Name].md`) and a tracking issue with the `proposal` label. Also triggered by the `idea:` alias.
## Available Extension Points
| Point | Location | Purpose |
|-------|----------|---------|
| `pre-create` | Before proposal creation | Duplicate detection, template customization |
| `post-create` | After proposal and issue created | Notifications, auto-assignment to branch |
---
## Prerequisites
- `gh pmu` extension installed
- `.gh-pmu.yml` configured in repository root
---
## Arguments
| Argument | Required | Description |
|----------|----------|-------------|
| `<title>` | No | Proposal title (e.g., `Dark Mode Support`) |
If no title provided, prompt the user for one.
**Alias:** `idea:` is treated identically to `proposal:` — same workflow, same output.
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
### Step 1: Parse Arguments
Extract `<title>` from command arguments.
**If empty:** Ask the user for a proposal title.
**If title contains special characters** (backticks, quotes): Escape for shell. On Windows, use temp file approach.
**Name conversion:** Convert title to file name: replace spaces with hyphens, Title-Case each word. Example: `dark mode support` → `Dark-Mode-Support`
### Step 2: Check for Existing Proposal
Check if `Proposal/[Name].md` already exists.
**If file exists:** Ask `Proposal/[Name].md already exists. Overwrite? (yes/no)`. If no → STOP.
### Step 3: Gather Description
Ask the user:
```
Briefly describe the proposal (problem and proposed solution):
```
**If user provides description:** Use it to populate the proposal template.
**If user declines or says "skip":** Create with placeholder sections.

<!-- USER-EXTENSION-START: pre-create -->
<!-- USER-EXTENSION-END: pre-create -->

### Step 4: Create Proposal Document
Ensure `Proposal/` directory exists (create if missing).
Create `Proposal/[Name].md` with standard template:
```markdown
# Proposal: [Title]
**Status:** Draft
**Created:** [YYYY-MM-DD]
**Author:** AI Assistant
**Tracking Issue:** (will be updated after issue creation)
---
## Problem Statement
[Problem description or "To be documented"]
## Proposed Solution
[Solution description or "To be documented"]
## Implementation Criteria
- [ ] [Criterion 1]
- [ ] [Criterion 2]
## Alternatives Considered
- [Alternative 1]: [Why not chosen]
## Impact Assessment
- **Scope:** [Files/components affected]
- **Risk:** [Low/Medium/High]
- **Effort:** [Estimate]
```
### Step 5: Create Tracking Issue
Build the issue body:
```markdown
## Proposal: [Title]
**File:** Proposal/[Name].md
### Summary
[Brief description from Step 3]
### Lifecycle
- [ ] Proposal reviewed
- [ ] Ready for PRD conversion
```
**Critical:** The issue body MUST include `**File:** Proposal/[Name].md` — required for `/create-prd` integration.
Create the issue:
```bash
gh pmu create --title "Proposal: {title}" --label proposal --status backlog --priority p2 --assignee @me -F .tmp-body.md
rm .tmp-body.md
```
**Note:** Always use `-F .tmp-body.md` for the body (never inline `--body`).
### Step 6: Update Proposal with Issue Reference
After issue creation, update the proposal document's tracking issue field:
```
**Tracking Issue:** #[issue-number]
```
### Step 7: Report and STOP
```
Created:
  Document: Proposal/[Name].md
  Issue: #$ISSUE_NUM — Proposal: {title}
  Status: Backlog
  Label: proposal

Say "/review-proposal #$ISSUE_NUM" or "/create-prd #$ISSUE_NUM", if ready
```

<!-- USER-EXTENSION-START: post-create -->
<!-- USER-EXTENSION-END: post-create -->

**STOP.** Do not begin work unless the user explicitly says "work", "implement the proposal", or "work issue".
---
## Error Handling
| Situation | Response |
|-----------|----------|
| No title provided | Prompt user for title |
| Empty title after prompt | "A proposal title is required." → STOP |
| Existing file, user declines overwrite | STOP without creating anything |
| `Proposal/` directory missing | Create it silently |
| `gh pmu create` fails | "Failed to create issue: {error}" → STOP |
| Special characters in title | Escape for shell safety |
---
**End of /proposal Command**
