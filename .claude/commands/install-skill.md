---
version: "v0.46.1"
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

## Execution Instructions

### List Mode (`--list`)

**Step 1: Load skill registry**

```bash
cat .claude/metadata/skill-registry.json
```

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
postgresql-integration        Guide through PostgreSQL setup and best practices    Available

Total: 23 skills
Enabled (in projectSkills): 6
```

**Status values:**
- `Enabled` - In `projectSkills` array (will appear in story templates)
- `Available` - In registry, not yet in `projectSkills`

**If `--installed` flag specified:** Always include Status column.

---

### Install Mode (`<skill-name>`)

**Step 1: Validate Prerequisites**

```bash
test -f framework-config.json || { echo "ERROR: framework-config.json not found"; exit 1; }
```

**Step 2: Verify skill exists in registry**

```bash
node -e "const r=JSON.parse(require('fs').readFileSync('.claude/metadata/skill-registry.json')); const s=r.skills.find(x=>x.name==='{skill}'); console.log(s ? 'FOUND' : 'NOT_FOUND')"
```

**Step 3: Update framework-config.json**

Add skill to `projectSkills` array:

```javascript
const config = JSON.parse(fs.readFileSync('framework-config.json'));
config.projectSkills = config.projectSkills || [];
if (!config.projectSkills.includes(skillName)) {
  config.projectSkills.push(skillName);
}
config.projectSkills.sort();
fs.writeFileSync('framework-config.json', JSON.stringify(config, null, 2));
```

**Step 4: Report Results**

```
Enabling skills...

✓ electron-development - Enabled (12 resources)
✓ playwright-setup - Enabled (2 resources)
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

**Messages:**
- `✓ {skill} - Enabled (n resources)` - Skill added to projectSkills
- `⊘ {skill} - Already enabled (skipped)` - Skill already in projectSkills
- `✗ {skill} - Not found in registry` - Skill doesn't exist

---

## Edge Cases

| Scenario | Handling |
|----------|----------|
| `framework-config.json` missing | Error: "Run framework installer first" |
| Skill not in registry | Error: "Not found in registry" (continue with others) |
| Already enabled | Skip with message (not an error) |

---

## Examples

### Enable Single Skill

```
> /install-skill electron-development

Enabling skills...

✓ electron-development - Enabled (12 resources)

Enabled: 1  Skipped: 0  Failed: 0
```

### Enable Multiple Skills

```
> /install-skill playwright-setup postgresql-integration error-handling-patterns

Enabling skills...

✓ playwright-setup - Enabled (2 resources)
✓ postgresql-integration - Enabled (3 resources)
✓ error-handling-patterns - Enabled (1 resources)

Enabled: 3  Skipped: 0  Failed: 0
```

### List Available Skills

```
> /install-skill --list

Available Skills:

Name                          Description                                          Status
────────────────────────────────────────────────────────────────────────────────────────────────────
api-versioning                Guide through API versioning strategies              Available
electron-development          Patterns for Electron app development                Enabled
tdd-red-phase                 Guide through RED phase of TDD cycle                 Enabled
...

Total: 23 skills
Enabled (in projectSkills): 6
```

---

## Related

- `/charter` - Automatically suggests and deploys skills during charter creation
- `Overview/Framework-Skills.md` - Complete skills documentation

---

**End of /install-skill Command**
