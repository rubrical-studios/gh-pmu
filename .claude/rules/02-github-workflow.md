# GitHub Workflow Integration
**Version:** v0.17.1
**Source:** Reference/GitHub-Workflow.md

---

**MUST READ:** At session startup and after compaction.

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

**Framework config (optional):** `framework: IDPF-Agile` enables workflow restrictions.
**Microsprint config:** `microsprint: stale_threshold_hours: 24`

## Prerequisites
```bash
gh extension install rubrical-studios/gh-pmu
```

## gh pmu Command Reference
**Issue Management:**
| Command | Replaces |
|---------|----------|
| `gh pmu create --title "..." --status backlog --assignee @me` | `gh issue create` + `gh pmu move` |
| `gh pmu move [#] --status [value]` | - |
| `gh pmu view [#]` | `gh issue view` |
| `gh pmu list --status [value]` | - |
| `gh pmu board` | - |

**Sub-Issue Management:**
| Command | Replaces |
|---------|----------|
| `gh pmu sub create --parent [#] --title "..."` | `gh issue create` + `gh pmu sub add` |
| `gh pmu sub add [parent] [child]` | - |
| `gh pmu sub list [#]` | - |
| `gh pmu split [#] --from=body` | Manual sub-issue creation |

**Bulk Operations:**
- `gh pmu move [#] --status done --recursive` - Update issue + all sub-issues
- `gh pmu triage --query "..." --apply status:backlog` - Bulk update
- `gh pmu intake --apply` - Add untracked issues

**Microsprint:** `start`, `current`, `add [#]`, `remove [#]`, `close`, `list`, `resolve`
**Release:** `start --branch release/vX.Y.Z`, `current`, `move [#] --release current` (recommended), `remove [#]`, `close [--tag]`, `list`
**Patch Releases:** Use `gh pmu release` with `patch/` branch naming (e.g., `--branch patch/v1.1.5`)

**Auto-Close:** Default Kanban template auto-closes issues when moved to `done`. `gh issue close` only needed for close reason or comment.

## Critical Rules
- **Issues close ONLY when user says "Done"** - Never close automatically, skip STOP checkpoint, or close because code shipped
- **Acceptance criteria must be checked** - All boxes checked before In Review or Done; evaluate criteria when moving to In Review
- **No auto-close keywords until Done** - Use `Refs #XXX` (not `Fixes/Closes/Resolves #XXX`) until user approves
- **All work on release branches** - Never push to main directly; work requires Release field; checkout release branch before working

### Commit Message Keywords
| Phase | Use | Avoid |
|-------|-----|-------|
| In Progress / In Review | `Refs #XXX`, `Part of #XXX` | `Fixes`, `Closes`, `Resolves` |
| After "Done" | `Fixes #XXX` | — |

## Framework Applicability
| Framework | Microsprint | Release | Patch |
|-----------|:-----------:|:-------:|:-----:|
| IDPF-Agile | Primary | Optional | - |
| IDPF-Structured | Optional | Primary | - |
| IDPF-LTS | - | - | Primary |
| IDPF-Vibe | Optional | - | - |

## Sprint-Release Binding
Each sprint scoped to one release; `microsprint start` requires active release; sprint issues must match release; work on release branch.

## Workflow Routing (CRITICAL)
**Step 1: Determine Framework** from `.gh-pmu.yml` or labels:
| Framework | Parent Label | Child Labels |
|-----------|--------------|--------------|
| IDPF-Agile | `epic` | `story` |
| IDPF-Structured | `requirement` | `implementation`, `qa-automation`, `qa-manual` |

**When user says "work #N":** `gh issue view [N] --repo {repository} --json labels --jq '.labels[].name'`

**IDPF-Agile:** `epic` label? → Yes: EPIC WORKFLOW (Section 4) | No: STANDARD (Section 1)
**IDPF-Structured:** `requirement` label? → Yes: REQUIREMENT WORKFLOW (Section 4-S) | No: STANDARD (Section 1)

**Section 4-S: Requirement Workflow**
1. `gh pmu move [req] --status in_progress`
2. `gh pmu sub list [req]` → Work sub-issues: implementation first, then qa-*
3. When all done → `gh pmu move [req] --status in_review`
4. **STOP** → `gh pmu move [req] --status done --recursive --yes`

**Trigger Words (Create Issue First):**
| Trigger | Section |
|---------|---------|
| `bug:`, `finding:` | 1 (Standard) |
| `enhancement:` | 1 (Standard) |
| `idea:` | 2 (Proposal alias) |
| `proposal:` | 2 (Proposal) |
| `prd:` | 7 (PRD) |

