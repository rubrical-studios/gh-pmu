---
version: "v0.44.0"
description: View, create, or manage project charter
argument-hint: "[update|refresh|validate]"
---
<!-- MANAGED -->
# /charter
Context-aware charter command. Shows summary if exists, starts creation if missing.
## Usage
| Command | Description |
|---------|-------------|
| `/charter` | Show charter summary (if exists) or start creation (if missing) |
| `/charter update` | Update specific charter sections |
| `/charter refresh` | Re-extract from code, merge with existing |
| `/charter validate` | Check current work against charter scope |
## Execution Instructions
**REQUIRED:** Before executing:
1. **Generate Todo List:** Parse workflow steps, use `TodoWrite` to create todos
2. **Track Progress:** Mark todos `in_progress` → `completed` as you work
3. **Post-Compaction:** If resuming, re-read this spec and regenerate todos
## Template Detection
**Pattern:** `/{[a-z][a-z0-9-]*}/`
| Scenario | Handling |
|----------|----------|
| ANY placeholder present | Treat as template |
| Empty sections, no placeholders | Treat as complete |
## Workflow
### /charter (No Arguments)
**Step 1: Check for charter**
```bash
test -f CHARTER.md
```
**Step 2: If exists, check for template placeholders**
```bash
grep -E '\{[a-z][a-z0-9-]*\}' CHARTER.md
```
**If TEMPLATE (has placeholders):** → Proceed to Step 3 (Extraction/Inception)
**If COMPLETE (no placeholders):**
1. Read and display charter summary
2. Show: Project name, vision, current focus, tech stack
3. Mention: "Run `/charter update` to modify, `/charter validate` to check scope"
**Step 3: If no charter OR template detected**
**Charter is mandatory.** Automatically proceed:
1. Check if codebase has existing code
2. If has code → Use extraction mode
3. If empty → Use inception mode
4. Proceed directly to charter creation (no skip option)
### Extraction Mode (Existing Projects)
**Step 1:** Load `Skills/codebase-analysis/SKILL.md`
**Step 2:** Analyze codebase (tech stack, architecture, test parsing, NFR detection)
**Step 3:** Present findings, ask user to confirm/adjust
**Step 4:** Generate CHARTER.md and Inception/ artifacts from confirmed findings
### Inception Mode (New Projects)
#### Essential Questions (Always Asked)
| # | Question | Maps To |
|---|----------|---------|
| 1 | What are you building? (1-2 sentences) | CHARTER.md Vision |
| 2 | What problem does it solve? | Inception/Charter-Details.md Problem Statement |
| 3 | What technology/language? | CHARTER.md Tech Stack |
| 4 | What's in scope for v1? (3-5 items) | CHARTER.md In Scope |
| 5 | What testing framework? (conditional) | Inception/Test-Strategy.md Framework |
**Note:** Q5 only asked for testable projects (skip for docs/config repos).
#### Testing Framework Question (Conditional)
| Tech Stack | Ask? | Options |
|------------|:----:|---------|
| TypeScript/JS/Node | Yes | Jest, Vitest, Bun test |
| Python | Yes | pytest, unittest |
| Go | Yes | testing, testify |
| Documentation-only | Skip | N/A |
#### Complexity-Triggered Questions
| Trigger | Follow-Up |
|---------|-----------|
| **Web app** | "Will users need to log in?" / "What data will you store?" |
| **API service** | "Who will consume this API?" |
| **Multi-user** | "What access levels are needed?" |
| **Data handling** | "Any sensitive/personal data?" / "Compliance requirements?" |
**Max 1-2 complexity questions to avoid overwhelming user.**
#### Dynamic Follow-Up Generation
- Analyze baseline answers for gaps/ambiguities
- Simple projects: 0-1 follow-ups; Complex: 2-4
- Skip questions already answered indirectly
#### Artifact Generation from Answers
**Answer-to-Artifact Mapping:**
| Answer | Primary Artifact |
|--------|------------------|
| What building? | CHARTER.md → Vision |
| What problem? | Inception/Charter-Details.md → Problem Statement |
| What technology? | CHARTER.md → Tech Stack |
| What's in scope? | CHARTER.md → In Scope |
| Testing framework? | Inception/Test-Strategy.md → Framework |
**Generation Process:**
1. Create lifecycle directory structure:
   ```bash
   mkdir -p Inception Construction/Test-Plans Construction/Design-Decisions Construction/Tech-Debt Transition
   ```
