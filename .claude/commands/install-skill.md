---
version: "v0.49.1"
description: Install skills from the framework to your project
argument-hint: "<skill-name> | --list [--installed]"
---

<!-- MANAGED -->
# /install-skill
Enable skills from the IDPF framework hub for use in development sessions.
---
## Usage
| Command | Description |
|---------|-------------|
| `/install-skill <skill-name>` | Enable a specific skill |
| `/install-skill <skill1> <skill2>` | Enable multiple skills |
| `/install-skill --list` | List available skills |
| `/install-skill --list --installed` | Show installation status |
---
## Prerequisites
- Framework must be installed (`framework-config.json`)
- Skill registry must exist (`.claude/metadata/skill-registry.json`)
---
## List Mode (`--list`)
**Step 1: Load skill registry**
Use the Read tool directly on `.claude/metadata/skill-registry.json` (do NOT use Glob — it doesn't follow symlinks in user projects).
**Step 2: Load projectSkills from config**
```bash
node -e "const c=JSON.parse(require('fs').readFileSync('framework-config.json')); console.log(JSON.stringify(c.projectSkills || []))"
```
**Step 3: Display Format**
```
Available Skills:

Name                          Description                                          Status
────────────────────────────────────────────────────────────────────────────────────────────────────
electron-development          Patterns and solutions for Electron app development  Enabled
playwright-setup              Installation verification and troubleshooting        Available

Total: 23 skills
Enabled (in projectSkills): 6
```
**Status:** `Enabled` = in projectSkills, `Available` = in registry but not in projectSkills.
If `--installed`: always include Status column.
---
## Install Mode (`<skill-name>`)
**Step 1:** Validate prerequisites — `framework-config.json` must exist.
**Step 2:** Verify skill exists in registry.
**Step 3:** Update `framework-config.json` — add to `projectSkills` array (sorted, no duplicates).
**Step 4:** Report results:
```
Enabling skills...

✓ electron-development - Enabled (12 resources)
⊘ tdd-red-phase - Already enabled (skipped)
✗ nonexistent-skill - Not found in registry

Enabled: 2  Skipped: 1  Failed: 1
```
---
## Output Symbols
| Symbol | Meaning |
|--------|---------|
| ✓ | Successfully enabled |
| ⊘ | Already enabled (skipped) |
| ✗ | Failed (not found) |
---
## Edge Cases
| Scenario | Handling |
|----------|----------|
| `framework-config.json` missing | Error: "Run framework installer first" |
| Skill not in registry | Error: "Not found in registry" (continue with others) |
| Already enabled | Skip with message (not an error) |
---
## Related
- `/charter` - Automatically suggests and deploys skills during charter creation
- `Overview/Framework-Skills.md` - Complete skills documentation
---
**End of /install-skill Command**
