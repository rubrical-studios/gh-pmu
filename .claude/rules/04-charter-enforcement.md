# Charter Enforcement
**Version:** v0.29.3
**Source:** Reference/Charter-Enforcement.md
**Purpose:** Define checkpoints for charter validation and scope conflict handling
---
## Overview
Charter enforcement validates work items align with project scope. Validation is **conversational, not blocking**. Users can expand scope, proceed anyway, or revise work.
---
## Checkpoint Detection
| Checkpoint | Trigger | What to Load |
|------------|---------|--------------|
| Proposal Creation | File in `Proposal/` | CHARTER.md, Inception/Scope-Boundaries.md |
| Proposal Update | Edit existing proposal | CHARTER.md, Inception/Scope-Boundaries.md |
| PRD Creation | `create-prd` or manual PRD | CHARTER.md, Inception/Scope-Boundaries.md |
| PRD Update | Edit existing PRD | CHARTER.md, Inception/Scope-Boundaries.md |
| Enhancement Issue | `enhancement:` trigger | CHARTER.md (summary only) |
| Add-Story Command | `Add-Story` or story issue | CHARTER.md (summary only) |
| Create-Backlog | `Create-Backlog` from PRD | CHARTER.md, Inception/Scope-Boundaries.md |
**Detection Logic:**
1. Check if CHARTER.md exists (if NO: skip validation)
2. Check if CHARTER.md is template (regex: `/{[a-z][a-z0-9-]*}/`) - if template: skip
3. Identify checkpoint type from action
4. Load appropriate charter artifacts
5. Proceed to validation
**Checkpoint Priority:**
| Priority | Checkpoint | Validation Depth |
|----------|------------|------------------|
| High | Create-Backlog, PRD creation | Full scope check |
| Medium | Proposal creation, Enhancement | Summary scope check |
| Low | Add-Story | Quick alignment check |
---
## Scope Validation
**Validation Process:**
1. Extract key concepts from work item (title keywords, problem statement, proposed solution, features)
2. Compare against charter scope (In Scope, Out of Scope, Current Focus)
3. Determine alignment: ALIGNED, POSSIBLY MISALIGNED, or CONFLICTS
| Category | Description | Action |
|----------|-------------|--------|
| Aligned | Work fits in-scope items | Proceed without interruption |
| Possibly Misaligned | Not in scope but not excluded | Ask for confirmation |
| Conflicts | Matches out-of-scope items | Flag conflict, offer resolution |
---
## Conversational Resolution
**Resolution Options:**
1. Expand charter scope to include this
2. Defer to future release
3. Proceed anyway (creates scope drift)
4. Revise the work item to fit current scope
| User Choice | Action |
|-------------|--------|
| Expand scope | Update Inception/Scope-Boundaries.md, sync CHARTER.md if needed |
| Defer | Note in work item: "Deferred - out of current scope" |
| Proceed anyway | Log warning, continue with work |
| Revise | Help user modify work item to fit scope |
**Expansion Flow:**
1. User provides description of scope item to add
2. Add to Inception/Scope-Boundaries.md "In Scope" with rationale
3. Update CHARTER.md if item is major
4. Commit: "Expand charter scope: add [item]"
5. Continue with original work
---
## /prepare-release Validation Gate
**Pre-Release Check:**
1. List all issues assigned to release
2. Check each against charter scope
3. Report: "X of Y issues aligned" and flagged misalignments
4. Ask: "Proceed with release? Or review flagged issues?"
| Finding | Severity | Release Blocked? |
|---------|----------|------------------|
| All aligned | Info | No |
| Some possibly misaligned | Warning | No (user decides) |
| Conflicts with out-of-scope | Error | No (user decides, but warned) |
---
## Token Budget
| Checkpoint Type | Artifacts Loaded | Estimated Tokens |
|-----------------|------------------|------------------|
| High priority | CHARTER.md + Scope-Boundaries.md | ~600-1,000 |
| Medium priority | CHARTER.md only | ~150-200 |
| Low priority | CHARTER.md (vision + scope sections) | ~100-150 |
---
## Configuration
**Opt-Out:**
- Per-session: "Skip charter validation for this session"
- Permanent: Create `.no-charter-enforcement` file
**Quiet Mode:** "Validate silently - only warn on conflicts" (skips "possibly misaligned" confirmations)
---
## Framework Exclusions
IDPF framework files are **automatically excluded** from charter scope validation.
**Excluded Patterns:**
| Pattern | Description |
|---------|-------------|
| `.claude/**` | Rules, hooks, scripts, commands |
| `IDPF-*/**` | Framework directories |
| `Overview/**`, `Assistant/**`, `Reference/**` | Framework documentation |
| `System-Instructions/**`, `Skills/**` | Specialists and skills |
| `Templates/**`, `.min-mirror/**` | Templates and minimized files |
| `framework-config.json`, `framework-manifest.json`, `.gh-pmu.yml` | Config files |
**Charter Scope Applies To:** `src/`, `lib/`, `app/`, `tests/`, project docs, Proposals, PRDs
---
**End of Charter Enforcement**
