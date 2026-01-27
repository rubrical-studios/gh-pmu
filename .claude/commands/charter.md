---
version: "v0.33.3"
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
| `/charter` | Show charter summary (if exists) or start creation (if missing) |
| `/charter update` | Update specific charter sections |
| `/charter refresh` | Re-extract from code, merge with existing |
| `/charter validate` | Check current work against charter scope |

---

## Execution Instructions

**REQUIRED:** Before executing this command:

1. **Generate Todo List:** Parse the workflow steps in this spec, then use `TodoWrite` to create todos
2. **Track Progress:** Mark todos `in_progress` → `completed` as you work
3. **Post-Compaction:** If resuming after context compaction, re-read this spec and regenerate todos

**Todo Generation Rules:**
- One todo per major workflow step
- Use the step description as the todo content

---

## Template Detection

When checking if `CHARTER.md` is a template vs a completed charter:

**Detection Pattern:**
```
Regex: /{[a-z][a-z0-9-]*}/
```

**Common Placeholders:**

| Pattern | Context |
|---------|---------|
| `{project-name}` | Title line |
| `{date}` | Last Updated field |
| `{language}` | Tech Stack table |
| `{framework}` | Tech Stack table |
| `{database}` | Tech Stack table |

**Edge Cases:**

| Scenario | Handling |
|----------|----------|
| ANY placeholder present | Treat as template |
| Literal braces in content | Regex avoids false positives (requires lowercase letter after `{`) |
| Empty sections, no placeholders | Treat as complete (user can elaborate later) |

---

## Workflow

### /charter (No Arguments)

**Step 1: Check for charter**

```bash
test -f CHARTER.md
```

**Step 2: If CHARTER.md exists, check for template placeholders**

Detect if the file is an unfilled template by checking for placeholder patterns:

```
Regex: /{[a-z][a-z0-9-]*}/
```

Common placeholders: `{project-name}`, `{date}`, `{language}`, `{framework}`, `{database}`

```bash
# Template detection (conceptual)
grep -E '\{[a-z][a-z0-9-]*\}' CHARTER.md
```

**If CHARTER.md is a TEMPLATE (has placeholders):**
→ Treat as "no charter" and proceed to Step 3 (Extraction/Inception)

**If CHARTER.md is COMPLETE (no placeholders):**
1. Read and display charter summary
2. Show: Project name, vision, current focus, tech stack
3. Mention: "Run `/charter update` to modify, `/charter validate` to check scope"

**Step 3: If no charter OR template detected**

**Charter is mandatory.** Automatically proceed with creation:

1. Check if codebase has existing code
2. If has code → Use extraction mode (analyze existing code)
3. If empty → Use inception mode (guided questions)
4. Proceed directly to charter creation (no skip option)

---

### Extraction Mode (Existing Projects)

When code exists and user chooses to create charter from existing code:

**Step 1: Load analysis skill**

Load `Skills/codebase-analysis/SKILL.md` for analysis capabilities.

**Step 2: Analyze codebase**

Use codebase-analysis skill:
- Tech stack detection (package.json, go.mod, etc.)
- Architecture inference (directory patterns)
- Test parsing (feature extraction from test names)
- NFR detection (code patterns)

**Step 3: Present findings**

Show extracted information and ask user to confirm/adjust:
- Detected tech stack
- Inferred architecture
- Extracted features from tests
- Identified NFRs

**Step 4: Generate charter artifacts**

Create CHARTER.md and Inception/ artifacts from confirmed findings.

---

### Inception Mode (New Projects)

When no code exists and user chooses to create charter, use guided questioning:

#### Essential Questions (Always Asked)

| # | Question | Maps To |
|---|----------|---------|
| 1 | What are you building? (1-2 sentences) | CHARTER.md Vision |
| 2 | What problem does it solve? | Inception/Charter-Details.md Problem Statement |
| 3 | What technology/language? | CHARTER.md Tech Stack, Inception/Tech-Stack.md |
| 4 | What's in scope for v1? (3-5 items) | CHARTER.md In Scope, Inception/Scope-Boundaries.md |
| 5 | What testing framework do you prefer? | Inception/Test-Strategy.md Framework |

