# Charter Enforcement
**Version:** v0.33.3
**Source:** Reference/Charter-Enforcement.md
**Purpose:** Define checkpoints for charter validation and scope conflict resolution.
## Overview
Charter enforcement validates work items align with project scope. Validation is **conversational, not blocking**—users choose to expand scope, proceed anyway, or revise work.
## Checkpoint Detection
| Checkpoint | Trigger | What to Load |
|------------|---------|--------------|
| Proposal Creation | User creates file in `Proposal/` | CHARTER.md, Inception/Scope-Boundaries.md |
| Proposal Update | User edits existing proposal | CHARTER.md, Inception/Scope-Boundaries.md |
| PRD Creation | User runs `create-prd` | CHARTER.md, Inception/Scope-Boundaries.md |
| PRD Update | User edits existing PRD | CHARTER.md, Inception/Scope-Boundaries.md |
| Enhancement Issue | `enhancement:` trigger | CHARTER.md (summary only) |
| Add-Story Command | `Add-Story` or story issue | CHARTER.md (summary only) |
| Create-Backlog | Create-Backlog from PRD | CHARTER.md, Inception/Scope-Boundaries.md |
### Detection Logic
1. Check if CHARTER.md exists → If NO: Skip validation
2. Check if CHARTER.md is template (has `{placeholder}`) → If template: Skip
3. Identify checkpoint type from action
4. Load appropriate charter artifacts
5. Proceed to validation
### Checkpoint Priority
| Priority | Checkpoint | Depth |
|----------|------------|-------|
| High | Create-Backlog, PRD creation | Full scope check |
| Medium | Proposal creation, Enhancement | Summary scope check |
| Low | Add-Story | Quick alignment check |
## Scope Validation
### Validation Process
1. Extract key concepts from work item (title, problem, solution, features)
2. Compare against charter scope (In Scope, Out of Scope, Current Focus)
3. Determine alignment: ALIGNED, POSSIBLY MISALIGNED, CONFLICTS
### Alignment Categories
| Category | Description | Action |
|----------|-------------|--------|
| Aligned | Work fits in-scope items | Proceed without interruption |
| Possibly Misaligned | Work not explicitly in scope | Ask for confirmation |
| Conflicts | Work matches out-of-scope items | Flag conflict, offer resolution |
## Conversational Resolution
When misalignment detected:
```
"This [proposal/feature/story] involves [X], which isn't currently in scope.

Options:
1. Expand charter scope to include this
2. Defer to future release
3. Proceed anyway (creates scope drift)
4. Revise the work item to fit current scope"
```
### Resolution Actions
| Choice | Action |
|--------|--------|
| Expand scope | Update Inception/Scope-Boundaries.md |
| Defer | Note in work item: "Deferred - out of current scope" |
| Proceed anyway | Log warning, continue |
| Revise | Help modify work item to fit scope |
## /prepare-release Validation Gate
Pre-Release Check:
1. List all issues assigned to release
2. For each issue: Check against charter scope
3. Report: "X of Y issues aligned" + flagged items
4. Ask: "Proceed with release? Or review flagged issues?"
**Validation is advisory only. User always has final decision.**
## Configuration
### Opt-Out
- **Per-session:** "Skip charter validation for this session"
- **Permanent:** Create `.no-charter-enforcement` file
### Quiet Mode
"Validate silently - only warn on conflicts"
## Framework Exclusions
IDPF framework files are **automatically excluded** from charter scope validation.
### Excluded Patterns
`.claude/**`, `IDPF-*/**`, `Overview/**`, `Assistant/**`, `Reference/**`, `System-Instructions/**`, `Skills/**`, `Templates/**`, `.min-mirror/**`, `framework-config.json`, `framework-manifest.json`, `.gh-pmu.yml`
### Charter Scope Applies To
- Application source code
- Project tests
- Project configuration (non-framework)
- Project documentation
- Proposals and PRDs
---
**End of Charter Enforcement**
