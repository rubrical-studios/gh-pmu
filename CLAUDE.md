# Claude Code - Project Instructions

**Process Framework:** IDPF-Agile
**Domain Specialists:** Backend-Specialist, API-Integration-Specialist, Security-Engineer, PRD-Analyst
**Primary Specialist:** API-Integration-Specialist

---

## Rules Auto-Loading

Rules are automatically loaded from `.claude/rules/`:
- `01-anti-hallucination.md` - Software development quality rules
- `02-github-workflow.md` - GitHub issue management integration (if enabled)
- `03-startup.md` - Session initialization and specialist loading

**No manual file reading required at startup.**

---

## Framework Configuration

This project uses the IDPF Framework ecosystem.
**Configuration:** See `framework-config.json` for framework location and project type.
**Framework Path:** `E:\Projects\virtual-ai-studio-dist`

---

## On-Demand Documentation

Load detailed documentation when needed:

| When Working On | Load File |
|-----------------|-----------|
| Framework workflow | `E:\Projects\virtual-ai-studio-dist/IDPF-Agile/Agile-Core.md` |
| Domain specialist | `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/{specialist}.md` |
| Testing patterns | `.claude/skills/test-writing-patterns/SKILL.md` |

---

## Available Commands

| Command | Purpose |
|---------|---------|
| `/switch-role` | Switch active domain specialist mid-session |
| `/add-role` | Add a new domain specialist to your project |


---

## Project-Specific Instructions

<!-- Add your project-specific instructions below this line -->

paint a pretty ascii picture on session startup

<!-- These will be preserved during framework updates -->

---

**End of Claude Code Instructions**
