# Proposal: Beta Deployment Workflow

**Status:** Draft
**Author:** API-Integration-Specialist
**Date:** 2025-12-26

---

## Problem Statement

Currently, `/prepare-release` is designed for production releases that:
1. Merge to `main` branch
2. Create version tags visible in the repository
3. Create GitHub Releases (marked as "Latest")

For beta/preview deployments (like testing the SQLite cache feature), this workflow:
- Pollutes the `main` branch with incomplete features
- Mixes prerelease tags with stable tags
- Shows prereleases prominently on the Releases page

Teams need a way to:
1. Deploy beta versions for testing without polluting main
2. Run `gh pmu` (stable) and `gh pmub` (beta) side-by-side during testing
3. Eventually merge the feature branch to main for stable release

---

## Decided Approach

Based on requirements gathering, the following decisions were made:

| Decision | Choice |
|----------|--------|
| Repository structure | **Monorepo** (single repo, feature branches) |
| Beta workflow | Work on branch → `/prepare-beta` → merge to main when ready |
| Release type | **GitHub Prerelease** (visible but marked "Pre-release") |
| Binary name | **`gh-pmub`** (allows side-by-side with `gh pmu`) |
| Install method | **Install script** (`install-beta.sh` / `install-beta.ps1`) |
| Uninstall method | **Manual** (document `gh extension remove pmub`) |
| Version format | **`v{current}+beta.N`** (e.g., `v0.9.2+beta.1`) |
| Beta numbering | **Auto-increment** (detect existing, increment automatically) |

---

## Workflow

```
feature/sqlite-cache (branch)
    │
    ├── Develop feature
    │
    ▼
/prepare-beta
    │
    ├── Detect current stable version (e.g., v0.9.2)
    ├── Find existing betas, auto-increment (→ v0.9.2+beta.3)
    ├── Build gh-pmub binary
    ├── Create prerelease tag
    ├── Publish GitHub Prerelease with:
    │   ├── gh-pmub binaries (all platforms)
    │   └── install-beta.sh / install-beta.ps1
    └── Output install instructions

    │
    ▼
[Test with gh pmub alongside gh pmu]
    │
    ▼
Merge feature branch to main
    │
    ▼
/prepare-release v1.0.0
    │
    └── Normal release → gh pmu (stable)
```

---

## Side-by-Side Installation

### Extension Directory Structure

```
~/.local/share/gh/extensions/
├── gh-pmu/
│   └── gh-pmu.exe       # stable (v0.9.2)
└── gh-pmub/
    └── gh-pmub.exe      # beta (v0.9.2+beta.1)
```

### Install Script

`install-beta.sh` (Linux/macOS):
```bash
#!/bin/bash
VERSION=${1:-latest}
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Download gh-pmub binary from prerelease
curl -L "https://github.com/rubrical-studios/gh-pmu/releases/download/${VERSION}/gh-pmub_${OS}_${ARCH}.tar.gz" | tar xz

# Install to gh extensions directory
mkdir -p ~/.local/share/gh/extensions/gh-pmub
mv gh-pmub ~/.local/share/gh/extensions/gh-pmub/

echo "Installed gh-pmub ${VERSION}"
echo "Run: gh pmub --version"
```

`install-beta.ps1` (Windows):
```powershell
param([string]$Version = "latest")

$ExtPath = "$env:LOCALAPPDATA\gh\extensions\gh-pmub"
New-Item -ItemType Directory -Force -Path $ExtPath

# Download and extract
$Url = "https://github.com/rubrical-studios/gh-pmu/releases/download/$Version/gh-pmub_windows_amd64.zip"
Invoke-WebRequest -Uri $Url -OutFile "$env:TEMP\gh-pmub.zip"
Expand-Archive -Path "$env:TEMP\gh-pmub.zip" -DestinationPath $ExtPath -Force

Write-Host "Installed gh-pmub $Version"
Write-Host "Run: gh pmub --version"
```

### Uninstall (Manual)

```bash
# Linux/macOS
rm -rf ~/.local/share/gh/extensions/gh-pmub

# Windows
Remove-Item -Recurse "$env:LOCALAPPDATA\gh\extensions\gh-pmub"

# Or via gh CLI
gh extension remove pmub
```

---

## Version Format

**Format:** `v{current_stable}+beta.{N}`

**Examples:**
- Stable: `v0.9.2`
- First beta: `v0.9.2+beta.1`
- Second beta: `v0.9.2+beta.2`
- After merge, new stable: `v1.0.0`

**Auto-increment logic:**
```bash
# Get current stable version
STABLE=$(git describe --tags --abbrev=0 --match "v[0-9]*" | grep -v "+beta")

# Find highest beta for this stable
LATEST_BETA=$(gh release list --json tagName -q '.[] | select(.tagName | startswith("'$STABLE'+beta")) | .tagName' | sort -V | tail -1)

# Increment
if [ -z "$LATEST_BETA" ]; then
  NEXT="$STABLE+beta.1"
else
  N=$(echo $LATEST_BETA | grep -oP '\+beta\.\K\d+')
  NEXT="$STABLE+beta.$((N+1))"
fi
```

---

## Config Compatibility

Both `gh pmu` and `gh pmub` read the same `.gh-pmu.yml` config file:

```yaml
# .gh-pmu.yml - shared by both extensions
project:
  owner: rubrical-studios
  number: 1
repositories:
  - rubrical-studios/gh-pmu
```

No separate config needed for beta testing.

---

## /prepare-beta Command

New command: `.claude/commands/prepare-beta.md`

**Steps:**
1. Verify not on `main` branch
2. Determine current stable version
3. Auto-increment beta number
4. Build `gh-pmub` binary (via GoReleaser or manual)
5. Create prerelease tag
6. Create GitHub Prerelease (marked as pre-release)
7. Include install scripts in release assets
8. Output install instructions

---

## Acceptance Criteria

- [ ] `/prepare-beta` command created
- [ ] Beta deployment works from any non-main branch
- [ ] Main branch remains untouched
- [ ] Prerelease tag created (e.g., `v0.9.2+beta.1`)
- [ ] GitHub Release marked as prerelease (not "Latest")
- [ ] Binary built as `gh-pmub`
- [ ] Install scripts provided (`install-beta.sh`, `install-beta.ps1`)
- [ ] `gh pmu` and `gh pmub` can run simultaneously
- [ ] Both commands read same `.gh-pmu.yml` config
- [ ] Beta number auto-increments
- [ ] Uninstall instructions documented

---

## First Use Case: SQLite Cache

The first feature to test with this workflow will be the **Local SQLite Cache** (Proposal #455):
- Work on `feature/sqlite-cache` branch
- Deploy as `gh pmub` for testing
- Test cache functionality alongside stable `gh pmu`
- Merge to main when validated

---

**End of Proposal**
