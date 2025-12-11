# GitHub Workflow Integration

**Version:** 1.0
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

## Prerequisites
```bash
gh extension install rubrical-studios/gh-pmu
```

## Critical Rules
⚠️ **NEVER close issues automatically** - Wait for "Done"
⚠️ **NEVER skip STOP checkpoint** - Report and wait
⚠️ **Issues stay open** until explicit approval

## Workflows

### 1. Bug Workflow
**Triggers:** "I found...", "bug", "broken", "finding:", unexpected behavior
**Step 1 (AUTO):** Create issue with `--label "bug"`
**Step 2 (WAIT):** Wait for "work issue", "fix that"
**Step 3:** `gh pmu move --status in_progress` → Work → Commit → `--status in_review`
**STOP:** Report and wait for "Done"
**Step 4:** `gh pmu move --status done` → `gh issue close`

### 2. Enhancement Workflow
**Triggers:** "I would like...", "Can you add...", "New feature..."
Same flow as Bug, use `--label "enhancement"`

### 3. Proposal Workflow
**Triggers:** "Create a proposal for...", "Design document..."
**Step 1 (AUTO):** Create `Proposal/[Name].md` + issue
**Step 2 (WAIT):** Wait for "implement the proposal"
**Step 3:** Implement → `git mv Proposal/[Name].md Proposal/Implemented/` → Commit
**STOP:** Wait for "Done"

### 4. Sub-Issue Workflow
**Triggers:** "Create sub-issues for...", "Break into phases..."
**Step 1:** Create sub-issues with parent reference
**Step 2:** Link: `gh pmu sub add [parent] [sub] --repo {repo}`
**Step 3:** Ask: "Should I label parent as 'epic'?"
**Step 4 (if yes):** Add `--add-label "epic"`, `--remove-label "story"` from parent, add "story" to subs

### 5. Create-Issues Workflow (PRD)
**Triggers:** "Create-Issues", "Create-Issues-Agile", "Create-Issues-Structured"
**Framework Detection:**
- `PRD-Agile-*.md` → Agile
- `PRD-Structured-*.md` → Structured
- `backlog/` → Agile
- Ambiguous → Ask user

**Agile:** Feature Area → Epic → Capabilities → Stories (linked as sub-issues)
**Structured:** REQ-XXX → Requirement → Implementation + QA sub-issues (linked)

### 6. Reopen Workflow
**Triggers:** "reopen issue #[N]"
`gh issue reopen [N]` → `gh pmu move --status ready`

### 7. Idea Workflow
**Triggers:** "I have an idea...", "What if...", "New idea:"
**Step 1 (AUTO):** Create lightweight `Proposal/[Idea].md` + issue with `--label "idea"`
**Step 2:** Iterate conversationally
**Step 3:** Wait for "convert to PRD"

### 8. Proposal-to-PRD Workflow
**Triggers:** "convert to PRD", "ready for PRD", "make PRD", `prd: [name]`, `prd: for #N`
**Step 1:** Read proposal, load IDPF-PRD, load anti-hallucination rules
**Step 2:** Run phases (Discovery → Elicitation → Specification → Generation)
**Step 3:** Generate `PRD/PRD-[Name].md`, update proposal status, `--remove-label "idea" --add-label "prd"`

## Trigger Words
`bug:`, `enhancement:`, `finding:`, `idea:`, `proposal:`, `prd:`
Create issue first (except `prd:` → Proposal-to-PRD workflow directly)

## Session Behavior Summary
- Auto-detect bugs/enhancements/proposals/ideas
- Create issues immediately
- Wait for instruction before working
- Update board status via `gh pmu move`
- Reference issues in commits
- Wait for "Done" before closing
- Move implemented proposals to `Proposal/Implemented/`
- Link sub-issues with `gh pmu sub add`

## Confirmation Message
```
GitHub Workflow Activated
━━━━━━━━━━━━━━━━━━━━━━━━━━
Project: {repo}
Board: https://github.com/users/{owner}/projects/{number}/views/1
Issues created auto. Work on instruction. Close on "Done".
```

## Manual Overrides
- "don't create an issue" - Skip creation
- "label as [X]" - Custom label
- "keep issue open" - Don't close
