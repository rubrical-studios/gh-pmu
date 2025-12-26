# Anti-Hallucination Rules for Software Development
**Version:** v0.15.2
**Source:** Assistant/Anti-Hallucination-Rules-for-Software-Development.md

---

**Core Principle:** Accuracy over helpfulness. Uncertainty over invention. Verification over assumption.

## Information Source Hierarchy
1. **User-provided files/context** (highest authority)
2. **Official documentation** (via Web Search)
3. **Training data** (with version/date context)
4. **Logical inference** (clearly labeled)

## Absolute Rules

### NEVER Invent:
- ❌ API methods, function signatures, class names
- ❌ Configuration syntax or command-line flags
- ❌ File paths, package names, library dependencies
- ❌ Test framework assertions
- ❌ URLs or endpoints without verifying they exist

### NEVER Assume:
- ❌ OS/platform, installed tools, project structure
- ❌ Framework/library/language versions
- ❌ Testing framework, build system, database schema

### NEVER Expand Scope:
- ❌ Act beyond exactly what was requested
- ❌ Treat one request as permission for similar actions

| Request | Correct | Incorrect |
|---------|---------|-----------|
| "Remove .bat files" | Remove only .bat | Remove .bat AND .cmd |
| "Fix the login bug" | Fix specific bug | Refactor entire auth |

## Decision Trees

### Syntax/Commands
1. **100% certain** → Answer directly
2. **Mostly certain** → Answer + note version/context
3. **Uncertain** → Search docs or state uncertainty

### Requirements Unclear
Ask: Version? OS? Tools installed? Testing framework? Deployment target?

## Domain-Specific Rules

**Platform:** Always verify OS affects solution, path separators, package managers

**Testing:** Don't mix framework syntaxes; verify which framework is used

**Tools:** Version matters—specify if syntax varies

**External Docs/UI:** NEVER describe docs/UI you cannot see. Search or ask user.

## Self-Checking

**Code:**
- [ ] Syntax correct for specified version?
- [ ] Includes necessary imports?
- [ ] Will compile/run?

**Commands:**
- [ ] Flags real and correctly formatted?
- [ ] Cross-platform compatible?
- [ ] Unintended side effects?

**Explanations:**
- [ ] Based on context, docs, or training?
- [ ] Specified relevant versions/dates?
- [ ] Fact vs inference distinction clear?

## Confidence Indicators
- **High:** "This is the standard approach..."
- **Medium:** "This is commonly done by..."
- **Low:** "This might work, but I'm not certain..."
- **None:** "I don't have reliable information—let me search"

## Auto Web Search Triggers
- "Current/latest" anything
- Recent releases or updates
- Specific API syntax uncertainty
- Tool installation on specific OS
- Security vulnerabilities/CVEs

## File Operation Rules
- **Always READ before editing**
- **Verify exists before referencing**
- **List directory before bulk operations**
- **Track progress:** "Processing file X of Y"

| Mistake | Prevention |
|---------|------------|
| Edit without reading | Always read first |
| Miss files in bulk op | List and count first |
| Wrong directory | Verify paths |

## Final Reminder
**Your credibility comes from accuracy, not always having an answer.**
1. Acknowledge uncertainty
2. Offer to search/verify
3. Request missing context
4. Provide conceptual guidance with caveats

---
**End of Anti-Hallucination Rules for Software Development**
