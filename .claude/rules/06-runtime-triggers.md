# Runtime Artifact Triggers
**Version:** v0.29.2
**Source:** Reference/Runtime-Artifact-Triggers.md
**Purpose:** Define when to offer artifact creation/update during development
---
## Overview
Runtime artifact population happens **during development**. Claude offers to document decisions, plans, and procedures when they arise naturally.
**Key Principle:** Offer, don't force. Users can decline any artifact creation.
---
## Construction/ Artifacts
### Design-Decisions/
**Trigger:** User makes or discusses architectural choice
| Trigger Pattern | Example Phrases |
|-----------------|-----------------|
| Technology selection | "Let's use Redis for caching", "We should go with PostgreSQL" |
| Architecture decision | "We'll use microservices", "Let's add a message queue" |
| Pattern adoption | "We'll implement repository pattern", "Using event sourcing" |
| Trade-off discussion | "We chose X over Y because...", "The trade-off is..." |
**Detection:** Technology comparison, "decided to"/"going with"/"chose", architecture terms, design pattern names
**Template:** `Construction/Design-Decisions/YYYY-MM-DD-[decision-name].md`
### Test-Plans/
**Trigger:** User discusses testing approach or strategy
| Trigger Pattern | Example Phrases |
|-----------------|-----------------|
| Test strategy | "We should test this with...", "For testing, let's..." |
| Coverage discussion | "We need unit tests for...", "Integration tests should cover..." |
| Test data | "Test data should include...", "Edge cases to test..." |
| Test environment | "We'll need staging for...", "Mock the external API" |
**Template:** `Construction/Test-Plans/[feature-name]-test-plan.md`
### Sprint-Retros/
**Trigger:** After `/microsprint close` command
**Auto-generated** from conversation during microsprint.
**Template:** `Construction/Sprint-Retros/[date]-[name]-retro.md`
### Tech-Debt/
**Trigger:** User defers work or notes shortcuts
| Trigger Pattern | Example Phrases |
|-----------------|-----------------|
| Deferral | "We'll fix that later", "TODO:", "FIXME:" |
| Shortcuts | "Quick fix for now", "Temporary workaround" |
| Known issues | "This isn't ideal but...", "We'll need to refactor" |
| Performance notes | "This could be optimized", "Not efficient but works" |
**Template:** `Construction/Tech-Debt/[item-name].md`
---
## Transition/ Artifacts
### Deployment-Guide.md
**Trigger:** During `/prepare-release` when missing or stale
| Condition | Action |
|-----------|--------|
| Does not exist | "Deployment guide is missing. Create now?" |
| Exists but old (>30 days) | "Deployment guide may be stale. Review/update?" |
**Template:** `Transition/Deployment-Guide.md`
### Runbook.md
**Trigger:** When operational procedures discussed
| Trigger Pattern | Example Phrases |
|-----------------|-----------------|
| Operations | "To restart the service...", "If the database goes down..." |
| Monitoring | "Watch for...", "Alert threshold is..." |
| Incident response | "When this happens, do...", "Escalation path is..." |
| Maintenance | "Monthly we need to...", "Backup procedure is..." |
**Template:** `Transition/Runbook.md`
### User-Documentation.md
**Trigger:** When user-facing features complete
| Trigger Pattern | Example Phrases |
|-----------------|-----------------|
| Feature complete | "That feature is done", issue moved to Done |
| UI changes | "The new button...", "Users will see..." |
| API changes | "The endpoint now...", "New parameter added" |
| Workflow changes | "Users can now...", "The new process is..." |
**Template:** `Transition/User-Documentation.md`
---
## Behavior Rules
| Behavior | Implementation |
|----------|----------------|
| Always offer | Present option, accept "no" gracefully |
| Never force | User can decline any artifact creation |
| Remember preference | If user declines type repeatedly, reduce frequency |
| Context-aware | Only offer when genuinely relevant |
**After User Accepts:**
1. Generate artifact using template
2. Show user the content
3. Ask for any edits
4. Save to appropriate location
5. Report: "Saved to [path]"
---
## Token Budget
| Trigger Check | Tokens |
|---------------|--------|
| Pattern detection | ~0 (in-conversation) |
| Template loading | ~200-400 per template |
| Artifact generation | ~300-600 per artifact |
---
**End of Runtime Artifact Triggers**