Create issue → Report number → **Wait for "work"**

## BLOCKING: Status Change Prerequisites
**Before `--status in_review`:**
1. `gh issue view [#] --json body -q '.body' > .tmp-issue-[#].md`
2. Review checkboxes, change `[ ]` to `[x]` for completed
3. `gh issue edit [#] --body-file .tmp-issue-[#].md`
4. `rm .tmp-issue-[#].md`
5. Verify: `gh issue view [#] --json body | grep -c "\[x\]"`
6. Now: `gh pmu move [#] --status in_review`

**Before `--status done`:**
1. `gh issue view [#] --json body | grep "\[ \]"`
2. If ANY unchecked boxes → DO NOT PROCEED
3. Now: `gh pmu move [#] --status done`

## Workflows
### 1. Standard Issue (Bug/Enhancement)
**Step 1 (AUTO):**
```bash
gh pmu create --repo {repository} --title "[Bug|Enhancement]: ..." --label [bug|enhancement] --body "..." --status backlog --priority p2 --assignee @me
```
**Step 2 (WAIT):** Wait for "work issue", "fix that", "implement that"
**Step 3:** `gh pmu move --status in_progress` → Work → Check criteria → `--status in_review`
**STOP:** Report and wait for "Done"
**Step 4:** `gh pmu move --status done` (auto-closes)

### 2. Proposal Workflow
**Step 1 (AUTO):** Create `Proposal/[Name].md` + issue via `gh pmu create --label proposal --assignee @me`
**Step 2 (WAIT):** Wait for "implement the proposal", "work issue"
**Step 3:** Implement → `git mv Proposal/[Name].md Proposal/Implemented/` → Check criteria → `--status in_review`
**STOP:** Report and wait for "Done"
**Step 4:** `gh pmu move --status done`

### 3. Sub-Issue Workflow
**Option A:** `gh pmu split [parent] --from=body` (from checklist)
**Option B:** `gh pmu sub create --parent [#] --title "..."`
Then ask: "Label parent as 'epic'? (yes/no)"
If yes: `gh issue edit [parent] --add-label "epic"`, add "story" to sub-issues

### 4. Epic Workflow
**CRITICAL:** Takes precedence when issue has "epic" label
**Detection:** `gh issue view [#] --json labels | grep -q "epic"`
**Step 0:** `gh pmu move [epic] --status in_progress`
**Step 1:** `gh pmu sub list [epic] --json` → Sort by number
**Step 2:** For each sub-issue: `--status in_progress` → Work → Check criteria → `--status in_review`
**Step 3:** Check epic criteria → `gh pmu move [epic] --status in_review`
**STOP:** Report and wait for "Done"
**Step 4:** `gh pmu move [epic] --status done --recursive --yes`

### 5. PRD to Issues
- **Agile:** `Create-Backlog` → Epics + Stories (see IDPF-Agile/Agile-Commands.md)
- **Structured:** `Create-Requirements` → REQ issues with Implementation + QA sub-issues

### 6. Reopen Workflow
`gh issue reopen [#]` → `gh pmu move [#] --status ready`

### 7. Proposal-to-PRD
1. Load IDPF-PRD framework + anti-hallucination rules
2. Run Discovery → Elicitation → Specification → Generation phases
3. Create `PRD/PRD-[Name].md`, update proposal status
4. Change label from "proposal" to "prd"

### 8. PRD Completion
1. Verify all linked issues are Done
2. Update PRD status to Complete
3. `git mv PRD/PRD-[Name].md PRD/Implemented/`

### 9. Microsprint Workflow
**Start:** `gh pmu microsprint start [--name "theme"]`
**Add:** `gh pmu microsprint add [#]` or `gh pmu move [#] --microsprint current`
**Close:** `gh pmu microsprint close [--skip-retro] [--commit]`
**Artifacts:** `Microsprints/[name]/review.md`, `retro.md`
**Team model:** One active microsprint shared by team. Join/Wait/Cancel prompt if another is active.
**Stale detection:** >24h old prompts Close/Abandon/Resume

### 10-11. Release/Patch Workflow
**Start:** `gh pmu release start --branch "release/v1.2.0"` (or `patch/v1.1.5` for patches)
**Add:** `gh pmu move [#] --release current`
**Close:** `gh pmu release close [--tag]`
**Artifacts:** `Releases/[release|patch]/vX.Y.Z/[release|patch]-notes.md`, `changelog.md`

### 12. PR-Only Main Merges
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
