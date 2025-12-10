# Claude Code - Project Instructions

**Purpose:** Automatic initialization with IDPF Framework integration
**Process Framework:** IDPF-Agile
**Domain Specialists:** Backend-Specialist, API-Integration-Specialist, Security-Engineer, PRD-Analyst
**Primary Specialist:** API-Integration-Specialist

---

## Framework Configuration

This project uses the IDPF Framework ecosystem.
**Configuration:** See `framework-config.json` for framework location and project type.

---

## Startup Procedure

When starting a new session in this repository, **IMMEDIATELY** perform these steps:

### Step 1: Confirm Date

State the date from your environment information and ask the user to confirm it is correct. **Wait for the user to respond before proceeding to Step 2.**

```
"According to my environment information, today's date is YYYY-MM-DD. Is this correct?"
```

If the user responds "no", prompt for the correct date in YYYY-MM-DD format.

This ensures accurate timestamps in commits and documentation.

### Step 2: Load Configuration

Read `framework-config.json` to get the `frameworkPath`.

### Step 3: Load Startup Instructions and Framework Core

Read these files in order:
1. `STARTUP.md` - Condensed essential rules and guidelines
2. `E:\Projects\process-docs/IDPF-Agile/Agile-Core.md` - Core framework workflow


### Step 4: Load Primary Domain Specialist

Read `framework-config.json` to get the `primarySpecialist` value.

If a primary specialist is configured (not null):
1. Read the specialist file: `E:\Projects\process-docs/System-Instructions/Domain/[primarySpecialist].md`
2. Note the active role for your ready message

If no primary specialist is configured, skip this step.

### Step 5: Read GitHub Workflow Integration (MUST READ)

Read the GitHub Workflow file to activate issue management:

```
E:\Projects\process-docs/Reference/GitHub-Workflow.md
```

If `.gh-pmu.yml` does not exist, ask user if they have a GitHub repo and project.
If yes, run `gh pmu init`. If no, skip GitHub integration.

### Step 6: Confirm Ready

Confirm initialization is complete and ask the user what they would like to work on.
If a primary specialist was loaded, include it in your ready message: "Active Role: [specialist-name]"

**Do NOT proceed with any other work until the startup sequence is complete.**

---

## Available Commands

After completing the startup procedure, display available commands:

| Command | Purpose |
|---------|---------|
| `/switch-role` | Switch active domain specialist mid-session |
| `/add-role` | Add a new domain specialist to your project |


---

## Post-Compact Procedure

**MUST REREAD AFTER COMPACTION:** After any compact operation (manual or automatic), immediately re-read:

```
E:\Projects\process-docs/Reference/GitHub-Workflow.md
```

This ensures GitHub workflow rules persist across context resets.

---

## Project-Specific Instructions

<!-- Add your project-specific instructions below this line -->

paint a pretty ascii picture on session startup

<!-- These will be preserved during framework updates -->

---

**End of Claude Code Instructions**
