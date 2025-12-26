---
description: Open a new release with branch and tracker
argument-hint: <version> (e.g., v1.2.0)
---

# Open Release

Opens a new release branch and creates a tracker issue for the release lifecycle.

## Prerequisites

- `gh pmu` extension installed
- `.gh-pmu.yml` configured in repository root
- Git working directory clean (no uncommitted changes)

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `$ARGUMENTS` | Yes | Version string (e.g., `v1.2.0`) |

## Workflow

### Step 1: Validate Arguments

Version must be provided:
```bash
# If no version provided, prompt user
```

Version should start with `v` (e.g., `v1.2.0`). Add prefix if missing.

### Step 2: Check Working Directory

```bash
git status --porcelain
```

If changes exist, prompt user to commit or stash before proceeding.

### Step 3: Create Release

```bash
gh pmu release start --branch "release/$VERSION"
```

This automatically:
1. Creates git branch `release/$VERSION`
2. Creates tracker issue in GitHub project
3. Sets the release as "current" for subsequent operations

### Step 4: Switch to Release Branch

```bash
git checkout "release/$VERSION"
```

### Step 5: Report Completion

Output:
```
Release $VERSION opened.

Branch: release/$VERSION
Tracker: #[tracker-issue-number]

Next steps:
1. Assign issues: gh pmu move [#] --release current
2. Work issues: work #N
3. When ready: /prepare-release $VERSION
4. After deploy: /close-release
```

---

## Release Lifecycle

```
/open-release v1.2.0        ◄── YOU ARE HERE
    └── Creates branch + tracker
         │
         ▼
    [Work on release branch]
         │
         ▼
/prepare-release v1.2.0
    └── PR → merge → tag → deploy
         │
         ▼
/close-release
    └── GitHub Release → cleanup
```

---

**End of Open Release**
