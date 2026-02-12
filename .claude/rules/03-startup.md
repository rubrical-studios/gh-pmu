# Session Startup (Hub)

**Version:** 0.42.0
**Type:** Central Hub

---

## Startup Sequence

When starting a new session:

1. **Gather Information**: Collect session data (see table below)
2. **Charter Detection**: Check project charter status
3. **Display Session Initialized**: Show consolidated status block
4. **Ask**: What would you like to work on?

### Session Information Sources

| Field | Source |
|-------|--------|
| Date | Environment/system date |
| Repository | `basename $(git rev-parse --show-toplevel)` |
| Branch | `git branch --show-current` + clean/dirty status |
| Process Framework | `framework-config.json` → `processFramework` |
| Framework Version | `framework-manifest.json` → `version` |
| Active Role | `framework-config.json` → `domainSpecialist` |
| Charter Status | `Active` or `Pending` |
| GitHub Workflow | `gh pmu --version` |

---

## Charter Detection (Mandatory)

**Charter is mandatory.** Check for project charter at startup:

1. Check CHARTER.md exists: `test -f CHARTER.md`
2. Check for template placeholders: `/{[a-z][a-z0-9-]*}/`

### Charter Status

- **Active** (exists, no placeholders): Proceed to display
- **Pending** (missing or template): Auto-run `/charter` command

**BLOCKING:** Session startup does not complete until charter is configured.

---

## Display Session Initialized Block

**Date appears ONLY here.** Format:

```
Session Initialized
- Date: {date}
- Repository: {repo-name}
- Branch: {branch} ({clean|dirty})
- Process Framework: {framework}
- Framework Version: {version}
- Active Role: {specialist}
- Charter Status: {Active|Pending}
- GitHub Workflow: Active via gh pmu {version}
```

If Charter Status is Pending, display blocking message and run `/charter`.

---

## On-Demand Loading

| When Needed | Load From |
|-------------|-----------|
| Framework workflow | `E:\Projects\idpf-central-hub/{framework}/` |
| Domain specialist | `E:\Projects\idpf-central-hub/System-Instructions/Domain/Base/{specialist}.md` |
| Skill usage | `.claude/skills/{skill-name}/SKILL.md` |
| Charter management | Run `/charter` command |

---

**End of Session Startup**
