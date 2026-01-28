<!-- Archived: 2026-01-27T23:24:25.962Z -->
<!-- Original: E:\Projects\gh-pmu\.claude\commands\create-branch.md -->
<!-- Reason: Rogue edits detected outside USER-EXTENSION blocks -->

---
version: "v0.33.2"
description: Create a branch with tracker issue (project)
argument-hint: <branch-name> (e.g., release/v0.16.0, my-feature, bugfix-123)
---

<!-- EXTENSIBLE -->
# /create-branch

Creates a new branch and associated tracker issue for any branch type (release, patch, feature, hotfix, etc.).

## Available Extension Points

| Point | Location | Purpose |
|-------|----------|---------|
| `pre-create` | Before branch creation | Custom validation, environment checks |
| `post-create` | After branch created | Notifications, setup scripts, CI triggers |

---

## Prerequisites

- `gh pmu` extension installed
- `.gh-pmu.yml` configured in repository root
- Git working directory clean (no uncommitted changes)

---

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `$1` | Yes | Branch name (any valid git branch name) |

---

## Workflow

### Step 1: Validate Arguments

Branch name must be a valid git branch name (no spaces, no special characters that git rejects).

If invalid or empty, report error and stop.

### Step 2: Check Working Directory

```bash
git status --porcelain
```

If changes exist, prompt to commit or stash.

<!-- USER-EXTENSION-START: pre-create -->
### Verify Config
```bash
git status --porcelain .gh-pmu.yml
```
**If `success` is false, STOP.**
<!-- USER-EXTENSION-END: pre-create -->

### Step 3: Create Branch with Tracker

```bash
gh pmu branch start --name "$BRANCH"
```

This creates:
- Git branch named `$BRANCH`
- Tracker issue titled "Branch: $BRANCH" with `branch` label

### Step 4: Switch to Branch

```bash
git checkout "$BRANCH"
```

### Step 5: Push Branch to Remote

```bash
git push -u origin "$BRANCH"
```

This establishes the remote branch and sets up upstream tracking.

<!-- USER-EXTENSION-START: post-create -->
<!-- USER-EXTENSION-END: post-create -->

### Step 6: Report Completion

```
Branch created.

Branch: $BRANCH
Tracker: #[tracker-issue-number]

Next steps:
1. Assign issues: /assign-branch #N #N ...
2. Work issues: work #N
3. When ready: /prepare-release
```

---

**End of Create Branch**