**Note:** Question 5 is conditional - only asked for testable projects (see Testing Framework Question below).

**Questioning Flow:**

```
Question 1: "What are you building? (1-2 sentences describing the project)"
  → User answers
  → Stored for CHARTER.md Vision

Question 2: "What problem does it solve for users?"
  → User answers
  → Stored for Charter-Details.md Problem Statement

Question 3: "What technology/language will you use?"
  → User answers
  → Stored for CHARTER.md Tech Stack

Question 4: "What's in scope for v1? (List 3-5 key features or capabilities)"
  → User answers (can be comma-separated or bulleted)
  → Stored for CHARTER.md In Scope

Question 5 (Conditional): "What testing framework do you prefer?"
  → Only asked if tech stack is testable (see below)
  → Options presented based on detected tech stack
  → Stored for Test-Strategy.md Framework
```

#### Testing Framework Question (Conditional)

**TDD is mandated** per IDPF-Agile for all projects where TDD can be used. This question identifies the preferred framework, not the testing philosophy.

**When to Ask:**

| Tech Stack Detected | Ask Question? | Default Options |
|---------------------|:-------------:|-----------------|
| TypeScript, JavaScript, Node | ✅ Yes | Jest, Vitest, Bun test |
| Python | ✅ Yes | pytest, unittest |
| Go | ✅ Yes | testing (built-in), testify |
| Rust | ✅ Yes | cargo test (built-in) |
| Java, Kotlin | ✅ Yes | JUnit, TestNG |
| C#, .NET | ✅ Yes | xUnit, NUnit, MSTest |
| Documentation-only, Config repos | ❌ Skip | N/A |
| Unknown/Other | ✅ Yes | "Not sure yet" option |

**Question Format:**

```
"What testing framework do you prefer?"

[Options based on tech from Q3]
- Option 1: [Primary framework for stack]
- Option 2: [Alternative framework]
- Option 3: Not sure yet (TBD)
```

**Example for TypeScript:**
```
"What testing framework do you prefer?"
- Jest (most popular, good IDE support)
- Vitest (fast, Vite-native)
- Bun test (if using Bun runtime)
- Not sure yet
```

**Skip Detection:**

```
If Q3 answer contains:
  - "documentation", "docs", "config", "terraform", "ansible"
  - No code-related tech detected
→ Skip Q5, set Test-Strategy.md framework to "N/A - non-code project"
```

**After essential questions:**
- Check if answers suggest complexity (see Complexity-Triggered Questions)
- If simple project, proceed to artifact generation
- If complex indicators detected, ask follow-up questions

#### Complexity-Triggered Questions

Detect complexity from user's answers and ask relevant follow-ups:

| Trigger (Detected In Answers) | Follow-Up Questions | Maps To |
|-------------------------------|---------------------|---------|
| **Web app** (mentions "web", "frontend", "UI", "browser") | "Will users need to log in?" / "What data will you store?" | Inception/Constraints.md (Auth), Inception/Architecture.md |
| **API service** (mentions "API", "REST", "GraphQL", "service") | "Who will consume this API?" / "Any rate limiting needs?" | Inception/Charter-Details.md (Target Users), Inception/Constraints.md |
| **Multi-user** (mentions "users", "team", "roles", "permissions") | "What access levels are needed?" / "How are permissions managed?" | Inception/Constraints.md (Security) |
| **Data handling** (mentions "data", "database", "storage", "PII") | "Any sensitive/personal data?" / "Compliance requirements (GDPR, HIPAA)?" | Inception/Constraints.md (Compliance) |
| **External integrations** (mentions third-party services) | "What external services?" / "Any constraints from those services?" | Inception/Tech-Stack.md, Inception/Constraints.md |

**Trigger Detection:**
```
For each user answer:
  1. Check for trigger keywords
  2. If triggered, add relevant questions to queue
  3. Skip questions already answered indirectly
  4. Ask 1-2 complexity questions max (avoid overwhelming)
```

