---
description: Close release with GitHub Release and cleanup
argument-hint: [--skip-release-page]
---

# Close Release

Closes a release by creating a GitHub Release page and cleaning up the release branch.

## Prerequisites

- Active release context (run `gh pmu release current` to verify)
- Release has been deployed (tag exists on main)
- All release issues are Done
- On main branch (post-merge)

## Arguments

| Argument | Required | Description |
|----------|----------|-------------|
| `--skip-release-page` | No | Skip GitHub Release page creation |

## Workflow

### Step 1: Verify Release Context

```bash
gh pmu release current
```

If no release active, error and exit.

Extract track and version from current release:
- `release/v1.2.0` → track=`release`, version=`v1.2.0`
- `patch/v1.1.1` → track=`patch`, version=`v1.1.1`
- `hotfix/v1.0.1` → track=`hotfix`, version=`v1.0.1`

### Step 2: Verify Deployment Complete

```bash
# Check that tag exists
git tag -l "$VERSION"

# Verify on main branch
git branch --show-current
```

If tag missing, warn that `/prepare-release` may not have completed.

### Step 3: Verify All Issues Done

```bash
gh pmu release current --json issues
```

If any issues not Done, warn user and confirm before proceeding.

### Step 4: Create GitHub Release

Unless `--skip-release-page` is specified:

```bash
gh release create "$VERSION" \
  --title "Release $VERSION" \
  --generate-notes
```

### Step 5: Close Tracker Issue

```bash
gh pmu release close
```

### Step 6: Delete Release Branch

```bash
# Delete remote branch
git push origin --delete "release/$VERSION"

# Delete local branch (if not current)
git branch -d "release/$VERSION"
```

### Step 7: Report Completion

Output:
```
Release $VERSION closed.

GitHub Release: https://github.com/[owner]/[repo]/releases/tag/$VERSION

Cleanup:
  - Tracker issue closed
  - Branch release/$VERSION deleted

Release lifecycle complete!
```

---

## Release Lifecycle

```
/open-release v1.2.0
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
/close-release              ◄── YOU ARE HERE
    └── GitHub Release → cleanup
```

---

**End of Close Release**
