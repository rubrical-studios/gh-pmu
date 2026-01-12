---
version: "v0.23.2"
description: Change domain specialist for this project
argument-hint: [specialist-name] (optional)
---
# /change-domain-expert
Change the active domain specialist for this project.
---
## Prerequisites
- Framework v0.17.0+ installed
- `framework-config.json` exists
---
## Available Base Experts
| # | Specialist | Focus Area |
|---|------------|------------|
| 1 | Full-Stack-Developer | End-to-end web development |
| 2 | Backend-Specialist | Server-side systems and APIs |
| 3 | Frontend-Specialist | UI/UX and client-side |
| 4 | Mobile-Specialist | iOS, Android, cross-platform |
| 5 | Desktop-Application-Developer | Native desktop apps |
| 6 | Embedded-Systems-Engineer | Hardware-software integration |
| 7 | Game-Developer | Game engines and graphics |
| 8 | ML-Engineer | Machine learning and AI |
| 9 | Data-Engineer | Data pipelines and warehousing |
| 10 | Cloud-Solutions-Architect | Cloud infrastructure |
| 11 | SRE-Specialist | Reliability and operations |
| 12 | Systems-Programmer-Specialist | Low-level systems |
---
## Workflow
### Step 1: Read Current Configuration
```bash
cat framework-config.json
```
Extract `frameworkPath` and current `projectType.domainSpecialist`.
### Step 2: Select New Specialist
**If argument provided:** Use specified name.
**If no argument:** Present numbered list (1-12).
### Step 3: Validate Selection
Must be one of 12 Base Experts.
### Step 4: Update framework-config.json
Update `projectType.domainSpecialist`.
### Step 5: Update CLAUDE.md
Update `**Domain Specialist:**` line and On-Demand Documentation table.
### Step 6: Update .claude/rules/03-startup.md
Update: metadata line, specialist file path, "Active Role" message.
### Step 7: Load New Specialist
```bash
cat "{frameworkPath}/System-Instructions/Domain/Base/{new-specialist}.md"
```
### Step 8: Report Completion
```
Domain specialist changed successfully.
Previous: {old-specialist}
New: {new-specialist}
```
---
**End of Change Domain Expert**
