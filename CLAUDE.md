# Claude Code - Project Instructions

**Purpose:** Automatic initialization with IDPF Framework integration
**Process Framework:** IDPF-Agile
**Domain Specialists:** Backend-Specialist, API-Integration-Specialist, PRD-Analyst

---

## Framework Configuration

This project uses the IDPF Framework ecosystem.
**Configuration:** See `framework-config.json` for framework location and project type.

---

## Startup Procedure

When starting a new session in this repository, **IMMEDIATELY** perform these steps:

### Step 1: Load Configuration

Read `framework-config.json` to get the `frameworkPath`.

### Step 2: Load Startup Instructions

Read `STARTUP.md` - this contains condensed essential rules and guidelines.

### Step 3: Configure GitHub Integration (if needed)

If `.gh-pm.yml` does not exist, ask user if they have a GitHub repo and project.
If yes, run `gh pm init`. If no, skip.

If `.claude/commands/gh-workflow.md` has unreplaced placeholders, prompt user for values.

### Step 4: Confirm Ready

Confirm initialization is complete and ask the user what they would like to work on.

**Do NOT proceed with any other work until the startup sequence is complete.**

---

## Expansion Commands

Use these to load full documentation when needed:
- `/expand-rules` - Load complete Anti-Hallucination Rules
- `/expand-framework` - Load full process framework documentation
- `/expand-domain` - Load full Domain Specialist instructions

---

## Project-Specific Instructions

<-- Add your project-specific instructions below this line -->
<-- These will be preserved during framework updates -->

---

**End of Claude Code Instructions**
