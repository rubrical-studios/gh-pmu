---
version: 0.16.0
description: Open a new release with branch and tracker
argument-hint: <version> (e.g., v1.2.0)
---

<!-- EXTENSIBLE: v0.16.0 -->
# /open-release

Opens a new release branch and creates a tracker issue.

## Available Extension Points

| Point | Location | Purpose |
|-------|----------|---------|
| `pre-create` | Before branch creation | Custom validation |
| `post-create` | After release created | Notifications, CI triggers |

---

## Prerequisites

- `gh pmu` extension installed
- `.gh-pmu.yml` configured
- Git working directory clean

---

## Workflow

### Step 1: Validate Arguments

Version must be provided (e.g., `v1.2.0`). Add `v` prefix if missing.

### Step 2: Check Working Directory

```bash
git status --porcelain
```

<!-- USER-EXTENSION-START: pre-create -->
### Verify Config

```bash
node .claude/scripts/open-release/verify-config.js
```

The script outputs JSON: `{"success": true/false, "message": "..."}`

**If `success` is false, STOP and report the error.**
<!-- USER-EXTENSION-END: pre-create -->

### Step 3: Create Release

```bash
gh pmu release start --branch "release/$VERSION"
```

### Step 4: Switch to Release Branch

```bash
git checkout "release/$VERSION"
```

<!-- USER-EXTENSION-START: post-create -->
<!-- USER-EXTENSION-END: post-create -->

### Step 5: Report Completion

```
Release $VERSION opened.

Branch: release/$VERSION
Tracker: #[issue-number]

Next steps:
1. Assign issues: gh pmu move [#] --release current
2. Work issues: work #N
3. When ready: /prepare-release $VERSION
4. After deploy: /close-release
```

---

**End of Open Release**