**Example Flow:**
```
User answers Q1: "A web app for tracking personal expenses"
  → Triggers: "web app" + "personal" (PII)

Follow-up Q5: "Will users need to log in?"
  → User: "Yes, email/password"
  → Stored for Constraints.md (Auth required)

Follow-up Q6: "Any compliance needs for personal financial data?"
  → User: "No, it's personal use only"
  → Stored for Constraints.md (No compliance)
```

#### Dynamic Follow-Up Generation

Generate context-specific follow-ups based on user's actual answers (NOT from static question banks):

**Generation Rules:**

1. **Analyze baseline answers** - Look for gaps, ambiguities, or implied requirements
2. **Project-type awareness** - CLI vs web vs API vs library have different needs
3. **Depth scaling** - Simple projects get 0-1 follow-ups, complex get 2-4
4. **No redundancy** - Skip questions already answered indirectly

**Follow-Up Categories:**

| Answer Pattern | Dynamic Follow-Up |
|----------------|-------------------|
| Vague scope item ("handle user data") | "What specifically happens with user data?" |
| Implicit constraint ("needs to be fast") | "Any specific performance requirements (response time, throughput)?" |
| Missing boundary ("file uploads") | "Any size limits or file type restrictions?" |
| Unclear priority ("reporting features") | "Is reporting core to v1 or a nice-to-have?" |

**Question Count Scaling:**

| Project Complexity | Total Questions |
|--------------------|-----------------|
| Simple (CLI tool, utility) | 4 essential only |
| Medium (web app, API) | 4-6 questions |
| Complex (multi-service, compliance) | 6-8 questions |

**Generation Prompt:**
```
Based on the user's answers:
1. Vision: "{answer1}"
2. Problem: "{answer2}"
3. Tech: "{answer3}"
4. Scope: "{answer4}"

Generate 0-2 follow-up questions to clarify:
- Ambiguous scope items
- Implied but unstated requirements
- Missing constraints for this project type
- Unclear priorities

Do NOT ask about topics already covered in answers.
```

#### Artifact Generation from Answers

After all questions answered, generate lifecycle artifacts:

**Answer-to-Artifact Mapping:**

| Answer | Primary Artifact | Secondary Artifacts |
|--------|------------------|---------------------|
| What are you building? | CHARTER.md → Vision | Inception/Charter-Details.md → Vision (Full) |
| What problem? | Inception/Charter-Details.md → Problem Statement | CHARTER.md → Vision (context) |
| What technology? | CHARTER.md → Tech Stack | Inception/Tech-Stack.md → Core Stack |
| What's in scope? | CHARTER.md → In Scope | Inception/Scope-Boundaries.md → In Scope |
| Testing framework? | Inception/Test-Strategy.md → Framework | Inception/Tech-Stack.md → Dev Dependencies |
| TDD mandate (implicit) | Inception/Constraints.md → Development Workflow | - |
| Auth requirements? | Inception/Constraints.md → Security | - |
| Data handling? | Inception/Constraints.md → Compliance | - |
| External services? | Inception/Tech-Stack.md → Dependencies | Inception/Constraints.md |

**Generation Process:**

```
1. Collect all answers into structured data

2. Generate CHARTER.md using Templates/Lifecycle/CHARTER.md
   - Fill Vision, Tech Stack, In Scope from answers
   - Set Status: Draft
   - Set Last Updated: today

3. Generate Inception/ artifacts:
   - Charter-Details.md (vision, problem, target users)
   - Tech-Stack.md (language, framework, dependencies, test framework)
   - Scope-Boundaries.md (in-scope items)
   - Constraints.md (TDD workflow constraint, NFRs from complexity questions)
   - Architecture.md (basic structure if tech answered)
   - Test-Strategy.md (TDD philosophy + framework from Q5)
   - Milestones.md (TBD placeholder)

4. Create Construction/ directories (empty):
   - Construction/Test-Plans/.gitkeep
   - Construction/Design-Decisions/.gitkeep
   - Construction/Sprint-Retros/.gitkeep
   - Construction/Tech-Debt/.gitkeep
   - Construction/README.md (describes subdirectory purposes)

5. Create Transition/ directory (empty templates):
   - Transition/Deployment-Guide.md (TBD template)
   - Transition/Runbook.md (TBD template)
   - Transition/User-Documentation.md (TBD template)

6. Use "TBD" for sections without answers:
   - "TBD - To be determined during development"
   - Highlights what still needs input

7. Commit all artifacts in single commit:
   - "Initialize project charter and lifecycle structure"
```

