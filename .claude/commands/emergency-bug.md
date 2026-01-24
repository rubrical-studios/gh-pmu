---
version: "v0.31.0"
description: Create P0 emergency issue (project)
argument-hint: "<description>"
---
<!-- MANAGED -->
# /emergency-bug
Create a high-priority (P0) issue with emergency label for critical bugs requiring immediate attention.
## Arguments
| Argument | Description |
|----------|-------------|
| `<description>` | Brief description of the emergency issue |
## Usage
```bash
/emergency-bug Production API returning 500 errors
/emergency-bug "Database connection timeouts affecting users"
```
## Workflow
### Step 1: Validate Description
**If no description:** **ASK USER:** What is the emergency issue?
### Step 2: Create Issue
```bash
gh pmu create --repo {repository} \
  --title "🚨 Emergency: {description}" \
  --label "bug" \
  --label "emergency" \
  --body "{emergency_body}" \
  --status ready \
  --priority p0 \
  --assignee @me
```
**Issue Body Template:**
```markdown
## 🚨 Emergency Bug
### Description
{description}
### Reported
**Date:** {date}
**Reported by:** {user}
### Impact
- [ ] Production affected
- [ ] Users impacted
- [ ] Data at risk
### Triage Checklist
- [ ] Severity confirmed
- [ ] Reproduction steps identified
- [ ] Workaround available?
- [ ] Root cause investigation started
### Priority
**P0 - Must Have** - Emergency requiring immediate attention.
### Resolution
_To be updated during fix._
```
### Step 3: Report Creation
```
🚨 Emergency issue created: #{issue_num}
Title: {title}
Priority: P0 (highest)
Status: Ready
Labels: bug, emergency
Next steps:
1. Investigate immediately
2. Update triage checklist
3. Work the issue: work #{issue_num}
```
## When to Use
**Use for:** Production outages, security vulnerabilities, data loss, service-critical failures
**Not for:** Feature requests, minor bugs, planned work
**End of /emergency-bug Command**
