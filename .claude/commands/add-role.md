# Add Domain Specialist Role

Add a new domain specialist to your project that wasn't selected during installation.

## All Available Specialists

1. Accessibility-Specialist
2. API-Integration-Specialist
3. Backend-Specialist
4. Cloud-Solutions-Architect
5. Database-Engineer
6. Data-Engineer
7. Desktop-Application-Developer
8. DevOps-Engineer
9. Embedded-Systems-Engineer
10. Frontend-Specialist
11. Full-Stack-Developer
12. Game-Developer
13. Graphics-Engineer-Specialist
14. ML-Engineer
15. Mobile-Specialist
16. Performance-Engineer
17. Platform-Engineer
18. PRD-Analyst
19. QA-Test-Engineer
20. Security-Engineer
21. SRE-Specialist
22. Systems-Programmer-Specialist
23. Technical-Writer-Specialist

## Currently Installed

Read `framework-config.json` to see which specialists are already in your `domainSpecialists` array.

## Instructions

When invoked:

### Step 1: Read Current Configuration

Read `framework-config.json` to get:
- `frameworkPath` - path to framework files
- `domainSpecialists` - currently installed specialists
- `primarySpecialist` - current default role

### Step 2: Display Available Roles

Show specialists NOT already in `domainSpecialists`:

```
Currently installed: [list from domainSpecialists]

Available to add:
[numbered list of specialists NOT in domainSpecialists]

Select a role to add (number):
```

If all specialists are already installed, inform the user and suggest `/switch-role` instead.

### Step 3: Update Configuration

Edit `framework-config.json`:
1. Add the new specialist to the `domainSpecialists` array
2. Ask if user wants to set it as `primarySpecialist`

**Example edit:**
```json
// Before
"domainSpecialists": ["Backend-Specialist", "Frontend-Specialist"],

// After
"domainSpecialists": ["Backend-Specialist", "Frontend-Specialist", "Security-Engineer"],
```

### Step 4: Load New Specialist Instructions

Read the new specialist's instruction file:

- Accessibility-Specialist: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Accessibility-Specialist.md`
- API-Integration-Specialist: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/API-Integration-Specialist.md`
- Backend-Specialist: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Backend-Specialist.md`
- Cloud-Solutions-Architect: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Cloud-Solutions-Architect.md`
- Database-Engineer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Database-Engineer.md`
- Data-Engineer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Data-Engineer.md`
- Desktop-Application-Developer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Desktop-Application-Developer.md`
- DevOps-Engineer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/DevOps-Engineer.md`
- Embedded-Systems-Engineer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Embedded-Systems-Engineer.md`
- Frontend-Specialist: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Frontend-Specialist.md`
- Full-Stack-Developer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Full-Stack-Developer.md`
- Game-Developer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Game-Developer.md`
- Graphics-Engineer-Specialist: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Graphics-Engineer-Specialist.md`
- ML-Engineer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/ML-Engineer.md`
- Mobile-Specialist: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Mobile-Specialist.md`
- Performance-Engineer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Performance-Engineer.md`
- Platform-Engineer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Platform-Engineer.md`
- PRD-Analyst: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/PRD-Analyst.md`
- QA-Test-Engineer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/QA-Test-Engineer.md`
- Security-Engineer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Security-Engineer.md`
- SRE-Specialist: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/SRE-Specialist.md`
- Systems-Programmer-Specialist: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Systems-Programmer-Specialist.md`
- Technical-Writer-Specialist: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Technical-Writer-Specialist.md`

### Step 5: Confirm Addition

**Response format:**
```
✓ Added Security-Engineer to project

Updated framework-config.json:
  • domainSpecialists: [..., Security-Engineer]
  • primarySpecialist: [unchanged or new value]

Loading Security-Engineer...

✓ Now operating as: Security-Engineer
  Focus areas: [from specialist file]

Use /switch-role to change between installed specialists.
```

## File Paths

- Accessibility-Specialist: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Accessibility-Specialist.md`
- API-Integration-Specialist: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/API-Integration-Specialist.md`
- Backend-Specialist: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Backend-Specialist.md`
- Cloud-Solutions-Architect: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Cloud-Solutions-Architect.md`
- Database-Engineer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Database-Engineer.md`
- Data-Engineer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Data-Engineer.md`
- Desktop-Application-Developer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Desktop-Application-Developer.md`
- DevOps-Engineer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/DevOps-Engineer.md`
- Embedded-Systems-Engineer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Embedded-Systems-Engineer.md`
- Frontend-Specialist: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Frontend-Specialist.md`
- Full-Stack-Developer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Full-Stack-Developer.md`
- Game-Developer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Game-Developer.md`
- Graphics-Engineer-Specialist: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Graphics-Engineer-Specialist.md`
- ML-Engineer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/ML-Engineer.md`
- Mobile-Specialist: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Mobile-Specialist.md`
- Performance-Engineer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Performance-Engineer.md`
- Platform-Engineer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Platform-Engineer.md`
- PRD-Analyst: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/PRD-Analyst.md`
- QA-Test-Engineer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/QA-Test-Engineer.md`
- Security-Engineer: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Security-Engineer.md`
- SRE-Specialist: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/SRE-Specialist.md`
- Systems-Programmer-Specialist: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Systems-Programmer-Specialist.md`
- Technical-Writer-Specialist: `E:\Projects\virtual-ai-studio-dist/System-Instructions/Domain/Technical-Writer-Specialist.md`

## Usage

User says: `/add-role` or "add security specialist" or "I need to add DevOps"

## Natural Language Triggers

- "add [role]"
- "install [role]"
- "I need [role] capabilities"
- "add a new specialist"
