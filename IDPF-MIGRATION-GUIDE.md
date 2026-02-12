# IDPF Migration Guide: gh-pmu

**Date:** 2026-02-11
**From:** Legacy install (v0.33.3, copy-based)
**To:** Hub install (v0.42.0, symlink-based via `install-project-existing.js`)
**Hub:** `E:\Projects\idpf-central-hub`

---

## Overview

The `install-project-existing.js` script will convert this project from a **copy-based** IDPF installation (where all files were copied into `.claude/`) to a **hub-based** installation (where `rules/`, `hooks/`, `scripts/shared/`, `metadata/`, and `skills/` are symlinked to the central hub, and commands are copied with extension-block preservation).

---

## Before Running the Script

### 1. Commit or Stash All Changes

Ensure a clean working tree. The script modifies `.claude/` extensively.

```bash
git status
git add -A && git commit -m "Pre-IDPF-migration snapshot"
```

### 2. Back Up Extensible Commands

Both `prepare-release.md` and `prepare-beta.md` contain **heavy project-specific customizations** in their `USER-EXTENSION` blocks. Back them up for post-migration diffing:

```bash
mkdir -p /tmp/gh-pmu-backup
cp .claude/commands/prepare-release.md /tmp/gh-pmu-backup/
cp .claude/commands/prepare-beta.md /tmp/gh-pmu-backup/
```

### 3. Sprint Content Will Be Deleted

The following sprint-related files are in `.claude/scripts/shared/` and will be destroyed when the installer replaces that directory with a symlink. **No backup needed -- these are intentionally being removed.**

| File | Disposition |
|------|-------------|
| `end-sprint.js` | Delete (no backup) |
| `plan-sprint.js` | Delete (no backup) |
| `sprint-retro.js` | Delete (no backup) |
| `sprint-status.js` | Delete (no backup) |
| `package.json` | Delete (no backup) |
| `package-lock.json` | Delete (no backup) |
| `node_modules/` | Delete (no backup) |
| `lib/poll.js` | Delete (no backup) |

The following sprint-related **commands** will NOT be deleted by the installer (it only adds hub commands, never removes project-specific ones). **Delete these manually after the script runs:**

- `.claude/commands/end-sprint.md`
- `.claude/commands/plan-sprint.md`
- `.claude/commands/sprint-retro.md`
- `.claude/commands/sprint-status.md`

### 4. Understand Extension Block Preservation (CRITICAL)

The installer's `mergeExtensionBlocks()` function reads the existing file, extracts non-empty `USER-EXTENSION` blocks, then injects them into the new hub version of the command. The command structure updates but your customizations survive.

**`prepare-release.md` has customizations in 10 of 13 extension points:**

| Extension Point | Content |
|-----------------|---------|
| `post-analysis` | Documentation review checklist + E2E impact analysis (`analyze-e2e-impact.js`) |
| `pre-validation` | Incomplete issues handling + lint gate (`lint.js`) |
| `post-validation` | Go test suite + coverage gate (`coverage.js`) + E2E gate (`run-e2e-gate.js`) |
| `post-prepare` | CI wait (`wait-for-ci.js`) |
| `post-pr-create` | CI wait (`wait-for-ci.js`) |
| `pre-tag` | Important rules (never skip CI, never auto-tag, etc.) |
| `post-tag` | Release monitor (`monitor-release.js`), release notes, asset cleanup |
| `checklist-before-tag` | Coverage gate + CI passing |
| `checklist-after-tag` | CI jobs + release assets + release notes |
| `checklist-close` | Tracker closed, release closed, branch deleted, GH release created |

**`prepare-beta.md` has customizations in 8 of 9 extension points:**

| Extension Point | Content |
|-----------------|---------|
| `post-analysis` | Commit analysis + version recommendation + E2E impact analysis |
| `pre-validation` | Lint gate (`lint.js`) |
| `post-validation` | Coverage gate + E2E gate |
| `post-prepare` | CI wait |
| `pre-tag` | Important rules + beta tag authorization (`.release-authorized` file) |
| `post-tag` | Monitor beta build + release notes |
| `checklist-before-tag` | Coverage gate |
| `checklist-after-tag` | Beta build monitored |