2. Generate CHARTER.md (Vision, Tech Stack, In Scope, Status: Draft)
3. Generate Inception/ artifacts (Charter-Details, Tech-Stack, Scope-Boundaries, Constraints, Architecture, Test-Strategy, Milestones)
4. Create Construction/ structure with .gitkeep files and README.md
5. Create Transition/ artifacts (Deployment-Guide, Runbook, User-Documentation)
6. Use "TBD" for sections without answers
7. Commit all artifacts: "Initialize project charter and lifecycle structure"
### /charter update
**Step 1:** Read current CHARTER.md and Inception/Charter-Details.md
**Step 2:** Ask what to update (Vision, Current Focus, Tech Stack, Scope, Milestones)
**Step 3:** Apply updates, sync to CHARTER.md if vision changes, update Last Updated date
**Step 4:** If Tech Stack modified, trigger skill and recipe suggestions (NEW items only)
### /charter refresh
**Step 1:** Load `Skills/codebase-analysis/SKILL.md`
**Step 2:** Analyze codebase
**Step 3:** Compare with existing Inception/ artifacts, identify differences
**Step 4:** Present diff, ask for confirmation
**Step 5:** Merge changes, commit "Charter refresh"
**Step 6:** If tech stack changed, trigger skill and recipe suggestions (NEW items only)
### /charter validate
**Step 1:** Load CHARTER.md and Inception/Scope-Boundaries.md
**Step 2:** Identify current work (issue, recent commits, staged changes)
**Step 3:** Compare against in-scope/out-of-scope items
**Step 4:** Report
| Finding | Action |
|---------|--------|
| Aligned | Proceed normally |
| Possibly out of scope | Ask user to confirm intent |
| Clearly out of scope | Suggest updating charter or revising work |
## Project Skills Selection
After charter creation, suggest relevant skills based on tech stack:
**Step 1:** Load `.claude/metadata/skill-registry.json`
**Step 2:** Match charter content against skill triggers (case-insensitive)
**Step 3: ASK USER:**
```
Based on your charter, these skills may be relevant:
- electron-development (matched: electron, vite...)
Include these skills? (yes/no/edit list)
```
**Step 4:** Store in framework-config.json `projectSkills` array
**Step 4b:** Deploy skills via `node .claude/scripts/shared/install-skill.js <skill-names...>`
**Step 5:** Report installed skills
## Extension Recipe Suggestions
After skill selection, suggest relevant extension recipes.
**Triggers:** `/charter` (creation), `/charter update` (if Tech Stack modified), `/charter refresh`
**Skip if:** `"extensionSuggestions": false` or no release commands installed
**Step 1:** Load `.claude/metadata/recipe-tech-mapping.json`
**Step 2:** Match tech stack against indicators
**Step 3:** Filter already-installed recipes (check extension points for content)
**Step 4: ASK USER:**
```
Extension Recipes Available:
- nodejs-tests: Run npm test before release validation
- dependency-audit: Check for vulnerabilities
Install? (y/n/select)
```
**Step 5:** Implement selected recipes (insert template into extension markers)
**Step 6:** Report results
| Edge Case | Handling |
|-----------|----------|
| Extension point has content | Skip: "{point} already configured" |
| No release commands | Skip: "Extension recipes require release commands" |
## Token Budget
| Artifact | Tokens |
|----------|--------|
| CHARTER.md | ~150-200 |
| Charter-Details.md | ~1,200-1,500 |
| Scope-Boundaries.md | ~500-800 |
## Related Commands
- `/charter update` - Modify charter sections
- `/charter refresh` - Sync charter with codebase
- `/charter validate` - Check scope alignment
**End of /charter Command**
