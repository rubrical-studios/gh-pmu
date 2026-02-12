# GitHub Workflow Integration
**Version:** v0.42.0
**Source:** Reference/GitHub-Workflow.md
---
**MUST READ:** At session startup and after compaction.
**Source:** Reference/GitHub-Workflow.md
## Project Configuration
**Read from `.gh-pmu.yml`:**
```yaml
project:
    owner: {owner}
    number: {number}
repositories:
    - {owner}/{repo}
fields:
    status:
        values: {backlog, in_progress, in_review, done}
    priority:
        values: {p0, p1, p2}
```
Use alias (left side) in commands: `gh pmu move 90 --status in_progress`
**If missing:** Run `gh pmu init`
**After project creation:** Set default repository in project settings (Settings → Default repository). Without this, issues may go to wrong repo.
**Framework config (optional):** `framework: IDPF-Agile` enables workflow restrictions.
## Prerequisites
```bash
gh extension install rubrical-studios/gh-pmu
```
## Branch Semantics
**Key Principle:** Any branch (except `main`) can produce a release.
Branch naming requires `{prefix}/{name}` format. Prefix is organizational convention, not functional:
| Prefix | Convention | Can Release? |
|--------|------------|--------------|
| `release/` | Version releases | Yes |
| `patch/` | Hotfixes | Yes |
| `idpf/` | Framework dev | Yes |
| `feature/` | Feature work | Yes |
| `hotfix/` | Urgent fixes | Yes |
**Invalid:** Branch names without prefix (e.g., `my-branch`) fail validation.
**`main` branch:** Protected (no direct pushes), PR-only, tag source after merge.
**Working branch:** Any non-main branch. Term "release branch" avoided—all working branches can produce releases.
## gh pmu Command Reference
**Issue Management:**
| Command | Replaces |
|---------|----------|
| `gh pmu create --title "..." [-F body.md] --status backlog --assignee @me` | `gh issue create` + `gh pmu move` |
| `gh pmu move [#...] --status [value]` | - |
| `gh pmu view [#] [-b] [-c] [-w]` | `gh issue view` |
| `gh pmu edit [#] [-F body.md] [--body-stdin]` | `gh issue edit` |
| `gh pmu list --status [value]` | - |
| `gh pmu board` | - |
**View/Edit flags:** `-b` exports body to file; `-c` shows comments; `-w` opens in browser; `-F` reads body from file; `--body-stdin` reads from stdin; `--body-stdout` outputs to stdout; `--json=fields` outputs JSON (use `=` on Windows)
**JSON output (use `=` syntax on Windows):**
```bash
gh pmu view 123 --json=status --jq='.status'
gh pmu view 123 --json=number,title,status
```
**Sub-Issue Management:**
| Command | Replaces |
|---------|----------|
| `gh pmu sub create --parent [#] --title "..."` | `gh issue create` + `gh pmu sub add` |
| `gh pmu sub add [parent] [child]` | - |
| `gh pmu sub list [#]` | - |
| `gh pmu split [#] --from=body` | Manual sub-issue creation |
**Bulk Operations:**
- `gh pmu move [#] [#] [#] --status done` - Update multiple issues at once
- `gh pmu move [#] --status done --recursive` - Update issue + all sub-issues
- `gh pmu move [#] --recursive --dry-run` - Preview recursive changes
- `gh pmu triage --query "..." --apply status:backlog` - Bulk update
- `gh pmu intake --apply` - Add untracked issues
**Move Flags:** `--status`, `--branch` (replaces `--release`), `--backlog` (clear fields), `--recursive`, `--dry-run`, `--depth N`, `-f/--force`, `--yes`
**Multi-Issue:** `gh pmu move 42 43 44 --status in_progress` - more efficient than parallel calls
**Deprecation:** `--release` flag deprecated, use `--branch` instead.
**Branch:** `start --name release/vX.Y.Z`, `current`, `reopen [name]`, `move [#] --branch current` (recommended), `remove [#]`, `close [--tag]`, `list`
**Deprecation:** `gh pmu release` deprecated, use `gh pmu branch` instead.
**Patch Releases:** Use `gh pmu branch` with `patch/` branch naming (e.g., `--name patch/v1.1.5`)
**Auto-Close:** Default Kanban template auto-closes issues when moved to `done`. `gh issue close` only needed for close reason or comment.
## Slash Command Preference
Prefer slash commands over raw `gh pmu` commands:
| Instead of | Use |
|------------|-----|
| `bug:` inline issue creation | `/bug <title>` |
| `enhancement:` inline issue creation | `/enhancement <title>` |
| `proposal:` / `idea:` inline creation | `/proposal <title>` |
| `work #N` inline routing | `/work #N` |
| `done` / `gh pmu move --status done` (in_review → done) | `/done [#N]` |
| `gh pmu branch start` | `/create-branch` |
| `gh pmu branch list` | `/switch-branch` |
| `gh pmu branch close` (releases) | `/prepare-release` |
| `gh pmu branch close` (features) | `/merge-branch` |
| `gh pmu move [#] --branch` | `/assign-branch` |
| Manual issue review | `/review-issue #N` |
| Manual proposal review | `/review-proposal #N` |
| Manual PRD review | `/review-prd #N` |
**Use raw commands for:** debugging, uncovered operations, user request, complex bulk ops.
## Critical Rules
- **Issues close ONLY when user says "Done"** - Never close automatically, skip STOP checkpoint, or close because code shipped
- **Acceptance criteria must be checked** - All boxes checked before In Review or Done; evaluate criteria when moving to In Review
- **No auto-close keywords until Done** - Use `Refs #XXX` (not `Fixes/Closes/Resolves #XXX`) until user approves
- **All work on working branches** - Never push to main directly; work requires branch tracker; checkout working branch before working (see Branch Semantics)
- **Work requires explicit trigger** - After "evaluate" or "assess" commands, STOP after analysis. Never implement until user says "work", "fix that", or "implement that". Clarifying questions ≠ work permission.
### Analysis vs Work (HARD STOP)
**Analysis keywords:** evaluate, analyze, assess, investigate, check, verify
**Note:** `review` is a **tracked action** that routes to `/review-issue` — NOT an analysis keyword.
**When these appear with issue reference (#N):**
1. Report findings only
2. **STOP** - Do not implement
3. Use **read-only** commands (queries, not mutations)
4. End with: "Analysis complete. Say 'work' to implement."
**Forbidden during analysis:**
- Running scripts with side effects
- Creating/editing/moving issues (except reading)
- Making code changes
- Any `gh pmu move`, `git commit`, `node script.js` that modifies state
### Commit Message Keywords
| Phase | Use | Avoid |
|-------|-----|-------|
| In Progress / In Review | `Refs #XXX`, `Part of #XXX` | `Fixes`, `Closes`, `Resolves` |
| After "Done" | `Fixes #XXX` | — |
## Framework Applicability
| Framework | Release | Patch |
|-----------|:-------:|:-----:|
| IDPF-Agile | ● | ● |
| IDPF-Vibe | — | — |
## Workflow Routing (CRITICAL)
**Step 1: Determine Framework** from `.gh-pmu.yml` or labels:
| Framework | Parent Label | Child Labels |
|-----------|--------------|--------------|
| IDPF-Agile | `epic` | `story` |
**When user says "work #N":** The `/work` command handles validation, branch checks, issue type detection, auto-todo extraction, PRD tracker auto-move, and framework dispatch. See `/work` command spec.
**Trigger Words (Create Issue First):**
| Trigger | Command | Description |
|---------|---------|-------------|
| `bug:` | `/bug` | Create bug issue with standard template |
| `enhancement:` | `/enhancement` | Create enhancement issue with standard template |
| `idea:` | `/proposal` | Alias for proposal |
| `proposal:` | `/proposal` | Create proposal document + tracking issue |
Each command creates the issue, reports the number, and STOPs. Do NOT proceed to implementation until user says "work".
**Review Command Routing:** `review #N` → `/review-issue N`. `review #42 #43` → `/review-issue 42 43`. Without issue number, no routing.
**Review Commands:**
| Command | Target | Criteria |
|---------|--------|----------|
| `/review-issue #N` | Issue body | Type-specific: bug, enhancement, story, epic, generic |
| `/review-proposal #N` | Proposal file from issue | Completeness, consistency, feasibility, quality |
| `/review-prd #N` | PRD file from issue | Requirements, user stories, AC, NFRs, test plan |
**Review Metadata:** `**Reviews:** N` field — added on first review, incremented on subsequent.
**Review Log:** `/review-proposal` and `/review-prd` maintain append-only table:
```markdown
## Review Log
| # | Date | Reviewer | Findings Summary |
|---|------|----------|------------------|
| 1 | YYYY-MM-DD | Claude | [Summary] |
```
Rows are append-only — never edit or delete existing rows.
## Workflows
### 1. Standard Issue (Bug/Enhancement)
Create via `/bug` or `/enhancement` (or triggers) → STOPs. `/work` handles validation, TDD, AC, in_review. `/done` closes.
See `/bug`, `/enhancement`, `/work`, `/done` command specs.
### 2. Proposal Workflow
Create via `/proposal` (or triggers) → STOPs. Work/done follow standard flow.
**Required Body Format:** Issue body MUST include `**File:** Proposal/[Name].md` for `/create-prd` integration.
**Criteria Placement:** Issue = lifecycle only; File = implementation details; PRD = user stories
See `/proposal`, `/work`, `/done` command specs.
### 3. Sub-Issue Workflow
**Option A:** `gh pmu split [parent] --from=body` (from checklist)
**Option B:** `gh pmu sub create --parent [#] --title "..."`
Then ask: "Label parent as 'epic'? (yes/no)"
If yes: `gh issue edit [parent] --add-label "epic"`, add "story" to sub-issues
### 4. Epic Workflow
**CRITICAL:** Takes precedence when issue has "epic" label. Detection mandatory — check labels before routing.
Epic flow: `/work` detects epic → in_progress → sub-issues sequentially (each: in_progress → work → AC → in_review → **STOP per sub-issue** → "done" → next). All done → epic AC → in_review → **STOP** → "done" → recursive close.
**Never skip per-sub-issue STOP boundary.**
See `/work` (Steps 4, 7, 12) and `/done` command specs.
### 5. PRD to Issues
`/create-backlog` → Epics + Stories from approved PRD. See command spec.
### 6. Reopen Workflow
`gh issue reopen [#]` → `gh pmu move [#] --status ready`
### 6.5. Branch Reopen Workflow
`gh pmu branch reopen [branch-name]` - Reopen closed branch tracker (e.g., `release/v1.2.0`, `patch/v1.1.5`)
### 7. Proposal-to-PRD
`/create-prd` runs IDPF-PRD phases → generates `PRD/PRD-[Name].md`, updates proposal status, changes label to `prd`. See command spec.
### 8. PRD Completion
`/complete-prd` verifies stories done → updates status → moves to `PRD/Implemented/`. See command spec.
### 9. Release Workflow
**Applicable:** IDPF-Agile
`/create-branch` (e.g., `release/v1.2.0`) → `gh pmu move [#] --branch current` → `/prepare-release` for PR, merge, tag, artifacts.
See `/create-branch` and `/prepare-release` command specs.
### 10. Patch Workflow
**Applicable:** IDPF-Agile
Same as Release (Section 9) with `patch/` prefix. `/create-branch` with `patch/vX.Y.Z` → `/prepare-release`.
### 11. PR-Only Main Merges
All work via PRs to main. Never push directly.
1. `gh pr create --base main --head release/vX.Y.Z`
2. Wait for review/approval
3. Merge via PR
**Blocked:** `git push origin main`, direct commits to main
**Allowed:** Push to release/patch branches, create/merge PRs
## Visibility Commands
`gh pmu board` (Kanban), `gh pmu list --status [value]`, `gh pmu triage --query "..." --apply status:backlog`, `gh pmu intake --apply`
## CI/CD Rate Limiting
See **ci-cd-pipeline-design** skill for GitHub API best practices: rate limits, auth strategies (PATs, GitHub Apps), exponential backoff, workflow cascade prevention.
Reference: `Skills/ci-cd-pipeline-design/SKILL.md` → "GitHub API Best Practices"
## Manual Overrides
- "don't create an issue" → Skip issue creation
- "label this as [label]" → Use specified label
- "keep the issue open" → Don't close
**End of GitHub Workflow Integration**
