---
version: "v0.25.0"
description: View, create, or manage project charter
argument-hint: "[update|refresh|validate]"
---
<!-- MANAGED -->
# /charter
Context-aware charter command. Shows summary if exists, starts creation if missing.
---
## Usage
| Command | Description |
|---------|-------------|
| `/charter` | Show summary or start creation |
| `/charter update` | Update specific sections |
| `/charter refresh` | Re-extract from code, merge with existing |
| `/charter validate` | Check work against scope |
---
## Template Detection
**Pattern:** `/{[a-z][a-z0-9-]*}/`
**Common Placeholders:** `{project-name}`, `{date}`, `{language}`, `{framework}`, `{database}`
| Scenario | Handling |
|----------|----------|
| ANY placeholder | Treat as template |
| Empty sections, no placeholders | Treat as complete |
---
## Workflow
### /charter (No Arguments)
**Step 1:** Check for charter: `test -f CHARTER.md`
**Step 2:** If exists, check for placeholders
```bash
grep -E '\{[a-z][a-z0-9-]*\}' CHARTER.md
```
**If TEMPLATE:** Proceed to extraction/inception
**If COMPLETE:** Show summary, mention update/validate commands
**Step 3:** If no charter or template:
1. Check if codebase has code
2. Has code → Offer extraction mode
3. Empty → Offer inception mode
4. Options: Create now, Skip, Never ask (creates `.no-charter`)
---
### Inception Mode
#### Essential Questions
| # | Question | Maps To |
|---|----------|---------|
| 1 | What are you building? | CHARTER.md Vision |
| 2 | What problem does it solve? | Problem Statement |
| 3 | What technology/language? | Tech Stack |
| 4 | What's in scope for v1? | In Scope |
#### Complexity-Triggered Questions
| Trigger | Follow-Up |
|---------|-----------|
| **Web app** | "Will users need to log in?" |
| **API service** | "Who will consume this API?" |
| **Multi-user** | "What access levels needed?" |
| **Data handling** | "Any sensitive/personal data?" |
| **External integrations** | "What external services?" |
Detection rules:
- Check answers for trigger keywords
- Add relevant questions to queue (1-2 max)
- Skip already-answered topics
#### Artifact Generation
| Answer | Primary Artifact |
|--------|------------------|
| What building? | CHARTER.md → Vision |
| What problem? | Charter-Details.md → Problem |
| What tech? | CHARTER.md → Tech Stack |
| What in scope? | CHARTER.md → In Scope |
**Process:**
1. Collect answers
2. Generate CHARTER.md from template
3. Generate Inception/ artifacts
4. Create Construction/ directories
5. Create Transition/ templates
6. Use "TBD" for missing sections
7. Commit: "Initialize project charter and lifecycle structure"
---
### /charter update
1. Read current charter and Charter-Details.md
2. Ask what to update: Vision, Focus, Tech Stack, Scope, Milestones
3. Apply updates, sync CHARTER.md if vision changes
4. Update date, commit
---
### /charter refresh
1. Analyze codebase (manifests, source, README)
2. Compare with existing Inception/ artifacts
3. Present diff to user
4. Merge confirmed changes, commit
---
### /charter validate
1. Load CHARTER.md and Scope-Boundaries.md
2. Identify current work (issue, commits, staged)
3. Compare against scope
4. Report: Aligned → proceed; Possibly out → confirm; Clearly out → suggest update
---
## Token Budget
| Artifact | Loaded | Tokens |
|----------|--------|--------|
| CHARTER.md | Always | ~150-200 |
| Charter-Details.md | update/validate | ~1,200-1,500 |
| Scope-Boundaries.md | validate | ~500-800 |
---
**End of /charter Command**
