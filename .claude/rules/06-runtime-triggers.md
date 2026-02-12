# Runtime Artifact Triggers
**Version:** v0.42.0
**Source:** Reference/Runtime-Artifact-Triggers.md
**Purpose:** Define when to offer artifact creation/update during development and transition phases.
**Key Principle:** Offer, don't force. Users can decline any artifact creation.
## Construction/ Artifacts
### Design-Decisions/
**Trigger:** User makes or discusses architectural choice
| Pattern | Examples |
|---------|----------|
| Technology selection | "Let's use Redis", "We should go with PostgreSQL" |
| Architecture decision | "Microservices approach", "Add a message queue" |
| Pattern adoption | "Repository pattern", "Event sourcing" |
| Trade-off discussion | "We chose X over Y because..." |
**Offer:** "Would you like me to document this decision in Construction/Design-Decisions/?"
### Test-Plans/
**Trigger:** User discusses testing approach or strategy
| Pattern | Examples |
|---------|----------|
| Test strategy | "We should test this with...", "For testing, let's..." |
| Coverage discussion | "We need unit tests for...", "Integration tests should cover..." |
| Test data | "Edge cases to test..." |
**Offer:** "Would you like me to capture this test plan in Construction/Test-Plans/?"
### Tech-Debt/
**Trigger:** User defers work or notes shortcuts
| Pattern | Examples |
|---------|----------|
| Deferral | "We'll fix that later", "TODO:", "FIXME:" |
| Shortcuts | "Quick fix for now", "Temporary workaround" |
| Known issues | "This isn't ideal but...", "We'll need to refactor" |
**Offer:** "Would you like me to track this in Construction/Tech-Debt/?"
## Transition/ Artifacts
### Deployment-Guide.md
**Trigger:** During `/prepare-release` when missing or stale
- File doesn't exist → prompt to create
- Last updated > 30 days → prompt to review
### Runbook.md
**Trigger:** When operational procedures discussed
| Pattern | Examples |
|---------|----------|
| Operations | "To restart the service...", "If the database goes down..." |
| Monitoring | "Watch for...", "Alert threshold is..." |
| Incident response | "When this happens, do..." |
**Offer:** "Would you like me to add this to Transition/Runbook.md?"
### User-Documentation.md
**Trigger:** When user-facing features complete
| Pattern | Examples |
|---------|----------|
| Feature complete | "That feature is done" |
| UI changes | "The new button...", "Users will see..." |
| API changes | "The endpoint now...", "New parameter added" |
**Offer:** "Would you like to update Transition/User-Documentation.md?"
## Behavior Rules
| Behavior | Implementation |
|----------|----------------|
| Always offer | Present option, accept "no" gracefully |
| Never force | User can decline any artifact |
| Remember preference | If user declines repeatedly, reduce frequency |
| Context-aware | Only offer when genuinely relevant |
### Offer Format
```
"I noticed you [made a decision/discussed testing/etc.].
Would you like me to document this in [artifact location]? (yes/no)"
```
---
**End of Runtime Artifact Triggers**
