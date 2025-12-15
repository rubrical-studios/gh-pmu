# Workflow Commands

gh-pmu provides three workflow command groups for managing development at different cadences:

| Workflow | Cadence | Use Case |
|----------|---------|----------|
| `microsprint` | Hours | AI-assisted development batches, rapid iteration |
| `release` | Days/Weeks | Version-based deployment, feature releases |
| `patch` | Hours/Days | Hotfix deployment, urgent bug fixes |

## Microsprint

Microsprints are hour-scale work batches designed for AI-assisted development workflows. They help organize focused work sessions with automatic artifact generation.

### Starting a Microsprint

```bash
# Start with auto-generated name (YYYY-MM-DD-a pattern)
gh pmu microsprint start

# Start with custom suffix
gh pmu microsprint start --name "auth-refactor"
# Creates: Microsprint: 2025-01-15-auth-refactor
```

A tracker issue is created with the `microsprint` label to track the session.

### Managing Issues

```bash
# Add issue to current microsprint
gh pmu microsprint add 42

# Remove issue from microsprint
gh pmu microsprint remove 42

# Create new issue directly in microsprint
gh pmu create --title "Fix login bug" --microsprint current
```

### Viewing Status

```bash
# Show current microsprint details
gh pmu microsprint current

# Update tracker issue body with latest progress
gh pmu microsprint current --refresh

# List all microsprints (open and closed)
gh pmu microsprint list
```

### Closing with Artifacts

```bash
# Interactive close with retrospective prompts
gh pmu microsprint close
# Prompts for: What Went Well, What Could Be Improved, Action Items

# Skip prompts, generate empty template
gh pmu microsprint close --skip-retro

# Auto-commit generated artifacts
gh pmu microsprint close --commit
```

**Generated artifacts:**
```
Microsprints/{name}/
  review.md    # Issue summary with status
  retro.md     # Retrospective notes
```

### Multi-User Support

Microsprints support team-wide workflows:

```bash
# If another user has an active microsprint, you'll be prompted:
# - Join their microsprint
# - Work without a microsprint
# - Cancel

# Resolve multiple active microsprints
gh pmu microsprint resolve
```

---

## Release

Releases are version-based deployment workflows for feature releases. They follow semantic versioning and generate deployment artifacts.

### Starting a Release

```bash
# Interactive version prompt (shows latest git tag)
gh pmu release start

# Specify version directly
gh pmu release start --version 1.2.0

# Specify release track
gh pmu release start --version 1.2.0 --track main
gh pmu release start --version 1.2.0 --track lts
```

A tracker issue is created with the `release` label.

### Managing Issues

```bash
# Add issue to current release
gh pmu release add 42

# Remove issue from release
gh pmu release remove 42

# Create new issue directly in release
gh pmu create --title "Add dark mode" --status in_progress
gh pmu release add <new-issue-number>
```

### Viewing Status

```bash
# Show current release details
gh pmu release current

# List all releases
gh pmu release list
```

### Closing a Release

```bash
# Close release and generate artifacts
gh pmu release close
```

**Generated artifacts** (configurable):
```
Releases/{version}/
  release-notes.md    # Summary of included issues
  changelog.md        # Changes for this version
```

### Release Tracks

Configure different release tracks for multi-branch projects:

```yaml
# .gh-pmu.yml
release:
  tracks:
    - name: main
      prefix: ""           # Artifacts: Releases/v1.2.0/
    - name: lts
      prefix: "lts-"       # Artifacts: Releases/lts-v1.2.0/
```

---

## Patch

Patches are hotfix deployment workflows for urgent bug fixes. They include validation to ensure patches only contain appropriate changes.

### Starting a Patch

```bash
# Start patch with version
gh pmu patch start --version 1.2.1
```

Version must be a valid patch increment from the latest git tag.

### Managing Issues

```bash
# Add issue to patch (with validation)
gh pmu patch add 42

# Skip validation warnings
gh pmu patch add 42 --force

# Remove issue from patch
gh pmu patch remove 42
```

**Validation rules:**
- **Error**: Issue has `breaking-change` label (incompatible with patches)
- **Warning**: Issue lacks `bug`, `fix`, or `hotfix` label

### Viewing Status

```bash
# Show current patch details
gh pmu patch current

# List all patches
gh pmu patch list
```

### Closing a Patch

```bash
# Close patch and generate artifacts
gh pmu patch close

# Close and create git tag
gh pmu patch close --tag
```

---

## Configuration

### Project Fields

Workflows require project fields to function. Run `gh pmu init` to auto-create:

| Field | Type | Purpose |
|-------|------|---------|
| `Microsprint` | Text | Track microsprint assignment |
| `Release` | Text | Track release assignment |

### Labels

Workflows use labels for tracker issues:

| Label | Used By |
|-------|---------|
| `microsprint` | Microsprint tracker issues |
| `release` | Release tracker issues |
| `patch` | Patch tracker issues |

Run `gh pmu init` to auto-create these labels.

### Release Artifacts

Configure artifact generation in `.gh-pmu.yml`:

```yaml
release:
  artifacts:
    directory: "Releases"      # Base directory (default)
    release_notes: true        # Generate release-notes.md
    changelog: true            # Generate changelog.md

  tracks:
    - name: main
      prefix: ""               # No prefix: Releases/v1.2.0/
    - name: lts
      prefix: "lts-"           # Prefixed: Releases/lts-v1.2.0/

  # Active releases (auto-synced by init)
  active:
    - "1.2.0"
```

---

## Workflow Integration

### With Move Command

```bash
# Move issue and assign to current microsprint
gh pmu move 42 --status in_progress --microsprint current
```

### With Create Command

```bash
# Create issue in current microsprint
gh pmu create --title "New feature" --microsprint current
```

### Checking Active Workflows

```bash
# See what workflows are active
gh pmu microsprint current
gh pmu release current
gh pmu patch current
```

---

## See Also

- [Commands Reference](commands.md) - Full command documentation
- [Configuration Guide](configuration.md) - `.gh-pmu.yml` setup
- [Sub-Issues Guide](sub-issues.md) - Hierarchy management
