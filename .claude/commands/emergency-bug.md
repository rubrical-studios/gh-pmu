---
version: "v0.33.2"
description: Create P0 emergency issue (project)
argument-hint: "<description> (e.g., \"API 500 errors\")"
---

<!-- MANAGED -->
# /emergency-bug

Create a high-priority (P0) issue with emergency label for critical bugs that need immediate attention.

---

## Arguments

| Argument | Description |
|----------|-------------|
| `<description>` | Brief description of the emergency issue |

---

## Usage

```bash
/emergency-bug Production API returning 500 errors
/emergency-bug "Database connection timeouts affecting users"
```

---

## Workflow

### Step 1: Validate Description

**If no description provided:**

**ASK USER:** What is the emergency issue? (Brief description)

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

**P0 - Must Have** - This is an emergency requiring immediate attention.

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

---

## When to Use

Use `/emergency-bug` for:
- Production outages
- Security vulnerabilities (active exploitation)
- Data loss incidents
- Service-critical failures

**Not for:**
- Feature requests (use standard issue)
- Minor bugs (use `/bug` or standard issue)
- Planned work items

---

## Related Commands

| Situation | Command |
|-----------|---------|
| Regular bug report | Standard issue creation |
| Work the emergency | `work #{issue_num}` |
| Close when resolved | `done` |

---

**End of /emergency-bug Command**