**Verified (v0.42.0):** Extension point names in the hub match the project exactly (13 points in `prepare-release`, 9 in `prepare-beta`). No renames or removals detected. The merge should preserve all customizations.

**Residual risk:** If the hub updates in the future and renames extension points, a subsequent re-run of the installer could drop customizations. The post-migration diff (Step 2 in After) will catch this.

**Script path references in extensions:**

Truly project-owned (survive migration as-is):
- `.claude/scripts/e2e/analyze-e2e-impact.js`
- `.claude/scripts/e2e/run-e2e-gate.js`
- `.claude/scripts/prepare-release/coverage.js`
- `.claude/scripts/prepare-release/lint.js`
- `.claude/scripts/close-release/monitor-release.js`

**Wrong path** -- these are duplicates of hub shared scripts placed in `scripts/framework/` from the old install. After migration, fix the references in the extension blocks and delete the directory:

| Current (wrong) | Correct (post-migration) |
|------------------|--------------------------|
| `.claude/scripts/framework/wait-for-ci.js` | `.claude/scripts/shared/wait-for-ci.js` |
| `.claude/scripts/framework/update-release-notes.js` | `.claude/scripts/shared/update-release-notes.js` |
| `.claude/scripts/framework/analyze-commits.js` | `.claude/scripts/shared/analyze-commits.js` |
| `.claude/scripts/framework/recommend-version.js` | `.claude/scripts/shared/recommend-version.js` |

All four exist in the hub's `shared/` directory and will be available via the symlink.

### 5. Note `.claude/.manifest.json`

The project has a `.manifest.json` tracking deployed script checksums (v0.33.3). The installer does **not** update or remove this file. It will become stale after migration. Remove it post-migration.

### 6. Note `framework-config.json` Format Changes

Current format (v0.33.3 legacy):
```json
{
  "frameworkVersion": "0.33.3",
  "extensibleCommands": ["create-branch", ...],
  "frameworkScripts": {},
  "userScripts": { ... },
  "projectType": {
    "processFramework": "IDPF-Agile",
    "domainSpecialist": "Backend-Specialist"
  },
  "frameworkPath": "E:\\Projects\\idpf-praxis-dist"
}
```

The installer will **merge TDD skills** into this file (adding `projectSkills` array) but will NOT restructure it to the new format. The `projectType` nested structure and `frameworkScripts`/`userScripts`/`extensibleCommands` fields are from the old schema.

---

## Run the Script

```bash
node E:/Projects/idpf-central-hub/install-project-existing.js --target E:/Projects/gh-pmu
```

The interactive prompts will ask:
1. **Project name** -- default: `gh-pmu`
2. **Process framework** -- select: `IDPF-Agile`
3. **Domain specialist** -- select: `Backend-Specialist`
4. **GitHub integration** -- will be skipped (`.gh-pmu.yml` already exists)

---

## After Running the Script

### 1. Verify Symlinks Created

Check that these are now junctions pointing to the hub:

```bash
cmd /c "dir /AL .claude"
```

