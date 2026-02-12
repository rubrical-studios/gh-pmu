# Claude Code - Project Instructions

**Process Framework:** IDPF-Agile
**Domain Specialist:** Backend-Specialist

---

## Rules Auto-Loading

Rules are automatically loaded from `.claude/rules/`:
- `01-anti-hallucination.md` - Software development quality rules
- `02-github-workflow.md` - GitHub issue management integration (if enabled)
- `03-startup.md` - Session initialization and specialist loading
- `04-charter-enforcement.md`
- `05-windows-shell.md`
- `06-runtime-triggers.md`

**No manual file reading required at startup.**

---

## Framework Configuration

This project uses the IDPF Framework ecosystem.
**Configuration:** See `framework-config.json` for framework location and project type.
**Framework Path:** `E:\Projects\idpf-central-hub`

---

## On-Demand Documentation

Load detailed documentation when needed:

| When Working On | Load File |
|-----------------|-----------|
| Framework workflow | `E:\Projects\idpf-central-hub/IDPF-Agile/Agile-Core.md` |
| Domain specialist | `E:\Projects\idpf-central-hub/System-Instructions/Domain/Base/{specialist}.md` |
| Testing patterns | `.claude/skills/test-writing-patterns/SKILL.md` |

---

## Project-Specific Instructions

<!-- Add your project-specific instructions below this line -->

On session startup:

1. display a random quote, then
2. paint a pretty ascii picture 

<!-- These will be preserved during framework updates -->

---

**End of Claude Code Instructions**
