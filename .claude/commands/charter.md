---
version: "v0.46.2"
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
| Rust | Yes | cargo test |
| Java/Kotlin | Yes | JUnit, TestNG |
| C#/.NET | Yes | xUnit, NUnit, MSTest |
| Documentation-only | Skip | N/A |
**Skip Detection:** If Q3 contains "documentation", "docs", "config", "terraform", "ansible" → skip Q5, set framework to "N/A - non-code project"
#### Complexity-Triggered Questions
| Trigger | Follow-Up |
|---------|-----------|
| **Web app** | "Will users need to log in?" / "What data will you store?" |
| **API service** | "Who will consume this API?" |
| **Multi-user** | "What access levels are needed?" |
| **Data handling** | "Any sensitive/personal data?" / "Compliance requirements?" |
| **External integrations** | "What external services?" / "Any constraints?" |
**Max 1-2 complexity questions to avoid overwhelming user.**
#### Dynamic Follow-Up Generation
- Analyze baseline answers for gaps/ambiguities
- Simple projects: 0-1 follow-ups; Complex: 2-4
- Skip questions already answered indirectly
| Project Complexity | Total Questions |
|--------------------|-----------------|
| Simple (CLI, utility) | 4 essential only |
| Medium (web app, API) | 4-6 questions |
| Complex (multi-service) | 6-8 questions |
#### Review Mode Question (Always Asked)
After essential questions and complexity follow-ups, ask about review mode:
**ASK USER (single-select via AskUserQuestion):**
```
What review mode should be used for this project?
- Solo: Single developer - skip team-oriented criteria
- Team (Recommended): 2-10 developers - include sizing, priorities, dependencies
- Enterprise: Large teams - all criteria plus effort estimation and risk assessment
```
**Default:** "team" if user doesn't select or non-interactive.
**After answer:**
1. Write `reviewMode` to `framework-config.json` (lowercase: "solo", "team", or "enterprise")
2. Show confirmation with mode-specific explanation
#### Artifact Generation from Answers
**Answer-to-Artifact Mapping:**
| Answer | Primary Artifact |
|--------|------------------|
| What building? | CHARTER.md → Vision |
| What problem? | Inception/Charter-Details.md → Problem Statement |
| What technology? | CHARTER.md → Tech Stack |
| What's in scope? | CHARTER.md → In Scope |
| Testing framework? | Inception/Test-Strategy.md → Framework |
| Review mode? | framework-config.json → reviewMode |
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
**Note:** Directories created after questions to avoid orphaned directories if user abandons mid-flow.
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
After charter creation, suggest relevant skills based on tech stack using `.claude/metadata/skill-keywords.json`.
**Step 1:** Load `.claude/metadata/skill-keywords.json` (contains `skillKeywords` and `groupKeywords`) and `.claude/metadata/skill-registry.json` (for descriptions)
**If `skill-keywords.json` missing:** Warn and skip (non-blocking).
**Step 2:** Match tech stack keywords against skillKeywords entries (case-insensitive, whole-word). Collect all skills with at least 1 keyword match as candidates — no false positive from partial string matching. Also match groupKeywords — if group keyword matches, add ALL group.skills. Deduplicate against existing `projectSkills`.
**If tech stack unknown:** Skip. **If zero matches found:** Report "No matching skills" and continue.
**Step 3:** Present candidates via `AskUserQuestion` with multi-select. Show skill name and description for each candidate.
**Step 3b: Existing Project — Additive Merge:** Read existing `projectSkills`, filter already-present candidates. If all relevant skills enabled, report and skip. Present only NEW candidates. Merge additively.
**Step 4:** Store confirmed skills in `framework-config.json` `projectSkills` array, sorted alphabetically. Additive merge with existing.
**Step 4b:** Deploy skills via `node .claude/scripts/shared/install-skill.js <skill-names...>`
**Step 5:** Report installed skills
## Extension Recipe Suggestions
After skill selection, suggest relevant extension recipes.
**Triggers:** `/charter` (creation), `/charter update` (if Tech Stack modified), `/charter refresh`
**Skip if:** `"extensionSuggestions": false` or no release commands installed
**Step 1:** Load `.claude/metadata/recipe-tech-mapping.json`
**Step 2:** Match tech stack against indicators and groupMappings
**Step 3:** Filter already-installed recipes (check extension points for content)
**Step 4: ASK USER:**
```
Extension Recipes Available:
- nodejs-tests: Run npm test before release validation
- dependency-audit: Check for vulnerabilities
Install? (y/n/select)
```
**Step 5:** Implement selected recipes (insert template between `USER-EXTENSION-START/END` markers)
**Step 6:** Report results
| Edge Case | Handling |
|-----------|----------|
| Extension point has content | Skip: "{point} already configured" |
| No release commands | Skip: "Extension recipes require release commands" |
| All suggestions installed | Report: "Extension recipes are up to date" |
## Praxis Diagram Configuration
After extension recipe suggestions, check if project uses Praxis Diagram and generate `.praxis-diagram.json`.
**Triggers:** `/charter` (creation), `/charter refresh`
**Skip if:** `.praxis-diagram.json` already exists
**Step 1:** Use `detectShapeLibrary()` from `.claude/scripts/shared/praxis-diagram-config.js`
| Tech Stack Contains | Shape Library |
|---------------------|---------------|
| `flowbite-svelte` or (`flowbite` + `svelte`) | `flowbite-svelte` |
| `flowbite-react` or (`flowbite` + `react`) | `flowbite-react` |
| Neither | `null` → prompt user |
**Step 2:** If undetectable, ASK USER via `AskUserQuestion` (options from `getAvailableShapeLibraries()`). Include "Skip" option.
**Step 3:** Generate `.praxis-diagram.json` via `generateConfig(shapeLibrary)`:
```json
{
  "shapes": "flowbite-svelte"
}
```
**Step 4:** Report configuration result
## Token Budget
| Artifact | Tokens |
|----------|--------|
| CHARTER.md | ~150-200 |
| Charter-Details.md | ~1,200-1,500 |
| Scope-Boundaries.md | ~500-800 |
| skill-keywords.json | ~300-500 |
## Related Commands
- `/charter update` - Modify charter sections
- `/charter refresh` - Sync charter with codebase
- `/charter validate` - Check scope alignment
**End of /charter Command**
