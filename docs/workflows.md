# Workflow Commands

gh-pmu provides workflow command groups for managing development at different cadences:

| Workflow | Cadence | Use Case |
|----------|---------|----------|
| `microsprint` | Hours | AI-assisted development batches, rapid iteration |
| `branch` | Days/Weeks | Branch-based deployment, feature releases, patches, hotfixes |

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

## Branch

Branch workflows are used for feature releases, patches, and hotfixes. The branch name is used literally for all artifacts.

### Starting a Branch

```bash
# Start a feature release
gh pmu branch start --branch release/v2.0.0

# Start a patch release
gh pmu branch start --branch patch/v1.9.1

# Start a hotfix
gh pmu branch start --branch hotfix-auth-bypass
```

The command creates the git branch and a tracker issue with the `branch` label.

### Managing Issues

```bash
# Add issue to current branch
gh pmu branch add 42

# Remove issue from branch
gh pmu branch remove 42

# Create new issue directly in branch
gh pmu create --title "Add dark mode" --status in_progress --branch current
```

### Viewing Status

```bash
# Show current branch details
gh pmu branch current

# List all branches
gh pmu branch list

# Reopen a closed branch
gh pmu branch reopen release/v1.9.0
```

### Closing a Branch

```bash
# Close branch and generate artifacts
gh pmu branch close

# Close with git tag
gh pmu branch close --tag
```

**Generated artifacts** (configurable):
```
Releases/{branch}/
  release-notes.md    # Summary of included issues
  changelog.md        # Changes for this version
```

Examples:
- `Releases/release/v2.0.0/`
- `Releases/patch/v1.9.1/`
- `Releases/hotfix-auth-bypass/`

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
| `branch` | Branch tracker issues |

Run `gh pmu init` to auto-create these labels.

### Branch Artifacts

Configure artifact generation in `.gh-pmu.yml`:

```yaml
branch:
  artifacts:
    directory: "Releases"      # Base directory (default)
    release_notes: true        # Generate release-notes.md
    changelog: true            # Generate changelog.md
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
gh pmu branch current
```

---

## See Also

- [Commands Reference](commands.md) - Full command documentation
- [Configuration Guide](configuration.md) - `.gh-pmu.yml` setup
- [Sub-Issues Guide](sub-issues.md) - Hierarchy management
