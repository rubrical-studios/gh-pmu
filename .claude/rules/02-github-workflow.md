# GitHub Workflow Integration

**Version:** 1.0
**Source:** Reference/GitHub-Workflow.md

---

**Version:** 1.5
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
| `gh pmu create --title "..." --status backlog` | `gh issue create` + `gh pmu move` |
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
**Release:** `start --version X.Y.Z`, `current`, `add [#]`, `remove [#]`, `close [--tag]`, `list`
**Patch:** `start --version X.Y.Z`, `current`, `add [#]`, `remove [#]`, `close [--tag]`, `list`

**Auto-Close:** Default Kanban template auto-closes issues when moved to `done`. `gh issue close` only needed for close reason or comment.

## Critical Rules
- **NEVER close issues automatically** - Wait for "Done"
- **NEVER skip STOP checkpoint** - Report and wait
- **Issues stay open** until explicit approval
- **NEVER mark Done with unchecked boxes** - All acceptance criteria must be checked
- **In Review requires checkbox evaluation** - Check all criteria when moving to In Review
- **NEVER close issues because code shipped** - Release != approval. Each issue needs separate "Done".

## Framework Applicability
| Framework | Microsprint | Release | Patch |
|-----------|:-----------:|:-------:|:-----:|
| IDPF-Agile | Primary | Optional | - |
| IDPF-Structured | Optional | Primary | - |
| IDPF-LTS | - | - | Primary |
| IDPF-Vibe | Optional | - | - |

## Workflow Routing (CRITICAL)
**When user says "work #N":**
```bash
gh issue view [N] --repo {repository} --json labels --jq '.labels[].name'
```
```
Has "epic" label? ─── YES ──► EPIC WORKFLOW (Section 4)
         │
         NO ──► STANDARD ISSUE WORKFLOW (Section 1)
```

**Trigger Words (Create Issue First):**
| Trigger | Section |
|---------|---------|
| `bug:`, `finding:` | 1 (Standard) |
| `enhancement:` | 1 (Standard) |
| `idea:` | 7 (Idea) |
| `proposal:` | 2 (Proposal) |
| `prd:` | 8 (PRD) |

Create issue → Report number → **Wait for "work"**

## BLOCKING: Status Change Prerequisites
**Before `--status in_review`:**
1. `gh issue view [#] --json body -q '.body' > /tmp/issue-[#].md`
2. Review checkboxes, change `[ ]` to `[x]` for completed
3. `gh issue edit [#] --body-file /tmp/issue-[#].md`
4. Verify: `gh issue view [#] --json body | grep -c "\[x\]"`
5. Now: `gh pmu move [#] --status in_review`

**Before `--status done`:**
1. `gh issue view [#] --json body | grep "\[ \]"`
2. If ANY unchecked boxes → DO NOT PROCEED
3. Now: `gh pmu move [#] --status done`

## Workflows
### 1. Standard Issue (Bug/Enhancement)
**Step 1 (AUTO):**
```bash
gh pmu create --repo {repository} --title "[Bug|Enhancement]: ..." --label [bug|enhancement] --body "..." --status backlog --priority p2
```
**Step 2 (WAIT):** Wait for "work issue", "fix that", "implement that"
**Step 3:** `gh pmu move --status in_progress` → Work → Check criteria → `--status in_review`
**STOP:** Report and wait for "Done"
**Step 4:** `gh pmu move --status done` (auto-closes)

### 2. Proposal Workflow
**Step 1 (AUTO):** Create `Proposal/[Name].md` + issue via `gh pmu create --label proposal`
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

### 5. Create-Issues (PRD)
- `PRD-Agile-*.md` → Use `Create-Backlog` (IDPF-Agile/Agile-Commands.md)
- `PRD-Structured-*.md` → Create REQ issues with Implementation + QA sub-issues

### 6. Reopen Workflow
`gh issue reopen [#]` → `gh pmu move [#] --status ready`

### 7. Idea Workflow
**Step 1:** Create `Proposal/[Idea-Name].md` (minimal) + issue via `gh pmu create --label idea`
**Step 2:** Iterate conversationally
**Step 3:** "convert to PRD" → Section 8

### 8. Proposal-to-PRD
1. Load IDPF-PRD framework + anti-hallucination rules
2. Run Discovery → Elicitation → Specification → Generation phases
3. Create `PRD/PRD-[Name].md`, update proposal status
4. Change label from "idea" to "prd"

### 9. PRD Completion
1. Verify all linked issues are Done
2. Update PRD status to Complete
3. `git mv PRD/PRD-[Name].md PRD/Implemented/`

### 10. Microsprint Workflow
**Start:** `gh pmu microsprint start [--name "theme"]`
**Add:** `gh pmu microsprint add [#]` or `gh pmu move [#] --microsprint current`
**Close:** `gh pmu microsprint close [--skip-retro] [--commit]`
**Artifacts:** `Microsprints/[name]/review.md`, `retro.md`
**Team model:** One active microsprint shared by team. Join/Wait/Cancel prompt if another is active.
**Stale detection:** >24h old prompts Close/Abandon/Resume

### 11. Release Workflow
**Start:** `gh pmu release start --version "1.2.0" [--name "Phoenix"]`
**Add:** `gh pmu release add [#]`
**Close:** `gh pmu release close [--tag]`
**Artifacts:** `Releases/v1.2.0/release-notes.md`, `changelog.md`

### 12. Patch Workflow
**Start:** `gh pmu patch start --version "1.1.5"`
**Add:** `gh pmu patch add [#]` (warns if not bug/hotfix, errors if breaking-change)
**Close:** `gh pmu patch close [--tag]`
**Artifacts:** `Patches/v1.1.5/patch-notes.md`

## Visibility Commands
```bash
gh pmu board                    # Kanban view
gh pmu board --status in_progress
gh pmu list --status in_review
gh pmu triage --query "..." --apply status:backlog
gh pmu intake --apply "status:backlog,priority:p2"
```

## Shell Limitations
**Heredocs with backticks fail.** Use `--body-file` with temp file instead.
**Command substitution fails.** Run commands separately, use literal values.

## CI/CD Rate Limiting
See **ci-cd-pipeline-design** skill for GitHub API best practices: rate limits, auth strategies (PATs, GitHub Apps), exponential backoff, workflow cascade prevention.
Reference: `Skills/ci-cd-pipeline-design/SKILL.md` → "GitHub API Best Practices"

## Manual Overrides
- "don't create an issue" → Skip issue creation
- "label this as [label]" → Use specified label
- "keep the issue open" → Don't close

**End of GitHub Workflow Integration**
