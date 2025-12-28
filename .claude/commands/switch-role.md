---
version: 0.16.0
description: Switch to a different domain specialist role
---

# Switch Domain Specialist Role

Switch to a different domain specialist role and make it the default for future sessions.

## Available Roles

1. API-Integration-Specialist
2. Backend-Specialist
3. Database-Engineer
4. PRD-Analyst
5. Security-Engineer

## Instructions

When invoked:

### Step 1: Read Current Configuration

Read `framework-config.json` to get the current `primarySpecialist` value.

### Step 2: Display Options and Get Selection

Show available roles (mark current primary if set) and ask user to select one:

```
Available roles:
1. API-Integration-Specialist
2. Backend-Specialist
3. Database-Engineer
4. PRD-Analyst
5. Security-Engineer

Current default: [primarySpecialist or "None"]

Select a role (1-5):
```

### Step 3: Update Configuration (Persist Selection)

Edit `framework-config.json` to set the new `primarySpecialist` value.

**Example edit:**
- Change `"primarySpecialist": "Backend-Specialist"` to `"primarySpecialist": "Frontend-Specialist"`

### Step 4: Load New Specialist Instructions

Read the new specialist's instruction file:

- API-Integration-Specialist: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/API-Integration-Specialist.md`
- Backend-Specialist: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Backend-Specialist.md`
- Database-Engineer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Database-Engineer.md`
- PRD-Analyst: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/PRD-Analyst.md`
- Security-Engineer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Security-Engineer.md`

### Step 5: Confirm Switch

**Response format:**
```
⊘ Deactivating: [Previous-Role]

✓ Updated framework-config.json (new default: [New-Role])

Loading [New-Role]...

✓ Now operating exclusively as: [New-Role]
  Focus areas: [from specialist file]

  This role will load automatically in future sessions.
  Previous role instructions are now inactive.

What would you like to work on?
```

## Context Management

Previous role instructions remain in conversation history but are explicitly deprioritized. The new role takes exclusive precedence for all subsequent work.

## File Paths

- API-Integration-Specialist: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/API-Integration-Specialist.md`
- Backend-Specialist: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Backend-Specialist.md`
- Database-Engineer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Database-Engineer.md`
- PRD-Analyst: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/PRD-Analyst.md`
- Security-Engineer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Security-Engineer.md`

## Usage

User says: `/switch-role` or "switch to frontend" or "I need backend help now"

## Natural Language Triggers

- "switch to [role]"
- "I need [role] help"
- "change to [role] mode"
- "activate [role]"