**Example Output:**

CHARTER.md:
```markdown
# Project Charter: Expense Tracker

**Status:** Draft
**Last Updated:** 2025-01-15

## Vision
A web app for tracking personal expenses to help users understand spending patterns.

## Tech Stack
| Layer | Technology |
|-------|------------|
| Language | TypeScript |
| Framework | Next.js |
| Database | PostgreSQL |

## In Scope (v1)
- Expense entry with categories
- Monthly spending reports
- Budget setting and alerts
```

Inception/Test-Strategy.md:
```markdown
# Test Strategy

## Testing Philosophy

**TDD (Test-Driven Development)** is mandated per IDPF-Agile.

All features follow the RED-GREEN-REFACTOR cycle:
1. **RED:** Write failing test first
2. **GREEN:** Write minimal code to pass
3. **REFACTOR:** Improve code while tests pass

## Framework

| Type | Framework | Rationale |
|------|-----------|-----------|
| Unit Tests | Vitest | User preference, fast execution |
| Integration | Vitest | Consistent tooling |
| E2E | TBD | To be determined based on needs |

## Coverage Targets

| Type | Target | Notes |
|------|--------|-------|
| Unit | 80%+ | Core business logic |
| Integration | Key flows | Cross-component verification |
| E2E | Critical paths | User journey validation |

## Test Organization

```
tests/
├── unit/           # Unit tests mirror src/ structure
├── integration/    # Cross-component tests
└── e2e/            # End-to-end user journeys
```
```

---

### /charter update

Updates specific sections of the charter.

**Step 1: Read current charter**

```bash
cat CHARTER.md
cat Inception/Charter-Details.md
```

**Step 2: Ask what to update**

Present sections:
- Vision
- Current Focus
- Tech Stack
- Scope (In/Out)
- Milestones

**Step 3: Apply updates**

1. Update `Inception/Charter-Details.md` with changes
2. If vision changes, sync to `CHARTER.md`
3. Update `Last Updated` date
4. Commit changes

**Step 4: Trigger suggestions (if Tech Stack modified)**

If user updated Tech Stack section:
1. Run **Project Skills Selection** (see above) - suggest NEW skills only
2. Run **Extension Recipe Suggestions** (see above) - suggest NEW recipes only

**Note:** Only suggest items not already installed. Compare against `projectSkills` array and existing extension point content.

---

### /charter refresh

Re-extracts project info from code and merges with existing charter.

**Step 1: Load analysis skill**

Load `Skills/codebase-analysis/SKILL.md` for analysis capabilities.

**Step 2: Analyze codebase**

Use codebase-analysis skill:
- Tech stack detection (package.json, go.mod, etc.)
- Architecture inference (directory patterns)
- Test parsing (feature extraction)
- NFR detection (code patterns)

**Step 3: Compare with existing**

1. Read current `Inception/` artifacts
2. Identify differences:
   - New dependencies
   - Changed architecture
   - Missing scope items

**Step 4: Present diff**

Show user what changed and ask for confirmation.

**Step 5: Merge changes**

1. Update relevant `Inception/` files
2. Sync `CHARTER.md` if vision/focus changed
3. Commit with "Charter refresh" message

**Step 6: Trigger suggestions (if tech stack changed)**

If tech stack detection found new indicators:
1. Run **Project Skills Selection** - suggest NEW skills only
2. Run **Extension Recipe Suggestions** - suggest NEW recipes only

**Note:** Only suggest items not already installed. This allows projects to pick up new skills/recipes as they adopt new technologies.

---

### /charter validate

Checks current work against charter scope.

**Step 1: Load charter context**

```bash
cat CHARTER.md
cat Inception/Scope-Boundaries.md
```

**Step 2: Identify current work**