Expected symlinks:
- `.claude/rules/` -> `E:\Projects\idpf-central-hub\.claude\rules\`
- `.claude/hooks/` -> `E:\Projects\idpf-central-hub\.claude\hooks\`
- `.claude/scripts/shared/` -> `E:\Projects\idpf-central-hub\.claude\scripts\shared\`
- `.claude/metadata/` -> `E:\Projects\idpf-central-hub\.claude\metadata\`
- `.claude/skills/` -> `E:\Projects\idpf-central-hub\.claude\skills\`

**Hub metadata is populated (v0.42.0):** The hub's `.claude/metadata/` now contains all 8 registries: `ci-features.json`, `extension-points.json`, `extension-recipes.json`, `recipe-tech-mapping.json`, `review-extensions.json`, `review-mode-criteria.json`, `skill-keywords.json`, `skill-registry.json`. Symlink will provide full metadata access.

**WARNING -- Hub hooks include `pre-push`:** The hub's `.claude/hooks/` contains a `pre-push` git hook in addition to `workflow-trigger.js`. Symlinking hooks will bring the hub's `pre-push` hook into this project. Verify it does not conflict with gh-pmu's release/tagging workflow (the `prepare-beta.md` extension uses a `.release-authorized` file to bypass push restrictions).

### 2. Verify Extension Block Preservation (CRITICAL)

Diff the extensible commands against the backups to confirm customizations survived:

```bash
diff /tmp/gh-pmu-backup/prepare-release.md .claude/commands/prepare-release.md
diff /tmp/gh-pmu-backup/prepare-beta.md .claude/commands/prepare-beta.md
```

Check that all customized extension blocks are present:
- `post-analysis` still has documentation review + E2E impact analysis
- `pre-validation` still has lint gate
- `post-validation` still has coverage gate + E2E gate
- `pre-tag` still has important rules + beta authorization
- `post-tag` still has release monitor + cleanup

If any blocks are missing, restore them manually from the backup files.

### 3. Fix `scripts/framework/` Path References

The extension blocks in `prepare-release.md` and `prepare-beta.md` reference scripts at `.claude/scripts/framework/` -- these are stale copies from the old install. The canonical versions now live in `scripts/shared/` (via hub symlink).

**In `.claude/commands/prepare-release.md`**, find and replace in extension blocks:
- `scripts/framework/wait-for-ci.js` -> `scripts/shared/wait-for-ci.js`
- `scripts/framework/update-release-notes.js` -> `scripts/shared/update-release-notes.js`

**In `.claude/commands/prepare-beta.md`**, find and replace in extension blocks:
- `scripts/framework/analyze-commits.js` -> `scripts/shared/analyze-commits.js`
- `scripts/framework/recommend-version.js` -> `scripts/shared/recommend-version.js`
- `scripts/framework/wait-for-ci.js` -> `scripts/shared/wait-for-ci.js`
- `scripts/framework/update-release-notes.js` -> `scripts/shared/update-release-notes.js`

Then delete the stale directory:

```bash
rm -rf .claude/scripts/framework
```

### 4. Delete Sprint Content

The installer doesn't remove project-specific commands. Delete them manually:

```bash
rm .claude/commands/end-sprint.md
rm .claude/commands/plan-sprint.md
rm .claude/commands/sprint-retro.md
rm .claude/commands/sprint-status.md
```

Also verify the sprint scripts in `scripts/shared/` were removed (they should be gone since `shared/` is now a symlink to the hub).

### 5. Verify New Commands Added

The hub has 30 commands total. The script copies all of them. These 12 are new to this project:

`bug.md`, `ci.md`, `done.md`, `enhancement.md`, `install-skill.md`, `proposal.md`, `resolve-review.md`, `review-issue.md`, `review-prd.md`, `review-proposal.md`, `review-test-plan.md`, `work.md`

Verify project-specific command survived: `emergency-bug.md`

**Note:** Framework-dev-only commands (`audit-hallucination.md`, `audit-minimization.md`, `gap-analysis.md`, `minimize-files.md`, `skill-validate.md`) are NOT in this hub -- they only exist in the `idpf-praxis` dev repo.

### 6. Update `CLAUDE.md`

The installer does NOT modify `CLAUDE.md`. It currently references `E:\Projects\idpf-praxis-dist` with absolute paths. Update:

- Change `frameworkPath` references to `E:\Projects\idpf-central-hub`
- The "Rules Auto-Loading" section lists 3 rules (`01`, `02`, `03`) but the hub provides 6 rules -- update the list to include `04-charter-enforcement.md`, `05-windows-shell.md`, `06-runtime-triggers.md`
- The "On-Demand Documentation" paths should reference the hub path

### 7. Update `framework-config.json`

The installer adds `projectSkills` but doesn't restructure the legacy format. Manually update to the new schema:

```json
{
  "frameworkVersion": "0.42.0",
  "installedDate": "2026-02-11",
  "processFramework": "IDPF-Agile",
  "reviewMode": "solo",
  "frameworkPath": "../idpf-central-hub",
  "domainSpecialist": "Backend-Specialist",
  "extensibleCommands": [
    "create-branch", "prepare-release", "prepare-beta",
    "merge-branch", "destroy-branch", "create-prd",
    "work", "done", "bug", "enhancement", "proposal",
    "review-proposal", "review-prd", "review-issue", "ci"
  ],
  "projectSkills": [
    "tdd-failure-recovery", "tdd-green-phase",
    "tdd-red-phase", "tdd-refactor-phase"
  ]
}
```

**New fields in v0.42.0:**
- `reviewMode`: Controls which review criteria are asked (`solo` for single developer, `team` for small teams, `enterprise` for large orgs). Default: `solo`.

Remove the legacy fields: `frameworkScripts`, `userScripts`, `projectType` (flatten `processFramework` and `domainSpecialist` to top level). Remove `frameworkPath` pointing to old `idpf-praxis-dist`.

### 8. Clean Up Stale Files

| File | Action |
|------|--------|
| `.claude/.manifest.json` | Remove (stale -- tracked copy-based deploys, not relevant with symlinks) |
| `.claude/archive/` | Review and remove if not needed |

### 9. Update `.gitignore`

The installer adds entries for symlinked directories. Verify these were added:

```
.claude/hooks
.claude/metadata
.claude/rules
.claude/scripts/shared
.claude/skills
run_claude.cmd
run_claude.sh
runp_claude.cmd
runp_claude.sh
```

### 10. Verify Extension Directories Created

The installer creates empty extension directories with `.gitkeep`:

- `.claude/extensions/`
- `.claude/scripts/create-branch/` (may already exist)
- `.claude/scripts/prepare-release/` (may already exist)
- `.claude/scripts/work/` (new)
- `.claude/scripts/done/` (new)
- `.claude/scripts/bug/` (new)
- `.claude/scripts/ci/` (new — v0.42.0)
- `.claude/scripts/enhancement/` (new)
- `.claude/scripts/proposal/` (new)
- `.claude/scripts/review-proposal/` (new)
- `.claude/scripts/review-prd/` (new)
- `.claude/scripts/review-issue/` (new)

### 11. Test the Integration

```bash
# Start Claude Code in the project
cd E:/Projects/gh-pmu
claude

