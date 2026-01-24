---
version: "v0.31.0"
description: View, create, or manage project charter
argument-hint: "[update|refresh|validate]"
---
<!-- MANAGED -->
# /charter
Context-aware charter command. Shows summary if exists, starts creation if missing.
## Usage
| Command | Description |
|---------|-------------|
| `/charter` | Show summary (if exists) or start creation |
| `/charter update` | Update specific charter sections |
| `/charter refresh` | Re-extract from code, merge with existing |
| `/charter validate` | Check current work against charter scope |
## Template Detection
**Pattern:** `/{[a-z][a-z0-9-]*}/`
**Common:** `{project-name}`, `{date}`, `{language}`, `{framework}`, `{database}`
| Scenario | Handling |
|----------|----------|
| ANY placeholder present | Treat as template |
| Empty sections, no placeholders | Treat as complete |
## Workflow
### /charter (No Arguments)
**Step 1:** Check for charter: `test -f CHARTER.md`
**Step 2:** If exists, check for template placeholders
**If TEMPLATE (has placeholders):** Treat as "no charter" → proceed to Step 3
**If COMPLETE (no placeholders):**
1. Display charter summary (vision, focus, tech stack)
2. Mention: "Run `/charter update` to modify, `/charter validate` to check scope"
**Step 3: If no charter OR template detected**
1. Check if codebase has existing code
2. If has code → Offer extraction mode
3. If empty → Offer inception mode
4. Present options: Create now, Skip, Never ask (creates `.no-charter`)
### Extraction Mode (Existing Projects)
**Step 1:** Load `Skills/codebase-analysis/SKILL.md`
**Step 2:** Analyze codebase (tech stack, architecture, tests, NFRs)
**Step 3:** Present findings for user confirmation
**Step 4:** Generate CHARTER.md and Inception/ artifacts
### Inception Mode (New Projects)
#### Essential Questions
| # | Question | Maps To |
|---|----------|---------|
| 1 | What are you building? | CHARTER.md Vision |
| 2 | What problem does it solve? | Charter-Details.md Problem Statement |
| 3 | What technology/language? | CHARTER.md Tech Stack |
| 4 | What's in scope for v1? (3-5 items) | CHARTER.md In Scope |
| 5 | Testing framework? (conditional) | Test-Strategy.md Framework |
**Q5 conditional:** Only for testable projects (TypeScript, Python, Go, Rust, Java, C#). Skip for docs-only/config repos.
#### Complexity-Triggered Questions
| Trigger | Follow-Up |
|---------|-----------|
| **Web app** | "Will users log in?" / "What data stored?" |
| **API service** | "Who consumes?" / "Rate limiting?" |
| **Multi-user** | "Access levels?" / "Permission management?" |
| **Data handling** | "Sensitive data?" / "Compliance (GDPR, HIPAA)?" |
Ask 1-2 complexity questions max based on triggers in user answers.
#### Artifact Generation
**Answer-to-Artifact Mapping:**
| Answer | Primary Artifact | Secondary |
|--------|------------------|-----------|
| What building | CHARTER.md Vision | Charter-Details.md |
| Problem | Charter-Details.md Problem | - |
| Technology | CHARTER.md Tech Stack | Tech-Stack.md |
| Scope | CHARTER.md In Scope | Scope-Boundaries.md |
| Testing | Test-Strategy.md | Tech-Stack.md Dev Deps |
**Generation Process:**
1. Generate CHARTER.md from Templates/Lifecycle/CHARTER.md
2. Generate Inception/ artifacts (Charter-Details, Tech-Stack, Scope-Boundaries, Constraints, Architecture, Test-Strategy, Milestones)
3. Create Construction/ directories (.gitkeep placeholders)
4. Create Transition/ templates (Deployment-Guide, Runbook, User-Documentation)
5. Use "TBD" for sections without answers
6. Commit all artifacts
### /charter update
**Step 1:** Read current CHARTER.md and Inception/Charter-Details.md
**Step 2:** Ask what to update (Vision, Focus, Tech Stack, Scope, Milestones)
**Step 3:** Apply updates, sync to CHARTER.md if vision changes, update date, commit
### /charter refresh
**Step 1:** Load codebase-analysis skill
**Step 2:** Analyze codebase (tech stack, architecture, tests, NFRs)
**Step 3:** Compare with existing Inception/ artifacts
**Step 4:** Present diff for user confirmation
**Step 5:** Merge changes, sync CHARTER.md if needed, commit
### /charter validate
**Step 1:** Load CHARTER.md and Inception/Scope-Boundaries.md
**Step 2:** Identify current work (issue, recent commits, staged changes)
**Step 3:** Validate against scope
**Step 4:** Report:
| Finding | Action |
|---------|--------|
| Aligned | Proceed normally |
| Possibly out of scope | Ask user to confirm |
| Clearly out of scope | Suggest updating charter or revising work |
## Token Budget
| Artifact | Loaded | Tokens |
|----------|--------|--------|
| CHARTER.md | Always | ~150-200 |
| Charter-Details.md | For update/validate | ~1,200-1,500 |
| Scope-Boundaries.md | For validate | ~500-800 |
**End of /charter Command**