Check:
- Current issue being worked (from conversation)
- Recent commits
- Staged changes

**Step 3: Validate alignment**

Compare work against:
- In-scope items
- Out-of-scope items
- Current focus

**Step 4: Report**

| Finding | Action |
|---------|--------|
| Aligned with scope | Proceed normally |
| Possibly out of scope | Ask user to confirm intent |
| Clearly out of scope | Suggest updating charter or revising work |

---

## Project Skills Selection

After charter creation (extraction or inception), suggest relevant skills based on the tech stack:

### Step 1: Load Skill Registry

```bash
cat .claude/metadata/skill-registry.json
```

Parse the registry to get skill names, descriptions, and triggers.

### Step 2: Match Charter Against Triggers

Compare charter content (tech stack, vision, scope) against skill triggers:

| Charter Content | Matched Skill |
|-----------------|---------------|
| Tech stack mentions "Electron" | `electron-development` |
| Scope mentions "E2E testing" | `playwright-setup` |
| Tech stack mentions "PostgreSQL" | `postgresql-integration` |

**Matching Logic:**
```
For each skill in registry:
  For each trigger in skill.triggers:
    If charter content contains trigger (case-insensitive):
      Add skill to matches
```

### Step 3: Present Matches to User

If matches found:

**ASK USER:**
```
Based on your charter, these skills may be relevant:

- electron-development (8 triggers matched: electron, vite, playwright...)
- playwright-setup (E2E testing mentioned in scope)

Include these skills in your project? (yes/no/edit list)
```

Provide options:
- **Yes** - Add all matched skills
- **No** - Skip skill binding
- **Edit list** - Select specific skills to include

### Step 4: Store in framework-config.json

If user confirms skills:

```javascript
// Read existing config
const config = JSON.parse(fs.readFileSync('framework-config.json'));

// Add projectSkills array
config.projectSkills = ["electron-development", "playwright-setup"];

// Write back
fs.writeFileSync('framework-config.json', JSON.stringify(config, null, 2));
```

**Note:** If `framework-config.json` doesn't exist (pre-install project), create it with minimal structure:

```json
{
  "projectSkills": ["electron-development", "playwright-setup"]
}
```

### Step 4b: Deploy Selected Skills

After storing skill references, deploy the actual skill packages to the project:

```bash
# Use shared install-skill.js script
node .claude/scripts/shared/install-skill.js <skill-names...>
```

**Deployment Logic:**
```
For each skill in confirmed list:
  1. Read frameworkPath from framework-config.json
  2. Locate package: {frameworkPath}/Skills/Packaged/{skill}.zip
  3. Extract to: .claude/skills/{skill}/
  4. Verify SKILL.md exists
  5. Report result
```

**Edge Cases:**

| Scenario | Handling |
|----------|----------|
| `frameworkPath` not set | Warn: "Cannot deploy skills - framework path unknown. Skills saved for later deployment." |
| Package not found | Warn: "Skill {name} not found in framework" |
| Skill already deployed | Skip with note: "Skill {name} already installed" |
| Extraction fails | Error with details, continue with others |

### Step 5: Report

```
Project skills configured and deployed:
✓ electron-development - Installed (12 resources)
✓ playwright-setup - Installed (2 resources)

Skills will be reported at session startup.
Run `/install-skill --list` to see available skills.
Run `/charter update` to modify skills later.
```

---

## Extension Recipe Suggestions

After skill selection, suggest relevant extension recipes based on the detected tech stack.

**Triggers:**
- `/charter` (creation) - Always after skill selection
- `/charter update` - If Tech Stack section is modified
- `/charter refresh` - Always after re-detection

**Skip if:**
- `framework-config.json` has `"extensionSuggestions": false`
- No release commands are installed (check for `.claude/commands/prepare-release.md`)

### Step 1: Load Recipe Mapping

```bash
cat .claude/metadata/recipe-tech-mapping.json
```

Parse the mapping to get tech indicator → recipe associations.

### Step 2: Match Charter Tech Stack Against Indicators

For each tech indicator detected during charter creation/refresh:

