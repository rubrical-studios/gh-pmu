---
description: Prepare release with PR, merge to main, and tag
argument-hint: [version] (e.g., v1.2.0)
---

# Prepare Release

Validate, create PR to main, merge, and tag for deployment.

## Pre-Checks

1. **Verify on release branch:**
   ```bash
   git branch --show-current
   ```
   Must be on a release branch (e.g., `release/v1.2.0`), not `main`.

2. **Check for open sprints in this release:**
   ```bash
   gh pmu microsprint current
   ```
   If sprints are open, close them first:
   ```bash
   gh pmu microsprint close
   ```

3. **Check for incomplete issues in release:**
   ```bash
   gh pmu release current --json issues | jq '.[] | select(.status != "done")'
   ```
   If incomplete issues exist, prompt user:
   - Transfer to next release
   - Return to backlog
   - Block release (cannot close with open issues)

## Phase 1: Version Updates

1. Determine version from `$ARGUMENTS` or prompt user
2. Update version files:
   - `package.json` / `Cargo.toml` / etc. (project-specific)
   - `CHANGELOG.md` — add new version section
3. **ASK USER:** Confirm version updates

## Phase 2: Git Preparation

1. Commit version updates:
   ```bash
   git add -A
   git commit -m "chore: prepare release $VERSION"
   ```

2. Create PR from release branch to main:
   ```bash
   gh pr create --base main --head $(git branch --show-current) \
     --title "Release $VERSION" \
     --body "## Release $VERSION\n\nSee CHANGELOG.md for details."
   ```

3. **ASK USER:** Approve and merge PR (or wait for CI)

4. After PR merged, switch to main and pull:
   ```bash
   git checkout main
   git pull origin main
   ```

## Phase 3: Tag on Main

1. Create tag on main:
   ```bash
   git tag -a $VERSION -m "Release $VERSION"
   git push origin $VERSION
   ```

2. **ASK USER:** Confirm tag pushed (triggers deployment if CI configured)

## Phase 4: Verification

1. Verify tag exists:
   ```bash
   git tag -l "$VERSION"
   ```

2. Verify deployment (if applicable):
   - Check CI/CD pipeline status
   - Verify deployed version

## Next Step

After deployment is verified, run `/close-release` to:
- Create GitHub Release page
- Close gh-pmu tracker issue
- Delete release branch

## Post-Release Reminder

**Releasing code does NOT close related issues.**
Issues included in this release still require explicit user approval ("Done") to close.
Do NOT auto-close issues just because they shipped.

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
/prepare-release v1.2.0     ◄── YOU ARE HERE
    └── PR → merge → tag → deploy
         │
         ▼
/close-release
    └── GitHub Release → cleanup
```

---

**End of Prepare Release**