# Verify session startup loads rules from hub
# Verify /charter shows the existing charter
# Verify gh pmu commands still work
```

### 12. Commit the Migration

```bash
git add -A
git commit -m "Migrate to IDPF hub-based installation (v0.42.0)"
```

---

## Risk Summary

| Risk | Impact | Mitigation |
|------|--------|------------|
| Extension blocks silently dropped | High | Back up commands before, diff after (Steps 2/2) |
| `framework-config.json` schema mismatch | Low | Manual update post-migration |
| `CLAUDE.md` stale references | Low | Manual update post-migration |
| `.manifest.json` stale | Low | Remove post-migration |
| Project-specific commands overwritten | None | Installer only adds, doesn't remove non-hub commands |

---

## New Capabilities from v0.42.0

After migration, this project gains access to these v0.42.0 features (not available at the v0.38.0 target date):

| Feature | Description |
|---------|-------------|
| `/ci` command | CI workflow management: status, list features, validate YAML, add/remove features, recommendations |
| `/done` diff verification | `done-verify.js` checks for hallucinated completions before closing issues |
| `/done` background CI | `ci-watch.js` monitors CI in background after push, reports pass/fail |
| Review mode | `reviewMode` in `framework-config.json` controls which criteria are asked (`solo`/`team`/`enterprise`) |
| Auto-evaluate reviews | Objective criteria auto-evaluated, only subjective criteria prompt the user |
| `/resolve-review` | Structured resolution of review findings with two-pass processing |
| `/review-test-plan` | Review test plans against their PRD |
| 24 skills | New `drawio-generation` skill for Draw.io XML diagram generation |
| `ci-hints.js` | Contextual hints for `/ci` subcommands |
| `review-mode.js` | Shared lib for review mode configuration queries |

**Note:** The `/ci add` and `/ci remove` commands manage GitHub Actions workflow YAML directly. If gh-pmu uses custom CI steps, review the feature registry (`ci-features.json`) before using `/ci add` to ensure compatibility.

---

**End of Migration Guide**