```
For each techIndicator in detectedTechStack:
  If techIndicator exists in recipe-tech-mapping.json:
    Add associated recipes to suggestions
```

Also check `groupMappings` for broader matches (e.g., any CI config → ci-gate-before-merge).

### Step 3: Filter Already-Installed Recipes

Check which extension points already have content:

```bash
# For each suggested recipe, check its target extension point
# Example: nodejs-tests targets prepare-release:pre-validation
grep -A2 "USER-EXTENSION-START: pre-validation" .claude/commands/prepare-release.md
```

**If extension point has content:** Remove recipe from suggestions (already configured).

### Step 4: Present Suggestions

If matches found (after filtering):

**ASK USER:**
```
Extension Recipes Available for Your Project
─────────────────────────────────────────────

Based on your tech stack, these extension recipes may be useful:

┌─────────────────────┬──────────────────────────────────────────────────┐
│ Recipe              │ Description                                      │
├─────────────────────┼──────────────────────────────────────────────────┤
│ nodejs-tests        │ Run npm test before release validation           │
│ dependency-audit    │ Check for known vulnerabilities                  │
│ ci-gate-before-merge│ Block merge/tag until CI passes                  │
└─────────────────────┴──────────────────────────────────────────────────┘

Install these recipes to your release commands? (y/n/select):
```

**Response options:**
- `y` / `yes` - Install all suggested recipes
- `n` / `no` - Skip recipe installation
- `select` - Present numbered list for individual selection

**If no matches:**
```
No extension recipes matched your tech stack.
Run `/extensions recipes` to browse all available recipes.
```

### Step 5: Implement Selected Recipes

For each confirmed recipe:

1. **Load recipe template** from `extension-recipes.json`
2. **Determine target extension point** from recipe's `appliesTo` field
3. **Locate extension markers** in command file:
   ```
   <!-- USER-EXTENSION-START: {point} -->
   <!-- USER-EXTENSION-END: {point} -->
   ```
4. **Insert template** between markers
5. **Report result**

**Implementation Logic:**
```
For each recipe in confirmedRecipes:
  template = extensionRecipes.categories[category].recipes[recipe].template
  targetPoint = recipe.appliesTo[0]  // e.g., "prepare-release:pre-validation"
  [command, point] = targetPoint.split(":")

  commandFile = ".claude/commands/{command}.md"

  Find: <!-- USER-EXTENSION-START: {point} -->
  Insert template after START marker

  Report: "✓ {recipe} → {command}:{point}"
```

### Step 6: Report Results

```
Implementing extension recipes...

✓ nodejs-tests → prepare-release:pre-validation
✓ dependency-audit → prepare-release:pre-validation
✓ ci-gate-before-merge → prepare-release:pre-tag

3 recipes implemented.
Run `/extensions list` to view all extension points.
Run `/extensions edit <command>:<point>` to modify.
```

### Edge Cases

| Scenario | Handling |
|----------|----------|
| Extension point already has content | Skip with note: "{point} already configured, skipping {recipe}" |
| No release commands installed | Skip entire section: "Extension recipes require release commands" |
| Recipe prerequisites not met | Note in output: "{recipe} requires {prerequisite}" |
| All suggestions already installed | Report: "Extension recipes are up to date" |

### Opt-Out

To disable extension recipe suggestions, add to `framework-config.json`:

```json
{
  "extensionSuggestions": false
}
```

---

## Token Budget

| Artifact | Loaded | Tokens |
|----------|--------|--------|
| `CHARTER.md` | Always for /charter | ~150-200 |
| `Inception/Charter-Details.md` | For update/validate | ~1,200-1,500 |
| `Inception/Scope-Boundaries.md` | For validate | ~500-800 |
| `skill-registry.json` | For skill selection | ~200-400 |
| `recipe-tech-mapping.json` | For recipe suggestions | ~300-500 |
| `extension-recipes.json` | For recipe templates | ~800-1,200 |

---

## Related Commands

- `/charter update` - Modify charter sections
- `/charter refresh` - Sync charter with codebase
- `/charter validate` - Check scope alignment

---

**End of /charter Command**
